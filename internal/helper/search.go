package helper

import (
	"context"
	"fmt"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

const DefaultAPITimeout = 5 * time.Minute

// adapted from tf vsphere provider internals
func FromID(client *govmomi.Client, resourceType, id string) (object.Reference, error) {
	finder := find.NewFinder(client.Client, false)

	ref := types.ManagedObjectReference{
		Type:  resourceType,
		Value: id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultAPITimeout)
	defer cancel()
	obj, err := finder.ObjectReference(ctx, ref)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func Datacenter(client *govmomi.Client, path string) (*object.Datacenter, error) {
	finder := find.NewFinder(client.Client, false)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultAPITimeout)
	defer cancel()

	obj, err := finder.Datacenter(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("Finding network: %s", err)
	}

	return obj, nil
}

func Network(client *govmomi.Client, dc *object.Datacenter, networkPath string) (object.NetworkReference, error) {
	finder := find.NewFinder(client.Client, false)
	finder.SetDatacenter(dc)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultAPITimeout)
	defer cancel()

	obj, err := finder.Network(ctx, networkPath)
	if err != nil {
		return nil, fmt.Errorf("Finding network: %s", err)
	}

	return obj, nil
}
