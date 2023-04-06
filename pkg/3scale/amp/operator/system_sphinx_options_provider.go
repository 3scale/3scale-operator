package operator

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/helper"
)

type SystemSphinxOptionsProvider struct {
	apimanager *appsv1alpha1.APIManager
	options    *component.SystemSphinxOptions
}

func NewSystemSphinxOptionsProvider(apimanager *appsv1alpha1.APIManager) *SystemSphinxOptionsProvider {
	return &SystemSphinxOptionsProvider{
		apimanager: apimanager,
		options:    component.NewSystemSphinxOptions(),
	}
}

func (s *SystemSphinxOptionsProvider) GetOptions() (*component.SystemSphinxOptions, error) {
	s.options.ImageTag = product.ThreescaleRelease
	s.options.Labels = s.labels()
	s.options.PodTemplateLabels = s.podTemplateLabels()
	s.setResourceRequirementsOptions()
	s.setNodeAffinityAndTolerationsOptions()
	s.setPVCOptions()

	err := s.options.Validate()
	if err != nil {
		return nil, fmt.Errorf("GetSystemOptions validating: %w", err)
	}

	return s.options, nil
}

func (s *SystemSphinxOptionsProvider) setResourceRequirementsOptions() {
	s.options.ContainerResourceRequirements = v1.ResourceRequirements{}
	if *s.apimanager.Spec.ResourceRequirementsEnabled {
		s.options.ContainerResourceRequirements = component.DefaultSphinxContainerResourceRequirements()
	}
	// DeploymentConfig-level ResourceRequirements CR fields have priority over
	// spec.resourceRequirementsEnabled, overwriting that setting when they are
	// defined
	if s.apimanager.Spec.System.SphinxSpec.Resources != nil {
		s.options.ContainerResourceRequirements = *s.apimanager.Spec.System.SphinxSpec.Resources
	}
}

func (s *SystemSphinxOptionsProvider) setNodeAffinityAndTolerationsOptions() {
	s.options.Affinity = s.apimanager.Spec.System.SphinxSpec.Affinity
	s.options.Tolerations = s.apimanager.Spec.System.SphinxSpec.Tolerations
}

func (s *SystemSphinxOptionsProvider) commonLabels() map[string]string {
	return map[string]string{
		"app":                  *s.apimanager.Spec.AppLabel,
		"threescale_component": "system",
	}
}

func (s *SystemSphinxOptionsProvider) labels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "sphinx"
	return labels
}

func (s *SystemSphinxOptionsProvider) podTemplateLabels() map[string]string {
	labels := helper.MeteringLabels("system-sphinx", helper.ApplicationType)

	for k, v := range s.labels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "system-sphinx"

	return labels
}

func (s *SystemSphinxOptionsProvider) setPVCOptions() {
	// Default values
	s.options.PVCOptions = component.SphinxPVCOptions{
		StorageClass:    nil,
		VolumeName:      "",
		StorageRequests: resource.MustParse("1Gi"),
	}

	if s.apimanager.Spec.System != nil &&
		s.apimanager.Spec.System.SphinxSpec != nil &&
		s.apimanager.Spec.System.SphinxSpec.PVC != nil {
		if s.apimanager.Spec.System.SphinxSpec.PVC.StorageClassName != nil {
			s.options.PVCOptions.StorageClass = s.apimanager.Spec.System.SphinxSpec.PVC.StorageClassName
		}
		if s.apimanager.Spec.System.SphinxSpec.PVC.Resources != nil {
			s.options.PVCOptions.StorageRequests = s.apimanager.Spec.System.SphinxSpec.PVC.Resources.Requests
		}
		if s.apimanager.Spec.System.SphinxSpec.PVC.VolumeName != nil {
			s.options.PVCOptions.VolumeName = *s.apimanager.Spec.System.SphinxSpec.PVC.VolumeName
		}
	}
}
