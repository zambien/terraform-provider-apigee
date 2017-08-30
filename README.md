# terraform-provider-apigee

A Terraform Apigee provider.

Allows Terraform deployments and management of Apigee API proxies, deployments, products, and target servers.

## Installation

To to releases and pull the appropriate release for your system: https://github.com/zambien/terraform-provider-apigee/releases

See here for info on how to install the plugin:

https://www.terraform.io/docs/plugins/basics.html

An example of how to do this would be:

1. Make a terraform providers folder in home
`mkdir -p ~/terraform-providers`

2. Download plugin for linux into your home directory
`curl -L https://github.com/zambien/terraform-provider-apigee/releases/download/v0.0.5/terraform-provider-apigee-v0.0.5-linux64 -o ~/terraform-providers/terraform-provider-apigee-v0.0.5-linux64`

3. Add the providers clause if you don't already have one.  Warning, this command will overwrite your .terraformrc!
```
cat << EOF > ~/.terraformrc
providers {
    apigee = "$HOME/terraform-providers/terraform-provider-apigee-v0.0.5-linux64"
}
EOF
```

## TFVARS for provider

```
APIGEE_BASE_URI="https://someinternalapigee.yourdomain.suffix" # optional... defaults to Apigee's SaaS
APIGEE_ORG="my-really-cool-apigee-org-name"
APIGEE_USER="some_dude@domain.suffix"
APIGEE_PASSWORD="for_the_love_of_pete_please_use_a_strong_password"
```

## Simple Example

```

variable "org" { default = "my-really-cool-apigee-org-name" }
variable "env" { default = "test" }

provider "apigee" {
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
   scopes = [""]

   attributes {
      access = "public" # this one is needed to expose the proxy.  The rest of the attributes are custom attrs.  Weird.

      custom1 = "customval1"
      custom2 = "customval2"
   }
}

# A proxy deployment
resource "apigee_api_proxy_deployment" "helloworld_proxy_deployment" {
   proxy_name   = "${apigee_api_proxy.helloworld_proxy.name}"
   org          = "${var.org}"
   env          = "${var.env}"
   revision     = "${apigee_api_proxy.helloworld_proxy.revision}"
}

# A target server
resource "apigee_target_server" "helloworld_target_server" {
   name = "helloworld_target_server"
   host = "somehost.thatexists.com"
   env = "${var.env}"
   enabled = true
   port = 8080

   ssl_info {
      ssl_enabled = false
      client_auth_enabled = false
      key_store = ""
      trust_store = ""
      key_alias = ""
      ignore_validation_errors = false
      ciphers = [""]
      protocols = [""]

   }
}

```

## Contributions
Please read [our contribution guidelines.](https://github.com/zambien/terraform-provider-apigee/blob/master/.github/CONTRIBUTING.md)

## Important Known Issues

Right now if you rev your proxy bundle then apply your deployment will not update automatically if you reference that proxy rev (as in the example above).

To work around the issue you can apply twice:
```
terraform apply && terraform apply
```

Or manually change the revision number in a variable or in the script...
```
resource "apigee_api_proxy_deployment" "helloworld_proxy_deployment" {
   proxy_name   = "${apigee_api_proxy.helloworld_proxy.name}"
   org          = "${var.org}"
   env          = "${var.env}"
   revision     = 4 # the known next revision number
}
```

This is happening due to a known issue in Terraform that should be fixed soon:
https://github.com/hashicorp/terraform/issues/15857