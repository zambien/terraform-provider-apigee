package apigee

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/hashicorp/terraform/helper/schema"
	//"github.com/mitchellh/mapstructure"
	"github.com/zambien/go-apigee-edge"
	"log"
	"strings"
)

func resourceDeveloperApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeveloperAppCreate,
		Read:   resourceDeveloperAppRead,
		Update: resourceDeveloperAppUpdate,
		Delete: resourceDeveloperAppDelete,

		Schema: map[string]*schema.Schema{
			"developer_email": {
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
			"key_expires_in": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"credentials": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"consumer_key": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"consumer_secret": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"issued_at": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"expires_at": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						/*
						"api_products": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},*/
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
			"developer_id": {
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

func resourceDeveloperAppCreate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceDeveloperAppCreate START")

	client := meta.(*apigee.EdgeClient)

	u1, _ := uuid.NewV4()
	d.SetId(u1.String())

	DeveloperAppData, err := setDeveloperAppData(d)
	if err != nil {
		log.Printf("[ERROR] resourceDeveloperAppCreate error in setDeveloperAppData: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceDeveloperAppCreate error in setDeveloperAppData: %s", err.Error())
	}

	log.Printf("[DEBUG] resourceDeveloperAppCreate sending object: %+v\n", DeveloperAppData)

	_, _, e := client.DeveloperApps.Create(d.Get("developer_email").(string), DeveloperAppData)
	if e != nil {
		log.Printf("[ERROR] resourceDeveloperAppCreate error in developer app creation: %s", e.Error())
		return fmt.Errorf("[ERROR] resourceDeveloperAppCreate error in developer app creation: %s", e.Error())
	}

	return resourceDeveloperAppRead(d, meta)
}

func resourceDeveloperAppRead(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceDeveloperAppRead START")
	client := meta.(*apigee.EdgeClient)

	DeveloperAppData, _, err := client.DeveloperApps.Get(d.Get("developer_email").(string), d.Get("name").(string))
	if err != nil {
		log.Printf("[ERROR] resourceDeveloperAppRead error getting developer apps: %s", err.Error())
		if strings.Contains(err.Error(), "404 ") {
			log.Printf("[DEBUG] resourceDeveloperAppRead 404 encountered.  Removing state for developer app: %#v", d.Get("name").(string))
			d.SetId("")
			return nil
		} else {
			log.Printf("[ERROR] resourceDeveloperAppRead error error getting developer apps: %s", err.Error())
			return fmt.Errorf("[ERROR] resourceDeveloperAppRead error getting developer apps: %s", err.Error())
		}
	}

	log.Printf("[DEBUG] resourceDeveloperAppRead DeveloperAppData: %+v\n", DeveloperAppData)

	//Scopes and apiProducts are tricky.  These actually result in an array which will always have
	//one element unless an outside API is called.
	//Get the most recent scopes from the last credentials set
	scopes := flattenStringList(DeveloperAppData.Credentials[len(DeveloperAppData.Credentials)-1].Scopes)

	//Apigee does not return products in the order you send them
	//Get the most recent api products from the last credentials set
	oldApiProducts := getStringList("api_products", d)
	newApiProducts := apiProductsListFromCredentials(DeveloperAppData.Credentials[len(DeveloperAppData.Credentials)-1].ApiProducts)

	if !arraySortedEqual(oldApiProducts, newApiProducts) {
		d.Set("api_products", newApiProducts)
	} else {
		d.Set("api_products", oldApiProducts)
	}


	d.Set("name", DeveloperAppData.Name)
	d.Set("attributes", DeveloperAppData.Attributes)
	d.Set("scopes", scopes)
	d.Set("callback_url", DeveloperAppData.CallbackUrl)
	d.Set("app_id", DeveloperAppData.AppId)
	d.Set("developer_id", DeveloperAppData.DeveloperId)
	d.Set("status", DeveloperAppData.Status)


	//For some reason this is not ever set and there are no errors.  I have followed the syntax here:
	// https://stackoverflow.com/questions/54033185/writing-a-terraform-provider-with-nested-map
	//and here: https://learn.hashicorp.com/tutorials/terraform/provider-complex-read
	//to no avail. We may need to update to lastest plugin sdk.
	//For now just set the last consumer key and secret as simple strings.
	/*
	credentials := flattenCredentials(DeveloperAppData.Credentials)
	if cred_err := d.Set("credentials", credentials); err != nil {
		return fmt.Errorf("[ERROR] resourceDeveloperAppRead error setting credentials: %s", cred_err.Error())
	}*/


	d.Set("credentials", make([]interface{}, 0))
	d.Set("consumer_key", DeveloperAppData.Credentials[len(DeveloperAppData.Credentials)-1].ConsumerKey)
	d.Set("consumer_secret", DeveloperAppData.Credentials[len(DeveloperAppData.Credentials)-1].ConsumerSecret)

	return nil
}

func resourceDeveloperAppUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceDeveloperAppUpdate START")

	client := meta.(*apigee.EdgeClient)

	DeveloperAppData, err := setDeveloperAppData(d)
	if err != nil {
		log.Printf("[ERROR] resourceDeveloperAppUpdate error in setDeveloperAppData: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceDeveloperAppUpdate error in setDeveloperAppData: %s", err.Error())
	}

	_, _, e := client.DeveloperApps.Update(d.Get("developer_email").(string), DeveloperAppData)
	if e != nil {
		log.Printf("[ERROR] resourceDeveloperAppUpdate error in developer app update: %s", e.Error())
		return fmt.Errorf("[ERROR] resourceDeveloperAppUpdate error in developer app update: %s", e.Error())
	}

	return resourceDeveloperAppRead(d, meta)
}

func resourceDeveloperAppDelete(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceDeveloperAppDelete START")

	client := meta.(*apigee.EdgeClient)

	_, err := client.DeveloperApps.Delete(d.Get("developer_email").(string), d.Get("name").(string))
	if err != nil {
		log.Printf("[ERROR] resourceDeveloperAppDelete error in developer app delete: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceDeveloperAppDelete error in developer app delete: %s", err.Error())
	}

	return nil
}

func setDeveloperAppData(d *schema.ResourceData) (apigee.DeveloperApp, error) {

	log.Print("[DEBUG] setDeveloperAppData START")

	apiProducts := []string{""}
	if d.Get("api_products") != nil {
		apiProducts = getStringList("api_products", d)
	}

	scopes := []string{""}
	if d.Get("scopes") != nil {
		scopes = getStringList("scopes", d)
	}
	log.Printf("[DEBUG] setDeveloperAppData scopes: %+v\n", scopes)

	attributes := []apigee.Attribute{}
	if d.Get("attributes") != nil {
		attributes = attributesFromMap(d.Get("attributes").(map[string]interface{}))
	}

	DeveloperApp := apigee.DeveloperApp{
		Name:         d.Get("name").(string),
		Attributes:   attributes,
		ApiProducts:  apiProducts,
		KeyExpiresIn: d.Get("key_expires_in").(int),
		Scopes:       scopes,
		CallbackUrl:  d.Get("callback_url").(string),
	}

	return DeveloperApp, nil
}
