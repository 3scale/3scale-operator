package operator

import (
	"context"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
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
	s.AddKnownTypes(appsv1alpha1.SchemeGroupVersion, apimanager)
	err := appsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)

	baseReconciler := reconcilers.NewBaseReconciler(cl, s, clientAPIReader, ctx, log)
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

	err = apimanagerLogicReconciler.ReconcileResource(&v1.ConfigMap{}, desiredConfigmap, reconcilers.CreateOnlyMutator)
	if err != nil {
		t.Fatal(err)
	}

	reconciledConfigmap := &v1.ConfigMap{}

	objectKey, err := client.ObjectKeyFromObject(desiredConfigmap)
	if err != nil {
		t.Fatal(err)
	}

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
