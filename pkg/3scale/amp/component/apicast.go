package component

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/common"
	"k8s.io/api/policy/v1beta1"

	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Apicast struct {
	Options *ApicastOptions
}

func NewApicast(options *ApicastOptions) *Apicast {
	return &Apicast{Options: options}
}

func (apicast *Apicast) Objects() []common.KubernetesObject {
	stagingDeploymentConfig := apicast.StagingDeploymentConfig()
	productionDeploymentConfig := apicast.ProductionDeploymentConfig()
	stagingService := apicast.StagingService()
	productionService := apicast.ProductionService()
	environmentConfigMap := apicast.EnvironmentConfigMap()

	objects := []common.KubernetesObject{
		stagingDeploymentConfig,
		productionDeploymentConfig,
		stagingService,
		productionService,
		environmentConfigMap,
	}
	return objects
}

func (apicast *Apicast) PDBObjects() []common.KubernetesObject {
	stagingPDB := apicast.StagingPodDisruptionBudget()
	prodPDB := apicast.ProductionPodDisruptionBudget()
	return []common.KubernetesObject{
		stagingPDB,
		prodPDB,
	}
}

func (apicast *Apicast) StagingService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "apicast-staging",
			Labels: map[string]string{
				"app":                          apicast.Options.AppLabel,
				"threescale_component":         "apicast",
				"threescale_component_element": "staging",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "gateway",
					Protocol:   v1.ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
				v1.ServicePort{
					Name:       "management",
					Protocol:   v1.ProtocolTCP,
					Port:       8090,
					TargetPort: intstr.FromInt(8090),
				},
			},
			Selector: map[string]string{"deploymentConfig": "apicast-staging"},
		},
	}
}

func (apicast *Apicast) ProductionService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "apicast-production",
			Labels: map[string]string{
				"app":                          apicast.Options.AppLabel,
				"threescale_component":         "apicast",
				"threescale_component_element": "production",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "gateway",
					Protocol:   v1.ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
				v1.ServicePort{
					Name:       "management",
					Protocol:   v1.ProtocolTCP,
					Port:       8090,
					TargetPort: intstr.FromInt(8090),
				},
			},
			Selector: map[string]string{"deploymentConfig": "apicast-production"},
		},
	}
}

func (apicast *Apicast) StagingDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: "apps.openshift.io/v1", Kind: "DeploymentConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "apicast-staging",
			Labels: map[string]string{
				"app":                          apicast.Options.AppLabel,
				"threescale_component":         "apicast",
				"threescale_component_element": "staging",
			},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: apicast.Options.StagingReplicas,
			Selector: map[string]string{
				"deploymentConfig": "apicast-staging",
			},
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
					TimeoutSeconds:      &[]int64{1800}[0],
					UpdatePeriodSeconds: &[]int64{1}[0],
				},
				Type: appsv1.DeploymentStrategyTypeRolling,
			},
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnConfigChange,
				},
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							"apicast-staging",
						},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: fmt.Sprintf("amp-apicast:%s", apicast.Options.ImageTag),
						},
					},
				},
			},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"deploymentConfig":             "apicast-staging",
						"app":                          apicast.Options.AppLabel,
						"threescale_component":         "apicast",
						"threescale_component_element": "staging",
					},
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   "9421",
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "amp",
					Containers: []v1.Container{
						v1.Container{
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									ContainerPort: 8080,
									Protocol:      v1.ProtocolTCP,
								},
								v1.ContainerPort{
									ContainerPort: 8090,
									Protocol:      v1.ProtocolTCP,
								},
								v1.ContainerPort{
									ContainerPort: 9421,
									Protocol:      v1.ProtocolTCP,
									Name:          "metrics",
								},
							},
							Env:             apicast.buildApicastStagingEnv(),
							Image:           "amp-apicast:latest",
							ImagePullPolicy: v1.PullIfNotPresent,
							Name:            "apicast-staging",
							Resources:       apicast.Options.StagingResourceRequirements,
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{HTTPGet: &v1.HTTPGetAction{
									Path: "/status/live",
									Port: intstr.FromInt(8090),
								}},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      5,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{HTTPGet: &v1.HTTPGetAction{
									Path: "/status/ready",
									Port: intstr.FromInt(8090),
								}},
								InitialDelaySeconds: 15,
								TimeoutSeconds:      5,
								PeriodSeconds:       30,
							},
						},
					},
				},
			},
		},
	}
}

func (apicast *Apicast) ProductionDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: "apps.openshift.io/v1", Kind: "DeploymentConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "apicast-production",
			Labels: map[string]string{
				"app":                          apicast.Options.AppLabel,
				"threescale_component":         "apicast",
				"threescale_component_element": "production",
			},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: apicast.Options.ProductionReplicas,
			Selector: map[string]string{
				"deploymentConfig": "apicast-production",
			},
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
					TimeoutSeconds:      &[]int64{1800}[0],
					UpdatePeriodSeconds: &[]int64{1}[0],
				},
				Type: appsv1.DeploymentStrategyTypeRolling,
			},
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnConfigChange,
				},
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							"system-master-svc",
							"apicast-production",
						},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: fmt.Sprintf("amp-apicast:%s", apicast.Options.ImageTag),
						},
					},
				},
			},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"deploymentConfig":             "apicast-production",
						"app":                          apicast.Options.AppLabel,
						"threescale_component":         "apicast",
						"threescale_component_element": "production",
					},
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   "9421",
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "amp",
					InitContainers: []v1.Container{
						v1.Container{
							Name:    "system-master-svc",
							Image:   "amp-apicast:latest",
							Command: []string{"sh", "-c", "until $(curl --output /dev/null --silent --fail --head http://system-master:3000/status); do sleep $SLEEP_SECONDS; done"},
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name:  "SLEEP_SECONDS",
									Value: "1",
								},
							},
						},
					},
					Containers: []v1.Container{
						v1.Container{
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									ContainerPort: 8080,
									Protocol:      v1.ProtocolTCP,
								},
								v1.ContainerPort{
									ContainerPort: 8090,
									Protocol:      v1.ProtocolTCP,
								},
								v1.ContainerPort{
									ContainerPort: 9421,
									Protocol:      v1.ProtocolTCP,
									Name:          "metrics",
								},
							},
							Env:             apicast.buildApicastProductionEnv(),
							Image:           "amp-apicast:latest",
							ImagePullPolicy: v1.PullIfNotPresent,
							Name:            "apicast-production",
							Resources:       apicast.Options.ProductionResourceRequirements,
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{HTTPGet: &v1.HTTPGetAction{
									Path: "/status/live",
									Port: intstr.FromInt(8090),
								}},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      5,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{HTTPGet: &v1.HTTPGetAction{
									Path: "/status/ready",
									Port: intstr.FromInt(8090),
								}},
								InitialDelaySeconds: 15,
								TimeoutSeconds:      5,
								PeriodSeconds:       30,
							},
						},
					},
				},
			},
		},
	}
}

func (apicast *Apicast) buildApicastCommonEnv() []v1.EnvVar {
	return []v1.EnvVar{
		envVarFromSecret("THREESCALE_PORTAL_ENDPOINT", "system-master-apicast", "PROXY_CONFIGS_ENDPOINT"),
		envVarFromSecret("BACKEND_ENDPOINT_OVERRIDE", "backend-listener", "service_endpoint"),
		envVarFromConfigMap("APICAST_MANAGEMENT_API", "apicast-environment", "APICAST_MANAGEMENT_API"),
		envVarFromConfigMap("OPENSSL_VERIFY", "apicast-environment", "OPENSSL_VERIFY"),
		envVarFromConfigMap("APICAST_RESPONSE_CODES", "apicast-environment", "APICAST_RESPONSE_CODES"),
	}
}

func (apicast *Apicast) buildApicastStagingEnv() []v1.EnvVar {
	result := []v1.EnvVar{}
	result = append(result, apicast.buildApicastCommonEnv()...)
	result = append(result,
		envVarFromValue("APICAST_CONFIGURATION_LOADER", "lazy"),
		envVarFromValue("APICAST_CONFIGURATION_CACHE", "0"),
		envVarFromValue("THREESCALE_DEPLOYMENT_ENV", "staging"),
	)
	return result
}

func (apicast *Apicast) buildApicastProductionEnv() []v1.EnvVar {
	result := []v1.EnvVar{}
	result = append(result, apicast.buildApicastCommonEnv()...)
	result = append(result,
		envVarFromValue("APICAST_CONFIGURATION_LOADER", "boot"),
		envVarFromValue("APICAST_CONFIGURATION_CACHE", "300"),
		envVarFromValue("THREESCALE_DEPLOYMENT_ENV", "production"),
	)
	return result
}

func (apicast *Apicast) EnvironmentConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "apicast-environment",
			Labels: map[string]string{"threescale_component": "apicast", "app": apicast.Options.AppLabel},
		},
		Data: map[string]string{
			"APICAST_MANAGEMENT_API": apicast.Options.ManagementAPI,
			"OPENSSL_VERIFY":         apicast.Options.OpenSSLVerify,
			"APICAST_RESPONSE_CODES": apicast.Options.ResponseCodes,
		},
	}
}

func (apicast *Apicast) StagingPodDisruptionBudget() *v1beta1.PodDisruptionBudget {
	return &v1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "apicast-staging",
			Labels: map[string]string{
				"app":                          apicast.Options.AppLabel,
				"threescale_component":         "apicast",
				"threescale_component_element": "staging",
			},
		},
		Spec: v1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deploymentConfig": "apicast-staging"},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}

func (apicast *Apicast) ProductionPodDisruptionBudget() *v1beta1.PodDisruptionBudget {
	return &v1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "apicast-production",
			Labels: map[string]string{
				"app":                          apicast.Options.AppLabel,
				"threescale_component":         "apicast",
				"threescale_component_element": "production",
			},
		},
		Spec: v1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deploymentConfig": "apicast-production"},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}
