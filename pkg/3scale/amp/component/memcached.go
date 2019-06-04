package component

import (
	"fmt"
	"github.com/3scale/3scale-operator/pkg/apis/common"
	"github.com/3scale/3scale-operator/pkg/helper"

	appsv1 "github.com/openshift/api/apps/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Memcached struct {
	// TemplateParameters
	// TemplateObjects
	// CLI Flags??? should be in this object???
	options []string
	Options *MemcachedOptions
}

type MemcachedOptions struct {
	memcachedNonRequiredOptions
	memcachedRequiredOptions
}

type memcachedRequiredOptions struct {
	appLabel string
}

type memcachedNonRequiredOptions struct {
}

func NewMemcached(options []string) *Memcached {
	redis := &Memcached{
		options: options,
	}
	return redis
}

type MemcachedOptionsProvider interface {
	GetMemcachedOptions() *MemcachedOptions
}
type CLIMemcachedOptionsProvider struct {
}

func (o *CLIMemcachedOptionsProvider) GetMemcachedOptions() (*MemcachedOptions, error) {
	rob := MemcachedOptionsBuilder{}
	rob.AppLabel("${APP_LABEL}")
	res, err := rob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Memcached Options - %s", err)
	}
	return res, nil
}

func (m *Memcached) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	// TODO move this outside this specific method
	optionsProvider := CLIMemcachedOptionsProvider{}
	memcachedOpts, err := optionsProvider.GetMemcachedOptions()
	_ = err
	m.Options = memcachedOpts
	m.buildParameters(template)
	m.addObjectsIntoTemplate(template)
}

func (m *Memcached) GetObjects() ([]common.KubernetesObject, error) {
	objects := m.buildObjects()
	return objects, nil
}

func (m *Memcached) addObjectsIntoTemplate(template *templatev1.Template) {
	objects := m.buildObjects()
	template.Objects = append(template.Objects, helper.WrapRawExtensions(objects)...)
}

func (m *Memcached) PostProcess(template *templatev1.Template, otherComponents []Component) {

}

func (m *Memcached) buildParameters(template *templatev1.Template) {
	parameters := []templatev1.Parameter{
		// 	- name: Memcached_USER
		// 	displayName: Memcached User
		// 	description: Username for Memcached user that will be used for accessing the database.
		// 	value: "Memcached"
		// 	required: true
		// - name: Memcached_PASSWORD
		// 	displayName: Memcached Password
		// 	description: Password for the Memcached usem.
		// 	generate: expression
		// 	from: "[a-z0-9]{8}"
		// 	required: true
		// - name: Memcached_DATABASE
		// 	displayName: Memcached Database Name
		// 	description: Name of the Memcached database accessed.
		// 	value: "system"
		// 	required: true
		// - name: Memcached_ROOT_PASSWORD
		// 	displayName: Memcached Root password.
		// 	description: Password for Root usem.
		// 	generate: expression
		// 	from: "[a-z0-9]{8}"
		// 	required: true
	}
	template.Parameters = append(template.Parameters, parameters...)
}

func (m *Memcached) buildObjects() []common.KubernetesObject {
	systemMemcachedDeploymentConfig := m.buildSystemMemcachedDeploymentConfig()

	objects := []common.KubernetesObject{
		systemMemcachedDeploymentConfig,
	}
	return objects
}

func (m *Memcached) buildSystemMemcachedDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-memcache",
			Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "memcache", "app": m.Options.appLabel},
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
					Labels: map[string]string{"threescale_component": "system", "threescale_component_element": "memcache", "app": m.Options.appLabel, "deploymentConfig": "system-memcache"},
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
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("250m"),
									v1.ResourceMemory: resource.MustParse("96Mi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("50m"),
									v1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
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
								Handler: v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{"sh", "-c", "echo version | nc $HOSTNAME 11211 | grep VERSION"}},
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
