package operator

import (
	"context"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestAMPImagesReconciler(t *testing.T) {
	var (
		name           = "example-apimanager"
		namespace      = "operator-unittest"
		wildcardDomain = "test.3scale.net"
		log            = logf.Log.WithName("operator_test")
		appLabel       = "someLabel"
		trueValue      = true
	)

	ctx := context.TODO()

	apimanager := &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				AppLabel:                     &appLabel,
				ImageStreamTagImportInsecure: &trueValue,
				WildcardDomain:               wildcardDomain,
			},
		},
	}
	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.GroupVersion, apimanager)
	err := appsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	err = imagev1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	err = routev1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)
	recorder := record.NewFakeRecorder(10000)
	clientset := fakeclientset.NewSimpleClientset()

	baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

	imagesReconciler := NewAMPImagesReconciler(baseAPIManagerLogicReconciler)
	_, err = imagesReconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName string
		objName  string
		obj      runtime.Object
	}{
		{"backendCreated", "amp-backend", &imagev1.ImageStream{}},
		{"zyncCreated", "amp-zync", &imagev1.ImageStream{}},
		{"apicastCreated", "amp-apicast", &imagev1.ImageStream{}},
		{"systemCreated", "amp-system", &imagev1.ImageStream{}},
		{"zyncPostgresqlCreated", "zync-database-postgresql", &imagev1.ImageStream{}},
		{"systemMemcachedCreated", "system-memcached", &imagev1.ImageStream{}},
		// TODO: service account created by AMPImagesReconciler. Should not be there.
		{"serviceAccountCreated", "amp", &v1.ServiceAccount{}},
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

func TestAMPImagesReconcilerWithAllExternalDatabases(t *testing.T) {
	var (
		name           = "example-apimanager"
		namespace      = "operator-unittest"
		wildcardDomain = "test.3scale.net"
		log            = logf.Log.WithName("operator_test")
		appLabel       = "someLabel"
		trueValue      = true
	)

	ctx := context.TODO()

	apimanager := &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				AppLabel:                     &appLabel,
				ImageStreamTagImportInsecure: &trueValue,
				WildcardDomain:               wildcardDomain,
			},
			HighAvailability: &appsv1alpha1.HighAvailabilitySpec{
				Enabled:                     true,
				ExternalZyncDatabaseEnabled: &trueValue,
			},
			ExternalComponents: appsv1alpha1.AllComponentsExternal(),
		},
	}
	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.GroupVersion, apimanager)
	err := appsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	err = imagev1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	err = routev1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)
	recorder := record.NewFakeRecorder(10000)
	clientset := fakeclientset.NewSimpleClientset()

	baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

	imagesReconciler := NewAMPImagesReconciler(baseAPIManagerLogicReconciler)
	_, err = imagesReconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName   string
		objName    string
		obj        runtime.Object
		hasToExist bool
	}{
		{"backendCreated", "amp-backend", &imagev1.ImageStream{}, true},
		{"zyncCreated", "amp-zync", &imagev1.ImageStream{}, true},
		{"apicastCreated", "amp-apicast", &imagev1.ImageStream{}, true},
		{"systemCreated", "amp-system", &imagev1.ImageStream{}, true},
		{"zyncPostgresqlCreated", "zync-database-postgresql", &imagev1.ImageStream{}, false},
		{"systemMemcachedCreated", "system-memcached", &imagev1.ImageStream{}, true},
		// TODO: service account created by AMPImagesReconciler. Should not be there.
		{"serviceAccountCreated", "amp", &v1.ServiceAccount{}, true},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			obj := tc.obj
			namespacedName := types.NamespacedName{
				Name:      tc.objName,
				Namespace: namespace,
			}
			err = cl.Get(context.TODO(), namespacedName, obj)
			if tc.hasToExist {
				if err != nil {
					subT.Errorf("error fetching object %s: %v", tc.objName, err)
				}
			} else {
				if err == nil || !errors.IsNotFound(err) {
					subT.Errorf("object %s that shouldn't exist exists or different error than NotFound returned: %v", tc.objName, err)
				}
			}

		})
	}
}
