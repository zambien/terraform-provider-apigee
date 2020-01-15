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

func TestAccProxyDeployment_Updated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProxyDeploymentDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckProxyDeploymentConfigRequired,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyDeploymentExists("apigee_api_proxy_deployment.foo_api_proxy_deployment", "foo_proxy_terraformed"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy_deployment.foo_api_proxy_deployment", "proxy_name", "foo_proxy_terraformed"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy_deployment.foo_api_proxy_deployment", "revision", "1"),
				),
			},

			resource.TestStep{
				Config: testAccCheckProxyDeploymentConfigUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyDeploymentExists("apigee_api_proxy_deployment.foo_api_proxy_deployment", "foo_proxy_terraformed"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy_deployment.foo_api_proxy_deployment", "proxy_name", "foo_proxy_terraformed"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy_deployment.foo_api_proxy_deployment", "revision", "2"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy_deployment.foo_api_proxy_deployment", "delay", "2"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy_deployment.foo_api_proxy_deployment", "override", "true"),
				),
			},
		},
	})
}

func testAccCheckProxyDeploymentDestroy(s *terraform.State) error {

	client := testAccProvider.Meta().(*apigee.EdgeClient)

	if err := proxyDeploymentDestroyHelper(s, client); err != nil {
		return err
	}
	return nil
}

func testAccCheckProxyDeploymentExists(n string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*apigee.EdgeClient)
		if err := proxyDeploymentExistsHelper(s, client, name); err != nil {
			log.Printf("Error in testAccCheckProxyDeploymentExists: %s", err)
			return err
		}
		return nil
	}
}

const testAccCheckProxyDeploymentConfigRequired = `
resource "apigee_api_proxy" "foo_api_proxy" {
   name  		= "foo_proxy_terraformed"
   bundle       = "test-fixtures/helloworld_proxy.zip"
   bundle_sha   = "${filebase64sha256("test-fixtures/helloworld_proxy.zip")}"
}

resource "apigee_api_proxy_deployment" "foo_api_proxy_deployment" {
   proxy_name   = "${apigee_api_proxy.foo_api_proxy.name}"
   org          = "zambien-trial"
   env          = "test"
   revision     = "1"
}
`

const testAccCheckProxyDeploymentConfigUpdated = `
resource "apigee_api_proxy" "foo_api_proxy" {
   name  		= "foo_proxy_terraformed"
   bundle       = "test-fixtures/helloworld_proxy2.zip"
   bundle_sha   = "${filebase64sha256("test-fixtures/helloworld_proxy2.zip")}"
}


resource "apigee_api_proxy_deployment" "foo_api_proxy_deployment" {
   proxy_name   = "${apigee_api_proxy.foo_api_proxy.name}"
   env          = "test"
   revision     = "2"
   delay		= "2"
   override 	= true
}
`

func proxyDeploymentDestroyHelper(s *terraform.State, client *apigee.EdgeClient) error {

	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No proxy deployment ID is set")
		}

		_, _, err := client.Proxies.GetDeployments("foo_proxy deployment")

		if err != nil {
			if strings.Contains(err.Error(), "404 ") {
				return nil
			}
			return fmt.Errorf("Received an error retrieving proxy deployment  %+v\n", err)
		}
	}

	return fmt.Errorf("ProxyDeployment still exists")
}

func proxyDeploymentExistsHelper(s *terraform.State, client *apigee.EdgeClient, name string) error {

	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No proxy deployment ID is set")
		}

		if proxyDeploymentData, _, err := client.Proxies.GetDeployments(name); err != nil {
			return fmt.Errorf("Received an error retrieving proxy deployment  %+v\n", proxyDeploymentData)
		} else {
			log.Printf("Created proxy deployment name: %s", proxyDeploymentData.Name)
		}

	}
	return nil
}
