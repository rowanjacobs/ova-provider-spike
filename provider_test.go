package main_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	main "github.com/rowanjacobs/ova-provider-spike"
)

var testAccProvider *schema.Provider
var testAccProviders map[string]terraform.ResourceProvider

func init() {
	testAccProvider = main.Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"ova": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := main.Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = main.Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("VSPHERE_USER"); v == "" {
		t.Fatal("VSPHERE_USER must be set for acceptance tests")
	}

	if v := os.Getenv("VSPHERE_PASSWORD"); v == "" {
		t.Fatal("VSPHERE_PASSWORD must be set for acceptance tests")
	}

	if v := os.Getenv("VSPHERE_SERVER"); v == "" {
		t.Fatal("VSPHERE_SERVER must be set for acceptance tests")
	}
}

//// testAccProviderMeta returns a instantiated VSphereClient for this provider.
//// It's useful in state migration tests where a provider connection is actually
//// needed, and we don't want to go through the regular provider configure
//// channels (so this function doesn't interfere with the testAccProvider
//// package global and standard acceptance tests).
////
//// Note we lean on environment variables for most of the provider configuration
//// here and this will fail if those are missing. A pre-check is not run.
//func testAccProviderMeta(t *testing.T) (interface{}, error) {
//	t.Helper()
//	d := schema.TestResourceDataRaw(t, testAccProvider.Schema, make(map[string]interface{}))
//	return providerConfigure(d)
//}
