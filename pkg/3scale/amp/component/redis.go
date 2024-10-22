package component

import (
	"path"

	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	k8sappsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	BackendRedisDeploymentName = "backend-redis"
	SystemRedisDeploymentName  = "system-redis"
)

const (
	redisConfigVolumeName              = "redis-config"
	backendRedisObjectMetaName         = "backend-redis"
	backendRedisDeploymentSelectorName = backendRedisObjectMetaName
	backendRedisStorageVolumeName      = "backend-redis-storage"
	backendRedisConfigMapKey           = "redis.conf"
	backendRedisContainerName          = "backend-redis"
	backendRedisConfigPath             = "/etc/redis.d/"
)

type Redis struct {
	Options *RedisOptions
}

func NewRedis(options *RedisOptions) *Redis {
	return &Redis{Options: options}
}

func (redis *Redis) BackendDeployment() *k8sappsv1.Deployment {
	return &k8sappsv1.Deployment{
		TypeMeta:   redis.buildDeploymentTypeMeta(),
		ObjectMeta: redis.buildDeploymentObjectMeta(),
		Spec:       redis.buildDeploymentSpec(),
	}
}

func (redis *Redis) buildDeploymentTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       reconcilers.DeploymentKind,
		APIVersion: reconcilers.DeploymentAPIVersion,
	}
}

func (redis *Redis) buildDeploymentObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:   backendRedisObjectMetaName,
		Labels: redis.Options.BackendRedisLabels,
	}
}

func (redis *Redis) buildDeploymentSpec() k8sappsv1.DeploymentSpec {
	var redisReplicas int32 = 1

	return k8sappsv1.DeploymentSpec{
		Template: redis.buildPodTemplateSpec(),
		Strategy: redis.buildDeploymentStrategy(),
		Selector: redis.buildDeploymentSelector(),
		Replicas: &redisReplicas,
	}
}

func (redis *Redis) buildDeploymentStrategy() k8sappsv1.DeploymentStrategy {
	return k8sappsv1.DeploymentStrategy{
		Type: k8sappsv1.RecreateDeploymentStrategyType,
	}
}

func (redis *Redis) getSelectorLabels() map[string]string {
	return map[string]string{
		reconcilers.DeploymentLabelSelector: backendRedisDeploymentSelectorName,
	}
}

func (redis *Redis) buildDeploymentSelector() *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: redis.getSelectorLabels(),
	}
}

func (redis *Redis) buildPodTemplateSpec() v1.PodTemplateSpec {
	return v1.PodTemplateSpec{
		Spec: v1.PodSpec{
			Affinity:                  redis.Options.BackendRedisAffinity,
			Tolerations:               redis.Options.BackendRedisTolerations,
			ServiceAccountName:        "amp", //TODO make this configurable via flag
			Volumes:                   redis.buildPodVolumes(),
			Containers:                redis.buildPodContainers(),
			PriorityClassName:         redis.Options.BackendRedisPriorityClassName,
			TopologySpreadConstraints: redis.Options.BackendRedisTopologySpreadConstraints,
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:      redis.Options.BackendRedisPodTemplateLabels,
			Annotations: redis.Options.BackendRedisPodTemplateAnnotations,
		},
	}
}

func (redis *Redis) buildPodVolumes() []v1.Volume {
	return []v1.Volume{
		{
			Name: backendRedisStorageVolumeName,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: backendRedisStorageVolumeName,
				},
			},
		},
		{
			Name: redisConfigVolumeName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{ //The name of the ConfigMap
						Name: redisConfigVolumeName,
					},
					Items: []v1.KeyToPath{
						{
							Key:  backendRedisConfigMapKey,
							Path: backendRedisConfigMapKey,
						},
					},
				},
			},
		},
	}
}

func (redis *Redis) buildPodContainers() []v1.Container {
	return []v1.Container{
		{
			Image:           redis.Options.BackendImage,
			ImagePullPolicy: v1.PullIfNotPresent,
			Name:            backendRedisContainerName,
			Env:             redis.buildEnv(),
			Resources:       redis.buildPodContainerResourceLimits(),
			ReadinessProbe:  redis.buildPodContainerReadinessProbe(),
			LivenessProbe:   redis.buildPodContainerLivenessProbe(),
			VolumeMounts:    redis.buildPodContainerVolumeMounts(),
		},
	}
}

func (redis *Redis) buildPodContainerResourceLimits() v1.ResourceRequirements {
	return *redis.Options.BackendRedisContainerResourceRequirements
}

func (redis *Redis) buildPodContainerReadinessProbe() *v1.Probe {
	return &v1.Probe{
		ProbeHandler: v1.ProbeHandler{
			Exec: &v1.ExecAction{
				Command: []string{
					"container-entrypoint",
					"bash",
					"-c",
					"redis-cli set liveness-probe \"`date`\" | grep OK",
				},
			},
		},
		InitialDelaySeconds: 10,
		PeriodSeconds:       30,
		TimeoutSeconds:      1,
	}
}

func (redis *Redis) buildPodContainerLivenessProbe() *v1.Probe {
	return &v1.Probe{
		InitialDelaySeconds: 10,
		PeriodSeconds:       10,
		ProbeHandler: v1.ProbeHandler{
			TCPSocket: &v1.TCPSocketAction{
				Port: intstr.FromInt32(6379),
			},
		},
	}
}

func (redis *Redis) buildPodContainerVolumeMounts() []v1.VolumeMount {
	return []v1.VolumeMount{
		{
			Name: backendRedisStorageVolumeName,
			// https://github.com/sclorg/redis-container/ images have
			// redis data directory hardcoded on /var/lib/redis/data
			MountPath: "/var/lib/redis/data",
		},
		{
			Name:      redisConfigVolumeName,
			MountPath: backendRedisConfigPath,
		},
	}
}

func (redis *Redis) BackendService() *v1.Service {
	return &v1.Service{
		ObjectMeta: redis.buildServiceObjectMeta(),
		TypeMeta:   redis.buildServiceTypeMeta(),
		Spec:       redis.buildServiceSpec(),
	}
}

func (redis *Redis) buildServiceObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:   "backend-redis",
		Labels: redis.Options.BackendRedisLabels,
	}
}

func (redis *Redis) buildServiceTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: "v1",
	}
}

func (redis *Redis) buildServiceSpec() v1.ServiceSpec {
	return v1.ServiceSpec{
		Ports:    redis.buildServicePorts(),
		Selector: redis.buildServiceSelector(),
	}
}

func (redis *Redis) buildServicePorts() []v1.ServicePort {
	return []v1.ServicePort{
		{
			Port:       6379,
			TargetPort: intstr.FromInt32(6379),
			Protocol:   v1.ProtocolTCP,
		},
	}
}

func (redis *Redis) buildServiceSelector() map[string]string {
	return map[string]string{
		reconcilers.DeploymentLabelSelector: backendRedisDeploymentSelectorName,
	}
}

func (redis *Redis) BackendPVC() *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: redis.buildPVCObjectMeta(),
		TypeMeta:   redis.buildPVCTypeMeta(),
		Spec:       redis.buildPVCSpec(),
		// TODO be able to configure StorageClass in case one wants to be used
	}
}

func (redis *Redis) buildPVCObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:   backendRedisStorageVolumeName,
		Labels: redis.Options.BackendRedisLabels,
	}
}

func (redis *Redis) buildPVCTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "PersistentVolumeClaim",
		APIVersion: "v1",
	}
}

func (redis *Redis) buildPVCSpec() v1.PersistentVolumeClaimSpec {
	return v1.PersistentVolumeClaimSpec{
		AccessModes: []v1.PersistentVolumeAccessMode{
			v1.ReadWriteOnce, // TODO be able to configure this because we have different volume access modes for different claims
		},
		Resources: v1.VolumeResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceStorage: resource.MustParse("1Gi"),
			},
		},
		StorageClassName: redis.Options.BackendRedisPVCStorageClass,
	}
}

func (redis *Redis) BackendRedisSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   BackendSecretBackendRedisSecretName,
			Labels: redis.Options.BackendCommonLabels,
		},
		StringData: map[string]string{
			BackendSecretBackendRedisStorageURLFieldName:           redis.Options.BackendStorageURL,
			BackendSecretBackendRedisQueuesURLFieldName:            redis.Options.BackendQueuesURL,
			BackendSecretBackendRedisStorageSentinelHostsFieldName: redis.Options.BackendRedisStorageSentinelHosts,
			BackendSecretBackendRedisStorageSentinelRoleFieldName:  redis.Options.BackendRedisStorageSentinelRole,
			BackendSecretBackendRedisQueuesSentinelHostsFieldName:  redis.Options.BackendRedisQueuesSentinelHosts,
			BackendSecretBackendRedisQueuesSentinelRoleFieldName:   redis.Options.BackendRedisQueuesSentinelRole,

			// TLS
			BackendSecretBackendRedisConfigCAFile:                  redis.Options.BackendConfigCAFile,
			BackendSecretBackendRedisConfigClientCertificate:       redis.Options.BackendConfigClientCertificate,
			BackendSecretBackendRedisConfigPrivateKey:              redis.Options.BackendConfigPrivateKey,
			BackendSecretBackendRedisConfigSSL:                     redis.Options.BackendConfigSSL,
			BackendSecretBackendRedisConfigQueuesCAFile:            redis.Options.BackendConfigQueuesCAFile,
			BackendSecretBackendRedisConfigQueuesClientCertificate: redis.Options.BackendConfigQueuesClientCertificate,
			BackendSecretBackendRedisConfigQueuesPrivateKey:        redis.Options.BackendConfigQueuesPrivateKey,
			BackendSecretBackendRedisConfigQueuesSSL:               redis.Options.BackendConfigQueuesSSL,
		},
		Type: v1.SecretTypeOpaque,
	}
}

////// Begin System Redis

func (redis *Redis) SystemDeployment() *k8sappsv1.Deployment {
	var redisReplicas int32 = 1

	return &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: reconcilers.DeploymentAPIVersion, Kind: reconcilers.DeploymentKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemRedisDeploymentName,
			Labels: redis.Options.SystemRedisLabels,
		},
		Spec: k8sappsv1.DeploymentSpec{
			Strategy: k8sappsv1.DeploymentStrategy{
				Type: k8sappsv1.RecreateDeploymentStrategyType,
			},
			MinReadySeconds: 0,
			Replicas:        &redisReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					reconcilers.DeploymentLabelSelector: SystemRedisDeploymentName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      redis.Options.SystemRedisPodTemplateLabels,
					Annotations: redis.Options.SystemRedisPodTemplateAnnotations,
				},
				Spec: v1.PodSpec{
					Affinity:           redis.Options.SystemRedisAffinity,
					Tolerations:        redis.Options.SystemRedisTolerations,
					ServiceAccountName: "amp", //TODO make this configurable via flag
					Volumes: []v1.Volume{
						{
							Name: "system-redis-storage",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "system-redis-storage",
									ReadOnly:  false,
								},
							},
						}, {
							Name: "redis-config",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "redis-config",
									},
									Items: []v1.KeyToPath{
										{
											Key:  "redis.conf",
											Path: "redis.conf",
										},
									},
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:      "system-redis",
							Image:     redis.Options.SystemImage,
							Env:       redis.buildEnv(),
							Resources: *redis.Options.SystemRedisContainerResourceRequirements,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "system-redis-storage",
									ReadOnly:  false,
									MountPath: "/var/lib/redis/data",
								},
								{
									Name:      "redis-config",
									ReadOnly:  false,
									MountPath: "/etc/redis.d/",
								},
							},
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 6379,
										},
									},
								},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      0,
								PeriodSeconds:       5,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									Exec: &v1.ExecAction{
										Command: []string{"container-entrypoint", "bash", "-c", "redis-cli set liveness-probe \"`date`\" | grep OK"},
									},
								},
								InitialDelaySeconds: 30,
								TimeoutSeconds:      5,
								PeriodSeconds:       10,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							TerminationMessagePath: "/dev/termination-log",
							ImagePullPolicy:        v1.PullIfNotPresent,
						},
					},
					PriorityClassName:         redis.Options.SystemRedisPriorityClassName,
					TopologySpreadConstraints: redis.Options.SystemRedisTopologySpreadConstraints,
				},
			},
		},
	}
}

func (redis *Redis) SystemService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-redis",
			Labels: redis.Options.SystemRedisLabels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "redis",
					Protocol:   v1.ProtocolTCP,
					Port:       6379,
					TargetPort: intstr.FromInt32(6379),
				},
			},
			Selector: map[string]string{reconcilers.DeploymentLabelSelector: "system-redis"},
		},
	}
}

func (redis *Redis) SystemPVC() *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-redis-storage",
			Labels: redis.Options.SystemRedisLabels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{"storage": resource.MustParse("1Gi")},
			},
			StorageClassName: redis.Options.SystemRedisPVCStorageClass,
		},
	}
}

func (redis *Redis) SystemRedisSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSecretSystemRedisSecretName,
			Labels: redis.Options.SystemCommonLabels,
		},
		StringData: map[string]string{
			SystemSecretSystemRedisURLFieldName:  redis.Options.SystemRedisURL,
			SystemSecretSystemRedisSentinelHosts: redis.Options.SystemRedisSentinelsHosts,
			SystemSecretSystemRedisSentinelRole:  redis.Options.SystemRedisSentinelsRole,

			// TLS
			SystemSecretSystemRedisCAFile:            redis.Options.SystemRedisCAFile,
			SystemSecretSystemRedisClientCertificate: redis.Options.SystemRedisClientCertificate,
			SystemSecretSystemRedisPrivateKey:        redis.Options.SystemRedisPrivateKey,
			SystemSecretSystemRedisSSL:               redis.Options.SystemRedisSSL,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func (redis *Redis) buildEnv() []v1.EnvVar {
	// https://github.com/sclorg/redis-container/ images have
	// redis data directory hardcoded on /var/lib/redis/data
	return []v1.EnvVar{
		helper.EnvVarFromValue("REDIS_CONF", path.Join(backendRedisConfigPath, backendRedisConfigMapKey)),
	}
}

////// End System Redis

// //// Redis Config Map Begin
type RedisConfigMap struct {
	Options *RedisConfigMapOptions
}

func NewRedisConfigMap(options *RedisConfigMapOptions) *RedisConfigMap {
	return &RedisConfigMap{Options: options}
}

func (r *RedisConfigMap) ConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
		ObjectMeta: r.buildConfigMapObjectMeta(),
		Data:       r.buildConfigMapData(),
	}
}

func (r *RedisConfigMap) buildConfigMapObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      redisConfigVolumeName,
		Labels:    r.Options.Labels,
		Namespace: r.Options.Namespace,
	}
}

func (r *RedisConfigMap) buildConfigMapData() map[string]string {
	return map[string]string{
		"redis.conf": r.getRedisConfData(),
	}
}

func (r *RedisConfigMap) getRedisConfData() string { // TODO read this from a real file
	return `protected-mode no

port 6379

timeout 0
tcp-keepalive 300

daemonize no
supervised no

loglevel notice

databases 16

save 900 1
save 300 10
save 60 10000

stop-writes-on-bgsave-error yes

rdbcompression yes
rdbchecksum yes

dbfilename dump.rdb

slave-serve-stale-data yes
slave-read-only yes

repl-diskless-sync no
repl-disable-tcp-nodelay no

appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
no-appendfsync-on-rewrite no
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
aof-load-truncated yes

lua-time-limit 5000

activerehashing no

aof-rewrite-incremental-fsync yes
dir /var/lib/redis/data

rename-command REPLICAOF ""
rename-command SLAVEOF ""
`
}

////// Redis Config Map End

func BackendCommonLabels(appLabel string) map[string]string {
	return map[string]string{
		"app":                  appLabel,
		"threescale_component": "backend",
	}
}
func BackendRedisLabels(appLabel string) map[string]string {
	labels := BackendCommonLabels(appLabel)
	labels["threescale_component_element"] = "redis"
	return labels
}
func SystemCommonLabels(appLabel string) map[string]string {
	return map[string]string{
		"app":                  appLabel,
		"threescale_component": "system",
	}
}
func SystemRedisLabels(appLabel string) map[string]string {
	labels := SystemCommonLabels(appLabel)
	labels["threescale_component_element"] = "redis"
	return labels
}
