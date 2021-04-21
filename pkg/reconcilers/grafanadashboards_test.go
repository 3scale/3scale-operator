package reconcilers

import (
	"testing"

	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
)

func TestGenericGrafanaDashboardsMutatorWhenCopied(t *testing.T) {
	desired := &grafanav1alpha1.GrafanaDashboard{
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Name: "desiredName",
			Json: `{"somekey": "somevalue"}`,
		},
	}

	existingTmp := desired.DeepCopyObject()
	existing, ok := existingTmp.(*grafanav1alpha1.GrafanaDashboard)
	if !ok {
		t.Fatal("grafanadashboard copy did not work")
	}

	update, err := GenericGrafanaDashboardsMutator(desired, existing)
	if err != nil {
		t.Fatal(err)
	}

	if update {
		t.Fatal("when existing and desired are cloned, reconciler reported update needed")
	}
}

func TestGenericGrafanaDashboardsMutatorWhenDiff(t *testing.T) {
	desired := &grafanav1alpha1.GrafanaDashboard{
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Name: "desiredName",
			Json: `{"somekey": "somevalue"}`,
		},
	}

	existingTmp := desired.DeepCopyObject()
	existing, ok := existingTmp.(*grafanav1alpha1.GrafanaDashboard)
	if !ok {
		t.Fatal("grafanadashboard copy did not work")
	}

	existing.Spec.Name = "ExistingName"
	existing.Spec.Json = `{"some_existing_key": "some_existing_value"}`

	update, err := GenericGrafanaDashboardsMutator(desired, existing)
	if err != nil {
		t.Fatal(err)
	}

	if !update {
		t.Fatal("when existing and desired are different, reconciler reported not update needed")
	}

	if existing.Spec.Name != desired.Spec.Name {
		t.Errorf("Spec.Name does not match. got [%s], expected [%s]", existing.Spec.Name, desired.Spec.Name)
	}

	if existing.Spec.Json != desired.Spec.Json {
		t.Errorf("Spec.Json does not match. got [%s], expected [%s]", existing.Spec.Json, desired.Spec.Json)
	}
}
