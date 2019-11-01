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

func TestAccCompany_Updated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCompanyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckCompanyConfigRequired,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompanyExists("apigee_company.foo_company", "foo_company"),
					resource.TestCheckResourceAttr(
						"apigee_company.foo_company", "name", "foo_company"),
				),
			},
			resource.TestStep{
				Config: testAccCheckCompanyConfigUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompanyExists("apigee_company.foo_company", "foo_company_updated"),
					resource.TestCheckResourceAttr(
						"apigee_company.foo_company", "name", "foo_company_updated"),
					resource.TestCheckResourceAttr(
						"apigee_company.foo_company", "display_name", "some longer foo description for foo company"),
					resource.TestCheckResourceAttr(
						"apigee_company.foo_company", "attributes.DisplayName", "my-awesome-foo-company"),
				),
			},
		},
	})
}

func testAccCheckCompanyDestroy(s *terraform.State) error {

	client := testAccProvider.Meta().(*apigee.EdgeClient)

	if err := companyDestroyHelper(s, client); err != nil {
		return err
	}
	return nil
}

func testAccCheckCompanyExists(n string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*apigee.EdgeClient)
		if err := companyExistsHelper(s, client, name); err != nil {
			log.Printf("Error in testAccCheckCompanyExists: %s", err)
			return err
		}
		return nil
	}
}

const testAccCheckCompanyConfigRequired = `
resource "apigee_company" "foo_company" {
   name = "foo_company"
}
`

const testAccCheckCompanyConfigUpdated = `
resource "apigee_company" "foo_company" {
   name = "foo_company_updated"
   display_name = "some longer foo description for foo company"
   attributes = {
      DisplayName = "my-awesome-foo-company"
   }
}
`

func companyDestroyHelper(s *terraform.State, client *apigee.EdgeClient) error {

	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No company ID is set")
		}

		_, _, err := client.Companies.Get("foo_company")

		if err != nil {
			if strings.Contains(err.Error(), "404 ") {
				return nil
			}
			return fmt.Errorf("Received an error retrieving company  %+v\n", err)
		}
	}

	return fmt.Errorf("Company still exists")
}

func companyExistsHelper(s *terraform.State, client *apigee.EdgeClient, name string) error {

	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No company ID is set")
		}

		if companyData, _, err := client.Companies.Get(name); err != nil {
			return fmt.Errorf("Received an error retrieving company  %+v\n", companyData)
		} else {
			log.Printf("Created company name: %s", companyData.Name)
		}

	}
	return nil
}
