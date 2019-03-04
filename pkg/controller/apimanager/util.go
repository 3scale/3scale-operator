package apimanager

import (
	"reflect"

	v1 "k8s.io/api/core/v1"
)

func secretStringDataToData(stringData map[string]string) map[string][]byte {
	data := map[string][]byte{}
	for k, v := range stringData {
		data[k] = []byte(v)
	}
	return data
}

func secretDataToStringData(data map[string][]byte) map[string]string {
	stringData := map[string]string{}
	for k, v := range data {
		stringData[k] = string(v)
	}
	return stringData
}

// As of now we define two secrets as equal when
// their labels, annotations, finalizers, ownerReferences and
// data is equal
func secretsEqual(s1, s2 *v1.Secret) bool {
	if !reflect.DeepEqual(s1.Labels, s2.Labels) {
		return false
	}

	if !reflect.DeepEqual(s1.Annotations, s2.Annotations) {
		return false
	}

	if !reflect.DeepEqual(s1.Finalizers, s2.Finalizers) {
		return false
	}

	if !reflect.DeepEqual(s1.OwnerReferences, s2.OwnerReferences) {
		return false
	}

	if !reflect.DeepEqual(s1.Data, s2.Data) {
		return false
	}

	return true
}
