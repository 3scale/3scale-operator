package component

import (
	"fmt"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	k8sappsv1 "k8s.io/api/apps/v1"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	SystemSearchdDeploymentName = "system-searchd"
	SystemSearchdPVCName        = "system-searchd"
	SystemSearchdServiceName    = "system-searchd"
	SystemSearchdDBVolumeName   = "system-searchd-database"
)

type SystemSearchd struct {
	Options *SystemSearchdOptions
}

func NewSystemSearchd(options *SystemSearchdOptions) *SystemSearchd {
	return &SystemSearchd{Options: options}
}

func (s *SystemSearchd) Service() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSearchdServiceName,
			Labels: s.Options.Labels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "searchd",
					Protocol:   v1.ProtocolTCP,
					Port:       9306,
					TargetPort: intstr.FromInt(9306),
				},
			},
			Selector: map[string]string{reconcilers.DeploymentLabelSelector: "system-searchd"},
		},
	}
}

func (s *SystemSearchd) Deployment() *k8sappsv1.Deployment {
	var searchdReplicas int32 = 1

	return &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: reconcilers.DeploymentAPIVersion, Kind: reconcilers.DeploymentKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:        SystemSearchdDeploymentName,
			Labels:      s.Options.Labels,
			Annotations: s.DeploymentAnnotations(),
		},
		Spec: k8sappsv1.DeploymentSpec{
			Replicas: &searchdReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					reconcilers.DeploymentLabelSelector: SystemSearchdDeploymentName,
				},
			},
			Strategy: k8sappsv1.DeploymentStrategy{
				Type: k8sappsv1.RecreateDeploymentStrategyType,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      s.Options.PodTemplateLabels,
					Annotations: s.Options.PodTemplateAnnotations,
				},
				Spec: v1.PodSpec{
					Affinity:           s.Options.Affinity,
					Tolerations:        s.Options.Tolerations,
					ServiceAccountName: "amp",
					Volumes: []v1.Volume{
						{
							Name: SystemSearchdDBVolumeName,
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: SystemSearchdPVCName,
									ReadOnly:  false,
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:            SystemSearchdDeploymentName,
							Image:           "system-searchd:latest",
							ImagePullPolicy: v1.PullIfNotPresent,
							VolumeMounts: []v1.VolumeMount{
								{
									ReadOnly:  false,
									Name:      SystemSearchdDBVolumeName,
									MountPath: "/var/lib/searchd",
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
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.FromInt(9306),
									},
								},
								InitialDelaySeconds: 30,
								TimeoutSeconds:      10,
								PeriodSeconds:       30,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							Resources: s.Options.ContainerResourceRequirements,
						},
					},
					PriorityClassName:         s.Options.PriorityClassName,
					TopologySpreadConstraints: s.Options.TopologySpreadConstraints,
				},
			},
		},
	}
}

func (s *SystemSearchd) PVC() *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSearchdPVCName,
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

func (s *SystemSearchd) DeploymentAnnotations() map[string]string {
	imageTriggerString := reconcilers.CreateImageTriggerAnnotationString([]reconcilers.ContainerImage{
		{
			Name: SystemSearchdDeploymentName,
			Tag:  fmt.Sprintf("system-searchd:%v", s.Options.ImageTag),
		},
	})
	return map[string]string{reconcilers.DeploymentImageTriggerAnnotation: imageTriggerString}

}
