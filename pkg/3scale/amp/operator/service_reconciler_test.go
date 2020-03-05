package operator

import (
	"context"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestServiceBaseReconcilerCreate(t *testing.T) {
	var (
		name      = "example-apimanager"
		namespace = "operator-unittest"
		log       = logf.Log.WithName("operator_test")
	)
	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatalf("Unable to get config: (%v)", err)
	}
	apimanager := &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{},
	}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.SchemeGroupVersion, apimanager)

	// Objects to track in the fake client.
	objs := []runtime.Object{}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)

	baseReconciler := NewBaseReconciler(cl, clientAPIReader, s, log, cfg)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseLogicReconciler, apimanager)
	createOnlyReconciler := NewCreateOnlySvcReconciler()

	reconciler := NewServiceBaseReconciler(baseAPIManagerLogicReconciler, createOnlyReconciler)

	desired := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mySvc",
			Namespace: namespace,
		},
	}

	err = reconciler.Reconcile(desired)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "mySvc",
		Namespace: namespace,
	}
	existing := &v1.Service{}
	err = cl.Get(context.TODO(), namespacedName, existing)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceBaseReconcilerUpdateOwnerRef(t *testing.T) {
	var (
		name      = "example-apimanager"
		namespace = "operator-unittest"
		log       = logf.Log.WithName("operator_test")
	)
	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatalf("Unable to get config: (%v)", err)
	}
	apimanager := &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{},
	}
	existing := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mySvc",
			Namespace: namespace,
		},
	}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.SchemeGroupVersion, apimanager)

	// Objects to track in the fake client.
	objs := []runtime.Object{existing}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)

	baseReconciler := NewBaseReconciler(cl, clientAPIReader, s, log, cfg)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseLogicReconciler, apimanager)
	createOnlyReconciler := NewCreateOnlySvcReconciler()

	reconciler := NewServiceBaseReconciler(baseAPIManagerLogicReconciler, createOnlyReconciler)

	desired := existing.DeepCopy()

	err = reconciler.Reconcile(desired)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "mySvc",
		Namespace: namespace,
	}
	reconciled := &v1.Service{}
	err = cl.Get(context.TODO(), namespacedName, reconciled)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Fatal(err)
	}

	if len(reconciled.GetOwnerReferences()) != 1 {
		t.Fatal("reconciled does not have owner reference")
	}

	if reconciled.GetOwnerReferences()[0].Name != name {
		t.Fatalf("reconciled owner reference is not apimanager, expected: %s, got: %s", name, reconciled.GetOwnerReferences()[0].Name)
	}
}

type myCustomSvcReconciler struct {
}

func (r *myCustomSvcReconciler) IsUpdateNeeded(desired, existing *v1.Service) bool {
	existing.Spec.Ports = []v1.ServicePort{
		v1.ServicePort{
			Name:     "http",
			Protocol: v1.ProtocolTCP,
			Port:     4000,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Type(intstr.Int),
				IntVal: 3000,
			},
		},
	}
	return true
}

func newCustomSvcReconciler() ServiceReconciler {
	return &myCustomSvcReconciler{}
}

func TestServiceBaseReconcilerUpdateNeeded(t *testing.T) {
	var (
		name      = "example-apimanager"
		namespace = "operator-unittest"
		log       = logf.Log.WithName("operator_test")
	)
	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatalf("Unable to get config: (%v)", err)
	}
	apimanager := &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{},
	}
	existing := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mySvc",
			Namespace: namespace,
		},
	}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.SchemeGroupVersion, apimanager)

	// existing does not need to be updated to set owner reference
	err = controllerutil.SetControllerReference(apimanager, existing, s)
	if err != nil {
		t.Fatal(err)
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{existing}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)

	baseReconciler := NewBaseReconciler(cl, clientAPIReader, s, log, cfg)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseLogicReconciler, apimanager)
	customReconciler := newCustomSvcReconciler()

	reconciler := NewServiceBaseReconciler(baseAPIManagerLogicReconciler, customReconciler)

	desired := existing.DeepCopy()

	err = reconciler.Reconcile(desired)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "mySvc",
		Namespace: namespace,
	}
	reconciled := &v1.Service{}
	err = cl.Get(context.TODO(), namespacedName, reconciled)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Fatal(err)
	}

	if reconciled.Spec.Ports == nil || len(reconciled.Spec.Ports) == 0 {
		t.Fatal("reconciled does not have reconciled data")
	}

	if reconciled.Spec.Ports[0].Port != 4000 {
		t.Fatalf("reconciled have reconciled data. Expected: 4000, got: %d", reconciled.Spec.Ports[0].Port)
	}
}
