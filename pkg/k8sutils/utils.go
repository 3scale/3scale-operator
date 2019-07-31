package k8sutils

import (
	v1 "k8s.io/api/core/v1"
)

func SecretStringDataFromData(secret v1.Secret) map[string]string {
	stringData := map[string]string{}

	for k, v := range secret.Data {
		stringData[k] = string(v)
	}
	return stringData
}
