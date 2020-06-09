package component

import (
	"fmt"

	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type BackendRedis struct {
	Options *BackendRedisOptions
}

func NewBackendRedis(options *BackendRedisOptions) *BackendRedis {
	return &BackendRedis{Options: options}
}

func (backendRedis *BackendRedis) DeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta:   backendRedis.buildDeploymentConfigTypeMeta(),
		ObjectMeta: backendRedis.buildDeploymentConfigObjectMeta(),
		Spec:       backendRedis.buildDeploymentConfigSpec(),
	}
}

func (backendRedis *BackendRedis) buildDeploymentConfigTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "DeploymentConfig",
		APIVersion: "apps.openshift.io/v1",
	}
}

func (backendRedis *BackendRedis) buildDeploymentConfigObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:   backendRedisObjectMetaName,
		Labels: backendRedis.Options.RedisLabels,
	}
}

const (
	backendRedisObjectMetaName    = "backend-redis"
	backendRedisDCSelectorName    = backendRedisObjectMetaName
	backendRedisStorageVolumeName = "backend-redis-storage"
	backendRedisConfigVolumeName  = "redis-config"
	backendRedisConfigMapKey      = "redis.conf"
	backendRedisContainerName     = "backend-redis"
	backendRedisContainerCommand  = "/opt/rh/rh-redis32/root/usr/bin/redis-server"
)

func (backendRedis *BackendRedis) buildDeploymentConfigSpec() appsv1.DeploymentConfigSpec {
	return appsv1.DeploymentConfigSpec{
		Template: backendRedis.buildPodTemplateSpec(),
		Strategy: backendRedis.buildDeploymentStrategy(),
		Selector: backendRedis.buildDeploymentConfigSelector(),
		Replicas: 1,
		Triggers: backendRedis.buildDeploymentConfigTriggers(),
	}
}

func (backendRedis *BackendRedis) buildDeploymentStrategy() appsv1.DeploymentStrategy {
	return appsv1.DeploymentStrategy{
		Type: appsv1.DeploymentStrategyTypeRecreate,
	}
}

func (backendRedis *BackendRedis) getSelectorLabels() map[string]string {
	return map[string]string{
		"deploymentConfig": backendRedisDCSelectorName,
	}
}

func (backendRedis *BackendRedis) buildDeploymentConfigSelector() map[string]string {
	return backendRedis.getSelectorLabels()
}

func (backendRedis *BackendRedis) buildDeploymentConfigTriggers() appsv1.DeploymentTriggerPolicies {
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
					Name: fmt.Sprintf("backend-redis:%s", backendRedis.Options.ImageTag),
				},
			},
		},
	}
}

func (backendRedis *BackendRedis) buildPodTemplateSpec() *v1.PodTemplateSpec {
	return &v1.PodTemplateSpec{
		Spec: v1.PodSpec{
			ServiceAccountName: "amp", //TODO make this configurable via flag
			Volumes:            backendRedis.buildPodVolumes(),
			Containers:         backendRedis.buildPodContainers(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: backendRedis.Options.PodTemplateLabels,
		},
	}
}

func (backendRedis *BackendRedis) buildPodVolumes() []v1.Volume {
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

func (backendRedis *BackendRedis) buildPodContainers() []v1.Container {
	return []v1.Container{
		v1.Container{
			Image:           "backend-redis:latest",
			ImagePullPolicy: v1.PullIfNotPresent,
			Name:            backendRedisContainerName,
			Command:         backendRedis.buildPodContainerCommand(),
			Args:            backendRedis.buildPodContainerCommandArgs(),
			Resources:       backendRedis.buildPodContainerResourceLimits(),
			ReadinessProbe:  backendRedis.buildPodContainerReadinessProbe(),
			LivenessProbe:   backendRedis.buildPodContainerLivenessProbe(),
			VolumeMounts:    backendRedis.buildPodContainerVolumeMounts(),
		},
	}
}

func (backendRedis *BackendRedis) buildPodContainerCommand() []string {
	return []string{
		backendRedisContainerCommand,
	}
}

func (backendRedis *BackendRedis) buildPodContainerCommandArgs() []string {
	return []string{
		"/etc/redis.d/redis.conf",
		"--daemonize",
		"no",
	}
}

func (backendRedis *BackendRedis) buildPodContainerResourceLimits() v1.ResourceRequirements {
	return *backendRedis.Options.ContainerResourceRequirements
}

func (backendRedis *BackendRedis) buildPodContainerReadinessProbe() *v1.Probe {
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

func (backendRedis *BackendRedis) buildPodContainerLivenessProbe() *v1.Probe {
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

func (backendRedis *BackendRedis) buildPodContainerVolumeMounts() []v1.VolumeMount {
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

func (backendRedis *BackendRedis) Service() *v1.Service {
	return &v1.Service{
		ObjectMeta: backendRedis.buildServiceObjectMeta(),
		TypeMeta:   backendRedis.buildServiceTypeMeta(),
		Spec:       backendRedis.buildServiceSpec(),
	}
}

func (backendRedis *BackendRedis) buildServiceObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:   "backend-redis",
		Labels: backendRedis.Options.RedisLabels,
	}
}

func (backendRedis *BackendRedis) buildServiceTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: "v1",
	}
}

func (backendRedis *BackendRedis) buildServiceSpec() v1.ServiceSpec {
	return v1.ServiceSpec{
		Ports:    backendRedis.buildServicePorts(),
		Selector: backendRedis.buildServiceSelector(),
	}
}

func (backendRedis *BackendRedis) buildServicePorts() []v1.ServicePort {
	return []v1.ServicePort{
		v1.ServicePort{
			Port:       6379,
			TargetPort: intstr.FromInt(6379),
			Protocol:   v1.ProtocolTCP,
		},
	}
}

func (backendRedis *BackendRedis) buildServiceSelector() map[string]string {
	return map[string]string{
		"deploymentConfig": backendRedisDCSelectorName,
	}
}

func (backendRedis *BackendRedis) PersistentVolumeClaim() *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: backendRedis.buildPVCObjectMeta(),
		TypeMeta:   backendRedis.buildPVCTypeMeta(),
		Spec:       backendRedis.buildPVCSpec(),
		// TODO be able to configure StorageClass in case one wants to be used
	}
}

func (backendRedis *BackendRedis) buildPVCObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:   backendRedisStorageVolumeName,
		Labels: backendRedis.Options.RedisLabels,
	}
}

func (backendRedis *BackendRedis) buildPVCTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "PersistentVolumeClaim",
		APIVersion: "v1",
	}
}

func (backendRedis *BackendRedis) buildPVCSpec() v1.PersistentVolumeClaimSpec {
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

func (backendRedis *BackendRedis) ImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "backend-redis",
			Labels: backendRedis.Options.BackendCommonLabels,
			Annotations: map[string]string{
				"openshift.io/display-name": "Backend Redis",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: backendRedis.Options.AmpRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "Backend " + backendRedis.Options.AmpRelease + " Redis",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: backendRedis.Options.Image,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: *backendRedis.Options.InsecureImportPolicy,
					},
				},
			},
		},
	}
}
