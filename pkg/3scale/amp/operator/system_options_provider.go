package operator

import (
	"fmt"
	"net/url"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/helper"
)

type SystemOptionsProvider struct {
	apimanager   *appsv1alpha1.APIManager
	namespace    string
	client       client.Client
	options      *component.SystemOptions
	secretSource *helper.SecretSource
}

func NewSystemOptionsProvider(apimanager *appsv1alpha1.APIManager, namespace string, client client.Client) *SystemOptionsProvider {
	return &SystemOptionsProvider{
		apimanager:   apimanager,
		namespace:    namespace,
		client:       client,
		options:      component.NewSystemOptions(),
		secretSource: helper.NewSecretSource(client, namespace),
	}
}

func (s *SystemOptionsProvider) GetSystemOptions() (*component.SystemOptions, error) {
	s.options.ImageTag = product.ThreescaleRelease
	s.options.ApicastRegistryURL = *s.apimanager.Spec.Apicast.RegistryURL
	s.options.TenantName = *s.apimanager.Spec.TenantName
	s.options.WildcardDomain = s.apimanager.Spec.WildcardDomain

	s.options.CommonLabels = s.commonLabels()
	s.options.CommonAppLabels = s.commonAppLabels()
	s.options.AppPodTemplateLabels = s.appPodTemplateLabels()
	s.options.CommonSidekiqLabels = s.commonSidekiqLabels()
	s.options.SidekiqPodTemplateLabels = s.sidekiqPodTemplateLabels()
	s.options.ProviderUILabels = s.providerUILabels()
	s.options.MasterUILabels = s.masterUILabels()
	s.options.DeveloperUILabels = s.developerUILabels()
	s.options.SphinxLabels = s.sphinxLabels()
	s.options.SphinxPodTemplateLabels = s.sphinxPodTemplateLabels()
	s.options.MemcachedLabels = s.memcachedLabels()
	s.options.SMTPLabels = s.smtpLabels()

	err := s.setSecretBasedOptions()
	if err != nil {
		return nil, fmt.Errorf("GetSystemOptions reading secret options: %w", err)
	}

	s.setResourceRequirementsOptions()
	s.setNodeAffinityAndTolerationsOptions()
	err = s.setFileStorageOptions()
	if err != nil {
		return nil, err
	}
	s.setReplicas()

	s.options.SideKiqMetrics = true
	s.options.AppMetrics = true
	s.options.IncludeOracleOptionalSettings = true

	s.options.Namespace = s.namespace

	err = s.options.Validate()
	if err != nil {
		return nil, fmt.Errorf("GetSystemOptions validating: %w", err)
	}

	return s.options, nil
}

func (s *SystemOptionsProvider) setSecretBasedOptions() error {
	err := s.setSystemMemcachedOptions()
	if err != nil {
		return fmt.Errorf("unable to create System Memcached secret options - %s", err)
	}

	err = s.setSystemRecaptchaOptions()
	if err != nil {
		return fmt.Errorf("unable to create System Recaptcha secret options - %s", err)
	}

	err = s.setSystemEventHookOptions()
	if err != nil {
		return fmt.Errorf("unable to create System Event Hooks secret options - %s", err)
	}

	err = s.setSystemAppOptions()
	if err != nil {
		return fmt.Errorf("unable to create System App secret options - %s", err)
	}

	err = s.setSystemSeedOptions()
	if err != nil {
		return fmt.Errorf("unable to create System Seed secret options - %s", err)
	}

	err = s.setSystemMasterApicastOptions()
	if err != nil {
		return fmt.Errorf("unable to create System Master Apicast secret options - %s", err)
	}

	err = s.setSystemSMTPOptions()
	if err != nil {
		return fmt.Errorf("unable to create System SMTP secret options - %s", err)
	}

	err = s.setBackendOptions()
	if err != nil {
		return err
	}

	return nil
}

func (s *SystemOptionsProvider) setSystemMemcachedOptions() error {
	val, err := s.secretSource.FieldValue(
		component.SystemSecretSystemMemcachedSecretName,
		component.SystemSecretSystemMemcachedServersFieldName,
		component.DefaultMemcachedServers())
	if err != nil {
		return err
	}
	s.options.MemcachedServers = val

	return nil
}

func (s *SystemOptionsProvider) setSystemRecaptchaOptions() error {
	recaptchaPublicKey, err := s.secretSource.FieldValue(
		component.SystemSecretSystemRecaptchaSecretName,
		component.SystemSecretSystemRecaptchaPublicKeyFieldName,
		component.DefaultRecaptchaPublickey())
	if err != nil {
		return err
	}
	s.options.RecaptchaPublicKey = &recaptchaPublicKey

	recaptchaPrivateKey, err := s.secretSource.FieldValue(
		component.SystemSecretSystemRecaptchaSecretName,
		component.SystemSecretSystemRecaptchaPrivateKeyFieldName,
		component.DefaultRecaptchaPrivatekey())
	if err != nil {
		return err
	}
	s.options.RecaptchaPrivateKey = &recaptchaPrivateKey

	return nil
}

func (s *SystemOptionsProvider) setSystemEventHookOptions() error {
	val, err := s.secretSource.FieldValue(
		component.SystemSecretSystemEventsHookSecretName,
		component.SystemSecretSystemEventsHookPasswordFieldName,
		component.DefaultBackendSharedSecret())
	if err != nil {
		return err
	}
	s.options.BackendSharedSecret = val

	val, err = s.secretSource.FieldValue(
		component.SystemSecretSystemEventsHookSecretName,
		component.SystemSecretSystemEventsHookURLFieldName,
		component.DefaultEventHooksURL())
	if err != nil {
		return err
	}
	s.options.EventHooksURL = val

	return nil
}

func (s *SystemOptionsProvider) setSystemAppOptions() error {
	val, err := s.secretSource.FieldValue(
		component.SystemSecretSystemAppSecretName,
		component.SystemSecretSystemAppSecretKeyBaseFieldName,
		component.DefaultSystemAppSecretKeyBase())
	if err != nil {
		return err
	}
	s.options.AppSecretKeyBase = val

	val, err = s.secretSource.FieldValue(
		component.SystemSecretSystemAppSecretName,
		component.SystemSecretSystemAppUserSessionTTLFieldName,
		component.DefaultUserSessionTTL())
	if err != nil {
		return err
	}
	s.options.UserSessionTTL = &val

	return nil
}

func (s *SystemOptionsProvider) setSystemSeedOptions() error {
	cases := []struct {
		field       *string
		secretName  string
		secretField string
		defValue    string
	}{
		{
			&s.options.MasterName,
			component.SystemSecretSystemSeedSecretName,
			component.SystemSecretSystemSeedMasterDomainFieldName,
			component.DefaultSystemMasterName(),
		},
		{
			&s.options.MasterUsername,
			component.SystemSecretSystemSeedSecretName,
			component.SystemSecretSystemSeedMasterUserFieldName,
			component.DefaultSystemMasterUsername(),
		},
		{
			&s.options.MasterPassword,
			component.SystemSecretSystemSeedSecretName,
			component.SystemSecretSystemSeedMasterPasswordFieldName,
			component.DefaultSystemMasterPassword(),
		},
		{
			&s.options.AdminUsername,
			component.SystemSecretSystemSeedSecretName,
			component.SystemSecretSystemSeedAdminUserFieldName,
			component.DefaultSystemAdminUsername(),
		},
		{
			&s.options.AdminPassword,
			component.SystemSecretSystemSeedSecretName,
			component.SystemSecretSystemSeedAdminPasswordFieldName,
			component.DefaultSystemAdminPassword(),
		},
		{
			&s.options.AdminAccessToken,
			component.SystemSecretSystemSeedSecretName,
			component.SystemSecretSystemSeedAdminAccessTokenFieldName,
			component.DefaultSystemAdminAccessToken(),
		},
		{
			&s.options.MasterAccessToken,
			component.SystemSecretSystemSeedSecretName,
			component.SystemSecretSystemSeedMasterAccessTokenFieldName,
			component.DefaultSystemMasterAccessToken(),
		},
	}

	for _, option := range cases {
		val, err := s.secretSource.FieldValue(option.secretName, option.secretField, option.defValue)
		if err != nil {
			return err
		}
		*option.field = val
	}

	adminEmail, err := s.secretSource.FieldValue(
		component.SystemSecretSystemSeedSecretName,
		component.SystemSecretSystemSeedAdminEmailFieldName,
		component.DefaultSystemAdminEmail())
	if err != nil {
		return err
	}
	s.options.AdminEmail = &adminEmail

	return nil
}

func (s *SystemOptionsProvider) setSystemMasterApicastOptions() error {
	val, err := s.secretSource.FieldValue(
		component.SystemSecretSystemMasterApicastSecretName,
		component.SystemSecretSystemMasterApicastAccessToken,
		component.DefaultSystemMasterApicastAccessToken())
	if err != nil {
		return err
	}
	s.options.ApicastAccessToken = val

	val, err = s.secretSource.FieldValue(
		component.SystemSecretSystemMasterApicastSecretName,
		component.SystemSecretSystemMasterApicastProxyConfigsEndpointFieldName,
		component.DefaultApicastSystemMasterProxyConfigEndpoint(s.options.ApicastAccessToken))
	if err != nil {
		return err
	}
	s.options.ApicastSystemMasterProxyConfigEndpoint = val

	return nil
}

func (s *SystemOptionsProvider) setSystemSMTPOptions() error {
	smtpSecretOptions := component.SystemSMTPSecretOptions{}
	cases := []struct {
		field       **string
		secretName  string
		secretField string
		defValue    string
	}{
		{
			&smtpSecretOptions.Address,
			component.SystemSecretSystemSMTPSecretName,
			component.SystemSecretSystemSMTPAddressFieldName,
			component.DefaultSystemSMTPAddress(),
		},
		{
			&smtpSecretOptions.Authentication,
			component.SystemSecretSystemSMTPSecretName,
			component.SystemSecretSystemSMTPAuthenticationFieldName,
			component.DefaultSystemSMTPAuthentication(),
		},
		{
			&smtpSecretOptions.Domain,
			component.SystemSecretSystemSMTPSecretName,
			component.SystemSecretSystemSMTPDomainFieldName,
			component.DefaultSystemSMTPDomain(),
		},
		{
			&smtpSecretOptions.OpenSSLVerifyMode,
			component.SystemSecretSystemSMTPSecretName,
			component.SystemSecretSystemSMTPOpenSSLVerifyModeFieldName,
			component.DefaultSystemSMTPOpenSSLVerifyMode(),
		},
		{
			&smtpSecretOptions.Password,
			component.SystemSecretSystemSMTPSecretName,
			component.SystemSecretSystemSMTPPasswordFieldName,
			component.DefaultSystemSMTPPassword(),
		},
		{
			&smtpSecretOptions.Port,
			component.SystemSecretSystemSMTPSecretName,
			component.SystemSecretSystemSMTPPortFieldName,
			component.DefaultSystemSMTPPort(),
		},
		{
			&smtpSecretOptions.Username,
			component.SystemSecretSystemSMTPSecretName,
			component.SystemSecretSystemSMTPUserNameFieldName,
			component.DefaultSystemSMTPUsername(),
		},
		{
			&smtpSecretOptions.FromAddress,
			component.SystemSecretSystemSMTPSecretName,
			component.SystemSecretSystemSMTPFromAddressFieldName,
			component.DefaultSystemSMTPFromAddress(),
		},
	}

	for _, option := range cases {
		val, err := s.secretSource.FieldValue(option.secretName, option.secretField, option.defValue)
		if err != nil {
			return err
		}
		*option.field = &val
	}

	s.options.SmtpSecretOptions = smtpSecretOptions
	return nil
}

func (s *SystemOptionsProvider) setBackendOptions() error {
	rawURL, err := s.secretSource.FieldValue(
		component.BackendSecretBackendListenerSecretName,
		component.BackendSecretBackendListenerServiceEndpointFieldName,
		component.DefaultBackendServiceEndpoint())
	if err != nil {
		return err
	}

	urlObj, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("'%s' field of '%s' secret must have 'scheme://user:password@host/path' format", component.BackendSecretBackendListenerServiceEndpointFieldName, component.BackendSecretBackendListenerSecretName)
	}

	if urlObj.Path == "" {
		urlObj.Path = "/internal/"
	}

	s.options.BackendServiceEndpoint = urlObj.String()

	return nil
}

func (s *SystemOptionsProvider) setResourceRequirementsOptions() {
	if *s.apimanager.Spec.ResourceRequirementsEnabled {
		s.options.AppMasterContainerResourceRequirements = component.DefaultAppMasterContainerResourceRequirements()
		s.options.AppProviderContainerResourceRequirements = component.DefaultAppProviderContainerResourceRequirements()
		s.options.AppDeveloperContainerResourceRequirements = component.DefaultAppDeveloperContainerResourceRequirements()
		s.options.SidekiqContainerResourceRequirements = component.DefaultSidekiqContainerResourceRequirements()
		s.options.SphinxContainerResourceRequirements = component.DefaultSphinxContainerResourceRequirements()
	} else {
		s.options.AppMasterContainerResourceRequirements = &v1.ResourceRequirements{}
		s.options.AppProviderContainerResourceRequirements = &v1.ResourceRequirements{}
		s.options.AppDeveloperContainerResourceRequirements = &v1.ResourceRequirements{}
		s.options.SidekiqContainerResourceRequirements = &v1.ResourceRequirements{}
		s.options.SphinxContainerResourceRequirements = &v1.ResourceRequirements{}
	}

	// DeploymentConfig-level ResourceRequirements CR fields have priority over
	// spec.resourceRequirementsEnabled, overwriting that setting when they are
	// defined
	if s.apimanager.Spec.System.AppSpec.MasterContainerResources != nil {
		s.options.AppMasterContainerResourceRequirements = s.apimanager.Spec.System.AppSpec.MasterContainerResources
	}
	if s.apimanager.Spec.System.AppSpec.ProviderContainerResources != nil {
		s.options.AppProviderContainerResourceRequirements = s.apimanager.Spec.System.AppSpec.ProviderContainerResources
	}
	if s.apimanager.Spec.System.AppSpec.DeveloperContainerResources != nil {
		s.options.AppDeveloperContainerResourceRequirements = s.apimanager.Spec.System.AppSpec.DeveloperContainerResources
	}
	if s.apimanager.Spec.System.SidekiqSpec.Resources != nil {
		s.options.SidekiqContainerResourceRequirements = s.apimanager.Spec.System.SidekiqSpec.Resources
	}
	if s.apimanager.Spec.System.SphinxSpec.Resources != nil {
		s.options.SphinxContainerResourceRequirements = s.apimanager.Spec.System.SphinxSpec.Resources
	}
}

func (s *SystemOptionsProvider) setNodeAffinityAndTolerationsOptions() {
	s.options.AppAffinity = s.apimanager.Spec.System.AppSpec.Affinity
	s.options.AppTolerations = s.apimanager.Spec.System.AppSpec.Tolerations
	s.options.SidekiqAffinity = s.apimanager.Spec.System.SidekiqSpec.Affinity
	s.options.SidekiqTolerations = s.apimanager.Spec.System.SidekiqSpec.Tolerations
	s.options.SphinxAffinity = s.apimanager.Spec.System.SphinxSpec.Affinity
	s.options.SphinxTolerations = s.apimanager.Spec.System.SphinxSpec.Tolerations
}

func (s *SystemOptionsProvider) setFileStorageOptions() error {
	if s.apimanager.IsS3STSEnabled() {
		s.options.S3FileStorageOptions = &component.S3FileStorageOptions{
			ConfigurationSecretName: s.apimanager.Spec.System.FileStorageSpec.S3.ConfigurationSecretRef.Name,
			STSEnabled:              true,
			STSAudience:             "openshift", // default value when "audience" is not specified
		}

		if s.apimanager.Spec.System.FileStorageSpec.S3.STS.Audience != nil {
			s.options.S3FileStorageOptions.STSAudience = *s.apimanager.Spec.System.FileStorageSpec.S3.STS.Audience
		}

		// Validate it exists
		_, err := s.secretSource.RequiredFieldValueFromRequiredSecret(
			s.apimanager.Spec.System.FileStorageSpec.S3.ConfigurationSecretRef.Name,
			component.AwsRoleArn)
		if err != nil {
			return err
		}

		tokenPath, err := s.secretSource.RequiredFieldValueFromRequiredSecret(
			s.apimanager.Spec.System.FileStorageSpec.S3.ConfigurationSecretRef.Name,
			component.AwsWebIdentityTokenFile)
		if err != nil {
			return err
		}

		s.options.S3FileStorageOptions.STSTokenMountRelativePath = filepath.Base(tokenPath)
		s.options.S3FileStorageOptions.STSTokenMountPath = filepath.Dir(tokenPath)

		// Validate it exists
		_, err = s.secretSource.RequiredFieldValueFromRequiredSecret(
			s.apimanager.Spec.System.FileStorageSpec.S3.ConfigurationSecretRef.Name,
			component.AwsBucket)
		if err != nil {
			return err
		}

		// Validate it exists
		_, err = s.secretSource.RequiredFieldValueFromRequiredSecret(
			s.apimanager.Spec.System.FileStorageSpec.S3.ConfigurationSecretRef.Name,
			component.AwsRegion)
		if err != nil {
			return err
		}

		return nil
	}

	if s.apimanager.IsS3IAMEnabled() {
		s.options.S3FileStorageOptions = &component.S3FileStorageOptions{
			ConfigurationSecretName: s.apimanager.Spec.System.FileStorageSpec.S3.ConfigurationSecretRef.Name,
			STSEnabled:              false,
		}

		// Validate it exists
		_, err := s.secretSource.RequiredFieldValueFromRequiredSecret(
			s.apimanager.Spec.System.FileStorageSpec.S3.ConfigurationSecretRef.Name,
			component.AwsAccessKeyID)
		if err != nil {
			return err
		}

		// Validate it exists
		_, err = s.secretSource.RequiredFieldValueFromRequiredSecret(
			s.apimanager.Spec.System.FileStorageSpec.S3.ConfigurationSecretRef.Name,
			component.AwsSecretAccessKey)
		if err != nil {
			return err
		}

		// Validate it exists
		_, err = s.secretSource.RequiredFieldValueFromRequiredSecret(
			s.apimanager.Spec.System.FileStorageSpec.S3.ConfigurationSecretRef.Name,
			component.AwsBucket)
		if err != nil {
			return err
		}

		// Validate it exists
		_, err = s.secretSource.RequiredFieldValueFromRequiredSecret(
			s.apimanager.Spec.System.FileStorageSpec.S3.ConfigurationSecretRef.Name,
			component.AwsRegion)
		if err != nil {
			return err
		}

		return nil
	}

	// defaults to PVC
	var storageClassName *string
	var volumeName *string
	storageRequests := component.DefaultSharedStorageResources()
	if s.apimanager.Spec.System != nil &&
		s.apimanager.Spec.System.FileStorageSpec != nil &&
		s.apimanager.Spec.System.FileStorageSpec.PVC != nil {
		storageClassName = s.apimanager.Spec.System.FileStorageSpec.PVC.StorageClassName
		volumeName = s.apimanager.Spec.System.FileStorageSpec.PVC.VolumeName
		if s.apimanager.Spec.System.FileStorageSpec.PVC.Resources != nil {
			storageRequests = s.apimanager.Spec.System.FileStorageSpec.PVC.Resources.Requests
		}
	}

	s.options.PvcFileStorageOptions = &component.PVCFileStorageOptions{
		StorageClass:    storageClassName,
		VolumeName:      volumeName,
		StorageRequests: storageRequests,
	}

	return nil
}

func (s *SystemOptionsProvider) setReplicas() {
	s.options.AppReplicas = 1
	if s.apimanager.Spec.System.AppSpec.Replicas != nil {
		s.options.AppReplicas = int32(*s.apimanager.Spec.System.AppSpec.Replicas)
	}

	s.options.SidekiqReplicas = 1
	if s.apimanager.Spec.System.SidekiqSpec.Replicas != nil {
		s.options.SidekiqReplicas = int32(*s.apimanager.Spec.System.SidekiqSpec.Replicas)
	}
}

func (s *SystemOptionsProvider) commonLabels() map[string]string {
	return map[string]string{
		"app":                  *s.apimanager.Spec.AppLabel,
		"threescale_component": "system",
	}
}

func (s *SystemOptionsProvider) commonAppLabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "app"
	return labels
}

func (s *SystemOptionsProvider) appPodTemplateLabels() map[string]string {
	labels := helper.MeteringLabels("system-app", helper.ApplicationType)

	for k, v := range s.commonAppLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "system-app"

	return labels
}

func (s *SystemOptionsProvider) commonSidekiqLabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "sidekiq"
	return labels
}

func (s *SystemOptionsProvider) sidekiqPodTemplateLabels() map[string]string {
	labels := helper.MeteringLabels("system-sidekiq", helper.ApplicationType)

	for k, v := range s.commonSidekiqLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "system-sidekiq"

	return labels
}

func (s *SystemOptionsProvider) providerUILabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "provider-ui"
	return labels
}

func (s *SystemOptionsProvider) masterUILabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "master-ui"
	return labels
}

func (s *SystemOptionsProvider) developerUILabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "developer-ui"
	return labels
}

func (s *SystemOptionsProvider) sphinxLabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "sphinx"
	return labels
}

func (s *SystemOptionsProvider) memcachedLabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "memcache"
	return labels
}

func (s *SystemOptionsProvider) smtpLabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "smtp"
	return labels
}

func (s *SystemOptionsProvider) sphinxPodTemplateLabels() map[string]string {
	labels := helper.MeteringLabels("system-sphinx", helper.ApplicationType)

	for k, v := range s.sphinxLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "system-sphinx"

	return labels
}
