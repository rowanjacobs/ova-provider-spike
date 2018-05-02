package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/rowanjacobs/ova-provider-spike/internal/helper"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
)

func resourceTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceTemplateCreate,
		Read:   resourceTemplateRead,
		Update: resourceTemplateUpdate,
		Delete: resourceTemplateDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"path": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"datastore_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the template's datastore. The template configuration is placed here, along with any virtual disks that are created without datastores.",
			},
			"datacenter": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the template's datacenter.",
			},
			"folder": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the folder to locate the template in.",
				// TODO: not sure if we need this
				StateFunc: helper.NormalizePath,
			},
			"resource_pool_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of a resource pool to put the template in.",
			},
		},
		// TODO: datastore, folder, resource pool
	}
}

func resourceTemplateCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*govmomi.Client)

	name := d.Get("name").(string)
	d.SetId(name)

	path := d.Get("path").(string)

	// TODO: find datastore, folder, and resource pool
	poolID := d.Get("resource_pool_id").(string)
	poolObj, err := helper.FromID(client, "ResourcePool", poolID)
	if err != nil {
		return err
	}
	pool := poolObj.(*object.ResourcePool)

	datastoreID := d.Get("datastore_id").(string)
	datastoreObj, err := helper.FromID(client, "Datastore", datastoreID)
	if err != nil {
		return err
	}
	datastore := datastoreObj.(*object.Datastore)

	folder, err := helper.FromAbsolutePath(client, d.Get("folder").(string))
	if err != nil {
		return err
	}

	urls, err := helper.Import(context.Background(), path, client, pool, datastore, folder)

	if err != nil {
		return err
	}

	fmt.Printf("urls: %+v\n\n", urls)

	// vm := object.NewVirtualMachine(m.(*vim25.Client), *moref)

	return resourceTemplateRead(d, m)
}

func resourceTemplateRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceTemplateUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceTemplateDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
