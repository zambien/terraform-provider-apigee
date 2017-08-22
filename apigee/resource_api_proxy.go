package apigee

import (
	"strings"

	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/satori/go.uuid"
	"github.com/zambien/go-apigee-edge"
	"log"
	"strconv"
)

func resourceApiProxy() *schema.Resource {
	return &schema.Resource{
		Create: resourceApiProxyCreate,
		Read:   resourceApiProxyRead,
		Update: resourceApiProxyUpdate,
		Delete: resourceApiProxyDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bundle": {
				Type:     schema.TypeString,
				Required: true,
			},
			"bundle_sha": {
				Type:     schema.TypeString,
				Required: true,
			},
			//bundle_sha is not used but is a workaround for: https://github.com/hashicorp/terraform/issues/15857
			"output_sha": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"revision": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceApiProxyCreate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceApiProxyCreate START")

	client := meta.(*apigee.EdgeClient)

	u1 := uuid.NewV4()

	proxyRev, _, e := client.Proxies.Import(d.Get("name").(string), d.Get("bundle").(string))

	if e != nil {
		return fmt.Errorf("error creating api_proxy: %s", e.Error())
	}

	d.SetId(u1.String())
	d.Set("name", d.Get("name").(string))
	d.Set("revision", proxyRev.Revision.String())
	d.Set("output_sha", d.Get("bundle_sha").(string))

	return resourceApiProxyRead(d, meta)
}

func resourceApiProxyRead(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceApiProxyRead START")

	client := meta.(*apigee.EdgeClient)

	u, _, err := client.Proxies.Get(d.Get("name").(string))
	if err != nil {
		d.SetId("")
		if strings.Contains(err.Error(), "404 ") {
			return nil
		}
		return err
	}

	latest_rev := strconv.Itoa(len(u.Revisions))

	d.Set("output_sha", d.Get("bundle_sha").(string))
	d.Set("revision", latest_rev)
	d.Set("name", u.Name)

	return nil
}

func resourceApiProxyUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceApiProxyUpdate START")

	client := meta.(*apigee.EdgeClient)

	if d.HasChange("name") {
		log.Printf("[INFO] resourceApiProxyUpdate name changed to: %#v\n", d.Get("name"))
	}

	if d.HasChange("bundle_sha") {
		log.Printf("[INFO] resourceApiProxyUpdate bundle_sha changed to: %#v\n", d.Get("bundle_sha"))
	}

	proxyRev, _, e := client.Proxies.Import(d.Get("name").(string), d.Get("bundle").(string))
	if e != nil {
		return fmt.Errorf("error creating api_proxy: %s", e.Error())
	}

	d.Set("revision", proxyRev.Revision.String())
	d.Set("output_sha", d.Get("bundle_sha").(string))

	return resourceApiProxyRead(d, meta)
}

func resourceApiProxyDelete(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceApiProxyDelete START")

	client := meta.(*apigee.EdgeClient)

	_, _, err := client.Proxies.Delete(d.Get("name").(string))
	if err != nil {
		return err
	}

	return nil
}
