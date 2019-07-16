package operator

import (
	"context"
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type ServiceReconciler struct {
	BaseReconciler
	ObjectMetaMerger ObjectMetaMerger
}

func NewServiceReconciler(baseReconciler BaseReconciler, objectMetaMerger ObjectMetaMerger) ServiceReconciler {
	return ServiceReconciler{
		BaseReconciler:   baseReconciler,
		ObjectMetaMerger: objectMetaMerger,
	}
}

func (r *ServiceReconciler) Reconcile(desiredService *v1.Service) error {
	objectInfo := ObjectInfo(desiredService)

	existingService := &v1.Service{}
	err := r.Client().Get(context.TODO(), types.NamespacedName{Name: desiredService.Name, Namespace: desiredService.Namespace}, existingService)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.Client().Create(context.TODO(), desiredService)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			r.Logger().Info(fmt.Sprintf("Created object %s", objectInfo))
			return nil
		}
		return err
	}

	needsUpdate, err := r.ensureService(existingService, desiredService)
	if err != nil {
		return err
	}

	if needsUpdate {
		r.Logger().Info(fmt.Sprintf("Updating Service %s", objectInfo))
		err := r.Client().Update(context.TODO(), existingService)
		if err != nil {
			r.Logger().Error(err, fmt.Sprintf("error updating Service %s", objectInfo))
			return err
		}
	}

	return nil
}

func (r *ServiceReconciler) ensureService(updated, desired *v1.Service) (bool, error) {
	changed := false

	objectMetaChanged, err := r.ObjectMetaMerger.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta)
	if err != nil {
		return false, err
	}
	if objectMetaChanged {
		changed = true
	}

	desired.Spec.ClusterIP = updated.Spec.ClusterIP
	desired.Spec.Type = updated.Spec.Type
	desired.Spec.SessionAffinity = updated.Spec.SessionAffinity

	if !reflect.DeepEqual(updated.Spec, desired.Spec) {
		updated.Spec = desired.Spec
		changed = true
	}

	return changed, nil
}
