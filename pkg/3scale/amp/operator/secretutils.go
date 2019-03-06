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

// Returns a new map containing the contents of `from` and the
// contents of `to`. The value for entries with duplicated keys
// will be that of `from`
func mergeSecretData(from, to map[string][]byte) map[string][]byte {
	result := map[string][]byte{}
	for key := range to {
		val := make([]byte, len(to[key]))
		copy(val, to[key])
		result[key] = val
	}

	for key := range from {
		val := make([]byte, len(from[key]))
		copy(val, from[key])
		result[key] = val
	}

	return result
}
