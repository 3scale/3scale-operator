package operator

import (
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/helper"
)

type ApicastOptionsProvider struct {
	apimanager     *appsv1alpha1.APIManager
	apicastOptions *component.ApicastOptions
	client         client.Client
	secretSource   *helper.SecretSource
}

func NewApicastOptionsProvider(apimanager *appsv1alpha1.APIManager, client client.Client) *ApicastOptionsProvider {
	return &ApicastOptionsProvider{
		apimanager:     apimanager,
		apicastOptions: component.NewApicastOptions(),
		client:         client,
		secretSource:   helper.NewSecretSource(client, apimanager.Namespace),
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
	a.apicastOptions.Namespace = a.apimanager.Namespace
	a.apicastOptions.ProductionWorkers = a.apimanager.Spec.Apicast.ProductionSpec.Workers
	a.apicastOptions.ProductionLogLevel = a.apimanager.Spec.Apicast.ProductionSpec.LogLevel
	a.apicastOptions.StagingLogLevel = a.apimanager.Spec.Apicast.StagingSpec.LogLevel

	a.setResourceRequirementsOptions()
	a.setNodeAffinityAndTolerationsOptions()
	a.setReplicas()

	err = a.setCustomPolicies()
	if err != nil {
		return nil, err
	}

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

	// DeploymentConfig-level ResourceRequirements CR fields have priority over
	// spec.resourceRequirementsEnabled, overwriting that setting when they are
	// defined
	if a.apimanager.Spec.Apicast.ProductionSpec.Resources != nil {
		a.apicastOptions.ProductionResourceRequirements = *a.apimanager.Spec.Apicast.ProductionSpec.Resources
	}

	if a.apimanager.Spec.Apicast.StagingSpec.Resources != nil {
		a.apicastOptions.StagingResourceRequirements = *a.apimanager.Spec.Apicast.StagingSpec.Resources
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

func (a *ApicastOptionsProvider) setCustomPolicies() error {
	for idx, customPolicySpec := range a.apimanager.Spec.Apicast.ProductionSpec.CustomPolicies {
		// CR Validation ensures secret name is not nil
		err := a.validateCustomPolicySecret(customPolicySpec.SecretRef.Name)
		if err != nil {
			errors := field.ErrorList{}
			customPoliciesIdxFldPath := field.NewPath("spec").
				Child("apicast").
				Child("productionSpec").
				Child("customPolicies").Index(idx)
			errors = append(errors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, err.Error()))
			return errors.ToAggregate()
		}

		a.apicastOptions.ProductionCustomPolicies = append(a.apicastOptions.ProductionCustomPolicies, component.CustomPolicy{
			Name:      customPolicySpec.Name,
			Version:   customPolicySpec.Version,
			SecretRef: *customPolicySpec.SecretRef,
		})
	}

	// TODO(eastizle): DRY!!
	for idx, customPolicySpec := range a.apimanager.Spec.Apicast.StagingSpec.CustomPolicies {
		// CR Validation ensures secret name is not nil
		err := a.validateCustomPolicySecret(customPolicySpec.SecretRef.Name)
		if err != nil {
			errors := field.ErrorList{}
			customPoliciesIdxFldPath := field.NewPath("spec").
				Child("apicast").
				Child("stagingSpec").
				Child("customPolicies").Index(idx)
			errors = append(errors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, err.Error()))
			return errors.ToAggregate()
		}

		a.apicastOptions.StagingCustomPolicies = append(a.apicastOptions.StagingCustomPolicies, component.CustomPolicy{
			Name:      customPolicySpec.Name,
			Version:   customPolicySpec.Version,
			SecretRef: *customPolicySpec.SecretRef,
		})
	}

	return nil
}

func (a *ApicastOptionsProvider) validateCustomPolicySecret(name string) error {
	_, err := a.secretSource.RequiredFieldValueFromRequiredSecret(name, "init.lua")
	if err != nil {
		return err
	}

	_, err = a.secretSource.RequiredFieldValueFromRequiredSecret(name, "apicast-policy.json")
	if err != nil {
		return err
	}

	return nil
}
