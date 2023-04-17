package operator

import (
	"context"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestSystemSearchdReconciler(t *testing.T) {
	var (
		log = logf.Log.WithName("operator_test")
	)

	ctx := context.TODO()

	apimanager := testSearchdBasicApimanager()
	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.GroupVersion, apimanager)
	err := appsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	// Create a fake client to mock API calls.
	cl := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	clientAPIReader := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)

	baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

	reconciler := NewSystemSearchdReconciler(baseAPIManagerLogicReconciler)
	_, err = reconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName string
		objName  string
		obj      k8sclient.Object
	}{
		{"PVC", component.SystemSearchdPVCName, &v1.PersistentVolumeClaim{}},
		{"Service", component.SystemSearchdServiceName, &v1.Service{}},
		{"DC", component.SystemSearchdDeploymentName, &appsv1.DeploymentConfig{}},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			obj := tc.obj
			namespacedName := types.NamespacedName{
				Name:      tc.objName,
				Namespace: namespace,
			}
			err = cl.Get(context.TODO(), namespacedName, obj)
			// object must exist, that is all required to be tested
			if err != nil {
				subT.Errorf("error fetching object %s: %v", tc.objName, err)
			}
		})
	}
}

func TestUpgradeFromSphinx(t *testing.T) {
	var (
		log        = logf.Log.WithName("upgrade_test")
		ctx        = context.TODO()
		s          = scheme.Scheme
		apimanager = testSearchdBasicApimanager()
	)

	s.AddKnownTypes(appsv1alpha1.GroupVersion, apimanager)
	err := appsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName     string
		initialState []runtime.Object
	}{
		{"Start from scratch", nil},
		{"only sphinx dc/service", []runtime.Object{
			testOldSphinxDC(namespace), testOldSphinxSvc(namespace),
		}},
		{"only searchd dc/service", []runtime.Object{
			testsearchdDC(namespace), testsearchdSvc(namespace),
		}},
		{"both sphinx and searchd dc/service", []runtime.Object{
			testOldSphinxDC(namespace), testOldSphinxSvc(namespace),
			testsearchdDC(namespace), testsearchdSvc(namespace),
		}},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			objs := []runtime.Object{}
			cl := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
			clientAPIReader := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
			clientset := fakeclientset.NewSimpleClientset()
			recorder := record.NewFakeRecorder(10000)
			baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
			baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

			reconciler := NewSystemSearchdReconciler(baseAPIManagerLogicReconciler)
			_, err = reconciler.Reconcile()
			if err != nil {
				subT.Fatal(err)
			}

			// Sphinx Service should not be there
			sphinxServiceKey := client.ObjectKey{Name: "system-sphinx", Namespace: namespace}
			err = cl.Get(ctx, sphinxServiceKey, &v1.Service{})
			if err == nil {
				subT.Fatalf("reading an object expected to be deleted: %s", sphinxServiceKey)
			}
			if !errors.IsNotFound(err) {
				subT.Fatalf("unexpected error reading object %s: %v", sphinxServiceKey, err)
			}

			// Sphinx DC should not be there
			sphinxDCKey := client.ObjectKey{Name: "system-sphinx", Namespace: namespace}
			err = cl.Get(ctx, sphinxDCKey, &appsv1.DeploymentConfig{})
			if err == nil {
				subT.Fatalf("reading an object expected to be deleted: %s", sphinxDCKey)
			}
			if !errors.IsNotFound(err) {
				subT.Fatalf("unexpected error reading object %s: %v", sphinxDCKey, err)
			}

			opts := &component.SystemSearchdOptions{}
			searchd := component.NewSystemSearchd(opts)
			// Searchd Service should be there
			searchdServiceKey := client.ObjectKey{Name: searchd.Service().Name, Namespace: namespace}
			err = cl.Get(ctx, searchdServiceKey, &v1.Service{})
			if err != nil {
				subT.Fatalf("error fetching object %s: %v", searchdServiceKey, err)
			}
			// Searchd DC should be there
			searchDCKey := client.ObjectKey{Name: searchd.DeploymentConfig().Name, Namespace: namespace}
			err = cl.Get(ctx, searchDCKey, &appsv1.DeploymentConfig{})
			if err != nil {
				subT.Fatalf("error fetching object %s: %v", searchDCKey, err)
			}
		})
	}
}

func testOldSphinxDC(namespace string) *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "system-sphinx",
			Namespace: namespace,
		},
	}
}

func testOldSphinxSvc(namespace string) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "system-sphinx",
			Namespace: namespace,
		},
	}
}

func testsearchdSvc(namespace string) *v1.Service {
	opts := &component.SystemSearchdOptions{}
	svc := component.NewSystemSearchd(opts).Service()
	svc.Namespace = namespace
	return svc
}

func testsearchdDC(namespace string) *appsv1.DeploymentConfig {
	opts := &component.SystemSearchdOptions{}
	dc := component.NewSystemSearchd(opts).DeploymentConfig()
	dc.Namespace = namespace
	return dc
}
