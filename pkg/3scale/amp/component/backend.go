package component

import (
	"strconv"

	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	routev1 "github.com/openshift/api/route/v1"
	k8sappsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	BackendListenerName      = "backend-listener"
	BackendWorkerName        = "backend-worker"
	BackendCronName          = "backend-cron"
	BackendInitContainerName = "backend-redis-svc"
)

const (
	BackendSecretBackendRedisSecretName                    = "backend-redis"
	BackendSecretBackendRedisStorageURLFieldName           = "REDIS_STORAGE_URL"
	BackendSecretBackendRedisQueuesURLFieldName            = "REDIS_QUEUES_URL"
	BackendSecretBackendRedisStorageSentinelHostsFieldName = "REDIS_STORAGE_SENTINEL_HOSTS"
	BackendSecretBackendRedisStorageSentinelRoleFieldName  = "REDIS_STORAGE_SENTINEL_ROLE"
	BackendSecretBackendRedisQueuesSentinelHostsFieldName  = "REDIS_QUEUES_SENTINEL_HOSTS"
	BackendSecretBackendRedisQueuesSentinelRoleFieldName   = "REDIS_QUEUES_SENTINEL_ROLE"

	// TLS
	BackendSecretBackendRedisConfigCAFile                  = "CONFIG_REDIS_CA_FILE"
	BackendSecretBackendRedisConfigClientCertificate       = "CONFIG_REDIS_CERT"
	BackendSecretBackendRedisConfigPrivateKey              = "CONFIG_REDIS_PRIVATE_KEY"
	BackendSecretBackendRedisConfigSSL                     = "CONFIG_REDIS_SSL"
	BackendSecretBackendRedisConfigQueuesCAFile            = "CONFIG_QUEUES_CA_FILE"
	BackendSecretBackendRedisConfigQueuesClientCertificate = "CONFIG_QUEUES_CERT"
	BackendSecretBackendRedisConfigQueuesPrivateKey        = "CONFIG_QUEUES_PRIVATE_KEY"
	BackendSecretBackendRedisConfigQueuesSSL               = "CONFIG_QUEUES_SSL"
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

func (backend *Backend) WorkerDeployment(containerImage string) *k8sappsv1.Deployment {
	return &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: reconcilers.DeploymentAPIVersion, Kind: reconcilers.DeploymentKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:   BackendWorkerName,
			Labels: backend.Options.CommonWorkerLabels,
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
			Replicas:        &backend.Options.WorkerReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					reconcilers.DeploymentLabelSelector: BackendWorkerName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      backend.Options.WorkerPodTemplateLabels,
					Annotations: backend.Options.WorkerPodTemplateAnnotations,
				},
				Spec: v1.PodSpec{
					Affinity:    backend.Options.WorkerAffinity,
					Tolerations: backend.Options.WorkerTolerations,
					InitContainers: []v1.Container{
						{
							Name:  "backend-redis-svc",
							Image: containerImage,
							Command: []string{
								"/opt/app/entrypoint.sh",
								"sh",
								"-c",
								"until rake connectivity:redis_storage_queue_check; do sleep $SLEEP_SECONDS; done",
							},
							Env: append(backend.buildBackendCommonEnv(), helper.EnvVarFromValue("SLEEP_SECONDS", "1")),
						},
					},
					Containers: []v1.Container{
						{
							Name:            BackendWorkerName,
							Image:           containerImage,
							Args:            []string{"bin/3scale_backend_worker", "run"},
							Env:             backend.buildBackendWorkerEnv(),
							Resources:       backend.Options.WorkerResourceRequirements,
							ImagePullPolicy: v1.PullIfNotPresent,
							Ports:           backend.workerPorts(),
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Port:   intstr.FromInt32(9421),
										Path:   "/metrics",
										Scheme: v1.URISchemeHTTP,
									},
								},
								FailureThreshold:    3,
								InitialDelaySeconds: 30,
								PeriodSeconds:       30,
								SuccessThreshold:    1,
								TimeoutSeconds:      60,
							},
						},
					},
					ServiceAccountName:        "amp",
					PriorityClassName:         backend.Options.PriorityClassNameWorker,
					TopologySpreadConstraints: backend.Options.TopologySpreadConstraintsWorker,
				},
			},
		},
	}
}

func (backend *Backend) CronDeployment(containerImage string) *k8sappsv1.Deployment {
	return &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: reconcilers.DeploymentAPIVersion, Kind: reconcilers.DeploymentKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:   BackendCronName,
			Labels: backend.Options.CommonCronLabels,
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
			Replicas:        &backend.Options.CronReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					reconcilers.DeploymentLabelSelector: BackendCronName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      backend.Options.CronPodTemplateLabels,
					Annotations: backend.Options.CronPodTemplateAnnotations,
				},
				Spec: v1.PodSpec{
					Affinity:    backend.Options.CronAffinity,
					Tolerations: backend.Options.CronTolerations,
					InitContainers: []v1.Container{
						{
							Name:  "backend-redis-svc",
							Image: containerImage,
							Command: []string{
								"/opt/app/entrypoint.sh",
								"sh",
								"-c",
								"until rake connectivity:redis_storage_queue_check; do sleep $SLEEP_SECONDS; done",
							},
							Env: append(backend.buildBackendCommonEnv(), helper.EnvVarFromValue("SLEEP_SECONDS", "1")),
						},
					},
					Containers: []v1.Container{
						{
							Name:            "backend-cron",
							Image:           containerImage,
							Args:            []string{"touch /tmp/healthy && backend-cron"},
							Env:             backend.buildBackendCronEnv(),
							Resources:       backend.Options.CronResourceRequirements,
							ImagePullPolicy: v1.PullIfNotPresent,
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									Exec: &v1.ExecAction{
										Command: []string{"cat", "/tmp/healthy"},
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       5,
							},
						},
					},
					ServiceAccountName:        "amp",
					PriorityClassName:         backend.Options.PriorityClassNameCron,
					TopologySpreadConstraints: backend.Options.TopologySpreadConstraintsCron,
				},
			},
		},
	}
}

func (backend *Backend) ListenerDeployment(containerImage string) *k8sappsv1.Deployment {
	return &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: reconcilers.DeploymentAPIVersion, Kind: reconcilers.DeploymentKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:   BackendListenerName,
			Labels: backend.Options.CommonListenerLabels,
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
			Replicas:        &backend.Options.ListenerReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					reconcilers.DeploymentLabelSelector: BackendListenerName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      backend.Options.ListenerPodTemplateLabels,
					Annotations: backend.Options.ListenerPodTemplateAnnotations,
				},
				Spec: v1.PodSpec{
					Affinity:    backend.Options.ListenerAffinity,
					Tolerations: backend.Options.ListenerTolerations,
					Containers: []v1.Container{
						{
							Name:      BackendListenerName,
							Image:     containerImage,
							Args:      []string{"bin/3scale_backend", "start", "-e", "production", "-p", "3000", "-x", "/dev/stdout"},
							Ports:     backend.listenerPorts(),
							Env:       backend.buildBackendListenerEnv(),
							Resources: backend.Options.ListenerResourceRequirements,
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{TCPSocket: &v1.TCPSocketAction{
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: 3000}},
								},
								InitialDelaySeconds: 30,
								TimeoutSeconds:      0,
								PeriodSeconds:       10,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{HTTPGet: &v1.HTTPGetAction{
									Path: "/status",
									Port: intstr.IntOrString{
										Type:   intstr.Int,
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
					ServiceAccountName:        "amp",
					PriorityClassName:         backend.Options.PriorityClassNameListener,
					TopologySpreadConstraints: backend.Options.TopologySpreadConstraintsListener,
				},
			},
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
				{
					Name:     "http",
					Protocol: v1.ProtocolTCP,
					Port:     3000,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 3000,
					},
				},
			},
			Selector: map[string]string{reconcilers.DeploymentLabelSelector: BackendListenerName},
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

func (backend *Backend) buildBackendCommonEnv() []v1.EnvVar {
	return []v1.EnvVar{
		helper.EnvVarFromSecret("CONFIG_REDIS_PROXY", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageURLFieldName),
		helper.EnvVarFromSecret("CONFIG_REDIS_SENTINEL_HOSTS", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageSentinelHostsFieldName),
		helper.EnvVarFromSecret("CONFIG_REDIS_SENTINEL_ROLE", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageSentinelRoleFieldName),
		helper.EnvVarFromSecret("CONFIG_QUEUES_MASTER_NAME", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisQueuesURLFieldName),
		helper.EnvVarFromSecret("CONFIG_QUEUES_SENTINEL_HOSTS", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisQueuesSentinelHostsFieldName),
		helper.EnvVarFromSecret("CONFIG_QUEUES_SENTINEL_ROLE", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisQueuesSentinelRoleFieldName),
		helper.EnvVarFromConfigMap("RACK_ENV", "backend-environment", "RACK_ENV"),
		// TLS
		helper.EnvVarFromSecret("CONFIG_REDIS_CA_FILE", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisConfigCAFile),
		helper.EnvVarFromSecret("CONFIG_REDIS_CERT", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisConfigClientCertificate),
		helper.EnvVarFromSecret("CONFIG_REDIS_PRIVATE_KEY", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisConfigPrivateKey),
		helper.EnvVarFromSecret("CONFIG_REDIS_SSL", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisConfigSSL),
		helper.EnvVarFromSecret("CONFIG_QUEUES_CA_FILE", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisConfigQueuesCAFile),
		helper.EnvVarFromSecret("CONFIG_QUEUES_CERT", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisConfigQueuesClientCertificate),
		helper.EnvVarFromSecret("CONFIG_QUEUES_PRIVATE_KEY", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisConfigQueuesPrivateKey),
		helper.EnvVarFromSecret("CONFIG_QUEUES_SSL", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisConfigQueuesSSL),
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

func (backend *Backend) WorkerPodDisruptionBudget() *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   BackendWorkerName,
			Labels: backend.Options.CommonWorkerLabels,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{reconcilers.DeploymentLabelSelector: BackendWorkerName},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}

func (backend *Backend) CronPodDisruptionBudget() *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "backend-cron",
			Labels: backend.Options.CommonCronLabels,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{reconcilers.DeploymentLabelSelector: "backend-cron"},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}

func (backend *Backend) ListenerPodDisruptionBudget() *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   BackendListenerName,
			Labels: backend.Options.CommonListenerLabels,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{reconcilers.DeploymentLabelSelector: BackendListenerName},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}

func (backend *Backend) listenerPorts() []v1.ContainerPort {
	ports := []v1.ContainerPort{
		{HostPort: 0, ContainerPort: 3000, Protocol: v1.ProtocolTCP},
	}

	if backend.Options.ListenerMetrics {
		ports = append(ports, v1.ContainerPort{Name: "metrics", ContainerPort: BackendListenerMetricsPort, Protocol: v1.ProtocolTCP})
	}

	return ports
}

func (backend *Backend) workerPorts() []v1.ContainerPort {
	var ports []v1.ContainerPort

	if backend.Options.WorkerMetrics {
		ports = append(ports, v1.ContainerPort{Name: "metrics", ContainerPort: BackendWorkerMetricsPort, Protocol: v1.ProtocolTCP})
	}

	return ports
}
