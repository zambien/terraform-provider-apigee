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
			"test": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"credentials": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"consumer_key": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"consumer_secret": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"ssl_info": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ssl_enabled": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"client_auth_enabled": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"key_store": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"trust_store": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"key_alias": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"ciphers": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"protocols": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
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
			"consumer_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"consumer_secret": {
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
	//one element unless an outside API is called.
	//Get the most recent scopes from the last credentials set
	scopes := flattenStringList(CompanyAppData.Credentials[len(CompanyAppData.Credentials)-1].Scopes)

	//TBD: credentials complex list.  See comments in resource_developer_app.go
	d.Set("credentials", make([]interface{}, 0))

	//Apigee does not return products in the order you send them
	//Get the most recent api products from the last credentials set
	oldApiProducts := getStringList("api_products", d)
	newApiProducts := apiProductsListFromCredentials(CompanyAppData.Credentials[len(CompanyAppData.Credentials)-1].ApiProducts)

	if !arraySortedEqual(oldApiProducts, newApiProducts) {
		d.Set("api_products", newApiProducts)
	} else {
		d.Set("api_products", oldApiProducts)
	}

	d.Set("test","tester")
	d.Set("name", CompanyAppData.Name)
	d.Set("attributes", CompanyAppData.Attributes)
	d.Set("scopes", scopes)
	d.Set("callback_url", CompanyAppData.CallbackUrl)
	d.Set("app_id", CompanyAppData.AppId)
	d.Set("company_name", CompanyAppData.CompanyName)
	d.Set("status", CompanyAppData.Status)
	d.Set("consumer_key", CompanyAppData.Credentials[len(CompanyAppData.Credentials)-1].ConsumerKey)
    d.Set("consumer_secret", CompanyAppData.Credentials[len(CompanyAppData.Credentials)-1].ConsumerSecret)

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
