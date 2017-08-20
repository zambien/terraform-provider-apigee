# terraform-provider-apigee

A Terraform Apigee provider.

Allows Terraform deployments and management of Apigee API proxies and target servers.

I plan to create the Product resource next and will see where interest goes from there.

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

# A proxy deployment
resource "apigee_api_proxy_deployment" "helloworld_proxy_deployment" {
   proxy_name   = "${apigee_api_proxy.helloworld_proxy.name}"
   org          = "${var.org}"
   env          = "${var.env}"
   revision     = "${apigee_api_proxy.helloworld_proxy.revision}"
}

# A target server
resource "apigee_target_servers" "helloworld_target_server" {
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
Please read [our contribution guidelines.](https://github.com/zambien/terraform-provider-apigee/blob/master/CONTRIBUTING.md)

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