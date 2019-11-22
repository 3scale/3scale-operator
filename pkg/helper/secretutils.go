package helper

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func GetSecretDataValueOrDefault(secretData map[string][]byte, fieldName string, defaultValue string) string {
	if value, exists := secretData[fieldName]; exists {
		return string(value)
	}
	return defaultValue
}

func GetSecret(name string, namespace string, client k8sclient.Client) (*v1.Secret, error) {
	secret := &v1.Secret{}

	objKey := k8sclient.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}

	err := client.Get(context.TODO(), objKey, secret)
	if err != nil {
		return secret, err
	}

	return secret, nil
}

func GetSecretDataValue(secretData map[string][]byte, fieldName string) *string {
	if value, exists := secretData[fieldName]; exists {
		resultStr := string(value)
		return &resultStr
	} else {
		return nil
	}
}

func GetSecretDataFromStringData(secretStringData map[string]string) map[string][]byte {
	result := map[string][]byte{}
	for k, v := range secretStringData {
		result[k] = []byte(v)
	}
	return result
}

func GetSecretStringDataFromData(secretData map[string][]byte) map[string]string {
	result := map[string]string{}
	for k, v := range secretData {
		result[k] = string(v)
	}
	return result
}

// Returns a new map containing the contents of `from` and the
// contents of `to`. The value for entries with duplicated keys
// will be that of `from`
func MergeSecretData(from, to map[string][]byte) map[string][]byte {
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

type SecretSource struct {
	client    client.Client
	namespace string
}

func NewSecretSource(client client.Client, namespace string) *SecretSource {
	return &SecretSource{
		client:    client,
		namespace: namespace,
	}
	// TODO implement caching??
}

func (s *SecretSource) FieldValue(secretName, fieldName string, def string) (*string, error) {
	return s.fieldReader(secretName, fieldName, false, false, def)
}

func (s *SecretSource) FieldValueFromRequiredSecret(secretName, fieldName string, def string) (*string, error) {
	return s.fieldReader(secretName, fieldName, true, false, def)
}

func (s *SecretSource) RequiredFieldValue(secretName, fieldName string) (*string, error) {
	return s.fieldReader(secretName, fieldName, false, true, "")
}

func (s *SecretSource) RequiredFieldValueFromRequiredSecret(secretName, fieldName string) (*string, error) {
	return s.fieldReader(secretName, fieldName, true, true, "")
}

func (s *SecretSource) fieldReader(secretName, fieldName string, secretRequired, fieldRequired bool, def string) (*string, error) {
	secret, err := GetSecret(secretName, s.namespace, s.client)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
		// secret not found
		if secretRequired {
			return nil, err
		}
	}
	// when secret is not found, it behaves like an empty secret
	result := GetSecretDataValue(secret.Data, fieldName)
	if fieldRequired && result == nil {
		return nil, fmt.Errorf("Secret field '%s' is required in secret '%s'", fieldName, secretName)
	}

	if result == nil {
		result = &def
	}

	return result, nil
}
