package component

import (
	"fmt"
	"strconv"

	"github.com/3scale/3scale-operator/pkg/helper"

	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	BackendListenerName = "backend-listener"
	BackendWorkerName   = "backend-worker"
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

const (
	BackendWorkerMetricsPort   = 9421
	BackendListenerMetricsPort = 9394
)

var (
	BackendWorkerMetricsPortStr   = strconv.FormatInt(BackendWorkerMetricsPort, 10)
	BackendListenerMetricsPortStr = strconv.FormatInt(BackendListenerMetricsPort, 10)
)

type Backend struct {
	Options *BackendOptions
}

func NewBackend(options *BackendOptions) *Backend {
	return &Backend{Options: options}
}

func (backend *Backend) WorkerDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   BackendWorkerName,
			Labels: backend.Options.CommonWorkerLabels,
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
						ContainerNames: []string{"backend-redis-svc", BackendWorkerName},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: fmt.Sprintf("amp-backend:%s", backend.Options.ImageTag)}}},
			},
			Replicas: backend.Options.WorkerReplicas,
			Selector: map[string]string{"deploymentConfig": BackendWorkerName},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: backend.Options.WorkerPodTemplateLabels,
				},
				Spec: v1.PodSpec{
					Affinity:    backend.Options.WorkerAffinity,
					Tolerations: backend.Options.WorkerTolerations,
					InitContainers: []v1.Container{
						v1.Container{
							Name:  "backend-redis-svc",
							Image: "amp-backend:latest",
							Command: []string{
								"/opt/app/entrypoint.sh",
								"sh",
								"-c",
								"until rake connectivity:redis_storage_queue_check; do sleep $SLEEP_SECONDS; done",
							}, Env: append(backend.buildBackendCommonEnv(), helper.EnvVarFromValue("SLEEP_SECONDS", "1")),
						},
					},
					Containers: []v1.Container{
						v1.Container{
							Name:            BackendWorkerName,
							Image:           "amp-backend:latest",
							Args:            []string{"bin/3scale_backend_worker", "run"},
							Env:             backend.buildBackendWorkerEnv(),
							Resources:       backend.Options.WorkerResourceRequirements,
							ImagePullPolicy: v1.PullIfNotPresent,
							Ports: []v1.ContainerPort{
								v1.ContainerPort{Name: "metrics", ContainerPort: BackendWorkerMetricsPort, Protocol: v1.ProtocolTCP},
							},
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
			Labels: backend.Options.CommonCronLabels,
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
							Name: fmt.Sprintf("amp-backend:%s", backend.Options.ImageTag)}}},
			},
			Replicas: backend.Options.CronReplicas,
			Selector: map[string]string{"deploymentConfig": "backend-cron"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: backend.Options.CronPodTemplateLabels,
				},
				Spec: v1.PodSpec{
					Affinity:    backend.Options.CronAffinity,
					Tolerations: backend.Options.CronTolerations,
					InitContainers: []v1.Container{
						v1.Container{
							Name:  "backend-redis-svc",
							Image: "amp-backend:latest",
							Command: []string{
								"/opt/app/entrypoint.sh",
								"sh",
								"-c",
								"until rake connectivity:redis_storage_queue_check; do sleep $SLEEP_SECONDS; done",
							}, Env: append(backend.buildBackendCommonEnv(), helper.EnvVarFromValue("SLEEP_SECONDS", "1")),
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
			Name:   BackendListenerName,
			Labels: backend.Options.CommonListenerLabels,
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
						ContainerNames: []string{BackendListenerName},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: fmt.Sprintf("amp-backend:%s", backend.Options.ImageTag)}}},
			},
			Replicas: backend.Options.ListenerReplicas,
			Selector: map[string]string{"deploymentConfig": BackendListenerName},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: backend.Options.ListenerPodTemplateLabels,
				},
				Spec: v1.PodSpec{
					Affinity:    backend.Options.ListenerAffinity,
					Tolerations: backend.Options.ListenerTolerations,
					Containers: []v1.Container{
						v1.Container{
							Name:  BackendListenerName,
							Image: "amp-backend:latest",
							Args:  []string{"bin/3scale_backend", "start", "-e", "production", "-p", "3000", "-x", "/dev/stdout"},
							Ports: []v1.ContainerPort{
								v1.ContainerPort{HostPort: 0,
									ContainerPort: 3000,
									Protocol:      v1.ProtocolTCP},
								v1.ContainerPort{Name: "metrics", ContainerPort: BackendListenerMetricsPort, Protocol: v1.ProtocolTCP},
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
			Name:   BackendListenerName,
			Labels: backend.Options.CommonListenerLabels,
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
			Selector: map[string]string{"deploymentConfig": BackendListenerName},
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
			Labels: backend.Options.CommonLabels,
		},
		Spec: routev1.RouteSpec{
			Host: "backend-" + backend.Options.TenantName + "." + backend.Options.WildcardDomain,
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: BackendListenerName,
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
			Labels: backend.Options.CommonLabels,
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
			Name:   BackendSecretBackendRedisSecretName,
			Labels: backend.Options.CommonLabels,
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
		helper.EnvVarFromSecret("CONFIG_REDIS_PROXY", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageURLFieldName),
		helper.EnvVarFromSecret("CONFIG_REDIS_SENTINEL_HOSTS", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageSentinelHostsFieldName),
		helper.EnvVarFromSecret("CONFIG_REDIS_SENTINEL_ROLE", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageSentinelRoleFieldName),
		helper.EnvVarFromSecret("CONFIG_QUEUES_MASTER_NAME", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisQueuesURLFieldName),
		helper.EnvVarFromSecret("CONFIG_QUEUES_SENTINEL_HOSTS", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisQueuesSentinelHostsFieldName),
		helper.EnvVarFromSecret("CONFIG_QUEUES_SENTINEL_ROLE", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisQueuesSentinelRoleFieldName),
		helper.EnvVarFromConfigMap("RACK_ENV", "backend-environment", "RACK_ENV"),
	}
}

func (backend *Backend) buildBackendWorkerEnv() []v1.EnvVar {
	result := []v1.EnvVar{}
	result = append(result, backend.buildBackendCommonEnv()...)
	result = append(result,
		helper.EnvVarFromSecret("CONFIG_EVENTS_HOOK", "system-events-hook", "URL"),
		helper.EnvVarFromSecret("CONFIG_EVENTS_HOOK_SHARED_SECRET", "system-events-hook", "PASSWORD"),
	)

	if backend.Options.WorkerMetrics {
		result = append(result,
			v1.EnvVar{Name: "CONFIG_WORKER_PROMETHEUS_METRICS_PORT", Value: BackendWorkerMetricsPortStr},
			v1.EnvVar{Name: "CONFIG_WORKER_PROMETHEUS_METRICS_ENABLED", Value: "true"},
		)
	}

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
		helper.EnvVarFromValue("PUMA_WORKERS", "16"),
		helper.EnvVarFromSecret("CONFIG_INTERNAL_API_USER", BackendSecretInternalApiSecretName, BackendSecretInternalApiUsernameFieldName),
		helper.EnvVarFromSecret("CONFIG_INTERNAL_API_PASSWORD", BackendSecretInternalApiSecretName, BackendSecretInternalApiPasswordFieldName),
	)

	if backend.Options.ListenerMetrics {
		result = append(result,
			v1.EnvVar{Name: "CONFIG_LISTENER_PROMETHEUS_METRICS_PORT", Value: BackendListenerMetricsPortStr},
			v1.EnvVar{Name: "CONFIG_LISTENER_PROMETHEUS_METRICS_ENABLED", Value: "true"},
		)
	}
	return result
}

func (backend *Backend) InternalAPISecretForSystem() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   BackendSecretInternalApiSecretName,
			Labels: backend.Options.CommonLabels,
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
			Name:   BackendSecretBackendListenerSecretName,
			Labels: backend.Options.CommonLabels,
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
			Name:   BackendWorkerName,
			Labels: backend.Options.CommonWorkerLabels,
		},
		Spec: v1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deploymentConfig": BackendWorkerName},
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
			Name:   "backend-cron",
			Labels: backend.Options.CommonCronLabels,
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
			Name:   BackendListenerName,
			Labels: backend.Options.CommonListenerLabels,
		},
		Spec: v1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deploymentConfig": BackendListenerName},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}
