package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorHighAvailabilityOptionsProvider) GetHighAvailabilityOptions() (*component.HighAvailabilityOptions, error) {
	hob := component.HighAvailabilityOptionsBuilder{}

	err := o.setBackendRedisOptions(&hob)
	if err != nil {
		return nil, err
	}
	err = o.setSystemRedisOptions(&hob)
	if err != nil {
		return nil, err
	}
	err = o.setSystemDatabaseOptions(&hob)
	if err != nil {
		return nil, err
	}

	res, err := hob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create HighAvailability Options - %s", err)
	}
	return res, nil
}

func (o *OperatorHighAvailabilityOptionsProvider) setBackendRedisOptions(builder *component.HighAvailabilityOptionsBuilder) error {
	currSecret, err := getSecret(component.BackendSecretBackendRedisSecretName, o.Namespace, o.Client)
	if err != nil {
		return err
	}

	secretData := currSecret.Data
	var result *string
	result = getSecretDataValue(secretData, component.BackendSecretBackendRedisStorageURLFieldName)
	if result == nil {
		return fmt.Errorf("Secret field '%s' is required in secret '%s'", component.BackendSecretBackendRedisStorageURLFieldName, component.BackendSecretBackendRedisSecretName)
	}
	builder.BackendRedisStorageEndpoint(*result)

	result = getSecretDataValue(secretData, component.BackendSecretBackendRedisQueuesURLFieldName)
	if result == nil {
		return fmt.Errorf("Secret field '%s' is required in secret '%s'", component.BackendSecretBackendRedisQueuesURLFieldName, component.BackendSecretBackendRedisSecretName)
	}
	builder.BackendRedisQueuesEndpoint(*result)

	return nil
}

func (o *OperatorHighAvailabilityOptionsProvider) setSystemRedisOptions(builder *component.HighAvailabilityOptionsBuilder) error {
	currSecret, err := getSecret(component.SystemSecretSystemRedisSecretName, o.Namespace, o.Client)
	if err != nil {
		return err
	}

	secretData := currSecret.Data
	var result *string
	result = getSecretDataValue(secretData, component.SystemSecretSystemRedisURLFieldName)
	if result == nil {
		return fmt.Errorf("Secret field '%s' is required in secret '%s'", component.SystemSecretSystemRedisURLFieldName, component.SystemSecretSystemRedisSecretName)
	}
	builder.SystemRedisURL(*result)

	result = getSecretDataValue(secretData, component.SystemSecretSystemRedisMessageBusRedisURLFieldName)
	if result == nil {
		return fmt.Errorf("Secret field '%s' is required in secret '%s'", component.SystemSecretSystemRedisMessageBusRedisURLFieldName, component.SystemSecretSystemRedisSecretName)
	}
	builder.SystemMessageBusRedisURL(*result)

	return nil
}

func (o *OperatorHighAvailabilityOptionsProvider) setSystemDatabaseOptions(builder *component.HighAvailabilityOptionsBuilder) error {
	currSecret, err := getSecret(component.SystemSecretSystemDatabaseSecretName, o.Namespace, o.Client)
	if err != nil {
		return err
	}

	secretData := currSecret.Data
	var result *string
	result = getSecretDataValue(secretData, component.SystemSecretSystemDatabaseURLFieldName)
	if result == nil {
		return fmt.Errorf("Secret field '%s' is required in secret '%s'", component.SystemSecretSystemDatabaseURLFieldName, component.SystemSecretSystemDatabaseSecretName)
	}
	builder.SystemDatabaseURL(*result)

	return nil
}
