package reconcilers

import (
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DefaultsOnlySecretMutator is useful for secrets pre-created by the user and when not all the fields are created.
// Fields referenced from deployments must exist,
// so defaults only reconciliation makes sure they exist with default values when user does doe pre-create them
func DefaultsOnlySecretMutator(existingObj, desiredObj client.Object) (bool, error) {
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

// SecretMutateFn is a function which mutates the existing Secret into it's desired state.
type SecretMutateFn func(desired, existing *v1.Secret) bool

func DeploymentSecretMutator(opts ...SecretMutateFn) MutateFn {
	return func(existingObj, desiredObj client.Object) (bool, error) {
		existing, ok := existingObj.(*v1.Secret)
		if !ok {
			return false, fmt.Errorf("%T is not a *v1.Secret", existingObj)
		}
		desired, ok := desiredObj.(*v1.Secret)
		if !ok {
			return false, fmt.Errorf("%T is not a *v1.Secret", desiredObj)
		}

		update := false

		// Loop through each option
		for _, opt := range opts {
			tmpUpdate := opt(desired, existing)
			update = update || tmpUpdate
		}

		return update, nil
	}
}

func SecretReconcileField(fieldName string) func(desired, existing *v1.Secret) bool {
	return func(desired, existing *v1.Secret) bool {
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
}

func SecretStringDataMutator(desired, existing *v1.Secret) bool {
	updated := false

	// StringData is merged to Data on write, so we need to compare the existing Data to the desired StringData
	// Before we can do this we need to convert the existing Data to StringData
	existingStringData := make(map[string]string)
	for key, bytes := range existing.Data {
		existingStringData[key] = string(bytes)
	}
	if !reflect.DeepEqual(existingStringData, desired.StringData) {
		updated = true
		existing.Data = nil // Need to clear the existing.Data because of how StringData is converted to Data
		existing.StringData = desired.StringData
	}

	return updated
}
