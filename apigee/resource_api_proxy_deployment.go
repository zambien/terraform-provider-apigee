package apigee

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/satori/go.uuid"
	"github.com/zambien/go-apigee-edge"
	"log"
	"strconv"
	"strings"
)

func resourceApiProxyDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceApiProxyDeploymentCreate,
		Read:   resourceApiProxyDeploymentRead,
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
				Default:  0,
			},
			"override": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceApiProxyDeploymentRead(d *schema.ResourceData, meta interface{}) (e error) {

	log.Print("[DEBUG] resourceApiProxyDeploymentRead START")
	log.Printf("[DEBUG] resourceApiProxyDeploymentRead proxy_name: %#v", d.Get("proxy_name").(string))

	client := meta.(*apigee.EdgeClient)

	found := false
	latestRevision := "0"

	if deployments, _, err := client.Proxies.GetDeployments(d.Get("proxy_name").(string)); err != nil {
		log.Printf("[ERROR] resourceApiProxyDeploymentRead error getting deployments: %s", err.Error())
		if strings.Contains(err.Error(), "404 ") {
			log.Printf("[DEBUG] resourceApiProxyDeploymentRead 404 encountered.  Removing state for deployment proxy_name: %#v", d.Get("proxy_name").(string))
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("[ERROR] resourceApiProxyDeploymentRead error reading deployments: %s", err.Error())
		}
	} else {

		log.Printf("[DEBUG] resourceApiProxyDeploymentRead deployments call fired for proxy_name: %#v", d.Get("proxy_name").(string))

		for _, environment := range deployments.Environments {
			log.Printf("[DEBUG] resourceApiProxyDeploymentRead checking revisions in deployed environment: %#v for expected environment: %#v\n", environment.Name, d.Get("env").(string))
			if environment.Name == d.Get("env").(string) {
				//We don't break.  Always get the last one if there are multiple deployments.
				for _, revision := range environment.Revision {
					log.Printf("[DEBUG] resourceApiProxyDeploymentRead checking deployed revision: %#v for expected revision: %#v\n", revision.Number.String(), d.Get("revision").(string))
					latestRevision = revision.Number.String()
					found = true
				}
			}
		}
	}

	if found {
		log.Printf("[DEBUG] resourceApiProxyDeploymentRead - deployment found. Revision is: %#v", latestRevision)
		d.Set("revision", latestRevision)
	} else {
		log.Print("[DEBUG] resourceApiProxyDeploymentRead - no deployment found")
		d.SetId("")
	}
	return nil
}

func resourceApiProxyDeploymentCreate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceApiProxyDeploymentCreate START")

	client := meta.(*apigee.EdgeClient)

	proxy_name := d.Get("proxy_name").(string)
	env := d.Get("env").(string)
	rev_int, _ := strconv.Atoi(d.Get("revision").(string))
	rev := apigee.Revision(rev_int)
	delay := int(d.Get("delay").(int))
	override := bool(d.Get("override").(bool))

	proxyDep, _, err := client.Proxies.Deploy(proxy_name, env, rev, delay, override)

	if err != nil {

		if strings.Contains(err.Error(), "conflicts with existing deployment path") {
			//create, fail, update
			log.Printf("[ERROR] resourceApiProxyDeploymentCreate error deploying: %s", err.Error())
			log.Print("[DEBUG] resourceApiProxyDeploymentCreate something got out of sync... maybe someone messing around in apigee directly.  Terraform OVERRIDE!!!")
			resourceApiProxyDeploymentUpdate(d, meta)
		} else {
			log.Printf("[ERROR] resourceApiProxyDeploymentCreate error deploying: %s", err.Error())
			return fmt.Errorf("[ERROR] resourceApiProxyDeploymentCreate error deploying: %s", err.Error())
		}
	}

	d.SetId(uuid.NewV4().String())
	d.Set("revision", proxyDep.Revision.String())

	return resourceApiProxyDeploymentRead(d, meta)
}

func resourceApiProxyDeploymentUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceApiProxyDeploymentUpdate START")

	client := meta.(*apigee.EdgeClient)

	proxy_name := d.Get("proxy_name").(string)
	env := d.Get("env").(string)
	rev_int, _ := strconv.Atoi(d.Get("revision").(string))
	rev := apigee.Revision(rev_int)
	delay := int(d.Get("delay").(int))
	override := bool(d.Get("override").(bool))

	//We must set delay and override here if not set.
	if delay == 0 {
		delay = 15 //seconds
	}
	if override == false {
		override = true
	}

	_, _, err := client.Proxies.ReDeploy(proxy_name, env, rev, delay, override)

	if err != nil {
		log.Printf("[ERROR] resourceApiProxyDeploymentUpdate error redeploying: %s", err.Error())
		if strings.Contains(err.Error(), " is already deployed into environment ") {
			return resourceApiProxyDeploymentRead(d, meta)
		}
		return fmt.Errorf("[ERROR] resourceApiProxyDeploymentUpdate error redeploying: %s", err.Error())
	}

	return resourceApiProxyDeploymentRead(d, meta)
}

func resourceApiProxyDeploymentDelete(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceApiProxyDeploymentDelete START")

	client := meta.(*apigee.EdgeClient)

	proxy_name := d.Get("proxy_name").(string)
	env := d.Get("env").(string)
	rev_int, _ := strconv.Atoi(d.Get("revision").(string))
	rev := apigee.Revision(rev_int)

	_, _, err := client.Proxies.Undeploy(proxy_name, env, rev)
	if err != nil {
		log.Printf("[ERROR] resourceApiProxyDeploymentDelete error undeploying: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceApiProxyDeploymentDelete error undeploying: %s", err.Error())
	}

	return nil
}
