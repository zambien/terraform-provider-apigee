package apigee

import (
	"strings"

	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zambien/go-apigee-edge"
)

func resourceSharedFlow() *schema.Resource {
	return &schema.Resource{
		Create: resourceSharedFlowCreate,
		Read:   resourceSharedFlowRead,
		Update: resourceSharedFlowUpdate,
		Delete: resourceSharedFlowDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSharedFlowImport,
		},

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
			//revision_sha is used as a workaround for: https://github.com/hashicorp/terraform/issues/15857
			"revision_sha": {
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

func resourceSharedFlowCreate(d *schema.ResourceData, meta interface{}) error {
	log.Print("[DEBUG] resourceSharedFlowCreate START")

	client := meta.(*apigee.EdgeClient)

	u1, _ := uuid.NewV4()

	sharedFlowRev, _, err := client.SharedFlows.Import(d.Get("name").(string), d.Get("bundle").(string))

	if err != nil {
		log.Printf("[ERROR] resourceSharedFlowCreate error importing shared_flow: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceSharedFlowCreate error importing shared_flow: %s", err.Error())
	}

	d.SetId(u1.String())
	d.Set("name", d.Get("name").(string))
	d.Set("revision", sharedFlowRev.Revision.String())
	d.Set("revision_sha", d.Get("bundle_sha").(string))

	return resourceSharedFlowRead(d, meta)
}

func resourceSharedFlowImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Print("[DEBUG] resourceSharedFlowImport START")

	client := meta.(*apigee.EdgeClient)
	sharedFlow, _, err := client.SharedFlows.Get(d.Id())
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("[DEBUG] resourceSharedFlowImport. Error getting deployment shared flow: %v", err)
	}
	latestRev := strconv.Itoa(len(sharedFlow.Revisions))
	d.Set("revision", latestRev)
	d.Set("name", d.Id())
	return []*schema.ResourceData{d}, nil
}

func resourceSharedFlowRead(d *schema.ResourceData, meta interface{}) error {
	log.Print("[DEBUG] resourceSharedFlowRead START")

	client := meta.(*apigee.EdgeClient)

	u, _, err := client.SharedFlows.Get(d.Get("name").(string))
	if err != nil {
		log.Printf("[ERROR] resourceSharedFlowRead error reading shared flows: %s", err.Error())
		if strings.Contains(err.Error(), "404 ") {
			log.Printf("[DEBUG] resourceSharedFlowRead 404 encountered.  Removing state for shared flow: %#v", d.Get("name").(string))
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("[ERROR] resourceSharedFlowRead error reading shared flows: %s", err.Error())
		}
	}

	latestRev := strconv.Itoa(len(u.Revisions))

	log.Printf("[DEBUG] resourceSharedFlowRead.  revision_sha before: %#v", d.Get("revision_sha").(string))
	d.Set("revision_sha", d.Get("bundle_sha").(string))
	log.Printf("[DEBUG] resourceSharedFlowRead.  revision_sha after: %#v", d.Get("revision_sha").(string))
	d.Set("revision", latestRev)
	d.Set("name", u.Name)

	return nil
}

func resourceSharedFlowUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceSharedFlowUpdate START")

	client := meta.(*apigee.EdgeClient)

	if d.HasChange("name") {
		log.Printf("[INFO] resourceSharedFlowUpdate name changed to: %#v\n", d.Get("name"))
	}

	if d.HasChange("bundle_sha") {
		log.Printf("[INFO] resourceSharedFlowUpdate bundle_sha changed to: %#v\n", d.Get("bundle_sha"))
	}

	sharedFlowRev, _, err := client.SharedFlows.Import(d.Get("name").(string), d.Get("bundle").(string))
	if err != nil {
		log.Printf("[ERROR] resourceSharedFlowUpdate error importing shared flow: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceSharedFlowUpdate error importing shared flow: %s", err.Error())
	}

	d.Set("revision", sharedFlowRev.Revision.String())
	d.Set("revision_sha", d.Get("bundle_sha").(string))

	return resourceSharedFlowRead(d, meta)
}

func resourceSharedFlowDelete(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceSharedFlowDelete START")

	client := meta.(*apigee.EdgeClient)

	//We have to handle retries in a special way here since this is a DELETE.  Note this used to work fine without retries.
	deleted := false
	tries := 0
	for !deleted && tries < 3 {
		_, _, err := client.SharedFlows.Delete(d.Get("name").(string))
		if err != nil {
			//This is a race condition with Apigee APIs.  Wait and try again.
			if strings.Contains(err.Error(), "Undeploy the shared flow and try again") {
				log.Printf("[ERROR] resourceSharedFlowDelete shared flow still exists.  We will wait and try again.")
				time.Sleep(5 * time.Second)
			} else {
				log.Printf("[ERROR] resourceSharedFlowDelete error deleting shared flow: %s", err.Error())
				return fmt.Errorf("[ERROR] resourceSharedFlowDelete error deleting api_proxshared flowy: %s", err.Error())
			}
		}
		deleted = true
		tries += tries
	}

	return nil
}
