package helper

import (
	"testing"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makeRoute(name, host string, admitted bool) routev1.Route {
	condStatus := corev1.ConditionFalse
	if admitted {
		condStatus = corev1.ConditionTrue
	}
	return routev1.Route{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       routev1.RouteSpec{Host: host},
		Status: routev1.RouteStatus{
			Ingress: []routev1.RouteIngress{
				{Conditions: []routev1.RouteIngressCondition{
					{Type: routev1.RouteAdmitted, Status: condStatus},
				}},
			},
		},
	}
}

func TestCheckRoutesReady(t *testing.T) {
	backendHost := "backend-3scale.example.com"
	apicastProdHost := "api-3scale-apicast-production.example.com"
	apicastStagHost := "api-3scale-apicast-staging.example.com"
	masterHost := "master.example.com"
	devPortalHost := "3scale.example.com"
	adminHost := "3scale-admin.example.com"

	allHosts := []string{backendHost, apicastProdHost, apicastStagHost, masterHost, devPortalHost, adminHost}
	allRoutes := []routev1.Route{
		makeRoute("backend", backendHost, true),
		makeRoute("apicast-prod", apicastProdHost, true),
		makeRoute("apicast-stag", apicastStagHost, true),
		makeRoute("master", masterHost, true),
		makeRoute("dev-portal", devPortalHost, true),
		makeRoute("admin", adminHost, true),
	}

	tests := []struct {
		name                 string
		expectedHosts        []string
		routes               []routev1.Route
		wantReady            bool
		wantNotReadyContains []string
	}{
		{
			name:          "all routes ready",
			expectedHosts: allHosts,
			routes:        allRoutes,
			wantReady:     true,
		},
		{
			name:          "one route not admitted",
			expectedHosts: allHosts,
			routes: []routev1.Route{
				makeRoute("backend", backendHost, true),
				makeRoute("apicast-prod", apicastProdHost, false),
				makeRoute("apicast-stag", apicastStagHost, true),
				makeRoute("master", masterHost, true),
				makeRoute("dev-portal", devPortalHost, true),
				makeRoute("admin", adminHost, true),
			},
			wantReady:            false,
			wantNotReadyContains: []string{apicastProdHost},
		},
		{
			name:          "one route missing",
			expectedHosts: allHosts,
			routes: []routev1.Route{
				makeRoute("backend", backendHost, true),
				makeRoute("apicast-stag", apicastStagHost, true),
				makeRoute("master", masterHost, true),
				makeRoute("dev-portal", devPortalHost, true),
				makeRoute("admin", adminHost, true),
			},
			wantReady:            false,
			wantNotReadyContains: []string{apicastProdHost},
		},
		{
			name:          "multiple routes not ready",
			expectedHosts: allHosts,
			routes: []routev1.Route{
				makeRoute("backend", backendHost, false),
				makeRoute("apicast-prod", apicastProdHost, false),
				makeRoute("apicast-stag", apicastStagHost, true),
				makeRoute("master", masterHost, true),
				makeRoute("dev-portal", devPortalHost, true),
				makeRoute("admin", adminHost, true),
			},
			wantReady:            false,
			wantNotReadyContains: []string{backendHost, apicastProdHost},
		},
		{
			name:          "only backend route required, ready",
			expectedHosts: []string{backendHost},
			routes:        []routev1.Route{makeRoute("backend", backendHost, true)},
			wantReady:     true,
		},
		{
			name:                 "only backend route required, not admitted",
			expectedHosts:        []string{backendHost},
			routes:               []routev1.Route{makeRoute("backend", backendHost, false)},
			wantReady:            false,
			wantNotReadyContains: []string{backendHost},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ready, notReadyHosts := CheckRoutesReady(tt.expectedHosts, tt.routes, logr.Discard())

			if ready != tt.wantReady {
				t.Errorf("ready = %v, want %v", ready, tt.wantReady)
			}
			if len(notReadyHosts) != len(tt.wantNotReadyContains) {
				t.Errorf("len(notReadyHosts) = %d, want %d; hosts: %v", len(notReadyHosts), len(tt.wantNotReadyContains), notReadyHosts)
			}
			for _, expected := range tt.wantNotReadyContains {
				found := false
				for _, h := range notReadyHosts {
					if h == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected %q in notReadyHosts %v", expected, notReadyHosts)
				}
			}
		})
	}
}
