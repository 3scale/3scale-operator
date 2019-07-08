package operator

import (
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type BaseControllerReconciler struct {
	BaseReconciler
}

func NewBaseControllerReconciler(b BaseReconciler) BaseControllerReconciler {
	return BaseControllerReconciler{
		BaseReconciler: b,
	}
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &BaseControllerReconciler{}

func (r *BaseControllerReconciler) Reconcile(reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}
