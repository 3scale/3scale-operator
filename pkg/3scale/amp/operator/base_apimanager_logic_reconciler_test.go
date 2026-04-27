package operator

import (
	"context"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	hpav2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestBaseAPIManagerLogicReconcilerUpdateOwnerRef(t *testing.T) {
	var (
		apimanagerName = "example-apimanager"
		namespace      = "operator-unittest"
		log            = logf.Log.WithName("operator_test")
	)

	ctx := context.TODO()

	apimanager := &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apimanagerName,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{},
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.GroupVersion, apimanager)

	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)

	baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	apimanagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

	desiredConfigmap := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myConfigmap",
			Namespace: namespace,
		},
		Data: map[string]string{
			"somekey": "somevalue",
		},
	}

	err := apimanagerLogicReconciler.ReconcileResource(&v1.ConfigMap{}, desiredConfigmap, reconcilers.CreateOnlyMutator)
	if err != nil {
		t.Fatal(err)
	}

	reconciledConfigmap := &v1.ConfigMap{}

	objectKey := client.ObjectKeyFromObject(desiredConfigmap)

	err = cl.Get(context.TODO(), objectKey, reconciledConfigmap)
	if err != nil {
		t.Errorf("error fetching existing: %v", err)
	}

	if len(reconciledConfigmap.GetOwnerReferences()) != 1 {
		t.Errorf("reconciled obj does not have owner reference")
	}

	if reconciledConfigmap.GetOwnerReferences()[0].Name != apimanagerName {
		t.Errorf("reconciled owner reference is not apimanager, expected: %s, got: %s", apimanagerName, reconciledConfigmap.GetOwnerReferences()[0].Name)
	}
}

func TestBaseAPIManagerLogicReconcilerHasPrometheusRules(t *testing.T) {
	var (
		apimanagerName = "example-apimanager"
		namespace      = "operator-unittest"
		log            = logf.Log.WithName("operator_test")
	)

	ctx := context.TODO()
	apimanager := &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apimanagerName,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{},
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	if err := monitoringv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := grafanav1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)

	prometheusAPIResourceList := &metav1.APIResourceList{
		GroupVersion: monitoringv1.SchemeGroupVersion.String(),
		APIResources: []metav1.APIResource{
			{Name: monitoringv1.PrometheusRuleName, Namespaced: true, Kind: monitoringv1.PrometheusRuleKind},
			{Name: monitoringv1.PodMonitorName, Namespaced: true, Kind: monitoringv1.PodMonitorsKind},
			{Name: monitoringv1.ServiceMonitorName, Namespaced: false, Kind: monitoringv1.ServiceMonitorsKind},
		},
	}
	grafanaAPIResourceList := &metav1.APIResourceList{
		GroupVersion: grafanav1alpha1.GroupVersion.String(),
		APIResources: []metav1.APIResource{
			{Name: "grafanadashboards", Namespaced: true, Kind: "GrafanaDashboard"},
		},
	}

	clientset.Resources = []*metav1.APIResourceList{
		prometheusAPIResourceList,
		grafanaAPIResourceList,
	}
	baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	apimanagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

	// Test uncached request. Resource should exist
	exists, err := apimanagerLogicReconciler.HasPrometheusRules()
	if err != nil {
		t.Fatalf("Unexpected error received")
	}
	if !exists {
		t.Fatalf("Unexpected exists value received. Expected: %t, got: %t", true, exists)
	}

	// Test cached request. It should return the same results as before
	exists, err = apimanagerLogicReconciler.HasPrometheusRules()
	if err != nil {
		t.Fatalf("Unexpected error received: %s", err)
	}
	if !exists {
		t.Fatalf("Unexpected exists value received. Expected: %t, got: %t", true, exists)
	}

	// Test that we are indeed receiving cached requests by simulating
	// the removal of CRD types. We now expect to still receive that the
	// resource exists even when we've removed it from the defined CRDs because
	// the cache should be working and not seeing the new change.
	clientset.Resources = []*metav1.APIResourceList{
		grafanaAPIResourceList,
	}
	exists, err = apimanagerLogicReconciler.HasPrometheusRules()
	if err != nil {
		t.Fatalf("Unexpected error received: %s", err)
	}
	if !exists {
		t.Fatalf("Unexpected exists value received. Expected: %t, got: %t", true, exists)
	}

	// Create a new APIManagerLogicReconciler to simulate a new state of cache
	// with the resource now removed. We now should receive that it does not
	// exist
	apimanagerLogicReconciler = NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)
	exists, err = apimanagerLogicReconciler.HasPrometheusRules()
	if err != nil {
		t.Fatalf("Unexpected error received: %s", err)
	}
	if exists {
		t.Fatalf("Unexpected exists value received. Expected: %t, got: %t", false, exists)
	}
}

func TestBaseAPIManagerLogicReconcilerHasGrafanaDashboards(t *testing.T) {
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	if err := appsv1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}

	if err := monitoringv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := grafanav1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}

	var (
		apimanagerName = "example-apimanager"
		namespace      = "operator-unittest"
		log            = logf.Log.WithName("operator_test")
	)

	ctx := context.TODO()
	apimanager := &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apimanagerName,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)

	prometheusAPIResourceList := &metav1.APIResourceList{
		GroupVersion: monitoringv1.SchemeGroupVersion.String(),
		APIResources: []metav1.APIResource{
			{Name: monitoringv1.PrometheusRuleName, Namespaced: true, Kind: monitoringv1.PrometheusRuleKind},
			{Name: monitoringv1.PodMonitorName, Namespaced: true, Kind: monitoringv1.PodMonitorsKind},
			{Name: monitoringv1.ServiceMonitorName, Namespaced: false, Kind: monitoringv1.ServiceMonitorsKind},
		},
	}
	grafanaAPIResourceList := &metav1.APIResourceList{
		GroupVersion: grafanav1alpha1.GroupVersion.String(),
		APIResources: []metav1.APIResource{
			{Name: "grafanadashboards", Namespaced: true, Kind: "GrafanaDashboard"},
		},
	}

	clientset.Resources = []*metav1.APIResourceList{
		prometheusAPIResourceList,
		grafanaAPIResourceList,
	}
	baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	apimanagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

	// Test uncached request. Resource should exist
	exists, err := apimanagerLogicReconciler.HasGrafanaDashboards()
	if err != nil {
		t.Fatalf("Unexpected error received")
	}
	if !exists {
		t.Fatalf("Unexpected exists value received. Expected: %t, got: %t", true, exists)
	}

	// Test cached request. It should return the same results as before
	exists, err = apimanagerLogicReconciler.HasGrafanaDashboards()
	if err != nil {
		t.Fatalf("Unexpected error received: %s", err)
	}
	if !exists {
		t.Fatalf("Unexpected exists value received. Expected: %t, got: %t", true, exists)
	}

	// Test that we are indeed receiving cached requests by simulating
	// the removal of CRD types. We now expect to still receive that the
	// resource exists even when we've removed it from the defined CRDs because
	// the cache should be working and not seeing the new change.
	clientset.Resources = []*metav1.APIResourceList{
		prometheusAPIResourceList,
	}
	exists, err = apimanagerLogicReconciler.HasGrafanaDashboards()
	if err != nil {
		t.Fatalf("Unexpected error received: %s", err)
	}
	if !exists {
		t.Fatalf("Unexpected exists value received. Expected: %t, got: %t", true, exists)
	}

	// Create a new APIManagerLogicReconciler to simulate a new state of cache
	// with the resource now removed. We now should receive that it does not
	// exist
	apimanagerLogicReconciler = NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)
	exists, err = apimanagerLogicReconciler.HasGrafanaV4Dashboards()
	if err != nil {
		t.Fatalf("Unexpected error received: %s", err)
	}
	if exists {
		t.Fatalf("Unexpected exists value received. Expected: %t, got: %t", false, exists)
	}
}

func TestBaseAPIManagerLogicReconcilerHasPodMonitors(t *testing.T) {
	var (
		apimanagerName = "example-apimanager"
		namespace      = "operator-unittest"
		log            = logf.Log.WithName("operator_test")
	)

	ctx := context.TODO()
	apimanager := &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apimanagerName,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{},
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	if err := monitoringv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := grafanav1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)

	prometheusAPIResourceList := &metav1.APIResourceList{
		GroupVersion: monitoringv1.SchemeGroupVersion.String(),
		APIResources: []metav1.APIResource{
			{Name: monitoringv1.PrometheusRuleName, Namespaced: true, Kind: monitoringv1.PrometheusRuleKind},
			{Name: monitoringv1.PodMonitorName, Namespaced: true, Kind: monitoringv1.PodMonitorsKind},
			{Name: monitoringv1.ServiceMonitorName, Namespaced: false, Kind: monitoringv1.ServiceMonitorsKind},
		},
	}
	grafanaAPIResourceList := &metav1.APIResourceList{
		GroupVersion: grafanav1alpha1.GroupVersion.String(),
		APIResources: []metav1.APIResource{
			{Name: "grafanadashboards", Namespaced: true, Kind: "GrafanaDashboard"},
		},
	}

	clientset.Resources = []*metav1.APIResourceList{
		prometheusAPIResourceList,
		grafanaAPIResourceList,
	}
	baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	apimanagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

	// Test uncached request. Resource should exist
	exists, err := apimanagerLogicReconciler.HasPodMonitors()
	if err != nil {
		t.Fatalf("Unexpected error received")
	}
	if !exists {
		t.Fatalf("Unexpected exists value received. Expected: %t, got: %t", true, exists)
	}

	// Test cached request. It should return the same results as before
	exists, err = apimanagerLogicReconciler.HasPodMonitors()
	if err != nil {
		t.Fatalf("Unexpected error received: %s", err)
	}
	if !exists {
		t.Fatalf("Unexpected exists value received. Expected: %t, got: %t", true, exists)
	}

	// Test that we are indeed receiving cached requests by simulating
	// the removal of CRD types. We now expect to still receive that the
	// resource exists even when we've removed it from the defined CRDs because
	// the cache should be working and not seeing the new change.
	clientset.Resources = []*metav1.APIResourceList{
		grafanaAPIResourceList,
	}
	exists, err = apimanagerLogicReconciler.HasPodMonitors()
	if err != nil {
		t.Fatalf("Unexpected error received: %s", err)
	}
	if !exists {
		t.Fatalf("Unexpected exists value received. Expected: %t, got: %t", true, exists)
	}

	// Create a new APIManagerLogicReconciler to simulate a new state of cache
	// with the resource now removed. We now should receive that it does not
	// exist
	apimanagerLogicReconciler = NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)
	exists, err = apimanagerLogicReconciler.HasPodMonitors()
	if err != nil {
		t.Fatalf("Unexpected error received: %s", err)
	}
	if exists {
		t.Fatalf("Unexpected exists value received. Expected: %t, got: %t", false, exists)
	}
}

func newTestReconciler(t *testing.T, apimanager *appsv1alpha1.APIManager, existingObjs ...runtime.Object) (*BaseAPIManagerLogicReconciler, client.Client) {
	t.Helper()
	s := scheme.Scheme
	if err := appsv1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}

	objs := append([]runtime.Object{apimanager}, existingObjs...)
	cl := fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(objs...).Build()
	clientAPIReader := fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(objs...).Build()
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)
	log := logf.Log.WithName("operator_test")

	baseReconciler := reconcilers.NewBaseReconciler(context.Background(), cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	return NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager), cl
}

func TestReconcileHpa(t *testing.T) {
	baseAM := func(annotations map[string]string) *appsv1alpha1.APIManager {
		am := &appsv1alpha1.APIManager{
			ObjectMeta: metav1.ObjectMeta{
				Name:        apimanagerName,
				Namespace:   namespace,
				Annotations: annotations,
			},
			Spec: appsv1alpha1.APIManagerSpec{
				APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
					WildcardDomain: wildcardDomain,
				},
			},
		}
		if _, err := am.SetDefaults(); err != nil {
			panic(err)
		}
		return am
	}

	hpaExists := func(t *testing.T, cl client.Client, name string) bool {
		t.Helper()
		obj := &hpav2.HorizontalPodAutoscaler{}
		err := cl.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, obj)
		if k8serrors.IsNotFound(err) {
			return false
		}
		if err != nil {
			t.Fatalf("unexpected error checking HPA %s: %v", name, err)
		}
		return true
	}

	tests := []struct {
		name        string
		apimanager  *appsv1alpha1.APIManager
		existingHPA *hpav2.HorizontalPodAutoscaler
		desired     *hpav2.HorizontalPodAutoscaler
		assert      func(t *testing.T, cl client.Client)
	}{
		{
			name:       "disabled with no existing HPA is a no-op",
			apimanager: baseAM(nil),
			desired:    component.DefaultHpa(component.BackendListenerName, namespace),
			assert: func(t *testing.T, cl client.Client) {
				if hpaExists(t, cl, component.BackendListenerName) {
					t.Error("HPA should not exist when disabled and was not pre-existing")
				}
			},
		},
		{
			name:        "disabled with existing HPA deletes it",
			apimanager:  baseAM(nil),
			existingHPA: component.DefaultHpa(component.BackendListenerName, namespace),
			desired:     component.DefaultHpa(component.BackendListenerName, namespace),
			assert: func(t *testing.T, cl client.Client) {
				if hpaExists(t, cl, component.BackendListenerName) {
					t.Error("HPA should have been deleted when disabled")
				}
			},
		},
		{
			name: "backend listener HPA enabled creates HPA",
			apimanager: func() *appsv1alpha1.APIManager {
				am := baseAM(nil)
				am.Spec.Backend.ListenerSpec.Hpa = true
				return am
			}(),
			desired: component.DefaultHpa(component.BackendListenerName, namespace),
			assert: func(t *testing.T, cl client.Client) {
				if !hpaExists(t, cl, component.BackendListenerName) {
					t.Error("HPA should have been created when enabled")
				}
			},
		},
		{
			name: "backend listener HPA enabled with async-disable annotation deletes HPA",
			apimanager: func() *appsv1alpha1.APIManager {
				am := baseAM(map[string]string{appsv1alpha1.DisableAsyncAnnotation: "true"})
				am.Spec.Backend.ListenerSpec.Hpa = true
				return am
			}(),
			existingHPA: component.DefaultHpa(component.BackendListenerName, namespace),
			desired:     component.DefaultHpa(component.BackendListenerName, namespace),
			assert: func(t *testing.T, cl client.Client) {
				if hpaExists(t, cl, component.BackendListenerName) {
					t.Error("HPA should have been deleted when async is disabled regardless of HPA flag")
				}
			},
		},
		{
			name: "backend worker HPA enabled creates HPA",
			apimanager: func() *appsv1alpha1.APIManager {
				am := baseAM(nil)
				am.Spec.Backend.WorkerSpec.Hpa = true
				return am
			}(),
			desired: component.DefaultHpa(component.BackendWorkerName, namespace),
			assert: func(t *testing.T, cl client.Client) {
				if !hpaExists(t, cl, component.BackendWorkerName) {
					t.Error("HPA should have been created when enabled")
				}
			},
		},
		{
			name: "apicast HPA enabled with async-disable annotation creates HPA",
			apimanager: func() *appsv1alpha1.APIManager {
				am := baseAM(map[string]string{appsv1alpha1.DisableAsyncAnnotation: "true"})
				am.Spec.Apicast.ProductionSpec.Hpa = true
				return am
			}(),
			desired: component.DefaultHpa(component.ApicastProductionName, namespace),
			assert: func(t *testing.T, cl client.Client) {
				if !hpaExists(t, cl, component.ApicastProductionName) {
					t.Error("Apicast HPA should be created regardless of async-disable annotation")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var existingObjs []runtime.Object
			if tt.existingHPA != nil {
				existingObjs = append(existingObjs, tt.existingHPA)
			}

			r, cl := newTestReconciler(t, tt.apimanager, existingObjs...)

			if err := r.ReconcileHpa(tt.desired, reconcilers.CreateOnlyMutator); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tt.assert(t, cl)
		})
	}
}

func TestBaseAPIManagerLogicReconcilerHasServiceMonitors(t *testing.T) {
	var (
		apimanagerName = "example-apimanager"
		namespace      = "operator-unittest"
		log            = logf.Log.WithName("operator_test")
	)

	ctx := context.TODO()
	apimanager := &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apimanagerName,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{},
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	if err := monitoringv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := grafanav1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)

	prometheusAPIResourceList := &metav1.APIResourceList{
		GroupVersion: monitoringv1.SchemeGroupVersion.String(),
		APIResources: []metav1.APIResource{
			{Name: monitoringv1.PrometheusRuleName, Namespaced: true, Kind: monitoringv1.PrometheusRuleKind},
			{Name: monitoringv1.PodMonitorName, Namespaced: true, Kind: monitoringv1.PodMonitorsKind},
			{Name: monitoringv1.ServiceMonitorName, Namespaced: false, Kind: monitoringv1.ServiceMonitorsKind},
		},
	}
	grafanaAPIResourceList := &metav1.APIResourceList{
		GroupVersion: grafanav1alpha1.GroupVersion.String(),
		APIResources: []metav1.APIResource{
			{Name: "grafanadashboards", Namespaced: true, Kind: "GrafanaDashboard"},
		},
	}

	clientset.Resources = []*metav1.APIResourceList{
		prometheusAPIResourceList,
		grafanaAPIResourceList,
	}
	baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	apimanagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

	// Test uncached request. Resource should exist
	exists, err := apimanagerLogicReconciler.HasServiceMonitors()
	if err != nil {
		t.Fatalf("Unexpected error received")
	}
	if !exists {
		t.Fatalf("Unexpected exists value received. Expected: %t, got: %t", true, exists)
	}

	// Test cached request. It should return the same results as before
	exists, err = apimanagerLogicReconciler.HasServiceMonitors()
	if err != nil {
		t.Fatalf("Unexpected error received: %s", err)
	}
	if !exists {
		t.Fatalf("Unexpected exists value received. Expected: %t, got: %t", true, exists)
	}

	// Test that we are indeed receiving cached requests by simulating
	// the removal of CRD types. We now expect to still receive that the
	// resource exists even when we've removed it from the defined CRDs because
	// the cache should be working and not seeing the new change.
	clientset.Resources = []*metav1.APIResourceList{
		grafanaAPIResourceList,
	}
	exists, err = apimanagerLogicReconciler.HasServiceMonitors()
	if err != nil {
		t.Fatalf("Unexpected error received: %s", err)
	}
	if !exists {
		t.Fatalf("Unexpected exists value received. Expected: %t, got: %t", true, exists)
	}

	// Create a new APIManagerLogicReconciler to simulate a new state of cache
	// with the resource now removed. We now should receive that it does not
	// exist
	apimanagerLogicReconciler = NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)
	exists, err = apimanagerLogicReconciler.HasServiceMonitors()
	if err != nil {
		t.Fatalf("Unexpected error received: %s", err)
	}
	if exists {
		t.Fatalf("Unexpected exists value received. Expected: %t, got: %t", false, exists)
	}
}
