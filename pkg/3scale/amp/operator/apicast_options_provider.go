package operator

import (
	"strconv"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
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
	a.apicastOptions.AppLabel = *a.apimanager.Spec.AppLabel
	a.apicastOptions.TenantName = *a.apimanager.Spec.TenantName
	a.apicastOptions.WildcardDomain = a.apimanager.Spec.WildcardDomain
	a.apicastOptions.ManagementAPI = *a.apimanager.Spec.Apicast.ApicastManagementAPI
	a.apicastOptions.ImageTag = product.ThreescaleRelease
	a.apicastOptions.OpenSSLVerify = strconv.FormatBool(*a.apimanager.Spec.Apicast.OpenSSLVerify)
	a.apicastOptions.ResponseCodes = strconv.FormatBool(*a.apimanager.Spec.Apicast.IncludeResponseCodes)

	a.setResourceRequirementsOptions()
	a.setReplicas()

	err := a.apicastOptions.Validate()
	return a.apicastOptions, err
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

func (a *ApicastOptionsProvider) setReplicas() {
	a.apicastOptions.ProductionReplicas = int32(*a.apimanager.Spec.Apicast.ProductionSpec.Replicas)
	a.apicastOptions.StagingReplicas = int32(*a.apimanager.Spec.Apicast.StagingSpec.Replicas)
}
