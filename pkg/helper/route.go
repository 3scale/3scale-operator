package helper

import (
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
)

// IsRouteReady returns true when all Ingresses of the Route have the
// "Admitted" Condition set to true
func IsRouteReady(route *routev1.Route) bool {
	routeStatusIngresses := route.Status.Ingress
	if routeStatusIngresses == nil || len(routeStatusIngresses) == 0 {
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
