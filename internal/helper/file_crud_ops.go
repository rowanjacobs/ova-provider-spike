package helper

import (
	"context"
	"errors"
	"io/ioutil"
	"os"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/nfc"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

func Import(ctx context.Context, path string, client *govmomi.Client, resourcePool *object.ResourcePool, dataStore *object.Datastore, folder *object.Folder) ([]string, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	manager := ovf.NewManager(client.Client)
	spec, err := manager.CreateImportSpec(ctx, string(contents), resourcePool, dataStore, types.OvfCreateImportSpecParams{})
	if err != nil {
		return nil, err
	}
	if spec.Error != nil {
		return nil, errors.New(spec.Error[0].LocalizedMessage)
	}

	lease, err := resourcePool.ImportVApp(ctx, spec.ImportSpec, folder, nil)
	if err != nil {
		return nil, err
	}

	info, err := lease.Wait(ctx, spec.FileItem)
	if err != nil {
		return nil, err
	}

	updater := lease.StartUpdater(ctx, info)
	defer updater.Done()

	itemURLs := []string{}
	for _, i := range info.Items {
		err = upload(ctx, lease, i)
		if err != nil {
			return nil, err
		}
		itemURLs = append(itemURLs, i.URL.String())
	}

	return itemURLs, lease.Complete(ctx)
}

// TODO: do we have to do anything special if the file is a tarball?
func upload(ctx context.Context, lease *nfc.Lease, item nfc.FileItem) error {
	file, err := os.Open(item.Path)
	defer file.Close()
	if err != nil {
		return err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	return lease.Upload(ctx, item, file, soap.Upload{ContentLength: fileInfo.Size()})
}
