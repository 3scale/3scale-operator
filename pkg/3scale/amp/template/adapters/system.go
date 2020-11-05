package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type System struct {
	generatePodDisruptionBudget bool
}

func NewSystemAdapter(generatePDB bool) Adapter {
	return NewAppenderAdapter(&System{generatePodDisruptionBudget: generatePDB})
}

func (s *System) Parameters() []templatev1.Parameter {
	return []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "WILDCARD_DOMAIN",
			Description: "Root domain for the wildcard routes. Eg. example.com will generate 3scale-admin.example.com.",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_BACKEND_USERNAME",
			Description: "Internal 3scale API username for internal 3scale api auth.",
			Value:       "3scale_api_user",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_BACKEND_PASSWORD",
			Description: "Internal 3scale API password for internal 3scale api auth.",
			Generate:    "expression",
			From:        "[a-z0-9]{8}",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_BACKEND_SHARED_SECRET",
			Description: "Shared secret to import events from backend to system.",
			Generate:    "expression",
			From:        "[a-z0-9]{8}",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_APP_SECRET_KEY_BASE",
			Description: "System application secret key base",
			Generate:    "expression",
			From:        "[a-f0-9]{128}",
			Required:    true,
		},
		templatev1.Parameter{
			Name:     "ADMIN_PASSWORD",
			Generate: "expression",
			From:     "[a-z0-9]{8}",
			Required: true,
		},
		templatev1.Parameter{
			Name:     "ADMIN_USERNAME",
			Value:    "admin",
			Required: true,
		},
		templatev1.Parameter{
			Name:     "ADMIN_EMAIL",
			Required: false,
		},
		templatev1.Parameter{
			Name:        "ADMIN_ACCESS_TOKEN",
			Description: "Admin Access Token with all scopes and write permissions for API access.",
			Generate:    "expression",
			From:        "[a-z0-9]{16}",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "MASTER_NAME",
			Description: "The root name which Master Admin UI will be available at.",
			Value:       "master",
			Required:    true,
		},
		templatev1.Parameter{
			Name:     "MASTER_USER",
			Value:    "master",
			Required: true,
		},
		templatev1.Parameter{
			Name:     "MASTER_PASSWORD",
			Generate: "expression",
			From:     "[a-z0-9]{8}",
			Required: true,
		},
		templatev1.Parameter{
			Name:     "MASTER_ACCESS_TOKEN",
			Generate: "expression",
			From:     "[a-z0-9]{8}",
			Required: true,
		},
		templatev1.Parameter{
			Name:        "RECAPTCHA_PUBLIC_KEY",
			Description: "reCAPTCHA site key (used in spam protection)",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "RECAPTCHA_PRIVATE_KEY",
			Description: "reCAPTCHA secret key (used in spam protection)",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_REDIS_URL",
			Description: "Define the external system-redis to connect to",
			Required:    true,
			Value:       "redis://system-redis:6379/1",
		},
		templatev1.Parameter{
			Name:        "SYSTEM_MESSAGE_BUS_REDIS_URL",
			Description: "Define the external system-redis message bus to connect to. By default the same value as SYSTEM_REDIS_URL but with the logical database incremented by 1 and the result applied mod 16",
		},
		templatev1.Parameter{
			Name:        "SYSTEM_REDIS_NAMESPACE",
			Description: "Define the namespace to be used by System's Redis Database. The empty value means not namespaced",
		},
		templatev1.Parameter{
			Name:        "SYSTEM_MESSAGE_BUS_REDIS_NAMESPACE",
			Description: "Define the namespace to be used by System's Message Bus Redis Database. The empty value means not namespaced",
		},
	}
}

func (s *System) Objects() ([]common.KubernetesObject, error) {
	systemOptions, err := s.options()
	if err != nil {
		return nil, err
	}
	systemComponent := component.NewSystem(systemOptions)
	objects := s.componentObjects(systemComponent)

	return objects, nil
}

func (s *System) componentObjects(c *component.System) []common.KubernetesObject {
	sharedStorage := c.SharedStorage()
	providerService := c.ProviderService()
	masterService := c.MasterService()
	developerService := c.DeveloperService()
	sphinxService := c.SphinxService()
	memcachedService := c.MemcachedService()

	appDeploymentConfig := c.AppDeploymentConfig()
	sidekiqDeploymentConfig := c.SidekiqDeploymentConfig()
	sphinxDeploymentConfig := c.SphinxDeploymentConfig()

	systemConfigMap := c.SystemConfigMap()
	environmentConfigMap := c.EnvironmentConfigMap()
	smtpSecret := c.SMTPSecret()

	eventsHookSecret := c.EventsHookSecret()

	redisSecret := c.RedisSecret()
	masterApicastSecret := c.MasterApicastSecret()

	seedSecret := c.SeedSecret()
	recaptchaSecret := c.RecaptchaSecret()
	appSecret := c.AppSecret()
	memcachedSecret := c.MemcachedSecret()

	objects := []common.KubernetesObject{
		sharedStorage,
		providerService,
		masterService,
		developerService,
		sphinxService,
		memcachedService,
		systemConfigMap,
		smtpSecret,
		environmentConfigMap,
		appDeploymentConfig,
		sidekiqDeploymentConfig,
		sphinxDeploymentConfig,
		eventsHookSecret,
		redisSecret,
		masterApicastSecret,
		seedSecret,
		recaptchaSecret,
		appSecret,
		memcachedSecret,
	}

	if s.generatePodDisruptionBudget {
		objects = append(objects, s.componentPDBObjects(c)...)
	}

	return objects
}

func (s *System) componentPDBObjects(c *component.System) []common.KubernetesObject {
	appPDB := c.AppPodDisruptionBudget()
	sidekiqPDB := c.SidekiqPodDisruptionBudget()

	return []common.KubernetesObject{
		appPDB,
		sidekiqPDB,
	}
}

func (s *System) options() (*component.SystemOptions, error) {
	o := component.NewSystemOptions()

	o.AdminAccessToken = "${ADMIN_ACCESS_TOKEN}"
	o.AdminPassword = "${ADMIN_PASSWORD}"
	o.AdminUsername = "${ADMIN_USERNAME}"
	adminEmail := "${ADMIN_EMAIL}"
	o.AdminEmail = &adminEmail
	o.AmpRelease = "${AMP_RELEASE}"
	o.ImageTag = "${AMP_RELEASE}"
	o.ApicastAccessToken = "${APICAST_ACCESS_TOKEN}"
	o.ApicastRegistryURL = "${APICAST_REGISTRY_URL}"
	o.MasterAccessToken = "${MASTER_ACCESS_TOKEN}"
	o.MasterName = "${MASTER_NAME}"
	o.MasterUsername = "${MASTER_USER}"
	o.MasterPassword = "${MASTER_PASSWORD}"
	recaptchaPublicKey := "${RECAPTCHA_PUBLIC_KEY}"
	o.RecaptchaPublicKey = &recaptchaPublicKey
	recaptchaPrivateKey := "${RECAPTCHA_PRIVATE_KEY}"
	o.RecaptchaPrivateKey = &recaptchaPrivateKey
	o.RedisURL = "${SYSTEM_REDIS_URL}"
	redisSentinelHosts := component.DefaultSystemRedisSentinelHosts()
	o.RedisSentinelHosts = &redisSentinelHosts
	redisSentinelRole := component.DefaultSystemRedisSentinelRole()
	o.RedisSentinelRole = &redisSentinelRole
	redisNamespace := "${SYSTEM_REDIS_NAMESPACE}"
	o.RedisNamespace = &redisNamespace
	messageBusRedisURL := "${SYSTEM_MESSAGE_BUS_REDIS_URL}"
	o.MessageBusRedisURL = &messageBusRedisURL
	messageBusRedisSentinelHosts := component.DefaultSystemMessageBusRedisSentinelHosts()
	o.MessageBusRedisSentinelHosts = &messageBusRedisSentinelHosts
	messageBusRedisSentinelRole := component.DefaultSystemMessageBusRedisSentinelRole()
	o.MessageBusRedisSentinelRole = &messageBusRedisSentinelRole
	messageBusRedisNamespace := "${SYSTEM_MESSAGE_BUS_REDIS_NAMESPACE}"
	o.MessageBusRedisNamespace = &messageBusRedisNamespace
	o.AppSecretKeyBase = "${SYSTEM_APP_SECRET_KEY_BASE}"
	o.BackendSharedSecret = "${SYSTEM_BACKEND_SHARED_SECRET}"
	o.TenantName = "${TENANT_NAME}"
	o.WildcardDomain = "${WILDCARD_DOMAIN}"
	o.PvcFileStorageOptions = &component.PVCFileStorageOptions{}
	o.MemcachedServers = component.DefaultMemcachedServers()
	o.EventHooksURL = component.DefaultEventHooksURL()
	o.ApicastSystemMasterProxyConfigEndpoint = component.DefaultApicastSystemMasterProxyConfigEndpoint(o.ApicastAccessToken)
	o.AppProviderContainerResourceRequirements = component.DefaultAppProviderContainerResourceRequirements()
	o.AppMasterContainerResourceRequirements = component.DefaultAppMasterContainerResourceRequirements()
	o.AppDeveloperContainerResourceRequirements = component.DefaultAppDeveloperContainerResourceRequirements()
	o.SphinxContainerResourceRequirements = component.DefaultSphinxContainerResourceRequirements()
	o.SidekiqContainerResourceRequirements = component.DefaultSidekiqContainerResourceRequirements()
	o.AppReplicas = component.DefaultAppReplicas()
	o.SidekiqReplicas = component.DefaultSidekiqReplicas()
	defaultSystemSMTPAddress := component.DefaultSystemSMTPAddress()
	defaultSystemSMTPAuthentication := component.DefaultSystemSMTPAuthentication()
	defaultSystemSMTPDomain := component.DefaultSystemSMTPDomain()
	defaultSystemSMTPOpenSSLVerifyMode := component.DefaultSystemSMTPOpenSSLVerifyMode()
	defaultSystemSMTPPassword := component.DefaultSystemSMTPPassword()
	defaultSystemSMTPPort := component.DefaultSystemSMTPPort()
	defaultSystemSMTPUsername := component.DefaultSystemSMTPUsername()
	o.SmtpSecretOptions = component.SystemSMTPSecretOptions{
		Address:           &defaultSystemSMTPAddress,
		Authentication:    &defaultSystemSMTPAuthentication,
		Domain:            &defaultSystemSMTPDomain,
		OpenSSLVerifyMode: &defaultSystemSMTPOpenSSLVerifyMode,
		Password:          &defaultSystemSMTPPassword,
		Port:              &defaultSystemSMTPPort,
		Username:          &defaultSystemSMTPUsername,
	}

	o.CommonLabels = s.commonLabels()
	o.CommonAppLabels = s.commonAppLabels()
	o.AppPodTemplateLabels = s.appPodTemplateLabels()
	o.CommonSidekiqLabels = s.commonSidekiqLabels()
	o.SidekiqPodTemplateLabels = s.sidekiqPodTemplateLabels()
	o.ProviderUILabels = s.providerUILabels()
	o.MasterUILabels = s.masterUILabels()
	o.DeveloperUILabels = s.developerUILabels()
	o.SphinxLabels = s.sphinxLabels()
	o.SphinxPodTemplateLabels = s.sphinxPodTemplateLabels()
	o.MemcachedLabels = s.memcachedLabels()
	o.SMTPLabels = s.smtpLabels()

	err := o.Validate()
	return o, err
}

func (s *System) commonLabels() map[string]string {
	return map[string]string{
		"app":                  "${APP_LABEL}",
		"threescale_component": "system",
	}
}

func (s *System) commonAppLabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "app"
	return labels
}

func (s *System) appPodTemplateLabels() map[string]string {
	labels := s.commonAppLabels()
	labels["deploymentConfig"] = "system-app"
	return labels
}

func (s *System) commonSidekiqLabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "sidekiq"
	return labels
}

func (s *System) sidekiqPodTemplateLabels() map[string]string {
	labels := s.commonSidekiqLabels()
	labels["deploymentConfig"] = "system-sidekiq"
	return labels
}

func (s *System) providerUILabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "provider-ui"
	return labels
}

func (s *System) masterUILabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "master-ui"
	return labels
}

func (s *System) developerUILabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "developer-ui"
	return labels
}

func (s *System) sphinxLabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "sphinx"
	return labels
}

func (s *System) sphinxPodTemplateLabels() map[string]string {
	labels := s.sphinxLabels()
	labels["deploymentConfig"] = "system-sphinx"
	return labels
}

func (s *System) memcachedLabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "memcache"
	return labels
}

func (s *System) smtpLabels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "smtp"
	return labels
}
