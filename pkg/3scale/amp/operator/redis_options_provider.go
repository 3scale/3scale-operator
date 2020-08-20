package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"

	v1 "k8s.io/api/core/v1"
)

type RedisOptionsProvider struct {
	apimanager *appsv1alpha1.APIManager
	options    *component.RedisOptions
}

func NewRedisOptionsProvider(apimanager *appsv1alpha1.APIManager) *RedisOptionsProvider {
	return &RedisOptionsProvider{
		apimanager: apimanager,
		options:    component.NewRedisOptions(),
	}
}

func (r *RedisOptionsProvider) GetRedisOptions() (*component.RedisOptions, error) {
	r.options.AmpRelease = product.ThreescaleRelease
	r.options.BackendImageTag = product.ThreescaleRelease
	r.options.SystemImageTag = product.ThreescaleRelease
	r.options.InsecureImportPolicy = r.apimanager.Spec.ImageStreamTagImportInsecure

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
	r.options.SystemRedisPodTemplateLabels = r.systemRedisPodTemplateLabels(r.options.SystemImage)
	r.options.BackendCommonLabels = r.backendCommonLabels()
	r.options.BackendRedisLabels = r.backendRedisLabels()
	r.options.BackendRedisPodTemplateLabels = r.backendRedisPodTemplateLabels(r.options.BackendImage)

	r.setResourceRequirementsOptions()
	r.setNodeAffinityAndTolerationsOptions()

	r.setPersistentVolumeClaimOptions()

	err := r.options.Validate()
	if err != nil {
		return nil, fmt.Errorf("GetRedisOptions validating: %w", err)
	}
	return r.options, nil
}

func (r *RedisOptionsProvider) setResourceRequirementsOptions() {
	if *r.apimanager.Spec.ResourceRequirementsEnabled {
		r.options.BackendRedisContainerResourceRequirements = component.DefaultBackendRedisContainerResourceRequirements()
		r.options.SystemRedisContainerResourceRequirements = component.DefaultSystemRedisContainerResourceRequirements()
	} else {
		r.options.BackendRedisContainerResourceRequirements = &v1.ResourceRequirements{}
		r.options.SystemRedisContainerResourceRequirements = &v1.ResourceRequirements{}
	}

	// DeploymentConfig-level ResourceRequirements CR fields have priority over
	// spec.resourceRequirementsEnabled, overwriting that setting when they are
	// defined
	if r.apimanager.Spec.Backend.RedisResources != nil {
		r.options.BackendRedisContainerResourceRequirements = r.apimanager.Spec.Backend.RedisResources
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

func (r *RedisOptionsProvider) systemRedisPodTemplateLabels(image string) map[string]string {
	labels := helper.MeteringLabels("system-redis", helper.ParseVersion(image), helper.ApplicationType)

	for k, v := range r.systemRedisLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "system-redis"

	return labels
}

func (r *RedisOptionsProvider) backendRedisPodTemplateLabels(image string) map[string]string {
	labels := helper.MeteringLabels("backend-redis", helper.ParseVersion(image), helper.ApplicationType)

	for k, v := range r.backendRedisLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "backend-redis"

	return labels
}
