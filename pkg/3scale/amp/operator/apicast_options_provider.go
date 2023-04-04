package operator

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
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

const (
	APIcastEnvironmentCMAnnotation = "apps.3scale.net/env-configmap-hash"
	PodPrioritySystemNodeCritical  = "system-node-critical"
)

func NewApicastOptionsProvider(apimanager *appsv1alpha1.APIManager, client client.Client) *ApicastOptionsProvider {
	return &ApicastOptionsProvider{
		apimanager:     apimanager,
		apicastOptions: component.NewApicastOptions(),
		client:         client,
		secretSource:   helper.NewSecretSource(client, apimanager.Namespace),
	}
}

func (a *ApicastOptionsProvider) GetApicastOptions() (*component.ApicastOptions, error) {
	a.apicastOptions.ManagementAPI = *a.apimanager.Spec.Apicast.ApicastManagementAPI
	a.apicastOptions.ImageTag = product.ThreescaleRelease
	a.apicastOptions.OpenSSLVerify = strconv.FormatBool(*a.apimanager.Spec.Apicast.OpenSSLVerify)
	a.apicastOptions.ResponseCodes = strconv.FormatBool(*a.apimanager.Spec.Apicast.IncludeResponseCodes)
	a.apicastOptions.ExtendedMetrics = true
	a.apicastOptions.CommonLabels = a.commonLabels()
	a.apicastOptions.CommonStagingLabels = a.commonStagingLabels()
	a.apicastOptions.CommonProductionLabels = a.commonProductionLabels()
	a.apicastOptions.StagingPodTemplateLabels = a.stagingPodTemplateLabels()
	a.apicastOptions.ProductionPodTemplateLabels = a.productionPodTemplateLabels()
	a.apicastOptions.Namespace = a.apimanager.Namespace
	a.apicastOptions.ProductionWorkers = a.apimanager.Spec.Apicast.ProductionSpec.Workers
	a.apicastOptions.ProductionLogLevel = a.apimanager.Spec.Apicast.ProductionSpec.LogLevel
	a.apicastOptions.StagingLogLevel = a.apimanager.Spec.Apicast.StagingSpec.LogLevel

	a.apicastOptions.ProductionHTTPSPort = a.apimanager.Spec.Apicast.ProductionSpec.HTTPSPort
	a.apicastOptions.ProductionHTTPSVerifyDepth = a.apimanager.Spec.Apicast.ProductionSpec.HTTPSVerifyDepth
	// when HTTPS certificate is provided and HTTPS port is not provided, assing default https port
	if a.apimanager.Spec.Apicast.ProductionSpec.HTTPSCertificateSecretRef != nil && a.apimanager.Spec.Apicast.ProductionSpec.HTTPSPort == nil {
		tmpDefaultPort := appsv1alpha1.DefaultHTTPSPort
		a.apicastOptions.ProductionHTTPSPort = &tmpDefaultPort
	}
	// when HTTPS port is provided and HTTPS Certificate secret is not provided,
	// Apicast will use some default certificate
	// Should the operator raise a warning?
	if a.apimanager.Spec.Apicast.ProductionSpec.HTTPSCertificateSecretRef != nil {
		a.apicastOptions.ProductionHTTPSCertificateSecretName = &a.apimanager.Spec.Apicast.ProductionSpec.HTTPSCertificateSecretRef.Name
	}

	a.apicastOptions.StagingHTTPSPort = a.apimanager.Spec.Apicast.StagingSpec.HTTPSPort
	a.apicastOptions.StagingHTTPSVerifyDepth = a.apimanager.Spec.Apicast.StagingSpec.HTTPSVerifyDepth
	// when HTTPS certificate is provided and HTTPS port is not provided, assing default https port
	if a.apimanager.Spec.Apicast.StagingSpec.HTTPSCertificateSecretRef != nil && a.apimanager.Spec.Apicast.StagingSpec.HTTPSPort == nil {
		tmpDefaultPort := appsv1alpha1.DefaultHTTPSPort
		a.apicastOptions.StagingHTTPSPort = &tmpDefaultPort
	}
	if a.apimanager.Spec.Apicast.StagingSpec.HTTPSCertificateSecretRef != nil {
		a.apicastOptions.StagingHTTPSCertificateSecretName = &a.apimanager.Spec.Apicast.StagingSpec.HTTPSCertificateSecretRef.Name
	}

	a.apicastOptions.ProductionServiceCacheSize = a.apimanager.Spec.Apicast.ProductionSpec.ServiceCacheSize
	a.apicastOptions.StagingServiceCacheSize = a.apimanager.Spec.Apicast.StagingSpec.ServiceCacheSize

	a.setResourceRequirementsOptions()
	a.setNodeAffinityAndTolerationsOptions()
	a.setReplicas()
	a.setPriorityClassNames()

	err := a.setCustomPolicies()
	if err != nil {
		return nil, err
	}

	err = a.setTracingConfiguration()
	if err != nil {
		return nil, err
	}

	err = a.setCustomEnvironments()
	if err != nil {
		return nil, err
	}

	a.setProxyConfigurations()

	// Pod Annotations. Used to rollout apicast deployment if any secrets/configmap changes
	a.apicastOptions.AdditionalPodAnnotations = a.additionalPodAnnotations()

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
	a.apicastOptions.ProductionReplicas = 1
	if a.apimanager.Spec.Apicast.ProductionSpec.Replicas != nil {
		a.apicastOptions.ProductionReplicas = int32(*a.apimanager.Spec.Apicast.ProductionSpec.Replicas)
	}

	a.apicastOptions.StagingReplicas = 1
	if a.apimanager.Spec.Apicast.StagingSpec.Replicas != nil {
		a.apicastOptions.StagingReplicas = int32(*a.apimanager.Spec.Apicast.StagingSpec.Replicas)
	}
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

func (a *ApicastOptionsProvider) stagingPodTemplateLabels() map[string]string {
	labels := helper.MeteringLabels("apicast-staging", helper.ApplicationType)

	for k, v := range a.commonStagingLabels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "apicast-staging"

	return labels
}

func (a *ApicastOptionsProvider) productionPodTemplateLabels() map[string]string {
	labels := helper.MeteringLabels("apicast-production", helper.ApplicationType)

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

func (a *ApicastOptionsProvider) setTracingConfiguration() error {
	err := a.setProductionTracingConfiguration()
	if err != nil {
		return err
	}

	err = a.setStagingTracingConfiguration()
	if err != nil {
		return err
	}

	return nil
}

func (a *ApicastOptionsProvider) setProductionTracingConfiguration() error {
	tracingIsEnabled := a.apimanager.IsAPIcastProductionOpenTracingEnabled()
	res := &component.APIcastTracingConfig{
		Enabled:        tracingIsEnabled,
		TracingLibrary: component.APIcastDefaultTracingLibrary,
	}
	if tracingIsEnabled {
		openTracingConfigSpec := a.apimanager.Spec.Apicast.ProductionSpec.OpenTracing
		if openTracingConfigSpec.TracingLibrary != nil {
			res.TracingLibrary = *a.apimanager.Spec.Apicast.ProductionSpec.OpenTracing.TracingLibrary
		}
		if openTracingConfigSpec.TracingConfigSecretRef != nil {
			namespacedName := types.NamespacedName{
				Name:      openTracingConfigSpec.TracingConfigSecretRef.Name, // CR Validation ensures not nil
				Namespace: a.apimanager.Namespace,
			}
			err := a.validateTracingConfigSecret(namespacedName)
			if err != nil {
				errors := field.ErrorList{}
				tracingConfigFldPath := field.NewPath("spec").Child("openTracing").Child("tracingConfigSecretRef")
				errors = append(errors, field.Invalid(tracingConfigFldPath, openTracingConfigSpec, err.Error()))
				return errors.ToAggregate()
			}
			res.TracingConfigSecretName = &openTracingConfigSpec.TracingConfigSecretRef.Name
		}
	}

	a.apicastOptions.ProductionTracingConfig = res

	return nil
}

func (a *ApicastOptionsProvider) setStagingTracingConfiguration() error {
	tracingIsEnabled := a.apimanager.IsAPIcastStagingOpenTracingEnabled()
	res := &component.APIcastTracingConfig{
		Enabled:        tracingIsEnabled,
		TracingLibrary: component.APIcastDefaultTracingLibrary,
	}
	if tracingIsEnabled {
		openTracingConfigSpec := a.apimanager.Spec.Apicast.StagingSpec.OpenTracing
		if openTracingConfigSpec.TracingLibrary != nil {
			res.TracingLibrary = *a.apimanager.Spec.Apicast.StagingSpec.OpenTracing.TracingLibrary
		}
		if openTracingConfigSpec.TracingConfigSecretRef != nil {
			namespacedName := types.NamespacedName{
				Name:      openTracingConfigSpec.TracingConfigSecretRef.Name, // CR Validation ensures not nil
				Namespace: a.apimanager.Namespace,
			}
			err := a.validateTracingConfigSecret(namespacedName)
			if err != nil {
				errors := field.ErrorList{}
				tracingConfigFldPath := field.NewPath("spec").Child("openTracing").Child("tracingConfigSecretRef")
				errors = append(errors, field.Invalid(tracingConfigFldPath, openTracingConfigSpec, err.Error()))
				return errors.ToAggregate()
			}
			res.TracingConfigSecretName = &openTracingConfigSpec.TracingConfigSecretRef.Name
		}
	}

	a.apicastOptions.StagingTracingConfig = res

	return nil
}

func (a *ApicastOptionsProvider) validateTracingConfigSecret(nn types.NamespacedName) error {
	secret := &v1.Secret{}
	err := a.client.Get(context.TODO(), nn, secret)

	if err != nil {
		// NotFoundError is also an error, it is required to exist
		return err
	}

	if _, ok := secret.Data[component.APIcastTracingConfigSecretKey]; !ok {
		return fmt.Errorf("Required secret key, %s not found", "config")
	}

	return nil
}

func (a *ApicastOptionsProvider) setCustomEnvironments() error {
	for idx, customEnvSpec := range a.apimanager.Spec.Apicast.ProductionSpec.CustomEnvironments {
		// CR Validation ensures secret name is not nil
		namespacedName := types.NamespacedName{
			Name:      customEnvSpec.SecretRef.Name,
			Namespace: a.apimanager.Namespace,
		}

		secret, err := a.customEnvironmentSecret(namespacedName)
		if err != nil {
			fieldErrors := field.ErrorList{}
			customEnvIdxFldPath := field.NewPath("spec").
				Child("apicast").
				Child("productionSpec").
				Child("customEnvironments").Index(idx)
			fieldErrors = append(fieldErrors, field.Invalid(customEnvIdxFldPath, customEnvSpec, err.Error()))
			return fieldErrors.ToAggregate()
		}

		a.apicastOptions.ProductionCustomEnvironments = append(a.apicastOptions.ProductionCustomEnvironments, secret)
	}

	// TODO(eastizle): DRY!!
	for idx, customEnvSpec := range a.apimanager.Spec.Apicast.StagingSpec.CustomEnvironments {
		// CR Validation ensures secret name is not nil
		namespacedName := types.NamespacedName{
			Name:      customEnvSpec.SecretRef.Name,
			Namespace: a.apimanager.Namespace,
		}

		secret, err := a.customEnvironmentSecret(namespacedName)
		if err != nil {
			fieldErrors := field.ErrorList{}
			customEnvIdxFldPath := field.NewPath("spec").
				Child("apicast").
				Child("stagingSpec").
				Child("customEnvironments").Index(idx)
			fieldErrors = append(fieldErrors, field.Invalid(customEnvIdxFldPath, customEnvSpec, err.Error()))
			return fieldErrors.ToAggregate()
		}

		a.apicastOptions.StagingCustomEnvironments = append(a.apicastOptions.StagingCustomEnvironments, secret)
	}

	return nil
}

func (a *ApicastOptionsProvider) customEnvironmentSecret(nn types.NamespacedName) (*v1.Secret, error) {
	secret := &v1.Secret{}
	err := a.client.Get(context.TODO(), nn, secret)

	if err != nil {
		// NotFoundError is also an error, it is required to exist
		return nil, err
	}

	if len(secret.Data) == 0 {
		return nil, errors.New("empty secret")
	}

	return secret, nil
}

func (a *ApicastOptionsProvider) setProxyConfigurations() {
	a.setStagingProxyConfigurations()
	a.setProductionProxyConfigurations()
}

func (a *ApicastOptionsProvider) setStagingProxyConfigurations() {
	a.apicastOptions.StagingAllProxy = a.apimanager.Spec.Apicast.StagingSpec.AllProxy
	a.apicastOptions.StagingHTTPProxy = a.apimanager.Spec.Apicast.StagingSpec.HTTPProxy
	a.apicastOptions.StagingHTTPSProxy = a.apimanager.Spec.Apicast.StagingSpec.HTTPSProxy
	a.apicastOptions.StagingNoProxy = a.apimanager.Spec.Apicast.StagingSpec.NoProxy

}

func (a *ApicastOptionsProvider) setProductionProxyConfigurations() {
	a.apicastOptions.ProductionAllProxy = a.apimanager.Spec.Apicast.ProductionSpec.AllProxy
	a.apicastOptions.ProductionHTTPProxy = a.apimanager.Spec.Apicast.ProductionSpec.HTTPProxy
	a.apicastOptions.ProductionHTTPSProxy = a.apimanager.Spec.Apicast.ProductionSpec.HTTPSProxy
	a.apicastOptions.ProductionNoProxy = a.apimanager.Spec.Apicast.ProductionSpec.NoProxy
}

func (a *ApicastOptionsProvider) additionalPodAnnotations() map[string]string {
	annotations := map[string]string{
		APIcastEnvironmentCMAnnotation: a.envConfigMapHash(),
	}

	return annotations
}

// APIcast environment hash
// When any of the fields used to compute the hash change the value, the hash will change
// and the apicast deployment will rollout
func (a *ApicastOptionsProvider) envConfigMapHash() string {
	h := fnv.New32a()
	h.Write([]byte(a.apicastOptions.ManagementAPI))
	h.Write([]byte(a.apicastOptions.OpenSSLVerify))
	h.Write([]byte(a.apicastOptions.ResponseCodes))
	val := h.Sum32()
	return fmt.Sprint(val)
}

func (a *ApicastOptionsProvider) setPriorityClassNames() {
	//a.apicastOptions.PriorityClassNameStaging = PodPrioritySystemNodeCritical
	if a.apimanager.Spec.Apicast.StagingSpec.PriorityClassName != nil {
		a.apicastOptions.PriorityClassNameStaging = *a.apimanager.Spec.Apicast.StagingSpec.PriorityClassName
	}

	//a.apicastOptions.PriorityClassNameProduction = PodPrioritySystemNodeCritical
	if a.apimanager.Spec.Apicast.ProductionSpec.PriorityClassName != nil {
		a.apicastOptions.PriorityClassNameProduction = *a.apimanager.Spec.Apicast.ProductionSpec.PriorityClassName
	}
}
