package apigee

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zambien/go-apigee-edge"
	"log"
	"strings"
	"github.com/satori/go.uuid"
	"fmt"
)

func resourceProduct() *schema.Resource {
	return &schema.Resource{
		Create: resourceProductCreate,
		Read:   resourceProductRead,
		Update: resourceProductUpdate,
		Delete: resourceProductDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"approval_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"api_resources": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"proxies": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"quota": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"quota_interval": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"quota_time_unit": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"scopes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

		},
	}
}

func resourceProductCreate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceProductCreate START")


	client := meta.(*apigee.EdgeClient)

	u1 := uuid.NewV4()
	d.SetId(u1.String())

	ProductData, err := setProductData(d)
	if err != nil {
		return fmt.Errorf("resourceProductCreate error in setProductData: %s", err.Error())
	}

	_, _, e := client.Products.Create(ProductData)
	if e != nil {
		return e
	}

	return resourceProductRead(d, meta)
}

func resourceProductRead(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceProductRead START")
	client := meta.(*apigee.EdgeClient)

	ProductData, _, err := client.Products.Get(d.Get("name").(string))
	if err != nil {
		d.SetId("")
		if strings.Contains(err.Error(), "404 ") {
			return nil
		}
		return err

	}

	apiResources := flattenStringList(ProductData.ApiResources)
	proxies := flattenStringList(ProductData.Proxies)
	scopes := flattenStringList(ProductData.Scopes)

	d.Set("name", ProductData.Name)
	if ProductData.DisplayName == "" {
		d.Set("display_name", ProductData.Name)
	} else {
		d.Set("display_name", ProductData.DisplayName)
	}
	d.Set("description", ProductData.Description)
	d.Set("approval_type", ProductData.ApprovalType)
	d.Set("attributes", ProductData.Attributes)
	d.Set("apiResource", apiResources)
	d.Set("proxies", proxies)
	d.Set("quota", ProductData.Quota)
	d.Set("quota_interval", ProductData.QuotaInterval)
	d.Set("quota_time_unit", ProductData.QuotaTimeUnit)
	d.Set("scopes", scopes)

	return nil
}

func resourceProductUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceProductUpdate START")

	client := meta.(*apigee.EdgeClient)

	ProductData, err := setProductData(d)
	if err != nil {
		return fmt.Errorf("resourceProductUpdate error in setProductData: %s", err.Error())
	}

	_, _, e := client.Products.Update(ProductData)
	if e != nil {
		return e
	}

	return resourceProductRead(d, meta)
}

func resourceProductDelete(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceProductDelete START")


	client := meta.(*apigee.EdgeClient)

	_, err := client.Products.Delete(d.Get("name").(string))
	if err != nil {
		return err
	}

	return nil
}

func setProductData(d *schema.ResourceData) (apigee.Product, error) {

	log.Print("[DEBUG] setProductData START")

	if d.Get("display_name") == "" {
		d.Set("display_name", d.Get("name"))
	}

	apiResources := []string{""}
	if d.Get("api_resources") != nil {
		apiResources = getStringList("api_resources", d)
	}

	proxies := []string{""}
	if d.Get("proxies") != nil {
		proxies = getStringList("proxies", d)
	}

	scopes := []string{""}
	if d.Get("scopes") != nil {
		scopes = getStringList("scopes", d)
	}

	attributes := []apigee.Attribute{}
	if d.Get("attributes") != nil {
		attributes = attributesFromMap(d.Get("attributes").(map[string]interface{}))
	}

	Product := apigee.Product{
		Name:    		d.Get("name").(string),
		DisplayName:    d.Get("display_name").(string),
		ApprovalType: 	d.Get("approval_type").(string),
		Attributes:    	attributes,
		Description:	d.Get("description").(string),
		ApiResources:	apiResources,
		Proxies:		proxies,
		Quota:			d.Get("quota").(string),
		QuotaInterval:	d.Get("quota_interval").(string),
		QuotaTimeUnit:	d.Get("quota_time_unit").(string),
		Scopes:			scopes,
	}

	return Product, nil
}