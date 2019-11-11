package apigee

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zambien/go-apigee-edge"
	"log"
	"strings"
)

func resourceDeveloper() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeveloperCreate,
		Read:   resourceDeveloperRead,
		Update: resourceDeveloperUpdate,
		Delete: resourceDeveloperDelete,

		Schema: map[string]*schema.Schema{
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"first_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"last_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"apps": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"developer_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDeveloperCreate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceDeveloperCreate START")

	client := meta.(*apigee.EdgeClient)

	u1, _ := uuid.NewV4()
	d.SetId(u1.String())

	DeveloperData, err := setDeveloperData(d)
	if err != nil {
		log.Printf("[ERROR] resourceDeveloperCreate error in setDeveloperData: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceDeveloperCreate error in setDeveloperData: %s", err.Error())
	}

	_, _, e := client.Developers.Create(DeveloperData)
	if e != nil {
		log.Printf("[ERROR] resourceDeveloperCreate error in developer creation: %s", e.Error())
		return fmt.Errorf("[ERROR] resourceDeveloperCreate error in developer creation: %s", e.Error())
	}

	return resourceDeveloperRead(d, meta)
}

func resourceDeveloperRead(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceDeveloperRead START")
	client := meta.(*apigee.EdgeClient)

	DeveloperData, _, err := client.Developers.Get(d.Get("email").(string))
	if err != nil {
		log.Printf("[ERROR] resourceDeveloperRead error getting developers: %s", err.Error())
		if strings.Contains(err.Error(), "404 ") {
			log.Printf("[DEBUG] resourceDeveloperRead 404 encountered.  Removing state for developer: %#v", d.Get("email").(string))
			d.SetId("")
			return nil
		} else {
			log.Printf("[ERROR] resourceDeveloperRead error error getting developers: %s", err.Error())
			return fmt.Errorf("[ERROR] resourceDeveloperRead error getting developers: %s", err.Error())
		}
	}

	apps := flattenStringList(DeveloperData.Apps)

	d.Set("email", DeveloperData.Email)
	d.Set("first_name", DeveloperData.FirstName)
	d.Set("last_name", DeveloperData.LastName)
	d.Set("user_name", DeveloperData.UserName)
	d.Set("attributes", DeveloperData.Attributes)
	d.Set("apps", apps)
	d.Set("developer_id", DeveloperData.DeveloperId)
	d.Set("status", DeveloperData.Status)

	return nil
}

func resourceDeveloperUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceDeveloperUpdate START")

	client := meta.(*apigee.EdgeClient)

	DeveloperData, err := setDeveloperData(d)
	if err != nil {
		log.Printf("[ERROR] resourceDeveloperUpdate error in setDeveloperData: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceDeveloperUpdate error in setDeveloperData: %s", err.Error())
	}

	_, _, e := client.Developers.Update(DeveloperData)
	if e != nil {
		log.Printf("[ERROR] resourceDeveloperUpdate error in developer update: %s", e.Error())
		return fmt.Errorf("[ERROR] resourceDeveloperUpdate error in developer update: %s", e.Error())
	}

	return resourceDeveloperRead(d, meta)
}

func resourceDeveloperDelete(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceDeveloperDelete START")

	client := meta.(*apigee.EdgeClient)

	_, err := client.Developers.Delete(d.Get("email").(string))
	if err != nil {
		log.Printf("[ERROR] resourceDeveloperDelete error in developer delete: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceDeveloperDelete error in developer delete: %s", err.Error())
	}

	return nil
}

func setDeveloperData(d *schema.ResourceData) (apigee.Developer, error) {

	log.Print("[DEBUG] setDeveloperData START")

	attributes := []apigee.Attribute{}
	if d.Get("attributes") != nil {
		attributes = attributesFromMap(d.Get("attributes").(map[string]interface{}))
	}

	Developer := apigee.Developer{
		Email:      d.Get("email").(string),
		FirstName:  d.Get("first_name").(string),
		LastName:   d.Get("last_name").(string),
		UserName:   d.Get("user_name").(string),
		Attributes: attributes,
	}

	return Developer, nil
}
