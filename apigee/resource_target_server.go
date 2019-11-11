package apigee

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zambien/go-apigee-edge"
)

func resourceTargetServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceTargetServerCreate,
		Read:   resourceTargetServerRead,
		Update: resourceTargetServerUpdate,
		Delete: resourceTargetServerDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTargetServerImport,
		},

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

func resourceTargetServerCreate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceTargetServerCreate START")

	client := meta.(*apigee.EdgeClient)

	u1, _ := uuid.NewV4()
	d.SetId(u1.String())

	targetServerData, err := setTargetServerData(d)
	if err != nil {
		log.Printf("[ERROR] resourceTargetServerCreate error in setTargetServerData: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceTargetServerCreate error in setTargetServerData: %s", err.Error())
	}

	_, _, e := client.TargetServers.Create(targetServerData, d.Get("env").(string))
	if e != nil {
		log.Printf("[ERROR] resourceTargetServerCreate error in create: %s", e.Error())
		return fmt.Errorf("[ERROR] resourceTargetServerCreate error in create: %s", e.Error())
	}

	return resourceTargetServerRead(d, meta)
}

func resourceTargetServerImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	log.Print("[DEBUG] resourceTargetServerImport START")
	client := meta.(*apigee.EdgeClient)
	splits := strings.Split(d.Id(), "_")
	if len(splits) < 1 {
		return []*schema.ResourceData{}, fmt.Errorf("[ERR] Wrong format of resource: %s. Please follow '{name}_{env}'", d.Id())
	}

	nameOffset := len(splits[len(splits)-1])
	name := d.Id()[:(len(d.Id())-nameOffset)-1]
	IDEnv := splits[len(splits)-1]

	targetServerData, _, err := client.TargetServers.Get(name, IDEnv)
	if err != nil {
		log.Printf("[ERROR] resourceTargetServerImport error getting target servers: %s", err.Error())
		if strings.Contains(err.Error(), "404 ") {
			return []*schema.ResourceData{}, fmt.Errorf("[Error] resourceTargetServerImport 404 encountered.  Removing state for target server: %#v", name)
		}
		return []*schema.ResourceData{}, fmt.Errorf("[ERROR] resourceTargetServerImport error getting target servers: %s", err.Error())
	}

	d.Set("name", targetServerData.Name)
	d.Set("host", targetServerData.Host)
	d.Set("enabled", targetServerData.Enabled)
	d.Set("port", targetServerData.Port)
	d.Set("env", IDEnv)

	if targetServerData.SSLInfo != nil {
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
	}

	return []*schema.ResourceData{d}, nil
}

func resourceTargetServerRead(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceTargetServerRead START")
	client := meta.(*apigee.EdgeClient)

	targetServerData, _, err := client.TargetServers.Get(d.Get("name").(string), d.Get("env").(string))
	if err != nil {
		log.Printf("[ERROR] resourceTargetServerRead error getting target servers: %s", err.Error())
		if strings.Contains(err.Error(), "404 ") {
			log.Printf("[DEBUG] resourceTargetServerRead 404 encountered.  Removing state for target server: %#v", d.Get("name").(string))
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("[ERROR] resourceTargetServerRead error getting target servers: %s", err.Error())
		}
	}

	port_str := strconv.Itoa(targetServerData.Port)

	d.Set("name", targetServerData.Name)
	d.Set("host", targetServerData.Host)
	d.Set("enabled", targetServerData.Enabled)
	d.Set("port", port_str)

	if targetServerData.SSLInfo != nil {
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
	}

	return nil
}

func resourceTargetServerUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceTargetServerUpdate START")

	client := meta.(*apigee.EdgeClient)

	targetServerData, err := setTargetServerData(d)
	if err != nil {
		log.Printf("[ERROR] resourceTargetServerUpdate error in setTargetServerData: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceTargetServerUpdate error in setTargetServerData: %s", err.Error())
	}

	_, _, e := client.TargetServers.Update(targetServerData, d.Get("env").(string))
	if e != nil {
		log.Printf("[ERROR] resourceTargetServerUpdate error in update: %s", e.Error())
		return fmt.Errorf("[ERROR] resourceTargetServerUpdate error in update: %s", e.Error())
	}

	return resourceTargetServerRead(d, meta)
}

func resourceTargetServerDelete(d *schema.ResourceData, meta interface{}) error {

	log.Print("[DEBUG] resourceTargetServerDelete START")
	client := meta.(*apigee.EdgeClient)

	_, err := client.TargetServers.Delete(d.Get("name").(string), d.Get("env").(string))
	if err != nil {
		log.Printf("[ERROR] resourceTargetServerDelete error in delete: %s", err.Error())
		return fmt.Errorf("[ERROR] resourceTargetServerDelete error in delete: %s", err.Error())
	}

	return nil
}

func setTargetServerData(d *schema.ResourceData) (apigee.TargetServer, error) {

	log.Print("[DEBUG] setTargetServerData START")

	port_int, _ := strconv.Atoi(d.Get("port").(string))

	var ssl_info *apigee.SSLInfo
	if d.Get("ssl_info") != nil {
		ciphers := []string{""}
		if d.Get("ssl_info.0.ciphers") != nil {
			ciphers = getStringList("ssl_info.0.ciphers", d)
		}

		protocols := []string{""}
		if d.Get("ssl_info.0.protocols") != nil {
			protocols = getStringList("ssl_info.0.protocols", d)
		}

		ssl_info = &apigee.SSLInfo{
			SSLEnabled:        d.Get("ssl_info.0.ssl_enabled").(string),
			ClientAuthEnabled: d.Get("ssl_info.0.client_auth_enabled").(string),
			KeyStore:          d.Get("ssl_info.0.key_store").(string),
			TrustStore:        d.Get("ssl_info.0.trust_store").(string),
			KeyAlias:          d.Get("ssl_info.0.key_alias").(string),
			Ciphers:           ciphers,
			//Ciphers: d.Get("ssl_info.0.ciphers").([]string),
			IgnoreValidationErrors: d.Get("ssl_info.0.ignore_validation_errors").(bool),
			Protocols:              protocols,
		}
	}

	targetServer := apigee.TargetServer{
		Name:    d.Get("name").(string),
		Host:    d.Get("host").(string),
		Enabled: d.Get("enabled").(bool),
		Port:    port_int,
		SSLInfo: ssl_info,
	}

	return targetServer, nil
}
