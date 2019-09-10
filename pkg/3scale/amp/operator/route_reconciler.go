package operator

import (
	"context"
	"fmt"

	"github.com/3scale/3scale-operator/pkg/helper"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type RoutesReconciler interface {
	IsUpdateNeeded(desired, existing *routev1.Route) bool
}

type RouteBaseReconciler struct {
	BaseAPIManagerLogicReconciler
	reconciler RoutesReconciler
}

func NewRouteBaseReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler, reconciler RoutesReconciler) *RouteBaseReconciler {
	return &RouteBaseReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
		reconciler:                    reconciler,
	}
}

func (r *RouteBaseReconciler) Reconcile(desired *routev1.Route) error {
	objectInfo := ObjectInfo(desired)
	existing := &routev1.Route{}
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

func (r *RouteBaseReconciler) isUpdateNeeded(desired, existing *routev1.Route) (bool, error) {
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

type CreateOnlyRouteReconciler struct {
}

func NewCreateOnlyRouteReconciler() *CreateOnlyRouteReconciler {
	return &CreateOnlyRouteReconciler{}
}

func (r *CreateOnlyRouteReconciler) IsUpdateNeeded(desired, existing *routev1.Route) bool {
	return false
}
