package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
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
	sumRate, err := helper.SumRateForOpenshiftVersion(r.Context(), r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	grafanaDashboard := component.KubernetesResourcesByNamespaceGrafanaV5Dashboard(sumRate, r.apiManager.Namespace, *r.apiManager.Spec.AppLabel)
	err = r.ReconcileGrafanaDashboards(grafanaDashboard, reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	grafanaDashboard = component.KubernetesResourcesByPodGrafanaV5Dashboard(sumRate, r.apiManager.Namespace, *r.apiManager.Spec.AppLabel)
	err = r.ReconcileGrafanaDashboards(grafanaDashboard, reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	grafanaV4Dashboard := component.KubernetesResourcesByNamespaceGrafanaV4Dashboard(sumRate, r.apiManager.Namespace, *r.apiManager.Spec.AppLabel)
	err = r.ReconcileGrafanaDashboards(grafanaV4Dashboard, reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	grafanaV4Dashboard = component.KubernetesResourcesByPodGrafanaV4Dashboard(sumRate, r.apiManager.Namespace, *r.apiManager.Spec.AppLabel)
	err = r.ReconcileGrafanaDashboards(grafanaV4Dashboard, reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	prometheusRule := component.KubeStateMetricsPrometheusRules(sumRate, r.apiManager.Namespace, *r.apiManager.Spec.AppLabel)
	err = r.ReconcilePrometheusRules(prometheusRule, reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
