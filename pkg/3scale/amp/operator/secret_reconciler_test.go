package operator

import (
	"context"
	"testing"

	"github.com/3scale/3scale-operator/pkg/helper"

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

func TestCreateOnlySecretReconciler(t *testing.T) {
	createOnlySecretReconciler := NewCreateOnlySecretReconciler()
	desiredSecret := &v1.Secret{}
	existingSecret := &v1.Secret{}
	if createOnlySecretReconciler.IsUpdateNeeded(desiredSecret, existingSecret) {
		t.Fatal("Create only secret reconciled reported update needed")
	}
}

func TestDefaultsOnlySecretReconciler(t *testing.T) {
	defaultsOnlySecretReconciler := NewDefaultsOnlySecretReconciler()
	desiredSecret := &v1.Secret{
		StringData: map[string]string{
			"a1": "a01Value",
			"a2": "a02Value",
		},
	}
	existingSecret := &v1.Secret{
		StringData: map[string]string{
			"a2": "other_a2_value",
			"a3": "a3Value",
		},
	}
	existingSecret.Data = helper.GetSecretDataFromStringData(existingSecret.StringData)
	if !defaultsOnlySecretReconciler.IsUpdateNeeded(desiredSecret, existingSecret) {
		t.Fatal("when defaults can be applied, reconciler reported no update needed")
	}

	_, ok := existingSecret.StringData["a1"]
	if !ok {
		t.Fatal("existingSecret does not have a1 data")
	}

	a2Value, ok := existingSecret.StringData["a2"]
	if !ok {
		t.Fatal("existingSecret does not have a2 data")
	}

	if a2Value != "other_a2_value" {
		t.Fatalf("existingSecret data not expected. Expected: 'other_a2_value', got: %s", a2Value)
	}

	_, ok = existingSecret.StringData["a3"]
	if !ok {
		t.Fatal("existingSecret does not have a3 data")
	}
}

func TestDefaultsOnlySecretReconcilerNoUpdateNeeded(t *testing.T) {
	defaultsOnlySecretReconciler := NewDefaultsOnlySecretReconciler()
	desiredSecret := &v1.Secret{
		StringData: map[string]string{
			"a1": "a1Value",
			"a2": "a2Value",
		},
	}
	existingSecret := &v1.Secret{
		StringData: map[string]string{
			"a1": "other_a1_value",
			"a2": "other_a2_value",
		},
	}
	existingSecret.Data = helper.GetSecretDataFromStringData(existingSecret.StringData)
	if defaultsOnlySecretReconciler.IsUpdateNeeded(desiredSecret, existingSecret) {
		t.Fatal("when defaults cannot be applied, reconciler reported update needed")
	}
}

func TestSecretBaseReconcilerCreate(t *testing.T) {
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
	createOnlySecretReconciler := NewCreateOnlySecretReconciler()

	secretReconciler := NewSecretBaseReconciler(baseAPIManagerLogicReconciler, createOnlySecretReconciler)

	desiredSecret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mySecret",
			Namespace: namespace,
		},
		StringData: map[string]string{
			"somekey": "somevalue",
		},
	}

	err = secretReconciler.Reconcile(desiredSecret)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      "mySecret",
		Namespace: namespace,
	}
	existingSecret := &v1.Secret{}
	err = cl.Get(context.TODO(), namespacedName, existingSecret)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Errorf("error fetching existing secret: %v", err)
	}
}

func TestSecretBaseReconcilerUpdateOwnerRef(t *testing.T) {
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
	existingSecret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mySecret",
			Namespace: namespace,
		},
		StringData: map[string]string{
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
	objs := []runtime.Object{existingSecret}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)

	baseReconciler := NewBaseReconciler(cl, clientAPIReader, s, log, cfg)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseLogicReconciler, apimanager)
	createOnlySecretReconciler := NewCreateOnlySecretReconciler()

	secretReconciler := NewSecretBaseReconciler(baseAPIManagerLogicReconciler, createOnlySecretReconciler)

	desiredSecret := existingSecret.DeepCopy()

	err = secretReconciler.Reconcile(desiredSecret)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      desiredSecret.Name,
		Namespace: namespace,
	}

	reconciledSecret := &v1.Secret{}
	err = cl.Get(context.TODO(), namespacedName, reconciledSecret)
	if err != nil {
		t.Fatal(err)
	}

	if len(reconciledSecret.GetOwnerReferences()) != 1 {
		t.Fatal("reconciled secret does not have owner reference")
	}

	if reconciledSecret.GetOwnerReferences()[0].Name != name {
		t.Fatalf("reconciled secret owner reference is not apimanager, expected: %s, got: %s", name, reconciledSecret.GetOwnerReferences()[0].Name)
	}
}

type myCustomSecretReconciler struct {
}

func (r *myCustomSecretReconciler) IsUpdateNeeded(desired, existing *v1.Secret) bool {
	if existing.StringData == nil {
		existing.StringData = map[string]string{}
	}
	existing.StringData["customKey"] = "customValue"
	return true
}

func newCustomSecretReconciler() SecretReconciler {
	return &myCustomSecretReconciler{}
}

func TestSecretBaseReconcilerUpdateNeeded(t *testing.T) {
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

	existingSecret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mySecret",
			Namespace: namespace,
		},
		StringData: map[string]string{},
	}

	// existing secret does not need to be updated to set owner reference
	err = controllerutil.SetControllerReference(apimanager, existingSecret, s)
	if err != nil {
		t.Fatal(err)
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{existingSecret}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)

	baseReconciler := NewBaseReconciler(cl, clientAPIReader, s, log, cfg)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseLogicReconciler, apimanager)
	customSecretReconciler := newCustomSecretReconciler()

	secretReconciler := NewSecretBaseReconciler(baseAPIManagerLogicReconciler, customSecretReconciler)

	desiredSecret := existingSecret.DeepCopy()

	err = secretReconciler.Reconcile(desiredSecret)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      desiredSecret.Name,
		Namespace: namespace,
	}

	reconciledSecret := &v1.Secret{}
	err = cl.Get(context.TODO(), namespacedName, reconciledSecret)
	if err != nil {
		t.Fatalf("error fetching reconciled secret: %v", err)
	}

	customValue, ok := reconciledSecret.StringData["customKey"]
	if !ok {
		t.Fatal("reconciled secret does not have reconciled data")
	}

	if customValue != "customValue" {
		t.Fatalf("reconciled secret have reconciled data. Expected: 'customValue', got: %s", customValue)
	}
}
