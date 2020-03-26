package apigee

import (
	"github.com/17media/structs"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zambien/go-apigee-edge"
	"reflect"
	"sort"
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

func apiProductsListFromCredentials(credentialApiProducts []apigee.CredentialApiProduct) []string {

	stringList := []string{}

	for _, apiProduct := range credentialApiProducts {
		stringList = append(stringList, apiProduct.ApiProduct)
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

func mapFromCredentials(credentials []apigee.Credential) []interface{} {

	result := make([]interface{}, 0, len(credentials))

	for _, elem := range credentials {
		credentialMap := structs.Map(elem)
		result = append(result, credentialMap)
	}

	return result
}

func arraySortedEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	a_copy := make([]string, len(a))
	b_copy := make([]string, len(b))

	copy(a_copy, a)
	copy(b_copy, b)

	sort.Strings(a_copy)
	sort.Strings(b_copy)

	return reflect.DeepEqual(a_copy, b_copy)
}

func updateResourceOnSortedArrayChange(d *schema.ResourceData, key string, newValues []string) {
	currentValues := getStringList(key, d)
	if currentValues != nil && !arraySortedEqual(currentValues, newValues) {
		d.Set(key, newValues)
	}
}
