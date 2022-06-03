package operator

import (
	"context"
	"testing"

	appsv1beta1 "github.com/3scale/3scale-operator/apis/apps/v1beta1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	imagev1 "github.com/openshift/api/image/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestSystemMySQLImageReconciler(t *testing.T) {
	var (
		appLabel  = "someLabel"
		name      = "example-apimanager"
		namespace = "operator-unittest"
		trueValue = true
		imageURL  = "mysql:test"
		log       = logf.Log.WithName("operator_test")
	)

	ctx := context.TODO()

	apimanager := &appsv1beta1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1beta1.APIManagerSpec{
			APIManagerCommonSpec: appsv1beta1.APIManagerCommonSpec{
				AppLabel:                     &appLabel,
				ImageStreamTagImportInsecure: &trueValue,
			},
			System: &appsv1beta1.SystemSpec{
				DatabaseSpec: &appsv1beta1.SystemDatabaseSpec{
					MySQL: &appsv1beta1.SystemMySQLSpec{
						Image: &imageURL,
					},
				},
			},
		},
	}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1beta1.GroupVersion, apimanager)
	err := imagev1.AddToScheme(s)
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

	reconciler := NewSystemMySQLImageReconciler(baseAPIManagerLogicReconciler)
	_, err = reconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	obj := &imagev1.ImageStream{}

	namespacedName := types.NamespacedName{
		Name:      "system-mysql",
		Namespace: namespace,
	}
	err = cl.Get(context.TODO(), namespacedName, obj)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Fatal(err)
	}
}
