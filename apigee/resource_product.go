package apigee

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zambien/go-apigee-edge"
)

func resourceProduct() *schema.Resource {
	return &schema.Resource{
		Create: resourceProductCreate,
		Read:   resourceProductRead,
		Update: resourceProductUpdate,
		Delete: resourceProductDelete,
		Importer: &schema.ResourceImporter{
			State: resourceProductImport,
		},

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
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"proxies": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
			"environments": {
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

	u1, _ := uuid.NewV4()
	d.SetId(u1.String())

	ProductData, err := setProductData(d)
	if err != nil {
		log.Printf("[ERROR] resourceProductCreate error in setProductData: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceProductCreate error in setProductData: %s", err.Error())
	}

	_, _, e := client.Products.Create(ProductData)
	if e != nil {
		log.Printf("[ERROR] resourceProductCreate error in product creation: %s", e.Error())
		return fmt.Errorf("[ERROR] resourceProductCreate error in product creation: %s", e.Error())
	}

	return resourceProductRead(d, meta)
}

func resourceProductImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Print("[DEBUG] resourceProductImport START")

	client := meta.(*apigee.EdgeClient)
	productData, _, err := client.Products.Get(d.Id())
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("[DEBUG] resourceProductImport. Error getting product: %v", err)
	}
	d.Set("name", d.Id())
	apiResources := flattenStringList(productData.ApiResources)
	proxies := flattenStringList(productData.Proxies)
	scopes := flattenStringList(productData.Scopes)
	environments := flattenStringList(productData.Environments)
	if productData.DisplayName == "" {
		d.Set("display_name", productData.Name)
	} else {
		d.Set("display_name", productData.DisplayName)
	}
	d.Set("display_name", productData.DisplayName)
	d.Set("description", productData.Description)
	d.Set("approval_type", productData.ApprovalType)
	d.Set("attributes", productData.Attributes)
	d.Set("apiResource", apiResources)
	d.Set("proxies", proxies)
	d.Set("quota", productData.Quota)
	d.Set("quota_interval", productData.QuotaInterval)
	d.Set("quota_time_unit", productData.QuotaTimeUnit)
	d.Set("environments", environments)
	d.Set("scopes", scopes)

	return []*schema.ResourceData{d}, nil
}

func resourceProductRead(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceProductRead START")
	client := meta.(*apigee.EdgeClient)

	ProductData, _, err := client.Products.Get(d.Get("name").(string))
	if err != nil {
		log.Printf("[ERROR] resourceProductRead error getting products: %s", err.Error())
		if strings.Contains(err.Error(), "404 ") {
			log.Printf("[DEBUG] resourceProductRead 404 encountered.  Removing state for product: %#v", d.Get("name").(string))
			d.SetId("")
			return nil
		} else {
			log.Printf("[ERROR] resourceProductRead error error getting products: %s", err.Error())
			return fmt.Errorf("[ERROR] resourceProductRead error getting products: %s", err.Error())
		}
	}

	d.Set("name", ProductData.Name)

	if ProductData.DisplayName == "" {
		d.Set("display_name", ProductData.Name)
	} else {
		d.Set("display_name", ProductData.DisplayName)
	}
	d.Set("display_name", ProductData.DisplayName)
	d.Set("description", ProductData.Description)
	d.Set("approval_type", ProductData.ApprovalType)
	d.Set("attributes", ProductData.Attributes)
	d.Set("quota", ProductData.Quota)
	d.Set("quota_interval", ProductData.QuotaInterval)
	d.Set("quota_time_unit", ProductData.QuotaTimeUnit)

	updateResourceOnSortedArrayChange(d, "apiResource", ProductData.ApiResources)
	updateResourceOnSortedArrayChange(d, "proxies", ProductData.Proxies)
	updateResourceOnSortedArrayChange(d, "environments", ProductData.Environments)
	updateResourceOnSortedArrayChange(d, "scopes", ProductData.Scopes)

	return nil
}

func resourceProductUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceProductUpdate START")

	client := meta.(*apigee.EdgeClient)

	ProductData, err := setProductData(d)
	if err != nil {
		log.Printf("[ERROR] resourceProductUpdate error in setProductData: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceProductUpdate error in setProductData: %s", err.Error())
	}

	_, _, e := client.Products.Update(ProductData)
	if e != nil {
		log.Printf("[ERROR] resourceProductUpdate error in product update: %s", e.Error())
		return fmt.Errorf("[ERROR] resourceProductUpdate error in product update: %s", e.Error())
	}

	return resourceProductRead(d, meta)
}

func resourceProductDelete(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceProductDelete START")

	client := meta.(*apigee.EdgeClient)

	_, err := client.Products.Delete(d.Get("name").(string))
	if err != nil {
		log.Printf("[ERROR] resourceProductDelete error in product delete: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceProductDelete error in product delete: %s", err.Error())
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

	environments := []string{""}
	if d.Get("environments") != nil {
		environments = getStringList("environments", d)
	}

	Product := apigee.Product{
		Name:          d.Get("name").(string),
		DisplayName:   d.Get("display_name").(string),
		ApprovalType:  d.Get("approval_type").(string),
		Attributes:    attributes,
		Description:   d.Get("description").(string),
		ApiResources:  apiResources,
		Proxies:       proxies,
		Quota:         d.Get("quota").(string),
		QuotaInterval: d.Get("quota_interval").(string),
		QuotaTimeUnit: d.Get("quota_time_unit").(string),
		Scopes:        scopes,
		Environments:  environments,
	}

	return Product, nil
}
