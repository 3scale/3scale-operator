package operator

import (
	"fmt"
	"reflect"
	"strings"

	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
)

func ApicastEnvCMMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*v1.ConfigMap)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.ConfigMap", existingObj)
	}
	desired, ok := desiredObj.(*v1.ConfigMap)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.ConfigMap", desiredObj)
	}

	update := false

	//	Check APICAST_MANAGEMENT_API
	fieldUpdated := reconcilers.ConfigMapReconcileField(desired, existing, "APICAST_MANAGEMENT_API")
	update = update || fieldUpdated

	//	Check OPENSSL_VERIFY
	fieldUpdated = reconcilers.ConfigMapReconcileField(desired, existing, "OPENSSL_VERIFY")
	update = update || fieldUpdated

	//	Check APICAST_RESPONSE_CODES
	fieldUpdated = reconcilers.ConfigMapReconcileField(desired, existing, "APICAST_RESPONSE_CODES")
	update = update || fieldUpdated

	return update, nil
}

type ApicastReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewApicastReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *ApicastReconciler {
	return &ApicastReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *ApicastReconciler) Reconcile() (reconcile.Result, error) {
	apicast, err := Apicast(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	stagingMutators := []reconcilers.DCMutateFn{
		reconcilers.DeploymentConfigImageChangeTriggerMutator,
		reconcilers.DeploymentConfigContainerResourcesMutator,
		reconcilers.DeploymentConfigAffinityMutator,
		reconcilers.DeploymentConfigTolerationsMutator,
		reconcilers.DeploymentConfigPodTemplateLabelsMutator,
		apicastLogLevelEnvVarMutator,
		apicastTracingConfigEnvVarsMutator,
		apicastEnvironmentEnvVarMutator,
		apicastHTTPSEnvVarMutator,
		apicastProxyConfigurationsEnvVarMutator,
		apicastServiceCacheSizeEnvVarMutator,
		apicastVolumeMountsMutator,
		apicastVolumesMutator,
		apicastCustomPolicyAnnotationsMutator,  // Should be always after volume mutator
		apicastTracingConfigAnnotationsMutator, // Should be always after volume mutator
		apicastCustomEnvAnnotationsMutator,     // Should be always after volume mutator
		portsMutator,
		apicastPodTemplateEnvConfigMapAnnotationsMutator,
		reconcilers.DeploymentConfigPriorityClassMutator,
	}

	if r.apiManager.Spec.Apicast.StagingSpec.Replicas != nil {
		stagingMutators = append(stagingMutators, reconcilers.DeploymentConfigReplicasMutator)
	}

	// Staging DC
	err = r.ReconcileDeploymentConfig(apicast.StagingDeploymentConfig(), reconcilers.DeploymentConfigMutator(stagingMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// add apicast production env var mutator
	productionMutators := []reconcilers.DCMutateFn{
		reconcilers.DeploymentConfigImageChangeTriggerMutator,
		reconcilers.DeploymentConfigContainerResourcesMutator,
		reconcilers.DeploymentConfigAffinityMutator,
		reconcilers.DeploymentConfigTolerationsMutator,
		reconcilers.DeploymentConfigPodTemplateLabelsMutator,
		apicastProductionWorkersEnvVarMutator,
		apicastLogLevelEnvVarMutator,
		apicastTracingConfigEnvVarsMutator,
		apicastEnvironmentEnvVarMutator,
		apicastHTTPSEnvVarMutator,
		apicastProxyConfigurationsEnvVarMutator,
		apicastServiceCacheSizeEnvVarMutator,
		apicastVolumeMountsMutator,
		apicastVolumesMutator,
		apicastCustomPolicyAnnotationsMutator,  // Should be always after volume mutator
		apicastTracingConfigAnnotationsMutator, // Should be always after volume mutator
		apicastCustomEnvAnnotationsMutator,     // Should be always after volume
		portsMutator,
		apicastPodTemplateEnvConfigMapAnnotationsMutator,
	}

	if r.apiManager.Spec.Apicast.ProductionSpec.Replicas != nil {
		productionMutators = append(productionMutators, reconcilers.DeploymentConfigReplicasMutator)
	}

	// Production DC
	productionDCMutator := reconcilers.DeploymentConfigMutator(
		productionMutators...,
	)

	err = r.ReconcileDeploymentConfig(apicast.ProductionDeploymentConfig(), productionDCMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Staging Service
	err = r.ReconcileService(apicast.StagingService(), getApiCastServiceMutator(r.apiManager.ObjectMeta.GetAnnotations()))
	if err != nil {
		return reconcile.Result{}, err
	}

	// Production Service
	err = r.ReconcileService(apicast.ProductionService(), getApiCastServiceMutator(r.apiManager.ObjectMeta.GetAnnotations()))
	if err != nil {
		return reconcile.Result{}, err
	}

	// Environment ConfigMap
	err = r.ReconcileConfigMap(apicast.EnvironmentConfigMap(), ApicastEnvCMMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Staging PDB
	err = r.ReconcilePodDisruptionBudget(apicast.StagingPodDisruptionBudget(), reconcilers.GenericPDBMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Production PDB
	err = r.ReconcilePodDisruptionBudget(apicast.ProductionPodDisruptionBudget(), reconcilers.GenericPDBMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	sumRate, err := helper.SumRateForOpenshiftVersion(r.Context(), r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileGrafanaDashboard(apicast.ApicastMainAppGrafanaDashboard(sumRate), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileGrafanaDashboard(apicast.ApicastServicesGrafanaDashboard(sumRate), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(apicast.ApicastPrometheusRules(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePodMonitor(apicast.ApicastProductionPodMonitor(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePodMonitor(apicast.ApicastStagingPodMonitor(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func getApiCastServiceMutator(apiManagerAnnotations map[string]string) reconcilers.MutateFn {
	disableApicastPortReconcile := "false"
	if apiManagerAnnotations == nil {
		apiManagerAnnotations = make(map[string]string)
	}
	if val, ok := apiManagerAnnotations["apps.3scale.net/disable-apicast-service-reconciler"]; ok {
		disableApicastPortReconcile = val
	}

	if disableApicastPortReconcile == "true" {
		return reconcilers.CreateOnlyMutator
	}
	return reconcilers.ServicePortMutator
}

func apicastProductionWorkersEnvVarMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	// Reconcile EnvVar only for "APICAST_WORKERS"
	return reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "APICAST_WORKERS"), nil
}

func apicastLogLevelEnvVarMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	// Reconcile EnvVar only for "APICAST_LOG_LEVEL"
	return reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "APICAST_LOG_LEVEL"), nil
}

func apicastTracingConfigEnvVarsMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	// Reconcile EnvVars related to opentracing
	var changed bool
	changed = reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "OPENTRACING_TRACER")

	tmpChanged := reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "OPENTRACING_CONFIG")
	changed = changed || tmpChanged

	return changed, nil
}

func apicastEnvironmentEnvVarMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	// Reconcile EnvVar only for "APICAST_ENVIRONMENT"
	return reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "APICAST_ENVIRONMENT"), nil
}

func apicastHTTPSEnvVarMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	// Reconcile EnvVars related to opentracing
	var changed bool

	for _, envVar := range []string{
		"APICAST_HTTPS_PORT",
		"APICAST_HTTPS_VERIFY_DEPTH",
		"APICAST_HTTPS_CERTIFICATE",
		"APICAST_HTTPS_CERTIFICATE_KEY",
	} {
		tmpChanged := reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, envVar)
		changed = changed || tmpChanged
	}

	return changed, nil
}

func apicastProxyConfigurationsEnvVarMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	// Reconcile EnvVars related to APIcast proxy-related configurations
	var changed bool

	for _, envVar := range []string{
		"ALL_PROXY",
		"HTTP_PROXY",
		"HTTPS_PROXY",
		"NO_PROXY",
	} {
		tmpChanged := reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, envVar)
		changed = changed || tmpChanged
	}

	return changed, nil
}

func apicastServiceCacheSizeEnvVarMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	return reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "APICAST_SERVICE_CACHE_SIZE"), nil
}

func portsMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	changed := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Containers[0].Ports, desired.Spec.Template.Spec.Containers[0].Ports) {
		changed = true
		existing.Spec.Template.Spec.Containers[0].Ports = desired.Spec.Template.Spec.Containers[0].Ports
	}

	return changed, nil
}

// volumeMountsMutator implements basic VolumeMount reconcilliation
// Added when in desired and not in existing
// Updated when in desired and in existing but not equal
// Existing not in desired will NOT be removed. Allows manually added volumemounts
func apicastVolumeMountsMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	changed := false
	existingContainer := &existing.Spec.Template.Spec.Containers[0]
	desiredContainer := &desired.Spec.Template.Spec.Containers[0]

	// Add desired not in existing
	for desiredIdx := range desiredContainer.VolumeMounts {
		existingIdx := helper.FindVolumeMountByMountPath(existingContainer.VolumeMounts, desiredContainer.VolumeMounts[desiredIdx])
		if existingIdx < 0 {
			existingContainer.VolumeMounts = append(existingContainer.VolumeMounts, desiredContainer.VolumeMounts[desiredIdx])
			changed = true
		} else if !reflect.DeepEqual(existingContainer.VolumeMounts[existingIdx], desiredContainer.VolumeMounts[desiredIdx]) {
			existingContainer.VolumeMounts[existingIdx] = desiredContainer.VolumeMounts[desiredIdx]
			changed = true
		}
	}

	// Check custom policy annotations in existing and not in desired to delete volumes associated
	// From the APIManager CR, operator does not know which custom policies have been deleted to reconcile volumes
	// Only volumes associated to custom policies are deleted. The operator still allows manually arbitrary mounted volumes
	existingCustomPolicyVolumeNames := component.ApicastPolicyVolumeNamesFromAnnotations(existing.Annotations)
	desiredCustomPolicyVolumeNames := component.ApicastPolicyVolumeNamesFromAnnotations(desired.Annotations)
	volumesToDelete := helper.ArrayStringDifference(existingCustomPolicyVolumeNames, desiredCustomPolicyVolumeNames)
	for _, volumeNameToDelete := range volumesToDelete {
		idx := helper.FindVolumeMountByName(existingContainer.VolumeMounts, volumeNameToDelete)
		if idx >= 0 {
			// Found a existing volume that needs to be removed
			// remove index
			existingContainer.VolumeMounts = append(existingContainer.VolumeMounts[:idx], existingContainer.VolumeMounts[idx+1:]...)
			changed = true
		}
	}

	// Check tracing config annotations in existing and not in desired to delete volumes associated
	// From the APIManager CR, operator does not know which custom policies have been deleted to reconcile volumes
	// Only volumes associated to custom policies are deleted. The operator still allows manually arbitrary mounted volumes
	existingTracingConfigVolumeNames := component.ApicastTracingConfigVolumeNamesFromAnnotations(existing.Annotations)
	desiredTracingConfigVolumeNames := component.ApicastTracingConfigVolumeNamesFromAnnotations(desired.Annotations)
	volumesToDelete = helper.ArrayStringDifference(existingTracingConfigVolumeNames, desiredTracingConfigVolumeNames)
	for _, volumeNameToDelete := range volumesToDelete {
		idx := helper.FindVolumeMountByName(existingContainer.VolumeMounts, volumeNameToDelete)
		if idx >= 0 {
			// Found a existing volume that needs to be removed
			// remove index
			existingContainer.VolumeMounts = append(existingContainer.VolumeMounts[:idx], existingContainer.VolumeMounts[idx+1:]...)
			changed = true
		}
	}

	// Check custom environment annotations in existing and not in desired to delete volumes associated
	// From the APIManager CR, operator does not know which custom environment have been deleted to reconcile volumes
	// Only volumes associated to custom environments are deleted. The operator still allows manually arbitrary mounted volumes
	existingCustomEnvVolumeNames := component.ApicastEnvVolumeNamesFromAnnotations(existing.Annotations)
	desiredCustomEnvVolumeNames := component.ApicastEnvVolumeNamesFromAnnotations(desired.Annotations)
	volumesToDelete = helper.ArrayStringDifference(existingCustomEnvVolumeNames, desiredCustomEnvVolumeNames)
	for _, volumeNameToDelete := range volumesToDelete {
		idx := helper.FindVolumeMountByName(existingContainer.VolumeMounts, volumeNameToDelete)
		if idx >= 0 {
			// Found a existing volume that needs to be removed
			// remove index
			existingContainer.VolumeMounts = append(existingContainer.VolumeMounts[:idx], existingContainer.VolumeMounts[idx+1:]...)
			changed = true
		}
	}

	// Check for existing volumeMounts associated to the TLS port that is no longer desired
	// Only the volumeMount associated to the TLS port is deleted. The operator still allows manually arbitrary mounted volumes
	existingIdx := helper.FindVolumeMountByName(existingContainer.VolumeMounts, component.HTTPSCertificatesVolumeName)
	desiredIdx := helper.FindVolumeMountByName(desiredContainer.VolumeMounts, component.HTTPSCertificatesVolumeName)
	if desiredIdx < 0 && existingIdx >= 0 {
		// volumeMount exists in existing and does not exist in desired => Remove from the list
		// shift all of the elements at the right of the deleting index by one to the left
		existingContainer.VolumeMounts = append(existingContainer.VolumeMounts[:existingIdx], existingContainer.VolumeMounts[existingIdx+1:]...)
		changed = true
	}

	return changed, nil
}

// volumeMountsMutator implements basic VolumeMount reconcilliation
// Added when in desired and not in existing
// Updated when in desired and in existing but not equal
// Existing not in desired will NOT be removed. Allows manually added volumemounts
func apicastVolumesMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	changed := false
	existingSpec := &existing.Spec.Template.Spec
	desiredSpec := &desired.Spec.Template.Spec

	// Add desired not in existing
	for desiredIdx := range desiredSpec.Volumes {
		existingIdx := helper.FindVolumeByName(existingSpec.Volumes, desiredSpec.Volumes[desiredIdx].Name)
		if existingIdx < 0 {
			existingSpec.Volumes = append(existingSpec.Volumes, desiredSpec.Volumes[desiredIdx])
			changed = true
		} else if !helper.VolumeFromSecretEqual(existingSpec.Volumes[existingIdx], desiredSpec.Volumes[desiredIdx]) {
			existingSpec.Volumes[existingIdx] = desiredSpec.Volumes[desiredIdx]
			changed = true
		}
	}

	// Check custom policy annotations in existing and not in desired to delete volumes associated
	// From the APIManager CR, operator does not know which custom policies have been deleted to reconcile volumes
	// Only volumes associated to custom policies are deleted. The operator still allows manually arbitrary mounted volumes
	existingCustomPolicyVolumeNames := component.ApicastPolicyVolumeNamesFromAnnotations(existing.Annotations)
	desiredCustomPolicyVolumeNames := component.ApicastPolicyVolumeNamesFromAnnotations(desired.Annotations)
	volumesToDelete := helper.ArrayStringDifference(existingCustomPolicyVolumeNames, desiredCustomPolicyVolumeNames)
	for _, volumeNameToDelete := range volumesToDelete {
		idx := helper.FindVolumeByName(existingSpec.Volumes, volumeNameToDelete)
		if idx >= 0 {
			// Found a existing volume that needs to be removed
			// remove index
			existingSpec.Volumes = append(existingSpec.Volumes[:idx], existingSpec.Volumes[idx+1:]...)
			changed = true
		}
	}

	// Check custom policy annotations in existing and not in desired to delete volumes associated
	// From the APIManager CR, operator does not know which tracing config have been deleted to reconcile volumes
	// Only volumes associated to tracing configs are deleted. The operator still allows manually arbitrary mounted volumes
	existingTracingConfigVolumeNames := component.ApicastTracingConfigVolumeNamesFromAnnotations(existing.Annotations)
	desiredTracingConfigVolumeNames := component.ApicastTracingConfigVolumeNamesFromAnnotations(desired.Annotations)
	volumesToDelete = helper.ArrayStringDifference(existingTracingConfigVolumeNames, desiredTracingConfigVolumeNames)
	for _, volumeNameToDelete := range volumesToDelete {
		idx := helper.FindVolumeByName(existingSpec.Volumes, volumeNameToDelete)
		if idx >= 0 {
			// Found a existing volume that needs to be removed
			// remove index
			existingSpec.Volumes = append(existingSpec.Volumes[:idx], existingSpec.Volumes[idx+1:]...)
			changed = true
		}
	}

	// Check custom environment annotations in existing and not in desired to delete volumes associated
	// From the APIManager CR, operator does not know which custom environments have been deleted to reconcile volumes
	// Only volumes associated to custom environments are deleted. The operator still allows manually arbitrary mounted volumes
	existingCustomEnvVolumeNames := component.ApicastEnvVolumeNamesFromAnnotations(existing.Annotations)
	desiredCustomEnvVolumeNames := component.ApicastEnvVolumeNamesFromAnnotations(desired.Annotations)
	volumesToDelete = helper.ArrayStringDifference(existingCustomEnvVolumeNames, desiredCustomEnvVolumeNames)
	for _, volumeNameToDelete := range volumesToDelete {
		idx := helper.FindVolumeByName(existingSpec.Volumes, volumeNameToDelete)
		if idx >= 0 {
			// Found a existing volume that needs to be removed
			// remove index
			existingSpec.Volumes = append(existingSpec.Volumes[:idx], existingSpec.Volumes[idx+1:]...)
			changed = true
		}
	}

	// Check for existing volume associated to the TLS port that is no longer desired
	// Only the volume associated to the TLS port is deleted. The operator still allows manually arbitrary mounted volumes
	existingIdx := helper.FindVolumeByName(existingSpec.Volumes, component.HTTPSCertificatesVolumeName)
	desiredIdx := helper.FindVolumeByName(desiredSpec.Volumes, component.HTTPSCertificatesVolumeName)
	if desiredIdx < 0 && existingIdx >= 0 {
		// volume exists in existing and does not exist in desired => Remove from the list
		// shift all of the elements at the right of the deleting index by one to the left
		existingSpec.Volumes = append(existingSpec.Volumes[:existingIdx], existingSpec.Volumes[existingIdx+1:]...)
		changed = true
	}

	return changed, nil
}

func apicastCustomPolicyAnnotationsMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	// It is expected that APIManagerMutator has already added desired annotations to the existing annotations
	// find existing custom policy annotations not in desired and delete them
	updated := false
	existingCustomPolicyVolumeNames := component.ApicastPolicyVolumeNamesFromAnnotations(existing.Annotations)
	desiredCustomPolicyVolumeNames := component.ApicastPolicyVolumeNamesFromAnnotations(desired.Annotations)
	if !helper.StringSliceEqualWithoutOrder(existingCustomPolicyVolumeNames, desiredCustomPolicyVolumeNames) {
		for key := range existing.Annotations {
			if strings.HasPrefix(key, component.CustomPoliciesAnnotationPartialKey) {
				delete(existing.Annotations, key)
			}
		}

		for key, val := range desired.Annotations {
			if strings.HasPrefix(key, component.CustomPoliciesAnnotationPartialKey) {
				existing.Annotations[key] = val
			}
		}

		updated = true
	}
	return updated, nil
}

func apicastTracingConfigAnnotationsMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	// It is expected that APIManagerMutator has already added desired annotations to the existing annotations
	// find existing tracing config volume annotations not in desired and delete them
	updated := false
	existingTracingConfigVolumeNames := component.ApicastTracingConfigVolumeNamesFromAnnotations(existing.Annotations)
	desiredTracingConfigVolumeNames := component.ApicastTracingConfigVolumeNamesFromAnnotations(desired.Annotations)
	if !helper.StringSliceEqualWithoutOrder(existingTracingConfigVolumeNames, desiredTracingConfigVolumeNames) {
		for key := range existing.Annotations {
			if strings.HasPrefix(key, component.APIcastTracingConfigAnnotationPartialKey) {
				delete(existing.Annotations, key)
			}
		}

		for key, val := range desired.Annotations {
			if strings.HasPrefix(key, component.APIcastTracingConfigAnnotationPartialKey) {
				existing.Annotations[key] = val
			}
		}

		updated = true
	}
	return updated, nil
}

func apicastCustomEnvAnnotationsMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	// It is expected that APIManagerMutator has already added desired annotations to the existing annotations
	// find existing custom environments annotations not in desired and delete them
	updated := false
	existingCustomEnvVolumeNames := component.ApicastEnvVolumeNamesFromAnnotations(existing.Annotations)
	desiredCustomEnvVolumeNames := component.ApicastEnvVolumeNamesFromAnnotations(desired.Annotations)
	if !helper.StringSliceEqualWithoutOrder(existingCustomEnvVolumeNames, desiredCustomEnvVolumeNames) {
		for key := range existing.Annotations {
			if strings.HasPrefix(key, component.CustomEnvironmentsAnnotationPartialKey) {
				delete(existing.Annotations, key)
			}
		}

		for key, val := range desired.Annotations {
			if strings.HasPrefix(key, component.CustomEnvironmentsAnnotationPartialKey) {
				existing.Annotations[key] = val
			}
		}

		updated = true
	}
	return updated, nil
}

func apicastPodTemplateEnvConfigMapAnnotationsMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	// Only reconcile the pod annotation regarding apicast-environment hash
	desiredVal, ok := desired.Spec.Template.Annotations[APIcastEnvironmentCMAnnotation]
	if !ok {
		return false, nil
	}

	updated := false
	existingVal, ok := existing.Spec.Template.Annotations[APIcastEnvironmentCMAnnotation]
	if !ok || existingVal != desiredVal {
		if existing.Spec.Template.Annotations == nil {
			existing.Spec.Template.Annotations = map[string]string{}
		}
		existing.Spec.Template.Annotations[APIcastEnvironmentCMAnnotation] = desiredVal
		updated = true
	}

	return updated, nil
}

func Apicast(apimanager *appsv1alpha1.APIManager, cl client.Client) (*component.Apicast, error) {
	optsProvider := NewApicastOptionsProvider(apimanager, cl)
	opts, err := optsProvider.GetApicastOptions()
	if err != nil {
		return nil, err
	}
	return component.NewApicast(opts), nil
}
