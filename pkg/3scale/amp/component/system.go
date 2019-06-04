package component

import (
	"fmt"
	"github.com/3scale/3scale-operator/pkg/apis/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"sort"

	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	SystemSecretSystemDatabaseSecretName            = "system-database"
	SystemSecretSystemDatabaseURLFieldName          = "URL"
	SystemSecretSystemDatabaseDatabaseNameFieldName = "DB_NAME"
	SystemSecretSystemDatabaseUserFieldName         = "DB_USER"
	SystemSecretSystemDatabasePasswordFieldName     = "DB_PASSWORD"
	SystemSecretSystemDatabaseRootPasswordFieldName = "DB_ROOT_PASSWORD"
)

const (
	SystemSecretSystemMemcachedSecretName       = "system-memcache"
	SystemSecretSystemMemcachedServersFieldName = "SERVERS"
)

const (
	SystemSecretSystemRecaptchaSecretName          = "system-recaptcha"
	SystemSecretSystemRecaptchaPublicKeyFieldName  = "PUBLIC_KEY"
	SystemSecretSystemRecaptchaPrivateKeyFieldName = "PRIVATE_KEY"
)

const (
	SystemSecretSystemEventsHookSecretName        = "system-events-hook"
	SystemSecretSystemEventsHookURLFieldName      = "URL"
	SystemSecretSystemEventsHookPasswordFieldName = "PASSWORD"
)

const (
	SystemSecretSystemRedisSecretName                  = "system-redis"
	SystemSecretSystemRedisURLFieldName                = "URL"
	SystemSecretSystemRedisMessageBusRedisURLFieldName = "MESSAGE_BUS_URL"
	SystemSecretSystemRedisNamespace                   = "NAMESPACE"
	SystemSecretSystemRedisMessageBusRedisNamespace    = "MESSAGE_BUS_NAMESPACE"
)

const (
	SystemSecretSystemAppSecretName             = "system-app"
	SystemSecretSystemAppSecretKeyBaseFieldName = "SECRET_KEY_BASE"
)

const (
	SystemSecretSystemSeedSecretName                 = "system-seed"
	SystemSecretSystemSeedMasterDomainFieldName      = "MASTER_DOMAIN"
	SystemSecretSystemSeedMasterAccessTokenFieldName = "MASTER_ACCESS_TOKEN"
	SystemSecretSystemSeedMasterUserFieldName        = "MASTER_USER"
	SystemSecretSystemSeedMasterPasswordFieldName    = "MASTER_PASSWORD"
	SystemSecretSystemSeedAdminAccessTokenFieldName  = "ADMIN_ACCESS_TOKEN"
	SystemSecretSystemSeedAdminUserFieldName         = "ADMIN_USER"
	SystemSecretSystemSeedAdminPasswordFieldName     = "ADMIN_PASSWORD"
	SystemSecretSystemSeedAdminEmailFieldName        = "ADMIN_EMAIL"
	SystemSecretSystemSeedTenantNameFieldName        = "TENANT_NAME"
)

const (
	SystemSecretSystemMasterApicastSecretName                    = "system-master-apicast"
	SystemSecretSystemMasterApicastProxyConfigsEndpointFieldName = "PROXY_CONFIGS_ENDPOINT"
	SystemSecretSystemMasterApicastBaseURL                       = "BASE_URL"
	SystemSecretSystemMasterApicastAccessToken                   = "ACCESS_TOKEN"
)

type System struct {
	options []string
	Options *SystemOptions
}

type SystemOptions struct {
	systemNonRequiredOptions
	systemRequiredOptions
}

type systemRequiredOptions struct {
	adminAccessToken    string
	adminPassword       string
	adminUsername       string
	ampRelease          string
	apicastAccessToken  string
	apicastRegistryURL  string
	appLabel            string
	masterAccessToken   string
	masterName          string
	masterUsername      string
	masterPassword      string
	recaptchaPublicKey  string
	recaptchaPrivateKey string
	appSecretKeyBase    string
	backendSharedSecret string
	tenantName          string
	wildcardDomain      string
	storageClassName    *string // should this be a string or *string? check what would be the difference between passing a "" and a nil pointer in the PersistentVolumeClaim corresponding field
}

type systemNonRequiredOptions struct {
	memcachedServers                       *string
	eventHooksURL                          *string
	redisURL                               *string
	redisNamespace                         *string
	messageBusRedisNamespace               *string
	messageBusRedisURL                     *string
	apicastSystemMasterProxyConfigEndpoint *string
	apicastSystemMasterBaseURL             *string
	adminEmail                             *string
}

func NewSystem(options []string) *System {
	system := &System{
		options: options,
	}
	return system
}

type SystemOptionsProvider interface {
	GetSystemOptions() *SystemOptions
}
type CLISystemOptionsProvider struct {
}

func (o *CLISystemOptionsProvider) GetSystemOptions() (*SystemOptions, error) {
	sob := SystemOptionsBuilder{}
	sob.AdminAccessToken("${ADMIN_ACCESS_TOKEN}")
	sob.AdminPassword("${ADMIN_PASSWORD}")
	sob.AdminUsername("${ADMIN_USERNAME}")
	sob.AdminEmail("${ADMIN_EMAIL}")
	sob.AmpRelease("${AMP_RELEASE}")
	sob.ApicastAccessToken("${APICAST_ACCESS_TOKEN}")
	sob.ApicastRegistryURL("${APICAST_REGISTRY_URL}")
	sob.MasterAccessToken("${MASTER_ACCESS_TOKEN}")
	sob.MasterName("${MASTER_NAME}")
	sob.MasterUsername("${MASTER_USER}")
	sob.MasterPassword("${MASTER_PASSWORD}")
	sob.AppLabel("${APP_LABEL}")
	sob.RecaptchaPublicKey("${RECAPTCHA_PUBLIC_KEY}")
	sob.RecaptchaPrivateKey("${RECAPTCHA_PRIVATE_KEY}")
	sob.RedisURL("${SYSTEM_REDIS_URL}")
	sob.RedisNamespace("${SYSTEM_REDIS_NAMESPACE}")
	sob.MessageBusRedisURL("${SYSTEM_MESSAGE_BUS_REDIS_URL}")
	sob.MessageBusRedisNamespace("${SYSTEM_MESSAGE_BUS_REDIS_NAMESPACE}")
	sob.AppSecretKeyBase("${SYSTEM_APP_SECRET_KEY_BASE}")
	sob.BackendSharedSecret("${SYSTEM_BACKEND_SHARED_SECRET}")
	sob.TenantName("${TENANT_NAME}")
	sob.WildcardDomain("${WILDCARD_DOMAIN}")
	sob.StorageClassName(nil)
	res, err := sob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create System Options - %s", err)
	}
	return res, nil
}

func (system *System) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	// TODO move this outside this specific method
	optionsProvider := CLISystemOptionsProvider{}
	systemOpts, err := optionsProvider.GetSystemOptions()
	_ = err
	system.Options = systemOpts
	system.buildParameters(template)
	system.addObjectsIntoTemplate(template)
}

func (system *System) GetObjects() ([]common.KubernetesObject, error) {
	objects := system.buildObjects()
	return objects, nil
}

func (system *System) addObjectsIntoTemplate(template *templatev1.Template) {
	objects := system.buildObjects()
	template.Objects = append(template.Objects, helper.WrapRawExtensions(objects)...)
}

func (system *System) PostProcess(template *templatev1.Template, otherComponents []Component) {

}

func (system *System) buildParameters(template *templatev1.Template) {
	parameters := []templatev1.Parameter{
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
	template.Parameters = append(template.Parameters, parameters...)
}

func (system *System) buildObjects() []common.KubernetesObject {
	systemSharedStorage := system.buildSystemSharedPVC()
	systemProviderService := system.buildSystemProviderService()
	systemMasterService := system.buildSystemMasterService()
	systemDeveloperService := system.buildSystemDeveloperService()
	systemProviderRoute := system.buildSystemProviderRoute()
	systemMasterRoute := system.buildSystemMasterRoute()
	systemDeveloperRoute := system.buildSystemDeveloperRoute()
	systemRedisService := system.buildSystemRedisService()
	systemSphinxService := system.buildSystemSphinxService()
	systemMemcachedService := system.buildSystemMemcachedService()

	systemAppDeploymentConfig := system.buildSystemAppDeploymentConfig()
	systemSidekiqDeploymentConfig := system.buildSystemSidekiqDeploymentConfig()
	systemSphinxDeploymentConfig := system.buildSystemSphinxDeploymentConfig()

	systemConfigMap := system.buildSystemConfigMap()
	systemEnvironmentConfigMap := system.buildSystemEnvironmentConfigMap()
	systemSmtpConfigMap := system.buildSystemSmtpConfigMap()

	systemEventsHookSecret := system.buildSystemEventsHookSecrets()

	systemRedisSecret := system.buildSystemRedisSecrets()
	systemMasterApicastSecret := system.buildSystemMasterApicastSecrets()

	systemSeedSecret := system.buildSystemSeedSecrets()
	systemRecaptchaSecret := system.buildSystemRecaptchaSecrets()
	systemAppSecret := system.buildSystemAppSecrets()
	systemMemcachedSecret := system.buildSystemMemcachedSecrets()

	objects := []common.KubernetesObject{
		systemSharedStorage,
		systemProviderService,
		systemMasterService,
		systemDeveloperService,
		systemProviderRoute,
		systemMasterRoute,
		systemDeveloperRoute,
		systemRedisService,
		systemSphinxService,
		systemMemcachedService,
		systemConfigMap,
		systemSmtpConfigMap,
		systemEnvironmentConfigMap,
		systemAppDeploymentConfig,
		systemSidekiqDeploymentConfig,
		systemSphinxDeploymentConfig,
		systemEventsHookSecret,
		systemRedisSecret,
		systemMasterApicastSecret,
		systemSeedSecret,
		systemRecaptchaSecret,
		systemAppSecret,
		systemMemcachedSecret,
	}
	return objects
}

func (system *System) getSystemBaseEnvsFromEnvConfigMap() []v1.EnvVar {
	result := []v1.EnvVar{}

	// Add system-base-env ConfigMap values to envvar sources
	cfg := system.buildSystemEnvironmentConfigMap()
	cfgmapkeys := make([]string, 0, len(cfg.Data))
	for key := range cfg.Data {
		cfgmapkeys = append(cfgmapkeys, key)
	}
	sort.Strings(cfgmapkeys)
	for _, key := range cfgmapkeys {
		envvar := envVarFromConfigMap(key, "system-environment", key)
		result = append(result, envvar)
	}

	return result
}

func (system *System) getSystemSmtpEnvsFromSMTPConfigMap() []v1.EnvVar {
	// Add smtp configmap to sources
	cfg := system.buildSystemSmtpConfigMap()
	cfgmapkeys := make([]string, 0, len(cfg.Data))
	for key := range cfg.Data {
		cfgmapkeys = append(cfgmapkeys, key)
	}
	sort.Strings(cfgmapkeys)

	// This cannot be used because the config map keys currently
	// do not have the same name than the envvar names in base_env
	// for _, key := range cfgmapkeys {
	// 	envvar := envVarFromConfigMap(key, "smtp", key)
	// 	result = append(result, envvar)
	// }

	result := []v1.EnvVar{
		envVarFromConfigMap("SMTP_ADDRESS", "smtp", "address"),
		envVarFromConfigMap("SMTP_USER_NAME", "smtp", "username"),
		envVarFromConfigMap("SMTP_PASSWORD", "smtp", "password"),
		envVarFromConfigMap("SMTP_DOMAIN", "smtp", "domain"),
		envVarFromConfigMap("SMTP_PORT", "smtp", "port"),
		envVarFromConfigMap("SMTP_AUTHENTICATION", "smtp", "authentication"),
		envVarFromConfigMap("SMTP_OPENSSL_VERIFY_MODE", "smtp", "openssl.verify.mode"),
	}

	return result
}

func (system *System) buildSystemSphinxEnv() []v1.EnvVar {
	result := []v1.EnvVar{}

	result = append(result,
		envVarFromConfigMap("RAILS_ENV", "system-environment", "RAILS_ENV"),
		envVarFromSecret("DATABASE_URL", SystemSecretSystemDatabaseSecretName, SystemSecretSystemDatabaseURLFieldName),
		envVarFromValue("THINKING_SPHINX_ADDRESS", "0.0.0.0"),
		envVarFromValue("THINKING_SPHINX_CONFIGURATION_FILE", "db/sphinx/production.conf"),
		envVarFromValue("THINKING_SPHINX_PID_FILE", "db/sphinx/searchd.pid"),
		envVarFromValue("DELTA_INDEX_INTERVAL", "5"),
		envVarFromValue("FULL_REINDEX_INTERVAL", "60"),
	)
	result = append(result, system.SystemRedisEnvVars()...)
	return result
}

func (system *System) SystemRedisEnvVars() []v1.EnvVar {
	result := []v1.EnvVar{}

	result = append(result,
		envVarFromSecret("REDIS_URL", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisURLFieldName),
		envVarFromSecret("MESSAGE_BUS_REDIS_URL", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisMessageBusRedisURLFieldName),
		envVarFromSecret("REDIS_NAMESPACE", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisNamespace),
		envVarFromSecret("MESSAGE_BUS_REDIS_NAMESPACE", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisMessageBusRedisNamespace),
	)
	return result
}

func (system *System) buildSystemBaseEnv() []v1.EnvVar {
	result := []v1.EnvVar{}

	baseEnvConfigMapEnvs := system.getSystemBaseEnvsFromEnvConfigMap()
	result = append(result, baseEnvConfigMapEnvs...)

	result = append(result,
		envVarFromSecret("DATABASE_URL", SystemSecretSystemDatabaseSecretName, SystemSecretSystemDatabaseURLFieldName),

		envVarFromSecret("MASTER_DOMAIN", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedMasterDomainFieldName),
		envVarFromSecret("MASTER_USER", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedMasterUserFieldName),
		envVarFromSecret("MASTER_PASSWORD", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedMasterPasswordFieldName),

		envVarFromSecret("ADMIN_ACCESS_TOKEN", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedAdminAccessTokenFieldName),
		envVarFromSecret("USER_LOGIN", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedAdminUserFieldName),
		envVarFromSecret("USER_PASSWORD", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedAdminPasswordFieldName),
		envVarFromSecret("USER_EMAIL", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedAdminEmailFieldName),
		envVarFromSecret("TENANT_NAME", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedTenantNameFieldName),

		envVarFromValue("THINKING_SPHINX_ADDRESS", "system-sphinx"),
		envVarFromValue("THINKING_SPHINX_CONFIGURATION_FILE", "/tmp/sphinx.conf"),

		envVarFromSecret("EVENTS_SHARED_SECRET", SystemSecretSystemEventsHookSecretName, SystemSecretSystemEventsHookPasswordFieldName),

		envVarFromSecret("RECAPTCHA_PUBLIC_KEY", SystemSecretSystemRecaptchaSecretName, SystemSecretSystemRecaptchaPublicKeyFieldName),
		envVarFromSecret("RECAPTCHA_PRIVATE_KEY", SystemSecretSystemRecaptchaSecretName, SystemSecretSystemRecaptchaPrivateKeyFieldName),

		envVarFromSecret("SECRET_KEY_BASE", SystemSecretSystemAppSecretName, SystemSecretSystemAppSecretKeyBaseFieldName),

		envVarFromSecret("MEMCACHE_SERVERS", SystemSecretSystemMemcachedSecretName, SystemSecretSystemMemcachedServersFieldName),

		envVarFromSecret("BACKEND_REDIS_URL", "backend-redis", "REDIS_STORAGE_URL"),
	)

	result = append(result, system.SystemRedisEnvVars()...)
	bckListenerApicastRouteEnv := envVarFromSecret("APICAST_BACKEND_ROOT_ENDPOINT", "backend-listener", "route_endpoint")
	bckListenerRouteEnv := envVarFromSecret("BACKEND_ROUTE", "backend-listener", "route_endpoint")
	result = append(result, bckListenerApicastRouteEnv, bckListenerRouteEnv)

	smtpEnvConfigMapEnvs := system.getSystemSmtpEnvsFromSMTPConfigMap()
	result = append(result, smtpEnvConfigMapEnvs...)

	apicastAccessToken := envVarFromSecret("APICAST_ACCESS_TOKEN", "system-master-apicast", "ACCESS_TOKEN")
	result = append(result, apicastAccessToken)

	// Add zync secret to envvars sources
	zyncAuthTokenVar := envVarFromSecret("ZYNC_AUTHENTICATION_TOKEN", "zync", "ZYNC_AUTHENTICATION_TOKEN")
	result = append(result, zyncAuthTokenVar)

	// Add backend internal api data to envvars sources
	systemBackendInternalAPIUser := envVarFromSecret("CONFIG_INTERNAL_API_USER", "backend-internal-api", "username")
	systemBackendInternalAPIPass := envVarFromSecret("CONFIG_INTERNAL_API_PASSWORD", "backend-internal-api", "password")
	result = append(result, systemBackendInternalAPIUser, systemBackendInternalAPIPass)

	return result
}

func (system *System) buildSystemAppPreHookEnv() []v1.EnvVar {
	return nil
}

func (system *System) buildSystemEnvironmentConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-environment",
			Labels: map[string]string{"threescale_component": "system", "app": system.Options.appLabel},
		},
		Data: map[string]string{
			"RAILS_ENV":              "production",
			"FORCE_SSL":              "true",
			"THREESCALE_SUPERDOMAIN": system.Options.wildcardDomain,
			"PROVIDER_PLAN":          "enterprise",
			"APICAST_REGISTRY_URL":   system.Options.apicastRegistryURL,
			"RAILS_LOG_TO_STDOUT":    "true",
			"RAILS_LOG_LEVEL":        "info",
			"THINKING_SPHINX_PORT":   "9306",
			"THREESCALE_SANDBOX_PROXY_OPENSSL_VERIFY_MODE": "VERIFY_NONE",
			"AMP_RELEASE":  system.Options.ampRelease,
			"SSL_CERT_DIR": "/etc/pki/tls/certs",
		},
	}
}

func (system *System) buildSystemMemcachedSecrets() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: SystemSecretSystemMemcachedSecretName,
			Labels: map[string]string{
				"app":                  system.Options.appLabel,
				"threescale_component": "system",
			},
		},
		StringData: map[string]string{
			SystemSecretSystemMemcachedServersFieldName: *system.Options.memcachedServers,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (system *System) buildSystemRecaptchaSecrets() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: SystemSecretSystemRecaptchaSecretName,
			Labels: map[string]string{
				"app":                  system.Options.appLabel,
				"threescale_component": "system",
			},
		},
		StringData: map[string]string{
			SystemSecretSystemRecaptchaPublicKeyFieldName:  system.Options.recaptchaPublicKey,
			SystemSecretSystemRecaptchaPrivateKeyFieldName: system.Options.recaptchaPrivateKey,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (system *System) buildSystemEventsHookSecrets() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: SystemSecretSystemEventsHookSecretName,
			Labels: map[string]string{
				"app":                  system.Options.appLabel,
				"threescale_component": "system",
			},
		},
		StringData: map[string]string{
			SystemSecretSystemEventsHookURLFieldName:      *system.Options.eventHooksURL,
			SystemSecretSystemEventsHookPasswordFieldName: system.Options.backendSharedSecret,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (system *System) buildSystemRedisSecrets() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: SystemSecretSystemRedisSecretName,
			Labels: map[string]string{
				"app":                  system.Options.appLabel,
				"threescale_component": "system",
			},
		},
		StringData: map[string]string{
			SystemSecretSystemRedisURLFieldName:                *system.Options.redisURL,
			SystemSecretSystemRedisMessageBusRedisURLFieldName: *system.Options.messageBusRedisURL,
			SystemSecretSystemRedisNamespace:                   *system.Options.redisNamespace,
			SystemSecretSystemRedisMessageBusRedisNamespace:    *system.Options.messageBusRedisNamespace,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (system *System) buildSystemAppSecrets() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: SystemSecretSystemAppSecretName, // TODO sure this should be a secret on its own?? maybe can join different secrets into one with more values?
			Labels: map[string]string{
				"app":                  system.Options.appLabel,
				"threescale_component": "system",
			},
		},
		StringData: map[string]string{
			SystemSecretSystemAppSecretKeyBaseFieldName: system.Options.appSecretKeyBase,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (system *System) buildSystemSeedSecrets() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: SystemSecretSystemSeedSecretName,
			Labels: map[string]string{
				"app":                  system.Options.appLabel,
				"threescale_component": "system",
			},
		},
		StringData: map[string]string{
			SystemSecretSystemSeedMasterDomainFieldName:      system.Options.masterName,
			SystemSecretSystemSeedMasterAccessTokenFieldName: system.Options.masterAccessToken,
			SystemSecretSystemSeedMasterUserFieldName:        system.Options.masterUsername,
			SystemSecretSystemSeedMasterPasswordFieldName:    system.Options.masterPassword,
			SystemSecretSystemSeedAdminAccessTokenFieldName:  system.Options.adminAccessToken,
			SystemSecretSystemSeedAdminUserFieldName:         system.Options.adminUsername,
			SystemSecretSystemSeedAdminPasswordFieldName:     system.Options.adminPassword,
			SystemSecretSystemSeedAdminEmailFieldName:        *system.Options.adminEmail,
			SystemSecretSystemSeedTenantNameFieldName:        system.Options.tenantName,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (system *System) buildSystemMasterApicastSecrets() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: SystemSecretSystemMasterApicastSecretName,
			Labels: map[string]string{
				"app":                  system.Options.appLabel,
				"threescale_component": "system",
			},
		},
		StringData: map[string]string{
			SystemSecretSystemMasterApicastProxyConfigsEndpointFieldName: *system.Options.apicastSystemMasterProxyConfigEndpoint,
			SystemSecretSystemMasterApicastBaseURL:                       *system.Options.apicastSystemMasterBaseURL,
			SystemSecretSystemMasterApicastAccessToken:                   system.Options.apicastAccessToken,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (system *System) buildSystemAppDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-app",
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "app", "app": system.Options.appLabel},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRolling,
				RollingParams: &appsv1.RollingDeploymentStrategyParams{
					UpdatePeriodSeconds: &[]int64{1}[0],
					IntervalSeconds:     &[]int64{1}[0],
					TimeoutSeconds:      &[]int64{1200}[0],
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Type(intstr.String),
						StrVal: "25%",
					},
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.Type(intstr.String),
						StrVal: "25%",
					},
					Pre: &appsv1.LifecycleHook{
						FailurePolicy: appsv1.LifecycleHookFailurePolicyRetry,
						ExecNewPod: &appsv1.ExecNewPodHook{
							// TODO the MASTER_ACCESS_TOKEN reference should be probably set as an envvar that gathers its value from the system-seed secret
							// but changing that probably has some implications during an upgrade process of the product
							Command:       []string{"bash", "-c", "bundle exec rake boot openshift:deploy " + "MASTER_ACCESS_TOKEN" + "=\"" + system.Options.masterAccessToken + "\""},
							Env:           system.buildSystemBaseEnv(),
							ContainerName: "system-master",
							Volumes:       []string{"system-storage"}},
					},
					Post: &appsv1.LifecycleHook{
						FailurePolicy: appsv1.LifecycleHookFailurePolicyAbort,
						ExecNewPod: &appsv1.ExecNewPodHook{
							Command:       []string{"bash", "-c", "bundle exec rake boot openshift:post_deploy"},
							ContainerName: "system-master"}}},
			},
			MinReadySeconds: 0,
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnConfigChange,
				}, appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic:      true,
						ContainerNames: []string{"system-provider", "system-developer", "system-master"},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "amp-system:latest"}}},
			},
			Replicas: 1,
			Selector: map[string]string{"deploymentConfig": "system-app"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "app", "app": system.Options.appLabel, "deploymentConfig": "system-app"},
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						v1.Volume{
							Name: "system-storage",
							VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
								ClaimName: "system-storage",
								ReadOnly:  false}},
						}, v1.Volume{
							Name: "system-config",
							VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "system",
								},
								Items: []v1.KeyToPath{
									v1.KeyToPath{
										Key:  "zync.yml",
										Path: "zync.yml",
									},
									v1.KeyToPath{
										Key:  "rolling_updates.yml",
										Path: "rolling_updates.yml",
									},
									v1.KeyToPath{
										Key:  "service_discovery.yml",
										Path: "service_discovery.yml",
									},
								}}}},
					},
					Containers: []v1.Container{
						v1.Container{
							Name:  "system-master",
							Image: "amp-system:latest",
							Args:  []string{"env", "TENANT_MODE=master", "PORT=3002", "container-entrypoint", "bundle", "exec", "unicorn", "-c", "config/unicorn.rb"},
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									Name:          "master",
									HostPort:      0,
									ContainerPort: 3002,
									Protocol:      v1.ProtocolTCP},
							},
							Env: system.buildSystemBaseEnv(),
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1000m"),
									v1.ResourceMemory: resource.MustParse("800Mi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("50m"),
									v1.ResourceMemory: resource.MustParse("600Mi"),
								},
							},
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "system-storage",
									ReadOnly:  false,
									MountPath: "/opt/system/public/system",
								}, v1.VolumeMount{
									Name:      "system-config",
									ReadOnly:  false,
									MountPath: "/opt/system-extra-configs"},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{TCPSocket: &v1.TCPSocketAction{
									Port: intstr.IntOrString{
										Type:   intstr.Type(intstr.String),
										StrVal: "master"}},
								},
								InitialDelaySeconds: 40,
								TimeoutSeconds:      10,
								PeriodSeconds:       10,
								SuccessThreshold:    0,
								FailureThreshold:    40,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{HTTPGet: &v1.HTTPGetAction{
									Path: "/check.txt",
									Port: intstr.IntOrString{
										Type:   intstr.Type(intstr.String),
										StrVal: "master",
									},
									Scheme: v1.URISchemeHTTP,
									HTTPHeaders: []v1.HTTPHeader{
										v1.HTTPHeader{
											Name:  "X-Forwarded-Proto",
											Value: "https"}}},
								},
								InitialDelaySeconds: 60,
								TimeoutSeconds:      10,
								PeriodSeconds:       30,
								SuccessThreshold:    0,
								FailureThreshold:    10,
							},
							ImagePullPolicy: v1.PullIfNotPresent,
							Stdin:           false,
							StdinOnce:       false,
							TTY:             false,
						}, v1.Container{
							Name:  "system-provider",
							Image: "amp-system:latest",
							Args:  []string{"env", "TENANT_MODE=provider", "PORT=3000", "container-entrypoint", "bundle", "exec", "unicorn", "-c", "config/unicorn.rb"},
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									Name:          "provider",
									HostPort:      0,
									ContainerPort: 3000,
									Protocol:      v1.ProtocolTCP},
							},
							Env: system.buildSystemBaseEnv(),
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1000m"),
									v1.ResourceMemory: resource.MustParse("800Mi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("50m"),
									v1.ResourceMemory: resource.MustParse("600Mi"),
								},
							},
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "system-storage",
									ReadOnly:  false,
									MountPath: "/opt/system/public/system",
								}, v1.VolumeMount{
									Name:      "system-config",
									ReadOnly:  false,
									MountPath: "/opt/system-extra-configs"},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{TCPSocket: &v1.TCPSocketAction{
									Port: intstr.IntOrString{
										Type:   intstr.Type(intstr.String),
										StrVal: "provider"}},
								},
								InitialDelaySeconds: 40,
								TimeoutSeconds:      10,
								PeriodSeconds:       10,
								SuccessThreshold:    0,
								FailureThreshold:    40,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{HTTPGet: &v1.HTTPGetAction{
									Path: "/check.txt",
									Port: intstr.IntOrString{
										Type:   intstr.Type(intstr.String),
										StrVal: "provider",
									},
									Scheme: v1.URISchemeHTTP,
									HTTPHeaders: []v1.HTTPHeader{
										v1.HTTPHeader{
											Name:  "X-Forwarded-Proto",
											Value: "https"}}},
								},
								InitialDelaySeconds: 60,
								TimeoutSeconds:      10,
								PeriodSeconds:       30,
								SuccessThreshold:    0,
								FailureThreshold:    10,
							},
							ImagePullPolicy: v1.PullIfNotPresent,
							Stdin:           false,
							StdinOnce:       false,
							TTY:             false,
						}, v1.Container{
							Name:  "system-developer",
							Image: "amp-system:latest",
							Args:  []string{"env", "PORT=3001", "container-entrypoint", "bundle", "exec", "unicorn", "-c", "config/unicorn.rb"},
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									Name:          "developer",
									HostPort:      0,
									ContainerPort: 3001,
									Protocol:      v1.ProtocolTCP},
							},
							Env: system.buildSystemBaseEnv(),
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1000m"),
									v1.ResourceMemory: resource.MustParse("800Mi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("50m"),
									v1.ResourceMemory: resource.MustParse("600Mi"),
								},
							},
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "system-storage",
									ReadOnly:  true,
									MountPath: "/opt/system/public/system",
								}, v1.VolumeMount{
									Name:      "system-config",
									ReadOnly:  false,
									MountPath: "/opt/system-extra-configs"},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{TCPSocket: &v1.TCPSocketAction{
									Port: intstr.IntOrString{
										Type:   intstr.Type(intstr.String),
										StrVal: "developer"}},
								},
								InitialDelaySeconds: 40,
								TimeoutSeconds:      10,
								PeriodSeconds:       10,
								SuccessThreshold:    0,
								FailureThreshold:    40,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{HTTPGet: &v1.HTTPGetAction{
									Path: "/check.txt",
									Port: intstr.IntOrString{
										Type:   intstr.Type(intstr.String),
										StrVal: "developer",
									},
									Scheme: v1.URISchemeHTTP,
									HTTPHeaders: []v1.HTTPHeader{
										v1.HTTPHeader{
											Name:  "X-Forwarded-Proto",
											Value: "https"}}},
								},
								InitialDelaySeconds: 60,
								TimeoutSeconds:      10,
								PeriodSeconds:       30,
								SuccessThreshold:    0,
								FailureThreshold:    10,
							},
							ImagePullPolicy: v1.PullIfNotPresent,
						},
					},
					ServiceAccountName: "amp",
				}},
		},
	}
}

func (system *System) buildSystemSidekiqDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-sidekiq",
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "sidekiq", "app": system.Options.appLabel},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRolling,
				RollingParams: &appsv1.RollingDeploymentStrategyParams{
					UpdatePeriodSeconds: &[]int64{1}[0],
					IntervalSeconds:     &[]int64{1}[0],
					TimeoutSeconds:      &[]int64{1200}[0],
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Type(intstr.String),
						StrVal: "25%",
					},
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.Type(intstr.String),
						StrVal: "25%"}},
			},
			MinReadySeconds: 0,
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnConfigChange,
				}, appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic:      true,
						ContainerNames: []string{"check-svc", "system-sidekiq"},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "amp-system:latest"}}},
			},
			Replicas: 1,
			Selector: map[string]string{"deploymentConfig": "system-sidekiq"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "sidekiq", "app": system.Options.appLabel, "deploymentConfig": "system-sidekiq"},
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						v1.Volume{
							Name: "system-tmp",
							VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{
								Medium: v1.StorageMediumMemory}},
						}, v1.Volume{
							Name: "system-storage",
							VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
								ClaimName: "system-storage",
								ReadOnly:  false}},
						}, v1.Volume{
							Name: "system-config",
							VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "system",
								},
								Items: []v1.KeyToPath{
									v1.KeyToPath{
										Key:  "zync.yml",
										Path: "zync.yml",
									}, v1.KeyToPath{
										Key:  "rolling_updates.yml",
										Path: "rolling_updates.yml",
									},
									v1.KeyToPath{
										Key:  "service_discovery.yml",
										Path: "service_discovery.yml",
									},
								}}}},
					},
					InitContainers: []v1.Container{
						v1.Container{
							Name:  "check-svc",
							Image: "amp-system:latest",
							Command: []string{
								"bash",
								"-c",
								"bundle exec sh -c \"until rake boot:redis && curl --output /dev/null --silent --fail --head http://system-master:3000/status; do sleep $SLEEP_SECONDS; done\"",
							},
							Env: []v1.EnvVar{
								envVarFromValue("SLEEP_SECONDS", "1"),
								envVarFromSecret("REDIS_URL", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisURLFieldName),
								envVarFromSecret("MESSAGE_BUS_REDIS_URL", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisMessageBusRedisURLFieldName),
								envVarFromSecret("REDIS_NAMESPACE", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisNamespace),
								envVarFromSecret("MESSAGE_BUS_REDIS_NAMESPACE", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisMessageBusRedisNamespace),
							},
						},
					},
					Containers: []v1.Container{
						v1.Container{
							Name:  "system-sidekiq",
							Image: "amp-system:latest",
							Args:  []string{"rake", "sidekiq:worker", "RAILS_MAX_THREADS=25"},
							Env:   system.buildSystemBaseEnv(),
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1000m"),
									v1.ResourceMemory: resource.MustParse("2Gi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("100m"),
									v1.ResourceMemory: resource.MustParse("500Mi"),
								},
							},
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "system-storage",
									ReadOnly:  false,
									MountPath: "/opt/system/public/system",
								}, v1.VolumeMount{
									Name:      "system-tmp",
									ReadOnly:  false,
									MountPath: "/tmp",
								}, v1.VolumeMount{
									Name:      "system-config",
									ReadOnly:  false,
									MountPath: "/opt/system-extra-configs"},
							},
							ImagePullPolicy: v1.PullIfNotPresent,
						},
					},
					ServiceAccountName: "amp",
				}},
		},
	}
}

func (system *System) buildSystemSharedPVC() *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-storage",
			Labels: map[string]string{
				"app":                          system.Options.appLabel,
				"threescale_component":         "system",
				"threescale_component_element": "app",
			},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteMany,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse("100Mi"),
				},
			},
		},
	}
}

func (system *System) buildSystemProviderService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-provider",
			Labels: map[string]string{
				"app":                          system.Options.appLabel,
				"threescale_component":         "system",
				"threescale_component_element": "provider-ui",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "http",
					Protocol:   v1.ProtocolTCP,
					Port:       3000,
					TargetPort: intstr.FromString("provider"),
				},
			},
			Selector: map[string]string{"deploymentConfig": "system-app"},
		},
	}
}

func (system *System) buildSystemMasterService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-master",
			Labels: map[string]string{
				"app":                          system.Options.appLabel,
				"threescale_component":         "system",
				"threescale_component_element": "master-ui",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "http",
					Protocol:   v1.ProtocolTCP,
					Port:       3000,
					TargetPort: intstr.FromString("master"),
				},
			},
			Selector: map[string]string{"deploymentConfig": "system-app"},
		},
	}
}

func (system *System) buildSystemDeveloperService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-developer",
			Labels: map[string]string{
				"app":                          system.Options.appLabel,
				"threescale_component":         "system",
				"threescale_component_element": "developer-ui",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "http",
					Protocol:   v1.ProtocolTCP,
					Port:       3000,
					TargetPort: intstr.FromString("developer"),
				},
			},
			Selector: map[string]string{"deploymentConfig": "system-app"},
		},
	}
}
func (system *System) buildSystemRedisService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-redis",
			Labels: map[string]string{
				"app":                          system.Options.appLabel,
				"threescale_component":         "system",
				"threescale_component_element": "redis",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "redis",
					Protocol:   v1.ProtocolTCP,
					Port:       6379,
					TargetPort: intstr.FromInt(6379),
				},
			},
			Selector: map[string]string{"deploymentConfig": "system-redis"},
		},
	}
}

func (system *System) buildSystemSphinxService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-sphinx",
			Labels: map[string]string{
				"app":                          system.Options.appLabel,
				"threescale_component":         "system",
				"threescale_component_element": "sphinx",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "sphinx",
					Protocol:   v1.ProtocolTCP,
					Port:       9306,
					TargetPort: intstr.FromInt(9306),
				},
			},
			Selector: map[string]string{"deploymentConfig": "system-sphinx"},
		},
	}
}

func (system *System) buildSystemMemcachedService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-memcache",
			Labels: map[string]string{
				"app":                          system.Options.appLabel,
				"threescale_component":         "system",
				"threescale_component_element": "memcache",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "memcache",
					Protocol:   v1.ProtocolTCP,
					Port:       11211,
					TargetPort: intstr.FromInt(11211),
				},
			},
			Selector: map[string]string{"deploymentConfig": "system-memcache"},
		},
	}
}

func (system *System) buildSystemSmtpConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "smtp",
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "smtp", "app": system.Options.appLabel},
		},
		Data: map[string]string{"address": "", "authentication": "", "domain": "", "openssl.verify.mode": "", "password": "", "port": "", "username": ""}}
}

func (system *System) buildSystemProviderRoute() *routev1.Route {
	return &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-provider-admin",
			Labels: map[string]string{"app": system.Options.appLabel, "threescale_component": "system", "threescale_component_element": "provider-ui"},
		},
		Spec: routev1.RouteSpec{
			Host: system.Options.tenantName + "-admin." + system.Options.wildcardDomain,
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "system-provider",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("http"),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyAllow},
		},
	}
}

func (system *System) buildSystemMasterRoute() *routev1.Route {
	return &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-master",
			Labels: map[string]string{"app": system.Options.appLabel, "threescale_component": "system", "threescale_component_element": "master-ui"},
		},
		Spec: routev1.RouteSpec{
			Host: system.Options.masterName + "." + system.Options.wildcardDomain,
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "system-master",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("http"),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyAllow},
		},
	}
}

func (system *System) buildSystemDeveloperRoute() *routev1.Route {
	return &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-developer",
			Labels: map[string]string{"app": system.Options.appLabel, "threescale_component": "system", "threescale_component_element": "developer-ui"},
		},
		Spec: routev1.RouteSpec{
			Host: system.Options.tenantName + "." + system.Options.wildcardDomain,
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "system-developer",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("http"),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyAllow},
		},
	}
}

func (system *System) buildSystemConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system",
			Labels: map[string]string{
				"app":                  system.Options.appLabel,
				"threescale_component": "system",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		Data: map[string]string{
			"zync.yml":              system.getSystemZyncConfData(),
			"rolling_updates.yml":   system.getSystemRollingUpdatesConfData(),
			"service_discovery.yml": system.getSystemServiceDiscoveryData(),
		},
	}
}

func (system *System) getSystemZyncConfData() string {
	return `production:
  endpoint: 'http://zync:8080'
  authentication:
    token: "<%= ENV.fetch('ZYNC_AUTHENTICATION_TOKEN') %>"
  connect_timeout: 5
  send_timeout: 5
  receive_timeout: 10
  root_url:
`
}

func (system *System) getSystemRollingUpdatesConfData() string {
	return `production:
  old_charts: false
  new_provider_documentation: false
  proxy_pro: false
  instant_bill_plan_change: false
  service_permissions: true
  async_apicast_deploy: false
  duplicate_application_id: true
  duplicate_user_key: true
  plan_changes_wizard: false
  require_cc_on_signup: false
  apicast_per_service: true
  new_notification_system: true
  cms_api: false
  apicast_v2: true
  forum: false
  published_service_plan_signup: true
  apicast_oidc: true
  policies: true
  policy_registry: true
  proxy_private_base_path: true
  service_mesh_integration: true
`
}

func (system *System) getSystemServiceDiscoveryData() string {
	return `production:
  enabled: <%= cluster_token_file_exists = File.exists?(cluster_token_file_path = '/var/run/secrets/kubernetes.io/serviceaccount/token') %>
  server_scheme: 'https'
  server_host: 'kubernetes.default.svc.cluster.local'
  server_port: 443
  bearer_token: "<%= File.read(cluster_token_file_path) if cluster_token_file_exists %>"
  authentication_method: service_account # can be service_account|oauth
  oauth_server_type: builtin # can be builtin|rh_sso
  client_id:
  client_secret:
  timeout: 1
  open_timeout: 1
  max_retry: 5
  verify_ssl: <%= OpenSSL::SSL::VERIFY_NONE %> # 0
`
}

func (system *System) buildSystemSphinxDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-sphinx",
			Labels: map[string]string{
				"app":                          system.Options.appLabel,
				"threescale_component":         "system",
				"threescale_component_element": "sphinx",
			},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnConfigChange,
				},
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							"system-master-svc",
							"system-sphinx",
						},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "amp-system:latest",
						},
					},
				},
			},
			Replicas: 1,
			Selector: map[string]string{"deploymentConfig": "system-sphinx"},
			Strategy: appsv1.DeploymentStrategy{
				RollingParams: &appsv1.RollingDeploymentStrategyParams{
					IntervalSeconds: &[]int64{1}[0],
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.Type(intstr.String),
						StrVal: "25%",
					},
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Type(intstr.String),
						StrVal: "25%",
					},
					TimeoutSeconds:      &[]int64{1200}[0],
					UpdatePeriodSeconds: &[]int64{1}[0],
				},
				Type: appsv1.DeploymentStrategyTypeRolling,
			},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                          system.Options.appLabel,
						"deploymentConfig":             "system-sphinx",
						"threescale_component":         "system",
						"threescale_component_element": "sphinx",
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "amp",
					InitContainers: []v1.Container{
						v1.Container{
							Name:    "system-master-svc",
							Image:   "amp-system:latest",
							Command: []string{"sh", "-c", "until $(curl --output /dev/null --silent --fail --head http://system-master:3000/status); do sleep $SLEEP_SECONDS; done"},
							Env: []v1.EnvVar{
								envVarFromValue("SLEEP_SECONDS", "1"),
							},
						},
					},
					Volumes: []v1.Volume{
						v1.Volume{
							Name: "system-sphinx-database",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{
									Medium: v1.StorageMediumDefault,
								},
							},
						},
					},
					Containers: []v1.Container{
						v1.Container{
							Name:            "system-sphinx",
							Image:           "amp-system:latest",
							ImagePullPolicy: v1.PullIfNotPresent,
							Args:            []string{"rake", "openshift:thinking_sphinx:start"},
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "system-sphinx-database",
									MountPath: "/opt/system/db/sphinx",
								},
							},
							Env: system.buildSystemSphinxEnv(),
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.FromInt(9306),
									},
								},
								InitialDelaySeconds: 60,
								PeriodSeconds:       10,
							},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1000m"),
									v1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("80m"),
									v1.ResourceMemory: resource.MustParse("250Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
}
