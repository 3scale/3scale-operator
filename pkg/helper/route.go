package helper

import (
	"context"
	"fmt"
	"sort"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListRoutes returns the routes in the given namespace, sorted by name.
func ListRoutes(k8sclient client.Client, namespace string, opts ...client.ListOption) ([]routev1.Route, error) {
	routeList := &routev1.RouteList{}
	listOpts := append([]client.ListOption{client.InNamespace(namespace)}, opts...)
	err := k8sclient.List(context.TODO(), routeList, listOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to list routes: %w", err)
	}
	routes := append([]routev1.Route(nil), routeList.Items...)
	sort.Slice(routes, func(i, j int) bool { return routes[i].Name < routes[j].Name })
	return routes, nil
}

// CheckRoutesReady checks whether all expectedHosts are present and admitted in routes.
// Returns (allReady, notReadyHosts).
func CheckRoutesReady(expectedHosts []string, routes []routev1.Route, logger logr.Logger) (bool, []string) {
	var notReadyHosts []string
	for _, expectedRouteHost := range expectedHosts {
		routeIdx := RouteFindByHost(routes, expectedRouteHost)
		if routeIdx == -1 {
			logger.V(1).Info("Status defaultRoutesReady: route not found", "expectedRouteHost", expectedRouteHost)
			notReadyHosts = append(notReadyHosts, expectedRouteHost)
		} else if !IsRouteReady(&routes[routeIdx]) {
			logger.V(1).Info("Status defaultRoutesReady: route not ready", "expectedRouteHost", expectedRouteHost)
			notReadyHosts = append(notReadyHosts, expectedRouteHost)
		}
	}
	return len(notReadyHosts) == 0, notReadyHosts
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
