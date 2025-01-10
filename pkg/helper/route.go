package helper

import (
	"context"
	"fmt"
	"sort"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DefaultRoutesReady returns true when the expected default routes are available
func DefaultRoutesReady(apimanager *appsv1alpha1.APIManager, k8sclient client.Client, logger logr.Logger) (bool, error) {
	var expectedRouteHosts []string
	wildcardDomain := apimanager.Spec.WildcardDomain
	if apimanager.Spec.TenantName != nil {
		expectedRouteHosts = []string{
			fmt.Sprintf("backend-%s.%s", *apimanager.Spec.TenantName, wildcardDomain), // Backend Listener route
		}
		if apimanager.IsZyncEnabled() {
			zyncRoutes := []string{
				fmt.Sprintf("api-%s-apicast-production.%s", *apimanager.Spec.TenantName, wildcardDomain), // Apicast Production default tenant Route
				fmt.Sprintf("api-%s-apicast-staging.%s", *apimanager.Spec.TenantName, wildcardDomain),    // Apicast Staging default tenant Route
				fmt.Sprintf("master.%s", wildcardDomain),                                                 // System's Master Portal Route
				fmt.Sprintf("%s.%s", *apimanager.Spec.TenantName, wildcardDomain),                        // System's default tenant Developer Portal Route
				fmt.Sprintf("%s-admin.%s", *apimanager.Spec.TenantName, wildcardDomain),                  // System's default tenant Admin Portal Route
			}
			expectedRouteHosts = append(expectedRouteHosts, zyncRoutes...)
		}
	} else {
		return false, nil
	}

	listOps := []client.ListOption{
		client.InNamespace(apimanager.Namespace),
	}

	routeList := &routev1.RouteList{}
	err := k8sclient.List(context.TODO(), routeList, listOps...)
	if err != nil {
		return false, fmt.Errorf("failed to list routes: %w", err)
	}

	routes := append([]routev1.Route(nil), routeList.Items...)
	sort.Slice(routes, func(i, j int) bool { return routes[i].Name < routes[j].Name })

	allDefaultRoutesReady := true
	for _, expectedRouteHost := range expectedRouteHosts {
		routeIdx := RouteFindByHost(routes, expectedRouteHost)
		if routeIdx == -1 {
			logger.V(1).Info("Status defaultRoutesReady: route not found", "expectedRouteHost", expectedRouteHost)
			allDefaultRoutesReady = false
		} else {
			matchedRoute := &routes[routeIdx]
			routeReady := IsRouteReady(matchedRoute)
			if !routeReady {
				logger.V(1).Info("Status defaultRoutesReady: route not ready", "expectedRouteHost", expectedRouteHost)
				allDefaultRoutesReady = false
			}
		}
	}

	return allDefaultRoutesReady, nil
}

// IsRouteReady returns true when all Ingresses of the Route have the
// "Admitted" Condition set to true
func IsRouteReady(route *routev1.Route) bool {
	routeStatusIngresses := route.Status.Ingress
	if len(routeStatusIngresses) == 0 {
		return false
	}

	for _, routeStatusIngress := range routeStatusIngresses {
		routeStatusIngressConditions := routeStatusIngress.Conditions
		isReady := false
		for _, routeStatusIngressCondition := range routeStatusIngressConditions {
			if routeStatusIngressCondition.Type == routev1.RouteAdmitted && routeStatusIngressCondition.Status == v1.ConditionTrue {
				isReady = true
				break
			}
		}
		if !isReady {
			return false
		}
	}

	return true
}

// RouteFindByHost returns the smallest index i at which a route with a given host is found
// or -1 if there is no such index.
func RouteFindByHost(a []routev1.Route, routeHost string) int {
	for i, n := range a {
		if n.Spec.Host == routeHost {
			return i
		}
	}
	return -1
}
