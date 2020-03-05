package operator

import (
	"context"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestRoleBindingBaseReconcilerCreate(t *testing.T) {
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
	err = rbacv1.AddToScheme(s)
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
	createOnlyReconciler := NewCreateOnlyRoleBindingReconciler()

	reconciler := NewRoleBindingBaseReconciler(baseAPIManagerLogicReconciler, createOnlyReconciler)

	desired := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myRoleBinding",
			Namespace: namespace,
		},
	}

	err = reconciler.Reconcile(desired)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "myRoleBinding",
		Namespace: namespace,
	}
	existing := &rbacv1.RoleBinding{}
	err = cl.Get(context.TODO(), namespacedName, existing)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Fatal(err)
	}
}

func TestRoleBindingBaseReconcilerUpdateOwnerRef(t *testing.T) {
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

	existing := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myRoleBinding",
			Namespace: namespace,
		},
	}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.SchemeGroupVersion, apimanager)
	err = rbacv1.AddToScheme(s)
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
	createOnlyReconciler := NewCreateOnlyRoleBindingReconciler()

	reconciler := NewRoleBindingBaseReconciler(baseAPIManagerLogicReconciler, createOnlyReconciler)

	desired := existing.DeepCopy()

	err = reconciler.Reconcile(desired)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "myRoleBinding",
		Namespace: namespace,
	}
	reconciled := &rbacv1.RoleBinding{}
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

type myCustomRoleBindingReconciler struct {
}

func (r *myCustomRoleBindingReconciler) IsUpdateNeeded(desired, existing *rbacv1.RoleBinding) bool {
	existing.RoleRef = rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     "myCustomRole",
	}
	return true
}

func newCustomRoleBindingReconciler() RoleBindingReconciler {
	return &myCustomRoleBindingReconciler{}
}

func TestRoleBindingBaseReconcilerUpdateNeeded(t *testing.T) {
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
	existing := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myRoleBinding",
			Namespace: namespace,
		},
	}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.SchemeGroupVersion, apimanager)
	err = rbacv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
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
	reconciler := NewRoleBindingBaseReconciler(baseAPIManagerLogicReconciler, newCustomRoleBindingReconciler())
	desired := existing.DeepCopy()
	err = reconciler.Reconcile(desired)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "myRoleBinding",
		Namespace: namespace,
	}
	reconciled := &rbacv1.RoleBinding{}
	err = cl.Get(context.TODO(), namespacedName, reconciled)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Fatal(err)
	}

	if reconciled.RoleRef.Name != "myCustomRole" {
		t.Fatalf("Expected: 'myCustomRole', got: %v", reconciled.RoleRef.Name)
	}
}
