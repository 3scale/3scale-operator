package operator

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"path"
	"sort"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/3scale/3scale-operator/apis/apps"
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
)

type ApicastOptionsProvider struct {
	apimanager     *appsv1alpha1.APIManager
	apicastOptions *component.ApicastOptions
	client         client.Client
	secretSource   *helper.SecretSource
}

const (
	APIcastEnvironmentCMAnnotation             = "apps.3scale.net/env-configmap-hash"
	PodPrioritySystemNodeCritical              = "system-node-critical"
	CustomPoliciesSecretResverAnnotationPrefix = "apimanager.apps.3scale.net/custompolicy-secret-resource-version-"
	OpentelemetrySecretResverAnnotationPrefix  = "apimanager.apps.3scale.net/opentelemtry-secret-resource-version-"
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
	a.setTopologySpreadConstraints()
	a.setPodTemplateAnnotations()

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

	// Retrieve opentelemtry staging configuration
	stagingOtelConfig, err := a.getOpenTelemetryStagingConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	a.apicastOptions.StagingOpentelemetry = stagingOtelConfig

	// Retrieve opentelemtry production configuration
	productionOtelConfig, err := a.getOpenTelemetryProductionConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	a.apicastOptions.ProductionOpentelemetry = productionOtelConfig

	a.setProxyConfigurations()

	// Pod Annotations. Used to rollout apicast deployment if any secrets/configmap changes
	a.apicastOptions.StagingAdditionalPodAnnotations = a.stagingAdditionalPodAnnotations()
	a.apicastOptions.ProductionAdditionalPodAnnotations = a.productionAdditionalPodAnnotations()

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

	// Deployment-level ResourceRequirements CR fields have priority over
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

	for k, v := range a.apimanager.Spec.Apicast.StagingSpec.Labels {
		labels[k] = v
	}

	labels[reconcilers.DeploymentLabelSelector] = "apicast-staging"

	return labels
}

func (a *ApicastOptionsProvider) productionPodTemplateLabels() map[string]string {
	labels := helper.MeteringLabels("apicast-production", helper.ApplicationType)

	for k, v := range a.commonProductionLabels() {
		labels[k] = v
	}

	for k, v := range a.apimanager.Spec.Apicast.ProductionSpec.Labels {
		labels[k] = v
	}

	labels[reconcilers.DeploymentLabelSelector] = "apicast-production"

	return labels
}

func (a *ApicastOptionsProvider) setCustomPolicies() error {
	for idx, customPolicySpec := range a.apimanager.Spec.Apicast.ProductionSpec.CustomPolicies {
		// CR Validation ensures secret name is not nil

		namespacedName := types.NamespacedName{
			Name:      customPolicySpec.SecretRef.Name, // CR Validation ensures not nil
			Namespace: a.apimanager.Namespace,
		}

		secret, err := a.validateCustomPolicySecret(context.TODO(), customPolicySpec.SecretRef.Name, namespacedName)
		if err != nil {
			fldErr := field.ErrorList{}
			customPoliciesIdxFldPath := field.NewPath("spec").
				Child("apicast").
				Child("productionSpec").
				Child("customPolicies").Index(idx)
			fldErr = append(fldErr, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, err.Error()))
			return fldErr.ToAggregate()
		}

		a.apicastOptions.ProductionCustomPolicies = append(a.apicastOptions.ProductionCustomPolicies, component.CustomPolicy{
			Name:    customPolicySpec.Name,
			Version: customPolicySpec.Version,
			Secret:  secret,
		})
	}

	// TODO(eastizle): DRY!!
	for idx, customPolicySpec := range a.apimanager.Spec.Apicast.StagingSpec.CustomPolicies {
		// CR Validation ensures secret name is not nil

		namespacedName := types.NamespacedName{
			Name:      customPolicySpec.SecretRef.Name, // CR Validation ensures not nil
			Namespace: a.apimanager.Namespace,
		}

		secret, err := a.validateCustomPolicySecret(context.TODO(), customPolicySpec.SecretRef.Name, namespacedName)
		if err != nil {
			fldErr := field.ErrorList{}
			customPoliciesIdxFldPath := field.NewPath("spec").
				Child("apicast").
				Child("stagingSpec").
				Child("customPolicies").Index(idx)
			fldErr = append(fldErr, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, err.Error()))
			return fldErr.ToAggregate()
		}

		a.apicastOptions.StagingCustomPolicies = append(a.apicastOptions.StagingCustomPolicies, component.CustomPolicy{
			Name:    customPolicySpec.Name,
			Version: customPolicySpec.Version,
			Secret:  secret,
		})
	}

	return nil
}

func (a *ApicastOptionsProvider) validateCustomPolicySecret(ctx context.Context, name string, nn types.NamespacedName) (*v1.Secret, error) {
	secret := &v1.Secret{}
	err := a.client.Get(ctx, nn, secret)

	if err != nil {
		// NotFoundError is also an error, it is required to exist
		return nil, err
	}

	_, err = a.secretSource.RequiredFieldValueFromRequiredSecret(name, "init.lua")
	if err != nil {
		return nil, err
	}

	_, err = a.secretSource.RequiredFieldValueFromRequiredSecret(name, "apicast-policy.json")
	if err != nil {
		return nil, err
	}

	return secret, nil
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
		TracingLibrary: apps.APIcastDefaultTracingLibrary,
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
				fldErr := field.ErrorList{}
				tracingConfigFldPath := field.NewPath("spec").Child("openTracing").Child("tracingConfigSecretRef")
				fldErr = append(fldErr, field.Invalid(tracingConfigFldPath, openTracingConfigSpec, err.Error()))
				return fldErr.ToAggregate()
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
		TracingLibrary: apps.APIcastDefaultTracingLibrary,
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
				fldErr := field.ErrorList{}
				tracingConfigFldPath := field.NewPath("spec").Child("openTracing").Child("tracingConfigSecretRef")
				fldErr = append(fldErr, field.Invalid(tracingConfigFldPath, openTracingConfigSpec, err.Error()))
				return fldErr.ToAggregate()
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
		return fmt.Errorf("required secret key, %s not found", "config")
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

func (a *ApicastOptionsProvider) stagingAdditionalPodAnnotations() map[string]string {
	annotations := map[string]string{
		APIcastEnvironmentCMAnnotation: a.envConfigMapHash(),
	}

	for idx := range a.apicastOptions.StagingCustomPolicies {
		// Secrets must exist
		// Annotation key includes the name of the secret
		annotationKey := fmt.Sprintf("%s%s", CustomPoliciesSecretResverAnnotationPrefix, a.apicastOptions.StagingCustomPolicies[idx].Secret.Name)
		annotations[annotationKey] = a.apicastOptions.StagingCustomPolicies[idx].Secret.ResourceVersion
	}

	if a.apimanager.OpenTelemetryEnabledForStaging() && a.isOpentelemetryPodAnnotationRequired(&a.apicastOptions.StagingOpentelemetry.Secret) {
		if a.apicastOptions.StagingOpentelemetry.Secret.Name != "" {
			annotationKey := fmt.Sprintf("%s%s", OpentelemetrySecretResverAnnotationPrefix, a.apicastOptions.StagingOpentelemetry.Secret.Name)
			annotations[annotationKey] = a.apicastOptions.StagingOpentelemetry.Secret.ResourceVersion
		}
	}

	return annotations
}

func (a *ApicastOptionsProvider) productionAdditionalPodAnnotations() map[string]string {
	annotations := map[string]string{
		APIcastEnvironmentCMAnnotation: a.envConfigMapHash(),
	}

	for idx := range a.apicastOptions.ProductionCustomPolicies {
		// Secrets must exist
		// Annotation key includes the name of the secret
		annotationKey := fmt.Sprintf("%s%s", CustomPoliciesSecretResverAnnotationPrefix, a.apicastOptions.ProductionCustomPolicies[idx].Secret.Name)
		annotations[annotationKey] = a.apicastOptions.ProductionCustomPolicies[idx].Secret.ResourceVersion
	}

	if a.apimanager.OpenTelemetryEnabledForProduction() && a.isOpentelemetryPodAnnotationRequired(&a.apicastOptions.ProductionOpentelemetry.Secret) {
		if a.apicastOptions.ProductionOpentelemetry.Secret.Name != "" {
			annotationKey := fmt.Sprintf("%s%s", OpentelemetrySecretResverAnnotationPrefix, a.apicastOptions.ProductionOpentelemetry.Secret.Name)
			annotations[annotationKey] = a.apicastOptions.ProductionOpentelemetry.Secret.ResourceVersion
		}
	}

	return annotations
}

func (a *ApicastOptionsProvider) isOpentelemetryPodAnnotationRequired(secret *v1.Secret) bool {
	existingLabels := secret.Labels

	if existingLabels != nil {
		if _, ok := existingLabels["apimanager.apps.3scale.net/watched-by"]; ok {
			return true
		}
	}

	return false
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
	if a.apimanager.Spec.Apicast.StagingSpec.PriorityClassName != nil {
		a.apicastOptions.PriorityClassNameStaging = *a.apimanager.Spec.Apicast.StagingSpec.PriorityClassName
	}
	if a.apimanager.Spec.Apicast.ProductionSpec.PriorityClassName != nil {
		a.apicastOptions.PriorityClassNameProduction = *a.apimanager.Spec.Apicast.ProductionSpec.PriorityClassName
	}
}

func (a *ApicastOptionsProvider) setTopologySpreadConstraints() {
	if a.apimanager.Spec.Apicast.StagingSpec.TopologySpreadConstraints != nil {
		a.apicastOptions.TopologySpreadConstraintsStaging = a.apimanager.Spec.Apicast.StagingSpec.TopologySpreadConstraints
	}
	if a.apimanager.Spec.Apicast.ProductionSpec.TopologySpreadConstraints != nil {
		a.apicastOptions.TopologySpreadConstraintsProduction = a.apimanager.Spec.Apicast.ProductionSpec.TopologySpreadConstraints
	}
}

func (a *ApicastOptionsProvider) setPodTemplateAnnotations() {
	a.apicastOptions.StagingPodTemplateAnnotations = a.apimanager.Spec.Apicast.StagingSpec.Annotations
	a.apicastOptions.ProductionPodTemplateAnnotations = a.apimanager.Spec.Apicast.ProductionSpec.Annotations
}

func (a *ApicastOptionsProvider) getOpenTelemetryStagingConfig(ctx context.Context) (component.OpentelemetryConfig, error) {
	res := component.OpentelemetryConfig{
		Enabled: false,
	}

	if a.apimanager.Spec.Apicast.StagingSpec.OpenTelemetry != nil && a.apimanager.Spec.Apicast.StagingSpec.OpenTelemetry.Enabled != nil {
		res = component.OpentelemetryConfig{
			Enabled: *a.apimanager.Spec.Apicast.StagingSpec.OpenTelemetry.Enabled,
		}
	} else {
		return res, nil
	}

	if !res.Enabled {
		return res, nil
	}

	// In the APIcast CR validation step it is checked that when enabled, the secret ref is not nil
	// Adding this to avoid panics
	if a.apimanager.Spec.Apicast.StagingSpec.OpenTelemetry.TracingConfigSecretRef == nil || a.apimanager.Spec.Apicast.StagingSpec.OpenTelemetry.TracingConfigSecretRef.Name == "" {
		fldPath := field.NewPath("spec").Child("apicast").Child("stagingSpec").Child("openTelemetry").Child("tracingConfigSecretRef")
		err := append(field.ErrorList{}, field.Invalid(fldPath, a.apimanager.Spec.Apicast.StagingSpec.OpenTelemetry.TracingConfigSecretRef, "tracing config secret name is empty"))
		return res, err.ToAggregate()
	}

	// Read secret and get first key in lexicographical order.
	// Defining some order is required because maps do not provide order semantics and
	// key consistency is required accross reconcile loops
	otelSecretKey := client.ObjectKey{
		Name:      a.apimanager.Spec.Apicast.StagingSpec.OpenTelemetry.TracingConfigSecretRef.Name,
		Namespace: a.apimanager.Namespace,
	}

	secret := &v1.Secret{}
	err := a.client.Get(ctx, otelSecretKey, secret)
	if err != nil {
		// NotFoundError is also an error, it is required to exist
		fldPath := field.NewPath("spec").Child("apicast").Child("stagingSpec").Child("openTelemetry").Child("tracingConfigSecretRef")
		err := append(field.ErrorList{}, field.Invalid(fldPath, a.apimanager.Spec.Apicast.StagingSpec.OpenTelemetry.TracingConfigSecretRef, err.Error()))
		return res, err.ToAggregate()
	}

	res.Secret = *secret

	secretKeys := helper.MapKeys(helper.GetSecretStringDataFromData(secret.Data))
	if len(secretKeys) == 0 {
		fldPath := field.NewPath("spec").Child("apicast").Child("stagingSpec").Child("openTelemetry").Child("tracingConfigSecretRef")
		err := append(field.ErrorList{}, field.Invalid(fldPath, a.apimanager.Spec.Apicast.StagingSpec.OpenTelemetry.TracingConfigSecretRef, "secret is empty, no key found"))
		return res, err.ToAggregate()
	}

	if a.apimanager.Spec.Apicast.StagingSpec.OpenTelemetry.TracingConfigSecretKey != nil &&
		*a.apimanager.Spec.Apicast.StagingSpec.OpenTelemetry.TracingConfigSecretKey != "" {
		res.ConfigFile = path.Join(component.OpentelemetryConfigMountBasePath, *a.apimanager.Spec.Apicast.StagingSpec.OpenTelemetry.TracingConfigSecretKey)
		return res, nil
	}
	sort.Strings(secretKeys)

	res.ConfigFile = path.Join(component.OpentelemetryConfigMountBasePath, secretKeys[0])

	return res, nil
}

func (a *ApicastOptionsProvider) getOpenTelemetryProductionConfig(ctx context.Context) (component.OpentelemetryConfig, error) {
	res := component.OpentelemetryConfig{
		Enabled: false,
	}

	if a.apimanager.Spec.Apicast.ProductionSpec.OpenTelemetry != nil && a.apimanager.Spec.Apicast.ProductionSpec.OpenTelemetry.Enabled != nil {
		res = component.OpentelemetryConfig{
			Enabled: *a.apimanager.Spec.Apicast.ProductionSpec.OpenTelemetry.Enabled,
		}
	} else {
		return res, nil
	}

	if !res.Enabled {
		return res, nil
	}

	// In the APIcast CR validation step it is checked that when enabled, the secret ref is not nil
	// Adding this to avoid panics
	if a.apimanager.Spec.Apicast.ProductionSpec.OpenTelemetry.TracingConfigSecretRef == nil || a.apimanager.Spec.Apicast.ProductionSpec.OpenTelemetry.TracingConfigSecretRef.Name == "" {
		fldPath := field.NewPath("spec").Child("apicast").Child("productionSpec").Child("openTelemetry").Child("tracingConfigSecretRef")
		err := append(field.ErrorList{}, field.Invalid(fldPath, a.apimanager.Spec.Apicast.ProductionSpec.OpenTelemetry.TracingConfigSecretRef, "tracing config secret name is empty"))
		return res, err.ToAggregate()
	}

	// Read secret and get first key in lexicographical order.
	// Defining some order is required because maps do not provide order semantics and
	// key consistency is required accross reconcile loops
	otelSecretKey := client.ObjectKey{
		Name:      a.apimanager.Spec.Apicast.ProductionSpec.OpenTelemetry.TracingConfigSecretRef.Name,
		Namespace: a.apimanager.Namespace,
	}

	secret := &v1.Secret{}
	err := a.client.Get(ctx, otelSecretKey, secret)
	if err != nil {
		// NotFoundError is also an error, it is required to exist
		fldPath := field.NewPath("spec").Child("apicast").Child("productionSpec").Child("openTelemetry").Child("tracingConfigSecretRef")
		err := append(field.ErrorList{}, field.Invalid(fldPath, a.apimanager.Spec.Apicast.ProductionSpec.OpenTelemetry.TracingConfigSecretRef, err.Error()))
		return res, err.ToAggregate()
	}

	res.Secret = *secret

	secretKeys := helper.MapKeys(helper.GetSecretStringDataFromData(secret.Data))
	if len(secretKeys) == 0 {
		fldPath := field.NewPath("spec").Child("apicast").Child("productionSpec").Child("openTelemetry").Child("tracingConfigSecretRef")
		err := append(field.ErrorList{}, field.Invalid(fldPath, a.apimanager.Spec.Apicast.ProductionSpec.OpenTelemetry.TracingConfigSecretRef, "secret is empty, no key found"))
		return res, err.ToAggregate()
	}

	sort.Strings(secretKeys)

	if a.apimanager.Spec.Apicast.ProductionSpec.OpenTelemetry.TracingConfigSecretKey != nil &&
		*a.apimanager.Spec.Apicast.ProductionSpec.OpenTelemetry.TracingConfigSecretKey != "" {
		res.ConfigFile = path.Join(component.OpentelemetryConfigMountBasePath, *a.apimanager.Spec.Apicast.ProductionSpec.OpenTelemetry.TracingConfigSecretKey)
		return res, nil
	}

	res.ConfigFile = path.Join(component.OpentelemetryConfigMountBasePath, secretKeys[0])

	return res, nil
}
