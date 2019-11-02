package apigee

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zambien/go-apigee-edge"
	"log"
	"regexp"
	"strings"
	"testing"
)

func TestAccDeveloperApp_Updated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDeveloperAppDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckDeveloperAppConfigRequired,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeveloperAppExists("apigee_developer_app.foo_developer_app", "foo_developer_app_name"),
					resource.TestCheckResourceAttr(
						"apigee_developer_app.foo_developer_app", "name", "foo_developer_app_name"),
					resource.TestCheckResourceAttr(
						"apigee_developer_app.foo_developer_app", "developer_email", "foo_developer_app_test_email@test.com"),
				),
			},
			resource.TestStep{
				Config: testAccCheckDeveloperAppConfigUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeveloperAppExists("apigee_developer_app.foo_developer_app_updated", "foo_developer_app_name_updated"),
					resource.TestCheckResourceAttr(
						"apigee_developer_app.foo_developer_app", "name", "foo_developer_app_name_updated"),
					resource.TestCheckResourceAttr(
						"apigee_developer_app.foo_developer_app", "developer_email", "foo_developer_app_test_email@test.com"),
					resource.TestCheckResourceAttr(
						"apigee_developer_app.foo_developer_app", "api_products.0", "foo_product"),
					resource.TestCheckResourceAttr(
						"apigee_developer_app.foo_developer_app", "api_products.1", "bbb_product"),
					resource.TestCheckResourceAttr(
						"apigee_developer_app.foo_developer_app", "api_products.2", "aaa_product"),
					resource.TestCheckResourceAttr(
						"apigee_developer_app.foo_developer_app", "scopes.0", "READ"),
					resource.TestCheckResourceAttr(
						"apigee_developer_app.foo_developer_app", "callback_url", "https://www.google.com"),
					//match integer
					resource.TestMatchResourceAttr(
						"apigee_developer_app.foo_developer_app", "key_expires_in", regexp.MustCompile("^[-+]?\\d+$")),
				),
			},
		},
	})
}

func testAccCheckDeveloperAppDestroy(s *terraform.State) error {

	client := testAccProvider.Meta().(*apigee.EdgeClient)

	if err := developerAppDestroyHelper(s, client); err != nil {
		return err
	}
	return nil
}

func testAccCheckDeveloperAppExists(n string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*apigee.EdgeClient)
		if err := developerAppExistsHelper(s, client, name); err != nil {
			log.Printf("Error in testAccCheckDeveloperAppExists: %s", err)
			return err
		}
		return nil
	}
}

const testAccCheckDeveloperAppConfigRequired = `
resource "apigee_developer" "foo_developer" {
   email = "foo_developer_app_test_email@test.com"
   first_name = "foo"
   last_name = "test"
   user_name = "footest"
}

resource "apigee_developer_app" "foo_developer_app" {
   name = "foo_developer_app_name"
   developer_email = "${apigee_developer.foo_developer.email}"
}
`

const testAccCheckDeveloperAppConfigUpdated = `
resource "apigee_developer" "foo_developer" {
   email = "foo_developer_app_test_email@test.com"
   first_name = "foo"
   last_name = "test"
   user_name = "footest"
}

resource "apigee_product" "aaa_product" {
   name = "aaa_product"
   approval_type = "auto"
   scopes = ["READ"]
}

resource "apigee_product" "bbb_product" {
   name = "bbb_product"
   approval_type = "auto"
   scopes = ["READ"]
}

resource "apigee_product" "foo_product" {
   name = "foo_product"
   approval_type = "auto"
   scopes = ["READ"]
}


resource "apigee_developer_app" "foo_developer_app" {
   name = "foo_developer_app_name_updated"
   developer_email = "${apigee_developer.foo_developer.email}"
   api_products = [
		"${apigee_product.foo_product.name}",		
		"${apigee_product.bbb_product.name}",		
		"${apigee_product.aaa_product.name}"
   ]
   scopes = ["READ"]
   callback_url = "https://www.google.com"
   key_expires_in = 123121515135
}
`

func developerAppDestroyHelper(s *terraform.State, client *apigee.EdgeClient) error {

	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No developer app ID is set")
		}

		_, _, err := client.DeveloperApps.Get("foo_developer_app_test_email@test.com", "foo_developer_app")

		if err != nil {
			if strings.Contains(err.Error(), "404 ") {
				return nil
			}
			return fmt.Errorf("Received an error retrieving developer app: %+v\n", err)
		}
	}

	return fmt.Errorf("DeveloperApp still exists")
}

func developerAppExistsHelper(s *terraform.State, client *apigee.EdgeClient, name string) error {

	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No developer app ID is set")
		}

		if developerAppData, _, err := client.DeveloperApps.Get("foo_developer_app_test_email@test.com", name); err != nil {
			return fmt.Errorf("Received an error retrieving developer app: %+v\n", err)
		} else {
			log.Printf("Created developer app: %s", developerAppData.Name)
		}

	}
	return nil
}
