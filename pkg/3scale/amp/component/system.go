package component

import (
	"sort"

	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	templatev1 "github.com/openshift/api/template/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type System struct {
	options []string
}

func NewSystem(options []string) *System {
	system := &System{
		options: options,
	}
	return system
}

func (system *System) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	system.buildParameters(template)
	system.buildObjects(template)
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
	}
	template.Parameters = append(template.Parameters, parameters...)
}

func (system *System) buildObjects(template *templatev1.Template) {
	systemSharedStorage := system.buildSystemSharedPVC()
	systemProviderService := system.buildSystemProviderService()
	systemMasterService := system.buildSystemMasterService()
	systemDeveloperService := system.buildSystemDeveloperService()
	systemProviderRoute := system.buildSystemProviderRoute()
	systemMasterRoute := system.buildSystemMasterRoute()
	systemDeveloperRoute := system.buildSystemDeveloperRoute()
	systemMysqlService := system.buildSystemMysqlService()
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

	systemDatabaseSecret := system.buildSystemDatabaseSecrets()
	systemSeedSecret := system.buildSystemSeedSecrets()
	systemRecaptchaSecret := system.buildSystemRecaptchaSecrets()
	systemAppSecret := system.buildSystemAppSecrets()
	systemMemcachedSecret := system.buildSystemMemcachedSecrets()

	objects := []runtime.RawExtension{
		runtime.RawExtension{Object: systemSharedStorage},
		runtime.RawExtension{Object: systemProviderService},
		runtime.RawExtension{Object: systemMasterService},
		runtime.RawExtension{Object: systemDeveloperService},
		runtime.RawExtension{Object: systemProviderRoute},
		runtime.RawExtension{Object: systemMasterRoute},
		runtime.RawExtension{Object: systemDeveloperRoute},
		runtime.RawExtension{Object: systemMysqlService},
		runtime.RawExtension{Object: systemRedisService},
		runtime.RawExtension{Object: systemSphinxService},
		runtime.RawExtension{Object: systemMemcachedService},
		runtime.RawExtension{Object: systemConfigMap},
		runtime.RawExtension{Object: systemSmtpConfigMap},
		runtime.RawExtension{Object: systemEnvironmentConfigMap},
		runtime.RawExtension{Object: systemAppDeploymentConfig},
		runtime.RawExtension{Object: systemSidekiqDeploymentConfig},
		runtime.RawExtension{Object: systemSphinxDeploymentConfig},
		runtime.RawExtension{Object: systemEventsHookSecret},
		runtime.RawExtension{Object: systemRedisSecret},
		runtime.RawExtension{Object: systemMasterApicastSecret},
		runtime.RawExtension{Object: systemDatabaseSecret},
		runtime.RawExtension{Object: systemSeedSecret},
		runtime.RawExtension{Object: systemRecaptchaSecret},
		runtime.RawExtension{Object: systemAppSecret},
		runtime.RawExtension{Object: systemMemcachedSecret},
	}
	template.Objects = append(template.Objects, objects...)
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
		envvar := createEnvVarFromConfigMap(key, "system-environment", key)
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
	// 	envvar := createEnvVarFromConfigMap(key, "smtp", key)
	// 	result = append(result, envvar)
	// }

	result := []v1.EnvVar{
		createEnvVarFromConfigMap("SMTP_ADDRESS", "smtp", "address"),
		createEnvVarFromConfigMap("SMTP_USER_NAME", "smtp", "username"),
		createEnvVarFromConfigMap("SMTP_PASSWORD", "smtp", "password"),
		createEnvVarFromConfigMap("SMTP_DOMAIN", "smtp", "domain"),
		createEnvVarFromConfigMap("SMTP_PORT", "smtp", "port"),
		createEnvVarFromConfigMap("SMTP_AUTHENTICATION", "smtp", "authentication"),
		createEnvVarFromConfigMap("SMTP_OPENSSL_VERIFY_MODE", "smtp", "openssl.verify.mode"),
	}

	return result
}

func (system *System) buildSystemSphinxEnv() []v1.EnvVar {
	result := []v1.EnvVar{}

	result = append(result,
		createEnvVarFromConfigMap("RAILS_ENV", "system-environment", "RAILS_ENV"),
		createEnvvarFromSecret("DATABASE_URL", "system-database", "URL"),
		createEnvVarFromValue("THINKING_SPHINX_ADDRESS", "0.0.0.0"),
		createEnvVarFromValue("THINKING_SPHINX_CONFIGURATION_FILE", "db/sphinx/production.conf"),
		createEnvVarFromValue("THINKING_SPHINX_PID_FILE", "db/sphinx/searchd.pid"),
		createEnvVarFromValue("DELTA_INDEX_INTERVAL", "5"),
		createEnvVarFromValue("FULL_REINDEX_INTERVAL", "60"),
	)
	return result
}

func (system *System) buildSystemBaseEnv() []v1.EnvVar {
	result := []v1.EnvVar{}

	baseEnvConfigMapEnvs := system.getSystemBaseEnvsFromEnvConfigMap()
	result = append(result, baseEnvConfigMapEnvs...)

	result = append(result,
		createEnvvarFromSecret("DATABASE_URL", "system-database", "URL"),

		createEnvvarFromSecret("MASTER_DOMAIN", "system-seed", "MASTER_DOMAIN"),
		createEnvvarFromSecret("MASTER_USER", "system-seed", "MASTER_USER"),
		createEnvvarFromSecret("MASTER_PASSWORD", "system-seed", "MASTER_PASSWORD"),

		createEnvvarFromSecret("ADMIN_ACCESS_TOKEN", "system-seed", "ADMIN_ACCESS_TOKEN"),
		createEnvvarFromSecret("USER_LOGIN", "system-seed", "ADMIN_USER"),
		createEnvvarFromSecret("USER_PASSWORD", "system-seed", "ADMIN_PASSWORD"),
		createEnvvarFromSecret("TENANT_NAME", "system-seed", "TENANT_NAME"),

		createEnvVarFromValue("THINKING_SPHINX_ADDRESS", "system-sphinx"),
		createEnvVarFromValue("THINKING_SPHINX_CONFIGURATION_FILE", "/tmp/sphinx.conf"),

		createEnvvarFromSecret("EVENTS_SHARED_SECRET", "system-events-hook", "PASSWORD"),

		createEnvvarFromSecret("RECAPTCHA_PUBLIC_KEY", "system-recaptcha", "PUBLIC_KEY"),
		createEnvvarFromSecret("RECAPTCHA_PRIVATE_KEY", "system-recaptcha", "PRIVATE_KEY"),

		createEnvvarFromSecret("SECRET_KEY_BASE", "system-app", "SECRET_KEY_BASE"),

		createEnvvarFromSecret("REDIS_URL", "system-redis", "URL"),

		createEnvvarFromSecret("MEMCACHE_SERVERS", "system-memcache", "SERVERS"),

		createEnvvarFromSecret("BACKEND_REDIS_URL", "backend-redis", "REDIS_STORAGE_URL"),
	)

	bckListenerApicastRouteEnv := createEnvvarFromSecret("APICAST_BACKEND_ROOT_ENDPOINT", "backend-listener", "route_endpoint")
	bckListenerRouteEnv := createEnvvarFromSecret("BACKEND_ROUTE", "backend-listener", "route_endpoint")
	result = append(result, bckListenerApicastRouteEnv, bckListenerRouteEnv)

	smtpEnvConfigMapEnvs := system.getSystemSmtpEnvsFromSMTPConfigMap()
	result = append(result, smtpEnvConfigMapEnvs...)

	apicastAccessToken := createEnvvarFromSecret("APICAST_ACCESS_TOKEN", "system-master-apicast", "ACCESS_TOKEN")
	result = append(result, apicastAccessToken)

	// Add zync secret to envvars sources
	zyncAuthTokenVar := createEnvvarFromSecret("ZYNC_AUTHENTICATION_TOKEN", "zync", "ZYNC_AUTHENTICATION_TOKEN")
	result = append(result, zyncAuthTokenVar)

	// Add backend internal api data to envvars sources
	systemBackendInternalAPIUser := createEnvvarFromSecret("CONFIG_INTERNAL_API_USER", "backend-internal-api", "username")
	systemBackendInternalAPIPass := createEnvvarFromSecret("CONFIG_INTERNAL_API_PASSWORD", "backend-internal-api", "password")
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
			Labels: map[string]string{"3scale.component": "system", "app": "${APP_LABEL}"},
		},
		Data: map[string]string{
			"RAILS_ENV":              "production",
			"FORCE_SSL":              "true",
			"THREESCALE_SUPERDOMAIN": "${WILDCARD_DOMAIN}",
			"PROVIDER_PLAN":          "enterprise",
			"APICAST_REGISTRY_URL":   "${APICAST_REGISTRY_URL}",
			"RAILS_LOG_TO_STDOUT":    "true",
			"RAILS_LOG_LEVEL":        "info",
			"THINKING_SPHINX_PORT":   "9306",
			"THREESCALE_SANDBOX_PROXY_OPENSSL_VERIFY_MODE": "VERIFY_NONE",
			"AMP_RELEASE":  "${AMP_RELEASE}",
			"SSL_CERT_DIR": "/etc/pki/tls/certs",
		},
	}
}

func (system *System) buildSystemDatabaseSecrets() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-database",
			Labels: map[string]string{
				"app":              "${APP_LABEL}",
				"3scale.component": "system",
			},
		},
		StringData: map[string]string{
			"URL": "mysql2://root:${MYSQL_ROOT_PASSWORD}@system-mysql/${MYSQL_DATABASE}",
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (system *System) buildSystemMemcachedSecrets() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-memcache",
			Labels: map[string]string{
				"app":              "${APP_LABEL}",
				"3scale.component": "system",
			},
		},
		StringData: map[string]string{
			"SERVERS": "system-memcache:11211",
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
			Name: "system-recaptcha",
			Labels: map[string]string{
				"app":              "${APP_LABEL}",
				"3scale.component": "system",
			},
		},
		StringData: map[string]string{
			"PUBLIC_KEY":  "${RECAPTCHA_PUBLIC_KEY}",
			"PRIVATE_KEY": "${RECAPTCHA_PRIVATE_KEY}",
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
			Name: "system-events-hook",
			Labels: map[string]string{
				"app":              "${APP_LABEL}",
				"3scale.component": "system",
			},
		},
		StringData: map[string]string{
			"URL":      "http://system-master:3000/master/events/import",
			"PASSWORD": "${SYSTEM_BACKEND_SHARED_SECRET}",
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
			Name: "system-redis",
			Labels: map[string]string{
				"app":              "${APP_LABEL}",
				"3scale.component": "system",
			},
		},
		StringData: map[string]string{
			"URL": "redis://system-redis:6379/1",
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
			Name: "system-app", // TODO sure this should be a secret on its own?? maybe can join different secrets into one with more values?
			Labels: map[string]string{
				"app":              "${APP_LABEL}",
				"3scale.component": "system",
			},
		},
		StringData: map[string]string{
			"SECRET_KEY_BASE": "${SYSTEM_APP_SECRET_KEY_BASE}",
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
			Name: "system-seed",
			Labels: map[string]string{
				"app":              "${APP_LABEL}",
				"3scale.component": "system",
			},
		},
		StringData: map[string]string{
			"MASTER_DOMAIN":      "${MASTER_NAME}",
			"MASTER_USER":        "${MASTER_USER}",
			"MASTER_PASSWORD":    "${MASTER_PASSWORD}",
			"ADMIN_ACCESS_TOKEN": "${ADMIN_ACCESS_TOKEN}",
			"ADMIN_USER":         "${ADMIN_USERNAME}",
			"ADMIN_PASSWORD":     "${ADMIN_PASSWORD}",
			"TENANT_NAME":        "${TENANT_NAME}",
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
			Name: "system-master-apicast",
			Labels: map[string]string{
				"app":              "${APP_LABEL}",
				"3scale.component": "system",
			},
		},
		StringData: map[string]string{
			"PROXY_CONFIGS_ENDPOINT": "http://${APICAST_ACCESS_TOKEN}@system-master:3000/master/api/proxy/configs",
			"BASE_URL":               "http://${APICAST_ACCESS_TOKEN}@system-master:3000",
			"ACCESS_TOKEN":           "${APICAST_ACCESS_TOKEN}",
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
			Labels: map[string]string{"3scale.component": "system", "3scale.component-element": "app", "app": "${APP_LABEL}"},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyType("Rolling"),
				RollingParams: &appsv1.RollingDeploymentStrategyParams{
					UpdatePeriodSeconds: &[]int64{1}[0],
					IntervalSeconds:     &[]int64{1}[0],
					TimeoutSeconds:      &[]int64{1200}[0],
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Type(1),
						IntVal: 0,
						StrVal: "25%",
					},
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.Type(1),
						IntVal: 0,
						StrVal: "25%",
					},
					Pre: &appsv1.LifecycleHook{
						FailurePolicy: appsv1.LifecycleHookFailurePolicy("Retry"),
						ExecNewPod: &appsv1.ExecNewPodHook{
							Command:       []string{"bash", "-c", "bundle exec rake boot openshift:deploy MASTER_ACCESS_TOKEN=\"${MASTER_ACCESS_TOKEN}\""},
							Env:           system.buildSystemBaseEnv(),
							ContainerName: "system-master",
							Volumes:       []string{"system-storage"}},
					},
					Post: &appsv1.LifecycleHook{
						FailurePolicy: appsv1.LifecycleHookFailurePolicy("Abort"),
						ExecNewPod: &appsv1.ExecNewPodHook{
							Command:       []string{"bash", "-c", "bundle exec rake boot openshift:post_deploy"},
							ContainerName: "system-master"}}},
			},
			MinReadySeconds: 0,
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerType("ConfigChange"),
				}, appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerType("ImageChange"),
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
					Labels: map[string]string{"3scale.component": "system", "3scale.component-element": "app", "app": "${APP_LABEL}", "deploymentConfig": "system-app"},
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
									Protocol:      v1.Protocol("TCP")},
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
										Type:   intstr.Type(1),
										IntVal: 0,
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
										Type:   intstr.Type(1),
										IntVal: 0,
										StrVal: "master",
									},
									Scheme: v1.URIScheme("HTTP"),
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
							ImagePullPolicy: v1.PullPolicy("IfNotPresent"),
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
									Protocol:      v1.Protocol("TCP")},
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
										Type:   intstr.Type(1),
										IntVal: 0,
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
										Type:   intstr.Type(1),
										IntVal: 0,
										StrVal: "provider",
									},
									Scheme: v1.URIScheme("HTTP"),
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
							ImagePullPolicy: v1.PullPolicy("IfNotPresent"),
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
									Protocol:      v1.Protocol("TCP")},
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
										Type:   intstr.Type(1),
										IntVal: 0,
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
										Type:   intstr.Type(1),
										IntVal: 0,
										StrVal: "developer",
									},
									Scheme: v1.URIScheme("HTTP"),
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
							ImagePullPolicy: v1.PullPolicy("IfNotPresent"),
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
			Labels: map[string]string{"3scale.component": "system", "3scale.component-element": "sidekiq", "app": "${APP_LABEL}"},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyType("Rolling"),
				RollingParams: &appsv1.RollingDeploymentStrategyParams{
					UpdatePeriodSeconds: &[]int64{1}[0],
					IntervalSeconds:     &[]int64{1}[0],
					TimeoutSeconds:      &[]int64{1200}[0],
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Type(1),
						IntVal: 0,
						StrVal: "25%",
					},
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.Type(1),
						IntVal: 0,
						StrVal: "25%"}},
			},
			MinReadySeconds: 0,
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerType("ConfigChange"),
				}, appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerType("ImageChange"),
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
					Labels: map[string]string{"3scale.component": "system", "3scale.component-element": "sidekiq", "app": "${APP_LABEL}", "deploymentConfig": "system-sidekiq"},
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						v1.Volume{
							Name: "system-tmp",
							VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{
								Medium: v1.StorageMedium("Memory")}},
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
								createEnvVarFromValue("SLEEP_SECONDS", "1"),
								createEnvvarFromSecret("REDIS_URL", "system-redis", "URL"),
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
							ImagePullPolicy: v1.PullPolicy("IfNotPresent"),
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
				"app":                      "${APP_LABEL}",
				"3scale.component":         "system",
				"3scale.component-element": "app",
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
				"app":                      "${APP_LABEL}",
				"3scale.component":         "system",
				"3scale.component-element": "provider-ui",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "http",
					Protocol:   v1.Protocol("TCP"),
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
				"app":                      "${APP_LABEL}",
				"3scale.component":         "system",
				"3scale.component-element": "master-ui",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "http",
					Protocol:   v1.Protocol("TCP"),
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
				"app":                      "${APP_LABEL}",
				"3scale.component":         "system",
				"3scale.component-element": "developer-ui",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "http",
					Protocol:   v1.Protocol("TCP"),
					Port:       3000,
					TargetPort: intstr.FromString("developer"),
				},
			},
			Selector: map[string]string{"deploymentConfig": "system-app"},
		},
	}
}

func (system *System) buildSystemMysqlService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-mysql",
			Labels: map[string]string{
				"app":                      "${APP_LABEL}",
				"3scale.component":         "system",
				"3scale.component-element": "mysql",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "system-mysql",
					Protocol:   v1.Protocol("TCP"),
					Port:       3306,
					TargetPort: intstr.FromInt(3306),
					NodePort:   0,
				},
			},
			Selector: map[string]string{"deploymentConfig": "system-mysql"},
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
				"app":                      "${APP_LABEL}",
				"3scale.component":         "system",
				"3scale.component-element": "redis",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "redis",
					Protocol:   v1.Protocol("TCP"),
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
				"app":                      "${APP_LABEL}",
				"3scale.component":         "system",
				"3scale.component-element": "sphinx",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "sphinx",
					Protocol:   v1.Protocol("TCP"),
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
				"app":                      "${APP_LABEL}",
				"3scale.component":         "system",
				"3scale.component-element": "memcache",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "memcache",
					Protocol:   v1.Protocol("TCP"),
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
			Labels: map[string]string{"3scale.component": "system", "3scale.component-element": "smtp", "app": "${APP_LABEL}"},
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
			Labels: map[string]string{"app": "${APP_LABEL}", "3scale.component": "system", "3scale.component-element": "provider-ui"},
		},
		Spec: routev1.RouteSpec{
			Host: "${TENANT_NAME}-admin.${WILDCARD_DOMAIN}",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "system-provider",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("http"),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationType("edge"),
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyType("Allow")},
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
			Labels: map[string]string{"app": "${APP_LABEL}", "3scale.component": "system", "3scale.component-element": "master-ui"},
		},
		Spec: routev1.RouteSpec{
			Host: "${MASTER_NAME}.${WILDCARD_DOMAIN}",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "system-master",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("http"),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationType("edge"),
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyType("Allow")},
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
			Labels: map[string]string{"app": "${APP_LABEL}", "3scale.component": "system", "3scale.component-element": "developer-ui"},
		},
		Spec: routev1.RouteSpec{
			Host: "${TENANT_NAME}.${WILDCARD_DOMAIN}",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "system-developer",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("http"),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationType("edge"),
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyType("Allow")},
		},
	}
}

func (system *System) buildSystemConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system",
			Labels: map[string]string{
				"app":              "${APP_LABEL}",
				"3scale.component": "system",
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
  proxy_private_base_path: true
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
				"app":                      "${APP_LABEL}",
				"3scale.component":         "system",
				"3scale.component-element": "sphinx",
			},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerType("ConfigChange"),
				},
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerType("ImageChange"),
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
						Type:   intstr.Type(1),
						StrVal: "25%",
					},
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Type(1),
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
						"app":                      "${APP_LABEL}",
						"deploymentConfig":         "system-sphinx",
						"3scale.component":         "system",
						"3scale.component-element": "sphinx",
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
								createEnvVarFromValue("SLEEP_SECONDS", "1"),
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
