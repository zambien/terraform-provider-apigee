package apigee

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zambien/go-apigee-edge"
)

func flattenStringList(list []string) []interface{} {

	vs := make([]interface{}, 0, len(list))

	for _, v := range list {
		vs = append(vs, &v)
	}

	return vs
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

func attributesFromMap(attributes map[string]interface{}) []apigee.Attribute {

	result := make([]apigee.Attribute, 0, len(attributes))

	for k, v := range attributes {
		t := apigee.Attribute{
			Name:  k,
			Value: v.(string),
		}
		result = append(result, t)
	}

	return result
}
