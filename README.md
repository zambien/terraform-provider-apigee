# terraform-provider-apigee [WIP]
A Terraform Apigee provider focused on Products and Proxies.

This is a work in progress.  Once it is to the point that is solves the use cases of proxy create and deploy I will submit PR to the terraform-providers repo.

## TFVARS for provider

```
APIGEE_BASE_URI="https://someinternalapigee.yourdomain.wtf" # optional... defaults to Apigee's SaaS
APIGEE_ORG="user-org-name"
APIGEE_USER="user@email.com"
APIGEE_PASSWORD="fortheloveofpetepleaseuseastrongpassword"
```

## Simple Example

```
provider "apigee" {
  org="my_really_cool_org_name"
  user="some_dude@domain.wtf"
  password="didupickastrongone?" # Generally speaking, don't put passwords in your tf files... pull from a Vault or something.
}

resource "apigee_api_proxy" "helloworld_proxy" {
   name_prefix =  "helloworld-terraformed" # used to prepend your proxy name.  Name will be name_prefix + a uuid.
   bundle = "${data.archive_file.bundle.output_path}" # Apigee APIs require a zip bundle to import a proxy.
   bundle_sha = "${data.archive_file.bundle.output_sha}" # The SHA is used to detect changes for plan/apply.
}

# This is a normal terraform offering and serves as an example of how you might create a proxy bundle.
data "archive_file" "bundle" {
   type        = "zip"
   source_dir = "${path.module}/proxy_files"
   output_path = "${path.module}/proxy_files_bundle/apiproxy.zip"
}

# Outputs
output "apigee_api_proxy_name" {
   value = ["${apigee_api_proxy.helloworld_proxy.name}"]
}

output "apigee_api_proxy_revision" {
   value = ["${apigee_api_proxy.helloworld_proxy.revision}"]
}
```