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

func resourceSharedFlowDelete(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceSharedFlowDelete START")

	client := meta.(*apigee.EdgeClient)

	_, _, err := client.SharedFlows.Delete(d.Get("name").(string))
	if err != nil {
		log.Printf("[ERROR] resourceSharedFlowDelete error deleting shard flow: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceSharedFlowDelete error deleting shared flow: %s", err.Error())
	}

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

func resourceSharedFlowCreate(d *schema.ResourceData, meta interface{}) error {
	log.Print("[DEBUG] resourceSharedFlowCreate START")

	client := meta.(*apigee.EdgeClient)

	u1, _ := uuid.NewV4()

	sharedFlowRev, _, err := client.SharedFlows.Import(d.Get("name").(string), d.Get("bundle").(string))

	if err != nil {
		log.Printf("[ERROR] resourceSharedFlowCreate error importing api_shared_flow: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceSharedFlowCreate error importing api_shared_flow: %s", err.Error())
	}

	d.SetId(u1.String())
	d.Set("name", d.Get("name").(string))
	d.Set("revision", sharedFlowRev.Revision.String())
	d.Set("revision_sha", d.Get("bundle_sha").(string))

	return resourceSharedFlowRead(d, meta)
}

func resourceSharedFlowRead(d *schema.ResourceData, meta interface{}) error {
	log.Print("[DEBUG] resourceSharedFlowRead START")

	client := meta.(*apigee.EdgeClient)

	u, _, err := client.Proxies.Get(d.Get("name").(string))
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

	latest_rev := strconv.Itoa(len(u.Revisions))

	log.Printf("[DEBUG] resourceSharedFlowRead.  revision_sha before: %#v", d.Get("revision_sha").(string))
	d.Set("revision_sha", d.Get("bundle_sha").(string))
	log.Printf("[DEBUG] resourceSharedFlowRead.  revision_sha after: %#v", d.Get("revision_sha").(string))
	d.Set("revision", latest_rev)
	d.Set("name", u.Name)

	return nil
}

func resourceSharedFlowImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Print("[DEBUG] resourceSharedFlowImport START")

	client := meta.(*apigee.EdgeClient)
	_, _, err := client.SharedFlows.Get(d.Id())

	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("[DEBUG] resourceSharedFlowImport. Error getting shared flow: %v", err)
	}

	return []*schema.ResourceData{}, nil

}
