
package main

import (
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zambien/go-apigee-edge"
	"github.com/satori/go.uuid"
	"fmt"
	"log"
	"strconv"
)

func resourceApiProxy() *schema.Resource {
	return &schema.Resource{
		Create: resourceApiProxyCreate,
		Read: 	resourceApiProxyRead,
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

	latest_rev:= strconv.Itoa(len(u.Revisions))

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

	if d.HasChange("name") {

		_, _, err := client.Proxies.Delete(d.Get("name").(string))
		if err != nil {
		 return err
		}

		u1 := uuid.NewV4()

		proxyRev, _, e := client.Proxies.Import(d.Get("name").(string), d.Get("bundle").(string))
		if e != nil {
		 return fmt.Errorf("error creating api_proxy: %s", e.Error())
		}

		d.SetId(u1.String())
		d.Set("name", d.Get("name").(string))
		d.Set("revision", proxyRev.Revision.String())

	} else if d.HasChange("bundle_sha") {

		proxyRev, _, e := client.Proxies.Import(d.Get("name").(string), d.Get("bundle").(string))
		if e != nil {
		 return fmt.Errorf("error creating api_proxy: %s", e.Error())
		}

		d.Set("revision", proxyRev.Revision.String())

	}

	return nil
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