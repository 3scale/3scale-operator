package operator

import (
	"context"
	"fmt"
	"k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PodDisruptionBudgetReconciler struct {
	BaseAPIManagerLogicReconciler
}

func NewPodDisruptionBudgetReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) *PodDisruptionBudgetReconciler {
	return &PodDisruptionBudgetReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r PodDisruptionBudgetReconciler) Reconcile(desired *v1beta1.PodDisruptionBudget) error {
	objectInfo := ObjectInfo(desired)
	existingPDB, err := r.getCurrentPodDisruptionBudget(types.NamespacedName{Name: desired.Name, Namespace: r.apiManager.GetNamespace()})
	if err != nil {
		r.Logger().Error(err, fmt.Sprintf("Error reading object %s. Requeuing request...", objectInfo))
		return err
	}

	if r.apiManager.IsPDBEnabled() && existingPDB == nil {
		return r.createResource(desired)
	}

	if r.apiManager.IsPDBEnabled() && existingPDB != nil && !reflect.DeepEqual(desired.Spec, existingPDB.Spec) {
		return r.updateResource(desired)
	}

	if !r.apiManager.IsPDBEnabled() && existingPDB != nil {
		return r.deleteResource(existingPDB)
	}

	return nil
}

func (r PodDisruptionBudgetReconciler) getCurrentPodDisruptionBudget(selector client.ObjectKey) (*v1beta1.PodDisruptionBudget, error) {
	existing := &v1beta1.PodDisruptionBudget{}
	err := r.Client().Get(context.TODO(), selector, existing)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
	} else {
		return existing.DeepCopy(), nil
	}
	return nil, nil
}
