package helper

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/nfc"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

func Import(ctx context.Context,
	ovfPath string,
	// networkMapOverrides map[string]string,
	client *govmomi.Client,
	resourcePool *object.ResourcePool,
	dataStore *object.Datastore,
	dc *object.Datacenter,
	folder *object.Folder,
) ([]string, error) {
	contents, err := ioutil.ReadFile(ovfPath)
	if err != nil {
		return nil, fmt.Errorf("failure reading file: %s", err)
	}

	envelope, err := ovf.Unmarshal(bytes.NewReader(contents))
	if err != nil {
		return nil, fmt.Errorf("failure unmarshalling ovf: %s", err)
	}

	// upsert network mappings
	networks := map[string]string{}
	if envelope.Network != nil {
		for _, net := range envelope.Network.Networks {
			networks[net.Name] = net.Name // this needs a remap
		}
	}
	// for original, new := range networkMapOverrides {
	// 	networks[original] = new // this needs a remap
	// }

	// form real network map with object references
	isp := types.OvfCreateImportSpecParams{NetworkMapping: []types.OvfNetworkMapping{}}
	for src, dst := range networks {
		net, err := Network(client, dc, dst)
		if err != nil {
			return nil, fmt.Errorf("failed finding network", err)
		}
		isp.NetworkMapping = append(isp.NetworkMapping, types.OvfNetworkMapping{
			Name:    src,
			Network: net.Reference(),
		})
	}

	// create an ovf manager, use it to create an import spec out of our CreateImportSpecParams
	manager := ovf.NewManager(client.Client)
	spec, err := manager.CreateImportSpec(ctx, string(contents), resourcePool, dataStore, isp)
	if err != nil {
		return nil, fmt.Errorf("failure creating import spec: %s", err)
	}
	if spec.Error != nil {
		return nil, fmt.Errorf("failure in import spec %+v\n%s\n", isp, spec.Error[0].LocalizedMessage)
	}

	// do a dance to execute the uploads
	lease, err := resourcePool.ImportVApp(ctx, spec.ImportSpec, folder, nil)
	if err != nil {
		return nil, fmt.Errorf("failure importing vapp: %s", err)
	}

	info, err := lease.Wait(ctx, spec.FileItem)
	if err != nil {
		return nil, fmt.Errorf("failure waiting on lease: %s", err)
	}

	updater := lease.StartUpdater(ctx, info)
	defer updater.Done()

	itemURLs := []string{}
	for _, i := range info.Items {
		err = upload(ctx, lease, ovfPath, i)
		if err != nil {
			return nil, fmt.Errorf("failure uploading: %s", err)
		}
		itemURLs = append(itemURLs, i.URL.String())
	}

	return itemURLs, lease.Complete(ctx)
}

// TODO: do we have to do anything special if the file is a tarball? answer: almost certainly yes...
func upload(ctx context.Context, lease *nfc.Lease, ovfPath string, item nfc.FileItem) error {
	file, err := os.Open(filepath.Join(filepath.Dir(ovfPath), item.Path))
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	err = lease.Upload(ctx, item, file, soap.Upload{ContentLength: fileInfo.Size()})
	if err != nil {
		return fmt.Errorf("Lease upload: %s", err)
	}

	return nil
}
