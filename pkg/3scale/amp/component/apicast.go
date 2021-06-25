package component

import (
	"fmt"
	"path"
	"strconv"

	"github.com/3scale/3scale-operator/pkg/helper"

	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ApicastStagingName    = "apicast-staging"
	ApicastProductionName = "apicast-production"

	CustomPoliciesMountBasePath = "/opt/app-root/src/policies"
)

type Apicast struct {
	Options *ApicastOptions
}

func NewApicast(options *ApicastOptions) *Apicast {
	return &Apicast{Options: options}
}

func (apicast *Apicast) StagingService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ApicastStagingName,
			Labels: apicast.Options.CommonStagingLabels,
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
			Selector: map[string]string{"deploymentConfig": ApicastStagingName},
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
			Name:   ApicastProductionName,
			Labels: apicast.Options.CommonProductionLabels,
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
			Selector: map[string]string{"deploymentConfig": ApicastProductionName},
		},
	}
}

func (apicast *Apicast) StagingDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{APIVersion: "apps.openshift.io/v1", Kind: "DeploymentConfig"},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ApicastStagingName,
			Labels: apicast.Options.CommonStagingLabels,
		},
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: apicast.Options.StagingReplicas,
			Selector: map[string]string{
				"deploymentConfig": ApicastStagingName,
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
							ApicastStagingName,
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
					Labels: apicast.Options.StagingPodTemplateLabels,
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   "9421",
					},
				},
				Spec: v1.PodSpec{
					Affinity:           apicast.Options.StagingAffinity,
					Tolerations:        apicast.Options.StagingTolerations,
					ServiceAccountName: "amp",
					Volumes:            apicast.stagingVolumes(),
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
							Name:            ApicastStagingName,
							Resources:       apicast.Options.StagingResourceRequirements,
							VolumeMounts:    apicast.stagingVolumeMounts(),
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
			Name:   ApicastProductionName,
			Labels: apicast.Options.CommonProductionLabels,
		},
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: apicast.Options.ProductionReplicas,
			Selector: map[string]string{
				"deploymentConfig": ApicastProductionName,
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
							ApicastProductionName,
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
					Labels: apicast.Options.ProductionPodTemplateLabels,
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   "9421",
					},
				},
				Spec: v1.PodSpec{
					Affinity:           apicast.Options.ProductionAffinity,
					Tolerations:        apicast.Options.ProductionTolerations,
					ServiceAccountName: "amp",
					Volumes:            apicast.productionVolumes(),
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
							Name:            ApicastProductionName,
							Resources:       apicast.Options.ProductionResourceRequirements,
							VolumeMounts:    apicast.productionVolumeMounts(),
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
	result := []v1.EnvVar{
		helper.EnvVarFromSecret("THREESCALE_PORTAL_ENDPOINT", "system-master-apicast", SystemSecretSystemMasterApicastProxyConfigsEndpointFieldName),
		helper.EnvVarFromSecret("BACKEND_ENDPOINT_OVERRIDE", BackendSecretBackendListenerSecretName, BackendSecretBackendListenerServiceEndpointFieldName),
		helper.EnvVarFromConfigMap("APICAST_MANAGEMENT_API", "apicast-environment", "APICAST_MANAGEMENT_API"),
		helper.EnvVarFromConfigMap("OPENSSL_VERIFY", "apicast-environment", "OPENSSL_VERIFY"),
		helper.EnvVarFromConfigMap("APICAST_RESPONSE_CODES", "apicast-environment", "APICAST_RESPONSE_CODES"),
	}

	if apicast.Options.ExtendedMetrics {
		result = append(result, helper.EnvVarFromValue("APICAST_EXTENDED_METRICS", "true"))
	}

	return result
}

func (apicast *Apicast) buildApicastStagingEnv() []v1.EnvVar {
	result := []v1.EnvVar{}
	result = append(result, apicast.buildApicastCommonEnv()...)
	result = append(result,
		helper.EnvVarFromValue("APICAST_CONFIGURATION_LOADER", "lazy"),
		helper.EnvVarFromValue("APICAST_CONFIGURATION_CACHE", "0"),
		helper.EnvVarFromValue("THREESCALE_DEPLOYMENT_ENV", "staging"),
	)
	if apicast.Options.StagingLogLevel != nil {
		result = append(result, helper.EnvVarFromValue("APICAST_LOG_LEVEL", *apicast.Options.StagingLogLevel))
	}
	return result
}

func (apicast *Apicast) buildApicastProductionEnv() []v1.EnvVar {
	result := []v1.EnvVar{}
	result = append(result, apicast.buildApicastCommonEnv()...)
	result = append(result,
		helper.EnvVarFromValue("APICAST_CONFIGURATION_LOADER", "boot"),
		helper.EnvVarFromValue("APICAST_CONFIGURATION_CACHE", "300"),
		helper.EnvVarFromValue("THREESCALE_DEPLOYMENT_ENV", "production"),
	)
	if apicast.Options.ProductionWorkers != nil {
		result = append(result, helper.EnvVarFromValue("APICAST_WORKERS", strconv.Itoa(int(*apicast.Options.ProductionWorkers))))
	}
	if apicast.Options.ProductionLogLevel != nil {
		result = append(result, helper.EnvVarFromValue("APICAST_LOG_LEVEL", *apicast.Options.ProductionLogLevel))
	}
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
			Labels: apicast.Options.CommonLabels,
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
			Name:   ApicastStagingName,
			Labels: apicast.Options.CommonStagingLabels,
		},
		Spec: v1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deploymentConfig": ApicastStagingName},
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
			Name:   ApicastProductionName,
			Labels: apicast.Options.CommonProductionLabels,
		},
		Spec: v1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deploymentConfig": ApicastProductionName},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}

func (apicast *Apicast) productionVolumeMounts() []v1.VolumeMount {
	var volumeMounts []v1.VolumeMount

	for _, customPolicy := range apicast.Options.ProductionCustomPolicies {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      policyVolumeName(customPolicy),
			MountPath: path.Join(CustomPoliciesMountBasePath, customPolicy.Name, customPolicy.Version),
			ReadOnly:  true,
		})
	}

	return volumeMounts
}

func (apicast *Apicast) stagingVolumeMounts() []v1.VolumeMount {
	var volumeMounts []v1.VolumeMount

	for _, customPolicy := range apicast.Options.StagingCustomPolicies {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      policyVolumeName(customPolicy),
			MountPath: path.Join(CustomPoliciesMountBasePath, customPolicy.Name, customPolicy.Version),
			ReadOnly:  true,
		})
	}

	return volumeMounts
}

func policyVolumeName(cp CustomPolicy) string {
	return fmt.Sprintf("policy-%s-%s", helper.DNS1123Name(cp.Version), helper.DNS1123Name(cp.Name))
}

func (apicast *Apicast) productionVolumes() []v1.Volume {
	var volumes []v1.Volume

	for _, customPolicy := range apicast.Options.ProductionCustomPolicies {
		volumes = append(volumes, v1.Volume{
			Name: policyVolumeName(customPolicy),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: customPolicy.SecretRef.Name,
				},
			},
		})
	}

	return volumes
}

func (apicast *Apicast) stagingVolumes() []v1.Volume {
	var volumes []v1.Volume

	for _, customPolicy := range apicast.Options.StagingCustomPolicies {
		volumes = append(volumes, v1.Volume{
			Name: policyVolumeName(customPolicy),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: customPolicy.SecretRef.Name,
				},
			},
		})
	}

	return volumes
}
