package reconcilers

import (
	v1 "k8s.io/api/core/v1"
)

func ConfigMapReconcileField(desired, existing *v1.ConfigMap, fieldName string) bool {
	updated := false

	if existingVal, ok := existing.Data[fieldName]; !ok {
		existing.Data[fieldName] = desired.Data[fieldName]
		updated = true
	} else {
		if desired.Data[fieldName] != existingVal {
			existing.Data[fieldName] = desired.Data[fieldName]
			updated = true
		}
	}
	return updated
}
