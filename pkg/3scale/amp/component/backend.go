package component

import (
	"github.com/3scale/3scale-operator/pkg/common"
	"k8s.io/api/policy/v1beta1"

	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
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
	Options *BackendOptions
}

func NewBackend(options *BackendOptions) *Backend {
	return &Backend{Options: options}
}

func (backend *Backend) Objects() []common.KubernetesObject {
	cronDeploymentConfig := backend.CronDeploymentConfig()
	listenerDeploymentConfig := backend.ListenerDeploymentConfig()
	listenerService := backend.ListenerService()
	listenerRoute := backend.ListenerRoute()
	workerDeploymentConfig := backend.WorkerDeploymentConfig()
	environmentConfigMap := backend.EnvironmentConfigMap()

	internalAPISecretForSystem := backend.InternalAPISecretForSystem()
	redisSecret := backend.RedisSecret()
	listenerSecret := backend.ListenerSecret()

	objects := []common.KubernetesObject{
		cronDeploymentConfig,
		listenerDeploymentConfig,
		listenerService,
		listenerRoute,
		workerDeploymentConfig,
		environmentConfigMap,
		internalAPISecretForSystem,
		redisSecret,
		listenerSecret,
	}
	return objects
}

func (backend *Backend) PDBObjects() []common.KubernetesObject {
	workerPDB := backend.WorkerPodDisruptionBudget()
	cronPDB := backend.CronPodDisruptionBudget()
	listenerPDB := backend.ListenerPodDisruptionBudget()
	return []common.KubernetesObject{
		workerPDB,
		cronPDB,
		listenerPDB,
	}
}

func (backend *Backend) WorkerDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "backend-worker",
			Labels: map[string]string{"threescale_component": "backend", "threescale_component_element": "worker", "app": backend.Options.AppLabel},
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
			Replicas: backend.Options.WorkerReplicas,
			Selector: map[string]string{"deploymentConfig": "backend-worker"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"threescale_component": "backend", "threescale_component_element": "worker", "app": backend.Options.AppLabel, "deploymentConfig": "backend-worker"},
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
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
							Name:            "backend-worker",
							Image:           "amp-backend:latest",
							Args:            []string{"bin/3scale_backend_worker", "run"},
							Env:             backend.buildBackendWorkerEnv(),
							Resources:       backend.Options.WorkerResourceRequirements,
							ImagePullPolicy: v1.PullIfNotPresent,
						},
					},
					ServiceAccountName: "amp"}},
		},
	}
}

func (backend *Backend) CronDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "backend-cron",
			Labels: map[string]string{"threescale_component": "backend", "threescale_component_element": "cron", "app": backend.Options.AppLabel},
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
			Replicas: backend.Options.CronReplicas,
			Selector: map[string]string{"deploymentConfig": "backend-cron"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"threescale_component": "backend", "threescale_component_element": "cron", "app": backend.Options.AppLabel, "deploymentConfig": "backend-cron"},
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
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
							Name:            "backend-cron",
							Image:           "amp-backend:latest",
							Args:            []string{"backend-cron"},
							Env:             backend.buildBackendCronEnv(),
							Resources:       backend.Options.CronResourceRequirements,
							ImagePullPolicy: v1.PullIfNotPresent,
						},
					},
					ServiceAccountName: "amp",
				}},
		},
	}
}

func (backend *Backend) ListenerDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "backend-listener",
			Labels: map[string]string{"threescale_component": "backend", "threescale_component_element": "listener", "app": backend.Options.AppLabel},
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
			Replicas: backend.Options.ListenerReplicas,
			Selector: map[string]string{"deploymentConfig": "backend-listener"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"threescale_component": "backend", "threescale_component_element": "listener", "app": backend.Options.AppLabel, "deploymentConfig": "backend-listener"},
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
						Env:       backend.buildBackendListenerEnv(),
						Resources: backend.Options.ListenerResourceRequirements,
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

func (backend *Backend) ListenerService() *v1.Service {
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
				"app":                          backend.Options.AppLabel,
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

func (backend *Backend) ListenerRoute() *routev1.Route {
	return &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "backend",
			Labels: map[string]string{"app": backend.Options.AppLabel, "threescale_component": "backend"},
		},
		Spec: routev1.RouteSpec{
			Host: "backend-" + backend.Options.TenantName + "." + backend.Options.WildcardDomain,
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

func (backend *Backend) EnvironmentConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "backend-environment",
			Labels: map[string]string{"threescale_component": "backend", "app": backend.Options.AppLabel},
		},
		Data: map[string]string{
			"RACK_ENV": "production",
		},
	}
}

func (backend *Backend) RedisSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: BackendSecretBackendRedisSecretName,
			Labels: map[string]string{
				"app":                  backend.Options.AppLabel,
				"threescale_component": "backend",
			},
		},
		StringData: map[string]string{
			BackendSecretBackendRedisStorageURLFieldName:           backend.Options.StorageURL,
			BackendSecretBackendRedisQueuesURLFieldName:            backend.Options.QueuesURL,
			BackendSecretBackendRedisStorageSentinelHostsFieldName: backend.Options.StorageSentinelHosts,
			BackendSecretBackendRedisStorageSentinelRoleFieldName:  backend.Options.StorageSentinelRole,
			BackendSecretBackendRedisQueuesSentinelHostsFieldName:  backend.Options.QueuesSentinelHosts,
			BackendSecretBackendRedisQueuesSentinelRoleFieldName:   backend.Options.QueuesSentinelRole,
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

func (backend *Backend) InternalAPISecretForSystem() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: BackendSecretInternalApiSecretName,
			Labels: map[string]string{
				"app":                  backend.Options.AppLabel,
				"threescale_component": "backend",
			},
		},
		StringData: map[string]string{
			BackendSecretInternalApiUsernameFieldName: backend.Options.SystemBackendUsername,
			BackendSecretInternalApiPasswordFieldName: backend.Options.SystemBackendPassword,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (backend *Backend) ListenerSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: BackendSecretBackendListenerSecretName,
			Labels: map[string]string{
				"app":                  backend.Options.AppLabel,
				"threescale_component": "backend",
			},
		},
		StringData: map[string]string{
			BackendSecretBackendListenerServiceEndpointFieldName: backend.Options.ServiceEndpoint,
			BackendSecretBackendListenerRouteEndpointFieldName:   backend.Options.RouteEndpoint,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (backend *Backend) WorkerPodDisruptionBudget() *v1beta1.PodDisruptionBudget {
	return &v1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "backend-worker",
			Labels: map[string]string{
				"app":                          backend.Options.AppLabel,
				"threescale_component":         "backend",
				"threescale_component_element": "worker",
			},
		},
		Spec: v1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deploymentConfig": "backend-worker"},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}

func (backend *Backend) CronPodDisruptionBudget() *v1beta1.PodDisruptionBudget {
	return &v1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "backend-cron",
			Labels: map[string]string{
				"app":                          backend.Options.AppLabel,
				"threescale_component":         "backend",
				"threescale_component_element": "cron",
			},
		},
		Spec: v1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deploymentConfig": "backend-cron"},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}

func (backend *Backend) ListenerPodDisruptionBudget() *v1beta1.PodDisruptionBudget {
	return &v1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "backend-listener",
			Labels: map[string]string{
				"app":                          backend.Options.AppLabel,
				"threescale_component":         "backend",
				"threescale_component_element": "listener",
			},
		},
		Spec: v1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deploymentConfig": "backend-listener"},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}
