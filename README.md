# terraform-provider-apigee
A Terraform Apigee provider focused on Products and Proxies.

## Example

```
provider "apigee" {
}

resource "apigee_api_proxy" "helloworld_proxy" {
   name_prefix =  "helloworld-terraformed"
   bundle = "${data.archive_file.bundle.output_path}"
   bundle_sha = "${data.archive_file.bundle.output_sha}"
}

data "archive_file" "bundle" {
   type        = "zip"
   source_dir = "${path.module}/proxy_files"
   output_path = "${path.module}/proxy_files_bundle/apiproxy.zip"
}

output "apigee_api_proxy_name" {
   value = ["${apigee_api_proxy.helloworld_proxy.name}"]
}

output "apigee_api_proxy_revision" {
   value = ["${apigee_api_proxy.helloworld_proxy.revision}"]
}
```