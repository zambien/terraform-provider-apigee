package apigee

import (
	"github.com/zambien/go-apigee-edge"
	"log"
)

// Config holds API and APP keys to authenticate to Datadog.
type Config struct {
	BaseURI string
	User    string
	Pass    string
	Org     string
}

// Client returns a new Apigee client.
func (c *Config) Client() (*apigee.EdgeClient, error) {

	auth := apigee.EdgeAuth{Username: c.User, Password: c.Pass}
	opts := &apigee.EdgeClientOptions{MgmtUrl: c.BaseURI, Org: c.Org, Auth: &auth, Debug: false}
	client, err := apigee.NewEdgeClient(opts)
	if err != nil {
		log.Printf("while initializing Edge client, error:\n%#v\n", err)
		return client, err
	}

	return client, nil
}
