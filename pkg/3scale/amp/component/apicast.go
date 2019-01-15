package component

import (
	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	templatev1 "github.com/openshift/api/template/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Apicast struct {
	options []string
}

func NewApicast(options []string) *Apicast {
	apicast := &Apicast{
		options: options,
	}
	return apicast
}

func (apicast *Apicast) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	apicast.buildParameters(template)
	apicast.buildObjects(template)
}

func (apicast *Apicast) PostProcess(template *templatev1.Template, otherComponents []Component) {

}

func (apicast *Apicast) buildParameters(template *templatev1.Template) {
	parameters := []templatev1.Parameter{
		{
			Name:        "APICAST_ACCESS_TOKEN",
			Description: "Read Only Access Token that is APIcast going to use to download its configuration.",
			Generate:    "expression",
			From:        "[a-z0-9]{8}",
			Required:    true,
		},
		{
			Name:        "APICAST_MANAGEMENT_API",
			Description: "Scope of the APIcast Management API. Can be disabled, status or debug. At least status required for health checks.",
			Value:       "status",
			Required:    false,
		},
		{
			Name:        "APICAST_OPENSSL_VERIFY",
			Description: "Turn on/off the OpenSSL peer verification when downloading the configuration. Can be set to true/false.",
			Value:       "false",
			Required:    false,
		},
		{
			Name:        "APICAST_RESPONSE_CODES",
			Description: "Enable logging response codes in APIcast.",
			Value:       "true",
			Required:    false,
		},
		{
			Name:        "APICAST_REGISTRY_URL",
			Description: "The URL to point to APIcast policies registry management",
			Value:       "http://apicast-staging:8090/policies",
			Required:    true,
		},
	}
	template.Parameters = append(template.Parameters, parameters...)
}

func (apicast *Apicast) buildObjects(template *templatev1.Template) {
	apicastStagingDeploymentConfig := apicast.buildApicastStagingDeploymentConfig()
	apicastProductionDeploymentConfig := apicast.buildApicastProductionDeploymentConfig()
	apicastStagingService := apicast.buildApicastStagingService()
	apicastProductionService := apicast.buildApicastProductionService()
	apicastStagingRoute := apicast.buildApicastStagingRoute()
	apicastProductionRoute := apicast.buildApicastProductionRoute()
	apicastEnvConfigMap := apicast.buildApicastEnvConfigMap()
	apicastRedisSecret := apicast.buildApicastRedisSecrets()

	objects := []runtime.RawExtension{
		{Object: apicastStagingDeploymentConfig},
		{Object: apicastProductionDeploymentConfig},
		{Object: apicastStagingService},
		{Object: apicastProductionService},
		{Object: apicastStagingRoute},
		{Object: apicastProductionRoute},
		{Object: apicastEnvConfigMap},
		{Object: apicastRedisSecret},
	}
	template.Objects = append(template.Objects, objects...)
}

func (apicast *Apicast) buildApicastStagingRoute() *routev1.Route {
	return &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "api-apicast-staging",
			Labels: map[string]string{"app": "${APP_LABEL}", "3scale.component": "apicast", "3scale.component-element": "staging"},
		},
		Spec: routev1.RouteSpec{
			Host: "api-${TENANT_NAME}-apicast-staging.${WILDCARD_DOMAIN}",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "apicast-staging",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("gateway"),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationType("edge"),
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyType("Allow")},
		},
	}
}

func (apicast *Apicast) buildApicastStagingService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "apicast-staging",
			Labels: map[string]string{
				"app":                      "${APP_LABEL}",
				"3scale.component":         "apicast",
				"3scale.component-element": "staging",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "gateway",
					Protocol:   v1.Protocol("TCP"),
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
				{
					Name:       "management",
					Protocol:   v1.Protocol("TCP"),
					Port:       8090,
					TargetPort: intstr.FromInt(8090),
				},
			},
			Selector: map[string]string{"deploymentConfig": "apicast-staging"},
		},
	}
}

func (apicast *Apicast) buildApicastProductionRoute() *routev1.Route {
	return &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "api-apicast-production",
			Labels: map[string]string{"app": "${APP_LABEL}", "3scale.component": "apicast", "3scale.component-element": "production"},
		},
		Spec: routev1.RouteSpec{
			Host: "api-${TENANT_NAME}-apicast-production.${WILDCARD_DOMAIN}",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "apicast-production",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("gateway"),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationType("edge"),
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyType("Allow")},
		},
	}
}

func (apicast *Apicast) buildApicastProductionService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "apicast-production",
			Labels: map[string]string{
				"app":                      "${APP_LABEL}",
				"3scale.component":         "apicast",
				"3scale.component-element": "production",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "gateway",
					Protocol:   v1.Protocol("TCP"),
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
				{
					Name:       "management",
					Protocol:   v1.Protocol("TCP"),
					Port:       8090,
					TargetPort: intstr.FromInt(8090),
				},
			},
			Selector: map[string]string{"deploymentConfig": "apicast-production"},
		},
	}
}

func (apicast *Apicast) buildApicastStagingDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: "apps.openshift.io/v1", Kind: "DeploymentConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "apicast-staging",
			Labels: map[string]string{
				"app":                      "${APP_LABEL}",
				"3scale.component":         "apicast",
				"3scale.component-element": "staging",
			},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: 1,
			Selector: map[string]string{
				"deploymentConfig": "apicast-staging",
			},
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
					TimeoutSeconds:      &[]int64{1800}[0],
					UpdatePeriodSeconds: &[]int64{1}[0],
				},
				Type: appsv1.DeploymentStrategyTypeRolling,
			},
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerType("ConfigChange"),
				},
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerType("ImageChange"),
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							"apicast-staging",
						},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "amp-apicast:latest",
						},
					},
				},
			},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"deploymentConfig":         "apicast-staging",
						"app":                      "${APP_LABEL}",
						"3scale.component":         "apicast",
						"3scale.component-element": "staging",
					},
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   "9421",
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "amp",
					Containers: []v1.Container{
						{
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 8080,
									Protocol:      v1.ProtocolTCP,
								},
								{
									ContainerPort: 8090,
									Protocol:      v1.ProtocolTCP,
								},
								{
									ContainerPort: 9421,
									Protocol:      v1.ProtocolTCP,
									Name:          "metrics",
								},
							},
							Env:             apicast.buildApicastStagingEnv(),
							Image:           "amp-apicast:latest",
							ImagePullPolicy: v1.PullPolicy("IfNotPresent"),
							Name:            "apicast-staging",
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("100m"),
									v1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("50m"),
									v1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{HTTPGet: &v1.HTTPGetAction{
									Path: "/status/live",
									Port: intstr.FromInt(8090),
								}},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      5,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{HTTPGet: &v1.HTTPGetAction{
									Path: "/status/ready",
									Port: intstr.FromInt(8090),
								}},
								InitialDelaySeconds: 15,
								TimeoutSeconds:      5,
								PeriodSeconds:       30,
							},
						},
					},
				},
			},
		},
	}
}

func (apicast *Apicast) buildApicastProductionDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: "apps.openshift.io/v1", Kind: "DeploymentConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "apicast-production",
			Labels: map[string]string{
				"app":                      "${APP_LABEL}",
				"3scale.component":         "apicast",
				"3scale.component-element": "production",
			},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: 1,
			Selector: map[string]string{
				"deploymentConfig": "apicast-production",
			},
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
					TimeoutSeconds:      &[]int64{1800}[0],
					UpdatePeriodSeconds: &[]int64{1}[0],
				},
				Type: appsv1.DeploymentStrategyTypeRolling,
			},
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
							"apicast-production",
						},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "amp-apicast:latest",
						},
					},
				},
			},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"deploymentConfig":         "apicast-production",
						"app":                      "${APP_LABEL}",
						"3scale.component":         "apicast",
						"3scale.component-element": "production",
					},
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   "9421",
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "amp",
					InitContainers: []v1.Container{
						{
							Name:    "system-master-svc",
							Image:   "amp-apicast:latest",
							Command: []string{"sh", "-c", "until $(curl --output /dev/null --silent --fail --head http://system-master:3000/status); do sleep $SLEEP_SECONDS; done"},
							Env: []v1.EnvVar{
								{
									Name:  "SLEEP_SECONDS",
									Value: "1",
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 8080,
									Protocol:      v1.ProtocolTCP,
								},
								{
									ContainerPort: 8090,
									Protocol:      v1.ProtocolTCP,
								},
								{
									ContainerPort: 9421,
									Protocol:      v1.ProtocolTCP,
									Name:          "metrics",
								},
							},
							Env:             apicast.buildApicastProductionEnv(),
							Image:           "amp-apicast:latest",
							ImagePullPolicy: v1.PullPolicy("IfNotPresent"),
							Name:            "apicast-production",
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1000m"),
									v1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("500m"),
									v1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{HTTPGet: &v1.HTTPGetAction{
									Path: "/status/live",
									Port: intstr.FromInt(8090),
								}},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      5,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{HTTPGet: &v1.HTTPGetAction{
									Path: "/status/ready",
									Port: intstr.FromInt(8090),
								}},
								InitialDelaySeconds: 15,
								TimeoutSeconds:      5,
								PeriodSeconds:       30,
							},
						},
					},
				},
			},
		},
	}
}

func (apicast *Apicast) buildApicastCommonEnv() []v1.EnvVar {
	return []v1.EnvVar{
		createEnvvarFromSecret("THREESCALE_PORTAL_ENDPOINT", "system-master-apicast", "PROXY_CONFIGS_ENDPOINT"),
		createEnvvarFromSecret("BACKEND_ENDPOINT_OVERRIDE", "backend-listener", "service_endpoint"),
		createEnvVarFromConfigMap("APICAST_MANAGEMENT_API", "apicast-environment", "APICAST_MANAGEMENT_API"),
		createEnvVarFromConfigMap("OPENSSL_VERIFY", "apicast-environment", "OPENSSL_VERIFY"),
		createEnvVarFromConfigMap("APICAST_RESPONSE_CODES", "apicast-environment", "APICAST_RESPONSE_CODES"),
	}
}

func (apicast *Apicast) buildApicastStagingEnv() []v1.EnvVar {
	result := []v1.EnvVar{}
	result = append(result, apicast.buildApicastCommonEnv()...)
	result = append(result,
		createEnvVarFromValue("APICAST_CONFIGURATION_LOADER", "lazy"),
		createEnvVarFromValue("APICAST_CONFIGURATION_CACHE", "0"),
		createEnvVarFromValue("THREESCALE_DEPLOYMENT_ENV", "staging"),
		createEnvvarFromSecret("REDIS_URL", "apicast-redis", "STAGING_URL"),
	)
	return result
}

func (apicast *Apicast) buildApicastProductionEnv() []v1.EnvVar {
	result := []v1.EnvVar{}
	result = append(result, apicast.buildApicastCommonEnv()...)
	result = append(result,
		createEnvVarFromValue("APICAST_CONFIGURATION_LOADER", "boot"),
		createEnvVarFromValue("APICAST_CONFIGURATION_CACHE", "300"),
		createEnvVarFromValue("THREESCALE_DEPLOYMENT_ENV", "production"),
		createEnvvarFromSecret("REDIS_URL", "apicast-redis", "PRODUCTION_URL"),
	)
	return result
}

func (apicast *Apicast) buildApicastEnvConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "apicast-environment",
			Labels: map[string]string{"3scale.component": "apicast", "app": "${APP_LABEL}"},
		},
		Data: map[string]string{
			"APICAST_MANAGEMENT_API": "${APICAST_MANAGEMENT_API}",
			"OPENSSL_VERIFY":         "${APICAST_OPENSSL_VERIFY}",
			"APICAST_RESPONSE_CODES": "${APICAST_RESPONSE_CODES}",
		},
	}
}

func (apicast *Apicast) buildApicastRedisSecrets() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "apicast-redis",
			Labels: map[string]string{
				"app":              "${APP_LABEL}",
				"3scale.component": "apicast",
			},
		},
		StringData: map[string]string{
			"PRODUCTION_URL": "redis://system-redis:6379/1",
			"STAGING_URL":    "redis://system-redis:6379/2",
		},
		Type: v1.SecretTypeOpaque,
	}
}
