package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	oprand "github.com/3scale/3scale-operator/pkg/crypto/rand"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (o *OperatorBackendOptionsProvider) GetBackendOptions() (*component.BackendOptions, error) {
	optProv := component.BackendOptionsBuilder{}
	optProv.AppLabel(*o.APIManagerSpec.AppLabel)
	optProv.TenantName(*o.APIManagerSpec.TenantName)
	optProv.WildcardDomain(o.APIManagerSpec.WildcardDomain)

	err := o.setSecretBasedOptions(&optProv)
	if err != nil {
		return nil, err
	}

	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Backend Options - %s", err)
	}
	return res, nil
}

func (o *OperatorBackendOptionsProvider) setSecretBasedOptions(b *component.BackendOptionsBuilder) error {
	err := o.setBackendInternalApiOptions(b)
	if err != nil {
		return fmt.Errorf("unable to create Backend Options - %s", err)
	}
	err = o.setBackendListenerOptions(b)
	if err != nil {
		return fmt.Errorf("unable to create Backend Options - %s", err)
	}
	err = o.setBackendRedisOptions(b)
	if err != nil {
		return fmt.Errorf("unable to create Backend Options - %s", err)
	}

	return nil
}

func (o *OperatorBackendOptionsProvider) setBackendInternalApiOptions(b *component.BackendOptionsBuilder) error {
	defaultSystemBackendUsername := "3scale_api_user"
	defaultSystemBackendPassword := oprand.String(8)

	currSecret, err := getSecret(component.BackendSecretInternalApiSecretName, o.Namespace, o.Client)
	if err != nil {
		if errors.IsNotFound(err) {
			// Set options defaults
			b.SystemBackendUsername(defaultSystemBackendUsername)
			b.SystemBackendPassword(defaultSystemBackendPassword)
		} else {
			return err
		}
	} else {
		// If a field of a secret already exists in the deployed secret then
		// We do not modify it. Otherwise we set a default value
		secretData := currSecret.Data
		b.SystemBackendUsername(getSecretDataValueOrDefault(secretData, component.BackendSecretInternalApiUsernameFieldName, defaultSystemBackendUsername))
		b.SystemBackendPassword(getSecretDataValueOrDefault(secretData, component.BackendSecretInternalApiPasswordFieldName, defaultSystemBackendPassword))
	}

	return nil
}

func (o *OperatorBackendOptionsProvider) setBackendListenerOptions(b *component.BackendOptionsBuilder) error {
	currSecret, err := getSecret(component.BackendSecretBackendListenerSecretName, o.Namespace, o.Client)
	if err != nil {
		if errors.IsNotFound(err) {
			// Do nothing because there are no required options for related to the Backend Secret Listener
		} else {
			return err
		}
	} else {
		secretData := currSecret.Data
		var result *string
		result = getSecretDataValue(secretData, component.BackendSecretBackendListenerServiceEndpointFieldName)
		if result != nil {
			b.ListenerServiceEndpoint(*result)
		}
		result = getSecretDataValue(secretData, component.BackendSecretBackendListenerRouteEndpointFieldName)
		if result != nil {
			b.ListenerRouteEndpoint(*result)
		}
	}

	return nil
}

func (o *OperatorBackendOptionsProvider) setBackendRedisOptions(b *component.BackendOptionsBuilder) error {
	currSecret, err := getSecret(component.BackendSecretBackendRedisSecretName, o.Namespace, o.Client)
	if err != nil {
		if errors.IsNotFound(err) {
			// Do nothing because there are no required options for related to the Backend Secret Listener
		} else {
			return err
		}
	} else {
		secretData := currSecret.Data
		var result *string
		result = getSecretDataValue(secretData, component.BackendSecretBackendRedisStorageURLFieldName)
		if result != nil {
			b.RedisStorageURL(*result)
		}
		result = getSecretDataValue(secretData, component.BackendSecretBackendRedisQueuesURLFieldName)
		if result != nil {
			b.RedisQueuesURL(*result)
		}
		result = getSecretDataValue(secretData, component.BackendSecretBackendRedisStorageSentinelHostsFieldName)
		if result != nil {
			b.RedisStorageSentinelHosts(*result)
		}
		result = getSecretDataValue(secretData, component.BackendSecretBackendRedisStorageSentinelRoleFieldName)
		if result != nil {
			b.RedisStorageSentinelRole(*result)
		}
		result = getSecretDataValue(secretData, component.BackendSecretBackendRedisQueuesSentinelHostsFieldName)
		if result != nil {
			b.RedisStorageSentinelHosts(*result)
		}
		result = getSecretDataValue(secretData, component.BackendSecretBackendRedisQueuesSentinelRoleFieldName)
		if result != nil {
			b.RedisQueuesSentinelRole(*result)
		}
	}

	return nil
}
