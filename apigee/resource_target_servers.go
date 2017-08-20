package apigee

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/satori/go.uuid"
	"github.com/zambien/go-apigee-edge"
	"log"
	"strconv"
	"strings"
)

func resourceTargetServers() *schema.Resource {
	return &schema.Resource{
		Create: resourceTargetServersCreate,
		Read:   resourceTargetServersRead,
		Update: resourceTargetServersUpdate,
		Delete: resourceTargetServersDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"host": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"env": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"port": {
				Type:     schema.TypeString,
				Required: true,
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
						"ignore_validation_errors": &schema.Schema{
							Type:     schema.TypeBool,
							Required: true,
						},
						"protocols": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func resourceTargetServersCreate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceTargetServersCreate START")

	client := meta.(*apigee.EdgeClient)

	u1 := uuid.NewV4()
	d.SetId(u1.String())

	targetServerData, err := setTargetServerData(d)
	if err != nil {
		return fmt.Errorf("resourceTargetServersCreate error in setTargetServerData: %s", err.Error())
	}

	_, _, e := client.TargetServers.Create(targetServerData, d.Get("env").(string))
	if e != nil {
		return e
	}

	return resourceTargetServersRead(d, meta)
}

func resourceTargetServersRead(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceTargetServersRead START")
	client := meta.(*apigee.EdgeClient)

	targetServerData, _, err := client.TargetServers.Get(d.Get("name").(string), d.Get("env").(string))
	if err != nil {
		d.SetId("")
		if strings.Contains(err.Error(), "404 ") {
			return nil
		}
		return err

	}

	d.Set("name", targetServerData.Name)
	d.Set("host", targetServerData.Host)
	d.Set("enabled", targetServerData.Enabled)
	d.Set("port", targetServerData.Port)

	protocols := flattenStringList(targetServerData.SSLInfo.Protocols)
	ciphers := flattenStringList(targetServerData.SSLInfo.Ciphers)

	d.Set("ssl_info.0.ssl_enabled", targetServerData.SSLInfo.SSLEnabled)
	d.Set("ssl_info.0.client_auth_enabled", targetServerData.SSLInfo.ClientAuthEnabled)
	d.Set("ssl_info.0.key_store", targetServerData.SSLInfo.KeyStore)
	d.Set("ssl_info.0.trust_store", targetServerData.SSLInfo.TrustStore)
	d.Set("ssl_info.0.key_alias", targetServerData.SSLInfo.KeyAlias)
	d.Set("ssl_info.0.ciphers", ciphers)
	d.Set("ssl_info.0.ignore_validation_errors", targetServerData.SSLInfo.IgnoreValidationErrors)
	d.Set("ssl_info.0.protocols", protocols)

	return nil
}

func resourceTargetServersUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceTargetServersUpdate START")

	client := meta.(*apigee.EdgeClient)

	targetServerData, err := setTargetServerData(d)
	if err != nil {
		return fmt.Errorf("resourceTargetServersUpdate error in setTargetServerData: %s", err.Error())
	}

	_, _, e := client.TargetServers.Update(targetServerData, d.Get("env").(string))
	if e != nil {
		return e
	}

	return resourceTargetServersRead(d, meta)
}

func resourceTargetServersDelete(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceTargetServersDelete START")
	client := meta.(*apigee.EdgeClient)

	_, err := client.TargetServers.Delete(d.Get("name").(string), d.Get("env").(string))
	if err != nil {
		return err
	}

	return nil
}

func setTargetServerData(d *schema.ResourceData) (apigee.TargetServer, error) {

	log.Print("[DEBUG] setTargetServerData START")

	port_int, _ := strconv.Atoi(d.Get("port").(string))

	ciphers := []string{""}
	if d.Get("ssl_info.0.ciphers") != nil {
		ciphers = getStringList("ssl_info.0.ciphers", d)
	}

	protocols := []string{""}
	if d.Get("ssl_info.0.protocols") != nil {
		protocols = getStringList("ssl_info.0.protocols", d)
	}

	targetServer := apigee.TargetServer{
		Name:    d.Get("name").(string),
		Host:    d.Get("host").(string),
		Enabled: d.Get("enabled").(bool),
		Port:    port_int,
		SSLInfo: apigee.SSLInfo{
			SSLEnabled:        d.Get("ssl_info.0.ssl_enabled").(string),
			ClientAuthEnabled: d.Get("ssl_info.0.client_auth_enabled").(string),
			KeyStore:          d.Get("ssl_info.0.key_store").(string),
			TrustStore:        d.Get("ssl_info.0.trust_store").(string),
			KeyAlias:          d.Get("ssl_info.0.key_alias").(string),
			Ciphers:           ciphers,
			//Ciphers: d.Get("ssl_info.0.ciphers").([]string),
			IgnoreValidationErrors: d.Get("ssl_info.0.ignore_validation_errors").(bool),
			Protocols:              protocols,
		},
	}

	return targetServer, nil
}

func getStringList(listName string, d *schema.ResourceData) []string {

	stringList := []string{}

	if attr, ok := d.GetOk(listName); ok {
		for _, s := range attr.([]interface{}) {
			if s != nil {
				stringList = append(stringList, s.(string))
			}
		}
	}

	return stringList
}

func flattenStringList(list []string) []interface{} {

	vs := make([]interface{}, 0, len(list))

	for _, v := range list {
		vs = append(vs, &v)
	}

	return vs
}
