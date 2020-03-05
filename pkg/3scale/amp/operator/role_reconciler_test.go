package operator

import (
	"context"
	"reflect"
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

func TestRoleBaseReconcilerCreate(t *testing.T) {
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
	createOnlyReconciler := NewCreateOnlyRoleReconciler()

	reconciler := NewRoleBaseReconciler(baseAPIManagerLogicReconciler, createOnlyReconciler)

	desired := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myRole",
			Namespace: namespace,
		},
	}

	err = reconciler.Reconcile(desired)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "myRole",
		Namespace: namespace,
	}
	existing := &rbacv1.Role{}
	err = cl.Get(context.TODO(), namespacedName, existing)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Fatal(err)
	}
}

func TestRoleBaseReconcilerUpdateOwnerRef(t *testing.T) {
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

	existing := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myRole",
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
	createOnlyReconciler := NewCreateOnlyRoleReconciler()

	reconciler := NewRoleBaseReconciler(baseAPIManagerLogicReconciler, createOnlyReconciler)

	desired := existing.DeepCopy()

	err = reconciler.Reconcile(desired)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "myRole",
		Namespace: namespace,
	}
	reconciled := &rbacv1.Role{}
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

type myCustomRoleReconciler struct {
}

func (r *myCustomRoleReconciler) IsUpdateNeeded(desired, existing *rbacv1.Role) bool {
	existing.Rules = []rbacv1.PolicyRule{
		rbacv1.PolicyRule{
			APIGroups: []string{"apps.openshift.io"},
			Resources: []string{
				"deploymentconfigs",
			},
			Verbs: []string{
				"get",
				"list",
			},
		},
	}
	return true
}

func newCustomRoleReconciler() RoleReconciler {
	return &myCustomRoleReconciler{}
}

func TestRoleBaseReconcilerUpdateNeeded(t *testing.T) {
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
	existing := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myRole",
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
	reconciler := NewRoleBaseReconciler(baseAPIManagerLogicReconciler, newCustomRoleReconciler())
	desired := existing.DeepCopy()
	err = reconciler.Reconcile(desired)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "myRole",
		Namespace: namespace,
	}
	reconciled := &rbacv1.Role{}
	err = cl.Get(context.TODO(), namespacedName, reconciled)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Fatal(err)
	}

	if reconciled.Rules == nil || len(reconciled.Rules) == 0 {
		t.Fatal("reconciled does not have reconciled data")
	}

	if !reflect.DeepEqual(reconciled.Rules[0].Verbs, []string{"get", "list"}) {
		t.Fatalf("reconciled have reconciled data. Expected: '%v', got: %v", []string{"get", "list"}, reconciled.Rules[0].Verbs)
	}
}
