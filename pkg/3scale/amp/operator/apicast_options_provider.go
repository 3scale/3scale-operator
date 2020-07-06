package operator

import (
	"fmt"
	"strconv"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/helper"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

type ApicastOptionsProvider struct {
	apimanager     *appsv1alpha1.APIManager
	apicastOptions *component.ApicastOptions
}

func NewApicastOptionsProvider(apimanager *appsv1alpha1.APIManager) *ApicastOptionsProvider {
	return &ApicastOptionsProvider{
		apimanager:     apimanager,
		apicastOptions: component.NewApicastOptions(),
	}
}

func (a *ApicastOptionsProvider) GetApicastOptions() (*component.ApicastOptions, error) {
	imageOpts, err := NewAmpImagesOptionsProvider(a.apimanager).GetAmpImagesOptions()
	if err != nil {
		return nil, fmt.Errorf("GetApicastOptions reading image options: %w", err)
	}

	a.apicastOptions.ManagementAPI = *a.apimanager.Spec.Apicast.ApicastManagementAPI
	a.apicastOptions.ImageTag = product.ThreescaleRelease
	a.apicastOptions.OpenSSLVerify = strconv.FormatBool(*a.apimanager.Spec.Apicast.OpenSSLVerify)
	a.apicastOptions.ResponseCodes = strconv.FormatBool(*a.apimanager.Spec.Apicast.IncludeResponseCodes)
	a.apicastOptions.ExtendedMetrics = true
	a.apicastOptions.CommonLabels = a.commonLabels()
	a.apicastOptions.CommonStagingLabels = a.commonStagingLabels()
	a.apicastOptions.CommonProductionLabels = a.commonProductionLabels()
	a.apicastOptions.StagingPodTemplateLabels = a.stagingPodTemplateLabels(imageOpts.ApicastImage)
	a.apicastOptions.ProductionPodTemplateLabels = a.productionPodTemplateLabels(imageOpts.ApicastImage)

	a.setResourceRequirementsOptions()
	a.setNodeAffinityAndTolerationsOptions()
	a.setReplicas()

	err = a.apicastOptions.Validate()
	if err != nil {
		return nil, fmt.Errorf("GetApicastOptions validating: %w", err)
	}
	return a.apicastOptions, nil
}

func (a *ApicastOptionsProvider) setResourceRequirementsOptions() {
	if *a.apimanager.Spec.ResourceRequirementsEnabled {
		a.apicastOptions.ProductionResourceRequirements = component.DefaultProductionResourceRequirements()
		a.apicastOptions.StagingResourceRequirements = component.DefaultStagingResourceRequirements()
	} else {
		a.apicastOptions.ProductionResourceRequirements = v1.ResourceRequirements{}
		a.apicastOptions.StagingResourceRequirements = v1.ResourceRequirements{}
	}
}

func (a *ApicastOptionsProvider) setNodeAffinityAndTolerationsOptions() {
	a.apicastOptions.StagingAffinity = a.apimanager.Spec.Apicast.StagingSpec.Affinity
	a.apicastOptions.StagingTolerations = a.apimanager.Spec.Apicast.StagingSpec.Tolerations
	a.apicastOptions.ProductionAffinity = a.apimanager.Spec.Apicast.ProductionSpec.Affinity
	a.apicastOptions.ProductionTolerations = a.apimanager.Spec.Apicast.ProductionSpec.Tolerations
}

func (a *ApicastOptionsProvider) setReplicas() {
	a.apicastOptions.ProductionReplicas = int32(*a.apimanager.Spec.Apicast.ProductionSpec.Replicas)
	a.apicastOptions.StagingReplicas = int32(*a.apimanager.Spec.Apicast.StagingSpec.Replicas)
}

func (a *ApicastOptionsProvider) commonLabels() map[string]string {
	return map[string]string{
		"app":                  *a.apimanager.Spec.AppLabel,
		"threescale_component": "apicast",
	}
}

func (a *ApicastOptionsProvider) commonStagingLabels() map[string]string {
	labels := a.commonLabels()
	labels["threescale_component_element"] = "staging"
	return labels
}

func (a *ApicastOptionsProvider) commonProductionLabels() map[string]string {
	labels := a.commonLabels()
	labels["threescale_component_element"] = "production"
	return labels
}

func (a *ApicastOptionsProvider) stagingPodTemplateLabels(image string) map[string]string {
	labels := helper.MeteringLabels("apicast-staging", helper.ParseVersion(image), helper.ApplicationType)

	for k, v := range a.commonStagingLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "apicast-staging"

	return labels
}

func (a *ApicastOptionsProvider) productionPodTemplateLabels(image string) map[string]string {
	labels := helper.MeteringLabels("apicast-production", helper.ParseVersion(image), helper.ApplicationType)

	for k, v := range a.commonProductionLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "apicast-production"

	return labels
}
