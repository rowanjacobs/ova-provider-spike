package main

import "github.com/hashicorp/terraform/helper/schema"

func resourceOvf() *schema.Resource {
	return &schema.Resource{
		Create: resourceOvfCreate,
		Read:   resourceOvfRead,
		Update: resourceOvfUpdate,
		Delete: resourceOvfDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"path": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceOvfCreate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)
	d.SetId(name)
	return nil
}

func resourceOvfRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceOvfUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceOvfDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
