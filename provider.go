// File : provider.go
package terraform_provider_apigee

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"log"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"baseUri": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
				DefaultFunc: schema.EnvDefaultFunc("APIGEE_BASE_URI", nil),
				Description: "Apigee Edge Base URI",
			},
			"userEmail": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
				DefaultFunc: schema.EnvDefaultFunc("APIGEE_USER_EMAIL", nil),
				Description: "Apigee Email Address",
			},
			"pass": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
				DefaultFunc: schema.EnvDefaultFunc("APIGEE_PASS", nil),
				Description: "Apigee Email Password",
			},
			"org": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
				DefaultFunc: schema.EnvDefaultFunc("APIGEE_ORG", nil),
				Description: "Apigee Organization",
			},
		},
		ResourcesMap: map[string]*schema.Resource{ },
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {

	config := Config{
		BaseURI : d.Get("baseUri").(string),
		User: d.Get("user").(string),
		Pass: d.Get("pass").(string),
		Org: d.Get("org").(string),
	}

	log.Println("[INFO] Initializing Apigee client")

	client, err := config.Client()
	if err != nil {
		return client, err
	}

	return client, nil
}
