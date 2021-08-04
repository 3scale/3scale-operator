package component

import (
	"fmt"
	"path"
	"strconv"
	"strings"

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

	CustomPoliciesMountBasePath               = "/opt/app-root/src/policies"
	CustomPoliciesAnnotationNameSegmentPrefix = "apicast-policy-volume"
	CustomPoliciesAnnotationPartialKey        = "apps.3scale.net/" + CustomPoliciesAnnotationNameSegmentPrefix
)

const (
	APIcastDefaultTracingLibrary                    = "jaeger"
	APIcastTracingConfigSecretKey                   = "config"
	APIcastTracingConfigMountBasePath               = "/opt/app-root/src/tracing-configs"
	APIcastTracingConfigAnnotationNameSegmentPrefix = "apicast-tracing-config-volume"
	APIcastTracingConfigAnnotationPartialKey        = "apps.3scale.net/" + APIcastTracingConfigAnnotationNameSegmentPrefix
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
			Name:        ApicastStagingName,
			Labels:      apicast.Options.CommonStagingLabels,
			Annotations: apicast.stagingDeploymentConfigAnnotations(),
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
			Name:        ApicastProductionName,
			Labels:      apicast.Options.CommonProductionLabels,
			Annotations: apicast.productionDeploymentConfigAnnotations(),
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

	stagingTracingConfig := apicast.Options.StagingTracingConfig
	if stagingTracingConfig.Enabled {
		result = append(result, helper.EnvVarFromValue("OPENTRACING_TRACER", stagingTracingConfig.TracingLibrary))

		if stagingTracingConfig.TracingConfigSecretName != nil {
			result = append(result,
				helper.EnvVarFromValue("OPENTRACING_CONFIG",
					path.Join(APIcastTracingConfigMountBasePath, stagingTracingConfig.VolumeName())))
		}
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

	productionTracingConfig := apicast.Options.ProductionTracingConfig
	if productionTracingConfig.Enabled {
		result = append(result, helper.EnvVarFromValue("OPENTRACING_TRACER", productionTracingConfig.TracingLibrary))

		if productionTracingConfig.TracingConfigSecretName != nil {
			result = append(result,
				helper.EnvVarFromValue("OPENTRACING_CONFIG", path.Join(APIcastTracingConfigMountBasePath, productionTracingConfig.VolumeName())))
		}
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
			Name:      customPolicy.VolumeName(),
			MountPath: path.Join(CustomPoliciesMountBasePath, customPolicy.Name, customPolicy.Version),
			ReadOnly:  true,
		})
	}

	if apicast.Options.ProductionTracingConfig.Enabled && apicast.Options.ProductionTracingConfig.TracingConfigSecretName != nil {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      apicast.Options.ProductionTracingConfig.VolumeName(),
			MountPath: APIcastTracingConfigMountBasePath,
		})
	}

	return volumeMounts
}

func (apicast *Apicast) stagingVolumeMounts() []v1.VolumeMount {
	var volumeMounts []v1.VolumeMount

	for _, customPolicy := range apicast.Options.StagingCustomPolicies {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      customPolicy.VolumeName(),
			MountPath: path.Join(CustomPoliciesMountBasePath, customPolicy.Name, customPolicy.Version),
			ReadOnly:  true,
		})
	}

	if apicast.Options.StagingTracingConfig.Enabled && apicast.Options.StagingTracingConfig.TracingConfigSecretName != nil {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      apicast.Options.StagingTracingConfig.VolumeName(),
			MountPath: APIcastTracingConfigMountBasePath,
		})
	}

	return volumeMounts
}

func (apicast *Apicast) productionVolumes() []v1.Volume {
	var volumes []v1.Volume

	for _, customPolicy := range apicast.Options.ProductionCustomPolicies {
		volumes = append(volumes, v1.Volume{
			Name: customPolicy.VolumeName(),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: customPolicy.SecretRef.Name,
				},
			},
		})
	}

	if apicast.Options.ProductionTracingConfig.Enabled && apicast.Options.ProductionTracingConfig.TracingConfigSecretName != nil {
		volumes = append(volumes, v1.Volume{
			Name: apicast.Options.ProductionTracingConfig.VolumeName(),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: *apicast.Options.ProductionTracingConfig.TracingConfigSecretName,
					Items: []v1.KeyToPath{
						v1.KeyToPath{
							Key:  APIcastTracingConfigSecretKey,
							Path: apicast.Options.ProductionTracingConfig.VolumeName(),
						},
					},
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
			Name: customPolicy.VolumeName(),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: customPolicy.SecretRef.Name,
				},
			},
		})
	}

	if apicast.Options.StagingTracingConfig.Enabled && apicast.Options.StagingTracingConfig.TracingConfigSecretName != nil {
		volumes = append(volumes, v1.Volume{
			Name: apicast.Options.StagingTracingConfig.VolumeName(),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: *apicast.Options.StagingTracingConfig.TracingConfigSecretName,
					Items: []v1.KeyToPath{
						v1.KeyToPath{
							Key:  APIcastTracingConfigSecretKey,
							Path: apicast.Options.StagingTracingConfig.VolumeName(),
						},
					},
				},
			},
		})
	}

	return volumes
}

func (apicast *Apicast) productionDeploymentConfigAnnotations() map[string]string {
	annotations := map[string]string{}

	for _, customPolicy := range apicast.Options.ProductionCustomPolicies {
		annotations[customPolicy.AnnotationKey()] = customPolicy.AnnotationValue()
	}

	productionTracingConfig := apicast.Options.ProductionTracingConfig
	if productionTracingConfig.Enabled && productionTracingConfig.TracingConfigSecretName != nil {
		annotations[productionTracingConfig.AnnotationKey()] = productionTracingConfig.VolumeName()
	}

	// keep backward compat
	if len(annotations) == 0 {
		return nil
	}

	return annotations
}

func (apicast *Apicast) stagingDeploymentConfigAnnotations() map[string]string {
	annotations := map[string]string{}

	for _, customPolicy := range apicast.Options.StagingCustomPolicies {
		annotations[customPolicy.AnnotationKey()] = customPolicy.AnnotationValue()
	}

	stagingTracingConfig := apicast.Options.StagingTracingConfig
	if stagingTracingConfig.Enabled && stagingTracingConfig.TracingConfigSecretName != nil {
		annotations[stagingTracingConfig.AnnotationKey()] = apicast.Options.StagingTracingConfig.VolumeName()
	}

	// keep backward compat
	if len(annotations) == 0 {
		return nil
	}

	return annotations
}

// AnnotationsValuesWithAnnotationKeyPrefix returns the annotation values from
// annotations whose keys have the prefix keyPrefix
func AnnotationsValuesWithAnnotationKeyPrefix(annotations map[string]string, keyPrefix string) []string {
	res := []string{}
	for key, val := range annotations {
		if strings.HasPrefix(key, keyPrefix) {
			res = append(res, val)
		}
	}

	return res
}

func ApicastPolicyVolumeNamesFromAnnotations(annotations map[string]string) []string {
	return AnnotationsValuesWithAnnotationKeyPrefix(annotations, CustomPoliciesAnnotationPartialKey)
}

func ApicastTracingConfigVolumeNamesFromAnnotations(annotations map[string]string) []string {
	return AnnotationsValuesWithAnnotationKeyPrefix(annotations, APIcastTracingConfigAnnotationPartialKey)
}
