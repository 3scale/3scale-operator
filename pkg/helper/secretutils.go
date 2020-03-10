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

type SecretCacheElement struct {
	Secret *v1.Secret
	Err    error
}

type SecretSource struct {
	client      client.Client
	namespace   string
	secretCache *MemoryCache
}

func NewSecretSource(client client.Client, namespace string) *SecretSource {
	return &SecretSource{
		client:      client,
		namespace:   namespace,
		secretCache: NewMemoryCache(),
	}
}

func (s *SecretSource) FieldValue(secretName, fieldName string, def string) (string, error) {
	return s.fieldReader(secretName, fieldName, false, false, def)
}

func (s *SecretSource) FieldValueFromRequiredSecret(secretName, fieldName string, def string) (string, error) {
	return s.fieldReader(secretName, fieldName, true, false, def)
}

func (s *SecretSource) RequiredFieldValueFromRequiredSecret(secretName, fieldName string) (string, error) {
	return s.fieldReader(secretName, fieldName, true, true, "")
}

func (s *SecretSource) fieldReader(secretName, fieldName string, secretRequired, fieldRequired bool, def string) (string, error) {
	secret, err := s.CachedSecret(secretName)
	if err != nil {
		if !errors.IsNotFound(err) {
			return "", err
		}
		// secret not found
		if secretRequired {
			return "", err
		}
	}
	// when secret is not found, it behaves like an empty secret
	result := GetSecretDataValue(secret.Data, fieldName)
	if fieldRequired && result == nil {
		return "", fmt.Errorf("Secret field '%s' is required in secret '%s'", fieldName, secretName)
	}

	if result == nil {
		result = &def
	}

	return *result, nil
}

func (s *SecretSource) CachedSecret(secretName string) (*v1.Secret, error) {
	var secret *v1.Secret
	secretElementI, err := s.secretCache.Get(secretName)
	if err != nil {
		if err != ErrNonExistentKey {
			return nil, err
		}

		// Key not found in cache, do the actual call
		secret, err = GetSecret(secretName, s.namespace, s.client)
		// Always store result, even when there is error.
		// Save calls when secret not found
		// Lifecycle of this cache instance is expected to be short
		// and limited (one reconciliation loop run)
		s.secretCache.Put(secretName, SecretCacheElement{Secret: secret, Err: err})
	} else {
		secretElement, ok := secretElementI.(SecretCacheElement)
		if !ok {
			return nil, fmt.Errorf("Unexpected error. Secret cache returned non secret object")
		}
		secret = secretElement.Secret
		err = secretElement.Err
	}

	return secret, err
}
