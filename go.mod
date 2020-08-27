module github.com/zambien/terraform-provider-apigee

go 1.13

require (
	github.com/17media/structs v0.0.0-20200317074636-7872972ebe57
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/hashicorp/terraform v0.12.13
	github.com/sethgrid/pester v0.0.0-20190127155807-68a33a018ad0 // indirect
	github.com/zambien/go-apigee-edge v0.0.0-20191101145538-e45257f96262
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
