package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type HighAvailabilityOptionsProvider struct {
	apimanager   *appsv1alpha1.APIManager
	namespace    string
	client       client.Client
	options      *component.HighAvailabilityOptions
	secretSource *helper.SecretSource
}

func NewHighAvailabilityOptionsProvider(apimanager *appsv1alpha1.APIManager, namespace string, client client.Client) *HighAvailabilityOptionsProvider {
	return &HighAvailabilityOptionsProvider{
		apimanager:   apimanager,
		namespace:    namespace,
		client:       client,
		options:      component.NewHighAvailabilityOptions(),
		secretSource: helper.NewSecretSource(client, namespace),
	}
}

func (h *HighAvailabilityOptionsProvider) GetHighAvailabilityOptions() (*component.HighAvailabilityOptions, error) {
	setOptionsFns := []func() error{}

	if h.apimanager.IsExternal(appsv1alpha1.BackendRedis) {
		setOptionsFns = append(setOptionsFns, h.setBackendRedisOptions)
	}
	if h.apimanager.IsExternal(appsv1alpha1.SystemRedis) {
		setOptionsFns = append(setOptionsFns, h.setSystemRedisOptions)
	}
	if h.apimanager.IsExternal(appsv1alpha1.SystemDatabase) {
		setOptionsFns = append(setOptionsFns, h.setSystemDatabaseOptions)
	}

	for _, setOptions := range setOptionsFns {
		if err := setOptions(); err != nil {
			return nil, err
		}
	}

	err := h.options.Validate()
	return h.options, err
}

func (h *HighAvailabilityOptionsProvider) setBackendRedisOptions() error {
	cases := []struct {
		field       *string
		secretName  string
		secretField string
	}{
		{
			&h.options.BackendRedisStorageEndpoint,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisStorageURLFieldName,
		},
		{
			&h.options.BackendRedisQueuesEndpoint,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisQueuesURLFieldName,
		},
	}

	for _, option := range cases {
		val, err := h.secretSource.RequiredFieldValueFromRequiredSecret(option.secretName, option.secretField)
		if err != nil {
			return err
		}
		*option.field = val
	}

	// Optional fields
	casesWithDefault := []struct {
		field       *string
		secretName  string
		secretField string
		defValue    string
	}{
		{
			&h.options.BackendRedisStorageSentinelHosts,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisStorageSentinelHostsFieldName,
			defaultBackendStorageSentinelHosts(),
		},
		{
			&h.options.BackendRedisStorageSentinelRole,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisStorageSentinelRoleFieldName,
			defaultBackendStorageSentinelRole(),
		},
		{
			&h.options.BackendRedisQueuesSentinelHosts,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisQueuesSentinelHostsFieldName,
			defaultBackendQueuesSentinelHosts(),
		},
		{
			&h.options.BackendRedisQueuesSentinelRole,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisQueuesSentinelRoleFieldName,
			defaultBackendQueuesSentinelRole(),
		},
	}

	for _, option := range casesWithDefault {
		val, err := h.secretSource.FieldValueFromRequiredSecret(option.secretName, option.secretField, option.defValue)
		if err != nil {
			return err
		}
		*option.field = val
	}

	return nil
}

func (h *HighAvailabilityOptionsProvider) setSystemRedisOptions() error {
	cases := []struct {
		field       *string
		secretName  string
		secretField string
	}{
		{
			&h.options.SystemRedisURL,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisURLFieldName,
		},
	}

	for _, option := range cases {
		val, err := h.secretSource.RequiredFieldValueFromRequiredSecret(option.secretName, option.secretField)
		if err != nil {
			return err
		}
		*option.field = val
	}

	// Optional fields
	casesWithDefault := []struct {
		field       *string
		secretName  string
		secretField string
		defValue    string
	}{
		{
			&h.options.SystemRedisSentinelsHosts,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisSentinelHosts,
			defaultSystemRedisSentinelHosts(),
		},
		{
			&h.options.SystemRedisSentinelsRole,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisSentinelRole,
			defaultSystemRedisSentinelRole(),
		},
	}

	for _, option := range casesWithDefault {
		val, err := h.secretSource.FieldValueFromRequiredSecret(option.secretName, option.secretField, option.defValue)
		if err != nil {
			return err
		}
		*option.field = val
	}

	return nil

}

func (h *HighAvailabilityOptionsProvider) setSystemDatabaseOptions() error {
	val, err := h.secretSource.RequiredFieldValueFromRequiredSecret(
		component.SystemSecretSystemDatabaseSecretName, component.SystemSecretSystemDatabaseURLFieldName)
	if err != nil {
		return err
	}
	h.options.SystemDatabaseURL = val
	return nil
}

func (h *HighAvailabilityOptionsProvider) backendRedisLabels() map[string]string {
	return map[string]string{
		"app":                  *h.apimanager.Spec.AppLabel,
		"threescale_component": "backend",
	}
}

func (h *HighAvailabilityOptionsProvider) SystemDatabaseLabels() map[string]string {
	return map[string]string{
		"app":                  *h.apimanager.Spec.AppLabel,
		"threescale_component": "system",
	}
}

func defaultSystemRedisSentinelHosts() string {
	return ""
}

func defaultSystemRedisSentinelRole() string {
	return ""
}

func defaultBackendStorageSentinelHosts() string {
	return ""
}

func defaultBackendStorageSentinelRole() string {
	return ""
}

func defaultBackendQueuesSentinelHosts() string {
	return ""
}

func defaultBackendQueuesSentinelRole() string {
	return ""
}
