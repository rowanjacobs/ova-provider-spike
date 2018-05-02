package helper

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
)

// adapted from terraform-provider-vsphere internal methods

type vSphereVersion struct {
	major int
	minor int
	patch int
}

var vmUUIDSearchIndexVersion = vSphereVersion{
	major: 6,
	minor: 5,
}

const DefaultAPITimeout = 5 * time.Minute

func FromUUID(client *govmomi.Client, uuid string) (*object.VirtualMachine, error) {
	log.Printf("[DEBUG] Locating virtual machine with UUID %q", uuid)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultAPITimeout)
	defer cancel()

	var result object.Reference
	version, err := parseVersion(client)
	if err != nil {
		return nil, err
	}
	expected := vmUUIDSearchIndexVersion
	if version.Older(expected) {
		result, err = virtualMachineFromContainerView(ctx, client, uuid)
	} else {
		result, err = virtualMachineFromSearchIndex(ctx, client, uuid)
	}

	if err != nil {
		return nil, err
	}

	// We need to filter our object through finder to ensure that the
	// InventoryPath field is populated, or else functions that depend on this
	// being present will fail.
	finder := find.NewFinder(client.Client, false)

	vm, err := finder.ObjectReference(ctx, result.Reference())
	if err != nil {
		return nil, err
	}

	// Should be safe to return here. If our reference returned here and is not a
	// VM, then we have bigger problems and to be honest we should be panicking
	// anyway.
	log.Printf("[DEBUG] VM %q found for UUID %q", vm.(*object.VirtualMachine).InventoryPath, uuid)
	return vm.(*object.VirtualMachine), nil
}

// virtualMachineFromSearchIndex gets the virtual machine reference via the
// SearchIndex MO and is the method used to fetch UUIDs on newer versions of
// vSphere.
func virtualMachineFromSearchIndex(ctx context.Context, client *govmomi.Client, uuid string) (object.Reference, error) {
	log.Printf("[DEBUG] Using SearchIndex to look up UUID %q", uuid)
	search := object.NewSearchIndex(client.Client)
	falseBool := false
	result, err := search.FindByUuid(ctx, nil, uuid, true, &falseBool)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, fmt.Errorf("virtual machine with UUID %q not found", uuid)
	}

	return result, nil
}

// virtualMachineFromContainerView is a compatability method that is
// used when the version of vSphere is too old to support using SearchIndex's
// FindByUuid method correctly. This is mainly to facilitate the ability to use
// FromUUID to find both templates in addition to virtual machines, which
// historically was not supported by FindByUuid.
func virtualMachineFromContainerView(ctx context.Context, client *govmomi.Client, uuid string) (object.Reference, error) {
	log.Printf("[DEBUG] Using ContainerView to look up UUID %q", uuid)
	m := view.NewManager(client.Client)

	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = v.Destroy(ctx); err != nil {
			log.Printf("[DEBUG] virtualMachineFromContainerView: Unexpected error destroying container view: %s", err)
		}
	}()

	var vms, results []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"config.uuid"}, &results)
	if err != nil {
		return nil, err
	}

	for _, result := range results {
		if result.Config == nil {
			continue
		}
		if result.Config.Uuid == uuid {
			vms = append(vms, result)
		}
	}

	switch {
	case len(vms) < 1:
		return nil, fmt.Errorf("virtual machine with UUID %q not found", uuid)
	case len(vms) > 1:
		return nil, fmt.Errorf("multiple virtual machines with UUID %q found", uuid)
	}

	return object.NewReference(client.Client, vms[0].Self), nil
}

func (v vSphereVersion) Older(other vSphereVersion) bool {
	vc := v.major<<16 + v.minor<<8 + v.patch
	vo := other.major<<16 + other.minor<<8 + other.patch
	return vo < vc
}

func parseVersion(client *govmomi.Client) (vSphereVersion, error) {
	version := client.Client.ServiceContent.About.Version
	s := strings.Split(version, ".")
	if len(s) > 3 {
		return vSphereVersion{}, fmt.Errorf("could not parse version string %q", version)
	}
	var err error
	major, err := strconv.Atoi(s[0])
	if err != nil {
		return vSphereVersion{}, fmt.Errorf("could not parse version string %q", version)
	}
	minor, err := strconv.Atoi(s[1])
	if err != nil {
		return vSphereVersion{}, fmt.Errorf("could not parse version string %q", version)
	}
	patch, err := strconv.Atoi(s[2])
	if err != nil {
		return vSphereVersion{}, fmt.Errorf("could not parse version string %q", version)
	}

	return vSphereVersion{major: major, minor: minor, patch: patch}, nil
}
