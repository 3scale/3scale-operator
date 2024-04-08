package reconcilers

import (
	v1 "k8s.io/api/core/v1"
	"strings"
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

func RedisConfigMapReconcileField(desired, existing *v1.ConfigMap, fieldName string) bool {
	updated := false

	if existingVal, ok := existing.Data[fieldName]; !ok {
		existing.Data[fieldName] = desired.Data[fieldName]
		updated = true
	} else {
		existingString := existingVal

		replicaOfString := "rename-command REPLICAOF \"\""
		if !strings.Contains(existingString, replicaOfString) {
			existingString = existingString + "\n" + replicaOfString
			existing.Data[fieldName] = existingString
			updated = true
		}

		slaveOfString := "rename-command SLAVEOF \"\""
		if !strings.Contains(existingString, slaveOfString) {
			existingString = existingString + "\n" + slaveOfString
			existing.Data[fieldName] = existingString
			updated = true
		}
	}
	return updated
}
