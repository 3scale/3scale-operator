package operator

import (
	"context"
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type SecretReconciler struct {
	BaseReconciler
	ObjectMetaMerger ObjectMetaMerger
}

func NewSecretReconciler(baseReconciler BaseReconciler, objectMetaMerger ObjectMetaMerger) SecretReconciler {
	return SecretReconciler{
		BaseReconciler:   baseReconciler,
		ObjectMetaMerger: objectMetaMerger,
	}
}

// TODO should this be a shared Secret reconcile behaviour or could there be
// different secrets where we reconcile them in a different way?
// For example, secrets where we just want to check if they are installed it
// but we don't want to reconcile them. That would make sense for example
// for system-seed secret.
// Also in the case where we wanted to reconcile them but we wanted to do it
// in a different way
func (r *SecretReconciler) Reconcile(desiredSecret *v1.Secret) error {
	objectInfo := ObjectInfo(desiredSecret)
	existingSecret := &v1.Secret{}
	err := r.Client().Get(context.TODO(), types.NamespacedName{Name: desiredSecret.Name, Namespace: desiredSecret.Namespace}, existingSecret)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.Client().Create(context.TODO(), desiredSecret)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			r.Logger().Info(fmt.Sprintf("Created object %s", objectInfo))
			return nil
		}
		return err
	}

	needsUpdate, err := r.ensureSecret(existingSecret, desiredSecret)

	if needsUpdate {
		r.Logger().Info(fmt.Sprintf("Updating Secret %s", objectInfo))
		err := r.Client().Update(context.TODO(), existingSecret)
		if err != nil {
			r.Logger().Error(err, fmt.Sprintf("error updating Service %s", objectInfo))
			return err
		}
	}

	return nil
}

func (r *SecretReconciler) ensureSecret(updated, desired *v1.Secret) (bool, error) {
	changed := false

	objectMetaChanged, err := r.ObjectMetaMerger.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta)
	if err != nil {
		return false, err
	}
	if objectMetaChanged {
		changed = true
	}

	// TODO writing on StringData has merge behaviour with existing Data content.
	// Do we want this or do we want to always overwrite everything? Take note that
	// in ConfigMap for example it seems there's no merge behavior by itself so
	// either we implement it by ourselves or we would have different behavior
	// between secret and configmap fields reconciliation
	// TODO there's also a case where a StringData field that is set to empty
	// is encoded as the empty string in the Data section but when there's another
	// update it is encoded to null, and that would be detected as a difference.
	// Would that be dangerous/a problem?
	updatedSecretStringData := getSecretStringDataFromData(updated.Data)
	if !reflect.DeepEqual(updatedSecretStringData, desired.StringData) {
		updated.StringData = desired.StringData
		changed = true
	}

	return changed, nil
}
