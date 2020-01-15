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

func TestAccProxy_Updated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckProxyConfigRequired,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists("apigee_api_proxy.foo_api_proxy", "foo_proxy_terraformed"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy.foo_api_proxy", "name", "foo_proxy_terraformed"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy.foo_api_proxy", "bundle", "test-fixtures/helloworld_proxy.zip"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy.foo_api_proxy", "revision", "1"),
				),
			},

			resource.TestStep{
				Config: testAccCheckProxyConfigUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists("apigee_api_proxy.foo_api_proxy", "foo_proxy_terraformed_updated"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy.foo_api_proxy", "name", "foo_proxy_terraformed_updated"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy.foo_api_proxy", "bundle", "test-fixtures/helloworld_proxy.zip"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy.foo_api_proxy", "revision", "1"),
				),
			},
		},
	})
}

func TestAccProxy_FailToDeleteWhenDeploymentExists(t *testing.T) {
	proxyName := "foo_proxy_terraformed_delete_test"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,

		CheckDestroy: testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckProxyConfigDeleteTest,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists("apigee_api_proxy.foo_api_proxy", proxyName),
					resource.TestCheckResourceAttr("apigee_api_proxy.foo_api_proxy", "name", proxyName),
				),
			},
			{
				Config:      testAccCheckProxyConfigDeleteTest,
				Destroy:     true,
				PreConfig:   deployProxy(t, proxyName),
				ExpectError: regexp.MustCompile("unable to delete ApiProxy"),
			},
			{
				Config:    testAccCheckProxyConfigDeleteTest,
				Destroy:   true,
				PreConfig: undeployProxy(t, proxyName),
			},
		},
	})
}

func testAccCheckProxyDestroy(s *terraform.State) error {

	client := testAccProvider.Meta().(*apigee.EdgeClient)

	if err := proxyDestroyHelper(s, client); err != nil {
		return err
	}
	return nil
}

func testAccCheckProxyExists(n string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*apigee.EdgeClient)
		if err := proxyExistsHelper(s, client, name); err != nil {
			log.Printf("Error in testAccCheckProxyExists: %s", err)
			return err
		}
		return nil
	}
}

const testAccCheckProxyConfigRequired = `
resource "apigee_api_proxy" "foo_api_proxy" {
   name  		= "foo_proxy_terraformed"
   bundle       = "test-fixtures/helloworld_proxy.zip"
   bundle_sha   = "${filebase64sha256("test-fixtures/helloworld_proxy.zip")}"
}
`

const testAccCheckProxyConfigUpdated = `
resource "apigee_api_proxy" "foo_api_proxy" {
   name  		= "foo_proxy_terraformed_updated"
   bundle       = "test-fixtures/helloworld_proxy.zip"
   bundle_sha   = "${filebase64sha256("test-fixtures/helloworld_proxy.zip")}"
}
`

const testAccCheckProxyConfigDeleteTest = `
resource "apigee_api_proxy" "foo_api_proxy" {
   name  		= "foo_proxy_terraformed_delete_test"
   bundle       = "test-fixtures/helloworld_proxy.zip"
   bundle_sha   = "${filebase64sha256("test-fixtures/helloworld_proxy.zip")}"
}
`

func proxyDestroyHelper(s *terraform.State, client *apigee.EdgeClient) error {

	for _, r := range s.RootModule().Resources {
		if r.Type != "apigee_api_proxy" {
			continue
		}
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No proxy ID is set")
		}

		_, _, err := client.Proxies.Get(r.Primary.Attributes["name"])

		if err != nil {
			if strings.Contains(err.Error(), "404 ") {
				return nil
			}
			return fmt.Errorf("Received an error retrieving proxy  %+v\n", err)
		}
	}

	return nil
}

func proxyExistsHelper(s *terraform.State, client *apigee.EdgeClient, name string) error {

	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No proxy ID is set")
		}

		if proxyData, _, err := client.Proxies.Get(name); err != nil {
			return fmt.Errorf("Received an error retrieving proxy  %+v\n", proxyData)
		} else {
			log.Printf("Created proxy name: %s", proxyData.Name)
		}

	}
	return nil
}

func deployProxy(t *testing.T, proxyName string) func() {
	return func() {
		client := testAccProvider.Meta().(*apigee.EdgeClient)
		_, _, err := client.Proxies.Deploy(proxyName, "test", 1, 1, false)
		if err != nil {
			t.Logf("[ERROR] Could not deploy proxy: %s, %s", proxyName, err)
		} else {
			t.Logf("Deployed proxy: %s", proxyName)
		}
	}
}

func undeployProxy(t *testing.T, proxyName string) func() {
	return func() {
		client := testAccProvider.Meta().(*apigee.EdgeClient)
		_, _, err := client.Proxies.Undeploy(proxyName, "test", 1)
		if err != nil {
			t.Logf("[ERROR] Could not undeploy proxy: %s, %s", proxyName, err)
		} else {
			t.Logf("Undeployed proxy: %s", proxyName)
		}
	}
}
