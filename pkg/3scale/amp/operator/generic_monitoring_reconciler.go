package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type GenericMonitoringReconciler struct {
	BaseAPIManagerLogicReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &GenericMonitoringReconciler{}

func NewGenericMonitoringReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) *GenericMonitoringReconciler {
	return &GenericMonitoringReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *GenericMonitoringReconciler) Reconcile() (reconcile.Result, error) {
	err := r.reconcileGrafanaDashboard(component.KubernetesResourcesByNamespaceGrafanaDashboard(r.apiManager.Namespace))
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileGrafanaDashboard(component.KubernetesResourcesByPodGrafanaDashboard(r.apiManager.Namespace))
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcilePrometheusRules(component.KubeStateMetricsPrometheusRules(r.apiManager.Namespace))
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
