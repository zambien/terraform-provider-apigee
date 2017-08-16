
package main

import (
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zambien/go-apigee-edge"
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"bundle": {
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

// TODO: Fix the issue with only zips working for bundles where folders should too
func resourceApiProxyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apigee.EdgeClient)

	/*
	if proxyRev, resp, err := client.Proxies.Import(d.Get("name").(string),""); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return fmt.Errorf("error creating api_proxy: %s", err.Error())
		}
		log.Printf("[INFO] Updating existing Datadog user %q", u.Handle)
	}*/

	log.Printf("[INFO] Logging in with URL: %#v\n", client.BaseURL.Host)
	log.Printf("[INFO] Logging in with user: %#v\n", client.BaseURL.User)
	log.Printf("[INFO] Logging in with: %#v\n", client.UserAgent)

	log.Printf("[INFO] Creating Proxy Name: %#v\n", d.Get("name").(string))
	log.Printf("[INFO] Creating Proxy From Bundle Location: %#v\n", d.Get("bundle").(string))

	proxyRev, resp, e := client.Proxies.Import(d.Get("name").(string), d.Get("bundle").(string))

	if e != nil {
		log.Printf("[ERROR] error creating api_proxy:: %#v\n", resp.StatusCode)
		log.Printf("[ERROR] error creating api_proxy:: %#v\n", resp.Status)
		log.Printf("[ERROR] error creating api_proxy:: %#v\n", resp.Body)
		return fmt.Errorf("error creating api_proxy: %s", e.Error())
	}

	log.Printf("status: %d\n", resp.StatusCode)
	log.Printf("status: %s\n", resp.Status)
	defer resp.Body.Close()
	log.Printf("proxyRev: %#v\n", proxyRev)

	/*
	var u datadog.User
	u.SetDisabled(d.Get("disabled").(bool))
	u.SetEmail(d.Get("email").(string))
	u.SetHandle(d.Get("handle").(string))
	u.SetIsAdmin(d.Get("is_admin").(bool))
	u.SetName(d.Get("name").(string))
	u.SetRole(d.Get("role").(string))

	// Datadog does not actually delete users, so CreateUser might return a 409.
	// We ignore that case and proceed, likely re-enabling the user.
	if _, err := client.CreateUser(u.Handle, u.Name); err != nil {
		if !strings.Contains(err.Error(), "API error 409 Conflict") {
			return fmt.Errorf("error creating user: %s", err.Error())
		}
		log.Printf("[INFO] Updating existing Datadog user %q", u.Handle)
	}

	if err := client.UpdateUser(u); err != nil {
		return fmt.Errorf("error creating user: %s", err.Error())
	}*/

	d.SetId(proxyRev.Name)

	return resourceApiProxyRead(d, meta)
}


func resourceApiProxyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apigee.EdgeClient)

	u, _, err := client.Proxies.Get(d.Get("name").(string))
	if err != nil {
		return err
	}

	//TODO add more than name
	d.Set("name", u.Name)

	return nil
}

//TODO: actually make it update
func resourceApiProxyUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apigee.EdgeClient)

	u, _, err := client.Proxies.Get(d.Get("name").(string))
	if err != nil {
		return err
	}

	//TODO add more than name
	d.Set("name", u.Name)

	return nil
}

//TODO: actually make it delete
func resourceApiProxyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apigee.EdgeClient)

	_, _, err := client.Proxies.Delete(d.Get("name").(string))
	if err != nil {
		return err
	}

	return nil
}