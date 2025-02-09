package helper

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
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
	client      k8sclient.Client
	namespace   string
	secretCache *MemoryCache
}

func NewSecretSource(client k8sclient.Client, namespace string) *SecretSource {
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

func ValidateTLSSecret(nn types.NamespacedName, client k8sclient.Client) error {
	secret := &v1.Secret{}
	err := client.Get(context.TODO(), nn, secret)
	if err != nil {
		return err
	}

	if secret.Type != v1.SecretTypeTLS {
		return fmt.Errorf("required kubernetes.io/tls secret type. Found %s", secret.Type)
	}

	if _, ok := secret.Data[v1.TLSCertKey]; !ok {
		return fmt.Errorf("required secret key, %s, not found", v1.TLSCertKey)
	}

	if _, ok := secret.Data[v1.TLSPrivateKeyKey]; !ok {
		return fmt.Errorf("required secret key, %s, not found", v1.TLSPrivateKeyKey)
	}

	return nil
}

func IsSecretWatchedBy3scale(secret *v1.Secret) bool {
	if secret == nil {
		return false
	}

	existingLabels := secret.Labels

	if existingLabels != nil {
		if _, ok := existingLabels["apimanager.apps.3scale.net/watched-by"]; ok {
			return true
		}
	}

	return false
}

func IsSecretWatchedBy3scaleBySecretName(client k8sclient.Client, secretName, namespace string) bool {
	secret := &v1.Secret{}
	secretKey := k8sclient.ObjectKey{
		Name:      secretName,
		Namespace: namespace,
	}
	err := client.Get(context.TODO(), secretKey, secret)
	if err != nil {
		return false
	}

	existingLabels := secret.Labels
	if existingLabels != nil {
		if _, ok := existingLabels["apimanager.apps.3scale.net/watched-by"]; ok {
			return true
		}
	}

	return false
}

func ValidateRedisURL(url string) error {
	redisPrefix := "rediss://"
	// Check if the URL scheme is "rediss"
	if !strings.HasPrefix(url, redisPrefix) {
		return fmt.Errorf("URL should start with "+redisPrefix+". Found %s", url)
	}
	// Remove the "rediss://" prefix
	url = strings.TrimPrefix(url, redisPrefix)
	// Split the URL into host:port and optional database number
	hostPortDB := strings.Split(url, "/")
	if len(hostPortDB) > 2 {
		return fmt.Errorf("URL has an invalid structure")
	}
	// Check if there's a database number part (e.g., after the slash)
	dbPart := ""
	if len(hostPortDB) == 2 {
		dbPart = hostPortDB[1]
	}
	// Split the host and port part
	hostPort := hostPortDB[0]
	hostPortParts := strings.Split(hostPort, ":")
	if len(hostPortParts) != 2 {
		return fmt.Errorf("URL must contain a host and port, e.g., 'host:port'")
	}
	// Validate the host part (IP or Domain)
	host := hostPortParts[0]
	if net.ParseIP(host) == nil {
		regex := `^(?i)([a-z0-9](-?[a-z0-9])*\.)+[a-z0-9](-?[a-z0-9])*$`
		re := regexp.MustCompile(regex)
		if !re.MatchString(host) {
			return fmt.Errorf("Invalid host (neither IP address nor valid domain): %s", host)
		}
	}
	// Validate the port (port part)
	portStr := hostPortParts[1]
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1024 || port > 65535 {
		return fmt.Errorf("Invalid port number: %s. Valid range is 1024-65535", portStr)
	}
	// Optionally, validate the database number (if it exists)
	if dbPart != "" {
		dbNum, err := strconv.Atoi(dbPart)
		if err != nil || dbNum < 0 {
			return fmt.Errorf("Invalid database number: %s", dbPart)
		}
	}
	return nil
}

func ValidateRedisSentinelHostsForTLS(sentinelHosts string) error {
	// 1. Validate Sentinel Hosts defined in system-redis or backend-redis secrets
	// - if it's a valid to work with Redis Master in TLS mode:
	// 1.1. Empty - Sentine is not defined. Redis Clients will work TLS Redis Master directly
	// 1.2. One of Sentinel Hosts is TLS (rediss://...), - All can use TLS: Redis master, sentinel, clients
	// 2. Not valid configurations:
	// 2.1. One of sentinel hosts has wrong URL format
	// 2.2. All Sentinel hosts are Non-TLS (redis://...). It's not valid as Master is TLS

	if len(sentinelHosts) == 0 {
		return nil
	}
	found := false
	hostList := strings.Split(sentinelHosts, ",")
	// Check if any of the Sentinel hosts uses 'rediss://' (TLS)
	for _, host := range hostList {
		host = strings.TrimSpace(host)
		parsedURL, err := url.Parse(host)
		if err != nil {
			return fmt.Errorf("URL has an invalid structure: %s", host)
		}
		if parsedURL.Scheme == "rediss" {
			// At least one secure Sentinel is found
			found = true
		}
	}
	// If no secure Sentinels found, return false
	if !found {
		return fmt.Errorf("no secure sentinel hosts found for secure redis master, hosts: %s", sentinelHosts)
	}
	return nil

}
