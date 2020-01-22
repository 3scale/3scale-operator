package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	oprand "github.com/3scale/3scale-operator/pkg/crypto/rand"
	"github.com/3scale/3scale-operator/pkg/helper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	err = o.setSystemSMTPOptions(builder)
	if err != nil {
		return fmt.Errorf("unable to create System SMTP secret options - %s", err)
	}

	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemMemcachedOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := helper.GetSecret(component.SystemSecretSystemMemcachedSecretName, o.Namespace, o.Client)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	secretData := currSecret.Data
	builder.MemcachedServers(helper.GetSecretDataValue(secretData, component.SystemSecretSystemMemcachedServersFieldName))
	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemRecaptchaOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := helper.GetSecret(component.SystemSecretSystemRecaptchaSecretName, o.Namespace, o.Client)
	defaultRecaptchaPublicKey := ""
	defaultRecaptchaPrivateKey := ""

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	secretData := currSecret.Data
	builder.RecaptchaPublicKey(helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemRecaptchaPublicKeyFieldName, defaultRecaptchaPublicKey))
	builder.RecaptchaPublicKey(helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemRecaptchaPrivateKeyFieldName, defaultRecaptchaPrivateKey))

	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemEventHookOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := helper.GetSecret(component.SystemSecretSystemEventsHookSecretName, o.Namespace, o.Client)
	defaultBackendSharedSecret := oprand.String(8)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	secretData := currSecret.Data
	builder.BackendSharedSecret(helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemEventsHookPasswordFieldName, defaultBackendSharedSecret))
	builder.EventHooksURL(helper.GetSecretDataValue(secretData, component.SystemSecretSystemEventsHookURLFieldName))
	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemRedisOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := helper.GetSecret(component.SystemSecretSystemRedisSecretName, o.Namespace, o.Client)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	secretData := currSecret.Data
	builder.RedisURL(helper.GetSecretDataValue(secretData, component.SystemSecretSystemRedisURLFieldName))
	builder.RedisSentinelHosts(helper.GetSecretDataValue(secretData, component.SystemSecretSystemRedisSentinelHosts))
	builder.RedisSentinelRole(helper.GetSecretDataValue(secretData, component.SystemSecretSystemRedisSentinelRole))
	builder.MessageBusRedisSentinelHosts(helper.GetSecretDataValue(secretData, component.SystemSecretSystemRedisMessageBusSentinelHosts))
	builder.MessageBusRedisSentinelRole(helper.GetSecretDataValue(secretData, component.SystemSecretSystemRedisMessageBusSentinelRole))
	builder.MessageBusRedisURL(helper.GetSecretDataValue(secretData, component.SystemSecretSystemRedisMessageBusRedisURLFieldName))
	builder.RedisNamespace(helper.GetSecretDataValue(secretData, component.SystemSecretSystemRedisNamespace))
	builder.MessageBusRedisNamespace(helper.GetSecretDataValue(secretData, component.SystemSecretSystemRedisMessageBusRedisNamespace))

	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemAppOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := helper.GetSecret(component.SystemSecretSystemAppSecretName, o.Namespace, o.Client)
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
	builder.AppSecretKeyBase(helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemAppSecretKeyBaseFieldName, defaultSecretKeyBase))

	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemSeedOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := helper.GetSecret(component.SystemSecretSystemSeedSecretName, o.Namespace, o.Client)
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
	builder.MasterName(helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedMasterDomainFieldName, defaultMasterDomain))
	builder.MasterUsername(helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedMasterUserFieldName, defaultMasterUser))
	builder.MasterPassword(helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedMasterPasswordFieldName, defaultMasterPassword))
	builder.AdminUsername(helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedAdminUserFieldName, defaultAdminUser))
	builder.AdminPassword(helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedAdminPasswordFieldName, defaultAdminPassword))
	builder.AdminAccessToken(helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedAdminAccessTokenFieldName, defaultAdminAccessToken))
	builder.MasterAccessToken(helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemSeedMasterAccessTokenFieldName, defaultMasterAccessToken))
	builder.AdminEmail(helper.GetSecretDataValue(secretData, component.SystemSecretSystemSeedAdminEmailFieldName))

	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemMasterApicastOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := helper.GetSecret(component.SystemSecretSystemMasterApicastSecretName, o.Namespace, o.Client)
	defaultSystemMasterApicastAccessToken := oprand.String(8)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	secretData := currSecret.Data
	// TODO we do not reconcile ProxyConfigEndpoint nor BaseURL fields because they are dependant on the TenantName
	builder.ApicastAccessToken(helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemMasterApicastAccessToken, defaultSystemMasterApicastAccessToken))

	return nil
}

func (o *OperatorSystemOptionsProvider) setSystemSMTPOptions(builder *component.SystemOptionsBuilder) error {
	currSecret, err := helper.GetSecret(component.SystemSecretSystemSMTPSecretName, o.Namespace, o.Client)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	secretData := currSecret.Data

	smtpSecretOptions := component.SystemSMTPSecretOptions{
		Address:           helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemSMTPAddressFieldName, ""),
		Authentication:    helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemSMTPAuthenticationFieldName, ""),
		Domain:            helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemSMTPDomainFieldName, ""),
		OpenSSLVerifyMode: helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemSMTPOpenSSLVerifyModeFieldName, ""),
		Password:          helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemSMTPPasswordFieldName, ""),
		Port:              helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemSMTPPortFieldName, ""),
		Username:          helper.GetSecretDataValueOrDefault(secretData, component.SystemSecretSystemSMTPUserNameFieldName, ""),
	}

	builder.SystemSMTPSecretOptions(smtpSecretOptions)

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
	if o.APIManagerSpec.System != nil &&
		o.APIManagerSpec.System.FileStorageSpec != nil &&
		o.APIManagerSpec.System.FileStorageSpec.S3 != nil {
		s3FileStorageSpec := o.APIManagerSpec.System.FileStorageSpec.S3
		b.S3FileStorageOptions(component.S3FileStorageOptions{
			ConfigurationSecretName: s3FileStorageSpec.ConfigurationSecretRef.Name,
		})
	} else {
		// default to PVC
		var storageClass *string
		if o.APIManagerSpec.System != nil &&
			o.APIManagerSpec.System.FileStorageSpec != nil &&
			o.APIManagerSpec.System.FileStorageSpec.PVC != nil {
			storageClass = o.APIManagerSpec.System.FileStorageSpec.PVC.StorageClassName
		}

		b.PVCFileStorageOptions(component.PVCFileStorageOptions{
			StorageClass: storageClass,
		})
	}
}

func (o *OperatorSystemOptionsProvider) setReplicas(sob *component.SystemOptionsBuilder) {
	sob.AppReplicas(int32(*o.APIManagerSpec.System.AppSpec.Replicas))
	sob.SidekiqReplicas(int32(*o.APIManagerSpec.System.SidekiqSpec.Replicas))
}

func System(cr *appsv1alpha1.APIManager, client client.Client) (*component.System, error) {
	optsProvider := OperatorSystemOptionsProvider{APIManagerSpec: &cr.Spec, Namespace: cr.Namespace, Client: client}
	opts, err := optsProvider.GetSystemOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystem(opts), nil
}
