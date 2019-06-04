package component

import (
	"fmt"
	"github.com/3scale/3scale-operator/pkg/apis/common"
	"github.com/3scale/3scale-operator/pkg/helper"

	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	BackendSecretBackendRedisSecretName                    = "backend-redis"
	BackendSecretBackendRedisStorageURLFieldName           = "REDIS_STORAGE_URL"
	BackendSecretBackendRedisQueuesURLFieldName            = "REDIS_QUEUES_URL"
	BackendSecretBackendRedisStorageSentinelHostsFieldName = "REDIS_STORAGE_SENTINEL_HOSTS"
	BackendSecretBackendRedisStorageSentinelRoleFieldName  = "REDIS_STORAGE_SENTINEL_ROLE"
	BackendSecretBackendRedisQueuesSentinelHostsFieldName  = "REDIS_QUEUES_SENTINEL_HOSTS"
	BackendSecretBackendRedisQueuesSentinelRoleFieldName   = "REDIS_QUEUES_SENTINEL_ROLE"
)

const (
	BackendSecretInternalApiSecretName        = "backend-internal-api"
	BackendSecretInternalApiUsernameFieldName = "username"
	BackendSecretInternalApiPasswordFieldName = "password"
)

const (
	BackendSecretBackendListenerSecretName               = "backend-listener"
	BackendSecretBackendListenerServiceEndpointFieldName = "service_endpoint"
	BackendSecretBackendListenerRouteEndpointFieldName   = "route_endpoint"
)

type Backend struct {
	options []string
	Options *BackendOptions
}

type backendRequiredOptions struct {
	appLabel              string
	systemBackendUsername string
	systemBackendPassword string
	tenantName            string
	wildcardDomain        string
}

type backendNonRequiredOptions struct {
	serviceEndpoint      *string
	routeEndpoint        *string
	storageURL           *string
	queuesURL            *string
	storageSentinelHosts *string
	storageSentinelRole  *string
	queuesSentinelHosts  *string
	queuesSentinelRole   *string
}

type BackendOptions struct {
	backendNonRequiredOptions
	backendRequiredOptions
}

func NewBackend(options []string) *Backend {
	backend := &Backend{
		options: options,
	}
	return backend
}

type BackendOptionsProvider interface {
	GetBackendOptions() *BackendOptions
}
type CLIBackendOptionsProvider struct {
}

func (o *CLIBackendOptionsProvider) GetBackendOptions() (*BackendOptions, error) {
	bob := BackendOptionsBuilder{}
	bob.AppLabel("${APP_LABEL}")
	bob.SystemBackendUsername("${SYSTEM_BACKEND_USERNAME}")
	bob.SystemBackendPassword("${SYSTEM_BACKEND_PASSWORD}")
	bob.TenantName("${TENANT_NAME}")
	bob.WildcardDomain("${WILDCARD_DOMAIN}")
	res, err := bob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Backend Options - %s", err)
	}
	return res, nil
}

func (backend *Backend) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	backend.buildParameters(template)

	// TODO move this outside this specific method
	optionsProvider := CLIBackendOptionsProvider{}
	backendOpts, err := optionsProvider.GetBackendOptions()
	_ = err
	backend.Options = backendOpts
	backend.buildParameters(template)
	backend.addObjectsIntoTemplate(template)
}

func (backend *Backend) GetObjects() ([]common.KubernetesObject, error) {
	objects := backend.buildObjects()
	return objects, nil
}

func (backend *Backend) addObjectsIntoTemplate(template *templatev1.Template) {
	objects := backend.buildObjects()
	template.Objects = append(template.Objects, helper.WrapRawExtensions(objects)...)
}

func (backend *Backend) PostProcess(template *templatev1.Template, otherComponents []Component) {

}

func (backend *Backend) buildParameters(template *templatev1.Template) {
	parameters := []templatev1.Parameter{}
	template.Parameters = append(template.Parameters, parameters...)
}

func (backend *Backend) buildObjects() []common.KubernetesObject {
	backendCronDeploymentConfig := backend.buildBackendCronDeploymentConfig()
	backendListenerDeploymentConfig := backend.buildBackendListenerDeploymentConfig()
	backendListenerService := backend.buildBackendListenerService()
	backendListenerRoute := backend.buildBackendListenerRoute()
	backendWorkerDeploymentConfig := backend.buildBackendWorkerDeploymentConfig()
	backendEnvConfigMap := backend.buildBackendEnvConfigMap()

	backendInternalApiCredsForSystem := backend.buildBackendInternalApiCredsForSystem()
	backendRedisSecrets := backend.buildBackendRedisSecrets()
	backendListenerSecrets := backend.buildBackendListenerSecrets()

	objects := []common.KubernetesObject{
		backendCronDeploymentConfig,
		backendListenerDeploymentConfig,
		backendListenerService,
		backendListenerRoute,
		backendWorkerDeploymentConfig,
		backendEnvConfigMap,
		backendInternalApiCredsForSystem,
		backendRedisSecrets,
		backendListenerSecrets,
	}
	return objects
}

func (backend *Backend) buildBackendWorkerDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "backend-worker",
			Labels: map[string]string{"threescale_component": "backend", "threescale_component_element": "worker", "app": backend.Options.appLabel},
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
						ContainerNames: []string{"backend-redis-svc", "backend-worker"},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "amp-backend:latest"}}},
			},
			Replicas: 1,
			Selector: map[string]string{"deploymentConfig": "backend-worker"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"threescale_component": "backend", "threescale_component_element": "worker", "app": backend.Options.appLabel, "deploymentConfig": "backend-worker"},
				},
				Spec: v1.PodSpec{InitContainers: []v1.Container{
					v1.Container{
						Name:  "backend-redis-svc",
						Image: "amp-backend:latest",
						Command: []string{
							"/opt/app/entrypoint.sh",
							"sh",
							"-c",
							"until rake connectivity:redis_storage_queue_check; do sleep $SLEEP_SECONDS; done",
						}, Env: append(backend.buildBackendCommonEnv(), envVarFromValue("SLEEP_SECONDS", "1")),
					},
				},
					Containers: []v1.Container{
						v1.Container{
							Name:  "backend-worker",
							Image: "amp-backend:latest",
							Args:  []string{"bin/3scale_backend_worker", "run"},
							Env:   backend.buildBackendWorkerEnv(),
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1000m"),
									v1.ResourceMemory: resource.MustParse("300Mi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("150m"),
									v1.ResourceMemory: resource.MustParse("50Mi"),
								},
							},
							ImagePullPolicy: v1.PullIfNotPresent,
						},
					},
					ServiceAccountName: "amp"}},
		},
	}
}

func (backend *Backend) buildBackendCronDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "backend-cron",
			Labels: map[string]string{"threescale_component": "backend", "threescale_component_element": "cron", "app": backend.Options.appLabel},
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
						ContainerNames: []string{"backend-redis-svc", "backend-cron"},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "amp-backend:latest"}}},
			},
			Replicas: 1,
			Selector: map[string]string{"deploymentConfig": "backend-cron"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"threescale_component": "backend", "threescale_component_element": "cron", "app": backend.Options.appLabel, "deploymentConfig": "backend-cron"},
				},
				Spec: v1.PodSpec{InitContainers: []v1.Container{
					v1.Container{
						Name:  "backend-redis-svc",
						Image: "amp-backend:latest",
						Command: []string{
							"/opt/app/entrypoint.sh",
							"sh",
							"-c",
							"until rake connectivity:redis_storage_queue_check; do sleep $SLEEP_SECONDS; done",
						}, Env: append(backend.buildBackendCommonEnv(), envVarFromValue("SLEEP_SECONDS", "1")),
					},
				},
					Containers: []v1.Container{
						v1.Container{
							Name:  "backend-cron",
							Image: "amp-backend:latest",
							Args:  []string{"backend-cron"},
							Env:   backend.buildBackendCronEnv(),
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("150m"),
									v1.ResourceMemory: resource.MustParse("80Mi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("50m"),
									v1.ResourceMemory: resource.MustParse("40Mi"),
								},
							},

							ImagePullPolicy: v1.PullIfNotPresent,
						},
					},
					ServiceAccountName: "amp",
				}},
		},
	}
}

func (backend *Backend) buildBackendListenerDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "backend-listener",
			Labels: map[string]string{"threescale_component": "backend", "threescale_component_element": "listener", "app": backend.Options.appLabel},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRolling,
				RollingParams: &appsv1.RollingDeploymentStrategyParams{
					UpdatePeriodSeconds: &[]int64{1}[0],
					IntervalSeconds:     &[]int64{1}[0],
					TimeoutSeconds:      &[]int64{600}[0],
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
						ContainerNames: []string{"backend-listener"},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "amp-backend:latest"}}},
			},
			Replicas: 1,
			Selector: map[string]string{"deploymentConfig": "backend-listener"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"threescale_component": "backend", "threescale_component_element": "listener", "app": backend.Options.appLabel, "deploymentConfig": "backend-listener"},
				},
				Spec: v1.PodSpec{Containers: []v1.Container{
					v1.Container{
						Name:  "backend-listener",
						Image: "amp-backend:latest",
						Args:  []string{"bin/3scale_backend", "start", "-e", "production", "-p", "3000", "-x", "/dev/stdout"},
						Ports: []v1.ContainerPort{
							v1.ContainerPort{HostPort: 0,
								ContainerPort: 3000,
								Protocol:      v1.ProtocolTCP},
						},
						Env: backend.buildBackendListenerEnv(),
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("1000m"),
								v1.ResourceMemory: resource.MustParse("700Mi"),
							},
							Requests: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("500m"),
								v1.ResourceMemory: resource.MustParse("550Mi"),
							},
						},

						LivenessProbe: &v1.Probe{
							Handler: v1.Handler{TCPSocket: &v1.TCPSocketAction{
								Port: intstr.IntOrString{
									Type:   intstr.Type(intstr.Int),
									IntVal: 3000}},
							},
							InitialDelaySeconds: 30,
							TimeoutSeconds:      0,
							PeriodSeconds:       10,
							SuccessThreshold:    0,
							FailureThreshold:    0,
						},
						ReadinessProbe: &v1.Probe{
							Handler: v1.Handler{HTTPGet: &v1.HTTPGetAction{
								Path: "/status",
								Port: intstr.IntOrString{
									Type:   intstr.Type(intstr.Int),
									IntVal: 3000}},
							},
							InitialDelaySeconds: 30,
							TimeoutSeconds:      5,
							PeriodSeconds:       0,
							SuccessThreshold:    0,
							FailureThreshold:    0,
						},
						ImagePullPolicy: v1.PullIfNotPresent,
					},
				},
					ServiceAccountName: "amp",
				}},
		},
	}
}

func (backend *Backend) buildBackendListenerService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "backend-listener",
			Labels: map[string]string{
				"threescale_component":         "backend",
				"threescale_component_element": "listener",
				"app":                          backend.Options.appLabel,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:     "http",
					Protocol: v1.ProtocolTCP,
					Port:     3000,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Type(intstr.Int),
						IntVal: 3000,
					},
				},
			},
			Selector: map[string]string{"deploymentConfig": "backend-listener"},
		},
	}
}

func (backend *Backend) buildBackendListenerRoute() *routev1.Route {
	return &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "backend",
			Labels: map[string]string{"app": backend.Options.appLabel, "threescale_component": "backend"},
		},
		Spec: routev1.RouteSpec{
			Host: "backend-" + backend.Options.tenantName + "." + backend.Options.wildcardDomain,
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "backend-listener",
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

func (backend *Backend) buildBackendEnvConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "backend-environment",
			Labels: map[string]string{"threescale_component": "backend", "app": backend.Options.appLabel},
		},
		Data: map[string]string{
			"RACK_ENV": "production",
		},
	}
}

func (backend *Backend) buildBackendRedisSecrets() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: BackendSecretBackendRedisSecretName,
			Labels: map[string]string{
				"app":                  backend.Options.appLabel,
				"threescale_component": "backend",
			},
		},
		StringData: map[string]string{
			BackendSecretBackendRedisStorageURLFieldName:           *backend.Options.storageURL,
			BackendSecretBackendRedisQueuesURLFieldName:            *backend.Options.queuesURL,
			BackendSecretBackendRedisStorageSentinelHostsFieldName: *backend.Options.storageSentinelHosts,
			BackendSecretBackendRedisStorageSentinelRoleFieldName:  *backend.Options.storageSentinelRole,
			BackendSecretBackendRedisQueuesSentinelHostsFieldName:  *backend.Options.queuesSentinelHosts,
			BackendSecretBackendRedisQueuesSentinelRoleFieldName:   *backend.Options.queuesSentinelRole,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (backend *Backend) buildBackendCommonEnv() []v1.EnvVar {
	return []v1.EnvVar{
		envVarFromSecret("CONFIG_REDIS_PROXY", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageURLFieldName),
		envVarFromSecret("CONFIG_REDIS_SENTINEL_HOSTS", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageSentinelHostsFieldName),
		envVarFromSecret("CONFIG_REDIS_SENTINEL_ROLE", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageSentinelRoleFieldName),
		envVarFromSecret("CONFIG_QUEUES_MASTER_NAME", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisQueuesURLFieldName),
		envVarFromSecret("CONFIG_QUEUES_SENTINEL_HOSTS", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisQueuesSentinelHostsFieldName),
		envVarFromSecret("CONFIG_QUEUES_SENTINEL_ROLE", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisQueuesSentinelRoleFieldName),
		envVarFromConfigMap("RACK_ENV", "backend-environment", "RACK_ENV"),
	}
}

func (backend *Backend) buildBackendWorkerEnv() []v1.EnvVar {
	result := []v1.EnvVar{}
	result = append(result, backend.buildBackendCommonEnv()...)
	result = append(result,
		envVarFromSecret("CONFIG_EVENTS_HOOK", "system-events-hook", "URL"),
		envVarFromSecret("CONFIG_EVENTS_HOOK_SHARED_SECRET", "system-events-hook", "PASSWORD"),
	)
	return result
}

func (backend *Backend) buildBackendCronEnv() []v1.EnvVar {
	result := []v1.EnvVar{}
	result = append(result, backend.buildBackendCommonEnv()...)
	return result
}

func (backend *Backend) buildBackendListenerEnv() []v1.EnvVar {
	result := []v1.EnvVar{}
	result = append(result, backend.buildBackendCommonEnv()...)
	result = append(result,
		envVarFromValue("PUMA_WORKERS", "16"),
		envVarFromSecret("CONFIG_INTERNAL_API_USER", BackendSecretInternalApiSecretName, BackendSecretInternalApiUsernameFieldName),
		envVarFromSecret("CONFIG_INTERNAL_API_PASSWORD", BackendSecretInternalApiSecretName, BackendSecretInternalApiPasswordFieldName),
	)
	return result
}

func (backend *Backend) buildBackendInternalApiCredsForSystem() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: BackendSecretInternalApiSecretName,
			Labels: map[string]string{
				"app":                  backend.Options.appLabel,
				"threescale_component": "backend",
			},
		},
		StringData: map[string]string{
			BackendSecretInternalApiUsernameFieldName: backend.Options.systemBackendUsername,
			BackendSecretInternalApiPasswordFieldName: backend.Options.systemBackendPassword,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (backend *Backend) buildBackendListenerSecrets() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: BackendSecretBackendListenerSecretName,
			Labels: map[string]string{
				"app":                  backend.Options.appLabel,
				"threescale_component": "backend",
			},
		},
		StringData: map[string]string{
			BackendSecretBackendListenerServiceEndpointFieldName: *backend.Options.serviceEndpoint,
			BackendSecretBackendListenerRouteEndpointFieldName:   *backend.Options.routeEndpoint,
		},
		Type: v1.SecretTypeOpaque,
	}
}
