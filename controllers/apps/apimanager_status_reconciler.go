package controllers

import (
	"context"
	"fmt"
	"sort"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/RHsyseng/operator-utils/pkg/olm"
	"github.com/go-logr/logr"
	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type APIManagerStatusReconciler struct {
	*reconcilers.BaseReconciler
	apimanagerResource *appsv1alpha1.APIManager
	logger             logr.Logger
}

func NewAPIManagerStatusReconciler(b *reconcilers.BaseReconciler, apimanagerResource *appsv1alpha1.APIManager) *APIManagerStatusReconciler {
	return &APIManagerStatusReconciler{
		BaseReconciler:     b,
		apimanagerResource: apimanagerResource,
		logger:             b.Logger().WithValues("Status Reconciler", apimanagerResource.Name),
	}
}

func (s *APIManagerStatusReconciler) Reconcile() (reconcile.Result, error) {
	s.logger.V(1).Info("START")

	newStatus, err := s.calculateStatus()
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to calculate status: %w", err)
	}

	equalStatus := s.apimanagerResource.Status.Equals(newStatus, s.logger)
	s.logger.V(1).Info("Status", "status is different", !equalStatus)
	if equalStatus {
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

		return reconcile.Result{}, fmt.Errorf("Failed to update status: %w", updateErr)
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

	deploymentStatus := olm.GetDeploymentConfigStatus(deployments)
	newStatus.Deployments = deploymentStatus

	return newStatus, nil
}

func (s *APIManagerStatusReconciler) expectedDeploymentNames(instance *appsv1alpha1.APIManager) []string {
	var systemDatabaseType component.SystemDatabaseType
	var externalRedisDatabases bool
	var externalZyncDatabase bool
	if instance.IsExternalDatabaseEnabled() {
		systemDatabaseType = component.SystemDatabaseTypeExternal
		externalRedisDatabases = true
		if instance.IsZyncExternalDatabaseEnabled() {
			externalZyncDatabase = true
		}
	} else {
		if instance.IsSystemPostgreSQLEnabled() {
			systemDatabaseType = component.SystemDatabaseTypeInternalPostgreSQL
		} else {
			systemDatabaseType = component.SystemDatabaseTypeInternalMySQL
		}
	}

	deploymentLister := component.DeploymentsLister{
		SystemDatabaseType:     systemDatabaseType,
		ExternalRedisDatabases: externalRedisDatabases,
		ExternalZyncDatabase:   externalZyncDatabase,
	}

	return deploymentLister.DeploymentNames()
}

func (s *APIManagerStatusReconciler) deploymentsAvailable(existingDeployments []appsv1.DeploymentConfig) bool {
	expectedDeploymentNames := s.expectedDeploymentNames(s.apimanagerResource)
	for _, deploymentName := range expectedDeploymentNames {
		foundExistingDCIdx := -1
		for idx, existingDC := range existingDeployments {
			if existingDC.Name == deploymentName {
				foundExistingDCIdx = idx
				break
			}
		}
		if foundExistingDCIdx == -1 || !helper.IsDeploymentConfigAvailable(&existingDeployments[foundExistingDCIdx]) {
			return false
		}
	}

	return true
}

func (s *APIManagerStatusReconciler) existingDeployments() ([]appsv1.DeploymentConfig, error) {
	expectedDeploymentNames := s.expectedDeploymentNames(s.apimanagerResource)

	var dcs []appsv1.DeploymentConfig
	for _, dcName := range expectedDeploymentNames {
		existingDeploymentConfig := &appsv1.DeploymentConfig{}
		err := s.Client().Get(context.Background(), types.NamespacedName{Namespace: s.apimanagerResource.Namespace, Name: dcName}, existingDeploymentConfig)
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		}
		if err != nil && errors.IsNotFound(err) {
			continue
		}

		for _, ownerRef := range existingDeploymentConfig.GetOwnerReferences() {
			if ownerRef.UID == s.apimanagerResource.UID {
				dcs = append(dcs, *existingDeploymentConfig)
				break
			}
		}
	}
	sort.Slice(dcs, func(i, j int) bool { return dcs[i].Name < dcs[j].Name })

	return dcs, nil
}

func (s *APIManagerStatusReconciler) apimanagerAvailableCondition(existingDeployments []appsv1.DeploymentConfig) (common.Condition, error) {
	deploymentsAvailable := s.deploymentsAvailable(existingDeployments)

	defaultRoutesReady, err := s.defaultRoutesReady()
	if err != nil {
		return common.Condition{}, err
	}

	newAvailableCondition := common.Condition{
		Type:   appsv1alpha1.APIManagerAvailableConditionType,
		Status: v1.ConditionFalse,
	}

	if deploymentsAvailable && defaultRoutesReady {
		newAvailableCondition.Status = v1.ConditionTrue
	}

	return newAvailableCondition, nil
}

func (s *APIManagerStatusReconciler) defaultRoutesReady() (bool, error) {
	wildcardDomain := s.apimanagerResource.Spec.WildcardDomain
	routeHosts := []string{
		fmt.Sprintf("backend-%s.%s", *s.apimanagerResource.Spec.TenantName, wildcardDomain),                // Backend Listener route
		fmt.Sprintf("api-%s-apicast-production.%s", *s.apimanagerResource.Spec.TenantName, wildcardDomain), // Apicast Production default tenant Route
		fmt.Sprintf("api-%s-apicast-staging.%s", *s.apimanagerResource.Spec.TenantName, wildcardDomain),    // Apicast Staging default tenant Route
		fmt.Sprintf("master.%s", wildcardDomain),                                                           // System's Master Portal Route
		fmt.Sprintf("%s.%s", *s.apimanagerResource.Spec.TenantName, wildcardDomain),                        // System's default tenant Developer Portal Route
		fmt.Sprintf("%s-admin.%s", *s.apimanagerResource.Spec.TenantName, wildcardDomain),                  // System's default tenant Admin Portal Route
	}

	listOps := []client.ListOption{
		client.InNamespace(s.apimanagerResource.Namespace),
	}

	routeList := &routev1.RouteList{}
	err := s.Client().List(context.TODO(), routeList, listOps...)
	if err != nil {
		return false, fmt.Errorf("Failed to list routes: %w", err)
	}

	var routes []routev1.Route
	for _, route := range routeList.Items {
		routes = append(routes, route)
	}
	sort.Slice(routes, func(i, j int) bool { return routes[i].Name < routes[j].Name })

	allDefaultRoutesReady := true
	for _, routeHost := range routeHosts {
		routeFound := false
		var matchedRoute *routev1.Route
		for _, route := range routes {
			// TODO should we check that there are no two routes
			// with the same host field value?
			if route.Spec.Host == routeHost {
				routeFound = true
				matchedRoute = &route
			}
		}

		if !routeFound {
			allDefaultRoutesReady = false
		}

		if routeFound {
			routeReady := helper.IsRouteReady(matchedRoute)
			if !routeReady {
				allDefaultRoutesReady = false
			}
		}
	}

	return allDefaultRoutesReady, nil
}
