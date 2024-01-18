package component

import (
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	k8sappsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	SystemMemcachedDeploymentName = "system-memcache"
)

type Memcached struct {
	Options *MemcachedOptions
}

func NewMemcached(options *MemcachedOptions) *Memcached {
	return &Memcached{Options: options}
}

func (m *Memcached) Deployment(containerImage string) *k8sappsv1.Deployment {
	var memcachedReplicas int32 = 1

	return &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: reconcilers.DeploymentAPIVersion, Kind: reconcilers.DeploymentKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemMemcachedDeploymentName,
			Labels: m.Options.DeploymentLabels,
		},
		Spec: k8sappsv1.DeploymentSpec{
			Strategy: k8sappsv1.DeploymentStrategy{
				Type: k8sappsv1.RecreateDeploymentStrategyType,
			},
			MinReadySeconds: 0,
			Replicas:        &memcachedReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					reconcilers.DeploymentLabelSelector: SystemMemcachedDeploymentName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      m.Options.PodTemplateLabels,
					Annotations: m.Options.PodTemplateAnnotations,
				},
				Spec: v1.PodSpec{
					Affinity:           m.Options.Affinity,
					Tolerations:        m.Options.Tolerations,
					ServiceAccountName: "amp", //TODO make this configurable via flag
					Containers: []v1.Container{
						{
							Name:    "memcache",
							Image:   containerImage,
							Command: []string{"memcached", "-m", "64"},
							Ports: []v1.ContainerPort{
								{HostPort: 0,
									ContainerPort: 11211,
									Protocol:      v1.ProtocolTCP},
							},
							Resources: m.Options.ResourceRequirements,
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 11211,
										},
									},
								},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      0,
								PeriodSeconds:       10,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 11211,
										},
									},
								},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      5,
								PeriodSeconds:       30,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							ImagePullPolicy: v1.PullIfNotPresent,
						},
					},
					PriorityClassName:         m.Options.PriorityClassName,
					TopologySpreadConstraints: m.Options.TopologySpreadConstraints,
				},
			},
		},
	}
}
