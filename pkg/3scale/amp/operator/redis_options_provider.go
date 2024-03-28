package operator

import (
	"context"
	"fmt"
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RedisOptionsProvider struct {
	apimanager   *appsv1alpha1.APIManager
	namespace    string
	client       client.Client
	options      *component.RedisOptions
	secretSource *helper.SecretSource
}

func NewRedisOptionsProvider(apimanager *appsv1alpha1.APIManager, namespace string, client client.Client) *RedisOptionsProvider {
	return &RedisOptionsProvider{
		apimanager:   apimanager,
		namespace:    namespace,
		client:       client,
		options:      component.NewRedisOptions(),
		secretSource: helper.NewSecretSource(client, namespace),
	}
}

func (r *RedisOptionsProvider) GetRedisOptions() (*component.RedisOptions, error) {
	r.options.AmpRelease = product.ThreescaleRelease
	r.options.BackendImageTag = product.ThreescaleRelease
	r.options.SystemImageTag = product.ThreescaleRelease

	r.options.BackendImage = BackendRedisImageURL()
	if r.apimanager.Spec.Backend != nil && r.apimanager.Spec.Backend.RedisImage != nil {
		r.options.BackendImage = *r.apimanager.Spec.Backend.RedisImage
	}

	r.options.SystemImage = SystemRedisImageURL()
	if r.apimanager.Spec.System != nil && r.apimanager.Spec.System.RedisImage != nil {
		r.options.SystemImage = *r.apimanager.Spec.System.RedisImage
	}

	r.options.SystemCommonLabels = r.systemCommonLabels()
	r.options.SystemRedisLabels = r.systemRedisLabels()
	r.options.SystemRedisPodTemplateLabels = r.systemRedisPodTemplateLabels()
	r.options.BackendCommonLabels = r.backendCommonLabels()
	r.options.BackendRedisLabels = r.backendRedisLabels()
	r.options.BackendRedisPodTemplateLabels = r.backendRedisPodTemplateLabels()
	r.options.SystemRedisPodTemplateAnnotations = r.systemRedisPodTemplateAnnotations()
	r.options.BackendRedisPodTemplateAnnotations = r.backendRedisPodTemplateAnnotations()

	r.setResourceRequirementsOptions()
	r.setNodeAffinityAndTolerationsOptions()

	r.setPersistentVolumeClaimOptions()
	r.setPriorityClassNames()
	r.setTopologySpreadConstraints()

	// Should the operator be reading redis secrets?
	// When HA is disabled, do we support external redis?
	// If answer is true, why does the operator deploy redis?
	// If the answer is no, then it would be sufficient to set default URL's (internal redis url)
	// to options and reconciliate secret for owner reference
	err := r.setSecretBasedOptions()
	if err != nil {
		return nil, fmt.Errorf("GetRedisOptions reading secret options: %w", err)
	}

	err = r.options.Validate()
	if err != nil {
		return nil, fmt.Errorf("GetRedisOptions validating: %w", err)
	}
	return r.options, nil
}

func (r *RedisOptionsProvider) setSecretBasedOptions() error {
	cases := []struct {
		field       *string
		secretName  string
		secretField string
		defValue    string
	}{
		{
			&r.options.BackendStorageURL,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisStorageURLFieldName,
			component.DefaultBackendRedisStorageURL(),
		},
		{
			&r.options.BackendQueuesURL,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisQueuesURLFieldName,
			component.DefaultBackendRedisQueuesURL(),
		},
		{
			&r.options.BackendRedisStorageSentinelHosts,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisStorageSentinelHostsFieldName,
			component.DefaultBackendStorageSentinelHosts(),
		},
		{
			&r.options.BackendRedisStorageSentinelRole,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisStorageSentinelRoleFieldName,
			component.DefaultBackendStorageSentinelRole(),
		},
		{
			&r.options.BackendRedisQueuesSentinelHosts,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisQueuesSentinelHostsFieldName,
			component.DefaultBackendQueuesSentinelHosts(),
		},
		{
			&r.options.BackendRedisQueuesSentinelRole,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisQueuesSentinelRoleFieldName,
			component.DefaultBackendQueuesSentinelRole(),
		},
		{
			&r.options.SystemRedisURL,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisURLFieldName,
			component.DefaultSystemRedisURL(),
		},
		{
			&r.options.SystemRedisSentinelsHosts,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisSentinelHosts,
			component.DefaultSystemRedisSentinelHosts(),
		},
		{
			&r.options.SystemRedisSentinelsRole,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisSentinelRole,
			component.DefaultSystemRedisSentinelRole(),
		},
		{
			&r.options.SystemRedisNamespace,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisNamespace,
			component.DefaultSystemRedisNamespace(),
		},
	}

	for _, option := range cases {
		val, err := r.secretSource.FieldValue(option.secretName, option.secretField, option.defValue)
		if err != nil {
			return err
		}
		*option.field = val
	}

	return nil
}

func (r *RedisOptionsProvider) setResourceRequirementsOptions() {
	if *r.apimanager.Spec.ResourceRequirementsEnabled {
		r.options.BackendRedisContainerResourceRequirements = component.DefaultBackendRedisContainerResourceRequirements()
		r.options.SystemRedisContainerResourceRequirements = component.DefaultSystemRedisContainerResourceRequirements()
	} else {
		r.options.BackendRedisContainerResourceRequirements = &v1.ResourceRequirements{}
		r.options.SystemRedisContainerResourceRequirements = &v1.ResourceRequirements{}
	}

	// Deployment-level ResourceRequirements CR fields have priority over
	// spec.resourceRequirementsEnabled, overwriting that setting when they are
	// defined
	if r.apimanager.Spec.Backend.RedisResources != nil {
		r.options.BackendRedisContainerResourceRequirements = r.apimanager.Spec.Backend.RedisResources
	}
	if r.apimanager.Spec.System.RedisResources != nil {
		r.options.SystemRedisContainerResourceRequirements = r.apimanager.Spec.System.RedisResources
	}
}

func (r *RedisOptionsProvider) setPersistentVolumeClaimOptions() {
	if r.apimanager.Spec.System != nil &&
		r.apimanager.Spec.System.RedisPersistentVolumeClaimSpec != nil {
		r.options.SystemRedisPVCStorageClass = r.apimanager.Spec.System.RedisPersistentVolumeClaimSpec.StorageClassName
	}
	if r.apimanager.Spec.Backend != nil &&
		r.apimanager.Spec.Backend.RedisPersistentVolumeClaimSpec != nil {
		r.options.BackendRedisPVCStorageClass = r.apimanager.Spec.Backend.RedisPersistentVolumeClaimSpec.StorageClassName
	}
}

func (r *RedisOptionsProvider) setNodeAffinityAndTolerationsOptions() {
	r.options.BackendRedisAffinity = r.apimanager.Spec.Backend.RedisAffinity
	r.options.BackendRedisTolerations = r.apimanager.Spec.Backend.RedisTolerations
	r.options.SystemRedisAffinity = r.apimanager.Spec.System.RedisAffinity
	r.options.SystemRedisTolerations = r.apimanager.Spec.System.RedisTolerations
}

func (r *RedisOptionsProvider) systemCommonLabels() map[string]string {
	return map[string]string{
		"app":                  *r.apimanager.Spec.AppLabel,
		"threescale_component": "system",
	}
}

func (r *RedisOptionsProvider) backendCommonLabels() map[string]string {
	return map[string]string{
		"app":                  *r.apimanager.Spec.AppLabel,
		"threescale_component": "backend",
	}
}

func (r *RedisOptionsProvider) systemRedisLabels() map[string]string {
	labels := r.systemCommonLabels()
	labels["threescale_component_element"] = "redis"
	return labels
}

func (r *RedisOptionsProvider) backendRedisLabels() map[string]string {
	labels := r.backendCommonLabels()
	labels["threescale_component_element"] = "redis"
	return labels
}

func (r *RedisOptionsProvider) systemRedisPodTemplateLabels() map[string]string {
	labels := helper.MeteringLabels("system-redis", helper.ApplicationType)

	for k, v := range r.systemRedisLabels() {
		labels[k] = v
	}

	for k, v := range r.apimanager.Spec.System.RedisLabels {
		labels[k] = v
	}

	labels[reconcilers.DeploymentLabelSelector] = "system-redis"

	return labels
}

func (r *RedisOptionsProvider) backendRedisPodTemplateLabels() map[string]string {
	labels := helper.MeteringLabels("backend-redis", helper.ApplicationType)

	for k, v := range r.backendRedisLabels() {
		labels[k] = v
	}

	for k, v := range r.apimanager.Spec.Backend.RedisLabels {
		labels[k] = v
	}

	labels[reconcilers.DeploymentLabelSelector] = "backend-redis"

	return labels
}

func (r *RedisOptionsProvider) setPriorityClassNames() {
	if r.apimanager.Spec.System != nil && r.apimanager.Spec.System.RedisPriorityClassName != nil {
		r.options.SystemRedisPriorityClassName = *r.apimanager.Spec.System.RedisPriorityClassName
	}
	if r.apimanager.Spec.Backend != nil && r.apimanager.Spec.Backend.RedisPriorityClassName != nil {
		r.options.BackendRedisPriorityClassName = *r.apimanager.Spec.Backend.RedisPriorityClassName
	}
}

func (r *RedisOptionsProvider) setTopologySpreadConstraints() {
	if r.apimanager.Spec.System != nil && r.apimanager.Spec.System.RedisTopologySpreadConstraints != nil {
		r.options.SystemRedisTopologySpreadConstraints = r.apimanager.Spec.System.RedisTopologySpreadConstraints
	}
	if r.apimanager.Spec.Backend != nil && r.apimanager.Spec.Backend.RedisTopologySpreadConstraints != nil {
		r.options.BackendRedisTopologySpreadConstraints = r.apimanager.Spec.Backend.RedisTopologySpreadConstraints
	}
}

func (r *RedisOptionsProvider) systemRedisPodTemplateAnnotations() map[string]string {
	annotations := make(map[string]string)
	for k, v := range r.apimanager.Spec.System.RedisAnnotations {
		annotations[k] = v
	}
	annotations["generationID"] = r.getRedisConfigCmResourceVersion()
	return annotations
}

func (r *RedisOptionsProvider) backendRedisPodTemplateAnnotations() map[string]string {
	annotations := make(map[string]string)
	for k, v := range r.apimanager.Spec.Backend.RedisAnnotations {
		annotations[k] = v
	}
	annotations["generationID"] = r.getRedisConfigCmResourceVersion()
	return annotations
}

func (r *RedisOptionsProvider) getRedisConfigCmResourceVersion() string {
	// get resourceVersion from CM and use it as Annotation "generationID" for redis Pods
	key := client.ObjectKey{
		Namespace: r.namespace,
		Name:      "redis-config",
	}
	cm := &v1.ConfigMap{}
	err := r.client.Get(context.Background(), key, cm)
	if err != nil {
		fmt.Printf("Failed to get ConfigMap: %v\n", err)
		return ""
	}
	resourceVersion := cm.GetResourceVersion()
	return resourceVersion
}
