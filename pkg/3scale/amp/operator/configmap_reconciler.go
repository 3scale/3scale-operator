package operator

import (
	"context"
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type ConfigMapReconciler struct {
	BaseReconciler
	ObjectMetaMerger ObjectMetaMerger
}

func NewConfigMapReconciler(baseReconciler BaseReconciler, objectMetaMerger ObjectMetaMerger) ConfigMapReconciler {
	return ConfigMapReconciler{
		BaseReconciler:   baseReconciler,
		ObjectMetaMerger: objectMetaMerger,
	}
}

func (r *ConfigMapReconciler) Reconcile(desiredConfigMap *v1.ConfigMap) error {
	objectInfo := ObjectInfo(desiredConfigMap)
	existingConfigMap := &v1.ConfigMap{}
	err := r.Client().Get(context.TODO(), types.NamespacedName{Name: desiredConfigMap.Name, Namespace: desiredConfigMap.Namespace}, existingConfigMap)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.Client().Create(context.TODO(), desiredConfigMap)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			r.Logger().Info(fmt.Sprintf("Created object %s", objectInfo))
			return nil
		}
		return err
	}

	needsUpdate, err := r.ensureConfigMap(existingConfigMap, desiredConfigMap)
	if err != nil {
		return err
	}

	if needsUpdate {
		r.Logger().Info(fmt.Sprintf("Updating ConfigMap %s", objectInfo))
		err := r.Client().Update(context.TODO(), existingConfigMap)
		if err != nil {
			r.Logger().Error(err, fmt.Sprintf("error updating Service %s", objectInfo))
			return err
		}
	}

	return nil
}

func (r *ConfigMapReconciler) ensureConfigMap(updated, desired *v1.ConfigMap) (bool, error) {
	changed := false

	objectMetaChanged, err := r.ObjectMetaMerger.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta)
	if err != nil {
		return false, err
	}
	if objectMetaChanged {
		changed = true
	}

	// TODO should be the reconciliation of ConfigMap data a merge behavior
	// instead of a replace one?
	// TODO should we reconcile BinaryData too???
	if !reflect.DeepEqual(updated.Data, desired.Data) {
		updated.Data = desired.Data
		changed = true
	}

	return changed, nil
}
