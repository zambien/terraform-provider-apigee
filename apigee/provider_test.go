// File : provider_test.go

package apigee

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"apigee": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("APIGEE_USER"); v == "" {
		t.Fatal("APIGEE_USER must be set for acceptance tests")
	}
	if v := os.Getenv("APIGEE_PASSWORD"); v == "" {
		t.Fatal("APIGEE_PASSWORD must be set for acceptance tests")
	}
	if v := os.Getenv("APIGEE_ORG"); v == "" {
		t.Fatal("APIGEE_ORG must be set for acceptance tests")
	}
}
