// File : provider.go
package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"github.com/zambien/go-apigee-edge"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"baseUri": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
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

		ResourcesMap: map[string]*schema.Resource{
			"apigee_api_proxy":  resourceApiProxy(),
		},

		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {

	auth := apigee.EdgeAuth{
		Username: d.Get("userEmail").(string),
		Password: d.Get("pass").(string),
	}

	/*
	config := Config{
		BaseURI : d.Get("baseUri").(string),
		Auth: authm,
		User: d.Get("userEmail").(string),
		Pass: d.Get("pass").(string),
		Org: d.Get("org").(string),
	}*/

	log.Println("[INFO] Initializing Apigee client")
	log.Printf("[INFO] Logging in with user: %#v\n", auth.Username)

	opts := &apigee.EdgeClientOptions{
		MgmtUrl : d.Get("baseUri").(string),
		Org: d.Get("org").(string),
		Auth: &auth, Debug: false,
	}
	//client, err := config.Client()
	client, err := apigee.NewEdgeClient(opts)
	if err != nil {
		return client, err
	}

	return client, nil
}
