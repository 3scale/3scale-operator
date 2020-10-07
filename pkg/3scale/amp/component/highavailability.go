package component

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HighAvailability struct {
	Options *HighAvailabilityOptions
}

var HighlyAvailableExternalDatabases = map[string]bool{
	"backend-redis": true,
	"system-redis":  true,
	"system-mysql":  true,
}

func NewHighAvailability(options *HighAvailabilityOptions) *HighAvailability {
	return &HighAvailability{Options: options}
}

func (ha *HighAvailability) SystemDatabaseSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSecretSystemDatabaseSecretName,
			Labels: ha.Options.SystemDatabaseLabels,
		},
		StringData: map[string]string{
			SystemSecretSystemDatabaseURLFieldName: ha.Options.SystemDatabaseURL,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (ha *HighAvailability) BackendRedisSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   BackendSecretBackendRedisSecretName,
			Labels: ha.Options.BackendRedisLabels,
		},
		StringData: map[string]string{
			BackendSecretBackendRedisStorageURLFieldName:           ha.Options.BackendRedisStorageEndpoint,
			BackendSecretBackendRedisQueuesURLFieldName:            ha.Options.BackendRedisQueuesEndpoint,
			BackendSecretBackendRedisStorageSentinelHostsFieldName: ha.Options.BackendRedisStorageSentinelHosts,
			BackendSecretBackendRedisStorageSentinelRoleFieldName:  ha.Options.BackendRedisStorageSentinelRole,
			BackendSecretBackendRedisQueuesSentinelHostsFieldName:  ha.Options.BackendRedisQueuesSentinelHosts,
			BackendSecretBackendRedisQueuesSentinelRoleFieldName:   ha.Options.BackendRedisQueuesSentinelRole,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (ha *HighAvailability) SystemRedisSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSecretSystemRedisSecretName,
			Labels: ha.Options.SystemRedisLabels,
		},
		StringData: map[string]string{
			SystemSecretSystemRedisURLFieldName:                ha.Options.SystemRedisURL,
			SystemSecretSystemRedisSentinelHosts:               ha.Options.SystemRedisSentinelsHosts,
			SystemSecretSystemRedisSentinelRole:                ha.Options.SystemRedisSentinelsRole,
			SystemSecretSystemRedisMessageBusRedisURLFieldName: ha.Options.SystemMessageBusRedisURL,
			SystemSecretSystemRedisMessageBusSentinelHosts:     ha.Options.SystemMessageBusRedisSentinelsHosts,
			SystemSecretSystemRedisMessageBusSentinelRole:      ha.Options.SystemMessageBusRedisSentinelsRole,
			SystemSecretSystemRedisNamespace:                   ha.Options.SystemRedisNamespace,
			SystemSecretSystemRedisMessageBusRedisNamespace:    ha.Options.SystemMessageBusRedisNamespace,
		},
		Type: v1.SecretTypeOpaque,
	}
}
