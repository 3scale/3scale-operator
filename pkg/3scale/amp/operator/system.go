package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	oprand "github.com/3scale/3scale-operator/pkg/crypto/rand"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (o *OperatorSystemOptionsProvider) GetSystemOptions() (*component.SystemOptions, error) {
	optProv := component.SystemOptionsBuilder{}

	productVersion := product.CurrentProductVersion()

	optProv.AppLabel(*o.APIManagerSpec.AppLabel)
	optProv.AmpRelease(string(productVersion))
	optProv.ApicastRegistryURL(*o.APIManagerSpec.Apicast.RegistryURL)
	optProv.TenantName(*o.APIManagerSpec.TenantName)
	optProv.WildcardDomain(o.APIManagerSpec.WildcardDomain)

	if o.APIManagerSpec.System.FileStorageSpec.PVC == nil {
		optProv.StorageClassName(nil)
	} else {
		optProv.StorageClassName(o.APIManagerSpec.System.FileStorageSpec.PVC.StorageClassName)
	}

	err := o.setSecretBasedOptions(&optProv)
	if err != nil {
		return nil, err
	}

	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create System Options - %s", err)
	}
	return res, nil
}

func (o *OperatorSystemOptionsProvider) setSecretBasedOptions(builder *component.SystemOptionsBuilder) error {
	err := o.setSystemMemcachedOptions(builder)
	if err != nil {
		return fmt.Errorf("unable to create System Memcached secret options - %s", err)
	}

	err = o.setSystemRecaptchaOptions(builder)
	if err != nil {
		return fmt.Errorf("unable to create System Recaptcha secret options - %s", err)
	}

	err = o.setSystemEventHookOptions(builder)
	if err != nil {
		return fmt.Errorf("unable to create System Event Hooks secret options - %s", err)
	}

	err = o.setSystemRedisOptions(builder)
	if err != nil {
		return fmt.Errorf("unable to create System Redis secret options - %s", err)
	}

	err = o.setSystemAppOptions(builder)
	if err != nil {
		return fmt.Errorf("unable to create System App secret options - %s", err)
	}

	err = o.setSystemSeedOptions(builder)
	if err != nil {
		return fmt.Errorf("unable to create System Seed secret options - %s", err)
	}

	err = o.setSystemMasterApicastOptions(builder)
	if err != nil {
		return fmt.Errorf("unable to create System Master Apicast secret options - %s", err)
	}

	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemMemcachedOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := getSecret(component.SystemSecretSystemMemcachedSecretName, o.Namespace, o.Client)

	if err != nil {
		if errors.IsNotFound(err) {
			// Do nothing because there are no required options for related to the Memcached servers secret
		} else {
			return err
		}
	} else {
		secretData := currSecret.Data
		var result *string
		result = getSecretDataValue(secretData, component.SystemSecretSystemMemcachedServersFieldName)
		if result != nil {
			builder.MemcachedServers(*result)
		}
	}
	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemRecaptchaOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := getSecret(component.SystemSecretSystemRecaptchaSecretName, o.Namespace, o.Client)
	defaultRecaptchaPublicKey := ""
	defaultRecaptchaPrivateKey := ""

	if err != nil {
		if errors.IsNotFound(err) {
			builder.RecaptchaPublicKey(defaultRecaptchaPublicKey)
			builder.RecaptchaPrivateKey(defaultRecaptchaPrivateKey)
		} else {
			return err
		}
	} else {
		secretData := currSecret.Data
		builder.RecaptchaPublicKey(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemRecaptchaPublicKeyFieldName, defaultRecaptchaPublicKey))
		builder.RecaptchaPublicKey(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemRecaptchaPrivateKeyFieldName, defaultRecaptchaPrivateKey))
	}
	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemEventHookOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := getSecret(component.SystemSecretSystemEventsHookSecretName, o.Namespace, o.Client)
	defaultBackendSharedSecret := oprand.String(8)

	if err != nil {
		if errors.IsNotFound(err) {
			builder.BackendSharedSecret(defaultBackendSharedSecret)
		} else {
			return err
		}
	} else {
		secretData := currSecret.Data
		builder.BackendSharedSecret(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemEventsHookPasswordFieldName, defaultBackendSharedSecret))
		var result *string
		result = getSecretDataValue(secretData, component.SystemSecretSystemEventsHookURLFieldName)
		if result != nil {
			builder.EventHooksURL(*result)
		}
	}
	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemRedisOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := getSecret(component.SystemSecretSystemRedisSecretName, o.Namespace, o.Client)

	if err != nil {
		if errors.IsNotFound(err) {
			// Do nothing because there are no required options for related to the System's Redis servers secret
		} else {
			return err
		}
	} else {
		secretData := currSecret.Data
		var result *string
		result = getSecretDataValue(secretData, component.SystemSecretSystemRedisURLFieldName)
		if result != nil {
			builder.RedisURL(*result)
		}

		result = getSecretDataValue(secretData, component.SystemSecretSystemRedisMessageBusRedisURLFieldName)
		if result != nil {
			builder.MessageBusRedisURL(*result)
		}

		result = getSecretDataValue(secretData, component.SystemSecretSystemRedisNamespace)
		if result != nil {
			builder.RedisNamespace(*result)
		}

		result = getSecretDataValue(secretData, component.SystemSecretSystemRedisMessageBusRedisNamespace)
		if result != nil {
			builder.MessageBusRedisNamespace(*result)
		}

	}
	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemAppOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := getSecret(component.SystemSecretSystemAppSecretName, o.Namespace, o.Client)
	// TODO is not exactly what we were generating
	// in OpenShift templates. We were generating
	// '[a-f0-9]{128}' . Ask system if there's some reason
	// for that and if we can change it. If must be that range
	// then we should create another function to generate
	// hexadecimal lowercase string output
	defaultSecretKeyBase := oprand.String(128)

	if err != nil {
		if errors.IsNotFound(err) {
			// Do nothing because there are no required options for related to the Memcached servers secret
			builder.AppSecretKeyBase(defaultSecretKeyBase)
		} else {
			return err
		}
	} else {
		secretData := currSecret.Data
		builder.AppSecretKeyBase(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemAppSecretKeyBaseFieldName, defaultSecretKeyBase))

	}
	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemSeedOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := getSecret(component.SystemSecretSystemSeedSecretName, o.Namespace, o.Client)
	defaultMasterDomain := "master"
	defaultMasterUser := "master"
	defaultMasterPassword := oprand.String(8)
	defaultAdminUser := "admin"
	defaultAdminPassword := oprand.String(8)
	defaultAdminAccessToken := oprand.String(16)
	//defaultSeedTenantName := *o.APIManagerSpec.TenantName // Fix this. Why is TENANT_NAME a secret in system seed? Does not seem a secret so should be directly gathered from the value
	defaultMasterAccessToken := oprand.String(8)
	if err != nil {
		if errors.IsNotFound(err) {
			// Do nothing because there are no required options for related to the Memcached servers secret
			builder.MasterName(defaultMasterDomain)
			builder.MasterUsername(defaultMasterUser)
			builder.MasterPassword(defaultMasterPassword)
			builder.AdminUsername(defaultAdminUser)
			builder.AdminPassword(defaultAdminPassword)
			builder.AdminAccessToken(defaultAdminAccessToken)
			builder.MasterAccessToken(defaultMasterAccessToken)
		} else {
			return err
		}
	} else {
		secretData := currSecret.Data
		builder.MasterName(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedMasterDomainFieldName, defaultMasterDomain))
		builder.MasterUsername(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedMasterUserFieldName, defaultMasterUser))
		builder.MasterPassword(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedMasterPasswordFieldName, defaultMasterPassword))
		builder.AdminUsername(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedAdminUserFieldName, defaultAdminUser))
		builder.AdminPassword(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedAdminPasswordFieldName, defaultAdminPassword))
		builder.AdminAccessToken(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedAdminAccessTokenFieldName, defaultAdminAccessToken))
		builder.MasterAccessToken(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedMasterAccessTokenFieldName, defaultMasterAccessToken))

		result := getSecretDataValue(secretData, component.SystemSecretSystemSeedAdminEmailFieldName)
		if result != nil {
			builder.AdminEmail(*result)
		}
	}
	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemMasterApicastOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := getSecret(component.SystemSecretSystemMasterApicastSecretName, o.Namespace, o.Client)
	defaultSystemMasterApicastAccessToken := oprand.String(8)

	if err != nil {
		if errors.IsNotFound(err) {
			// Do nothing because there are no required options for related to the secret
			builder.ApicastAccessToken(defaultSystemMasterApicastAccessToken)
		} else {
			return err
		}
	} else {
		secretData := currSecret.Data
		// TODO we do not reconcile ProxyConfigEndpoint nor BaseURL fields because they are dependant on the TenantName
		builder.ApicastAccessToken(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemMasterApicastAccessToken, defaultSystemMasterApicastAccessToken))
	}
	return nil
}
