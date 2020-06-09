package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"

	v1 "k8s.io/api/core/v1"
)

type BackendRedisOptionsProvider struct {
	apimanager *appsv1alpha1.APIManager
	options    *component.BackendRedisOptions
}

func NewBackendRedisOptionsProvider(apimanager *appsv1alpha1.APIManager) *BackendRedisOptionsProvider {
	return &BackendRedisOptionsProvider{
		apimanager: apimanager,
		options:    component.NewBackendRedisOptions(),
	}
}

func (r *BackendRedisOptionsProvider) GetBackendRedisOptions() (*component.BackendRedisOptions, error) {
	r.options.AmpRelease = product.ThreescaleRelease
	r.options.ImageTag = product.ThreescaleRelease
	r.options.InsecureImportPolicy = r.apimanager.Spec.ImageStreamTagImportInsecure

	r.options.Image = BackendRedisImageURL()
	if r.apimanager.Spec.Backend != nil && r.apimanager.Spec.Backend.RedisImage != nil {
		r.options.Image = *r.apimanager.Spec.Backend.RedisImage
	}

	r.options.BackendCommonLabels = r.backendCommonLabels()
	r.options.RedisLabels = r.backendRedisLabels()
	r.options.PodTemplateLabels = r.backendRedisPodTemplateLabels(r.options.Image)

	r.setResourceRequirementsOptions()

	err := r.options.Validate()
	if err != nil {
		return nil, fmt.Errorf("GetRedisOptions validating: %w", err)
	}
	return r.options, nil
}

func (r *BackendRedisOptionsProvider) setResourceRequirementsOptions() {
	if *r.apimanager.Spec.ResourceRequirementsEnabled {
		r.options.ContainerResourceRequirements = component.DefaultBackendRedisContainerResourceRequirements()
	} else {
		r.options.ContainerResourceRequirements = &v1.ResourceRequirements{}
	}
}

func (r *BackendRedisOptionsProvider) systemCommonLabels() map[string]string {
	return map[string]string{
		"app":                  *r.apimanager.Spec.AppLabel,
		"threescale_component": "system",
	}
}

func (r *BackendRedisOptionsProvider) backendCommonLabels() map[string]string {
	return map[string]string{
		"app":                  *r.apimanager.Spec.AppLabel,
		"threescale_component": "backend",
	}
}

func (r *BackendRedisOptionsProvider) backendRedisLabels() map[string]string {
	labels := r.backendCommonLabels()
	labels["threescale_component_element"] = "redis"
	return labels
}

func (r *BackendRedisOptionsProvider) backendRedisPodTemplateLabels(image string) map[string]string {
	labels := helper.MeteringLabels("backend-redis", helper.ParseVersion(image), helper.ApplicationType)

	for k, v := range r.backendRedisLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "backend-redis"

	return labels
}
