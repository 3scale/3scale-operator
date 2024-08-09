package operator

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"strings"

	k8sappsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/upgrade"
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
	ampImages, err := AmpImages(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	apicast, err := Apicast(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	stagingMutators := []reconcilers.DMutateFn{
		reconcilers.DeploymentAnnotationsMutator,
		reconcilers.DeploymentContainerResourcesMutator,
		reconcilers.DeploymentAffinityMutator,
		reconcilers.DeploymentTolerationsMutator,
		reconcilers.DeploymentPodTemplateLabelsMutator,
		apicastLogLevelEnvVarMutator,
		apicastTracingConfigEnvVarsMutator,
		apicastOpentelemetryConfigEnvVarsMutator,
		apicastEnvironmentEnvVarMutator,
		apicastHTTPSEnvVarMutator,
		apicastProxyConfigurationsEnvVarMutator,
		apicastServiceCacheSizeEnvVarMutator,
		apicastVolumeMountsMutator,
		apicastVolumesMutator,
		apicastCustomPolicyAnnotationsMutator,  // Should be always after volume mutator
		apicastTracingConfigAnnotationsMutator, // Should be always after volume mutator
		apicastOpentelemetryConfigAnnotationsMutator,
		apicastCustomEnvAnnotationsMutator, // Should be always after volume mutator
		portsMutator,
		apicastPodTemplateEnvConfigMapAnnotationsMutator,
		reconcilers.DeploymentPriorityClassMutator,
		reconcilers.DeploymentTopologySpreadConstraintsMutator,
		reconcilers.DeploymentPodTemplateAnnotationsMutator,
		reconcilers.DeploymentPodContainerImageMutator,
	}

	if r.apiManager.Spec.Apicast.StagingSpec.Replicas != nil {
		stagingMutators = append(stagingMutators, reconcilers.DeploymentReplicasMutator)
	}

	// Staging Deployment
	err = r.ReconcileDeployment(apicast.StagingDeployment(ampImages.Options.ApicastImage), reconcilers.DeploymentMutator(stagingMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// 3scale 2.14 -> 2.15
	isMigrated, err := upgrade.MigrateDeploymentConfigToDeployment(component.ApicastStagingName, r.apiManager.GetNamespace(), false, r.Client(), nil)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !isMigrated {
		return reconcile.Result{Requeue: true}, nil
	}

	// add apicast production env var mutator
	productionMutators := []reconcilers.DMutateFn{
		reconcilers.DeploymentAnnotationsMutator,
		reconcilers.DeploymentContainerResourcesMutator,
		reconcilers.DeploymentAffinityMutator,
		reconcilers.DeploymentTolerationsMutator,
		reconcilers.DeploymentPodTemplateLabelsMutator,
		apicastProductionWorkersEnvVarMutator,
		apicastLogLevelEnvVarMutator,
		apicastTracingConfigEnvVarsMutator,
		apicastOpentelemetryConfigEnvVarsMutator,
		apicastEnvironmentEnvVarMutator,
		apicastHTTPSEnvVarMutator,
		apicastProxyConfigurationsEnvVarMutator,
		apicastServiceCacheSizeEnvVarMutator,
		apicastVolumeMountsMutator,
		apicastVolumesMutator,
		apicastCustomPolicyAnnotationsMutator,  // Should be always after volume mutator
		apicastTracingConfigAnnotationsMutator, // Should be always after volume mutator
		apicastOpentelemetryConfigAnnotationsMutator,
		apicastCustomEnvAnnotationsMutator, // Should be always after volume
		portsMutator,
		apicastPodTemplateEnvConfigMapAnnotationsMutator,
		reconcilers.DeploymentPriorityClassMutator,
		reconcilers.DeploymentTopologySpreadConstraintsMutator,
		reconcilers.DeploymentPodTemplateAnnotationsMutator,
		reconcilers.DeploymentPodContainerImageMutator,
		reconcilers.DeploymentPodInitContainerImageMutator,
	}

	if r.apiManager.Spec.Apicast.ProductionSpec.Replicas != nil {
		productionMutators = append(productionMutators, reconcilers.DeploymentReplicasMutator)
	}

	// Production Deployment
	err = r.ReconcileDeployment(apicast.ProductionDeployment(ampImages.Options.ApicastImage), reconcilers.DeploymentMutator(productionMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// 3scale 2.14 -> 2.15
	isMigrated, err = upgrade.MigrateDeploymentConfigToDeployment(component.ApicastProductionName, r.apiManager.GetNamespace(), false, r.Client(), nil)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !isMigrated {
		return reconcile.Result{Requeue: true}, nil
	}

	serviceMutators := []reconcilers.MutateFn{
		getApiCastServiceMutator(r.apiManager.ObjectMeta.GetAnnotations()),
		reconcilers.ServiceSelectorMutator,
	}

	// Staging Service
	err = r.ReconcileService(apicast.StagingService(), reconcilers.ServiceMutator(serviceMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// Production Service
	err = r.ReconcileService(apicast.ProductionService(), reconcilers.ServiceMutator(serviceMutators...))
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

	err = r.ReconcileGrafanaDashboards(apicast.ApicastMainAppGrafanaV5Dashboard(sumRate), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ReconcileGrafanaDashboards(apicast.ApicastServicesGrafanaV5Dashboard(sumRate), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileGrafanaDashboards(apicast.ApicastMainAppGrafanaV4Dashboard(sumRate), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ReconcileGrafanaDashboards(apicast.ApicastServicesGrafanaV4Dashboard(sumRate), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(apicast.ApicastPrometheusRules(), reconcilers.RemovePrometheusRulesMutator)
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

	err = r.ReconcileHpa(component.DefaultHpa(component.ApicastProductionName, r.apiManager.Namespace), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err := r.reconcileAPImanagerCR(context.TODO())
	if err != nil {
		return ctrl.Result{}, err
	}
	if res.Requeue {
		return res, nil
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

func apicastProductionWorkersEnvVarMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	// Reconcile EnvVar only for "APICAST_WORKERS"
	return reconcilers.DeploymentEnvVarReconciler(desired, existing, "APICAST_WORKERS"), nil
}

func apicastLogLevelEnvVarMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	// Reconcile EnvVar only for "APICAST_LOG_LEVEL"
	return reconcilers.DeploymentEnvVarReconciler(desired, existing, "APICAST_LOG_LEVEL"), nil
}

func apicastTracingConfigEnvVarsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	// Reconcile EnvVars related to opentracing
	var changed bool
	changed = reconcilers.DeploymentEnvVarReconciler(desired, existing, "OPENTRACING_TRACER")

	tmpChanged := reconcilers.DeploymentEnvVarReconciler(desired, existing, "OPENTRACING_CONFIG")
	changed = changed || tmpChanged

	return changed, nil
}

func apicastOpentelemetryConfigEnvVarsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	// Reconcile EnvVars related to opentracing
	var changed bool
	changed = reconcilers.DeploymentEnvVarReconciler(desired, existing, "OPENTELEMETRY")

	tmpChanged := reconcilers.DeploymentEnvVarReconciler(desired, existing, "OPENTELEMETRY_CONFIG")
	changed = changed || tmpChanged

	return changed, nil
}

func apicastEnvironmentEnvVarMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	// Reconcile EnvVar only for "APICAST_ENVIRONMENT"
	return reconcilers.DeploymentEnvVarReconciler(desired, existing, "APICAST_ENVIRONMENT"), nil
}

func apicastHTTPSEnvVarMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	// Reconcile EnvVars related to opentracing
	var changed bool

	for _, envVar := range []string{
		"APICAST_HTTPS_PORT",
		"APICAST_HTTPS_VERIFY_DEPTH",
		"APICAST_HTTPS_CERTIFICATE",
		"APICAST_HTTPS_CERTIFICATE_KEY",
	} {
		tmpChanged := reconcilers.DeploymentEnvVarReconciler(desired, existing, envVar)
		changed = changed || tmpChanged
	}

	return changed, nil
}

func apicastProxyConfigurationsEnvVarMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	// Reconcile EnvVars related to APIcast proxy-related configurations
	var changed bool

	for _, envVar := range []string{
		"ALL_PROXY",
		"HTTP_PROXY",
		"HTTPS_PROXY",
		"NO_PROXY",
	} {
		tmpChanged := reconcilers.DeploymentEnvVarReconciler(desired, existing, envVar)
		changed = changed || tmpChanged
	}

	return changed, nil
}

func apicastServiceCacheSizeEnvVarMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	return reconcilers.DeploymentEnvVarReconciler(desired, existing, "APICAST_SERVICE_CACHE_SIZE"), nil
}

func portsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	changed := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Containers[0].Ports, desired.Spec.Template.Spec.Containers[0].Ports) {
		changed = true
		existing.Spec.Template.Spec.Containers[0].Ports = desired.Spec.Template.Spec.Containers[0].Ports
	}

	return changed, nil
}

// apicastVolumeMountsMutator implements basic VolumeMount reconciliation
// Added when in desired and not in existing
// Updated when in desired and in existing but not equal
// Existing not in desired will NOT be removed. Allows manually added volumemounts
func apicastVolumeMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
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

	// If desired volume mounts do not have opentelemtry but it is present in the existsing volume mounts, remove it because it is not required.
	existingOtlpVolumeMountIdx := helper.FindVolumeMountByName(existingContainer.VolumeMounts, component.OpentelemetryConfigurationVolumeName)
	if existingOtlpVolumeMountIdx >= 0 {
		desiredOtlpVolumeMountIdx := helper.FindVolumeMountByName(desiredContainer.VolumeMounts, component.OpentelemetryConfigurationVolumeName)
		if desiredOtlpVolumeMountIdx < 0 {
			existingContainer.VolumeMounts = append(existingContainer.VolumeMounts[:existingOtlpVolumeMountIdx], existingContainer.VolumeMounts[existingOtlpVolumeMountIdx+1:]...)
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

// apicastVolumesMutator implements basic VolumeMount reconciliation
// Added when in desired and not in existing
// Updated when in desired and in existing but not equal
// Existing not in desired will NOT be removed. Allows manually added volumemounts
func apicastVolumesMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
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

	// If desired volumes do not have opentelemtry but it is present in the existsing volumes, remove it because it is not required.
	existingOtlpVolumeMountIdx := helper.FindVolumeByName(existingSpec.Volumes, component.OpentelemetryConfigurationVolumeName)
	if existingOtlpVolumeMountIdx >= 0 {
		desiredOtlpVolumeMountIdx := helper.FindVolumeByName(desiredSpec.Volumes, component.OpentelemetryConfigurationVolumeName)
		if desiredOtlpVolumeMountIdx < 0 {
			existingSpec.Volumes = append(existingSpec.Volumes[:existingOtlpVolumeMountIdx], existingSpec.Volumes[existingOtlpVolumeMountIdx+1:]...)
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

func apicastCustomPolicyAnnotationsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
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

func apicastTracingConfigAnnotationsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
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

func apicastOpentelemetryConfigAnnotationsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	// It is expected that APIManagerMutator has already added desired annotations to the existing annotations
	// find existing tracing config volume annotations not in desired and delete them
	updated := false
	existingOpentelemetryConfigVolumeNames := component.ApicastOpentelemetryConfigVolumeNamesFromAnnotations(existing.Annotations)
	desiredOpentelemetryConfigVolumeNames := component.ApicastOpentelemetryConfigVolumeNamesFromAnnotations(desired.Annotations)
	if !helper.StringSliceEqualWithoutOrder(existingOpentelemetryConfigVolumeNames, desiredOpentelemetryConfigVolumeNames) {
		for key := range existing.Annotations {
			if strings.HasPrefix(key, component.APIcastOpentelemetryConfigAnnotationPartialKey) {
				delete(existing.Annotations, key)
			}
		}

		for key, val := range desired.Annotations {
			if strings.HasPrefix(key, component.APIcastOpentelemetryConfigAnnotationPartialKey) {
				existing.Annotations[key] = val
			}
		}

		updated = true
	}
	return updated, nil
}

func apicastCustomEnvAnnotationsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
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

func apicastPodTemplateEnvConfigMapAnnotationsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
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

func (r *ApicastReconciler) reconcileAPImanagerCR(ctx context.Context) (ctrl.Result, error) {

	changed := false
	changed, err := r.reconcileApimanagerSecretLabels(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	if changed {
		err = r.Client().Update(ctx, r.apiManager)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{Requeue: changed}, nil
}

func (r *ApicastReconciler) reconcileApimanagerSecretLabels(ctx context.Context) (bool, error) {
	secretUIDs, err := r.getSecretUIDs(ctx)
	if err != nil {
		return false, err
	}

	return replaceAPIManagerSecretLabels(r.apiManager, secretUIDs), nil
}

func (r *ApicastReconciler) getSecretUIDs(ctx context.Context) ([]string, error) {
	// production custom policy
	// staging custom policy

	secretKeys := []client.ObjectKey{}
	if r.apiManager.Spec.Apicast.ProductionSpec.CustomPolicies != nil {
		for _, customPolicy := range r.apiManager.Spec.Apicast.ProductionSpec.CustomPolicies {
			secretKeys = append(secretKeys, client.ObjectKey{
				Name:      customPolicy.SecretRef.Name,
				Namespace: r.apiManager.Namespace,
			})
		}
	}

	if r.apiManager.Spec.Apicast.StagingSpec.CustomPolicies != nil {
		for _, customPolicy := range r.apiManager.Spec.Apicast.StagingSpec.CustomPolicies {
			secretKeys = append(secretKeys, client.ObjectKey{
				Name:      customPolicy.SecretRef.Name,
				Namespace: r.apiManager.Namespace,
			})
		}
	}

	if r.apiManager.OpenTelemetryEnabledForStaging() {
		if r.apiManager.Spec.Apicast.StagingSpec.OpenTelemetry.TracingConfigSecretRef != nil {
			secretKeys = append(secretKeys, client.ObjectKey{
				Name:      r.apiManager.Spec.Apicast.StagingSpec.OpenTelemetry.TracingConfigSecretRef.Name,
				Namespace: r.apiManager.Namespace,
			})
		}
	}

	if r.apiManager.OpenTelemetryEnabledForProduction() {
		if r.apiManager.Spec.Apicast.ProductionSpec.OpenTelemetry.TracingConfigSecretRef != nil {
			secretKeys = append(secretKeys, client.ObjectKey{
				Name:      r.apiManager.Spec.Apicast.ProductionSpec.OpenTelemetry.TracingConfigSecretRef.Name,
				Namespace: r.apiManager.Namespace,
			})
		}
	}

	uids := []string{}
	for idx := range secretKeys {
		secret := &v1.Secret{}
		secretKey := secretKeys[idx]
		err := r.Client().Get(ctx, secretKey, secret)
		r.Logger().V(1).Info("read secret", "objectKey", secretKey, "error", err)
		if err != nil {
			return nil, err
		}
		uids = append(uids, string(secret.GetUID()))
	}

	return uids, nil
}

func Apicast(apimanager *appsv1alpha1.APIManager, cl client.Client) (*component.Apicast, error) {
	optsProvider := NewApicastOptionsProvider(apimanager, cl)
	opts, err := optsProvider.GetApicastOptions()
	if err != nil {
		return nil, err
	}
	return component.NewApicast(opts), nil
}

func GetSystemRedisSecret(apimanagerNs string, ctx context.Context, client client.Client, logger logr.Logger) bool {
	backendRedisSecret := &v1.Secret{}
	err := client.Get(ctx, types.NamespacedName{
		Name:      "system-redis",
		Namespace: apimanagerNs,
	}, backendRedisSecret)
	if err != nil {
		logger.Error(err, "Failed to get system-redis secret, cant check for authenticated redis sentinels, check the system-redis secret exists")
		return false
	}

	RedisSentinelHost := strings.Contains(string(backendRedisSecret.Data["SENTINEL_HOSTS"]), "@")

	return RedisSentinelHost
}
