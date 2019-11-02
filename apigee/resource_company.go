package apigee

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zambien/go-apigee-edge"
	"log"
	"strings"
)

func resourceCompany() *schema.Resource {
	return &schema.Resource{
		Create: resourceCompanyCreate,
		Read:   resourceCompanyRead,
		Update: resourceCompanyUpdate,
		Delete: resourceCompanyDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCompanyCreate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceCompanyCreate START")

	client := meta.(*apigee.EdgeClient)

	u1, _ := uuid.NewV4()
	d.SetId(u1.String())

	CompanyData, err := setCompanyData(d)
	if err != nil {
		log.Printf("[ERROR] resourceCompanyCreate error in setCompanyData: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceCompanyCreate error in setCompanyData: %s", err.Error())
	}

	_, _, e := client.Companies.Create(CompanyData)
	if e != nil {
		log.Printf("[ERROR] resourceCompanyCreate error in developer creation: %s", e.Error())
		return fmt.Errorf("[ERROR] resourceCompanyCreate error in developer creation: %s", e.Error())
	}

	return resourceCompanyRead(d, meta)
}

func resourceCompanyRead(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceCompanyRead START")
	client := meta.(*apigee.EdgeClient)

	CompanyData, _, err := client.Companies.Get(d.Get("name").(string))
	if err != nil {
		log.Printf("[ERROR] resourceCompanyRead error getting companies: %s", err.Error())
		if strings.Contains(err.Error(), "404 ") {
			log.Printf("[DEBUG] resourceCompanyRead 404 encountered.  Removing state for developer: %#v", d.Get("name").(string))
			d.SetId("")
			return nil
		} else {
			log.Printf("[ERROR] resourceCompanyRead error error getting companies: %s", err.Error())
			return fmt.Errorf("[ERROR] resourceCompanyRead error getting companies: %s", err.Error())
		}
	}

	apps := flattenStringList(CompanyData.Apps)

	if CompanyData.DisplayName == "" {
		d.Set("display_name", CompanyData.Name)
	} else {
		d.Set("display_name", CompanyData.DisplayName)
	}

	d.Set("name", CompanyData.Name)
	d.Set("attributes", CompanyData.Attributes)
	d.Set("apps", apps)
	d.Set("status", CompanyData.Status)

	return nil
}

func resourceCompanyUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceCompanyUpdate START")

	client := meta.(*apigee.EdgeClient)

	CompanyData, err := setCompanyData(d)
	if err != nil {
		log.Printf("[ERROR] resourceCompanyUpdate error in setCompanyData: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceCompanyUpdate error in setCompanyData: %s", err.Error())
	}

	_, _, e := client.Companies.Update(CompanyData)
	if e != nil {
		log.Printf("[ERROR] resourceCompanyUpdate error in developer update: %s", e.Error())
		return fmt.Errorf("[ERROR] resourceCompanyUpdate error in developer update: %s", e.Error())
	}

	return resourceCompanyRead(d, meta)
}

func resourceCompanyDelete(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceCompanyDelete START")

	client := meta.(*apigee.EdgeClient)

	_, err := client.Companies.Delete(d.Get("name").(string))
	if err != nil {
		log.Printf("[ERROR] resourceCompanyDelete error in developer delete: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceCompanyDelete error in developer delete: %s", err.Error())
	}

	return nil
}

func setCompanyData(d *schema.ResourceData) (apigee.Company, error) {

	log.Print("[DEBUG] setCompanyData START")

	if d.Get("display_name") == "" {
		d.Set("display_name", d.Get("name"))
	}

	attributes := []apigee.Attribute{}
	if d.Get("attributes") != nil {
		attributes = attributesFromMap(d.Get("attributes").(map[string]interface{}))
	}

	Company := apigee.Company{
		Name:        d.Get("name").(string),
		DisplayName: d.Get("display_name").(string),
		Attributes:  attributes,

		Status: d.Get("status").(string),
	}

	return Company, nil
}
