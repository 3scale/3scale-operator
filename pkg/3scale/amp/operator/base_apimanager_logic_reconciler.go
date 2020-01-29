package operator

import (
	"context"
	"fmt"
	"k8s.io/api/policy/v1beta1"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type BaseAPIManagerLogicReconciler struct {
	BaseLogicReconciler
	apiManager *appsv1alpha1.APIManager
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &BaseAPIManagerLogicReconciler{}

func NewBaseAPIManagerLogicReconciler(b BaseLogicReconciler, apiManager *appsv1alpha1.APIManager) BaseAPIManagerLogicReconciler {
	return BaseAPIManagerLogicReconciler{
		BaseLogicReconciler: b,
		apiManager:          apiManager,
	}
}

func (r BaseAPIManagerLogicReconciler) Reconcile() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func (r BaseAPIManagerLogicReconciler) NamespacedNameWithAPIManagerNamespace(obj metav1.Object) types.NamespacedName {
	return types.NamespacedName{Namespace: r.apiManager.GetNamespace(), Name: obj.GetName()}
}

func (r BaseAPIManagerLogicReconciler) setOwnerReference(obj common.KubernetesObject) error {
	err := controllerutil.SetControllerReference(r.apiManager, obj, r.Scheme())
	if err != nil {
		r.Logger().Error(err, "Error setting OwnerReference on object",
			"Kind", obj.GetObjectKind().GroupVersionKind().String(),
			"Namespace", obj.GetNamespace(),
			"Name", obj.GetName(),
		)
	}
	return err
}

func (r BaseAPIManagerLogicReconciler) ensureOwnerReference(obj common.KubernetesObject) (bool, error) {
	changed := false

	originalSize := len(obj.GetOwnerReferences())
	err := r.setOwnerReference(obj)
	if err != nil {
		return false, err
	}

	newSize := len(obj.GetOwnerReferences())
	if originalSize != newSize {
		changed = true
	}

	return changed, nil
}

func (r *BaseAPIManagerLogicReconciler) createResource(obj common.KubernetesObject) error {
	obj.SetNamespace(r.apiManager.GetNamespace())
	if err := r.setOwnerReference(obj); err != nil {
		return err
	}

	r.Logger().Info(fmt.Sprintf("Created object %s", ObjectInfo(obj)))
	return r.Client().Create(context.TODO(), obj) // don't wrap error
}

func (r *BaseAPIManagerLogicReconciler) updateResource(obj common.KubernetesObject) error {
	if err := r.setOwnerReference(obj); err != nil {
		return err
	}

	r.Logger().Info(fmt.Sprintf("Updated object %s", ObjectInfo(obj)))
	return r.Client().Update(context.TODO(), obj) // don't wrap error
}

func (r *BaseAPIManagerLogicReconciler) deleteResource(obj common.KubernetesObject) error {
	r.Logger().Info(fmt.Sprintf("Delete object %s", ObjectInfo(obj)))
	return r.Client().Delete(context.TODO(), obj)
}

func (r *BaseAPIManagerLogicReconciler) reconcilePodDisruptionBudget(desiredPDB *v1beta1.PodDisruptionBudget) error {
	reconciler := NewPodDisruptionBudgetReconciler(*r)
	return reconciler.Reconcile(desiredPDB)
}
