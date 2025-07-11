package reconcilers

import (
	"fmt"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
)

func ServiceMutator(opts ...MutateFn) MutateFn {
	return func(existingObj, desiredObj client.Object) (bool, error) {
		existing, ok := existingObj.(*v1.Service)
		if !ok {
			return false, fmt.Errorf("%T is not a *v1.Service", existingObj)
		}
		desired, ok := desiredObj.(*v1.Service)
		if !ok {
			return false, fmt.Errorf("%T is not a *v1.Service", desiredObj)
		}

		update := false

		// Loop through each option
		for _, opt := range opts {
			tmpUpdate, err := opt(existing, desired)
			if err != nil {
				return false, err
			}
			update = update || tmpUpdate
		}

		return update, nil
	}
}

func ServicePortMutator(existingObj, desiredObj client.Object) (bool, error) {
	existing, ok := existingObj.(*v1.Service)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.Service", existingObj)
	}
	desired, ok := desiredObj.(*v1.Service)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.Service", desiredObj)
	}

	updated := false

	if !reflect.DeepEqual(existing.Spec.Ports, desired.Spec.Ports) {
		updated = true
		existing.Spec.Ports = desired.Spec.Ports
	}

	return updated, nil
}

func ServiceSelectorMutator(existingObj, desiredObj client.Object) (bool, error) {
	existing, ok := existingObj.(*v1.Service)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.Service", existingObj)
	}
	desired, ok := desiredObj.(*v1.Service)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.Service", desiredObj)
	}

	updated := false

	if !reflect.DeepEqual(existing.Spec.Selector, desired.Spec.Selector) {
		updated = true
		existing.Spec.Selector = desired.Spec.Selector
	}

	return updated, nil
}
