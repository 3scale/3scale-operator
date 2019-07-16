package operator

import (
	"context"
	"fmt"
	"reflect"

	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type RouteReconciler struct {
	BaseReconciler
	ObjectMetaMerger ObjectMetaMerger
}

func NewRouteReconciler(baseReconciler BaseReconciler, objectMetaMerger ObjectMetaMerger) RouteReconciler {
	return RouteReconciler{
		BaseReconciler:   baseReconciler,
		ObjectMetaMerger: objectMetaMerger,
	}
}

func (r *RouteReconciler) Reconcile(desiredRoute *routev1.Route) error {
	objectInfo := ObjectInfo(desiredRoute)

	existingRoute := &routev1.Route{}
	err := r.Client().Get(context.TODO(), types.NamespacedName{Name: desiredRoute.Name, Namespace: desiredRoute.Namespace}, existingRoute)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.Client().Create(context.TODO(), desiredRoute)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			r.Logger().Info(fmt.Sprintf("Created object %s", objectInfo))
			return nil
		}
		return err
	}

	needsUpdate, err := r.ensureRoute(existingRoute, desiredRoute)
	if err != nil {
		return err
	}

	if needsUpdate {
		r.Logger().Info(fmt.Sprintf("Updating Route %s", objectInfo))
		err := r.Client().Update(context.TODO(), existingRoute)
		if err != nil {
			r.Logger().Error(err, fmt.Sprintf("error updating Service %s", objectInfo))
			return err
		}
	}

	return nil
}

func (r *RouteReconciler) ensureRoute(updated, desired *routev1.Route) (bool, error) {
	changed := false

	objectMetaChanged, err := r.ObjectMetaMerger.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta)
	if err != nil {
		return false, err
	}
	if objectMetaChanged {
		changed = true
	}

	// Set in the desired some fields that are automatically set
	// by Kubernetes controllers as defaults that are not defined in
	// our logic
	desired.Spec.WildcardPolicy = updated.Spec.WildcardPolicy
	desired.Spec.To.Weight = updated.Spec.To.Weight

	if !reflect.DeepEqual(updated.Spec, desired.Spec) {
		updated.Spec = desired.Spec
		changed = true
	}

	return changed, nil
}
