package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"

	v1 "k8s.io/api/core/v1"
)

type SystemRedisOptionsProvider struct {
	apimanager *appsv1alpha1.APIManager
	options    *component.SystemRedisOptions
}

func NewSystemRedisOptionsProvider(apimanager *appsv1alpha1.APIManager) *SystemRedisOptionsProvider {
	return &SystemRedisOptionsProvider{
		apimanager: apimanager,
		options:    component.NewSystemRedisOptions(),
	}
}

func (r *SystemRedisOptionsProvider) GetSystemRedisOptions() (*component.SystemRedisOptions, error) {
	r.options.AmpRelease = product.ThreescaleRelease
	r.options.ImageTag = product.ThreescaleRelease
	r.options.InsecureImportPolicy = r.apimanager.Spec.ImageStreamTagImportInsecure

	r.options.Image = SystemRedisImageURL()
	if r.apimanager.Spec.System != nil && r.apimanager.Spec.System.RedisImage != nil {
		r.options.Image = *r.apimanager.Spec.System.RedisImage
	}

	r.options.SystemCommonLabels = r.systemCommonLabels()
	r.options.RedisLabels = r.systemRedisLabels()
	r.options.PodTemplateLabels = r.systemRedisPodTemplateLabels(r.options.Image)

	r.setResourceRequirementsOptions()

	err := r.options.Validate()
	if err != nil {
		return nil, fmt.Errorf("GetRedisOptions validating: %w", err)
	}
	return r.options, nil
}

func (r *SystemRedisOptionsProvider) setResourceRequirementsOptions() {
	if *r.apimanager.Spec.ResourceRequirementsEnabled {
		r.options.ContainerResourceRequirements = component.DefaultSystemRedisContainerResourceRequirements()
	} else {
		r.options.ContainerResourceRequirements = &v1.ResourceRequirements{}
	}
}

func (r *SystemRedisOptionsProvider) systemCommonLabels() map[string]string {
	return map[string]string{
		"app":                  *r.apimanager.Spec.AppLabel,
		"threescale_component": "system",
	}
}

func (r *SystemRedisOptionsProvider) systemRedisLabels() map[string]string {
	labels := r.systemCommonLabels()
	labels["threescale_component_element"] = "redis"
	return labels
}

func (r *SystemRedisOptionsProvider) systemRedisPodTemplateLabels(image string) map[string]string {
	labels := helper.MeteringLabels("system-redis", helper.ParseVersion(image), helper.ApplicationType)

	for k, v := range r.systemRedisLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "system-redis"

	return labels
}
