package operator

import (
	"context"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	k8sappsv1 "k8s.io/api/apps/v1"

	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/api/core/v1"
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

func TestRedisBackendDCReconcilerCreate(t *testing.T) {
	var (
		appLabel       = "someLabel"
		name           = "example-apimanager"
		namespace      = "operator-unittest"
		trueValue      = true
		wildcardDomain = "test.3scale.net"
		tenantName     = "someTenant"
		log            = logf.Log.WithName("operator_test")
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
				ResourceRequirementsEnabled:  &trueValue,
				WildcardDomain:               wildcardDomain,
				TenantName:                   &tenantName,
			},
		},
	}
	_, err := apimanager.SetDefaults()
	if err != nil {
		t.Fatal(err)
	}

	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.GroupVersion, apimanager)
	err = imagev1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	err = k8sappsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	// 3scale 2.14 -> 2.15
	err = appsv1.Install(s)
	if err != nil {
		t.Fatal(err)
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)

	baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

	cases := []struct {
		testName              string
		reconcilerConstructor DependencyReconcilerConstructor
		expectedObjs          []struct {
			objName string
			obj     client.Object
		}
	}{
		{"backendRedis", NewBackendRedisDependencyReconciler, []struct {
			objName string
			obj     client.Object
		}{
			{"backend-redis", &k8sappsv1.Deployment{}},
			{"backend-redis", &v1.Service{}},
			{"redis-config", &v1.ConfigMap{}},
			{"backend-redis-storage", &v1.PersistentVolumeClaim{}},
			{"backend-redis", &imagev1.ImageStream{}},
		}},
		{"systemRedis", NewSystemRedisDependencyReconciler, []struct {
			objName string
			obj     client.Object
		}{
			{"system-redis", &k8sappsv1.Deployment{}},
			{"system-redis-storage", &v1.PersistentVolumeClaim{}},
			{"system-redis", &imagev1.ImageStream{}},
			{"system-redis", &v1.Service{}},
		}},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			reconciler := tc.reconcilerConstructor(baseAPIManagerLogicReconciler)
			_, err := reconciler.Reconcile()
			if err != nil {
				subT.Fatal(err)
			}

			for _, obj := range tc.expectedObjs {
				namespacedName := types.NamespacedName{
					Name:      obj.objName,
					Namespace: namespace,
				}
				err = cl.Get(context.TODO(), namespacedName, obj.obj)
				// object must exist, that is all required to be tested
				if err != nil {
					subT.Errorf("error fetching object %s: %v", obj.objName, err)
				}
			}
		})
	}
}
