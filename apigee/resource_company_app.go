package apigee

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zambien/go-apigee-edge"
	"log"
	"strings"
)

func resourceCompanyApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceCompanyAppCreate,
		Read:   resourceCompanyAppRead,
		Update: resourceCompanyAppUpdate,
		Delete: resourceCompanyAppDelete,

		Schema: map[string]*schema.Schema{
			"company_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"api_products": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"scopes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"callback_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"app_id": {
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

func resourceCompanyAppCreate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceCompanyAppCreate START")

	client := meta.(*apigee.EdgeClient)

	u1, _ := uuid.NewV4()
	d.SetId(u1.String())

	CompanyAppData, err := setCompanyAppData(d)
	if err != nil {
		log.Printf("[ERROR] resourceCompanyAppCreate error in setCompanyAppData: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceCompanyAppCreate error in setCompanyAppData: %s", err.Error())
	}

	log.Printf("[DEBUG] resourceCompanyAppCreate sending object: %+v\n", CompanyAppData)

	_, _, e := client.CompanyApps.Create(d.Get("company_name").(string), CompanyAppData)
	if e != nil {
		log.Printf("[ERROR] resourceCompanyAppCreate error in company app creation: %s", e.Error())
		return fmt.Errorf("[ERROR] resourceCompanyAppCreate error in company app creation: %s", e.Error())
	}

	return resourceCompanyAppRead(d, meta)
}

func resourceCompanyAppRead(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceCompanyAppRead START")
	client := meta.(*apigee.EdgeClient)

	CompanyAppData, _, err := client.CompanyApps.Get(d.Get("company_name").(string), d.Get("name").(string))
	if err != nil {
		log.Printf("[ERROR] resourceCompanyAppRead error getting company apps: %s", err.Error())
		if strings.Contains(err.Error(), "404 ") {
			log.Printf("[DEBUG] resourceCompanyAppRead 404 encountered.  Removing state for company app: %#v", d.Get("name").(string))
			d.SetId("")
			return nil
		} else {
			log.Printf("[ERROR] resourceCompanyAppRead error error getting company apps: %s", err.Error())
			return fmt.Errorf("[ERROR] resourceCompanyAppRead error getting company apps: %s", err.Error())
		}
	}

	log.Printf("[DEBUG] resourceCompanyAppRead CompanyAppData: %+v\n", CompanyAppData)

	//Scopes and apiProducts are tricky.  These actually result in an array which will always have
	//one element unless an outside API is called.  Since using terraform we assume you do everything there
	//you might only ever have one credential... we'll see.
	apiProducts := apiProductsListFromCredentials(CompanyAppData.Credentials[0].ApiProducts)
	scopes := flattenStringList(CompanyAppData.Credentials[0].Scopes)

	d.Set("name", CompanyAppData.Name)
	d.Set("api_products", apiProducts)
	d.Set("attributes", CompanyAppData.Attributes)
	d.Set("scopes", scopes)
	d.Set("callback_url", CompanyAppData.CallbackUrl)
	d.Set("app_id", CompanyAppData.AppId)
	d.Set("company_name", CompanyAppData.CompanyName)
	d.Set("status", CompanyAppData.Status)

	return nil
}

func resourceCompanyAppUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceCompanyAppUpdate START")

	client := meta.(*apigee.EdgeClient)

	CompanyAppData, err := setCompanyAppData(d)
	if err != nil {
		log.Printf("[ERROR] resourceCompanyAppUpdate error in setCompanyAppData: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceCompanyAppUpdate error in setCompanyAppData: %s", err.Error())
	}

	_, _, e := client.CompanyApps.Update(d.Get("company_name").(string), CompanyAppData)
	if e != nil {
		log.Printf("[ERROR] resourceCompanyAppUpdate error in company app update: %s", e.Error())
		return fmt.Errorf("[ERROR] resourceCompanyAppUpdate error in company app update: %s", e.Error())
	}

	return resourceCompanyAppRead(d, meta)
}

func resourceCompanyAppDelete(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceCompanyAppDelete START")

	client := meta.(*apigee.EdgeClient)

	_, err := client.CompanyApps.Delete(d.Get("company_name").(string), d.Get("name").(string))
	if err != nil {
		log.Printf("[ERROR] resourceCompanyAppDelete error in company app delete: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceCompanyAppDelete error in company app delete: %s", err.Error())
	}

	return nil
}

func setCompanyAppData(d *schema.ResourceData) (apigee.CompanyApp, error) {

	log.Print("[DEBUG] setCompanyAppData START")

	apiProducts := []string{""}
	if d.Get("api_products") != nil {
		apiProducts = getStringList("api_products", d)
	}

	scopes := []string{""}
	if d.Get("scopes") != nil {
		scopes = getStringList("scopes", d)
	}
	log.Printf("[DEBUG] setCompanyAppData scopes: %+v\n", scopes)

	attributes := []apigee.Attribute{}
	if d.Get("attributes") != nil {
		attributes = attributesFromMap(d.Get("attributes").(map[string]interface{}))
	}

	CompanyApp := apigee.CompanyApp{
		Name:        d.Get("name").(string),
		Attributes:  attributes,
		ApiProducts: apiProducts,
		Scopes:      scopes,
		CallbackUrl: d.Get("callback_url").(string),
	}

	return CompanyApp, nil
}
