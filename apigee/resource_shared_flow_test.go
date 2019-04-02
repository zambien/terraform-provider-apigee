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

func TestAccSharedFlow_Updated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSharedFlowDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSharedFlowConfigRequired,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSharedFlowExists("apigee_shared_flow.helloFlow", "helloFlow"),
					resource.TestCheckResourceAttr(
						"apigee_shared_flow.helloFlow", "name", "helloFlow"),
					resource.TestCheckResourceAttr(
						"apigee_shared_flow.helloFlow", "bundle", "test-fixtures/helloFlow_sharedFlow.zip"),
					resource.TestCheckResourceAttr(
						"apigee_shared_flow.helloFlow", "revision", "1"),
				),
			},

			resource.TestStep{
				Config: testAccCheckSharedFlowConfigUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSharedFlowExists("apigee_shared_flow.helloFlow", "helloFlow_updated"),
					resource.TestCheckResourceAttr(
						"apigee_shared_flow.helloFlow", "name", "helloFlow_updated"),
					resource.TestCheckResourceAttr(
						"apigee_shared_flow.helloFlow", "bundle", "test-fixtures/helloFlow_sharedFlow.zip"),
					resource.TestCheckResourceAttr(
						"apigee_shared_flow.helloFlow", "revision", "1"),
				),
			},
		},
	})
}

func testAccCheckSharedFlowDestroy(s *terraform.State) error {

	client := testAccProvider.Meta().(*apigee.EdgeClient)

	if err := sharedFlowDestroyHelper(s, client); err != nil {
		return err
	}
	return nil
}

func testAccCheckSharedFlowExists(n string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*apigee.EdgeClient)
		if err := sharedFlowExistsHelper(s, client, name); err != nil {
			log.Printf("Error in testAccCheckSharedFlowExists: %s", err)
			return err
		}
		return nil
	}
}

const testAccCheckSharedFlowConfigRequired = `
resource "apigee_shared_flow" "helloFlow" {
   name  = "helloFlow"
   bundle       = "test-fixtures/helloFlow_sharedFlow.zip"
   bundle_sha   = "${base64sha256(file("test-fixtures/helloFlow_sharedFlow.zip"))}"
}
`

const testAccCheckSharedFlowConfigUpdated = `
resource "apigee_shared_flow" "helloFlow" {
  name  = "helloFlow_updated"
  bundle       = "test-fixtures/helloFlow_sharedFlow.zip"
  bundle_sha   = "${base64sha256(file("test-fixtures/helloFlow_sharedFlow.zip"))}"
}
`

func sharedFlowDestroyHelper(s *terraform.State, client *apigee.EdgeClient) error {

	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No shared flow ID is set")
		}

		_, _, err := client.SharedFlows.Get("foo_flow")

		if err != nil {
			if strings.Contains(err.Error(), "404 ") {
				return nil
			}
			return fmt.Errorf("Received an error retrieving shared flow  %+v\n", err)
		}
	}

	return fmt.Errorf("Shared flow still exists")
}

func sharedFlowExistsHelper(s *terraform.State, client *apigee.EdgeClient, name string) error {
	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No shared flow ID is set")
		}

		if sharedFlowData, _, err := client.SharedFlows.Get(name); err != nil {
			return fmt.Errorf("Received an error retrieving shared flow  %+v\n", sharedFlowData)
		} else {
			log.Printf("Created shared flow name: %s", sharedFlowData.Name)
		}

	}
	return nil
}
