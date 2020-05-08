package reconcilers

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestConfigMapReconcileFieldCopiedObjects(t *testing.T) {
	desired := &v1.ConfigMap{
		Data: map[string]string{
			"a1": "a1Value",
		},
	}

	existingTmp := desired.DeepCopyObject()
	existing, ok := existingTmp.(*v1.ConfigMap)
	if !ok {
		t.Fatal("configMap copy did not work")
	}

	update := ConfigMapReconcileField(desired, existing, "a1")

	if update {
		t.Fatal("when existing and desired are cloned, reconciler reported update needed")
	}
}

func TestConfigMapReconcileFieldMissingField(t *testing.T) {
	desired := &v1.ConfigMap{
		Data: map[string]string{
			"a1": "a1Value",
		},
	}

	existing := &v1.ConfigMap{
		Data: map[string]string{
			"a2": "a2Value",
		},
	}

	update := ConfigMapReconcileField(desired, existing, "a1")

	if !update {
		t.Fatal("when field is missing, reconciler reported no update needed")
	}

	a1Value, ok := existing.Data["a1"]
	if !ok {
		t.Fatal("existing does not have a1 data")
	}

	if a1Value != "a1Value" {
		t.Fatalf("existing data not expected. Expected: 'a1Value', got: %s", a1Value)
	}
}

func TestConfigMapReconcileFieldDifferentValue(t *testing.T) {
	desired := &v1.ConfigMap{
		Data: map[string]string{
			"a1": "desiredA1Value",
		},
	}

	existing := &v1.ConfigMap{
		Data: map[string]string{
			"a1": "existingA1Value",
		},
	}

	update := ConfigMapReconcileField(desired, existing, "a1")

	if !update {
		t.Fatal("when field value is different, reconciler reported no update needed")
	}

	a1Value, ok := existing.Data["a1"]
	if !ok {
		t.Fatal("existing does not have a1 data")
	}

	if a1Value != "desiredA1Value" {
		t.Fatalf("existing data not expected. Expected: 'desiredA1Value', got: %s", a1Value)
	}
}
