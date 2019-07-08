package operator

import (
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type BaseLogicReconciler struct {
	BaseReconciler
}

type LogicReconciler interface {
	Reconcile() (reconcile.Result, error)
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &BaseLogicReconciler{}

func NewBaseLogicReconciler(b BaseReconciler) BaseLogicReconciler {
	return BaseLogicReconciler{
		BaseReconciler: b,
	}
}

func (r BaseLogicReconciler) Reconcile() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}
