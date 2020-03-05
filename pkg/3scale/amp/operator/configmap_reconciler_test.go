package operator

import (
	"context"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestCreateOnlyConfigMapReconciler(t *testing.T) {
	createOnlyConfigMapReconciler := NewCreateOnlyConfigMapReconciler()
	desired := &v1.ConfigMap{}
	existing := &v1.ConfigMap{}
	if createOnlyConfigMapReconciler.IsUpdateNeeded(desired, existing) {
		t.Fatal("Create only reconciler reported update needed")
	}
}

func TestConfigMapBaseReconcilerCreate(t *testing.T) {
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
	err = appsv1.AddToScheme(s)
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
	createOnlyConfigMapReconciler := NewCreateOnlyConfigMapReconciler()

	configmapReconciler := NewConfigMapBaseReconciler(baseAPIManagerLogicReconciler, createOnlyConfigMapReconciler)

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

	err = configmapReconciler.Reconcile(desiredConfigmap)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "myConfigmap",
		Namespace: namespace,
	}
	reconciledConfigmap := &v1.ConfigMap{}
	err = cl.Get(context.TODO(), namespacedName, reconciledConfigmap)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Errorf("error fetching existing: %v", err)
	}
}

func TestConfigMapBaseReconcilerUpdateOwnerRef(t *testing.T) {
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
	existingConfigMap := &v1.ConfigMap{
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
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.SchemeGroupVersion, apimanager)
	err = appsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{existingConfigMap}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)

	baseReconciler := NewBaseReconciler(cl, clientAPIReader, s, log, cfg)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseLogicReconciler, apimanager)
	createOnlyConfigmapReconciler := NewCreateOnlyConfigMapReconciler()

	configmapReconciler := NewConfigMapBaseReconciler(baseAPIManagerLogicReconciler, createOnlyConfigmapReconciler)

	desiredConfigmap := existingConfigMap.DeepCopy()

	err = configmapReconciler.Reconcile(desiredConfigmap)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      desiredConfigmap.Name,
		Namespace: namespace,
	}

	reconciledConfigmap := &v1.ConfigMap{}
	err = cl.Get(context.TODO(), namespacedName, reconciledConfigmap)
	if err != nil {
		t.Errorf("error fetching reconciled: %v", err)
	}

	if len(reconciledConfigmap.GetOwnerReferences()) != 1 {
		t.Errorf("reconciled obj does not have owner reference")
	}

	if reconciledConfigmap.GetOwnerReferences()[0].Name != name {
		t.Errorf("reconciled owner reference is not apimanager, expected: %s, got: %s", name, reconciledConfigmap.GetOwnerReferences()[0].Name)
	}
}

type myCustomConfigmapReconciler struct {
}

func (r *myCustomConfigmapReconciler) IsUpdateNeeded(desired, existing *v1.ConfigMap) bool {
	if existing.Data == nil {
		existing.Data = map[string]string{}
	}
	existing.Data["customKey"] = "customValue"
	return true
}

func newCustomConfigmapReconciler() ConfigMapReconciler {
	return &myCustomConfigmapReconciler{}
}

func TestConfigMapBaseReconcilerUpdateNeeded(t *testing.T) {
	// Test that update is done when reconciler tells
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
	err = appsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	existing := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myConfigmap",
			Namespace: namespace,
		},
		Data: map[string]string{},
	}

	// existing obj does not need to be updated to set owner reference
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
	customConfigmapReconciler := newCustomConfigmapReconciler()

	configmapReconciler := NewConfigMapBaseReconciler(baseAPIManagerLogicReconciler, customConfigmapReconciler)

	desired := existing.DeepCopy()

	err = configmapReconciler.Reconcile(desired)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      desired.Name,
		Namespace: namespace,
	}

	reconciled := &v1.ConfigMap{}
	err = cl.Get(context.TODO(), namespacedName, reconciled)
	if err != nil {
		t.Fatalf("error fetching reconciled: %v", err)
	}

	customValue, ok := reconciled.Data["customKey"]
	if !ok {
		t.Fatal("reconciled does not have reconciled data")
	}

	if customValue != "customValue" {
		t.Fatalf("reconciled have reconciled data. Expected: 'customValue', got: %s", customValue)
	}
}
