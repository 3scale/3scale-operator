package reconcilers

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/common"

	v1 "k8s.io/api/core/v1"
)

// DefaultsOnlySecretMutator is useful for secrets pre-created by the user and when not all the fields are created.
// Fields referenced from deployment configs must exist,
// so defaults only reconciliation makes sure they exist with default values when user does doe pre-create them
func DefaultsOnlySecretMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*v1.Secret)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.Secret", existingObj)
	}
	desired, ok := desiredObj.(*v1.Secret)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.Secret", desiredObj)
	}

	updated := false

	if existing.StringData == nil {
		existing.StringData = map[string]string{}
	}

	for k, v := range desired.StringData {
		if _, ok := existing.Data[k]; !ok {
			existing.StringData[k] = v
			updated = true
		}
	}

	return updated, nil
}

func SecretReconcileField(desired, existing *v1.Secret, fieldName string) bool {
	updated := false

	if existing.Data == nil {
		existing.Data = map[string][]byte{}
	}
	if existing.StringData == nil {
		existing.StringData = map[string]string{}
	}

	valB, ok := existing.Data[fieldName]
	if !ok {
		existing.StringData[fieldName] = desired.StringData[fieldName]
		updated = true
	} else {
		valStr := string(valB)
		if desired.StringData[fieldName] != valStr {
			// should merge existing key in Data struct
			existing.StringData[fieldName] = desired.StringData[fieldName]
			updated = true
		}
	}
	return updated
}
