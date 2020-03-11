package operator

import (
	"context"
	"reflect"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestPodDisruptionBudgetBaseReconcilerCreate(t *testing.T) {
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
		Spec: appsv1alpha1.APIManagerSpec{
			PodDisruptionBudget: &appsv1alpha1.PodDisruptionBudgetSpec{Enabled: true},
		},
	}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.SchemeGroupVersion, apimanager)
	err = policyv1beta1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)

	baseReconciler := NewBaseReconciler(cl, clientAPIReader, s, log, cfg)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseLogicReconciler, apimanager)

	reconciler := NewPodDisruptionBudgetReconciler(baseAPIManagerLogicReconciler)

	desired := &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myPodDisruptionBudget",
			Namespace: namespace,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"test": "mytest"},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: 1},
		},
	}

	err = reconciler.Reconcile(desired)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "myPodDisruptionBudget",
		Namespace: namespace,
	}
	reconciled := &policyv1beta1.PodDisruptionBudget{}
	err = cl.Get(context.TODO(), namespacedName, reconciled)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(desired.Spec, reconciled.Spec) {
		t.Errorf("Created PDB is not the same as desired")
	}

	if len(reconciled.GetOwnerReferences()) != 1 {
		t.Fatal("reconciled does not have owner reference")
	}

	if reconciled.GetOwnerReferences()[0].Name != name {
		t.Fatalf("reconciled owner reference is not apimanager, expected: %s, got: %s", name, reconciled.GetOwnerReferences()[0].Name)
	}
}

func TestPodDisruptionBudgetBaseReconcilerUpdate(t *testing.T) {
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
		Spec: appsv1alpha1.APIManagerSpec{
			PodDisruptionBudget: &appsv1alpha1.PodDisruptionBudgetSpec{Enabled: true},
		},
	}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.SchemeGroupVersion, apimanager)
	err = policyv1beta1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	existing := &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myPodDisruptionBudget",
			Namespace: namespace,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"test1": "mytest1"},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: 2},
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{existing}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)

	baseReconciler := NewBaseReconciler(cl, clientAPIReader, s, log, cfg)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseLogicReconciler, apimanager)

	reconciler := NewPodDisruptionBudgetReconciler(baseAPIManagerLogicReconciler)

	desired := &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myPodDisruptionBudget",
			Namespace: namespace,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"test": "mytest"},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: 1},
		},
	}

	err = reconciler.Reconcile(desired)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "myPodDisruptionBudget",
		Namespace: namespace,
	}
	reconciled := &policyv1beta1.PodDisruptionBudget{}
	err = cl.Get(context.TODO(), namespacedName, reconciled)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(desired.Spec, reconciled.Spec) {
		t.Errorf("Updated PDB is not the same as desired")
	}
}

func TestPodDisruptionBudgetBaseReconcilerDelete(t *testing.T) {
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
	err = policyv1beta1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	existing := &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myPodDisruptionBudget",
			Namespace: namespace,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"test1": "mytest1"},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: 2},
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{existing}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)

	baseReconciler := NewBaseReconciler(cl, clientAPIReader, s, log, cfg)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseLogicReconciler, apimanager)

	reconciler := NewPodDisruptionBudgetReconciler(baseAPIManagerLogicReconciler)

	desired := &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myPodDisruptionBudget",
			Namespace: namespace,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{},
	}

	err = reconciler.Reconcile(desired)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "myPodDisruptionBudget",
		Namespace: namespace,
	}
	reconciled := &policyv1beta1.PodDisruptionBudget{}
	err = cl.Get(context.TODO(), namespacedName, reconciled)
	// object should not exist, that is all required to be tested
	if !errors.IsNotFound(err) {
		t.Fatal(err)
	}
}
