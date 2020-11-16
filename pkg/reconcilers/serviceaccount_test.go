package reconcilers

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServiceAccountImagePullPolicyMutator(t *testing.T) {
	existingImagePullSecrets := []v1.LocalObjectReference{
		v1.LocalObjectReference{Name: "imagepullsecret3"},
		v1.LocalObjectReference{Name: "imagepullsecret4"},
		v1.LocalObjectReference{Name: "myserviceaccount-dockercfg-XXXXX"},
		v1.LocalObjectReference{Name: "imagepullsecret7"},
		v1.LocalObjectReference{Name: "myserviceaccount-token-XXXXX"},
	}
	existing := serviceAccountTestFactory(existingImagePullSecrets)

	desiredImagePullSecrets := []v1.LocalObjectReference{
		v1.LocalObjectReference{Name: "imagepullsecret4"},
		v1.LocalObjectReference{Name: "newimagepullsecret"},
		v1.LocalObjectReference{Name: "anotherpullsecret"},
		v1.LocalObjectReference{Name: "imagepullsecret3"},
	}
	desired := serviceAccountTestFactory(desiredImagePullSecrets)

	changed, err := ServiceAccountImagePullPolicyMutator(existing, desired)
	if err != nil {
		t.Fatal(err)
	}

	if !changed {
		t.Fatalf("No changes detected. Expected: %t, got %t", true, changed)
	}

	newExistingImagePullSecrets := []v1.LocalObjectReference{
		v1.LocalObjectReference{Name: "anotherpullsecret"},
		v1.LocalObjectReference{Name: "imagepullsecret3"},
		v1.LocalObjectReference{Name: "imagepullsecret4"},
		v1.LocalObjectReference{Name: "myserviceaccount-dockercfg-XXXXX"},
		v1.LocalObjectReference{Name: "myserviceaccount-token-XXXXX"},
		v1.LocalObjectReference{Name: "newimagepullsecret"},
	}
	newExisting := serviceAccountTestFactory(newExistingImagePullSecrets)

	if !reflect.DeepEqual(existing, newExisting) {
		t.Fatalf("Unexpected reconciliated ImagePullSecrets. Expected: %+v, Got: %+v", newExisting, existing)
	}
}

func serviceAccountTestFactory(imagePullSecrets []v1.LocalObjectReference) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myserviceaccount",
			Namespace: "someNs",
		},
		ImagePullSecrets: imagePullSecrets,
	}
}

func TestServiceAccountImagePullPolicyMutatorImagePullPoliciesAreOrderIndependent(t *testing.T) {
	existingImagePullSecrets := []v1.LocalObjectReference{
		v1.LocalObjectReference{Name: "imagepullsecret4"},
		v1.LocalObjectReference{Name: "imagepullsecret3"},
	}
	existing := serviceAccountTestFactory(existingImagePullSecrets)
	originalExisting := existing.DeepCopy()

	desiredImagePullSecrets := []v1.LocalObjectReference{
		v1.LocalObjectReference{Name: "imagepullsecret3"},
		v1.LocalObjectReference{Name: "imagepullsecret4"},
	}
	desired := serviceAccountTestFactory(desiredImagePullSecrets)

	changed, err := ServiceAccountImagePullPolicyMutator(existing, desired)
	if err != nil {
		t.Fatal(err)
	}

	if changed {
		t.Fatalf("Changes detected. Expected: %t, got %t", false, changed)
	}

	if !reflect.DeepEqual(existing, originalExisting) {
		t.Fatalf("Changed detected. No changes were expected. Original: %v, got %v", originalExisting, existing)
	}
}

func TestServiceAccountImagePullPolicyMutatorK8sAddedSecretsArePreserved(t *testing.T) {
	existingImagePullSecrets := []v1.LocalObjectReference{
		v1.LocalObjectReference{Name: "imagepullsecret4"},
		v1.LocalObjectReference{Name: "imagepullsecret3"},
		v1.LocalObjectReference{Name: "myserviceaccount-token-XXXXX"},
		v1.LocalObjectReference{Name: "myserviceaccount-dockercfg-XXXXX"},
	}
	existing := serviceAccountTestFactory(existingImagePullSecrets)

	desiredImagePullSecrets := []v1.LocalObjectReference{
		v1.LocalObjectReference{Name: "imagepullsecret9"},
		v1.LocalObjectReference{Name: "imagepullsecret8"},
		v1.LocalObjectReference{Name: "imagepullsecret4"},
	}
	desired := serviceAccountTestFactory(desiredImagePullSecrets)

	changed, err := ServiceAccountImagePullPolicyMutator(existing, desired)
	if err != nil {
		t.Fatal(err)
	}

	if !changed {
		t.Fatalf("No changes detected. Expected: %t, got %t", true, changed)
	}

	newExistingImagePullSecrets := []v1.LocalObjectReference{
		v1.LocalObjectReference{Name: "imagepullsecret4"},
		v1.LocalObjectReference{Name: "imagepullsecret8"},
		v1.LocalObjectReference{Name: "imagepullsecret9"},
		v1.LocalObjectReference{Name: "myserviceaccount-dockercfg-XXXXX"},
		v1.LocalObjectReference{Name: "myserviceaccount-token-XXXXX"},
	}
	newExisting := serviceAccountTestFactory(newExistingImagePullSecrets)

	if !reflect.DeepEqual(existing, newExisting) {
		t.Fatalf("Unexpected reconciliated ImagePullSecrets. Expected: %+v, Got: %+v", newExisting, existing)
	}
}
