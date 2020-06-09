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

type SystemRedis struct {
	Options *SystemRedisOptions
}

func NewSystemRedis(options *SystemRedisOptions) *SystemRedis {
	return &SystemRedis{Options: options}
}

func (systemRedis *SystemRedis) DeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-redis",
			Labels: systemRedis.Options.RedisLabels,
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
							Name: fmt.Sprintf("system-redis:%s", systemRedis.Options.ImageTag),
						},
					},
				},
			},
			Replicas: 1,
			Selector: map[string]string{"deploymentConfig": "system-redis"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: systemRedis.Options.PodTemplateLabels,
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
							Resources: *systemRedis.Options.ContainerResourceRequirements,
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

func (systemRedis *SystemRedis) Service() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-redis",
			Labels: systemRedis.Options.RedisLabels,
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

func (systemRedis *SystemRedis) PersistentVolumeClaim() *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-redis-storage",
			Labels: systemRedis.Options.RedisLabels,
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

func (systemRedis *SystemRedis) ImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-redis",
			Labels: systemRedis.Options.SystemCommonLabels,
			Annotations: map[string]string{
				"openshift.io/display-name": "System Redis",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: systemRedis.Options.AmpRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "System " + systemRedis.Options.AmpRelease + " Redis",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: systemRedis.Options.Image,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: *systemRedis.Options.InsecureImportPolicy,
					},
				},
			},
		},
	}
}
