package controllers

import (
	"context"
	"fmt"
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	subController "github.com/3scale/3scale-operator/controllers/subscription"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/apispkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	"github.com/RHsyseng/operator-utils/pkg/olm"
	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	k8sappsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sort"
)

type APIManagerStatusReconciler struct {
	*reconcilers.BaseReconciler
	apimanagerResource *appsv1alpha1.APIManager
	logger             logr.Logger
	preflightsErr      error
}

func NewAPIManagerStatusReconciler(b *reconcilers.BaseReconciler, apimanagerResource *appsv1alpha1.APIManager, preflightsError error) *APIManagerStatusReconciler {
	return &APIManagerStatusReconciler{
		BaseReconciler:     b,
		apimanagerResource: apimanagerResource,
		logger:             b.Logger().WithValues("Status Reconciler", client.ObjectKeyFromObject(apimanagerResource)),
		preflightsErr:      preflightsError,
	}
}

func (s *APIManagerStatusReconciler) Reconcile() (reconcile.Result, error) {
	s.logger.V(1).Info("START")
	newStatus, err := s.calculateStatus()
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to calculate status: %w", err)
	}

	apiManagerAvailable := false
	for _, s := range s.apimanagerResource.Status.Conditions {
		if s.Type == appsv1alpha1.APIManagerAvailableConditionType && s.IsTrue() {
			apiManagerAvailable = true
		}
	}

	equalStatus := s.apimanagerResource.Status.Equals(newStatus, s.logger)
	s.logger.V(1).Info("Status", "status is different", !equalStatus)
	if equalStatus && apiManagerAvailable {
		// Steady state
		s.logger.V(1).Info("Status was not updated")
		return reconcile.Result{}, nil
	}

	s.apimanagerResource.Status = *newStatus
	updateErr := s.Client().Status().Update(s.Context(), s.apimanagerResource)
	if updateErr != nil {
		// Ignore conflicts, resource might just be outdated.
		if errors.IsConflict(updateErr) {
			s.logger.Info("Failed to update status: resource might just be outdated")
			return reconcile.Result{Requeue: true}, nil
		}

		return reconcile.Result{}, fmt.Errorf("failed to update status: %w", updateErr)
	}

	if !apiManagerAvailable {
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

func (s *APIManagerStatusReconciler) calculateStatus() (*appsv1alpha1.APIManagerStatus, error) {
	newStatus := &appsv1alpha1.APIManagerStatus{}

	deployments, err := s.existingDeployments()
	if err != nil {
		return nil, err
	}

	newStatus.Conditions = s.apimanagerResource.Status.Conditions.Copy()

	availableCondition, err := s.apimanagerAvailableCondition(deployments)
	if err != nil {
		return nil, err
	}
	newStatus.Conditions.SetCondition(availableCondition)

	if s.preflightsErr == nil {
		s.reconcileHpaWarningMessages(&newStatus.Conditions, s.apimanagerResource)
		s.reconcileOpenTracingDeprecationMessage(&newStatus.Conditions, s.apimanagerResource)
	}

	if !helper.IsPreflightBypassed() {
		err = s.reconcilePreflightsStatus(&newStatus.Conditions, s.apimanagerResource)
		if err != nil {
			return nil, err
		}
	}

	deploymentStatus := olm.GetDeploymentStatus(deployments)
	newStatus.Deployments = deploymentStatus

	return newStatus, nil
}

func (s *APIManagerStatusReconciler) expectedDeploymentNames(instance *appsv1alpha1.APIManager) []string {
	var systemDatabaseType component.SystemDatabaseType
	var externalRedisDatabases bool
	var externalZyncDatabase bool

	if instance.IsExternal(appsv1alpha1.SystemDatabase) {
		systemDatabaseType = component.SystemDatabaseTypeExternal
	} else {
		if instance.IsSystemPostgreSQLEnabled() {
			systemDatabaseType = component.SystemDatabaseTypeInternalPostgreSQL
		} else {
			systemDatabaseType = component.SystemDatabaseTypeInternalMySQL
		}
	}
	if instance.IsExternal(appsv1alpha1.SystemRedis) {
		externalRedisDatabases = true
	}
	if instance.IsExternal(appsv1alpha1.ZyncDatabase) && s.apimanagerResource.IsZyncEnabled() {
		externalZyncDatabase = true
	}

	deploymentLister := component.DeploymentsLister{
		SystemDatabaseType:     systemDatabaseType,
		ExternalRedisDatabases: externalRedisDatabases,
		ExternalZyncDatabase:   externalZyncDatabase,
		IsZyncEnabled:          s.apimanagerResource.IsZyncEnabled(),
	}

	return deploymentLister.DeploymentNames()
}

func (s *APIManagerStatusReconciler) deploymentsAvailable(existingDeployments []k8sappsv1.Deployment) bool {
	expectedDeploymentNames := s.expectedDeploymentNames(s.apimanagerResource)
	for _, deploymentName := range expectedDeploymentNames {
		foundExistingDeploymentIdx := -1
		for idx, existingDeployment := range existingDeployments {
			if existingDeployment.Name == deploymentName {
				foundExistingDeploymentIdx = idx
				break
			}
		}
		if foundExistingDeploymentIdx == -1 || !helper.IsDeploymentAvailable(&existingDeployments[foundExistingDeploymentIdx]) {
			return false
		}
	}

	return true
}

func (s *APIManagerStatusReconciler) existingDeployments() ([]k8sappsv1.Deployment, error) {
	expectedDeploymentNames := s.expectedDeploymentNames(s.apimanagerResource)

	var deployments []k8sappsv1.Deployment
	for _, dName := range expectedDeploymentNames {
		existingDeployment := &k8sappsv1.Deployment{}
		err := s.Client().Get(context.Background(), types.NamespacedName{Namespace: s.apimanagerResource.Namespace, Name: dName}, existingDeployment)
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		}
		if err != nil && errors.IsNotFound(err) {
			continue
		}

		for _, ownerRef := range existingDeployment.GetOwnerReferences() {
			if ownerRef.UID == s.apimanagerResource.UID {
				deployments = append(deployments, *existingDeployment)
				break
			}
		}
	}
	sort.Slice(deployments, func(i, j int) bool { return deployments[i].Name < deployments[j].Name })

	return deployments, nil
}

func (s *APIManagerStatusReconciler) apimanagerAvailableCondition(existingDeployments []k8sappsv1.Deployment) (common.Condition, error) {
	deploymentsAvailable := s.deploymentsAvailable(existingDeployments)

	defaultRoutesReady, err := s.defaultRoutesReady()
	if err != nil {
		return common.Condition{}, err
	}

	newAvailableCondition := common.Condition{
		Type:   appsv1alpha1.APIManagerAvailableConditionType,
		Status: v1.ConditionFalse,
	}

	s.logger.V(1).Info("Status apimanagerAvailableCondition", "deploymentsAvailable", deploymentsAvailable, "defaultRoutesReady", defaultRoutesReady)
	if deploymentsAvailable && defaultRoutesReady {
		newAvailableCondition.Status = v1.ConditionTrue
	}

	return newAvailableCondition, nil
}

func (s *APIManagerStatusReconciler) defaultRoutesReady() (bool, error) {
	var expectedRouteHosts []string
	wildcardDomain := s.apimanagerResource.Spec.WildcardDomain
	if s.apimanagerResource.Spec.TenantName != nil {
		expectedRouteHosts = []string{
			fmt.Sprintf("backend-%s.%s", *s.apimanagerResource.Spec.TenantName, wildcardDomain), // Backend Listener route
		}
		if s.apimanagerResource.IsZyncEnabled() {
			zyncRoutes := []string{
				fmt.Sprintf("api-%s-apicast-production.%s", *s.apimanagerResource.Spec.TenantName, wildcardDomain), // Apicast Production default tenant Route
				fmt.Sprintf("api-%s-apicast-staging.%s", *s.apimanagerResource.Spec.TenantName, wildcardDomain),    // Apicast Staging default tenant Route
				fmt.Sprintf("master.%s", wildcardDomain),                                                           // System's Master Portal Route
				fmt.Sprintf("%s.%s", *s.apimanagerResource.Spec.TenantName, wildcardDomain),                        // System's default tenant Developer Portal Route
				fmt.Sprintf("%s-admin.%s", *s.apimanagerResource.Spec.TenantName, wildcardDomain),                  // System's default tenant Admin Portal Route
			}
			expectedRouteHosts = append(expectedRouteHosts, zyncRoutes...)
		}
	} else {
		return false, nil
	}

	listOps := []client.ListOption{
		client.InNamespace(s.apimanagerResource.Namespace),
	}

	routeList := &routev1.RouteList{}
	err := s.Client().List(context.TODO(), routeList, listOps...)
	if err != nil {
		return false, fmt.Errorf("failed to list routes: %w", err)
	}

	routes := append([]routev1.Route(nil), routeList.Items...)
	sort.Slice(routes, func(i, j int) bool { return routes[i].Name < routes[j].Name })

	allDefaultRoutesReady := true
	for _, expectedRouteHost := range expectedRouteHosts {
		routeIdx := helper.RouteFindByHost(routes, expectedRouteHost)
		if routeIdx == -1 {
			s.logger.V(1).Info("Status defaultRoutesReady: route not found", "expectedRouteHost", expectedRouteHost)
			allDefaultRoutesReady = false
		} else {
			matchedRoute := &routes[routeIdx]
			routeReady := helper.IsRouteReady(matchedRoute)
			if !routeReady {
				s.logger.V(1).Info("Status defaultRoutesReady: route not ready", "expectedRouteHost", expectedRouteHost)
				allDefaultRoutesReady = false
			}
		}
	}

	return allDefaultRoutesReady, nil
}

func (s *APIManagerStatusReconciler) reconcileHpaWarningMessages(conditions *common.Conditions, cr *appsv1alpha1.APIManager) {

	cond := common.Condition{
		Type:    appsv1alpha1.APIManagerWarningConditionType,
		Status:  v1.ConditionStatus(metav1.ConditionTrue),
		Reason:  "HPA & ResourceRequirementsEnabled false",
		Message: "HorizontalPodAutoScaling (HPA) can't function if ResourcesRequirementsEnabled is set to false as there would be no resources to compare to in order to scale",
	}
	// check if condition is already present
	foundCondition := conditions.GetConditionByMessage(cond.Message)

	// If hpa is enabled but the condition is not found, add it
	if !*cr.Spec.ResourceRequirementsEnabled && (cr.Spec.Backend.ListenerSpec.Hpa || cr.Spec.Backend.WorkerSpec.Hpa || cr.Spec.Apicast.ProductionSpec.Hpa) && foundCondition == nil {
		*conditions = append(*conditions, cond)
	}

	// if hpa is disabled and condition is found, remove it
	if !cr.Spec.Backend.ListenerSpec.Hpa && !cr.Spec.Backend.WorkerSpec.Hpa && !cr.Spec.Apicast.ProductionSpec.Hpa && foundCondition != nil {
		conditions.RemoveConditionByMessage(cond.Message)
	}

	cond = common.Condition{
		Type:    appsv1alpha1.APIManagerWarningConditionType,
		Status:  v1.ConditionStatus(metav1.ConditionTrue),
		Reason:  "HPA",
		Message: "HorizontalPodAutoscaling (Hpa) enabled overrides values applied to replicas",
	}

	// check if condition is already present
	foundCondition = conditions.GetConditionByMessage(cond.Message)

	// If hpa is enabled but the condition is not found, add it
	if (cr.Spec.Backend.ListenerSpec.Hpa || cr.Spec.Backend.WorkerSpec.Hpa || cr.Spec.Apicast.ProductionSpec.Hpa) && foundCondition == nil {
		*conditions = append(*conditions, cond)
	}

	// if hpa is disabled and condition is found, remove it
	if !cr.Spec.Backend.ListenerSpec.Hpa && !cr.Spec.Backend.WorkerSpec.Hpa && !cr.Spec.Apicast.ProductionSpec.Hpa && foundCondition != nil {
		conditions.RemoveConditionByMessage(cond.Message)
	}

	cond = common.Condition{
		Type:    appsv1alpha1.APIManagerWarningConditionType,
		Status:  v1.ConditionStatus(metav1.ConditionTrue),
		Reason:  "HPA",
		Message: "HorizontalPodAutoscaling (Hpa). Discovered async disabled annotation and Hpa enabled for backend. Hpa for backends will now be disabled.",
	}

	// check if condition is already present
	foundCondition = conditions.GetConditionByMessage(cond.Message)

	// If hpa is enabled but the condition is not found, add it
	if cr.IsBackendHpaEnabled() && cr.IsAsyncDisableAnnotationPresent() && foundCondition == nil {
		*conditions = append(*conditions, cond)
	}

	// if hpa is disabled and async disable is true and condition is found, remove it
	if foundCondition != nil && (!cr.IsBackendHpaEnabled() || !cr.IsAsyncDisableAnnotationPresent()) {
		conditions.RemoveConditionByMessage(cond.Message)
	}
}

func (s *APIManagerStatusReconciler) reconcileOpenTracingDeprecationMessage(conditions *common.Conditions, cr *appsv1alpha1.APIManager) {
	apicastStaging := "Apicast Staging"
	apicastProduction := "Apicast Production"

	// check if condition is already present
	stageFoundCondition := conditions.GetConditionByReason(apicastOpenTracingCondition(apicastStaging).Reason)
	prodFoundCondition := conditions.GetConditionByReason(apicastOpenTracingCondition(apicastProduction).Reason)

	// If opentracing is enabled on apicast staging but the condition is not found, add it
	if cr.IsAPIcastStagingOpenTracingEnabled() && stageFoundCondition == nil {
		*conditions = append(*conditions, apicastOpenTracingCondition(apicastStaging))
	}

	// If opentracing is enabled on apicast production but the condition is not found, add it
	if cr.IsAPIcastProductionOpenTracingEnabled() && prodFoundCondition == nil {
		*conditions = append(*conditions, apicastOpenTracingCondition(apicastProduction))
	}

	// if opentracing is disabled on apicast staging and condition is found, remove it
	if !cr.IsAPIcastStagingOpenTracingEnabled() && stageFoundCondition != nil {
		conditions.RemoveConditionByReason(apicastOpenTracingCondition(apicastStaging).Reason)
	}

	// if opentracing is disabled on apicast production and condition is found, remove it
	if !cr.IsAPIcastProductionOpenTracingEnabled() && prodFoundCondition != nil {
		conditions.RemoveConditionByReason(apicastOpenTracingCondition(apicastProduction).Reason)
	}
}

func apicastOpenTracingCondition(apicast string) common.Condition {
	return common.Condition{
		Type:    appsv1alpha1.APIManagerWarningConditionType,
		Status:  v1.ConditionStatus(metav1.ConditionTrue),
		Reason:  common.ConditionReason(fmt.Sprintf("%s OpenTracing Deprecation", apicast)),
		Message: "OpenTracing is deprecated, please use OpenTelemetry instead",
	}
}

func (s *APIManagerStatusReconciler) reconcilePreflightsStatus(conditions *common.Conditions, cr *appsv1alpha1.APIManager) error {
	prefligtsCondition := common.Condition{
		Type:    appsv1alpha1.APIManagerPreflightsConditionType,
		Status:  v1.ConditionStatus(metav1.ConditionTrue),
		Reason:  common.ConditionReason("PreflightsPass"),
		Message: "All requirements for the current version are met",
	}

	upgradeSuccessfulPreflight := "All requirement for incoming version are met. If using automatic upgrades the upgrade will start shortly, if manual, you can proceed with approval"
	requirementConfigMapNotFoundPreflight := "Requirement config map is not found yet, it should be generated shortly"
	freshInstallPreflightsErrorMessage := fmt.Sprintf("Preflights failed - %s - re-running preflights in 10 minutes", s.preflightsErr)
	upgradePreflightsErrorMessage := fmt.Sprintf("Preflights failed - %s - re-running preflights in 10 minutes", s.preflightsErr)
	multiMinorHopPreflightsMessage := fmt.Sprintf("Preflights failed - %s. Multi minor version hop detected. Reconciliation of this 3scale instance is stopped. Remove the operator and refer to official upgrade path for 3scale Operator", s.preflightsErr)

	reqConfigMap, err := subController.RetrieveRequirementsConfigMap(s.Client())
	if err != nil {
		if errors.IsNotFound(err) {
			prefligtsCondition.Message = requirementConfigMapNotFoundPreflight
			prefligtsCondition.Status = v1.ConditionStatus(metav1.ConditionFalse)
		} else {
			return err
		}
	}

	isMultiHopDetected, err := cr.IsMultiMinorHopDetected()
	if err != nil {
		return err
	}
	if isMultiHopDetected {
		prefligtsCondition.Message = multiMinorHopPreflightsMessage
		prefligtsCondition.Status = v1.ConditionStatus(metav1.ConditionFalse)
	}

	if cr.IsInFreshInstallationScenario() && s.preflightsErr != nil && !isMultiHopDetected {
		prefligtsCondition.Status = v1.ConditionStatus(metav1.ConditionFalse)
		prefligtsCondition.Message = freshInstallPreflightsErrorMessage
	}

	if !cr.IsInFreshInstallationScenario() && s.preflightsErr != nil && !isMultiHopDetected {
		prefligtsCondition.Status = v1.ConditionStatus(metav1.ConditionFalse)
		prefligtsCondition.Message = upgradePreflightsErrorMessage
	}

	if !cr.IsInFreshInstallationScenario() && s.preflightsErr == nil && (version.ThreescaleVersionMajorMinor() != reqConfigMap.Data[helper.RHTThreescaleVersion]) && !isMultiHopDetected {
		prefligtsCondition.Status = v1.ConditionStatus(metav1.ConditionTrue)
		prefligtsCondition.Message = upgradeSuccessfulPreflight
	}

	existingPreflightCondition := conditions.GetConditionByReason(prefligtsCondition.Reason)
	if existingPreflightCondition == nil {
		*conditions = append(*conditions, prefligtsCondition)
	} else {
		conditions.RemoveConditionByReason(prefligtsCondition.Reason)
		existingPreflightCondition.Message = prefligtsCondition.Message
		existingPreflightCondition.Status = prefligtsCondition.Status
		*conditions = append(*conditions, prefligtsCondition)
	}

	return nil
}
