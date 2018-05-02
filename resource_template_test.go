package main_test

import (
	"errors"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccResourceTemplate_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: testAccResourceVSphereTemplateCheckExists(false),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTemplateConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereTemplateCheckExists(true),
				),
			},
		},
	})
}

func testAccResourceVSphereTemplateCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetTemplate(s, "terraform-test-ovf")
		if err != nil {
			if ok, _ := regexp.MatchString("virtual machine with UUID \"[-a-f0-9]+\" not found", err.Error()); ok && !expected {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return errors.New("expected VM to be missing")
		}
		return nil
	}
}

const testAccResourceTemplateConfigBasic = `
resource "ova_template" "terraform-test-ovf" {
	name = "some-template"
	path = "some-ovf-path"
}
`
