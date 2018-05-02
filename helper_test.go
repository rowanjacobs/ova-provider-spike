package main_test

import (
	"fmt"

	"github.com/hashicorp/terraform/terraform"
	"github.com/rowanjacobs/ova-provider-spike/internal/helper"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
)

// testGetTemplate is a convenience method to fetch a template by resource name.
func testGetTemplate(s *terraform.State, resourceName string) (*object.VirtualMachine, error) {
	client := testGetClient()
	attributes, err := testGetAttributesForResource(s, fmt.Sprintf("ova_template.%s", resourceName))
	if err != nil {
		return nil, err
	}
	uuid, ok := attributes["uuid"]
	if !ok {
		return nil, fmt.Errorf("resource %q has no UUID", resourceName)
	}
	return helper.FromUUID(client, uuid)
}

func testGetClient() *govmomi.Client {
	return testAccProvider.Meta().(*govmomi.Client)
}

func testGetAttributesForResource(s *terraform.State, addr string) (map[string]string, error) {
	rs, ok := s.RootModule().Resources[addr]
	if !ok {
		return map[string]string{}, fmt.Errorf("%s not found in state", addr)
	}

	return rs.Primary.Attributes, nil
}
