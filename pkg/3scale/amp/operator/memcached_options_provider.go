package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"

	v1 "k8s.io/api/core/v1"
)

type MemcachedOptionsProvider struct {
	apimanager       *appsv1alpha1.APIManager
	memcachedOptions *component.MemcachedOptions
}

func NewMemcachedOptionsProvider(apimanager *appsv1alpha1.APIManager) *MemcachedOptionsProvider {
	return &MemcachedOptionsProvider{
		apimanager:       apimanager,
		memcachedOptions: component.NewMemcachedOptions(),
	}
}

func (m *MemcachedOptionsProvider) GetMemcachedOptions() (*component.MemcachedOptions, error) {
	m.memcachedOptions.ImageTag = product.ThreescaleRelease

	imageOpts, err := NewAmpImagesOptionsProvider(m.apimanager).GetAmpImagesOptions()
	if err != nil {
		return nil, fmt.Errorf("GetMemcachedOptions reading image options: %w", err)
	}
	m.memcachedOptions.DeploymentLabels = m.deploymentLabels()
	m.memcachedOptions.PodTemplateLabels = m.podTemplateLabels(imageOpts.SystemMemcachedImage)

	m.setResourceRequirementsOptions()
	m.setNodeAffinityAndTolerationsOptions()

	err = m.memcachedOptions.Validate()
	if err != nil {
		return nil, fmt.Errorf("GetMemcachedOptions validating: %w", err)
	}
	return m.memcachedOptions, nil
}

func (m *MemcachedOptionsProvider) setResourceRequirementsOptions() {
	if *m.apimanager.Spec.ResourceRequirementsEnabled {
		m.memcachedOptions.ResourceRequirements = component.DefaultMemcachedResourceRequirements()
	} else {
		m.memcachedOptions.ResourceRequirements = v1.ResourceRequirements{}
	}

	// DeploymentConfig-level ResourceRequirements CR fields have priority over
	// spec.resourceRequirementsEnabled, overwriting that setting when they are
	// defined
	if m.apimanager.Spec.System.MemcachedResources != nil {
		m.memcachedOptions.ResourceRequirements = *m.apimanager.Spec.System.MemcachedResources
	}
}

func (m *MemcachedOptionsProvider) setNodeAffinityAndTolerationsOptions() {
	m.memcachedOptions.Affinity = m.apimanager.Spec.System.MemcachedAffinity
	m.memcachedOptions.Tolerations = m.apimanager.Spec.System.MemcachedTolerations
}

func (m *MemcachedOptionsProvider) deploymentLabels() map[string]string {
	return map[string]string{
		"app":                          *m.apimanager.Spec.AppLabel,
		"threescale_component":         "system",
		"threescale_component_element": "memcache",
	}
}

func (m *MemcachedOptionsProvider) podTemplateLabels(image string) map[string]string {
	labels := helper.MeteringLabels("system-memcache", helper.ParseVersion(image), helper.ApplicationType)

	for k, v := range m.deploymentLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "system-memcache"

	return labels
}
