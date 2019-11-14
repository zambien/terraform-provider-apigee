package apigee

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zambien/go-apigee-edge"
)

func resourceSharedFlowDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceSharedFlowDeploymentCreate,
		Read:   resourceSharedFlowDeploymentRead,
		Update: resourceSharedFlowDeploymentUpdate,
		Delete: resourceSharedFlowDeploymentDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSharedFlowDeploymentImport,
		},

		Schema: map[string]*schema.Schema{
			"shared_flow_name": {
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

func resourceSharedFlowDeploymentImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Print("[DEBUG] resourceSharedFlowDeploymentImport START")
	client := meta.(*apigee.EdgeClient)

	splits := strings.Split(d.Id(), "_")
	if len(splits) < 2 {
		return []*schema.ResourceData{}, fmt.Errorf("[ERR] Wrong format of resource: %s. Please follow '{name}_{env}_deployment'", d.Id())
	}
	nameOffset := len(splits[len(splits)-1]) + len(splits[len(splits)-2])
	envOffset := len(splits[len(splits)-1])
	name := d.Id()[:(len(d.Id())-nameOffset)-2]
	IDEnv := d.Id()[len(name)+1 : (len(d.Id())-envOffset)-1]
	deployment, _, err := client.SharedFlows.GetDeployments(name)
	if err != nil {
		log.Printf("[DEBUG] resourceSharedFlowDeploymentImport. Error getting deployment: %v", err)
		return nil, nil
	}
	d.Set("org", deployment.Organization)
	d.Set("shared_flow_name", deployment.Name)
	d.Set("env", IDEnv)

	return []*schema.ResourceData{d}, nil
}

func resourceSharedFlowDeploymentRead(d *schema.ResourceData, meta interface{}) (e error) {
	log.Print("[DEBUG] resourceSharedFlowDeploymentRead START")
	log.Printf("[DEBUG] resourceSharedFlowDeploymentRead shared_flow_name: %#v", d.Get("shared_flow_name").(string))

	client := meta.(*apigee.EdgeClient)

	found := false
	matchedRevision := "0"

	if deployments, _, err := client.SharedFlows.GetDeployments(d.Get("shared_flow_name").(string)); err != nil {
		log.Printf("[ERROR] resourceSharedFlowDeploymentRead error getting deployments: %s", err.Error())
		if strings.Contains(err.Error(), "404 ") {
			log.Printf("[DEBUG] resourceSharedFlowDeploymentRead 404 encountered.  Removing state for deployment shared_flow_name: %#v", d.Get("shared_flow_name").(string))
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("[ERROR] resourceSharedFlowDeploymentRead error reading deployments: %s", err.Error())
		}
	} else {
		log.Printf("[DEBUG] resourceSharedFlowDeploymentRead deployments call fired for shared_flow_name: %#v", d.Get("shared_flow_name").(string))
		for _, environment := range deployments.Environments {
			log.Printf("[DEBUG] resourceSharedFlowDeploymentRead checking revisions in deployed environment: %#v for expected environment: %#v\n", environment.Name, d.Get("env").(string))
			if environment.Name == d.Get("env").(string) {
				//We don't break.  Always get the last one if there are multiple deployments.
				for _, revision := range environment.Revision {
					log.Printf("[DEBUG] resourceSharedFlowDeploymentRead checking deployed revision: %#v for expected revision: %#v\n", revision.Number.String(), d.Get("revision").(string))
					if d.Get("revision").(string) != "latest" && d.Get("revision").(string) == revision.Number.String() {
						matchedRevision = revision.Number.String()
						found = true
						break
					} else {
						matchedRevision = revision.Number.String()
					}
					found = true
				}
			}
		}
	}

	if found {
		log.Printf("[DEBUG] resourceSharedFlowDeploymentRead - deployment found. Revision is: %#v", matchedRevision)
		d.Set("revision", matchedRevision)
	} else {
		log.Print("[DEBUG] resourceSharedFlowDeploymentRead - no deployment found")
		d.SetId("")
	}
	return nil
}

func resourceSharedFlowDeploymentCreate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceSharedFlowDeploymentCreate START")

	client := meta.(*apigee.EdgeClient)

	sharedFlowName := d.Get("shared_flow_name").(string)
	env := d.Get("env").(string)
	revInt, _ := strconv.Atoi(d.Get("revision").(string))
	rev := apigee.Revision(revInt)
	delay := int(d.Get("delay").(int))
	override := bool(d.Get("override").(bool))

	if d.Get("revision").(string) == "latest" {
		// deploy latest
		rev, err := getLatestSharedFlowRevision(client, sharedFlowName)
		if err != nil {
			return fmt.Errorf("[ERROR] resourceSharedFlowDeploymentCreate error getting latest revision: %v", err)
		}
		_, _, err = client.SharedFlows.Deploy(sharedFlowName, env, apigee.Revision(rev), delay, override)
		if err != nil {
			return fmt.Errorf("[ERROR] resourceSharedFlowDeploymentCreate error deploying: %v", err)
		}
		log.Printf("[DEBUG] resourceSharedFlowDeploymentCreate Deployed revision %d of %s", rev, sharedFlowName)
		return resourceSharedFlowDeploymentRead(d, meta)
	}

	sharedFlowDep, _, err := client.SharedFlows.Deploy(sharedFlowName, env, rev, delay, override)

	if err != nil {

		if strings.Contains(err.Error(), "conflicts with existing deployment path") {
			//create, fail, update
			log.Printf("[ERROR] resourceSharedFlowDeploymentCreate error deploying: %s", err.Error())
			log.Print("[DEBUG] resourceSharedFlowDeploymentCreate something got out of sync... maybe someone messing around in apigee directly.  Terraform OVERRIDE!!!")
			resourceSharedFlowDeploymentUpdate(d, meta)
		} else {
			log.Printf("[ERROR] resourceSharedFlowDeploymentCreate error deploying: %s", err.Error())
			return fmt.Errorf("[ERROR] resourceSharedFlowDeploymentCreate error deploying: %s", err.Error())
		}
	}

	id, _ := uuid.NewV4()
	d.SetId(id.String())
	d.Set("revision", sharedFlowDep.Revision.String())

	return resourceSharedFlowDeploymentRead(d, meta)
}

func resourceSharedFlowDeploymentUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceSharedFlowDeploymentUpdate START")

	client := meta.(*apigee.EdgeClient)

	sharedFlowName := d.Get("shared_flow_name").(string)
	env := d.Get("env").(string)
	delay := int(d.Get("delay").(int))
	override := bool(d.Get("override").(bool))

	//We must set delay and override here if not set.
	if delay == 0 {
		delay = 15 //seconds
	}
	if override == false {
		override = true
	}

	if d.Get("revision").(string) == "latest" {
		// deploy latest
		rev, err := getLatestSharedFlowRevision(client, sharedFlowName)
		if err != nil {
			return fmt.Errorf("[ERROR] resourceSharedFlowDeploymentUpdate error getting latest revision: %v", err)
		}
		_, _, err = client.SharedFlows.ReDeploy(sharedFlowName, env, apigee.Revision(rev), delay, override)
		if err != nil {
			if strings.Contains(err.Error(), " is already deployed ") {
				return resourceSharedFlowDeploymentRead(d, meta)
			}
			return fmt.Errorf("[ERROR] resourceSharedFlowDeploymentUpdate error deploying: %v", err)
		}
		log.Printf("[DEBUG] resourceSharedFlowDeploymentUpdate Deployed revision %d of %s", rev, sharedFlowName)
		return resourceSharedFlowDeploymentRead(d, meta)
	}

	revInt, _ := strconv.Atoi(d.Get("revision").(string))
	rev := apigee.Revision(revInt)
	_, _, err := client.SharedFlows.ReDeploy(sharedFlowName, env, rev, delay, override)

	if err != nil {
		log.Printf("[ERROR] resourceSharedFlowDeploymentUpdate error redeploying: %s", err.Error())
		if strings.Contains(err.Error(), " is already deployed into environment ") {
			return resourceSharedFlowDeploymentRead(d, meta)
		}
		return fmt.Errorf("[ERROR] resourceSharedFlowDeploymentUpdate error redeploying: %s", err.Error())
	}

	return resourceSharedFlowDeploymentRead(d, meta)
}

func resourceSharedFlowDeploymentDelete(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceSharedFlowDeploymentDelete START")

	client := meta.(*apigee.EdgeClient)

	sharedFlowName := d.Get("shared_flow_name").(string)
	env := d.Get("env").(string)
	revInt, _ := strconv.Atoi(d.Get("revision").(string))
	rev := apigee.Revision(revInt)

	_, _, err := client.SharedFlows.Undeploy(sharedFlowName, env, rev)
	if err != nil {
		log.Printf("[ERROR] resourceSharedFlowDeploymentDelete error undeploying: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceSharedFlowDeploymentDelete error undeploying: %s", err.Error())
	}

	return nil
}

func getLatestSharedFlowRevision(client *apigee.EdgeClient, sharedFlowName string) (int, error) {
	sharedFlow, _, err := client.SharedFlows.Get(sharedFlowName)
	if err != nil {
		return -1, fmt.Errorf("[ERROR] getLatestSharedFlowRevision error reading shared flows: %s", err.Error())
	}
	return len(sharedFlow.Revisions), nil
}
