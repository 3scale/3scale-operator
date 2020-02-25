package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type HighAvailabilityOptionsProvider struct {
	namespace    string
	client       client.Client
	options      *component.HighAvailabilityOptions
	secretSource *helper.SecretSource
}

func NewHighAvailabilityOptionsProvider(namespace string, client client.Client) *HighAvailabilityOptionsProvider {
	return &HighAvailabilityOptionsProvider{
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

	// not required for operator, but required for templates
	h.options.AppLabel = "-"
	h.options.BackendRedisQueuesSentinelHosts = "-"
	h.options.BackendRedisQueuesSentinelRole = "-"
	h.options.BackendRedisStorageSentinelHosts = "-"
	h.options.BackendRedisStorageSentinelRole = "-"
	h.options.SystemRedisSentinelsHosts = "-"
	h.options.SystemRedisSentinelsRole = "-"
	h.options.SystemMessageBusRedisSentinelsHosts = "-"
	h.options.SystemMessageBusRedisSentinelsRole = "-"

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
