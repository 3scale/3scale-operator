package component

import (
	"sort"

	"github.com/3scale/3scale-operator/pkg/common"

	appsv1 "github.com/openshift/api/apps/v1"
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
	SystemSecretSystemRedisSentinelHosts               = "SENTINEL_HOSTS"
	SystemSecretSystemRedisSentinelRole                = "SENTINEL_ROLE"
	SystemSecretSystemRedisMessageBusSentinelHosts     = "MESSAGE_BUS_SENTINEL_HOSTS"
	SystemSecretSystemRedisMessageBusSentinelRole      = "MESSAGE_BUS_SENTINEL_ROLE"
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

const (
	SystemSecretSystemSMTPSecretName                 = "system-smtp"
	SystemSecretSystemSMTPAddressFieldName           = "address"
	SystemSecretSystemSMTPUserNameFieldName          = "username"
	SystemSecretSystemSMTPPasswordFieldName          = "password"
	SystemSecretSystemSMTPDomainFieldName            = "domain"
	SystemSecretSystemSMTPPortFieldName              = "port"
	SystemSecretSystemSMTPAuthenticationFieldName    = "authentication"
	SystemSecretSystemSMTPOpenSSLVerifyModeFieldName = "openssl.verify.mode"
)

const (
	SystemFileStoragePVCName = "system-storage"
)

type System struct {
	Options *SystemOptions
}

func NewSystem(options *SystemOptions) *System {
	return &System{Options: options}
}

func (system *System) Objects() []common.KubernetesObject {
	sharedStorage := system.SharedStorage()
	providerService := system.ProviderService()
	masterService := system.MasterService()
	developerService := system.DeveloperService()
	sphinxService := system.SphinxService()
	memcachedService := system.MemcachedService()

	appDeploymentConfig := system.AppDeploymentConfig()
	sidekiqDeploymentConfig := system.SidekiqDeploymentConfig()
	sphinxDeploymentConfig := system.SphinxDeploymentConfig()

	systemConfigMap := system.SystemConfigMap()
	environmentConfigMap := system.EnvironmentConfigMap()
	smtpSecret := system.SMTPSecret()

	eventsHookSecret := system.EventsHookSecret()

	redisSecret := system.RedisSecret()
	masterApicastSecret := system.MasterApicastSecret()

	seedSecret := system.SeedSecret()
	recaptchaSecret := system.RecaptchaSecret()
	appSecret := system.AppSecret()
	memcachedSecret := system.MemcachedSecret()

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
	return objects
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
		envvar := envVarFromConfigMap(key, "system-environment", key)
		result = append(result, envvar)
	}

	return result
}

func (system *System) getSystemSMTPEnvsFromSMTPSecret() []v1.EnvVar {
	result := []v1.EnvVar{
		envVarFromSecret("SMTP_ADDRESS", SystemSecretSystemSMTPSecretName, "address"),
		envVarFromSecret("SMTP_USER_NAME", SystemSecretSystemSMTPSecretName, "username"),
		envVarFromSecret("SMTP_PASSWORD", SystemSecretSystemSMTPSecretName, "password"),
		envVarFromSecret("SMTP_DOMAIN", SystemSecretSystemSMTPSecretName, "domain"),
		envVarFromSecret("SMTP_PORT", SystemSecretSystemSMTPSecretName, "port"),
		envVarFromSecret("SMTP_AUTHENTICATION", SystemSecretSystemSMTPSecretName, "authentication"),
		envVarFromSecret("SMTP_OPENSSL_VERIFY_MODE", SystemSecretSystemSMTPSecretName, "openssl.verify.mode"),
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
		envVarFromSecret("REDIS_SENTINEL_HOSTS", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisSentinelHosts),
		envVarFromSecret("REDIS_SENTINEL_ROLE", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisSentinelRole),
		envVarFromSecret("MESSAGE_BUS_REDIS_SENTINEL_HOSTS", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisMessageBusSentinelHosts),
		envVarFromSecret("MESSAGE_BUS_REDIS_SENTINEL_ROLE", SystemSecretSystemRedisSecretName, SystemSecretSystemRedisMessageBusSentinelRole),
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
	)

	result = append(result, system.SystemRedisEnvVars()...)
	result = append(result, system.BackendRedisEnvVars()...)
	bckListenerApicastRouteEnv := envVarFromSecret("APICAST_BACKEND_ROOT_ENDPOINT", "backend-listener", "route_endpoint")
	bckListenerRouteEnv := envVarFromSecret("BACKEND_ROUTE", "backend-listener", "route_endpoint")
	result = append(result, bckListenerApicastRouteEnv, bckListenerRouteEnv)

	smtpEnvSecretEnvs := system.getSystemSMTPEnvsFromSMTPSecret()
	result = append(result, smtpEnvSecretEnvs...)

	apicastAccessToken := envVarFromSecret("APICAST_ACCESS_TOKEN", SystemSecretSystemMasterApicastSecretName, "ACCESS_TOKEN")
	result = append(result, apicastAccessToken)

	// Add zync secret to envvars sources
	zyncAuthTokenVar := envVarFromSecret("ZYNC_AUTHENTICATION_TOKEN", "zync", "ZYNC_AUTHENTICATION_TOKEN")
	result = append(result, zyncAuthTokenVar)

	// Add backend internal api data to envvars sources
	systemBackendInternalAPIUser := envVarFromSecret("CONFIG_INTERNAL_API_USER", "backend-internal-api", "username")
	systemBackendInternalAPIPass := envVarFromSecret("CONFIG_INTERNAL_API_PASSWORD", "backend-internal-api", "password")
	result = append(result, systemBackendInternalAPIUser, systemBackendInternalAPIPass)

	if system.Options.s3FileStorageOptions != nil {
		result = append(result,
			envVarFromConfigMap("FILE_UPLOAD_STORAGE", "system-environment", "FILE_UPLOAD_STORAGE"),
			envVarFromSecret(AwsAccessKeyID, system.Options.s3FileStorageOptions.ConfigurationSecretName, AwsAccessKeyID),
			envVarFromSecret(AwsSecretAccessKey, system.Options.s3FileStorageOptions.ConfigurationSecretName, AwsSecretAccessKey),
			envVarFromSecret(AwsBucket, system.Options.s3FileStorageOptions.ConfigurationSecretName, AwsBucket),
			envVarFromSecret(AwsRegion, system.Options.s3FileStorageOptions.ConfigurationSecretName, AwsRegion),
			envVarFromSecretOptional(AwsProtocol, system.Options.s3FileStorageOptions.ConfigurationSecretName, AwsProtocol),
			envVarFromSecretOptional(AwsHostname, system.Options.s3FileStorageOptions.ConfigurationSecretName, AwsHostname),
			envVarFromSecretOptional(AwsPathStyle, system.Options.s3FileStorageOptions.ConfigurationSecretName, AwsPathStyle),
		)
	}

	return result
}

func (system *System) buildSystemAppPreHookEnv() []v1.EnvVar {
	result := []v1.EnvVar{}
	baseEnv := system.buildSystemBaseEnv()
	result = append(result, baseEnv...)
	result = append(result,
		envVarFromSecret("MASTER_ACCESS_TOKEN", SystemSecretSystemSeedSecretName, SystemSecretSystemSeedMasterAccessTokenFieldName),
	)
	return result
}

func (system *System) BackendRedisEnvVars() []v1.EnvVar {
	return []v1.EnvVar{
		envVarFromSecret("BACKEND_REDIS_URL", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageURLFieldName),
		envVarFromSecret("BACKEND_REDIS_SENTINEL_HOSTS", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageSentinelHostsFieldName),
		envVarFromSecret("BACKEND_REDIS_SENTINEL_ROLE", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageSentinelRoleFieldName),
	}
}

func (system *System) EnvironmentConfigMap() *v1.ConfigMap {
	res := &v1.ConfigMap{
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

	if system.Options.s3FileStorageOptions != nil {
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

func (system *System) RecaptchaSecret() *v1.Secret {
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

func (system *System) EventsHookSecret() *v1.Secret {
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

func (system *System) RedisSecret() *v1.Secret {
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
			SystemSecretSystemRedisSentinelHosts:               *system.Options.redisSentinelHosts,
			SystemSecretSystemRedisSentinelRole:                *system.Options.redisSentinelRole,
			SystemSecretSystemRedisMessageBusRedisURLFieldName: *system.Options.messageBusRedisURL,
			SystemSecretSystemRedisMessageBusSentinelHosts:     *system.Options.messageBusRedisSentinelHosts,
			SystemSecretSystemRedisMessageBusSentinelRole:      *system.Options.messageBusRedisSentinelRole,
			SystemSecretSystemRedisNamespace:                   *system.Options.redisNamespace,
			SystemSecretSystemRedisMessageBusRedisNamespace:    *system.Options.messageBusRedisNamespace,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (system *System) AppSecret() *v1.Secret {
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

func (system *System) SeedSecret() *v1.Secret {
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

func (system *System) MasterApicastSecret() *v1.Secret {
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

func (system *System) appPodVolumes() []v1.Volume {
	res := []v1.Volume{}
	if system.Options.pvcFileStorageOptions != nil {
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
				},
			},
		},
	}

	res = append(res, systemConfigVolume)
	return res
}

func (system *System) volumeNamesForSystemAppPreHookPod() []string {
	res := []string{}
	if system.Options.pvcFileStorageOptions != nil {
		res = append(res, SystemFileStoragePVCName)
	}
	return res
}

func (system *System) AppDeploymentConfig() *appsv1.DeploymentConfig {
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
							Command:       []string{"bash", "-c", "bundle exec rake boot openshift:deploy"},
							Env:           system.buildSystemAppPreHookEnv(),
							ContainerName: "system-master",
							Volumes:       system.volumeNamesForSystemAppPreHookPod()},
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
			Replicas: *system.Options.appReplicas,
			Selector: map[string]string{"deploymentConfig": "system-app"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "app", "app": system.Options.appLabel, "deploymentConfig": "system-app"},
				},
				Spec: v1.PodSpec{
					Volumes: system.appPodVolumes(),
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
							Env:          system.buildSystemBaseEnv(),
							Resources:    *system.Options.appMasterContainerResourceRequirements,
							VolumeMounts: system.appMasterContainerVolumeMounts(),
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
							Env:          system.buildSystemBaseEnv(),
							Resources:    *system.Options.appProviderContainerResourceRequirements,
							VolumeMounts: system.appProviderContainerVolumeMounts(),
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
							Env:          system.buildSystemBaseEnv(),
							Resources:    *system.Options.appDeveloperContainerResourceRequirements,
							VolumeMounts: system.appDeveloperContainerVolumeMounts(),
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
			Medium: v1.StorageMediumMemory}},
	}

	res = append(res, systemTmpVolume)
	if system.Options.pvcFileStorageOptions != nil {
		res = append(res, system.FileStorageVolume())
	}

	systemConfigVolume := v1.Volume{
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
			},
		},
		},
	}

	res = append(res, systemConfigVolume)
	return res
}

func (system *System) SidekiqDeploymentConfig() *appsv1.DeploymentConfig {
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
			Replicas: *system.Options.sidekiqReplicas,
			Selector: map[string]string{"deploymentConfig": "system-sidekiq"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "sidekiq", "app": system.Options.appLabel, "deploymentConfig": "system-sidekiq"},
				},
				Spec: v1.PodSpec{
					Volumes: system.SidekiqPodVolumes(),
					InitContainers: []v1.Container{
						v1.Container{
							Name:  "check-svc",
							Image: "amp-system:latest",
							Command: []string{
								"bash",
								"-c",
								"bundle exec sh -c \"until rake boot:redis && curl --output /dev/null --silent --fail --head http://system-master:3000/status; do sleep $SLEEP_SECONDS; done\"",
							},
							Env: append(system.SystemRedisEnvVars(), envVarFromValue("SLEEP_SECONDS", "1")),
						},
					},
					Containers: []v1.Container{
						v1.Container{
							Name:            "system-sidekiq",
							Image:           "amp-system:latest",
							Args:            []string{"rake", "sidekiq:worker", "RAILS_MAX_THREADS=25"},
							Env:             system.buildSystemBaseEnv(),
							Resources:       *system.Options.sidekiqContainerResourceRequirements,
							VolumeMounts:    system.sidekiqContainerVolumeMounts(),
							ImagePullPolicy: v1.PullIfNotPresent,
						},
					},
					ServiceAccountName: "amp",
				}},
		},
	}
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

func (system *System) appCommonContainerVolumeMounts(systemStorageReadonly bool) []v1.VolumeMount {
	res := []v1.VolumeMount{}
	if system.Options.pvcFileStorageOptions != nil {
		res = append(res, system.systemStorageVolumeMount(systemStorageReadonly))
	}
	res = append(res, system.systemConfigVolumeMount())

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
	if system.Options.pvcFileStorageOptions != nil {
		res = append(res, system.systemStorageVolumeMount(false))
	}
	systemTmpVolumeMount := v1.VolumeMount{
		Name:      "system-tmp",
		ReadOnly:  false,
		MountPath: "/tmp",
	}
	res = append(res, systemTmpVolumeMount)
	res = append(res, system.systemConfigVolumeMount())
	return res
}

func (system *System) SharedStorage() *v1.PersistentVolumeClaim {
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
			StorageClassName: system.Options.pvcFileStorageOptions.StorageClass,
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

func (system *System) ProviderService() *v1.Service {
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

func (system *System) MasterService() *v1.Service {
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

func (system *System) DeveloperService() *v1.Service {
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

func (system *System) SphinxService() *v1.Service {
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

func (system *System) MemcachedService() *v1.Service {
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

func (system *System) SMTPSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSecretSystemSMTPSecretName,
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "smtp", "app": system.Options.appLabel},
		},
		StringData: map[string]string{
			SystemSecretSystemSMTPAddressFieldName:           system.Options.smtpSecretOptions.Address,
			SystemSecretSystemSMTPAuthenticationFieldName:    system.Options.smtpSecretOptions.Authentication,
			SystemSecretSystemSMTPDomainFieldName:            system.Options.smtpSecretOptions.Domain,
			SystemSecretSystemSMTPOpenSSLVerifyModeFieldName: system.Options.smtpSecretOptions.OpenSSLVerifyMode,
			SystemSecretSystemSMTPPasswordFieldName:          system.Options.smtpSecretOptions.Password,
			SystemSecretSystemSMTPPortFieldName:              system.Options.smtpSecretOptions.Port,
			SystemSecretSystemSMTPUserNameFieldName:          system.Options.smtpSecretOptions.Username,
		},
	}
}

func (system *System) SystemConfigMap() *v1.ConfigMap {
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

func (system *System) SphinxDeploymentConfig() *appsv1.DeploymentConfig {
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
							Resources: *system.Options.sphinxContainerResourceRequirements,
						},
					},
				},
			},
		},
	}
}
