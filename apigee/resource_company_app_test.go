package apigee

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zambien/go-apigee-edge"
	"log"
	"strings"
	"testing"
)

func TestAccCompanyApp_Updated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCompanyAppDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckCompanyAppConfigRequired,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompanyAppExists("apigee_company_app.foo_company_app", "foo_company_app_name"),
					resource.TestCheckResourceAttr(
						"apigee_company_app.foo_company_app", "name", "foo_company_app_name"),
					resource.TestCheckResourceAttr(
						"apigee_company_app.foo_company_app", "company_name", "foo_company"),
				),
			},
			resource.TestStep{
				Config: testAccCheckCompanyAppConfigUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompanyAppExists("apigee_company_app.foo_company_app_updated", "foo_company_app_name_updated"),
					resource.TestCheckResourceAttr(
						"apigee_company_app.foo_company_app", "name", "foo_company_app_name_updated"),
					resource.TestCheckResourceAttr(
						"apigee_company_app.foo_company_app", "company_name", "foo_company"),
					resource.TestCheckResourceAttr(
						"apigee_company_app.foo_company_app", "api_products.0", "foo_product"),
					resource.TestCheckResourceAttr(
						"apigee_company_app.foo_company_app", "scopes.0", "READ"),
					resource.TestCheckResourceAttr(
						"apigee_company_app.foo_company_app", "callback_url", "https://www.google.com"),
				),
			},
		},
	})
}

func testAccCheckCompanyAppDestroy(s *terraform.State) error {

	client := testAccProvider.Meta().(*apigee.EdgeClient)

	if err := companyAppDestroyHelper(s, client); err != nil {
		return err
	}
	return nil
}

func testAccCheckCompanyAppExists(n string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*apigee.EdgeClient)
		if err := companyAppExistsHelper(s, client, name); err != nil {
			log.Printf("Error in testAccCheckCompanyAppExists: %s", err)
			return err
		}
		return nil
	}
}

const testAccCheckCompanyAppConfigRequired = `
resource "apigee_company" "foo_company" {
   name = "foo_company"
}

resource "apigee_company_app" "foo_company_app" {
   name = "foo_company_app_name"
   company_name = "${apigee_company.foo_company.name}"
}
`

const testAccCheckCompanyAppConfigUpdated = `
resource "apigee_company" "foo_company" {
   name = "foo_company"
}

resource "apigee_product" "foo_product" {
   name = "foo_product"
   approval_type = "auto"
   scopes = ["READ"]
}

resource "apigee_company_app" "foo_company_app" {
   name = "foo_company_app_name_updated"
   company_name = "${apigee_company.foo_company.name}"
   api_products = ["${apigee_product.foo_product.name}"]
   scopes = ["READ"]
   callback_url = "https://www.google.com"
}
`

func companyAppDestroyHelper(s *terraform.State, client *apigee.EdgeClient) error {

	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No company app ID is set")
		}

		_, _, err := client.CompanyApps.Get("foo_company", "foo_company_app")

		if err != nil {
			if strings.Contains(err.Error(), "404 ") {
				return nil
			}
			return fmt.Errorf("Received an error retrieving company app: %+v\n", err)
		}
	}

	return fmt.Errorf("CompanyApp still exists")
}

func companyAppExistsHelper(s *terraform.State, client *apigee.EdgeClient, name string) error {

	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No company app ID is set")
		}

		if companyAppData, _, err := client.CompanyApps.Get("foo_company", name); err != nil {
			return fmt.Errorf("Received an error retrieving company app: %+v\n", err)
		} else {
			log.Printf("Created company app: %s", companyAppData.Name)
		}

	}
	return nil
}
