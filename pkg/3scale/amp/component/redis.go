package component

import (
	"github.com/3scale/3scale-operator/pkg/common"

	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Redis struct {
	Options *RedisOptions
}

func NewRedis(options *RedisOptions) *Redis {
	return &Redis{Options: options}
}

func (redis *Redis) Objects() []common.KubernetesObject {
	backendRedisObjects := redis.buildBackendRedisObjects()
	systemRedisObjects := redis.buildSystemRedisObjects()

	objects := backendRedisObjects
	objects = append(objects, systemRedisObjects...)
	return objects
}

func (redis *Redis) buildBackendRedisObjects() []common.KubernetesObject {
	dc := redis.BackendDeploymentConfig()
	bs := redis.BackendService()
	cm := redis.BackendConfigMap()
	bpvc := redis.BackendPVC()
	bis := redis.BackendImageStream()
	objects := []common.KubernetesObject{
		dc,
		bs,
		cm,
		bpvc,
		bis,
	}
	return objects
}

func (redis *Redis) BackendDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta:   redis.buildDeploymentConfigTypeMeta(),
		ObjectMeta: redis.buildDeploymentConfigObjectMeta(),
		Spec:       redis.buildDeploymentConfigSpec(),
	}
}

func (redis *Redis) buildDeploymentConfigTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "DeploymentConfig",
		APIVersion: "apps.openshift.io/v1",
	}
}

const ()

func (redis *Redis) buildDeploymentConfigObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:   backendRedisObjectMetaName,
		Labels: redis.buildLabelsForDeploymentConfigObjectMeta(),
	}
}

const (
	backendRedisObjectMetaName    = "backend-redis"
	backendRedisDCSelectorName    = backendRedisObjectMetaName
	backendComponentNameLabel     = "backend"
	backendComponentElementLabel  = "redis"
	backendRedisStorageVolumeName = "backend-redis-storage"
	backendRedisConfigVolumeName  = "redis-config"
	backendRedisConfigMapKey      = "redis.conf"
	backendRedisContainerName     = "backend-redis"
	backendRedisContainerCommand  = "/opt/rh/rh-redis32/root/usr/bin/redis-server"
)

func (redis *Redis) buildLabelsForDeploymentConfigObjectMeta() map[string]string {
	return map[string]string{
		"app":                          redis.Options.AppLabel,
		"threescale_component":         backendComponentNameLabel,
		"threescale_component_element": backendComponentElementLabel,
	}
}

func (redis *Redis) buildDeploymentConfigSpec() appsv1.DeploymentConfigSpec {
	return appsv1.DeploymentConfigSpec{
		Template: redis.buildPodTemplateSpec(),
		Strategy: redis.buildDeploymentStrategy(),
		Selector: redis.buildDeploymentConfigSelector(),
		Replicas: 1, //TODO make this configurable via flag
		Triggers: redis.buildDeploymentConfigTriggers(),
	}
}

func (redis *Redis) buildDeploymentStrategy() appsv1.DeploymentStrategy {
	return appsv1.DeploymentStrategy{
		Type: appsv1.DeploymentStrategyTypeRecreate,
	}
}

func (redis *Redis) getSelectorLabels() map[string]string {
	return map[string]string{
		"deploymentConfig": backendRedisDCSelectorName,
	}
}

func (redis *Redis) buildDeploymentConfigSelector() map[string]string {
	return redis.getSelectorLabels()
}

func (redis *Redis) buildDeploymentConfigTriggers() appsv1.DeploymentTriggerPolicies {
	return appsv1.DeploymentTriggerPolicies{
		appsv1.DeploymentTriggerPolicy{
			Type: appsv1.DeploymentTriggerOnConfigChange,
		},
		appsv1.DeploymentTriggerPolicy{
			Type: appsv1.DeploymentTriggerOnImageChange,
			ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
				Automatic: true,
				ContainerNames: []string{
					"backend-redis",
				},
				From: v1.ObjectReference{
					Kind: "ImageStreamTag",
					Name: "backend-redis:latest",
				},
			},
		},
	}
}

func (redis *Redis) buildPodTemplateSpec() *v1.PodTemplateSpec {
	return &v1.PodTemplateSpec{
		Spec: v1.PodSpec{
			ServiceAccountName: "amp", //TODO make this configurable via flag
			Volumes:            redis.buildPodVolumes(),
			Containers:         redis.buildPodContainers(),
		},
		ObjectMeta: redis.buildPodObjectMeta(),
	}
}

func (redis *Redis) buildPodObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Labels: map[string]string{
			"deploymentConfig":             backendRedisDCSelectorName,
			"app":                          redis.Options.AppLabel,
			"threescale_component":         backendComponentNameLabel,
			"threescale_component_element": backendComponentElementLabel,
		},
	}
}

func (redis *Redis) buildPodVolumes() []v1.Volume {
	return []v1.Volume{
		v1.Volume{
			Name: backendRedisStorageVolumeName,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: backendRedisStorageVolumeName,
				},
			},
		},
		v1.Volume{
			Name: backendRedisConfigVolumeName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{ //The name of the ConfigMap
						Name: backendRedisConfigVolumeName,
					},
					Items: []v1.KeyToPath{
						v1.KeyToPath{
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
		v1.Container{
			Image:           "backend-redis:latest",
			ImagePullPolicy: v1.PullIfNotPresent,
			Name:            backendRedisContainerName,
			Command:         redis.buildPodContainerCommand(),
			Args:            redis.buildPodContainerCommandArgs(),
			Resources:       redis.buildPodContainerResourceLimits(),
			ReadinessProbe:  redis.buildPodContainerReadinessProbe(),
			LivenessProbe:   redis.buildPodContainerLivenessProbe(),
			VolumeMounts:    redis.buildPodContainerVolumeMounts(),
		},
	}
}

func (redis *Redis) buildPodContainerCommand() []string {
	return []string{
		backendRedisContainerCommand,
	}
}

func (redis *Redis) buildPodContainerCommandArgs() []string {
	return []string{
		"/etc/redis.d/redis.conf",
		"--daemonize",
		"no",
	}
}

func (redis *Redis) buildPodContainerResourceLimits() v1.ResourceRequirements {
	return *redis.Options.BackendRedisContainerResourceRequirements
}

func (redis *Redis) buildPodContainerReadinessProbe() *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
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
		Handler: v1.Handler{
			TCPSocket: &v1.TCPSocketAction{
				Port: intstr.FromInt(6379),
			},
		},
	}
}

func (redis *Redis) buildPodContainerVolumeMounts() []v1.VolumeMount {
	return []v1.VolumeMount{
		v1.VolumeMount{
			Name:      backendRedisStorageVolumeName,
			MountPath: "/var/lib/redis/data",
		},
		v1.VolumeMount{
			Name:      backendRedisConfigVolumeName,
			MountPath: "/etc/redis.d/",
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
		Labels: redis.buildLabelsForServiceObjectMeta(),
	}
}

func (redis *Redis) buildLabelsForServiceObjectMeta() map[string]string {
	return map[string]string{
		"app":                          redis.Options.AppLabel,
		"threescale_component":         "backend",
		"threescale_component_element": "redis",
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
		v1.ServicePort{
			Port:       6379,
			TargetPort: intstr.FromInt(6379),
			Protocol:   v1.ProtocolTCP,
		},
	}
}

func (redis *Redis) buildServiceSelector() map[string]string {
	return map[string]string{
		"deploymentConfig": backendRedisDCSelectorName,
	}
}

func (redis *Redis) BackendConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: redis.buildConfigMapObjectMeta(),
		TypeMeta:   redis.buildConfigMapTypeMeta(),
		Data:       redis.buildConfigMapData(),
	}
}

func (redis *Redis) buildConfigMapObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:   backendRedisConfigVolumeName,
		Labels: redis.buildLabelsForConfigMapObjectMeta(),
	}
}

func (redis *Redis) buildLabelsForConfigMapObjectMeta() map[string]string {
	return map[string]string{
		"app":                          redis.Options.AppLabel,
		"threescale_component":         "system", // TODO should also be redis???
		"threescale_component_element": "redis",
	}
}

func (redis *Redis) buildConfigMapTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "ConfigMap",
		APIVersion: "v1",
	}
}

func (redis *Redis) buildConfigMapData() map[string]string {
	return map[string]string{
		"redis.conf": redis.getRedisConfData(),
	}
}

func (redis *Redis) getRedisConfData() string { // TODO read this from a real file
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
`
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
		Labels: redis.buildLabelsForServiceObjectMeta(),
	}
}

func (redis *Redis) buildLabelsForPVCObjectMeta() map[string]string {
	return map[string]string{
		"app":                          redis.Options.AppLabel,
		"threescale_component":         "backend",
		"threescale_component_element": "redis",
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
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceStorage: resource.MustParse("1Gi"),
			},
		},
	}
}

func (redis *Redis) BackendImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "backend-redis",
			Labels: map[string]string{
				"app":                  redis.Options.AppLabel,
				"threescale_component": "backend",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "Backend Redis",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "Backend Redis (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: redis.Options.AmpRelease,
					},
				},
				imagev1.TagReference{
					Name: redis.Options.AmpRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "Backend " + redis.Options.AmpRelease + " Redis",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: redis.Options.BackendImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: *redis.Options.InsecureImportPolicy,
					},
				},
			},
		},
	}
}

////// Begin System Redis
func (redis *Redis) buildSystemRedisObjects() []common.KubernetesObject {
	systemRedisDC := redis.SystemDeploymentConfig()
	systemRedisPVC := redis.SystemPVC()
	systemRedisService := redis.SystemService()
	systemRedisImageStream := redis.SystemImageStream()

	objects := []common.KubernetesObject{
		systemRedisDC,
		systemRedisPVC,
		systemRedisService,
		systemRedisImageStream,
	}

	return objects
}

func (redis *Redis) SystemDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-redis",
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "redis", "app": redis.Options.AppLabel},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRecreate,
			},
			MinReadySeconds: 0,
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnConfigChange},
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							"system-redis",
						},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "system-redis:latest",
						},
					},
				},
			},
			Replicas: 1,
			Selector: map[string]string{"deploymentConfig": "system-redis"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "redis", "app": redis.Options.AppLabel, "deploymentConfig": "system-redis"},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "amp", //TODO make this configurable via flag
					Volumes: []v1.Volume{
						v1.Volume{
							Name: "system-redis-storage",
							VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
								ClaimName: "system-redis-storage",
								ReadOnly:  false}},
						}, v1.Volume{
							Name: "redis-config",
							VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "redis-config",
								},
								Items: []v1.KeyToPath{
									v1.KeyToPath{
										Key:  "redis.conf",
										Path: "redis.conf"}}}}},
					},
					Containers: []v1.Container{
						v1.Container{
							Name:      "system-redis",
							Image:     "system-redis:latest",
							Command:   []string{"/opt/rh/rh-redis32/root/usr/bin/redis-server"},
							Args:      []string{"/etc/redis.d/redis.conf", "--daemonize", "no"},
							Resources: *redis.Options.SystemRedisContainerResourceRequirements,
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "system-redis-storage",
									ReadOnly:  false,
									MountPath: "/var/lib/redis/data",
								}, v1.VolumeMount{
									Name:      "redis-config",
									ReadOnly:  false,
									MountPath: "/etc/redis.d/"},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{TCPSocket: &v1.TCPSocketAction{
									Port: intstr.IntOrString{
										Type:   intstr.Type(intstr.Int),
										IntVal: 6379}},
								},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      0,
								PeriodSeconds:       5,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{"container-entrypoint", "bash", "-c", "redis-cli set liveness-probe \"`date`\" | grep OK"}},
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
				}},
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
			Name: "system-redis",
			Labels: map[string]string{
				"app":                          redis.Options.AppLabel,
				"threescale_component":         "system",
				"threescale_component_element": "redis",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "redis",
					Protocol:   v1.ProtocolTCP,
					Port:       6379,
					TargetPort: intstr.FromInt(6379),
				},
			},
			Selector: map[string]string{"deploymentConfig": "system-redis"},
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
			Name: "system-redis-storage",
			Labels: map[string]string{
				"threescale_component":         "system",
				"threescale_component_element": "redis",
				"app":                          redis.Options.AppLabel,
			},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.PersistentVolumeAccessMode("ReadWriteOnce"),
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{"storage": resource.MustParse("1Gi")},
			},
		},
	}
}

func (redis *Redis) SystemImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-redis",
			Labels: map[string]string{
				"app":                  redis.Options.AppLabel,
				"threescale_component": "system",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "System Redis",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "System Redis (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: redis.Options.AmpRelease,
					},
				},
				imagev1.TagReference{
					Name: redis.Options.AmpRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "System " + redis.Options.AmpRelease + " Redis",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: redis.Options.SystemImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: *redis.Options.InsecureImportPolicy,
					},
				},
			},
		},
	}
}

////// End System Redis
