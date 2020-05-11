package reconcilers

import (
	"testing"

	"github.com/3scale/3scale-operator/pkg/helper"

	v1 "k8s.io/api/core/v1"
)

func TestDefaultsOnlySecretMutatorNoUpdateNeeded(t *testing.T) {
	desired := &v1.Secret{
		StringData: map[string]string{
			"a1": "a1Value",
			"a2": "a2Value",
		},
	}
	existing := &v1.Secret{
		StringData: map[string]string{
			"a1": "other_a1_value",
			"a2": "other_a2_value",
		},
	}
	existing.Data = helper.GetSecretDataFromStringData(existing.StringData)

	update, err := DefaultsOnlySecretMutator(existing, desired)
	if err != nil {
		t.Fatal(err)
	}

	if update {
		t.Fatal("when defaults cannot be applied, reconciler reported update needed")
	}
}

func TestDefaultsOnlySecretReconciler(t *testing.T) {
	desired := &v1.Secret{
		StringData: map[string]string{
			"a1": "a01Value",
			"a2": "a02Value",
		},
	}
	existing := &v1.Secret{
		StringData: map[string]string{
			"a2": "other_a2_value",
			"a3": "a3Value",
		},
	}
	existing.Data = helper.GetSecretDataFromStringData(existing.StringData)

	update, err := DefaultsOnlySecretMutator(existing, desired)
	if err != nil {
		t.Fatal(err)
	}

	if !update {
		t.Fatal("when defaults can be applied, reconciler reported no update needed")
	}

	_, ok := existing.StringData["a1"]
	if !ok {
		t.Fatal("existingSecret does not have a1 data")
	}

	a2Value, ok := existing.StringData["a2"]
	if !ok {
		t.Fatal("existingSecret does not have a2 data")
	}

	if a2Value != "other_a2_value" {
		t.Fatalf("existingSecret data not expected. Expected: 'other_a2_value', got: %s", a2Value)
	}

	_, ok = existing.StringData["a3"]
	if !ok {
		t.Fatal("existingSecret does not have a3 data")
	}
}
