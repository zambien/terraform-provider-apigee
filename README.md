# terraform-provider-apigee
A Terraform Apigee provider focused on Proxies and Deployments.

Allows Terraform deployments and management of Apigee API proxies.

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

# The API proxy in Apigee
resource "apigee_api_proxy" "helloworld_proxy" {
   name  = "helloworld-terraformed"                         # The proxy name.
   bundle       = "${data.archive_file.bundle.output_path}" # Apigee APIs require a zip bundle to import a proxy.
   bundle_sha   = "${data.archive_file.bundle.output_sha}"  # The SHA is used to detect changes for plan/apply.
}

# A proxy deployment in Apigee
resource "apigee_api_proxy_deployment" "helloworld_proxy_deployment" {
   proxy_name   = "${apigee_api_proxy.helloworld_proxy.name}"
   org          = "${var.org}"
   env          = "${var.env}"
   revision     = "${apigee_api_proxy.helloworld_proxy.revision}"
}

# Outputs
output "apigee_api_proxy_name" {
   value = ["${apigee_api_proxy.helloworld_proxy.name}"]
}

output "apigee_api_proxy_deployed_id" {
   value = ["${apigee_api_proxy_deployment.helloworld_proxy_deployment.id}"]
}

output "apigee_api_proxy_deployed_rev" {
   value = ["${apigee_api_proxy_deployment.helloworld_proxy_deployment.revision}"]
}

```

## Issues

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