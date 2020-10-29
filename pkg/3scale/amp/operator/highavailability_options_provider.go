package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
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
	err := h.setBackendRedisOptions()
	if err != nil {
		return nil, err
	}
	err = h.setSystemRedisOptions()
	if err != nil {
		return nil, err
	}
	err = h.setSystemDatabaseOptions()
	if err != nil {
		return nil, err
	}

	h.options.BackendRedisLabels = h.backendRedisLabels()
	h.options.SystemRedisLabels = h.SystemDatabaseLabels()
	h.options.SystemDatabaseLabels = h.SystemDatabaseLabels()

	err = h.options.Validate()
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
			component.DefaultBackendStorageSentinelHosts(),
		},
		{
			&h.options.BackendRedisStorageSentinelRole,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisStorageSentinelRoleFieldName,
			component.DefaultBackendStorageSentinelRole(),
		},
		{
			&h.options.BackendRedisQueuesSentinelHosts,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisQueuesSentinelHostsFieldName,
			component.DefaultBackendQueuesSentinelHosts(),
		},
		{
			&h.options.BackendRedisQueuesSentinelRole,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisQueuesSentinelRoleFieldName,
			component.DefaultBackendQueuesSentinelRole(),
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
		{
			&h.options.SystemMessageBusRedisURL,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisMessageBusRedisURLFieldName,
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
			component.DefaultSystemRedisSentinelHosts(),
		},
		{
			&h.options.SystemRedisSentinelsRole,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisSentinelRole,
			component.DefaultSystemRedisSentinelRole(),
		},
		{
			&h.options.SystemMessageBusRedisSentinelsHosts,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisMessageBusSentinelHosts,
			component.DefaultSystemMessageBusRedisSentinelHosts(),
		},
		{
			&h.options.SystemMessageBusRedisSentinelsRole,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisMessageBusSentinelRole,
			component.DefaultSystemMessageBusRedisSentinelRole(),
		},
		{
			&h.options.SystemRedisNamespace,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisNamespace,
			component.DefaultSystemRedisNamespace(),
		},
		{
			&h.options.SystemMessageBusRedisNamespace,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisMessageBusRedisNamespace,
			component.DefaultSystemMessageBusRedisNamespace(),
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
