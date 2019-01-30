package operator

import (
	"context"

	v1 "k8s.io/api/core/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func getSecretDataValueOrDefault(secretData map[string][]byte, fieldName string, defaultValue string) string {
	if value, exists := secretData[fieldName]; exists {
		return string(value)
	}
	return defaultValue
}

func getSecret(name string, namespace string, client k8sclient.Client) (*v1.Secret, error) {
	secret := &v1.Secret{}

	objKey := k8sclient.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}

	err := client.Get(context.TODO(), objKey, secret)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func getSecretDataValue(secretData map[string][]byte, fieldName string) *string {
	if value, exists := secretData[fieldName]; exists {
		resultStr := string(value)
		return &resultStr
	} else {
		return nil
	}
}

/* TODO unfinished
if exists in Current and not exists in expected then remove what's in current ???
ix exits in Current and exists in expected then do not modify current
if not exists in Current and exists in expected then set current to expected
if not exists in Current and not exists in expected then do nothing
*/
// Returns a new map containing the merged data of expected into current.
// If there are duplicate keys then the contents are NOT overwritten
// (this is, current has preference). Also, keys already existing in
// the current map but not existing in expected are removed
func mergeSecretData(current, expected map[string][]byte) map[string][]byte {
	result := map[string][]byte{}
	for key := range expected {
		if _, exists := current[key]; !exists {
			val := make([]byte, len(current[key]))
			copy(val, current[key])
			result[key] = val
		} else {
			val := make([]byte, len(current[key]))
			copy(val, expected[key])
			result[key] = val
		}
	}

	return result
}
