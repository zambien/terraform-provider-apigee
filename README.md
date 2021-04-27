# terraform-provider-apigee

A Terraform Apigee provider.

Allows Terraform deployments and management of Apigee API proxies, deployments, products, companies/developers/apps, and target servers.

https://registry.terraform.io/providers/zambien/apigee/

- [terraform-provider-apigee](#terraform-provider-apigee)
  * [TFVARS for provider](#tfvars-for-provider)
  * [Simple Example](#simple-example)
  * [Contributions](#contributions)
  * [Building](#building)
  * [Testing](#testing)
      - [Set env vars for test using username/password:](#set-env-vars-for-test-using-username-password-)
      - [Set env vars for test using access token:](#set-env-vars-for-test-using-access-token-)
  * [Releasing](#releasing)
  * [Known issues](#known-issues)
  
## TFVARS for provider

```
APIGEE_BASE_URI="https://someinternalapigee.yourdomain.suffix" # optional... defaults to Apigee's SaaS
APIGEE_ORG="my-really-cool-apigee-org-name"

# To authenticate with Apigee you can use user and password
APIGEE_USER="some_dude@domain.suffix"
APIGEE_PASSWORD="for_the_love_of_pete_please_use_a_strong_password"

# Or you can use an Access Token from Apigee OAuth
APIGEE_ACCESS_TOKEN="my-access-token"
```

## Simple Example

```

# note, to test and build the plugin locally uncomment the lines below or do something like it
# provider_version=0.0.x
# mkdir -p ~/.terraform.d/plugins/local/zambien/apigee/${provider_version}/linux_amd64/
# mv terraform-provider-apigee-v${provider_version}-linux64 ~/.terraform.d/plugins/local/zambien/apigee/${provider_version}/linux_amd64/terraform-provider-apigee_v${provider_version}

terraform {
  required_version = ">= 0.14"

  required_providers {
    apigee = {
      # pull from registry
      source = "zambien/apigee"
      # test locally built plugin
      # source = "local/zambien/apigee"
      version = "~> 0.0.23"
    }
  }
}

variable "org" { default = "my-really-cool-apigee-org-name" }
variable "env" { default = "test" }

provider "apigee" {
  base_uri      = "https://someinternalapigeemanagment.yourdomain.suffix"      # optional... defaults to Apigee's SaaS
  org           = "${var.org}"
  user          = "some_dude@domain.suffix"
  password      = "did_u_pick_a_strong_one?"                # Generally speaking, don't put passwords in your tf files... pull from a Vault or something.
}

# This is a normal terraform offering and serves as an example of how you might create a proxy bundle.
data "archive_file" "bundle" {
   type         = "zip"
   source_dir   = "${path.module}/proxy_files"
   output_path  = "${path.module}/proxy_files_bundle/apiproxy.zip"
}

# The API proxy
resource "apigee_api_proxy" "helloworld_proxy" {
   name  = "helloworld-terraformed"                         # The proxy name.
   bundle       = "${data.archive_file.bundle.output_path}" # Apigee APIs require a zip bundle to import a proxy.
   bundle_sha   = "${data.archive_file.bundle.output_sha}"  # The SHA is used to detect changes for plan/apply.
}

# A product
resource "apigee_product" "helloworld_product" {
   name = "helloworld-product"
   display_name = "helloworld-product" # The provider will assume display name is the same as name if you do not set it.
   description = "no one ever fills this out"
   approval_type = "auto"

   api_resources = ["/**"]
   proxies = ["${apigee_api_proxy.helloworld_proxy.name}"]

   # 1000 requests every 2 minutes
   quota = "1000"
   quota_interval = "2"
   quota_time_unit = "minute"

   # See here: http://docs.apigee.com/api-services/content/working-scopes
   scopes = ["READ"]

   attributes = {
      access = "public" # this one is needed to expose the proxy.  The rest of the attributes are custom attrs.  Weird.

      custom1 = "customval1"
      custom2 = "customval2"
   }
   
   environments = ["test"] # Optional.  If none are specified all are allowed per Apigee API.
}

# A proxy deployment
resource "apigee_api_proxy_deployment" "helloworld_proxy_deployment" {
   proxy_name   = "${apigee_api_proxy.helloworld_proxy.name}"
   org          = "${var.org}"
   env          = "${var.env}"

   # NOTE: revision = "latest" 
   # will deploy the latest revision of the api proxy 
   revision     = "latest"
   # OR revision = "1" # for specific revision
}

# A target server
# NOTE: If you want to use the import functionality the resource ID must follow {target_server_name}_{environment}
resource "apigee_target_server" "helloworld_target_server_testing" {
   name = "helloworld_target_server"
   host = "somehost.thatexists.com"
   env = "testing"
   enabled = true
   port = 8080

   ssl_info {
      ssl_enabled = "false"
      client_auth_enabled = "false"
      key_store = ""
      trust_store = ""
      key_alias = ""
      ignore_validation_errors = false
      ciphers = [""]
      protocols = [""]

   }
}

# A developer
resource "apigee_developer" "helloworld_developer" {
   email = "helloworld_email@test.com"                                  # required
   first_name = "helloworld"                                            # required
   last_name = "thelloworld1"                                           # required
   user_name = "helloworld1"                                            # required

   attributes = {                                                         # optional
      DisplayName = "my_awesome_app_updated"
      Notes = "notes_for_developer_app_updated"
	  custom_attribute_name = "custom_attribute_value"
   }
}

# A developer app

resource "apigee_developer_app" "helloworld_developer_app" {
   name = "helloworld_developer_app"                                    # required
   developer_email = "${apigee_developer.helloworld_developer.email}"   # developer email must exist
   api_products = ["${apigee_product.helloworld_product.name}"]         # list must exist
   scopes = ["READ"]                                                    # scopes must exist in the api_product
   callback_url = "https://www.google.com"                              # optional
   key_expires_in = 2592000000                                          # optional

   attributes = {                                                         # optional
      DisplayName = "my_awesome_developer_app"
      Notes = "notes_for_awesome_developer_app"
	  custom_attribute_name = "custom_attribute_value"
   }
}

# A company
resource "apigee_company" "helloworld_company" {
   name = "helloworld_company"                                          # required
   display_name = "some longer description for company"                 # optional

   attributes = {                                                         # optional
      DisplayName = "my-awesome-company"
   }
}

# A company app
resource "apigee_company_app" "helloworld_company_app" {
   name = "helloworld_company_app_name"
   company_name = "${apigee_company.helloworld_company.name}"
   api_products = ["${apigee_product.helloworld_product.name}"]
   scopes = ["READ"]
   callback_url = "https://www.google.com"
}

# Create the shared flow bundle pretty much the same way you create the proxy bundle.
data "archive_file" "sharedflow_bundle" {
   type         = "zip"
   source_dir   = "${path.module}/sharedflow_files"
   output_path  = "${path.module}/sharedflow_files_bundle/sharedflow.zip"
}

# The Shared Flow
resource "apigee_shared_flow" "helloworld_shared_flow" {
   name         = "helloworld-sharedflow-terraformed"                         # The shared flow's name.
   bundle       = "${data.archive_file.sharedflow_bundle.output_path}"        # Apigee APIs require a zip bundle to import a shared flow.
   bundle_sha   = "${data.archive_file.sharedflow_bundle.output_sha}"         # The SHA is used to detect changes for plan/apply.
}

# A Shared Flow deployment
resource "apigee_shared_flow_deployment" "helloworld_shared_flow_deployment" {
   shared_flow_name   = "${apigee_shared_flow.helloworld_shared_flow.name}"
   org                = "${var.org}"
   env                = "${var.env}"

   # NOTE: revision = "latest" 
   # will deploy the latest revision of the shared flow 
   revision     = "latest"
   # OR revision = "1" # for specific revision
}
```

## Contributions
Please read [our contribution guidelines.](https://github.com/zambien/terraform-provider-apigee/blob/master/.github/CONTRIBUTING.md)

## Building
Should be buildable on any terraform version at or higher than 0.9.3.  To build you would use the standard go build command.  For example for MacOS:

`GOOS=darwin GOARCH=amd64 go build -o terraform-provider-apigee-v0.0.X-darwin64`

Windows:
`GOOS=windows GOARCH=amd64 go build -o terraform-provider-apigee-v0.0.X-win64`

Linux:
`GOOS=linux GOARCH=amd64 go build -o terraform-provider-apigee-v0.0.X-linux64`

## Testing
To run tests, use the following commands.  Note that you will need your credentials setup for the tests to run. You can authenticate with your username/password OR an access token from Apigee OAuth.

#### Set env vars for test using username/password:
```
APIGEE_ORG="my-really-cool-apigee-org-name"
APIGEE_USER="some_dude@domain.suffix"
APIGEE_PASSWORD="for_the_love_of_pete_please_use_a_strong_password"
```

#### Set env vars for test using access token:
```
APIGEE_ORG="my-really-cool-apigee-org-name"
APIGEE_ACCESS_TOKEN="my-access-token"
```

From the project root:
`TF_ACC=1 go test -v ./apigee`

To run a single test:
`TF_ACC=1 go test -v ./apigee -run=TestAccDeveloperApp_Updated`

Running in debug mode and capturing debug in a file:
`rm -f /tmp/testlog.txt && TF_ACC=1 TF_LOG=DEBUG TF_LOG_PATH=/tmp/testlog.txt go test -v ./apigee`

## Releasing

We use goreleaser to release versions.  The steps to release are:

```
export GITHUB_TOKEN="A_GITHUB_TOKEN_THAT_HAS_CORRECT_ACCESS_ENTITLEMENTS"
git tag -a v0.0.x -m "Some description of the release"
export GPG_FINGERPRINT=[your_GPG_fingerprint]
touch /tmp/temp.txt && gpg --local-user ${GPG_FINGERPRINT}  --armor --detach-sign /tmp/temp.txt
goreleaser # actually create the release
```

You can read more about goreleaser here:

https://goreleaser.com/

## Known issues

* You will often find the need to run apply twice when updating a proxy.  This has to do with how terraform handles state.  This plugin will be rewritten to combine proxies and proxy deployments to resolve this issue in the future.
