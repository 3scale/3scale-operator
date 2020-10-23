package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type GenericMonitoringReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewGenericMonitoringReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *GenericMonitoringReconciler {
	return &GenericMonitoringReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *GenericMonitoringReconciler) Reconcile() (reconcile.Result, error) {
	err := r.ReconcileGrafanaDashboard(component.KubernetesResourcesByNamespaceGrafanaDashboard(r.apiManager.Namespace), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileGrafanaDashboard(component.KubernetesResourcesByPodGrafanaDashboard(r.apiManager.Namespace), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(component.KubeStateMetricsPrometheusRules(r.apiManager.Namespace), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
