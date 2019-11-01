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

func TestAccDeveloper_Updated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDeveloperDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckDeveloperConfigRequired,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeveloperExists("apigee_developer.foo_developer", "foo_developer_test_email@test.com"),
					resource.TestCheckResourceAttr(
						"apigee_developer.foo_developer", "email", "foo_developer_test_email@test.com"),
					resource.TestCheckResourceAttr(
						"apigee_developer.foo_developer", "first_name", "foo"),
					resource.TestCheckResourceAttr(
						"apigee_developer.foo_developer", "last_name", "test"),
					resource.TestCheckResourceAttr(
						"apigee_developer.foo_developer", "user_name", "footest"),
				),
			},

			resource.TestStep{
				Config: testAccCheckDeveloperConfigUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeveloperExists("apigee_developer.foo_developer", "foo_developer_test_email_updated@test.com"),
					resource.TestCheckResourceAttr(
						"apigee_developer.foo_developer", "email", "foo_developer_test_email_updated@test.com"),
					resource.TestCheckResourceAttr(
						"apigee_developer.foo_developer", "first_name", "foo-updated"),
					resource.TestCheckResourceAttr(
						"apigee_developer.foo_developer", "last_name", "test-updated"),
					resource.TestCheckResourceAttr(
						"apigee_developer.foo_developer", "user_name", "footest-updated"),
					resource.TestCheckResourceAttr(
						"apigee_developer.foo_developer", "attributes.DisplayName", "my-awesome-app-updated"),
					resource.TestCheckResourceAttr(
						"apigee_developer.foo_developer", "attributes.Notes", "notes_for_developer_app_updated"),
					resource.TestCheckResourceAttr(
						"apigee_developer.foo_developer", "attributes.custom_attribute_name", "custom_attribute_value_updated"),
				),
			},
		},
	})
}

func testAccCheckDeveloperDestroy(s *terraform.State) error {

	client := testAccProvider.Meta().(*apigee.EdgeClient)

	if err := developerDestroyHelper(s, client); err != nil {
		return err
	}
	return nil
}

func testAccCheckDeveloperExists(n string, email string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*apigee.EdgeClient)
		if err := developerExistsHelper(s, client, email); err != nil {
			log.Printf("Error in testAccCheckDeveloperExists: %s", err)
			return err
		}
		return nil
	}
}

const testAccCheckDeveloperConfigRequired = `
resource "apigee_developer" "foo_developer" {
   email = "foo_developer_test_email@test.com"
   first_name = "foo"
   last_name = "test"
   user_name = "footest"
}
`

const testAccCheckDeveloperConfigUpdated = `
resource "apigee_developer" "foo_developer" {
   email = "foo_developer_test_email_updated@test.com"
   first_name = "foo-updated"
   last_name = "test-updated"
   user_name = "footest-updated"
   attributes = {
      DisplayName = "my-awesome-app-updated"
      Notes = "notes_for_developer_app_updated"
	  custom_attribute_name = "custom_attribute_value_updated"
   }
}
`

func developerDestroyHelper(s *terraform.State, client *apigee.EdgeClient) error {

	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No developer ID is set")
		}

		_, _, err := client.Developers.Get("foo_developer")

		if err != nil {
			if strings.Contains(err.Error(), "404 ") {
				return nil
			}
			return fmt.Errorf("Received an error retrieving developer  %+v\n", err)
		}
	}

	return fmt.Errorf("Developer still exists")
}

func developerExistsHelper(s *terraform.State, client *apigee.EdgeClient, email string) error {

	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No developer ID is set")
		}

		if developerData, _, err := client.Developers.Get(email); err != nil {
			return fmt.Errorf("Received an error retrieving developer  %+v\n", developerData)
		} else {
			log.Printf("Created developer Email: %s", developerData.Email)
		}

	}
	return nil
}
