package reconcilers

import (
	"testing"

	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
)

func TestGenericGrafanaDashboardsMutatorWhenCopied(t *testing.T) {
	desired := &grafanav1alpha1.GrafanaDashboard{
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
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
			Json: `{"somekey": "somevalue"}`,
		},
	}

	existingTmp := desired.DeepCopyObject()
	existing, ok := existingTmp.(*grafanav1alpha1.GrafanaDashboard)
	if !ok {
		t.Fatal("grafanadashboard copy did not work")
	}

	existing.Spec.Json = `{"some_existing_key": "some_existing_value"}`

	update, err := GenericGrafanaDashboardsMutator(desired, existing)
	if err != nil {
		t.Fatal(err)
	}

	if !update {
		t.Fatal("when existing and desired are different, reconciler reported not update needed")
	}

	if existing.Spec.Json != desired.Spec.Json {
		t.Errorf("Spec.Json does not match. got [%s], expected [%s]", existing.Spec.Json, desired.Spec.Json)
	}
}
