package component

import (
	"github.com/3scale/3scale-operator/pkg/common"

	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Memcached struct {
	Options *MemcachedOptions
}

func NewMemcached(options *MemcachedOptions) *Memcached {
	return &Memcached{Options: options}
}

func (m *Memcached) Objects() []common.KubernetesObject {
	deploymentConfig := m.DeploymentConfig()

	objects := []common.KubernetesObject{
		deploymentConfig,
	}
	return objects
}

func (m *Memcached) DeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-memcache",
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "memcache", "app": m.Options.AppLabel},
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
					Type: appsv1.DeploymentTriggerOnConfigChange},
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							"memcache",
						},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "system-memcached:latest",
						},
					},
				},
			},
			Replicas: 1,
			Selector: map[string]string{"deploymentConfig": "system-memcache"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "memcache", "app": m.Options.AppLabel, "deploymentConfig": "system-memcache"},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "amp", //TODO make this configurable via flag
					Containers: []v1.Container{
						v1.Container{
							Name:    "memcache",
							Image:   "system-memcached:latest",
							Command: []string{"memcached", "-m", "64"},
							Ports: []v1.ContainerPort{
								v1.ContainerPort{HostPort: 0,
									ContainerPort: 11211,
									Protocol:      v1.ProtocolTCP},
							},
							Resources: m.Options.ResourceRequirements,
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{TCPSocket: &v1.TCPSocketAction{
									Port: intstr.IntOrString{
										Type:   intstr.Type(intstr.Int),
										IntVal: 11211}},
								},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      0,
								PeriodSeconds:       10,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{TCPSocket: &v1.TCPSocketAction{
									Port: intstr.IntOrString{
										Type:   intstr.Type(intstr.Int),
										IntVal: 11211}},
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
				}},
		},
	}
}
