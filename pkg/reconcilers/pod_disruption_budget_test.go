package reconcilers

import (
	"testing"

	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func pdbTestFactory(maxUnavailable int32) *policyv1beta1.PodDisruptionBudget {
	return &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myPodDisruptionBudget",
			Namespace: "someNs",
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"test1": "mytest1"},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: maxUnavailable},
		},
	}
}

func TestGenericPDBMutator(t *testing.T) {
	var existingMaxUnavailable int32 = 1
	var desiredMaxUnavailable int32 = 2

	existing := pdbTestFactory(existingMaxUnavailable)
	desired := pdbTestFactory(desiredMaxUnavailable)

	update, err := GenericPDBMutator(existing, desired)
	if err != nil {
		t.Fatal(err)
	}
	if !update {
		t.Fatal("when defaults can be applied, reconciler reported no update needed")
	}

	if existing.Spec.MaxUnavailable.IntVal != desiredMaxUnavailable {
		t.Fatalf("Maxunavailable not reconciled. Expected: %d, got: %d", desiredMaxUnavailable, existing.Spec.MaxUnavailable.IntVal)
	}
}
