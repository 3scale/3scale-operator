package component

import (
	"context"
	"crypto/md5"
	"fmt"
	"hash/fnv"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	k8sappsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ApicastStagingName                 = "apicast-staging"
	ApicastProductionName              = "apicast-production"
	ApicastProductionInitContainerName = "system-master-svc"

	CustomPoliciesMountBasePath               = "/opt/app-root/src/policies"
	CustomPoliciesAnnotationNameSegmentPrefix = "apicast-policy-volume"
	CustomPoliciesAnnotationPartialKey        = "apps.3scale.net/" + CustomPoliciesAnnotationNameSegmentPrefix

	CustomEnvironmentsMountBasePath               = "/opt/app-root/src/environments"
	CustomEnvironmentsAnnotationNameSegmentPrefix = "apicast-env-volume"
	CustomEnvironmentsAnnotationPartialKey        = "apps.3scale.net/" + CustomEnvironmentsAnnotationNameSegmentPrefix

	HTTPSCertificatesMountPath  = "/var/run/secrets/tls"
	HTTPSCertificatesVolumeName = "https-certificates"

	APIcastEnvironmentConfigMapName = "apicast-environment"

	OpentelemetryConfigurationVolumeName = "otel-volume"
	OpentelemetryConfigMountBasePath     = "/opt/app-root/src/otel-configs"
)

const (
	APIcastTracingConfigSecretKey                   = "config"
	APIcastTracingConfigMountBasePath               = "/opt/app-root/src/tracing-configs"
	APIcastTracingConfigAnnotationNameSegmentPrefix = "apicast-tracing-config-volume"
	APIcastTracingConfigAnnotationPartialKey        = "apps.3scale.net/" + APIcastTracingConfigAnnotationNameSegmentPrefix
	APIcastOpentelemetryConfigAnnotationPartialKey  = "apps.3scale.net/" + OpentelemetryConfigurationVolumeName
)

const (
	APIcastEnvironmentCMAnnotation             = "apimanager.apps.3scale.net/env-configmap-hash"
	HttpsCertSecretResverAnnotationPrefix      = "apimanager.apps.3scale.net/https-cert-secret-resource-version-"
	OpenTelemetrySecretResverAnnotationPrefix  = "apimanager.apps.3scale.net/opentelemetry-secret-resource-version-"
	CustomEnvSecretResverAnnotationPrefix      = "apimanager.apps.3scale.net/customenv-secret-resource-version-"
	CustomPoliciesSecretResverAnnotationPrefix = "apimanager.apps.3scale.net/custompolicy-secret-resource-version-"
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
			Ports:    apicast.stagingServicePorts(),
			Selector: map[string]string{reconcilers.DeploymentLabelSelector: ApicastStagingName},
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
			Ports:    apicast.productionServicePorts(),
			Selector: map[string]string{reconcilers.DeploymentLabelSelector: ApicastProductionName},
		},
	}
}

func (apicast *Apicast) StagingDeployment(ctx context.Context, k8sclient client.Client, containerImage string) (*k8sappsv1.Deployment, error) {
	watchedSecretAnnotations, err := ComputeWatchedSecretAnnotations(ctx, k8sclient, ApicastStagingName, apicast.Options.Namespace, apicast)
	if err != nil {
		return nil, err
	}

	return &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: reconcilers.DeploymentAPIVersion, Kind: reconcilers.DeploymentKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:        ApicastStagingName,
			Labels:      apicast.Options.CommonStagingLabels,
			Annotations: apicast.stagingDeploymentAnnotations(),
		},
		Spec: k8sappsv1.DeploymentSpec{
			Replicas: &apicast.Options.StagingReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					reconcilers.DeploymentLabelSelector: ApicastStagingName,
				},
			},
			Strategy: k8sappsv1.DeploymentStrategy{
				RollingUpdate: &k8sappsv1.RollingUpdateDeployment{
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "25%",
					},
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "25%",
					},
				},
				Type: k8sappsv1.RollingUpdateDeploymentStrategyType,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      apicast.Options.StagingPodTemplateLabels,
					Annotations: apicast.stagingPodAnnotations(watchedSecretAnnotations),
				},
				Spec: v1.PodSpec{
					Affinity:           apicast.Options.StagingAffinity,
					Tolerations:        apicast.Options.StagingTolerations,
					ServiceAccountName: "amp",
					Volumes:            apicast.stagingVolumes(),
					Containers: []v1.Container{
						{
							Ports:           apicast.stagingContainerPorts(),
							Env:             apicast.buildApicastStagingEnv(),
							Image:           containerImage,
							ImagePullPolicy: v1.PullIfNotPresent,
							Name:            ApicastStagingName,
							Resources:       apicast.Options.StagingResourceRequirements,
							VolumeMounts:    apicast.stagingVolumeMounts(),
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{HTTPGet: &v1.HTTPGetAction{
									Path: "/status/live",
									Port: intstr.FromInt32(8090),
								}},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      5,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{HTTPGet: &v1.HTTPGetAction{
									Path: "/status/ready",
									Port: intstr.FromInt32(8090),
								}},
								InitialDelaySeconds: 15,
								TimeoutSeconds:      5,
								PeriodSeconds:       30,
							},
						},
					},
					PriorityClassName:         apicast.Options.PriorityClassNameStaging,
					TopologySpreadConstraints: apicast.Options.TopologySpreadConstraintsStaging,
				},
			},
		},
	}, nil
}

func (apicast *Apicast) ProductionDeployment(ctx context.Context, k8sclient client.Client, containerImage string) (*k8sappsv1.Deployment, error) {
	watchedSecretAnnotations, err := ComputeWatchedSecretAnnotations(ctx, k8sclient, ApicastStagingName, apicast.Options.Namespace, apicast)
	if err != nil {
		return nil, err
	}

	return &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: reconcilers.DeploymentAPIVersion, Kind: reconcilers.DeploymentKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:        ApicastProductionName,
			Labels:      apicast.Options.CommonProductionLabels,
			Annotations: apicast.productionDeploymentAnnotations(),
		},
		Spec: k8sappsv1.DeploymentSpec{
			Replicas: &apicast.Options.ProductionReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					reconcilers.DeploymentLabelSelector: ApicastProductionName,
				},
			},
			Strategy: k8sappsv1.DeploymentStrategy{
				RollingUpdate: &k8sappsv1.RollingUpdateDeployment{
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "25%",
					},
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "25%",
					},
				},
				Type: k8sappsv1.RollingUpdateDeploymentStrategyType,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      apicast.Options.ProductionPodTemplateLabels,
					Annotations: apicast.productionPodAnnotations(watchedSecretAnnotations),
				},
				Spec: v1.PodSpec{
					Affinity:           apicast.Options.ProductionAffinity,
					Tolerations:        apicast.Options.ProductionTolerations,
					ServiceAccountName: "amp",
					Volumes:            apicast.productionVolumes(),
					InitContainers: []v1.Container{
						{
							Name:    ApicastProductionInitContainerName,
							Image:   containerImage,
							Command: []string{"sh", "-c", "until $(curl --output /dev/null --silent --fail --head http://system-master:3000/status); do sleep $SLEEP_SECONDS; done"},
							Env: []v1.EnvVar{
								{
									Name:  "SLEEP_SECONDS",
									Value: "1",
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Ports:           apicast.productionContainerPorts(),
							Env:             apicast.buildApicastProductionEnv(),
							Image:           containerImage,
							ImagePullPolicy: v1.PullIfNotPresent,
							Name:            ApicastProductionName,
							Resources:       apicast.Options.ProductionResourceRequirements,
							VolumeMounts:    apicast.productionVolumeMounts(),
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{HTTPGet: &v1.HTTPGetAction{
									Path: "/status/live",
									Port: intstr.FromInt32(8090),
								}},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      5,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{HTTPGet: &v1.HTTPGetAction{
									Path: "/status/ready",
									Port: intstr.FromInt32(8090),
								}},
								InitialDelaySeconds: 15,
								TimeoutSeconds:      5,
								PeriodSeconds:       30,
							},
						},
					},
					PriorityClassName:         apicast.Options.PriorityClassNameProduction,
					TopologySpreadConstraints: apicast.Options.TopologySpreadConstraintsProduction,
				},
			},
		},
	}, nil
}

func (apicast *Apicast) buildApicastCommonEnv() []v1.EnvVar {
	result := []v1.EnvVar{
		helper.EnvVarFromSecret("THREESCALE_PORTAL_ENDPOINT", "system-master-apicast", SystemSecretSystemMasterApicastProxyConfigsEndpointFieldName),
		helper.EnvVarFromSecret("BACKEND_ENDPOINT_OVERRIDE", BackendSecretBackendListenerSecretName, BackendSecretBackendListenerServiceEndpointFieldName),
		helper.EnvVarFromConfigMap("APICAST_MANAGEMENT_API", APIcastEnvironmentConfigMapName, "APICAST_MANAGEMENT_API"),
		helper.EnvVarFromConfigMap("OPENSSL_VERIFY", APIcastEnvironmentConfigMapName, "OPENSSL_VERIFY"),
		helper.EnvVarFromConfigMap("APICAST_RESPONSE_CODES", APIcastEnvironmentConfigMapName, "APICAST_RESPONSE_CODES"),
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

	var customEnvPaths []string
	for _, customEnvSecret := range apicast.Options.StagingCustomEnvironments {
		for fileKey := range customEnvSecret.Data {
			customEnvPaths = append(customEnvPaths, path.Join(CustomEnvironmentsMountBasePath, customEnvSecret.GetName(), fileKey))
		}
	}

	if len(customEnvPaths) > 0 {
		// Sort customenvPaths to ensure deterministic reconciliation
		sort.Strings(customEnvPaths)
		result = append(result, helper.EnvVarFromValue("APICAST_ENVIRONMENT", strings.Join(customEnvPaths, ":")))
	}

	if apicast.Options.StagingHTTPSPort != nil {
		result = append(result, helper.EnvVarFromValue("APICAST_HTTPS_PORT", strconv.FormatInt(int64(*apicast.Options.StagingHTTPSPort), 10)))
	}

	if apicast.Options.StagingHTTPSVerifyDepth != nil {
		result = append(result, helper.EnvVarFromValue("APICAST_HTTPS_VERIFY_DEPTH", strconv.FormatInt(*apicast.Options.StagingHTTPSVerifyDepth, 10)))
	}

	if apicast.Options.StagingHTTPSCertificateSecretName != nil {
		result = append(result,
			helper.EnvVarFromValue("APICAST_HTTPS_CERTIFICATE", fmt.Sprintf("%s/%s", HTTPSCertificatesMountPath, v1.TLSCertKey)),
			helper.EnvVarFromValue("APICAST_HTTPS_CERTIFICATE_KEY", fmt.Sprintf("%s/%s", HTTPSCertificatesMountPath, v1.TLSPrivateKeyKey)))
	}

	if apicast.Options.StagingAllProxy != nil {
		result = append(result, helper.EnvVarFromValue("ALL_PROXY", *apicast.Options.StagingAllProxy))
	}

	if apicast.Options.StagingHTTPProxy != nil {
		result = append(result, helper.EnvVarFromValue("HTTP_PROXY", *apicast.Options.StagingHTTPProxy))
	}

	if apicast.Options.StagingHTTPSProxy != nil {
		result = append(result, helper.EnvVarFromValue("HTTPS_PROXY", *apicast.Options.StagingHTTPSProxy))
	}

	if apicast.Options.StagingNoProxy != nil {
		result = append(result, helper.EnvVarFromValue("NO_PROXY", *apicast.Options.StagingNoProxy))
	}

	if apicast.Options.StagingServiceCacheSize != nil {
		result = append(result, helper.EnvVarFromValue("APICAST_SERVICE_CACHE_SIZE", fmt.Sprintf("%d", *apicast.Options.StagingServiceCacheSize)))
	}

	if apicast.Options.StagingOpentelemetry.Enabled {
		result = append(result, helper.EnvVarFromValue("OPENTELEMETRY", "1"))
		result = append(result, helper.EnvVarFromValue("OPENTELEMETRY_CONFIG", apicast.Options.StagingOpentelemetry.ConfigFile))
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

	var customEnvPaths []string
	for _, customEnvSecret := range apicast.Options.ProductionCustomEnvironments {
		for fileKey := range customEnvSecret.Data {
			customEnvPaths = append(customEnvPaths, path.Join(CustomEnvironmentsMountBasePath, customEnvSecret.GetName(), fileKey))
		}
	}

	if len(customEnvPaths) > 0 {
		// Sort customenvPaths to ensure deterministic reconciliation
		sort.Strings(customEnvPaths)
		result = append(result, helper.EnvVarFromValue("APICAST_ENVIRONMENT", strings.Join(customEnvPaths, ":")))
	}

	if apicast.Options.ProductionHTTPSPort != nil {
		result = append(result, helper.EnvVarFromValue("APICAST_HTTPS_PORT", strconv.FormatInt(int64(*apicast.Options.ProductionHTTPSPort), 10)))
	}

	if apicast.Options.ProductionHTTPSVerifyDepth != nil {
		result = append(result, helper.EnvVarFromValue("APICAST_HTTPS_VERIFY_DEPTH", strconv.FormatInt(*apicast.Options.ProductionHTTPSVerifyDepth, 10)))
	}

	if apicast.Options.ProductionHTTPSCertificateSecretName != nil {
		result = append(result,
			helper.EnvVarFromValue("APICAST_HTTPS_CERTIFICATE", path.Join(HTTPSCertificatesMountPath, v1.TLSCertKey)),
			helper.EnvVarFromValue("APICAST_HTTPS_CERTIFICATE_KEY", path.Join(HTTPSCertificatesMountPath, v1.TLSPrivateKeyKey)),
		)
	}

	if apicast.Options.ProductionAllProxy != nil {
		result = append(result, helper.EnvVarFromValue("ALL_PROXY", *apicast.Options.ProductionAllProxy))
	}

	if apicast.Options.ProductionHTTPProxy != nil {
		result = append(result, helper.EnvVarFromValue("HTTP_PROXY", *apicast.Options.ProductionHTTPProxy))
	}

	if apicast.Options.ProductionHTTPSProxy != nil {
		result = append(result, helper.EnvVarFromValue("HTTPS_PROXY", *apicast.Options.ProductionHTTPSProxy))
	}

	if apicast.Options.ProductionNoProxy != nil {
		result = append(result, helper.EnvVarFromValue("NO_PROXY", *apicast.Options.ProductionNoProxy))
	}

	if apicast.Options.ProductionServiceCacheSize != nil {
		result = append(result, helper.EnvVarFromValue("APICAST_SERVICE_CACHE_SIZE", fmt.Sprintf("%d", *apicast.Options.ProductionServiceCacheSize)))
	}

	if apicast.Options.ProductionOpentelemetry.Enabled {
		result = append(result, helper.EnvVarFromValue("OPENTELEMETRY", "1"))
		result = append(result, helper.EnvVarFromValue("OPENTELEMETRY_CONFIG", apicast.Options.ProductionOpentelemetry.ConfigFile))
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

func (apicast *Apicast) StagingPodDisruptionBudget() *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ApicastStagingName,
			Labels: apicast.Options.CommonStagingLabels,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{reconcilers.DeploymentLabelSelector: ApicastStagingName},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}

func (apicast *Apicast) ProductionPodDisruptionBudget() *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ApicastProductionName,
			Labels: apicast.Options.CommonProductionLabels,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{reconcilers.DeploymentLabelSelector: ApicastProductionName},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: PDB_MAX_UNAVAILABLE_POD_NUMBER},
		},
	}
}

func (apicast *Apicast) productionVolumeMounts() []v1.VolumeMount {
	volumeMounts := []v1.VolumeMount{}

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

	for _, customEnvSecret := range apicast.Options.ProductionCustomEnvironments {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      customEnvVolumeName(customEnvSecret),
			MountPath: path.Join(CustomEnvironmentsMountBasePath, customEnvSecret.GetName()),
			ReadOnly:  true,
		})
	}

	if apicast.Options.ProductionHTTPSCertificateSecretName != nil {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      HTTPSCertificatesVolumeName,
			MountPath: HTTPSCertificatesMountPath,
			ReadOnly:  true,
		})
	}

	if apicast.Options.ProductionOpentelemetry.Enabled {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      OpentelemetryConfigurationVolumeName,
			MountPath: OpentelemetryConfigMountBasePath,
			ReadOnly:  true,
		})
	}

	return volumeMounts
}

func (apicast *Apicast) stagingVolumeMounts() []v1.VolumeMount {
	volumeMounts := []v1.VolumeMount{}

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

	for _, customEnvSecret := range apicast.Options.StagingCustomEnvironments {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      customEnvVolumeName(customEnvSecret),
			MountPath: path.Join(CustomEnvironmentsMountBasePath, customEnvSecret.GetName()),
			ReadOnly:  true,
		})
	}

	if apicast.Options.StagingHTTPSCertificateSecretName != nil {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      HTTPSCertificatesVolumeName,
			MountPath: HTTPSCertificatesMountPath,
			ReadOnly:  true,
		})
	}

	if apicast.Options.StagingOpentelemetry.Enabled {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      OpentelemetryConfigurationVolumeName,
			MountPath: OpentelemetryConfigMountBasePath,
			ReadOnly:  true,
		})
	}

	return volumeMounts
}

func customEnvVolumeName(secret *v1.Secret) string {
	return fmt.Sprintf("custom-env-%s", secret.GetName())
}

func customEnvAnnotationKey(secret *v1.Secret) string {
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/
	// prefix/name: value
	// The name segment is required and must be 63 characters or less
	// Currently: len(CustomEnvironmentsAnnotationNameSegmentPrefix) + 32 (from the hash) = 50
	return fmt.Sprintf("%s-%x", CustomEnvironmentsAnnotationPartialKey, md5.Sum([]byte(customEnvVolumeName(secret))))
}

func customEnvAnnotationValue(secret *v1.Secret) string {
	return customEnvVolumeName(secret)
}

func (apicast *Apicast) productionVolumes() []v1.Volume {
	volumes := []v1.Volume{}

	for _, customPolicy := range apicast.Options.ProductionCustomPolicies {
		volumes = append(volumes, v1.Volume{
			Name: customPolicy.VolumeName(),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: customPolicy.Secret.Name,
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
						{
							Key:  APIcastTracingConfigSecretKey,
							Path: apicast.Options.ProductionTracingConfig.VolumeName(),
						},
					},
				},
			},
		})
	}

	if apicast.Options.ProductionOpentelemetry.Enabled {
		volumes = append(volumes, v1.Volume{
			Name: OpentelemetryConfigurationVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: apicast.Options.ProductionOpentelemetry.Secret.Name,
					Optional:   &[]bool{false}[0],
				},
			},
		})
	}

	for _, customEnvSecret := range apicast.Options.ProductionCustomEnvironments {
		volumes = append(volumes, v1.Volume{
			Name: customEnvVolumeName(customEnvSecret),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: customEnvSecret.GetName(),
				},
			},
		})
	}

	if apicast.Options.ProductionHTTPSCertificateSecretName != nil {
		volumes = append(volumes, v1.Volume{
			Name: HTTPSCertificatesVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: *apicast.Options.ProductionHTTPSCertificateSecretName,
				},
			},
		})
	}

	return volumes
}

func (apicast *Apicast) stagingVolumes() []v1.Volume {
	volumes := []v1.Volume{}

	for _, customPolicy := range apicast.Options.StagingCustomPolicies {
		volumes = append(volumes, v1.Volume{
			Name: customPolicy.VolumeName(),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: customPolicy.Secret.Name,
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
						{
							Key:  APIcastTracingConfigSecretKey,
							Path: apicast.Options.StagingTracingConfig.VolumeName(),
						},
					},
				},
			},
		})
	}

	if apicast.Options.StagingOpentelemetry.Enabled {
		volumes = append(volumes, v1.Volume{
			Name: OpentelemetryConfigurationVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: apicast.Options.StagingOpentelemetry.Secret.Name,
					Optional:   &[]bool{false}[0],
				},
			},
		})
	}

	for _, customEnvSecret := range apicast.Options.StagingCustomEnvironments {
		volumes = append(volumes, v1.Volume{
			Name: customEnvVolumeName(customEnvSecret),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: customEnvSecret.GetName(),
				},
			},
		})
	}

	if apicast.Options.StagingHTTPSCertificateSecretName != nil {
		volumes = append(volumes, v1.Volume{
			Name: HTTPSCertificatesVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: *apicast.Options.StagingHTTPSCertificateSecretName,
				},
			},
		})
	}

	return volumes
}

func (apicast *Apicast) productionDeploymentAnnotations() map[string]string {
	annotations := map[string]string{}

	for _, customPolicy := range apicast.Options.ProductionCustomPolicies {
		annotations[customPolicy.AnnotationKey()] = customPolicy.AnnotationValue()
	}

	productionTracingConfig := apicast.Options.ProductionTracingConfig
	if productionTracingConfig.Enabled && productionTracingConfig.TracingConfigSecretName != nil {
		annotations[productionTracingConfig.AnnotationKey()] = productionTracingConfig.VolumeName()
	}

	for _, customEnvSecret := range apicast.Options.ProductionCustomEnvironments {
		annotations[customEnvAnnotationKey(customEnvSecret)] = customEnvAnnotationValue(customEnvSecret)
	}

	return annotations
}

func (apicast *Apicast) stagingDeploymentAnnotations() map[string]string {
	annotations := map[string]string{}

	for _, customPolicy := range apicast.Options.StagingCustomPolicies {
		annotations[customPolicy.AnnotationKey()] = customPolicy.AnnotationValue()
	}

	stagingTracingConfig := apicast.Options.StagingTracingConfig
	if stagingTracingConfig.Enabled && stagingTracingConfig.TracingConfigSecretName != nil {
		annotations[stagingTracingConfig.AnnotationKey()] = apicast.Options.StagingTracingConfig.VolumeName()
	}

	for _, customEnvSecret := range apicast.Options.StagingCustomEnvironments {
		annotations[customEnvAnnotationKey(customEnvSecret)] = customEnvAnnotationValue(customEnvSecret)
	}

	return annotations
}

func (apicast *Apicast) productionContainerPorts() []v1.ContainerPort {
	ports := []v1.ContainerPort{
		{ContainerPort: 8080, Protocol: v1.ProtocolTCP},
		{ContainerPort: 8090, Protocol: v1.ProtocolTCP},
		{ContainerPort: 9421, Protocol: v1.ProtocolTCP, Name: "metrics"},
	}

	if apicast.Options.ProductionHTTPSPort != nil {
		ports = append(ports,
			v1.ContainerPort{Name: "httpsproxy", ContainerPort: *apicast.Options.ProductionHTTPSPort, Protocol: v1.ProtocolTCP})
	}

	return ports
}

func (apicast *Apicast) productionServicePorts() []v1.ServicePort {
	ports := []v1.ServicePort{
		{Name: "gateway", Protocol: v1.ProtocolTCP, Port: 8080, TargetPort: intstr.FromInt32(8080)},
		{Name: "management", Protocol: v1.ProtocolTCP, Port: 8090, TargetPort: intstr.FromInt32(8090)},
		{Name: "metrics", Protocol: v1.ProtocolTCP, Port: 9421, TargetPort: intstr.FromInt32(9421)},
	}

	if apicast.Options.ProductionHTTPSPort != nil {
		ports = append(ports,
			v1.ServicePort{Name: "httpsproxy", Port: *apicast.Options.ProductionHTTPSPort, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromString("httpsproxy")},
		)
	}

	return ports
}

func (apicast *Apicast) stagingContainerPorts() []v1.ContainerPort {
	ports := []v1.ContainerPort{
		{ContainerPort: 8080, Protocol: v1.ProtocolTCP},
		{ContainerPort: 8090, Protocol: v1.ProtocolTCP},
		{ContainerPort: 9421, Protocol: v1.ProtocolTCP, Name: "metrics"},
	}

	if apicast.Options.StagingHTTPSPort != nil {
		ports = append(ports,
			v1.ContainerPort{Name: "httpsproxy", ContainerPort: *apicast.Options.StagingHTTPSPort, Protocol: v1.ProtocolTCP})
	}

	return ports
}

func (apicast *Apicast) stagingServicePorts() []v1.ServicePort {
	ports := []v1.ServicePort{
		{Name: "gateway", Protocol: v1.ProtocolTCP, Port: 8080, TargetPort: intstr.FromInt32(8080)},
		{Name: "management", Protocol: v1.ProtocolTCP, Port: 8090, TargetPort: intstr.FromInt32(8090)},
		{Name: "metrics", Protocol: v1.ProtocolTCP, Port: 9421, TargetPort: intstr.FromInt32(9421)},
	}

	if apicast.Options.StagingHTTPSPort != nil {
		ports = append(ports,
			v1.ServicePort{Name: "httpsproxy", Port: *apicast.Options.StagingHTTPSPort, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromString("httpsproxy")},
		)
	}

	return ports
}

func (apicast *Apicast) stagingPodAnnotations(watchedSecretAnnotations map[string]string) map[string]string {
	annotations := map[string]string{
		"prometheus.io/scrape":         "true",
		"prometheus.io/port":           "9421",
		APIcastEnvironmentCMAnnotation: apicast.envConfigMapHash(),
	}

	for key, val := range watchedSecretAnnotations {
		annotations[key] = val
	}

	for key, val := range apicast.Options.StagingPodTemplateAnnotations {
		annotations[key] = val
	}

	return annotations
}

func (apicast *Apicast) productionPodAnnotations(watchedSecretAnnotations map[string]string) map[string]string {
	annotations := map[string]string{
		"prometheus.io/scrape":         "true",
		"prometheus.io/port":           "9421",
		APIcastEnvironmentCMAnnotation: apicast.envConfigMapHash(),
	}

	for key, val := range watchedSecretAnnotations {
		annotations[key] = val
	}

	for key, val := range apicast.Options.ProductionPodTemplateAnnotations {
		annotations[key] = val
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

func ApicastOpentelemetryConfigVolumeNamesFromAnnotations(annotations map[string]string) []string {
	return AnnotationsValuesWithAnnotationKeyPrefix(annotations, APIcastOpentelemetryConfigAnnotationPartialKey)
}

func ApicastEnvVolumeNamesFromAnnotations(annotations map[string]string) []string {
	return AnnotationsValuesWithAnnotationKeyPrefix(annotations, CustomEnvironmentsAnnotationPartialKey)
}

// APIcast environment hash
// When any of the fields used to compute the hash change, the hash will change and the apicast deployment will rollout
func (apicast *Apicast) envConfigMapHash() string {
	h := fnv.New32a()
	h.Write([]byte(apicast.Options.ManagementAPI))
	h.Write([]byte(apicast.Options.OpenSSLVerify))
	h.Write([]byte(apicast.Options.ResponseCodes))
	val := h.Sum32()
	return fmt.Sprint(val)
}
