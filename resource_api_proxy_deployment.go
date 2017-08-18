
package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zambien/go-apigee-edge"
	"log"
	"fmt"
	"github.com/satori/go.uuid"
	"strconv"
	"strings"
)

func resourceApiProxyDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceApiProxyDeploymentCreate,
		Read: 	resourceApiProxyDeploymentRead,
		Update: resourceApiProxyDeploymentUpdate,
		Delete: resourceApiProxyDeploymentDelete,

		Schema: map[string]*schema.Schema{
			"proxy_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"org": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"env": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"revision": {
				Type:     schema.TypeString,
				Required: true,
			},
			"delay": {
				Type:     schema.TypeInt,
				Optional: true,
				Default: 0,
			},
			"override": {
				Type:     schema.TypeBool,
				Optional: true,
				Default: false,
			},
		},
	}
}

func resourceApiProxyDeploymentRead(d *schema.ResourceData, meta interface{}) (e error) {

	log.Print("[DEBUG] resourceApiProxyDeploymentRead START")

	client := meta.(*apigee.EdgeClient)

	found := false

	if deployments, _, err := client.Proxies.GetDeployments(d.Get("proxy_name").(string)); err != nil {

		log.Printf("[ERROR] resourceApiProxyDeploymentRead error getting deployments: %s", e.Error())

	} else {

		log.Print("[DEBUG] resourceApiProxyDeploymentRead deployments call fired")

		// TODO: Maybe look at this for the first refactoring exercise... https://github.com/fatih/structs
		for _, environment := range deployments.Environments {
			log.Printf("[DEBUG] resourceApiProxyDeploymentRead checking deploys in environment: %#v for env: %#v\n", environment.Name, d.Get("env").(string))
			if environment.Name == d.Get("env").(string) {
				for _, revision := range environment.Revision {
					log.Printf("[DEBUG] resourceApiProxyDeploymentRead checking revision in revision: %#v for env: %#v\n", revision.Number.String(), d.Get("revision").(string))
					found = true
				}
			}
		}
	}

	if !found {
		log.Print("[INFO] resourceApiProxyDeploymentRead - no deployment found")
		d.SetId("")
	}
	return nil
}

func resourceApiProxyDeploymentCreate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceApiProxyDeploymentCreate START")

	client := meta.(*apigee.EdgeClient)

	proxy_name :=d.Get("proxy_name").(string)
	env := d.Get("env").(string)
	rev_int, _:= strconv.Atoi(d.Get("revision").(string))
	rev := apigee.Revision(rev_int)
	delay := int(d.Get("delay").(int))
	override := bool(d.Get("override").(bool))

	proxyDep, _, err := client.Proxies.Deploy(proxy_name, env, rev, delay, override)

	if err != nil {
		return fmt.Errorf("error deploying: %s", err.Error())
	}

	d.SetId(uuid.NewV4().String())
	d.Set("revision", proxyDep.Revision.String())

	return resourceApiProxyDeploymentRead(d, meta)
}

func resourceApiProxyDeploymentUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceApiProxyDeploymentUpdate START")

	client := meta.(*apigee.EdgeClient)

	proxy_name :=d.Get("proxy_name").(string)
	env := d.Get("env").(string)
	rev_int, _:= strconv.Atoi(d.Get("revision").(string))
	rev := apigee.Revision(rev_int)
	delay := int(d.Get("delay").(int))
	override := bool(d.Get("override").(bool))

	if d.HasChange("proxy_name") || d.HasChange("env") {

		log.Print("[INFO] resourceApiProxyDeploymentUpdate Change detected which requires undeploy and new deploy.")

		_, _, err := client.Proxies.Undeploy(proxy_name, env, rev)
		if err != nil {
			return fmt.Errorf("error undeploying: %s", err.Error())
		}

		proxyDep, _, err := client.Proxies.Deploy(proxy_name, env, rev, delay, override)

		if err != nil {
			return fmt.Errorf("error deploying: %s", err.Error())
		}

		d.SetId(uuid.NewV4().String())
		d.Set("revision", proxyDep.Revision.String())

	} else if d.HasChange("revision") {

		log.Print("[INFO] resourceApiProxyDeploymentUpdate Change detected which allows in place deploy.")

		//We must set delay and override here if not set.
		if delay == 0 {
			delay = 15 //seconds
		}
		if override == false {
			override = true
		}

		_, _, err := client.Proxies.ReDeploy(proxy_name, env, rev, delay, override)

		if err != nil {
			if strings.Contains(err.Error(), " is already deployed into environment ") {
				return resourceApiProxyDeploymentRead(d, meta)
			}
			return fmt.Errorf("error deploying: %s", err.Error())
		}
	}

	return resourceApiProxyDeploymentRead(d, meta)
}

func resourceApiProxyDeploymentDelete(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceApiProxyDeploymentDelete START")

	client := meta.(*apigee.EdgeClient)

	proxy_name :=d.Get("proxy_name").(string)
	env := d.Get("env").(string)
	rev_int, _:= strconv.Atoi(d.Get("revision").(string))
	rev := apigee.Revision(rev_int)

	_, _, err := client.Proxies.Undeploy(proxy_name, env, rev)
	if err != nil {
		return fmt.Errorf("error undeploying: %s", err.Error())
	}

	return nil
}