package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	oprand "github.com/3scale/3scale-operator/pkg/crypto/rand"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (o *OperatorSystemOptionsProvider) GetSystemOptions() (*component.SystemOptions, error) {
	optProv := component.SystemOptionsBuilder{}

	optProv.AppLabel(*o.APIManagerSpec.AppLabel)
	optProv.AmpRelease(product.ThreescaleRelease)
	optProv.ApicastRegistryURL(*o.APIManagerSpec.Apicast.RegistryURL)
	optProv.TenantName(*o.APIManagerSpec.TenantName)
	optProv.WildcardDomain(o.APIManagerSpec.WildcardDomain)

	err := o.setSecretBasedOptions(&optProv)
	if err != nil {
		return nil, err
	}

	o.setResourceRequirementsOptions(&optProv)
	o.setFileStorageOptions(&optProv)
	o.setReplicas(&optProv)

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

	if o.APIManagerSpec.System.FileStorageSpec != nil && o.APIManagerSpec.System.FileStorageSpec.S3 != nil {
		err = o.setAWSSecretOptions(builder)
		if err != nil {
			return fmt.Errorf("unable to create AWS S3 secret options - %s", err)
		}
	}

	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemMemcachedOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := getSecret(component.SystemSecretSystemMemcachedSecretName, o.Namespace, o.Client)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	secretData := currSecret.Data
	builder.MemcachedServers(getSecretDataValue(secretData, component.SystemSecretSystemMemcachedServersFieldName))
	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemRecaptchaOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := getSecret(component.SystemSecretSystemRecaptchaSecretName, o.Namespace, o.Client)
	defaultRecaptchaPublicKey := ""
	defaultRecaptchaPrivateKey := ""

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	secretData := currSecret.Data
	builder.RecaptchaPublicKey(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemRecaptchaPublicKeyFieldName, defaultRecaptchaPublicKey))
	builder.RecaptchaPublicKey(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemRecaptchaPrivateKeyFieldName, defaultRecaptchaPrivateKey))

	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemEventHookOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := getSecret(component.SystemSecretSystemEventsHookSecretName, o.Namespace, o.Client)
	defaultBackendSharedSecret := oprand.String(8)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	secretData := currSecret.Data
	builder.BackendSharedSecret(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemEventsHookPasswordFieldName, defaultBackendSharedSecret))
	builder.EventHooksURL(getSecretDataValue(secretData, component.SystemSecretSystemEventsHookURLFieldName))
	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemRedisOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := getSecret(component.SystemSecretSystemRedisSecretName, o.Namespace, o.Client)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	secretData := currSecret.Data
	builder.RedisURL(getSecretDataValue(secretData, component.SystemSecretSystemRedisURLFieldName))
	builder.RedisSentinelHosts(getSecretDataValue(secretData, component.SystemSecretSystemRedisSentinelHosts))
	builder.RedisSentinelRole(getSecretDataValue(secretData, component.SystemSecretSystemRedisSentinelRole))
	builder.MessageBusRedisSentinelHosts(getSecretDataValue(secretData, component.SystemSecretSystemRedisMessageBusSentinelHosts))
	builder.MessageBusRedisSentinelRole(getSecretDataValue(secretData, component.SystemSecretSystemRedisMessageBusSentinelRole))
	builder.MessageBusRedisURL(getSecretDataValue(secretData, component.SystemSecretSystemRedisMessageBusRedisURLFieldName))
	builder.RedisNamespace(getSecretDataValue(secretData, component.SystemSecretSystemRedisNamespace))
	builder.MessageBusRedisNamespace(getSecretDataValue(secretData, component.SystemSecretSystemRedisMessageBusRedisNamespace))

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

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	secretData := currSecret.Data
	builder.AppSecretKeyBase(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemAppSecretKeyBaseFieldName, defaultSecretKeyBase))

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
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	secretData := currSecret.Data
	builder.MasterName(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedMasterDomainFieldName, defaultMasterDomain))
	builder.MasterUsername(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedMasterUserFieldName, defaultMasterUser))
	builder.MasterPassword(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedMasterPasswordFieldName, defaultMasterPassword))
	builder.AdminUsername(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedAdminUserFieldName, defaultAdminUser))
	builder.AdminPassword(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedAdminPasswordFieldName, defaultAdminPassword))
	builder.AdminAccessToken(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedAdminAccessTokenFieldName, defaultAdminAccessToken))
	builder.MasterAccessToken(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedMasterAccessTokenFieldName, defaultMasterAccessToken))
	builder.AdminEmail(getSecretDataValue(secretData, component.SystemSecretSystemSeedAdminEmailFieldName))

	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemMasterApicastOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := getSecret(component.SystemSecretSystemMasterApicastSecretName, o.Namespace, o.Client)
	defaultSystemMasterApicastAccessToken := oprand.String(8)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	secretData := currSecret.Data
	// TODO we do not reconcile ProxyConfigEndpoint nor BaseURL fields because they are dependant on the TenantName
	builder.ApicastAccessToken(getSecretDataValueOrDefault(secretData, component.SystemSecretSystemMasterApicastAccessToken, defaultSystemMasterApicastAccessToken))

	return nil
}

func (o *OperatorSystemOptionsProvider) setResourceRequirementsOptions(b *component.SystemOptionsBuilder) {
	if !*o.APIManagerSpec.ResourceRequirementsEnabled {
		b.AppMasterContainerResourceRequirements(v1.ResourceRequirements{})
		b.AppProviderContainerResourceRequirements(v1.ResourceRequirements{})
		b.AppDeveloperContainerResourceRequirements(v1.ResourceRequirements{})
		b.SidekiqContainerResourceRequirements(v1.ResourceRequirements{})
		b.SphinxContainerResourceRequirements(v1.ResourceRequirements{})
	}
}

func (o *OperatorSystemOptionsProvider) setFileStorageOptions(b *component.SystemOptionsBuilder) {
	if o.APIManagerSpec.System.FileStorageSpec.PVC != nil {
		b.PVCFileStorageOptions(component.PVCFileStorageOptions{
			StorageClass: o.APIManagerSpec.System.FileStorageSpec.PVC.StorageClassName,
		})
	}

	if o.APIManagerSpec.System.FileStorageSpec.S3 != nil {
		s3FileStorageSpec := o.APIManagerSpec.System.FileStorageSpec.S3
		b.S3FileStorageOptions(component.S3FileStorageOptions{
			AWSAccessKeyId:       "",
			AWSSecretAccessKey:   "",
			AWSRegion:            s3FileStorageSpec.AWSRegion,
			AWSBucket:            s3FileStorageSpec.AWSBucket,
			AWSCredentialsSecret: s3FileStorageSpec.AWSCredentials.Name,
		})
	}
}

func (o *OperatorSystemOptionsProvider) setAWSSecretOptions(sob *component.SystemOptionsBuilder) error {
	awsCredentialsSecretName := o.APIManagerSpec.System.FileStorageSpec.S3.AWSCredentials.Name
	currSecret, err := getSecret(awsCredentialsSecretName, o.Namespace, o.Client)
	if err != nil {
		return err
	}

	// If a field of a secret already exists in the deployed secret then
	// We do not modify it. Otherwise we set a default value
	secretData := currSecret.Data
	var result *string
	result = getSecretDataValue(secretData, component.S3SecretAWSAccessKeyIdFieldName)
	if result == nil {
		return fmt.Errorf("Secret field '%s' is required in secret '%s'", component.S3SecretAWSAccessKeyIdFieldName, awsCredentialsSecretName)
	}
	awsAccessKeyID := *result

	result = getSecretDataValue(secretData, component.S3SecretAWSSecretAccessKeyFieldName)
	if result == nil {
		return fmt.Errorf("Secret field '%s' is required in secret '%s'", component.S3SecretAWSSecretAccessKeyFieldName, awsCredentialsSecretName)
	}
	awsSecretAccessKeyID := *result

	s3FileStorageSpec := o.APIManagerSpec.System.FileStorageSpec.S3
	sob.S3FileStorageOptions(component.S3FileStorageOptions{
		AWSAccessKeyId:       awsAccessKeyID,
		AWSSecretAccessKey:   awsSecretAccessKeyID,
		AWSRegion:            s3FileStorageSpec.AWSRegion,
		AWSBucket:            s3FileStorageSpec.AWSBucket,
		AWSCredentialsSecret: s3FileStorageSpec.AWSCredentials.Name,
	})

	return nil
}

func (o *OperatorSystemOptionsProvider) setReplicas(sob *component.SystemOptionsBuilder) {
	sob.AppReplicas(int32(*o.APIManagerSpec.System.AppSpec.Replicas))
	sob.SidekiqReplicas(int32(*o.APIManagerSpec.System.SidekiqSpec.Replicas))
}
