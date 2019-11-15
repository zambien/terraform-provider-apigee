package apigee

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zambien/go-apigee-edge"
)

func TestAccSharedFlowDeployment_Updated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSharedFlowDeploymentDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSharedFlowDeploymentConfigRequired,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSharedFlowDeploymentExists("apigee_shared_flow_deployment.foo_shared_flow_deployment", "foo_shared_flow_terraformed"),
					resource.TestCheckResourceAttr(
						"apigee_shared_flow_deployment.foo_shared_flow_deployment", "shared_flow_name", "foo_shared_flow_terraformed"),
					resource.TestCheckResourceAttr(
						"apigee_shared_flow_deployment.foo_shared_flow_deployment", "org", "zambien-trial"),
					resource.TestCheckResourceAttr(
						"apigee_shared_flow_deployment.foo_shared_flow_deployment", "revision", "1"),
				),
			},

			resource.TestStep{
				Config: testAccCheckSharedFlowDeploymentConfigUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSharedFlowDeploymentExists("apigee_shared_flow_deployment.foo_shared_flow_deployment", "foo_shared_flow_terraformed"),
					resource.TestCheckResourceAttr(
						"apigee_shared_flow_deployment.foo_shared_flow_deployment", "shared_flow_name", "foo_shared_flow_terraformed"),
					resource.TestCheckResourceAttr(
						"apigee_shared_flow_deployment.foo_shared_flow_deployment", "org", "zambien-trial"),
					resource.TestCheckResourceAttr(
						"apigee_shared_flow_deployment.foo_shared_flow_deployment", "revision", "2"),
					resource.TestCheckResourceAttr(
						"apigee_shared_flow_deployment.foo_shared_flow_deployment", "delay", "2"),
					resource.TestCheckResourceAttr(
						"apigee_shared_flow_deployment.foo_shared_flow_deployment", "override", "true"),
				),
			},
		},
	})
}

func testAccCheckSharedFlowDeploymentDestroy(s *terraform.State) error {

	client := testAccProvider.Meta().(*apigee.EdgeClient)

	if err := sharedFlowDeploymentDestroyHelper(s, client); err != nil {
		return err
	}
	return nil
}

func testAccCheckSharedFlowDeploymentExists(n string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*apigee.EdgeClient)
		if err := sharedFlowDeploymentExistsHelper(s, client, name); err != nil {
			log.Printf("Error in testAccCheckSharedFlowDeploymentExists: %s", err)
			return err
		}
		return nil
	}
}

const testAccCheckSharedFlowDeploymentConfigRequired = `
resource "apigee_shared_flow" "foo_shared_flow" {
   name  		= "foo_shared_flow_terraformed"
   bundle       = "test-fixtures/helloworld_shared_flow.zip"
   bundle_sha   = "${filebase64sha256("test-fixtures/helloworld_shared_flow.zip")}"
}

resource "apigee_shared_flow_deployment" "foo_shared_flow_deployment" {
   shared_flow_name   = apigee_shared_flow.foo_shared_flow.name
   org          = "zambien-trial"
   env          = "test"
   revision     = "1"
}
`

const testAccCheckSharedFlowDeploymentConfigUpdated = `
resource "apigee_shared_flow" "foo_shared_flow" {
   name  		= "foo_shared_flow_terraformed"
   bundle       = "test-fixtures/helloworld_shared_flow2.zip"
   bundle_sha   = "${filebase64sha256("test-fixtures/helloworld_shared_flow2.zip")}"
}


resource "apigee_shared_flow_deployment" "foo_shared_flow_deployment" {
   shared_flow_name   = apigee_shared_flow.foo_shared_flow.name
   org          = "zambien-trial"
   env          = "test"
   revision     = "2"
   delay		= "2"
   override 	= true
}
`

func sharedFlowDeploymentDestroyHelper(s *terraform.State, client *apigee.EdgeClient) error {

	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No shared flow deployment ID is set")
		}

		_, _, err := client.SharedFlows.GetDeployments("foo_shared_flow_deployment")

		if err != nil {
			if strings.Contains(err.Error(), "404 ") {
				return nil
			}
			return fmt.Errorf("Received an error retrieving shared flow deployment  %+v\n", err)
		}
	}

	return fmt.Errorf("SharedFlowDeployment still exists")
}

func sharedFlowDeploymentExistsHelper(s *terraform.State, client *apigee.EdgeClient, name string) error {

	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No shared flow deployment ID is set")
		}

		if sharedFlowDeploymentData, _, err := client.SharedFlows.GetDeployments(name); err != nil {
			return fmt.Errorf("Received an error retrieving shared flow deployment  %+v\n", sharedFlowDeploymentData)
		} else {
			log.Printf("Created shared flow deployment name: %s", sharedFlowDeploymentData.Name)
		}

	}
	return nil
}
