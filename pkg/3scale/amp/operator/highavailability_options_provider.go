package operator

import (
	"errors"
	"fmt"
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

type SecretField struct {
	field           *string
	secretFieldName string
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

	// Redis TLS fields
	var tlsFieldsErrs []error
	if h.apimanager.IsBackendRedisTLSEnabled() {
		requiredFields := []SecretField{
			{&h.options.BackendRedisSslCa, "REDIS_SSL_CA"},
			{&h.options.BackendRedisSslCert, "REDIS_SSL_CERT"},
			{&h.options.BackendRedisSslKey, "REDIS_SSL_KEY"},
		}
		err := h.validateRedisTLSFields(component.BackendSecretBackendRedisSecretName, requiredFields)
		if err != nil {
			tlsFieldsErrs = append(tlsFieldsErrs, fmt.Errorf("'backendRedisTLSEnabled: true' is set in apimanager. Secret validation errors: %v", err))
		}
	}
	if h.apimanager.IsQueuesRedisTLSEnabled() {
		requiredFields := []SecretField{
			{&h.options.BackendRedisQueuesSslCa, "REDIS_SSL_QUEUES_CA"},
			{&h.options.BackendRedisQueuesSslCert, "REDIS_SSL_QUEUES_CERT"},
			{&h.options.BackendRedisQueuesSslKey, "REDIS_SSL_QUEUES_KEY"},
		}
		err := h.validateRedisTLSFields(component.BackendSecretBackendRedisSecretName, requiredFields)
		if err != nil {
			tlsFieldsErrs = append(tlsFieldsErrs, fmt.Errorf("'queuesRedisTLSEnabled: true' is set in apimanager. Secret validation errors: %v", err))
		}
	}
	if len(tlsFieldsErrs) > 0 {
		return fmt.Errorf("validation errors for Redis TLS configuration in 'backend-redis' secret: %v", errors.Join(tlsFieldsErrs...))
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

	// Redis TLS fields
	if h.apimanager.IsSystemRedisTLSEnabled() {
		requiredFields := []SecretField{
			{&h.options.SystemRedisSslCa, "REDIS_SSL_CA"},
			{&h.options.SystemRedisSslCert, "REDIS_SSL_CERT"},
			{&h.options.SystemRedisSslKey, "REDIS_SSL_KEY"},
		}
		errs := h.validateRedisTLSFields(component.SystemSecretSystemRedisSecretName, requiredFields)
		if len(errs) > 0 {
			return fmt.Errorf("validation errors for Redis TLS configuration in 'system-redis' secret: %v", errors.Join(errs...))
		}
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

func (h *HighAvailabilityOptionsProvider) validateRedisTLSFields(secretName string, fields []SecretField) []error {
	var errs []error
	for _, field := range fields {
		val, err := h.secretSource.RequiredFieldValueFromRequiredSecret(secretName, field.secretFieldName)
		if err != nil {
			errs = append(errs, fmt.Errorf("%w", err))
		}
		*field.field = val
	}
	return errs
}
