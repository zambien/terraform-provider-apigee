
package terraform_provider_apigee

import (
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/DinoChiesa/go-apigee-edge"
)

func resourceApiProxy() *schema.Resource {
	return &schema.Resource{
		Exists: resourceApiProxyExists,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceApiProxyExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	client := meta.(*apigee.EdgeClient)

	if _, _, err := client.Proxies.Get(d.Get("name").(string)); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}