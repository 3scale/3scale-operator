package operator

import (
	"context"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	imagev1 "github.com/openshift/api/image/v1"
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

func TestImageStreamBaseReconcilerCreate(t *testing.T) {
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
	err = imagev1.AddToScheme(s)
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
	genericReconciler := NewImageStreamGenericReconciler()

	isReconciler := NewImageStreamBaseReconciler(baseAPIManagerLogicReconciler, genericReconciler)

	desired := &imagev1.ImageStream{
		TypeMeta:   metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		ObjectMeta: metav1.ObjectMeta{Name: "myIS", Namespace: namespace},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{},
		},
	}

	err = isReconciler.Reconcile(desired)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{Name: "myIS", Namespace: namespace}
	reconciled := &imagev1.ImageStream{}
	err = cl.Get(context.TODO(), namespacedName, reconciled)
	// object must exist, that is all required to be tested
	if err != nil {
		t.Fatal(err)
	}
}

func TestImageStreamBaseReconcilerUpdateOwnerRef(t *testing.T) {
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
	existing := &imagev1.ImageStream{
		TypeMeta:   metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		ObjectMeta: metav1.ObjectMeta{Name: "myIS", Namespace: namespace},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{},
		},
	}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.SchemeGroupVersion, apimanager)
	err = imagev1.AddToScheme(s)
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
	genericReconciler := NewImageStreamGenericReconciler()

	isReconciler := NewImageStreamBaseReconciler(baseAPIManagerLogicReconciler, genericReconciler)

	desired := existing.DeepCopy()

	err = isReconciler.Reconcile(desired)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{Name: desired.Name, Namespace: namespace}

	reconciled := &imagev1.ImageStream{}
	err = cl.Get(context.TODO(), namespacedName, reconciled)
	if err != nil {
		t.Fatal(err)
	}

	if len(reconciled.GetOwnerReferences()) != 1 {
		t.Errorf("reconciled obj does not have owner reference")
	}

	if reconciled.GetOwnerReferences()[0].Name != name {
		t.Errorf("reconciled owner reference is not apimanager, expected: %s, got: %s", name, reconciled.GetOwnerReferences()[0].Name)
	}
}

type myCustomISReconciler struct {
}

func (r *myCustomISReconciler) IsUpdateNeeded(desired, existing *imagev1.ImageStream) bool {
	newTag := imagev1.TagReference{
		Name: "latest",
		From: &v1.ObjectReference{
			Kind: "ImageStreamTag",
			Name: "3scale-1.4",
		},
	}
	existing.Spec.Tags = append(existing.Spec.Tags, newTag)
	return true
}

func newCustomISReconciler() ImageStreamReconciler {
	return &myCustomISReconciler{}
}

func TestImageStreamBaseReconcilerUpdateNeeded(t *testing.T) {
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
	err = imagev1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	existing := &imagev1.ImageStream{
		TypeMeta:   metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		ObjectMeta: metav1.ObjectMeta{Name: "myIS", Namespace: namespace},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{},
		},
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
	customReconciler := newCustomISReconciler()

	isReconciler := NewImageStreamBaseReconciler(baseAPIManagerLogicReconciler, customReconciler)

	desired := existing.DeepCopy()

	err = isReconciler.Reconcile(desired)
	if err != nil {
		t.Fatal(err)
	}

	namespacedName := types.NamespacedName{
		Name:      desired.Name,
		Namespace: namespace,
	}

	reconciled := &imagev1.ImageStream{}
	err = cl.Get(context.TODO(), namespacedName, reconciled)
	if err != nil {
		t.Fatal(err)
	}

	if len(reconciled.Spec.Tags) != 1 {
		t.Fatal("reconciled obj does not have reconciled data")
	}

	if reconciled.Spec.Tags[0].Name != "latest" {
		t.Fatal("reconciled obj does not have reconciled data")
	}
}

func TestImageStreamGenericReconciler(t *testing.T) {
	existing := &imagev1.ImageStream{
		TypeMeta:   metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		ObjectMeta: metav1.ObjectMeta{Name: "myIS", Namespace: "MyNS"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name:         "tag0",
					ImportPolicy: imagev1.TagImportPolicy{Insecure: false, Scheduled: true},
					From:         &v1.ObjectReference{Kind: "ImageStreamTag", Name: "3scale-0"},
				},
				imagev1.TagReference{
					Name:         "tag1",
					ImportPolicy: imagev1.TagImportPolicy{Insecure: false, Scheduled: true},
					From:         &v1.ObjectReference{Kind: "ImageStreamTag", Name: "3scale-1"},
				},
			},
		},
	}

	desired := &imagev1.ImageStream{
		TypeMeta:   metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		ObjectMeta: metav1.ObjectMeta{Name: "myIS", Namespace: "MyNS"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					// tag that should be updated
					Name:         "tag1",
					ImportPolicy: imagev1.TagImportPolicy{Insecure: true, Scheduled: false},
					From:         &v1.ObjectReference{Kind: "ImageStreamTag", Name: "3scale-1other"},
				},
				imagev1.TagReference{
					// tag that should be added
					Name:         "tag2",
					ImportPolicy: imagev1.TagImportPolicy{Insecure: false, Scheduled: true},
					From:         &v1.ObjectReference{Kind: "ImageStreamTag", Name: "3scale-2"},
				},
			},
		},
	}

	genericReconciler := NewImageStreamGenericReconciler()
	if !genericReconciler.IsUpdateNeeded(desired, existing) {
		t.Fatal("when defaults can be applied, reconciler reported no update needed")
	}

	if len(existing.Spec.Tags) != 3 {
		t.Fatalf("reconciled obj does not have expected number of tags. Expected: 3, got: %d", len(existing.Spec.Tags))
	}

	findTagReference := func(tagRefName string, tagRefS []imagev1.TagReference) int {
		for i := range tagRefS {
			if tagRefS[i].Name == tagRefName {
				return i
			}
		}
		return -1
	}

	// tag0 existed previously in obj, should be left untouched
	tag0Index := findTagReference("tag0", existing.Spec.Tags)
	if tag0Index < 0 {
		t.Fatal("reconciled obj does not have tag0")
	}

	// tag1 existed previously in obj, should be updated
	tag1Index := findTagReference("tag1", existing.Spec.Tags)
	if tag1Index < 0 {
		t.Fatal("reconciled obj does not have tag1")
	}
	tag1 := existing.Spec.Tags[1]
	// From and ImportPolicy fields should have been reconciled
	if tag1.From.Name != "3scale-1other" {
		t.Fatal("reconciled obj tag1 'from' was not reconciled")
	}

	if !tag1.ImportPolicy.Insecure {
		t.Fatal("reconciled obj tag1 'impoortpolicy.insecure' was not reconciled")
	}

	if tag1.ImportPolicy.Scheduled {
		t.Fatal("reconciled obj tag1 'impoortpolicy.scheduled' was not reconciled")
	}

	// tag2 did not exist previously in obj, should be appended
	tag2Index := findTagReference("tag2", existing.Spec.Tags)
	if tag2Index < 0 {
		t.Fatal("reconciled obj does not have tag2")
	}
}
