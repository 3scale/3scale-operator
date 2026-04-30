package controllers

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
)

func TestZyncRouteToAPIManagerMapperMap(t *testing.T) {
	const namespace = "test-ns"
	const otherNamespace = "other-ns"

	am := getTestAPIManager(namespace)
	am.Name = "main-apimanager"

	amOtherNs := getTestAPIManager(otherNamespace)
	amOtherNs.Name = "other-apimanager"

	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.GroupVersion, am, &appsv1alpha1.APIManagerList{})
	if err := routev1.Install(s); err != nil {
		t.Fatal(err)
	}

	newRoute := func(host, ns string) *routev1.Route {
		return &routev1.Route{
			TypeMeta:   metav1.TypeMeta{Kind: "Route", APIVersion: "route.openshift.io/v1"},
			ObjectMeta: metav1.ObjectMeta{Name: "a-route", Namespace: ns},
			Spec:       routev1.RouteSpec{Host: host},
		}
	}

	enqueuesMainAM := []reconcile.Request{
		{NamespacedName: types.NamespacedName{Name: "main-apimanager", Namespace: namespace}},
	}

	tests := []struct {
		name    string
		objects []runtime.Object
		route   *routev1.Route
		want    []reconcile.Request
	}{
		// Per-domain cases: each expected tenant host must trigger a reconcile.
		{
			name:    "backend listener route host enqueues APIManager",
			objects: []runtime.Object{am},
			route:   newRoute("backend-3scale.test.example.com", namespace),
			want:    enqueuesMainAM,
		},
		{
			name:    "apicast production route host enqueues APIManager",
			objects: []runtime.Object{am},
			route:   newRoute("api-3scale-apicast-production.test.example.com", namespace),
			want:    enqueuesMainAM,
		},
		{
			name:    "apicast staging route host enqueues APIManager",
			objects: []runtime.Object{am},
			route:   newRoute("api-3scale-apicast-staging.test.example.com", namespace),
			want:    enqueuesMainAM,
		},
		{
			name:    "master portal route host enqueues APIManager",
			objects: []runtime.Object{am},
			route:   newRoute("master.test.example.com", namespace),
			want:    enqueuesMainAM,
		},
		{
			name:    "developer portal route host enqueues APIManager",
			objects: []runtime.Object{am},
			route:   newRoute("3scale.test.example.com", namespace),
			want:    enqueuesMainAM,
		},
		{
			name:    "admin portal route host enqueues APIManager",
			objects: []runtime.Object{am},
			route:   newRoute("3scale-admin.test.example.com", namespace),
			want:    enqueuesMainAM,
		},
		{
			name:    "route host does not match any APIManager — no requests",
			objects: []runtime.Object{am},
			route:   newRoute("unrelated.example.com", namespace),
			want:    nil,
		},
		{
			name:    "no APIManagers in namespace — no requests",
			objects: []runtime.Object{},
			route:   newRoute("backend-3scale.test.example.com", namespace),
			want:    nil,
		},
		{
			name:    "route in different namespace matches only that namespace's APIManager",
			objects: []runtime.Object{am, amOtherNs},
			route:   newRoute("backend-3scale.test.example.com", otherNamespace),
			want: []reconcile.Request{
				{NamespacedName: types.NamespacedName{Name: "other-apimanager", Namespace: otherNamespace}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(tt.objects...).Build()
			mapper := &ZyncRouteToAPIManagerMapper{
				Context:   context.TODO(),
				K8sClient: cl,
				Logger:    logr.Discard(),
			}

			got := mapper.Map(context.TODO(), tt.route)

			if len(got) != len(tt.want) {
				t.Fatalf("Map() returned %d requests, want %d: got %v", len(got), len(tt.want), got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("Map()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
