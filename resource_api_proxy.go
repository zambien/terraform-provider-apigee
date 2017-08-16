
package main

import (
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zambien/go-apigee-edge"
	"github.com/satori/go.uuid"
	"fmt"
	"log"
)

func resourceApiProxy() *schema.Resource {
	return &schema.Resource{
		Create: resourceApiProxyCreate,
		Exists: resourceApiProxyExists,
		Read: resourceApiProxyRead,
		Update: resourceApiProxyUpdate,
		Delete: resourceApiProxyDelete,

		Schema: map[string]*schema.Schema{
			"name_prefix": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bundle": {
				Type:     schema.TypeString,
				Required: true,
			},
			"bundle_sha": {
				Type:     schema.TypeString,
				Required: true,
			},
			"revision": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceApiProxyExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	client := meta.(*apigee.EdgeClient)

	if _, _, err := client.Proxies.Get(d.Get("name").(string)); err != nil {
		d.SetId("")
		if strings.Contains(err.Error(), "404 ") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func resourceApiProxyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apigee.EdgeClient)

	u1 := uuid.NewV4()
	name := d.Get("name_prefix").(string) + "-" + u1.String()

	proxyRev, _, e := client.Proxies.Import(name, d.Get("bundle").(string))

	if e != nil {
		return fmt.Errorf("error creating api_proxy: %s", e.Error())
	}

	d.SetId(u1.String())
	d.Set("name", name)
	d.Set("revision", proxyRev.Revision.String())

	return resourceApiProxyRead(d, meta)
}


func resourceApiProxyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apigee.EdgeClient)

	u, _, err := client.Proxies.Get(d.Get("name").(string))
	if err != nil {
		return err
	}

	d.Set("name", u.Name)

	return nil
}

// TODO: Refactor for DRY
func resourceApiProxyUpdate(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*apigee.EdgeClient)

	log.Printf("resourceApiProxyUpdate name_prefix changed: %#v\n", d.HasChange("name_prefix"))
	log.Printf("resourceApiProxyUpdate bundle_sha changed: %#v\n", d.HasChange("bundle_sha"))

	 if d.HasChange("name_prefix") {

		  _, _, err := client.Proxies.Delete(d.Get("name").(string))
		 if err != nil {
			 return err
		 }

		 u1 := uuid.NewV4()
		 name := d.Get("name_prefix").(string) + "-" + u1.String()

		 proxyRev, _, e := client.Proxies.Import(name, d.Get("bundle").(string))
		 if e != nil {
			 return fmt.Errorf("error creating api_proxy: %s", e.Error())
		 }

		 d.SetId(u1.String())
		 d.Set("name", name)
		 d.Set("revision", proxyRev.Revision.String())

	 } else if d.HasChange("bundle_sha") {

		 name := d.Get("name").(string)

		 proxyRev, _, e := client.Proxies.Import(name, d.Get("bundle").(string))
		 if e != nil {
			 return fmt.Errorf("error creating api_proxy: %s", e.Error())
		 }

		 d.Set("revision", proxyRev.Revision.String())
	 }

	return nil
}

func resourceApiProxyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apigee.EdgeClient)

	_, _, err := client.Proxies.Delete(d.Get("name").(string))
	if err != nil {
		return err
	}

	return nil
}