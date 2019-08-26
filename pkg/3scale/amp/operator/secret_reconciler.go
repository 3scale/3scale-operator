package operator

import (
	"context"
	"fmt"

	"github.com/3scale/3scale-operator/pkg/helper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type SecretReconciler interface {
	IsUpdateNeeded(desired, existing *v1.Secret) bool
}

type SecretBaseReconciler struct {
	BaseAPIManagerLogicReconciler
	reconciler SecretReconciler
}

func NewSecretBaseReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler, reconciler SecretReconciler) *SecretBaseReconciler {
	return &SecretBaseReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
		reconciler:                    reconciler,
	}
}

func (r *SecretBaseReconciler) Reconcile(desired *v1.Secret) error {
	objectInfo := ObjectInfo(desired)
	existing := &v1.Secret{}
	err := r.Client().Get(
		context.TODO(),
		types.NamespacedName{Name: desired.Name, Namespace: r.apiManager.GetNamespace()},
		existing)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.createResource(desired)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			return nil
		}
		return err
	}

	update, err := r.isUpdateNeeded(desired, existing)
	if err != nil {
		return err
	}

	if update {
		return r.updateResource(existing)
	}

	return nil
}

func (r *SecretBaseReconciler) isUpdateNeeded(desired, existing *v1.Secret) (bool, error) {
	updated := helper.EnsureObjectMeta(&existing.ObjectMeta, &desired.ObjectMeta)

	updatedTmp, err := r.ensureOwnerReference(existing)
	if err != nil {
		return false, nil
	}

	updated = updated || updatedTmp

	updatedTmp = r.reconciler.IsUpdateNeeded(desired, existing)
	updated = updated || updatedTmp

	return updated, nil
}

type CreateOnlySecretReconciler struct {
}

func NewCreateOnlySecretReconciler() *CreateOnlySecretReconciler {
	return &CreateOnlySecretReconciler{}
}

func (r *CreateOnlySecretReconciler) IsUpdateNeeded(desired, existing *v1.Secret) bool {
	return false
}

type DefaultsOnlySecretReconciler struct {
}

// Reconciler for defaults only is useful for secrets pre-created by the user and when not all the fields are created.
// Fields referenced from deployment configs must exist,
// so defaults only reconciliation makes sure they exist with default values when user does doe pre-create them
func NewDefaultsOnlySecretReconciler() *DefaultsOnlySecretReconciler {
	return &DefaultsOnlySecretReconciler{}
}

func (r *DefaultsOnlySecretReconciler) IsUpdateNeeded(desired, existing *v1.Secret) bool {
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

	return updated
}

func SecretReconcileField(desired, existing *v1.Secret, fieldName string) bool {
	updated := false

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
