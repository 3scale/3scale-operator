package component

import (
	"context"
	"strconv"

	"sigs.k8s.io/controller-runtime/pkg/client"

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
	BackendSecretBackendRedisSecretName                      = "backend-redis"
	BackendSecretBackendRedisQueuesURLFieldName              = "REDIS_QUEUES_URL"
	BackendSecretBackendRedisQueuesUsernameFieldName         = "REDIS_QUEUES_USERNAME"
	BackendSecretBackendRedisQueuesPasswordFieldName         = "REDIS_QUEUES_PASSWORD"
	BackendSecretBackendRedisQueuesSentinelHostsFieldName    = "REDIS_QUEUES_SENTINEL_HOSTS"
	BackendSecretBackendRedisQueuesSentinelRoleFieldName     = "REDIS_QUEUES_SENTINEL_ROLE"
	BackendSecretBackendRedisQueuesSentinelUsernameFieldName = "REDIS_QUEUES_SENTINEL_USERNAME"
	BackendSecretBackendRedisQueuesSentinelPasswordFieldName = "REDIS_QUEUES_SENTINEL_PASSWORD"

	BackendSecretBackendRedisStorageURLFieldName              = "REDIS_STORAGE_URL"
	BackendSecretBackendRedisStorageUsernameFieldName         = "REDIS_STORAGE_USERNAME"
	BackendSecretBackendRedisStoragePasswordFieldName         = "REDIS_STORAGE_PASSWORD"
	BackendSecretBackendRedisStorageSentinelHostsFieldName    = "REDIS_STORAGE_SENTINEL_HOSTS"
	BackendSecretBackendRedisStorageSentinelRoleFieldName     = "REDIS_STORAGE_SENTINEL_ROLE"
	BackendSecretBackendRedisStorageSentinelUsernameFieldName = "REDIS_STORAGE_SENTINEL_USERNAME"
	BackendSecretBackendRedisStorageSentinelPasswordFieldName = "REDIS_STORAGE_SENTINEL_PASSWORD"
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

const (
	ConfigRedisCaFilePath     = "/tls/backend-redis/backend-redis-ca.crt"
	ConfigRedisClientCertPath = "/tls/backend-redis/backend-redis-client.crt"
	ConfigRedisPrivateKeyPath = "/tls/backend-redis/backend-redis-private.key"

	ConfigQueuesCaFilePath             = "/tls/queues/config-queues-ca.crt"
	ConfigQueuesClientCertPath         = "/tls/queues/config-queues-client.crt"
	ConfigQueuesPrivateKeyPath         = "/tls/queues/config-queues-private.key"
	BackendRedisSecretResverAnnotation = "apimanager.apps.3scale.net/backend-redis-secret-resource-version"
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

func (backend *Backend) WorkerDeployment(ctx context.Context, k8sclient client.Client, containerImage string) (*k8sappsv1.Deployment, error) {
	watchedSecretAnnotations, err := ComputeWatchedSecretAnnotations(ctx, k8sclient, BackendWorkerName, backend.Options.Namespace, backend)
	if err != nil {
		return nil, err
	}
	deploymentAnnotations := helper.MergeMapsStringString(watchedSecretAnnotations, backend.Options.WorkerPodTemplateAnnotations)

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
					Annotations: deploymentAnnotations,
				},
				Spec: v1.PodSpec{
					Affinity:    backend.Options.WorkerAffinity,
					Tolerations: backend.Options.WorkerTolerations,
					Volumes:     backend.backendVolumes(),
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
							VolumeMounts: backend.backendContainerVolumeMounts(),
							Env:          append(backend.buildBackendCommonEnv(), helper.EnvVarFromValue("SLEEP_SECONDS", "1")),
						},
					},
					Containers: []v1.Container{
						{
							Name:            BackendWorkerName,
							Image:           containerImage,
							Args:            []string{"bin/3scale_backend_worker", "run"},
							Env:             backend.buildBackendWorkerEnv(),
							Resources:       backend.Options.WorkerResourceRequirements,
							VolumeMounts:    backend.backendContainerVolumeMounts(),
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
	}, nil
}

func (backend *Backend) CronDeployment(ctx context.Context, k8sclient client.Client, containerImage string) (*k8sappsv1.Deployment, error) {
	watchedSecretAnnotations, err := ComputeWatchedSecretAnnotations(ctx, k8sclient, BackendCronName, backend.Options.Namespace, backend)
	if err != nil {
		return nil, err
	}
	deploymentAnnotations := helper.MergeMapsStringString(watchedSecretAnnotations, backend.Options.CronPodTemplateAnnotations)

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
					Annotations: deploymentAnnotations,
				},
				Spec: v1.PodSpec{
					Affinity:    backend.Options.CronAffinity,
					Tolerations: backend.Options.CronTolerations,
					Volumes:     backend.backendVolumes(),
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
							VolumeMounts: backend.backendContainerVolumeMounts(),
							Env:          append(backend.buildBackendCommonEnv(), helper.EnvVarFromValue("SLEEP_SECONDS", "1")),
						},
					},
					Containers: []v1.Container{
						{
							Name:            "backend-cron",
							Image:           containerImage,
							Args:            []string{"touch /tmp/healthy && backend-cron"},
							Env:             backend.buildBackendCronEnv(),
							VolumeMounts:    backend.backendContainerVolumeMounts(),
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
	}, nil
}

func (backend *Backend) ListenerDeployment(ctx context.Context, k8sclient client.Client, containerImage string) (*k8sappsv1.Deployment, error) {
	watchedSecretAnnotations, err := ComputeWatchedSecretAnnotations(ctx, k8sclient, BackendListenerName, backend.Options.Namespace, backend)
	if err != nil {
		return nil, err
	}
	deploymentAnnotations := helper.MergeMapsStringString(watchedSecretAnnotations, backend.Options.ListenerPodTemplateAnnotations)

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
					Annotations: deploymentAnnotations,
				},
				Spec: v1.PodSpec{
					Affinity:    backend.Options.ListenerAffinity,
					Tolerations: backend.Options.ListenerTolerations,
					Volumes:     backend.backendVolumes(),
					Containers: []v1.Container{
						{
							Name:         BackendListenerName,
							Image:        containerImage,
							Args:         []string{"bin/3scale_backend", "start", "-e", "production", "-p", "3000", "-x", "/dev/stdout"},
							Ports:        backend.listenerPorts(),
							Env:          backend.buildBackendListenerEnv(),
							Resources:    backend.Options.ListenerResourceRequirements,
							VolumeMounts: backend.backendContainerVolumeMounts(),
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 3000,
										},
									},
								},
								InitialDelaySeconds: 30,
								TimeoutSeconds:      0,
								PeriodSeconds:       10,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path: "/status",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 3000,
										},
									},
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
	}, nil
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
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyAllow,
			},
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
	result := []v1.EnvVar{}
	result = append(result,
		helper.EnvVarFromSecret("CONFIG_REDIS_PROXY", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageURLFieldName),
		helper.EnvVarFromSecret("CONFIG_REDIS_SENTINEL_HOSTS", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageSentinelHostsFieldName),
		helper.EnvVarFromSecret("CONFIG_REDIS_SENTINEL_ROLE", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageSentinelRoleFieldName),
		helper.EnvVarFromSecret("CONFIG_QUEUES_MASTER_NAME", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisQueuesURLFieldName),
		helper.EnvVarFromSecret("CONFIG_QUEUES_SENTINEL_HOSTS", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisQueuesSentinelHostsFieldName),
		helper.EnvVarFromSecret("CONFIG_QUEUES_SENTINEL_ROLE", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisQueuesSentinelRoleFieldName),
		// ACL
		helper.EnvVarFromSecretOptional("CONFIG_REDIS_USERNAME", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageUsernameFieldName),
		helper.EnvVarFromSecretOptional("CONFIG_REDIS_PASSWORD", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStoragePasswordFieldName),
		helper.EnvVarFromSecretOptional("CONFIG_REDIS_SENTINEL_USERNAME", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageSentinelUsernameFieldName),
		helper.EnvVarFromSecretOptional("CONFIG_REDIS_SENTINEL_PASSWORD", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisStorageSentinelPasswordFieldName),
		helper.EnvVarFromSecretOptional("CONFIG_QUEUES_USERNAME", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisQueuesUsernameFieldName),
		helper.EnvVarFromSecretOptional("CONFIG_QUEUES_PASSWORD", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisQueuesPasswordFieldName),
		helper.EnvVarFromSecretOptional("CONFIG_QUEUES_SENTINEL_USERNAME", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisQueuesSentinelUsernameFieldName),
		helper.EnvVarFromSecretOptional("CONFIG_QUEUES_SENTINEL_PASSWORD", BackendSecretBackendRedisSecretName, BackendSecretBackendRedisQueuesSentinelPasswordFieldName),
		helper.EnvVarFromConfigMap("RACK_ENV", "backend-environment", "RACK_ENV"),
	)
	if backend.Options.BackendRedisTLSEnabled {
		result = append(result, backend.BackendRedisTLSEnvVars()...)
	}
	if backend.Options.QueuesRedisTLSEnabled {
		result = append(result, backend.QueuesRedisTLSEnvVars()...)
	}
	return result
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

func (backend *Backend) BackendRedisTLSEnvVars() []v1.EnvVar {
	return []v1.EnvVar{
		helper.EnvVarFromValue("CONFIG_REDIS_CA_FILE", ConfigRedisCaFilePath),
		helper.EnvVarFromValue("CONFIG_REDIS_CERT", ConfigRedisClientCertPath),
		helper.EnvVarFromValue("CONFIG_REDIS_PRIVATE_KEY", ConfigRedisPrivateKeyPath),
		helper.EnvVarFromValue("CONFIG_REDIS_SSL", "1"),
	}
}

func (backend *Backend) QueuesRedisTLSEnvVars() []v1.EnvVar {
	return []v1.EnvVar{
		helper.EnvVarFromValue("CONFIG_QUEUES_CA_FILE", ConfigQueuesCaFilePath),
		helper.EnvVarFromValue("CONFIG_QUEUES_CERT", ConfigQueuesClientCertPath),
		helper.EnvVarFromValue("CONFIG_QUEUES_PRIVATE_KEY", ConfigQueuesPrivateKeyPath),
		helper.EnvVarFromValue("CONFIG_QUEUES_SSL", "1"),
	}
}

func (backend *Backend) backendVolumes() []v1.Volume {
	res := []v1.Volume{}

	if backend.Options.BackendRedisTLSEnabled {
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
	if backend.Options.QueuesRedisTLSEnabled {
		backendRedisTlsVolume := v1.Volume{
			Name: "queues-redis-tls",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: BackendSecretBackendRedisSecretName,
					Items: []v1.KeyToPath{
						{
							Key:  "REDIS_SSL_QUEUES_CA",
							Path: "config-queues-ca.crt",
						},
						{
							Key:  "REDIS_SSL_QUEUES_CERT",
							Path: "config-queues-client.crt",
						},
						{
							Key:  "REDIS_SSL_QUEUES_KEY",
							Path: "config-queues-private.key",
						},
					},
				},
			},
		}
		res = append(res, backendRedisTlsVolume)
	}
	return res
}

func (backend *Backend) backendContainerVolumeMounts() []v1.VolumeMount {
	res := []v1.VolumeMount{}
	if backend.Options.BackendRedisTLSEnabled {
		res = append(res, backend.backendRedisContainerVolumeMounts())
	}
	if backend.Options.QueuesRedisTLSEnabled {
		res = append(res, backend.queuesRedisContainerVolumeMounts())
	}
	return res
}

func (backend *Backend) backendRedisContainerVolumeMounts() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      "backend-redis-tls",
		ReadOnly:  false,
		MountPath: "/tls/backend-redis",
	}
}

func (backend *Backend) queuesRedisContainerVolumeMounts() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      "queues-redis-tls",
		ReadOnly:  false,
		MountPath: "/tls/queues",
	}
}
