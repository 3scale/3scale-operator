package component

import (
	"fmt"

	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	SystemSphinxDeploymentName = "system-sphinx"
	SystemSphinxPVCName        = "system-sphinx"
	SystemSphinxServiceName    = "system-sphinx"
	SystemSphinxDBVolumeName   = "system-sphinx-databasex"
)

type SystemSphinx struct {
	Options *SystemSphinxOptions
}

func NewSystemSphinx(options *SystemSphinxOptions) *SystemSphinx {
	return &SystemSphinx{Options: options}
}

func (s *SystemSphinx) Service() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSphinxServiceName,
			Labels: s.Options.Labels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "sphinx",
					Protocol:   v1.ProtocolTCP,
					Port:       9306,
					TargetPort: intstr.FromInt(9306),
				},
			},
			Selector: map[string]string{"deploymentConfig": "system-sphinx"},
		},
	}
}

func (s *SystemSphinx) DeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSphinxDeploymentName,
			Labels: s.Options.Labels,
		},
		Spec: appsv1.DeploymentConfigSpec{
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnConfigChange,
				},
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							"system-sphinx",
						},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: fmt.Sprintf("system-searchd:%s", s.Options.ImageTag),
						},
					},
				},
			},
			Replicas: 1,
			Selector: map[string]string{"deploymentConfig": SystemSphinxDeploymentName},
			Strategy: appsv1.DeploymentStrategy{
				RollingParams: &appsv1.RollingDeploymentStrategyParams{
					IntervalSeconds: &[]int64{1}[0],
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.Type(intstr.String),
						StrVal: "25%",
					},
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Type(intstr.String),
						StrVal: "25%",
					},
					TimeoutSeconds:      &[]int64{1200}[0],
					UpdatePeriodSeconds: &[]int64{1}[0],
				},
				Type: appsv1.DeploymentStrategyTypeRolling,
			},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: s.Options.PodTemplateLabels,
				},
				Spec: v1.PodSpec{
					Affinity:           s.Options.Affinity,
					Tolerations:        s.Options.Tolerations,
					ServiceAccountName: "amp",
					Volumes: []v1.Volume{
						v1.Volume{
							Name: SystemSphinxDBVolumeName,
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: SystemSphinxPVCName,
									ReadOnly:  false,
								},
							},
						},
					},
					Containers: []v1.Container{
						v1.Container{
							Name:            "system-sphinx",
							Image:           "system-searchd:latest",
							ImagePullPolicy: v1.PullIfNotPresent,
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									ReadOnly:  false,
									Name:      SystemSphinxDBVolumeName,
									MountPath: "/var/lib/sphinx",
								},
							},
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.FromInt(9306),
									},
								},
								InitialDelaySeconds: 60,
								PeriodSeconds:       10,
							},
							Resources: s.Options.ContainerResourceRequirements,
						},
					},
				},
			},
		},
	}
}

func (s *SystemSphinx) PVC() *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSphinxPVCName,
			Labels: s.Options.Labels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			StorageClassName: s.Options.PVCOptions.StorageClass,
			VolumeName:       s.Options.PVCOptions.VolumeName,
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.PersistentVolumeAccessMode("ReadWriteOnce"),
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: s.Options.PVCOptions.StorageRequests,
				},
			},
		},
	}
}
