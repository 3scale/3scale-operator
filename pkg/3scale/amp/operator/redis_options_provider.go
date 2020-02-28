package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
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
	r.options.AppLabel = *r.apimanager.Spec.AppLabel
	r.options.AmpRelease = product.ThreescaleRelease
	r.options.InsecureImportPolicy = r.apimanager.Spec.ImageStreamTagImportInsecure

	r.options.BackendImage = BackendRedisImageURL()
	if r.apimanager.Spec.Backend != nil && r.apimanager.Spec.Backend.RedisImage != nil {
		r.options.BackendImage = *r.apimanager.Spec.Backend.RedisImage
	}

	r.options.SystemImage = SystemRedisImageURL()
	if r.apimanager.Spec.System != nil && r.apimanager.Spec.System.RedisImage != nil {
		r.options.SystemImage = *r.apimanager.Spec.System.RedisImage
	}

	r.setResourceRequirementsOptions()

	err := r.options.Validate()
	return r.options, err
}

func (r *RedisOptionsProvider) setResourceRequirementsOptions() {
	if *r.apimanager.Spec.ResourceRequirementsEnabled {
		r.options.BackendRedisContainerResourceRequirements = component.DefaultBackendRedisContainerResourceRequirements()
		r.options.SystemRedisContainerResourceRequirements = component.DefaultSystemRedisContainerResourceRequirements()
	} else {
		r.options.BackendRedisContainerResourceRequirements = &v1.ResourceRequirements{}
		r.options.SystemRedisContainerResourceRequirements = &v1.ResourceRequirements{}
	}
}
