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
	grafanaDashboard := component.KubernetesResourcesByNamespaceGrafanaDashboard(r.apiManager.Namespace, *r.apiManager.Spec.AppLabel)
	err := r.ReconcileGrafanaDashboard(grafanaDashboard, reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	grafanaDashboard = component.KubernetesResourcesByPodGrafanaDashboard(r.apiManager.Namespace, *r.apiManager.Spec.AppLabel)
	err = r.ReconcileGrafanaDashboard(grafanaDashboard, reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	prometheusRule := component.KubeStateMetricsPrometheusRules(r.apiManager.Namespace, *r.apiManager.Spec.AppLabel)
	err = r.ReconcilePrometheusRules(prometheusRule, reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
