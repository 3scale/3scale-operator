package component

import (
	"context"
	"sort"
	"strconv"

	"sigs.k8s.io/controller-runtime/pkg/client"

	k8sappsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/3scale/3scale-operator/apis/apps"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
)

const (
	SystemSecretSystemDatabaseSecretName            = "system-database"
	SystemSecretSystemDatabaseURLFieldName          = "URL"
	SystemSecretSystemDatabaseDatabaseNameFieldName = "DB_NAME"
	SystemSecretSystemDatabaseUserFieldName         = "DB_USER"
	SystemSecretSystemDatabasePasswordFieldName     = "DB_PASSWORD"
	SystemSecretSystemDatabaseRootPasswordFieldName = "DB_ROOT_PASSWORD"
	SystemSecretDatabaseSslCa                       = "DATABASE_SSL_CA"
	SystemSecretDatabaseSslCert                     = "DATABASE_SSL_CERT"
	SystemSecretDatabaseSslKey                      = "DATABASE_SSL_KEY"
	SystemSecretDatabaseSslMode                     = "DATABASE_SSL_MODE"
	SystemSecretSslCa                               = "DB_SSL_CA"
	SystemSecretSslCert                             = "DB_SSL_CERT"
	SystemSecretSslKey                              = "DB_SSL_KEY"
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
	SystemSecretSystemRedisSecretName    = "system-redis"
	SystemSecretSystemRedisURLFieldName  = "URL"
	SystemSecretSystemRedisSentinelHosts = "SENTINEL_HOSTS"
	SystemSecretSystemRedisSentinelRole  = "SENTINEL_ROLE"

	// ACL
	SystemSecretSystemRedisUsernameFieldName         = "REDIS_USERNAME"
	SystemSecretSystemRedisPasswordFieldName         = "REDIS_PASSWORD"
	SystemSecretSystemRedisSentinelUsernameFieldName = "REDIS_SENTINEL_USERNAME"
	SystemSecretSystemRedisSentinelPasswordFieldName = "REDIS_SENTINEL_PASSWORD"
)

const (
	SystemSecretSystemAppSecretName              = "system-app"
	SystemSecretSystemAppSecretKeyBaseFieldName  = "SECRET_KEY_BASE"
	SystemSecretSystemAppUserSessionTTLFieldName = "USER_SESSION_TTL"
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
	SystemSecretSystemMasterApicastAccessToken                   = "ACCESS_TOKEN"
)

const (
	SystemSecretSystemSMTPSecretName                 = "system-smtp"
	SystemSecretSystemSMTPAddressFieldName           = "address"
	SystemSecretSystemSMTPUserNameFieldName          = "username"
	SystemSecretSystemSMTPPasswordFieldName          = "password"
	SystemSecretSystemSMTPDomainFieldName            = "domain"
	SystemSecretSystemSMTPPortFieldName              = "port"
	SystemSecretSystemSMTPAuthenticationFieldName    = "authentication"
	SystemSecretSystemSMTPOpenSSLVerifyModeFieldName = "openssl.verify.mode"
	SystemSecretSystemSMTPFromAddressFieldName       = "from_address"
)

const (
	SystemFileStoragePVCName = "system-storage"
)

const (
	SystemSidekiqName              = "system-sidekiq"
	SystemSideKiqInitContainerName = "check-svc"
	SystemAppDeploymentName        = "system-app"
	SystemAppPreHookJobName        = "system-app-pre"
	SystemAppPostHookJobName       = "system-app-post"

	SystemAppMasterContainerName    = "system-master"
	SystemAppProviderContainerName  = "system-provider"
	SystemAppDeveloperContainerName = "system-developer"
)

const (
	SystemAppPrometheusExporterPortEnvVarName     = "PROMETHEUS_EXPORTER_PORT"
	SystemSidekiqPrometheusExporterPortEnvVarName = "PROMETHEUS_EXPORTER_PORT"
	SystemSidekiqMetricsPort                      = 9394
	SystemAppMasterContainerPrometheusPort        = 9395
	SystemAppProviderContainerPrometheusPort      = 9396
	SystemAppDeveloperContainerPrometheusPort     = 9394
	SystemAppMasterContainerMetricsPortName       = "master-metrics"
	SystemAppProviderContainerMetricsPortName     = "prov-metrics"
	SystemAppDeveloperContainerMetricsPortName    = "dev-metrics"
)

const (
	SystemDatabaseSecretResverAnnotationPrefix = "apimanager.apps.3scale.net/systemdatabase-secret-resource-version-"
)

const (
	S3StsCredentialsSecretName = "s3-credentials"
)

const (
	redisCaFilePath            = "/tls/system-redis/system-redis-ca.crt"
	redisClientCertPath        = "/tls/system-redis/system-redis-client.crt"
	redisPrivateKeyPath        = "/tls/system-redis/system-redis-private.key"
	backendRedisCaFilePath     = "/tls/backend-redis/backend-redis-ca.crt"
	backendRedisClientCertPath = "/tls/backend-redis/backend-redis-client.crt"
	backendRedisPrivateKeyPath = "/tls/backend-redis/backend-redis-private.key"

	SystemRedisSecretResverAnnotation = "apimanager.apps.3scale.net/system-redis-secret-resource-version"
)

type System struct {
	Options *SystemOptions
}

func NewSystem(options *SystemOptions) *System {
	return &System{Options: options}
}

func (system *System) getSystemBaseEnvsFromEnvConfigMap() []v1.EnvVar {
	result := []v1.EnvVar{}

	// Add system-base-env ConfigMap values to envvar sources
	cfg := system.EnvironmentConfigMap()
	cfgmapkeys := make([]string, 0, len(cfg.Data))
	for key := range cfg.Data {
		cfgmapkeys = append(cfgmapkeys, key)
	}
	sort.Strings(cfgmapkeys)
	for _, key := range cfgmapkeys {
		envvar := helper.EnvVarFromConfigMap(key, "system-environment", key)
		result = append(result, envvar)
	}

	return result
}

func (system *System) getSystemSMTPEnvsFromSMTPSecret() []v1.EnvVar {
	result := []v1.EnvVar{
		helper.EnvVarFromSecret("SMTP_ADDRESS", SystemSecretSystemSMTPSecretName, SystemSecretSystemSMTPAddressFieldName),
		helper.EnvVarFromSecret("SMTP_USER_NAME", SystemSecretSystemSMTPSecretName, SystemSecretSystemSMTPUserNameFieldName),
		helper.EnvVarFromSecret("SMTP_PASSWORD", SystemSecretSystemSMTPSecretName, SystemSecretSystemSMTPPasswordFieldName),
		helper.EnvVarFromSecret("SMTP_DOMAIN", SystemSecretSystemSMTPSecretName, SystemSecretSystemSMTPDomainFieldName),
		helper.EnvVarFromSecret("SMTP_PORT", SystemSecretSystemSMTPSecretName, SystemSecretSystemSMTPPortFieldName),
		helper.EnvVarFromSecret("SMTP_AUTHENTICATION", SystemSecretSystemSMTPSecretName, SystemSecretSystemSMTPAuthenticationFieldName),
		helper.EnvVarFromSecret("SMTP_OPENSSL_VERIFY_MODE", SystemSecretSystemSMTPSecretName, SystemSecretSystemSMTPOpenSSLVerifyModeFieldName),
	}

	if system.Options.SmtpSecretOptions.FromAddress != nil &&
		*system.Options.SmtpSecretOptions.FromAddress != "" {
		result = append(result, helper.EnvVarFromSecret("NOREPLY_EMAIL", SystemSecretSystemSMTPSecretName, SystemSecretSystemSMTPFromAddressFieldName))
	}

	return result
}

func (system *System) SystemRedisEnvVars() []v1.EnvVar {
	result := []v1.EnvVar{}

	result = append(result,
		helper.EnvVarFromSecret("REDIS_URL", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisURLFieldName),
		helper.EnvVarFromSecret("REDIS_SENTINEL_HOSTS", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisSentinelHosts),
		helper.EnvVarFromSecret("REDIS_SENTINEL_ROLE", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisSentinelRole),
		// ACL
		helper.EnvVarFromSecretOptional("REDIS_USERNAME", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisUsernameFieldName),
		helper.EnvVarFromSecretOptional("REDIS_PASSWORD", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisPasswordFieldName),
		helper.EnvVarFromSecretOptional("REDIS_SENTINEL_USERNAME", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisSentinelUsernameFieldName),
		helper.EnvVarFromSecretOptional("REDIS_SENTINEL_PASSWORD", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisSentinelPasswordFieldName),
	)
	if system.Options.RedisTLSEnabled {
		result = append(result, system.SystemRedisTLSEnvVars()...)
		result = append(result, system.BackendRedisTLSEnvVars()...)
	} else {
		result = append(result, helper.EnvVarFromValue("REDIS_SSL", "0"))
	}

	return result
}

func (system *System) buildSystemBaseEnv() []v1.EnvVar {
	result := []v1.EnvVar{}
	baseEnvConfigMapEnvs := system.getSystemBaseEnvsFromEnvConfigMap()
	result = append(result, baseEnvConfigMapEnvs...)

	result = append(result,
		helper.EnvVarFromSecret("DATABASE_URL", SystemSecretSystemDatabaseSecretName, SystemSecretSystemDatabaseURLFieldName),

		helper.EnvVarFromSecret("MASTER_DOMAIN", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedMasterDomainFieldName),
		helper.EnvVarFromSecret("MASTER_USER", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedMasterUserFieldName),
		helper.EnvVarFromSecret("MASTER_PASSWORD", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedMasterPasswordFieldName),

		helper.EnvVarFromSecret("ADMIN_ACCESS_TOKEN", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedAdminAccessTokenFieldName),
		helper.EnvVarFromSecret("USER_LOGIN", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedAdminUserFieldName),
		helper.EnvVarFromSecret("USER_PASSWORD", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedAdminPasswordFieldName),
		helper.EnvVarFromSecret("USER_EMAIL", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedAdminEmailFieldName),
		helper.EnvVarFromSecret("TENANT_NAME", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedTenantNameFieldName),

		helper.EnvVarFromValue("THINKING_SPHINX_ADDRESS", SystemSearchdServiceName),

		helper.EnvVarFromSecret("EVENTS_SHARED_SECRET", SystemSecretSystemEventsHookSecretName, SystemSecretSystemEventsHookPasswordFieldName),

		helper.EnvVarFromSecret("RECAPTCHA_PUBLIC_KEY", SystemSecretSystemRecaptchaSecretName, SystemSecretSystemRecaptchaPublicKeyFieldName),
		helper.EnvVarFromSecret("RECAPTCHA_PRIVATE_KEY", SystemSecretSystemRecaptchaSecretName, SystemSecretSystemRecaptchaPrivateKeyFieldName),

		helper.EnvVarFromSecret("SECRET_KEY_BASE", SystemSecretSystemAppSecretName, SystemSecretSystemAppSecretKeyBaseFieldName),

		helper.EnvVarFromSecret("MEMCACHE_SERVERS", SystemSecretSystemMemcachedSecretName, SystemSecretSystemMemcachedServersFieldName),
	)

	if system.Options.SystemDbTLSEnabled {
		result = append(result,
			helper.EnvVarFromSecretOptional("DB_SSL_CA", SystemSecretSystemDatabaseSecretName, SystemSecretSslCa),
			helper.EnvVarFromSecretOptional("DB_SSL_CERT", SystemSecretSystemDatabaseSecretName, SystemSecretSslCert),
			helper.EnvVarFromSecretOptional("DB_SSL_KEY", SystemSecretSystemDatabaseSecretName, SystemSecretSslKey),
			helper.EnvVarFromSecretOptional("DATABASE_SSL_MODE", SystemSecretSystemDatabaseSecretName, SystemSecretDatabaseSslMode),
			helper.EnvVarFromValue("DATABASE_SSL_CA", helper.TlsCertPresent("DATABASE_SSL_CA", SystemSecretSystemDatabaseSecretName, system.Options.SystemDbTLSEnabled)),
			helper.EnvVarFromValue("DATABASE_SSL_CERT", helper.TlsCertPresent("DATABASE_SSL_CERT", SystemSecretSystemDatabaseSecretName, system.Options.SystemDbTLSEnabled)),
			helper.EnvVarFromValue("DATABASE_SSL_KEY", helper.TlsCertPresent("DATABASE_SSL_KEY", SystemSecretSystemDatabaseSecretName, system.Options.SystemDbTLSEnabled)),
		)
	}

	result = append(result, system.SystemRedisEnvVars()...)
	result = append(result, system.BackendRedisEnvVars()...)
	bckServiceEndpointEnv := helper.EnvVarFromSecret("BACKEND_URL", BackendSecretBackendListenerSecretName, BackendSecretBackendListenerServiceEndpointFieldName)
	bckPublicRouteEndpointEnv := helper.EnvVarFromSecret("BACKEND_PUBLIC_URL", BackendSecretBackendListenerSecretName, BackendSecretBackendListenerRouteEndpointFieldName)
	result = append(result, bckServiceEndpointEnv, bckPublicRouteEndpointEnv)

	smtpEnvSecretEnvs := system.getSystemSMTPEnvsFromSMTPSecret()
	result = append(result, smtpEnvSecretEnvs...)

	apicastAccessToken := helper.EnvVarFromSecret("APICAST_ACCESS_TOKEN", SystemSecretSystemMasterApicastSecretName, "ACCESS_TOKEN")
	result = append(result, apicastAccessToken)

	// Add zync secret to envvars sources
	if system.Options.ZyncEnabled {
		zyncAuthTokenVar := helper.EnvVarFromSecret("ZYNC_AUTHENTICATION_TOKEN", "zync", "ZYNC_AUTHENTICATION_TOKEN")
		result = append(result, zyncAuthTokenVar)
	}

	// Add backend internal api data to envvars sources
	systemBackendInternalAPIUser := helper.EnvVarFromSecret("CONFIG_INTERNAL_API_USER", "backend-internal-api", "username")
	systemBackendInternalAPIPass := helper.EnvVarFromSecret("CONFIG_INTERNAL_API_PASSWORD", "backend-internal-api", "password")
	result = append(result, systemBackendInternalAPIUser, systemBackendInternalAPIPass)

	if system.Options.S3FileStorageOptions != nil {
		result = append(result,
			helper.EnvVarFromSecret(apps.AwsBucket, system.Options.S3FileStorageOptions.ConfigurationSecretName, apps.AwsBucket),
			helper.EnvVarFromSecret(apps.AwsRegion, system.Options.S3FileStorageOptions.ConfigurationSecretName, apps.AwsRegion),
			helper.EnvVarFromSecretOptional(apps.AwsProtocol, system.Options.S3FileStorageOptions.ConfigurationSecretName, apps.AwsProtocol),
			helper.EnvVarFromSecretOptional(apps.AwsHostname, system.Options.S3FileStorageOptions.ConfigurationSecretName, apps.AwsHostname),
			helper.EnvVarFromSecretOptional(apps.AwsPathStyle, system.Options.S3FileStorageOptions.ConfigurationSecretName, apps.AwsPathStyle),
		)

		if system.Options.S3FileStorageOptions.STSEnabled {
			result = append(result,
				helper.EnvVarFromSecret(apps.AwsRoleArn, system.Options.S3FileStorageOptions.ConfigurationSecretName, apps.AwsRoleArn),
				helper.EnvVarFromSecret(apps.AwsWebIdentityTokenFile, system.Options.S3FileStorageOptions.ConfigurationSecretName, apps.AwsWebIdentityTokenFile),
			)
		} else {
			result = append(result,
				helper.EnvVarFromSecret(apps.AwsAccessKeyID, system.Options.S3FileStorageOptions.ConfigurationSecretName, apps.AwsAccessKeyID),
				helper.EnvVarFromSecret(apps.AwsSecretAccessKey, system.Options.S3FileStorageOptions.ConfigurationSecretName, apps.AwsSecretAccessKey),
			)
		}
	}

	return result
}

func (system *System) buildAppEnv() []v1.EnvVar {
	result := []v1.EnvVar{}
	result = append(result, helper.EnvVarFromSecret(SystemSecretSystemAppUserSessionTTLFieldName, SystemSecretSystemAppSecretName, SystemSecretSystemAppUserSessionTTLFieldName))
	return result
}

func (system *System) buildAppMasterContainerEnv() []v1.EnvVar {
	result := system.buildSystemBaseEnv()
	if system.Options.AppMetrics {
		result = append(result, helper.EnvVarFromValue(SystemAppPrometheusExporterPortEnvVarName, strconv.Itoa(SystemAppMasterContainerPrometheusPort)))
	}
	result = append(result, system.buildAppEnv()...)

	return result
}

func (system *System) buildAppProviderContainerEnv() []v1.EnvVar {
	result := system.buildSystemBaseEnv()
	if system.Options.AppMetrics {
		result = append(result, helper.EnvVarFromValue(SystemAppPrometheusExporterPortEnvVarName, strconv.Itoa(SystemAppProviderContainerPrometheusPort)))
	}
	result = append(result, system.buildAppEnv()...)

	return result
}

func (system *System) buildAppDeveloperContainerEnv() []v1.EnvVar {
	result := system.buildSystemBaseEnv()
	if system.Options.AppMetrics {
		result = append(result, helper.EnvVarFromValue(SystemAppPrometheusExporterPortEnvVarName, strconv.Itoa(SystemAppDeveloperContainerPrometheusPort)))
	}
	result = append(result, system.buildAppEnv()...)

	return result
}

func (system *System) buildSystemSidekiqContainerEnv() []v1.EnvVar {
	result := system.buildSystemBaseEnv()
	if system.Options.SideKiqMetrics {
		result = append(result, helper.EnvVarFromValue(SystemSidekiqPrometheusExporterPortEnvVarName, strconv.Itoa(SystemSidekiqMetricsPort)))
	}

	return result
}

func (system *System) buildSystemAppPreHookEnv() []v1.EnvVar {
	result := []v1.EnvVar{}
	baseEnv := system.buildSystemBaseEnv()
	result = append(result, baseEnv...)
	result = append(result,
		helper.EnvVarFromSecret("MASTER_ACCESS_TOKEN", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedMasterAccessTokenFieldName),
	)
	if system.Options.IncludeOracleOptionalSettings {
		result = append(result, helper.EnvVarFromSecretOptional("ORACLE_SYSTEM_PASSWORD", SystemSecretSystemDatabaseSecretName, "ORACLE_SYSTEM_PASSWORD"))
	}
	return result
}

func (system *System) buildSystemAppPostHookEnv() []v1.EnvVar {
	result := []v1.EnvVar{}
	baseEnv := system.buildSystemBaseEnv()
	result = append(result, baseEnv...)
	if system.Options.IncludeOracleOptionalSettings {
		result = append(result, helper.EnvVarFromSecretOptional("ORACLE_SYSTEM_PASSWORD", SystemSecretSystemDatabaseSecretName, "ORACLE_SYSTEM_PASSWORD"))
	}
	return result
}

func (system *System) BackendRedisEnvVars() []v1.EnvVar {
	result := []v1.EnvVar{}
	result = append(result,
		helper.EnvVarFromSecret("BACKEND_REDIS_URL", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageURLFieldName),
		helper.EnvVarFromSecret("BACKEND_REDIS_SENTINEL_HOSTS", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageSentinelHostsFieldName),
		helper.EnvVarFromSecret("BACKEND_REDIS_SENTINEL_ROLE", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageSentinelRoleFieldName),
		helper.EnvVarFromSecretOptional("BACKEND_REDIS_USERNAME", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageUsernameFieldName),
		helper.EnvVarFromSecretOptional("BACKEND_REDIS_PASSWORD", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStoragePasswordFieldName),
		helper.EnvVarFromSecretOptional("BACKEND_REDIS_SENTINEL_USERNAME", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageSentinelUsernameFieldName),
		helper.EnvVarFromSecretOptional("BACKEND_REDIS_SENTINEL_PASSWORD", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageSentinelPasswordFieldName),
	)
	if system.Options.RedisTLSEnabled {
		result = append(result, system.BackendRedisTLSEnvVars()...)
	} else {
		result = append(result, helper.EnvVarFromValue("BACKEND_REDIS_SSL", "0"))
	}
	return result
}

func (system *System) EnvironmentConfigMap() *v1.ConfigMap {
	res := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-environment",
			Labels: system.Options.CommonLabels,
		},
		Data: map[string]string{
			"RAILS_ENV":              "production",
			"FORCE_SSL":              "true",
			"THREESCALE_SUPERDOMAIN": system.Options.WildcardDomain,
			"PROVIDER_PLAN":          "enterprise",
			"APICAST_REGISTRY_URL":   system.Options.ApicastRegistryURL,
			"RAILS_LOG_TO_STDOUT":    "true",
			"RAILS_LOG_LEVEL":        "info",
			"THINKING_SPHINX_PORT":   "9306",
			"THREESCALE_SANDBOX_PROXY_OPENSSL_VERIFY_MODE": "VERIFY_NONE",
			"SSL_CERT_DIR": "/etc/pki/tls/certs",
		},
	}

	if system.Options.S3FileStorageOptions != nil {
		res.Data["FILE_UPLOAD_STORAGE"] = "s3"
	}

	return res
}

func (system *System) MemcachedSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSecretSystemMemcachedSecretName,
			Labels: system.Options.CommonLabels,
		},
		StringData: map[string]string{
			SystemSecretSystemMemcachedServersFieldName: system.Options.MemcachedServers,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (system *System) RecaptchaSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSecretSystemRecaptchaSecretName,
			Labels: system.Options.CommonLabels,
		},
		StringData: map[string]string{
			SystemSecretSystemRecaptchaPublicKeyFieldName:  *system.Options.RecaptchaPublicKey,
			SystemSecretSystemRecaptchaPrivateKeyFieldName: *system.Options.RecaptchaPrivateKey,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (system *System) EventsHookSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSecretSystemEventsHookSecretName,
			Labels: system.Options.CommonLabels,
		},
		StringData: map[string]string{
			SystemSecretSystemEventsHookURLFieldName:      system.Options.EventHooksURL,
			SystemSecretSystemEventsHookPasswordFieldName: system.Options.BackendSharedSecret,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (system *System) AppSecret() *v1.Secret {
	result := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSecretSystemAppSecretName, // TODO sure this should be a secret on its own?? maybe can join different secrets into one with more values?
			Labels: system.Options.CommonLabels,
		},
		StringData: map[string]string{
			SystemSecretSystemAppSecretKeyBaseFieldName:  system.Options.AppSecretKeyBase,
			SystemSecretSystemAppUserSessionTTLFieldName: *system.Options.UserSessionTTL,
		},
		Type: v1.SecretTypeOpaque,
	}

	return result
}

func (system *System) SeedSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSecretSystemSeedSecretName,
			Labels: system.Options.CommonLabels,
		},
		StringData: map[string]string{
			SystemSecretSystemSeedMasterDomainFieldName:      system.Options.MasterName,
			SystemSecretSystemSeedMasterAccessTokenFieldName: system.Options.MasterAccessToken,
			SystemSecretSystemSeedMasterUserFieldName:        system.Options.MasterUsername,
			SystemSecretSystemSeedMasterPasswordFieldName:    system.Options.MasterPassword,
			SystemSecretSystemSeedAdminAccessTokenFieldName:  system.Options.AdminAccessToken,
			SystemSecretSystemSeedAdminUserFieldName:         system.Options.AdminUsername,
			SystemSecretSystemSeedAdminPasswordFieldName:     system.Options.AdminPassword,
			SystemSecretSystemSeedAdminEmailFieldName:        *system.Options.AdminEmail,
			SystemSecretSystemSeedTenantNameFieldName:        system.Options.TenantName,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (system *System) MasterApicastSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSecretSystemMasterApicastSecretName,
			Labels: system.Options.CommonLabels,
		},
		StringData: map[string]string{
			SystemSecretSystemMasterApicastProxyConfigsEndpointFieldName: system.Options.ApicastSystemMasterProxyConfigEndpoint,
			SystemSecretSystemMasterApicastAccessToken:                   system.Options.ApicastAccessToken,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (system *System) appPodVolumes() []v1.Volume {
	res := []v1.Volume{}
	if system.Options.PvcFileStorageOptions != nil {
		res = append(res, system.FileStorageVolume())
	}

	systemConfigVolume := v1.Volume{
		Name: "system-config",
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "system",
				},
				Items: []v1.KeyToPath{
					{
						Key:  "zync.yml",
						Path: "zync.yml",
					},
					{
						Key:  "rolling_updates.yml",
						Path: "rolling_updates.yml",
					},
					{
						Key:  "service_discovery.yml",
						Path: "service_discovery.yml",
					},
				},
			},
		},
	}

	res = append(res, systemConfigVolume)
	if system.Options.SystemDbTLSEnabled {
		systemTlsVolume := v1.Volume{
			Name: "tls-secret",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: SystemSecretSystemDatabaseSecretName, // Name of the secret containing the TLS certs
					Items: []v1.KeyToPath{
						{
							Key:  SystemSecretSslCa,
							Path: "ca.crt", // Map the secret key to the ca.crt file in the container
						},
						{
							Key:  SystemSecretSslCert,
							Path: "tls.crt", // Map the secret key to the tls.crt file in the container
						},
						{
							Key:  SystemSecretSslKey,
							Path: "tls.key", // Map the secret key to the tls.key file in the container
						},
					},
				},
			},
		}
		res = append(res, systemTlsVolume)

		systemWritableTlsVolume := v1.Volume{
			Name: "writable-tls",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		}
		res = append(res, systemWritableTlsVolume)
	}

	if system.Options.RedisTLSEnabled {
		systemRedisTlsVolume := v1.Volume{
			Name: "system-redis-tls",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: SystemSecretSystemRedisSecretName, // Name of the secret containing the TLS certs
					Items: []v1.KeyToPath{
						{
							Key:  "REDIS_SSL_CA",
							Path: "system-redis-ca.crt",
						},
						{
							Key:  "REDIS_SSL_CERT",
							Path: "system-redis-client.crt",
						},
						{
							Key:  "REDIS_SSL_KEY",
							Path: "system-redis-private.key",
						},
					},
				},
			},
		}
		res = append(res, systemRedisTlsVolume)

		backendRedisTlsVolume := v1.Volume{
			Name: "backend-redis-tls",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: BackendSecretBackendRedisSecretName, // Name of the secret containing the TLS certs
					Items: []v1.KeyToPath{
						{
							Key:  "REDIS_SSL_CA",
							Path: "backend-redis-ca.crt",
						},
						{
							Key:  "REDIS_SSL_CERT",
							Path: "backend-redis-client.crt",
						},
						{
							Key:  "REDIS_SSL_KEY",
							Path: "backend-redis-private.key",
						},
					},
				},
			},
		}
		res = append(res, backendRedisTlsVolume)
	}

	if system.Options.S3FileStorageOptions != nil && system.Options.S3FileStorageOptions.STSEnabled {
		s3CredsProjectedVolume := v1.Volume{
			Name: S3StsCredentialsSecretName,
			VolumeSource: v1.VolumeSource{
				Projected: &v1.ProjectedVolumeSource{
					Sources: []v1.VolumeProjection{
						{
							ServiceAccountToken: &v1.ServiceAccountTokenProjection{
								Audience:          system.Options.S3FileStorageOptions.STSAudience,
								ExpirationSeconds: &[]int64{3600}[0],
								Path:              system.Options.S3FileStorageOptions.STSTokenMountRelativePath,
							},
						},
					},
				},
			},
		}
		res = append(res, s3CredsProjectedVolume)
	}

	return res
}

func (system *System) AppDeployment(ctx context.Context, k8sclient client.Client, containerImage string) (*k8sappsv1.Deployment, error) {
	watchedSecretAnnotations, err := ComputeWatchedSecretAnnotations(ctx, k8sclient, SystemAppDeploymentName, system.Options.Namespace, system)
	if err != nil {
		return nil, err
	}

	return &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: reconcilers.DeploymentAPIVersion, Kind: reconcilers.DeploymentKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemAppDeploymentName,
			Labels: system.Options.CommonAppLabels,
		},
		Spec: k8sappsv1.DeploymentSpec{
			Strategy: k8sappsv1.DeploymentStrategy{
				Type: k8sappsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &k8sappsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "25%",
					},
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "25%",
					},
				},
			},
			MinReadySeconds: 0,
			Replicas:        &system.Options.AppReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					reconcilers.DeploymentLabelSelector: SystemAppDeploymentName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      system.Options.AppPodTemplateLabels,
					Annotations: system.appPodAnnotations(watchedSecretAnnotations),
				},
				Spec: v1.PodSpec{
					Affinity:       system.Options.AppAffinity,
					Tolerations:    system.Options.AppTolerations,
					Volumes:        system.appPodVolumes(),
					InitContainers: system.systemInit(containerImage),
					Containers: []v1.Container{
						{
							Name:         SystemAppMasterContainerName,
							Image:        containerImage,
							Args:         []string{"env", "TENANT_MODE=master", "PORT=3002", "container-entrypoint", "bundle", "exec", "unicorn", "-c", "config/unicorn.rb"},
							Ports:        system.appMasterPorts(),
							Env:          system.buildAppMasterContainerEnv(),
							Resources:    *system.Options.AppMasterContainerResourceRequirements,
							VolumeMounts: system.appMasterContainerVolumeMounts(),
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.IntOrString{
											Type:   intstr.String,
											StrVal: "master",
										},
									},
								},
								InitialDelaySeconds: 40,
								TimeoutSeconds:      10,
								PeriodSeconds:       10,
								SuccessThreshold:    0,
								FailureThreshold:    40,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path: "/check.txt",
										Port: intstr.IntOrString{
											Type:   intstr.String,
											StrVal: "master",
										},
										Scheme: v1.URISchemeHTTP,
										HTTPHeaders: []v1.HTTPHeader{
											{
												Name:  "X-Forwarded-Proto",
												Value: "https",
											},
										},
									},
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
						},
						{
							Name:         SystemAppProviderContainerName,
							Image:        containerImage,
							Args:         []string{"env", "TENANT_MODE=provider", "PORT=3000", "container-entrypoint", "bundle", "exec", "unicorn", "-c", "config/unicorn.rb"},
							Ports:        system.appProviderPorts(),
							Env:          system.buildAppProviderContainerEnv(),
							Resources:    *system.Options.AppProviderContainerResourceRequirements,
							VolumeMounts: system.appProviderContainerVolumeMounts(),
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.IntOrString{
											Type:   intstr.String,
											StrVal: "provider",
										},
									},
								},
								InitialDelaySeconds: 40,
								TimeoutSeconds:      10,
								PeriodSeconds:       10,
								SuccessThreshold:    0,
								FailureThreshold:    40,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path: "/check.txt",
										Port: intstr.IntOrString{
											Type:   intstr.String,
											StrVal: "provider",
										},
										Scheme: v1.URISchemeHTTP,
										HTTPHeaders: []v1.HTTPHeader{
											{
												Name:  "X-Forwarded-Proto",
												Value: "https",
											},
										},
									},
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
						},
						{
							Name:         SystemAppDeveloperContainerName,
							Image:        containerImage,
							Args:         []string{"env", "PORT=3001", "container-entrypoint", "bundle", "exec", "unicorn", "-c", "config/unicorn.rb"},
							Ports:        system.appDeveloperPorts(),
							Env:          system.buildAppDeveloperContainerEnv(),
							Resources:    *system.Options.AppDeveloperContainerResourceRequirements,
							VolumeMounts: system.appDeveloperContainerVolumeMounts(),
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.IntOrString{
											Type:   intstr.String,
											StrVal: "developer",
										},
									},
								},
								InitialDelaySeconds: 40,
								TimeoutSeconds:      10,
								PeriodSeconds:       10,
								SuccessThreshold:    0,
								FailureThreshold:    40,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path: "/check.txt",
										Port: intstr.IntOrString{
											Type:   intstr.String,
											StrVal: "developer",
										},
										Scheme: v1.URISchemeHTTP,
										HTTPHeaders: []v1.HTTPHeader{
											{
												Name:  "X-Forwarded-Proto",
												Value: "https",
											},
										},
									},
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
					ServiceAccountName:        "amp",
					PriorityClassName:         system.Options.AppPriorityClassName,
					TopologySpreadConstraints: system.Options.AppTopologySpreadConstraints,
				},
			},
		},
	}, nil
}

func (system *System) AppPreHookJob(containerImage string, namespace string, currentSystemAppRevision int64) *batchv1.Job {
	var completions int32 = 1

	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      SystemAppPreHookJobName,
			Namespace: namespace,
			Labels:    system.Options.CommonAppLabels,
			Annotations: map[string]string{
				helper.SystemAppRevisionAnnotation: strconv.FormatInt(currentSystemAppRevision, 10),
			},
		},
		Spec: batchv1.JobSpec{
			Completions: &completions,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Volumes:        system.appPodVolumes(),
					InitContainers: system.systemInit(containerImage),
					Containers: []v1.Container{
						{
							Name:            SystemAppPreHookJobName,
							Image:           containerImage,
							Args:            []string{"bash", "-c", "bundle exec rake boot openshift:deploy"},
							Env:             system.buildSystemAppPreHookEnv(),
							Resources:       *system.Options.AppMasterContainerResourceRequirements,
							VolumeMounts:    system.appMasterContainerVolumeMounts(),
							ImagePullPolicy: v1.PullIfNotPresent,
						},
					},
					RestartPolicy:      v1.RestartPolicyNever,
					ServiceAccountName: "amp",
					PriorityClassName:  system.Options.AppPriorityClassName,
				},
			},
		},
	}
}

func (system *System) AppPostHookJob(containerImage string, namespace string, currentSystemAppRevision int64) *batchv1.Job {
	var completions int32 = 1

	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      SystemAppPostHookJobName,
			Namespace: namespace,
			Labels:    system.Options.CommonAppLabels,
			Annotations: map[string]string{
				helper.SystemAppRevisionAnnotation: strconv.FormatInt(currentSystemAppRevision, 10),
			},
		},
		Spec: batchv1.JobSpec{
			Completions: &completions,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Volumes:        system.appPodVolumes(),
					InitContainers: system.systemInit(containerImage),
					Containers: []v1.Container{
						{
							Name:            SystemAppPostHookJobName,
							Image:           containerImage,
							Args:            []string{"bash", "-c", "bundle exec rake boot openshift:post_deploy"},
							Env:             system.buildSystemAppPostHookEnv(),
							Resources:       *system.Options.AppMasterContainerResourceRequirements,
							VolumeMounts:    system.appMasterContainerVolumeMounts(),
							ImagePullPolicy: v1.PullIfNotPresent,
						},
					},
					RestartPolicy:      v1.RestartPolicyNever,
					ServiceAccountName: "amp",
					PriorityClassName:  system.Options.AppPriorityClassName,
				},
			},
		},
	}
}

func (system *System) FileStorageVolume() v1.Volume {
	return v1.Volume{
		Name: SystemFileStoragePVCName,
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: SystemFileStoragePVCName,
				ReadOnly:  false,
			},
		},
	}
}

func (system *System) SidekiqPodVolumes() []v1.Volume {
	res := []v1.Volume{}
	systemTmpVolume := v1.Volume{
		Name: "system-tmp",
		VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{
			Medium: v1.StorageMediumMemory,
		}},
	}

	res = append(res, systemTmpVolume)
	if system.Options.PvcFileStorageOptions != nil {
		res = append(res, system.FileStorageVolume())
	}

	systemConfigVolume := v1.Volume{
		Name: "system-config",
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "system",
				},
				Items: []v1.KeyToPath{
					{
						Key:  "zync.yml",
						Path: "zync.yml",
					},
					{
						Key:  "rolling_updates.yml",
						Path: "rolling_updates.yml",
					},
					{
						Key:  "service_discovery.yml",
						Path: "service_discovery.yml",
					},
				},
			},
		},
	}

	res = append(res, systemConfigVolume)

	if system.Options.SystemDbTLSEnabled {
		systemTlsVolume := v1.Volume{
			Name: "tls-secret",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: SystemSecretSystemDatabaseSecretName, // Name of the secret containing the TLS certs
					Items: []v1.KeyToPath{
						{
							Key:  SystemSecretSslCa,
							Path: "ca.crt", // Map the secret key to the ca.crt file in the container
						},
						{
							Key:  SystemSecretSslCert,
							Path: "tls.crt", // Map the secret key to the tls.crt file in the container
						},
						{
							Key:  SystemSecretSslKey,
							Path: "tls.key", // Map the secret key to the tls.key file in the container
						},
					},
				},
			},
		}
		res = append(res, systemTlsVolume)

		systemWritableTlsVolume := v1.Volume{
			Name: "writable-tls",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		}
		res = append(res, systemWritableTlsVolume)
	}
	if system.Options.S3FileStorageOptions != nil && system.Options.S3FileStorageOptions.STSEnabled {
		s3CredsProjectedVolume := v1.Volume{
			Name: S3StsCredentialsSecretName,
			VolumeSource: v1.VolumeSource{
				Projected: &v1.ProjectedVolumeSource{
					Sources: []v1.VolumeProjection{
						{
							ServiceAccountToken: &v1.ServiceAccountTokenProjection{
								Audience:          system.Options.S3FileStorageOptions.STSAudience,
								ExpirationSeconds: &[]int64{3600}[0],
								Path:              system.Options.S3FileStorageOptions.STSTokenMountRelativePath,
							},
						},
					},
				},
			},
		}
		res = append(res, s3CredsProjectedVolume)
	}
	if system.Options.RedisTLSEnabled {
		systemRedisTlsVolume := v1.Volume{
			Name: "system-redis-tls",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: SystemSecretSystemRedisSecretName,
					Items: []v1.KeyToPath{
						{
							Key:  "REDIS_SSL_CA",
							Path: "system-redis-ca.crt",
						},
						{
							Key:  "REDIS_SSL_CERT",
							Path: "system-redis-client.crt",
						},
						{
							Key:  "REDIS_SSL_KEY",
							Path: "system-redis-private.key",
						},
					},
				},
			},
		}
		res = append(res, systemRedisTlsVolume)

		backendRedisTlsVolume := v1.Volume{
			Name: "backend-redis-tls",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: BackendSecretBackendRedisSecretName,
					Items: []v1.KeyToPath{
						{
							Key:  "REDIS_SSL_CA",
							Path: "backend-redis-ca.crt",
						},
						{
							Key:  "REDIS_SSL_CERT",
							Path: "backend-redis-client.crt",
						},
						{
							Key:  "REDIS_SSL_KEY",
							Path: "backend-redis-private.key",
						},
					},
				},
			},
		}
		res = append(res, backendRedisTlsVolume)
	}
	return res
}

func (system *System) SidekiqDeployment(ctx context.Context, k8sclient client.Client, containerImage string) (*k8sappsv1.Deployment, error) {
	watchedSecretAnnotations, err := ComputeWatchedSecretAnnotations(ctx, k8sclient, SystemSidekiqName, system.Options.Namespace, system)
	if err != nil {
		return nil, err
	}

	return &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: reconcilers.DeploymentAPIVersion, Kind: reconcilers.DeploymentKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSidekiqName,
			Labels: system.Options.CommonSidekiqLabels,
		},
		Spec: k8sappsv1.DeploymentSpec{
			Strategy: k8sappsv1.DeploymentStrategy{
				Type: k8sappsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &k8sappsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "25%",
					},
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "25%",
					},
				},
			},
			MinReadySeconds: 0,
			Replicas:        &system.Options.SidekiqReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					reconcilers.DeploymentLabelSelector: SystemSidekiqName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      system.Options.SidekiqPodTemplateLabels,
					Annotations: system.sidekiqPodAnnotations(watchedSecretAnnotations),
				},
				Spec: v1.PodSpec{
					Affinity:       system.Options.SidekiqAffinity,
					Tolerations:    system.Options.SidekiqTolerations,
					Volumes:        system.SidekiqPodVolumes(),
					InitContainers: system.sidekiqInit(containerImage),
					Containers: []v1.Container{
						{
							Name:            SystemSidekiqName,
							Image:           containerImage,
							Args:            []string{"rake", "sidekiq:worker", "RAILS_MAX_THREADS=25"},
							Env:             system.buildSystemSidekiqContainerEnv(),
							Resources:       *system.Options.SidekiqContainerResourceRequirements,
							VolumeMounts:    system.sidekiqContainerVolumeMounts(),
							ImagePullPolicy: v1.PullIfNotPresent,
							Ports:           system.sideKiqPorts(),
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.FromInt32(9394),
									},
								},
								FailureThreshold:    40,
								InitialDelaySeconds: 30,
								PeriodSeconds:       30,
								SuccessThreshold:    1,
								TimeoutSeconds:      30,
							},
						},
					},
					ServiceAccountName:        "amp",
					PriorityClassName:         system.Options.SideKiqPriorityClassName,
					TopologySpreadConstraints: system.Options.SideKiqTopologySpreadConstraints,
				},
			},
		},
	}, nil
}

func (system *System) systemStorageVolumeMount(readOnly bool) v1.VolumeMount {
	return v1.VolumeMount{
		Name:      SystemFileStoragePVCName,
		ReadOnly:  readOnly,
		MountPath: "/opt/system/public/system",
	}
}

func (system *System) systemConfigVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      "system-config",
		ReadOnly:  false,
		MountPath: "/opt/system-extra-configs",
	}
}

func (system *System) systemTlsVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      "writable-tls",
		ReadOnly:  false,
		MountPath: "/tls",
	}
}

func (system *System) s3CredsProjectedVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      S3StsCredentialsSecretName,
		ReadOnly:  true,
		MountPath: system.Options.S3FileStorageOptions.STSTokenMountPath,
	}
}

func (system *System) appCommonContainerVolumeMounts(systemStorageReadonly bool) []v1.VolumeMount {
	res := []v1.VolumeMount{}
	if system.Options.PvcFileStorageOptions != nil {
		res = append(res, system.systemStorageVolumeMount(systemStorageReadonly))
	}

	if system.Options.S3FileStorageOptions != nil && system.Options.S3FileStorageOptions.STSEnabled {
		res = append(res, system.s3CredsProjectedVolumeMount())
	}

	res = append(res, system.systemConfigVolumeMount())
	if system.Options.RedisTLSEnabled {
		res = append(res, system.systemRedisTlsVolumeMount())
		res = append(res, system.backendRedisTlsVolumeMount())
	}

	if system.Options.SystemDbTLSEnabled {
		res = append(res, system.systemTlsVolumeMount())
	}

	return res
}

func (system *System) appMasterContainerVolumeMounts() []v1.VolumeMount {
	return system.appCommonContainerVolumeMounts(false)
}

func (system *System) appProviderContainerVolumeMounts() []v1.VolumeMount {
	return system.appCommonContainerVolumeMounts(false)
}

func (system *System) appDeveloperContainerVolumeMounts() []v1.VolumeMount {
	// TODO why system-app developer container has the system-config volume set to true? is it really necessary?
	// other containers in the same pod have it to false
	return system.appCommonContainerVolumeMounts(true)
}

func (system *System) sidekiqContainerVolumeMounts() []v1.VolumeMount {
	res := []v1.VolumeMount{}
	if system.Options.PvcFileStorageOptions != nil {
		res = append(res, system.systemStorageVolumeMount(false))
	}
	systemTmpVolumeMount := v1.VolumeMount{
		Name:      "system-tmp",
		ReadOnly:  false,
		MountPath: "/tmp",
	}
	res = append(res, systemTmpVolumeMount)
	res = append(res, system.systemConfigVolumeMount())
	if system.Options.SystemDbTLSEnabled {
		res = append(res, system.systemTlsVolumeMount())
	}
	if system.Options.RedisTLSEnabled {
		res = append(res, system.systemRedisTlsVolumeMount())
		res = append(res, system.backendRedisTlsVolumeMount())
	}
	if system.Options.S3FileStorageOptions != nil && system.Options.S3FileStorageOptions.STSEnabled {
		res = append(res, system.s3CredsProjectedVolumeMount())
	}

	return res
}

func (system *System) SharedStorage() *v1.PersistentVolumeClaim {
	volName := ""
	if system.Options.PvcFileStorageOptions.VolumeName != nil {
		volName = *system.Options.PvcFileStorageOptions.VolumeName
	}

	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-storage",
			Labels: system.Options.CommonAppLabels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			StorageClassName: system.Options.PvcFileStorageOptions.StorageClass,
			VolumeName:       volName,
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteMany,
			},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: system.Options.PvcFileStorageOptions.StorageRequests,
				},
			},
		},
	}
}

func (system *System) ProviderService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-provider",
			Labels: system.Options.ProviderUILabels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "http",
					Protocol:   v1.ProtocolTCP,
					Port:       3000,
					TargetPort: intstr.FromString("provider"),
				},
			},
			Selector: map[string]string{reconcilers.DeploymentLabelSelector: SystemAppDeploymentName},
		},
	}
}

func (system *System) MasterService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-master",
			Labels: system.Options.MasterUILabels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "http",
					Protocol:   v1.ProtocolTCP,
					Port:       3000,
					TargetPort: intstr.FromString("master"),
				},
			},
			Selector: map[string]string{reconcilers.DeploymentLabelSelector: SystemAppDeploymentName},
		},
	}
}

func (system *System) DeveloperService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-developer",
			Labels: system.Options.DeveloperUILabels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "http",
					Protocol:   v1.ProtocolTCP,
					Port:       3000,
					TargetPort: intstr.FromString("developer"),
				},
			},
			Selector: map[string]string{reconcilers.DeploymentLabelSelector: SystemAppDeploymentName},
		},
	}
}

func (system *System) MemcachedService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-memcache",
			Labels: system.Options.MemcachedLabels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "memcache",
					Protocol:   v1.ProtocolTCP,
					Port:       11211,
					TargetPort: intstr.FromInt32(11211),
				},
			},
			Selector: map[string]string{reconcilers.DeploymentLabelSelector: "system-memcache"},
		},
	}
}

func (system *System) SMTPSecret() *v1.Secret {
	res := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSecretSystemSMTPSecretName,
			Labels: system.Options.SMTPLabels,
		},
		StringData: map[string]string{
			SystemSecretSystemSMTPAddressFieldName:           *system.Options.SmtpSecretOptions.Address,
			SystemSecretSystemSMTPAuthenticationFieldName:    *system.Options.SmtpSecretOptions.Authentication,
			SystemSecretSystemSMTPDomainFieldName:            *system.Options.SmtpSecretOptions.Domain,
			SystemSecretSystemSMTPOpenSSLVerifyModeFieldName: *system.Options.SmtpSecretOptions.OpenSSLVerifyMode,
			SystemSecretSystemSMTPPasswordFieldName:          *system.Options.SmtpSecretOptions.Password,
			SystemSecretSystemSMTPPortFieldName:              *system.Options.SmtpSecretOptions.Port,
			SystemSecretSystemSMTPUserNameFieldName:          *system.Options.SmtpSecretOptions.Username,
		},
	}

	if system.Options.SmtpSecretOptions.FromAddress != nil {
		res.StringData[SystemSecretSystemSMTPFromAddressFieldName] = *system.Options.SmtpSecretOptions.FromAddress
	}

	return res
}

func (system *System) SystemConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system",
			Labels: system.Options.CommonLabels,
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
	// When zync is disabled, the endpoint need to be an empty string to prevent spamming sidekiq with ZyncWorker jobs
	if !system.Options.ZyncEnabled {
		return `production:
  endpoint: ''
  authentication:
    token: "<%= ENV.fetch('ZYNC_AUTHENTICATION_TOKEN') %>"
  connect_timeout: 5
  send_timeout: 5
  receive_timeout: 10
  root_url:
`
	}

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
	return `production: {}
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

func (system *System) AppPodDisruptionBudget() *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-app",
			Labels: system.Options.CommonAppLabels,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{reconcilers.DeploymentLabelSelector: SystemAppDeploymentName},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}

func (system *System) SidekiqPodDisruptionBudget() *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-sidekiq",
			Labels: system.Options.CommonSidekiqLabels,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{reconcilers.DeploymentLabelSelector: "system-sidekiq"},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}

func (system *System) sideKiqPorts() []v1.ContainerPort {
	var ports []v1.ContainerPort

	if system.Options.SideKiqMetrics {
		ports = append(ports, v1.ContainerPort{Name: "metrics", ContainerPort: SystemSidekiqMetricsPort, Protocol: v1.ProtocolTCP})
	}

	return ports
}

func (system *System) appMasterPorts() []v1.ContainerPort {
	var ports []v1.ContainerPort

	ports = append(ports, v1.ContainerPort{Name: "master", HostPort: 0, ContainerPort: 3002, Protocol: v1.ProtocolTCP})

	if system.Options.AppMetrics {
		ports = append(ports, v1.ContainerPort{Name: SystemAppMasterContainerMetricsPortName, ContainerPort: SystemAppMasterContainerPrometheusPort, Protocol: v1.ProtocolTCP})
	}

	return ports
}

func (system *System) appProviderPorts() []v1.ContainerPort {
	var ports []v1.ContainerPort

	ports = append(ports, v1.ContainerPort{Name: "provider", HostPort: 0, ContainerPort: 3000, Protocol: v1.ProtocolTCP})

	if system.Options.AppMetrics {
		ports = append(ports, v1.ContainerPort{Name: SystemAppProviderContainerMetricsPortName, ContainerPort: SystemAppProviderContainerPrometheusPort, Protocol: v1.ProtocolTCP})
	}

	return ports
}

func (system *System) appDeveloperPorts() []v1.ContainerPort {
	var ports []v1.ContainerPort

	ports = append(ports, v1.ContainerPort{Name: "developer", HostPort: 0, ContainerPort: 3001, Protocol: v1.ProtocolTCP})

	if system.Options.AppMetrics {
		ports = append(ports, v1.ContainerPort{Name: SystemAppDeveloperContainerMetricsPortName, ContainerPort: SystemAppDeveloperContainerPrometheusPort, Protocol: v1.ProtocolTCP})
	}

	return ports
}

func (system *System) systemInit(containerImage string) []v1.Container {
	if system.Options.SystemDbTLSEnabled {
		return []v1.Container{
			{
				Name:  "set-permissions",
				Image: containerImage, // Minimal image for chmod
				Command: []string{
					"sh",
					"-c",
					"cp /tls/* /writable-tls/ && chmod 0600 /writable-tls/*",
				},
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      "tls-secret",
						MountPath: "/tls",
						ReadOnly:  true,
					},
					{
						Name:      "writable-tls",
						MountPath: "/writable-tls",
						ReadOnly:  false, // Writable emptyDir volume
					},
				},
			},
		}
	} else {
		return []v1.Container{}
	}
}

func (system *System) sidekiqInit(containerImage string) []v1.Container {
	var containers []v1.Container
	// Base init container setup
	initContainer := v1.Container{
		Name:  SystemSideKiqInitContainerName,
		Image: containerImage,
		Command: []string{
			"bash",
			"-c",
			"bundle exec sh -c \"until rake boot:redis && curl --output /dev/null --silent --fail --head http://system-master:3000/status; do sleep $SLEEP_SECONDS; done\"",
		},
		Env: append(system.SystemRedisEnvVars(), helper.EnvVarFromValue("SLEEP_SECONDS", "1")),
	}

	// Append Redis TLS volume mounts if Redis TLS is enabled
	if system.Options.RedisTLSEnabled {
		initContainer.VolumeMounts = append(initContainer.VolumeMounts, system.systemRedisTlsVolumeMount())
		initContainer.VolumeMounts = append(initContainer.VolumeMounts, system.backendRedisTlsVolumeMount())
	}

	// Append SystemDb TLS volume mount if SystemDb TLS is enabled
	if system.Options.SystemDbTLSEnabled {
		// Set-permissions container for DB TLS
		containers = append(containers, v1.Container{
			Name:  "set-permissions",
			Image: containerImage,
			Command: []string{
				"sh",
				"-c",
				"cp /tls/* /writable-tls/ && chmod 0600 /writable-tls/*",
			},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "tls-secret",
					MountPath: "/tls",
					ReadOnly:  true,
				},
				{
					Name:      "writable-tls",
					MountPath: "/writable-tls",
					ReadOnly:  false, // Writable emptyDir volume
				},
			},
		})
	}

	containers = append(containers, initContainer)
	return containers
}

func (system *System) appPodAnnotations(watchedSecretAnnotations map[string]string) map[string]string {
	annotations := system.Options.AppPodTemplateAnnotations
	if annotations == nil {
		annotations = make(map[string]string)
	}
	for key, val := range watchedSecretAnnotations {
		annotations[key] = val
	}

	for key, val := range system.Options.AppPodTemplateAnnotations {
		annotations[key] = val
	}

	return annotations
}

func (system *System) sidekiqPodAnnotations(watchedSecretAnnotations map[string]string) map[string]string {
	annotations := system.Options.SideKiqPodTemplateAnnotations

	if annotations == nil {
		annotations = make(map[string]string)
	}
	for key, val := range watchedSecretAnnotations {
		annotations[key] = val
	}

	for key, val := range system.Options.SideKiqPodTemplateAnnotations {
		annotations[key] = val
	}

	return annotations
}

func (system *System) SystemRedisTLSEnvVars() []v1.EnvVar {
	return []v1.EnvVar{
		helper.EnvVarFromValue("REDIS_CA_FILE", redisCaFilePath),
		helper.EnvVarFromValue("REDIS_CLIENT_CERT", redisClientCertPath),
		helper.EnvVarFromValue("REDIS_PRIVATE_KEY", redisPrivateKeyPath),
		helper.EnvVarFromValue("REDIS_SSL", "1"),
	}
}

func (system *System) BackendRedisTLSEnvVars() []v1.EnvVar {
	return []v1.EnvVar{
		helper.EnvVarFromValue("BACKEND_REDIS_CA_FILE", backendRedisCaFilePath),
		helper.EnvVarFromValue("BACKEND_REDIS_CLIENT_CERT", backendRedisClientCertPath),
		helper.EnvVarFromValue("BACKEND_REDIS_PRIVATE_KEY", backendRedisPrivateKeyPath),
		helper.EnvVarFromValue("BACKEND_REDIS_SSL", "1"),
	}
}

func (system *System) systemRedisTlsVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      "system-redis-tls",
		ReadOnly:  false,
		MountPath: "/tls/system-redis",
	}
}

func (system *System) backendRedisTlsVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      "backend-redis-tls",
		ReadOnly:  false,
		MountPath: "/tls/backend-redis",
	}
}
