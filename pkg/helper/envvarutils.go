package helper

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var redisSecretsEnvPathMap = map[string]struct {
	path      string
	sslEnvVar string
}{
	"REDIS_CA_FILE":             {"/tls/system-redis/system-redis-ca.crt", "SSL_CA"},
	"REDIS_CLIENT_CERT":         {"/tls/system-redis/system-redis-client.crt", "SSL_CERT"},
	"REDIS_PRIVATE_KEY":         {"/tls/system-redis/system-redis-private.key", "SSL_KEY"},
	"CONFIG_REDIS_CA_FILE":      {"/tls/backend-redis-ca.crt", "SSL_CA"},
	"CONFIG_REDIS_CERT":         {"/tls/backend-redis-client.crt", "SSL_CERT"},
	"CONFIG_REDIS_PRIVATE_KEY":  {"/tls/backend-redis-private.key", "SSL_KEY"},
	"CONFIG_QUEUES_CA_FILE":     {"/tls/config-queues-ca.crt", "SSL_QUEUES_CA"},
	"CONFIG_QUEUES_CERT":        {"/tls/config-queues-client.crt", "SSL_QUEUES_CERT"},
	"CONFIG_QUEUES_PRIVATE_KEY": {"/tls/config-queues-private.key", "SSL_QUEUES_KEY"},
}

func EnvVarFromConfigMap(envVarName string, configMapName, configMapKey string) v1.EnvVar {
	return v1.EnvVar{
		Name: envVarName,
		ValueFrom: &v1.EnvVarSource{
			ConfigMapKeyRef: &v1.ConfigMapKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: configMapName,
				},
				Key: configMapKey,
			},
		},
	}
}

func EnvVarFromConfigMapOptional(envVarName string, configMapName, configMapKey string) v1.EnvVar {
	trueValue := true
	return v1.EnvVar{
		Name: envVarName,
		ValueFrom: &v1.EnvVarSource{
			ConfigMapKeyRef: &v1.ConfigMapKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: configMapName,
				},
				Key:      configMapKey,
				Optional: &trueValue,
			},
		},
	}
}

func EnvVarFromValue(name string, value string) v1.EnvVar {
	return v1.EnvVar{
		Name:  name,
		Value: value,
	}
}

func EnvVarFromSecret(envVarName string, secretName, secretKey string) v1.EnvVar {
	return v1.EnvVar{
		Name: envVarName,
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: secretName,
				},
				Key: secretKey,
			},
		},
	}
}

func EnvVarFromSecretOptional(envVarName string, secretName, secretKey string) v1.EnvVar {
	trueValue := true
	return v1.EnvVar{
		Name: envVarName,
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: secretName,
				},
				Key:      secretKey,
				Optional: &trueValue,
			},
		},
	}
}

// FindEnvVar returns the smallest index i at which x == a[i],
// or -1 if there is no such index.
func FindEnvVar(a []v1.EnvVar, x string) int {
	for i, n := range a {
		if n.Name == x {
			return i
		}
	}
	return -1
}

// EnsureEnvVar updates existingEnvVars with desired, reconciling any
// possible differences. It returns whether existingEnvVars has been modified.
func EnsureEnvVar(desired v1.EnvVar, existingEnvVars *[]v1.EnvVar) bool {
	update := false
	envVarExists := false
	for idx := range *existingEnvVars {
		if (*existingEnvVars)[idx].Name == desired.Name {
			envVarExists = true
			if !reflect.DeepEqual((*existingEnvVars)[idx], desired) {
				(*existingEnvVars)[idx] = desired
				update = true
			}
			break
		}
	}

	if !envVarExists {
		*existingEnvVars = append(*existingEnvVars, desired)
		update = true
	}

	return update
}

// RemoveDuplicateEnvVars removes duplicate env vars by name from a slice
func RemoveDuplicateEnvVars(envVars []v1.EnvVar) []v1.EnvVar {
	set := map[string]v1.EnvVar{}
	var result []v1.EnvVar

	for idx := range envVars {
		if _, ok := set[envVars[idx].Name]; !ok {
			set[envVars[idx].Name] = envVars[idx]
			result = append(result, envVars[idx])
		}
	}

	return result
}

// EnvVarReconciler implements basic env var reconciliation.
// Added when in desired and not in existing
// Updated when in desired and in existing but not equal
// Removed when not in desired and exists in existing Deployment
func EnvVarReconciler(desired []v1.EnvVar, existing *[]v1.EnvVar, envVar string) bool {
	update := false

	if existing == nil {
		*existing = make([]v1.EnvVar, 0)
	}

	desiredIdx := FindEnvVar(desired, envVar)
	existingIdx := FindEnvVar(*existing, envVar)

	if desiredIdx < 0 && existingIdx >= 0 {
		// env var exists in existing and does not exist in desired => Remove from the list
		// shift all of the elements at the right of the deleting index by one to the left
		*existing = append((*existing)[:existingIdx], (*existing)[existingIdx+1:]...)
		update = true
	} else if desiredIdx < 0 && existingIdx < 0 {
		// env var does not exist in existing and does not exist in desired => NOOP
	} else if desiredIdx >= 0 && existingIdx < 0 {
		// env var does not exist in existing and exists in desired => ADD it
		*existing = append(*existing, desired[desiredIdx])
		update = true
	} else {
		// env var exists in existing and exists in desired
		if !reflect.DeepEqual((*existing)[existingIdx], desired[desiredIdx]) {
			(*existing)[existingIdx] = desired[desiredIdx]
			update = true
		}
	}
	return update
}

func EnvVarPathFromRedisSecret(secretName string, envVar string) string {
	cfg, _ := config.GetConfig()
	client, _ := client.New(cfg, client.Options{})
	namespace, _ := GetOperatorNamespace()
	secret, err := GetSecret(secretName, namespace, client)
	if err != nil {
		return ""
	}
	envPathRecord, exists := redisSecretsEnvPathMap[envVar]
	if !exists {
		return ""
	}
	if sslData, exists := secret.Data[envPathRecord.sslEnvVar]; !exists || len(sslData) == 0 {
		return ""
	}
	secret.Data[envVar] = []byte(envPathRecord.path)
	err = client.Update(context.TODO(), secret)
	if err != nil {
		return ""
	}
	return envPathRecord.path
}
