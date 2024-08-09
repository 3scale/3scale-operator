package operator

import (
	"fmt"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/go-logr/logr"
	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	grafanav1beta1 "github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	routev1 "github.com/openshift/api/route/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	k8sappsv1 "k8s.io/api/apps/v1"
	hpa "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type BaseAPIManagerLogicReconciler struct {
	*reconcilers.BaseReconciler
	apiManager           *appsv1alpha1.APIManager
	logger               logr.Logger
	crdAvailabilityCache *baseAPIManagerLogicReconcilerCRDAvailabilityCache
}

type baseAPIManagerLogicReconcilerCRDAvailabilityCache struct {
	grafanaDashboardCRDV5Available *bool
	grafanaDashboardCRDV4Available *bool
	prometheusRuleCRDAvailable     *bool
	podMonitorCRDAvailable         *bool
	serviceMonitorCRDAvailable     *bool
}

func NewBaseAPIManagerLogicReconciler(b *reconcilers.BaseReconciler, apiManager *appsv1alpha1.APIManager) *BaseAPIManagerLogicReconciler {
	return &BaseAPIManagerLogicReconciler{
		BaseReconciler:       b,
		apiManager:           apiManager,
		logger:               b.Logger().WithValues("APIManager Controller", apiManager.Name),
		crdAvailabilityCache: &baseAPIManagerLogicReconcilerCRDAvailabilityCache{},
	}
}

func (r *BaseAPIManagerLogicReconciler) NamespacedNameWithAPIManagerNamespace(obj metav1.Object) types.NamespacedName {
	return types.NamespacedName{Namespace: r.apiManager.GetNamespace(), Name: obj.GetName()}
}

func (r *BaseAPIManagerLogicReconciler) ReconcilePodDisruptionBudget(desired *policyv1.PodDisruptionBudget, mutatefn reconcilers.MutateFn) error {
	if !r.apiManager.IsPDBEnabled() {
		common.TagObjectToDelete(desired)
	}
	return r.ReconcileResource(&policyv1.PodDisruptionBudget{}, desired, mutatefn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileDeployment(desired *k8sappsv1.Deployment, mutatefn reconcilers.MutateFn) error {
	return r.ReconcileResource(&k8sappsv1.Deployment{}, desired, mutatefn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileService(desired *v1.Service, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&v1.Service{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileConfigMap(desired *v1.ConfigMap, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&v1.ConfigMap{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileServiceAccount(desired *v1.ServiceAccount, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&v1.ServiceAccount{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileRoute(desired *routev1.Route, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&routev1.Route{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileSecret(desired *v1.Secret, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&v1.Secret{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcilePersistentVolumeClaim(desired *v1.PersistentVolumeClaim, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&v1.PersistentVolumeClaim{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileRole(desired *rbacv1.Role, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&rbacv1.Role{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileRoleBinding(desired *rbacv1.RoleBinding, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&rbacv1.RoleBinding{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileJob(desired *batchv1.Job, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&batchv1.Job{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileGrafanaDashboards(
	desired interface{},
	mutateFn reconcilers.MutateFn,
) error {
	// Check for CRD availability
	dashboardsAvailable, err := r.HasGrafanaDashboards()
	if err != nil {
		return err
	}

	// Reconcile based on the type of the desired object
	switch d := desired.(type) {
	case *grafanav1beta1.GrafanaDashboard:
		if dashboardsAvailable && *r.crdAvailabilityCache.grafanaDashboardCRDV5Available {
			if !r.apiManager.IsMonitoringEnabled() {
				common.TagObjectToDelete(d)
			}
			return r.ReconcileResource(d, d, mutateFn)
		}

	case *grafanav1alpha1.GrafanaDashboard:
		if dashboardsAvailable && *r.crdAvailabilityCache.grafanaDashboardCRDV4Available {
			if !r.apiManager.IsMonitoringEnabled() {
				common.TagObjectToDelete(d)
			}
			return r.ReconcileResource(d, d, mutateFn)
		}

	default:
		return fmt.Errorf("unsupported GrafanaDashboard type: %T", d)
	}

	// Log error only if neither v4 nor v5 CRDs are available
	if !dashboardsAvailable && r.apiManager.IsMonitoringEnabled() {
		errToLog := fmt.Errorf("Error creating grafana dashboard object '%s'. Install grafana-operator in your cluster to create grafana dashboard objects", getName(desired))
		r.EventRecorder().Eventf(r.apiManager, v1.EventTypeWarning, "ReconcileError", errToLog.Error())
		r.logger.Error(errToLog, "ReconcileError")
	}

	return nil
}

func getName(desired interface{}) string {
	switch d := desired.(type) {
	case *grafanav1beta1.GrafanaDashboard:
		return d.Name
	case *grafanav1alpha1.GrafanaDashboard:
		return d.Name
	default:
		return "unknown"
	}
}

func (r *BaseAPIManagerLogicReconciler) ReconcilePrometheusRules(desired *monitoringv1.PrometheusRule, mutateFn reconcilers.MutateFn) error {
	kindExists, err := r.HasPrometheusRules()
	if err != nil {
		return err
	}

	if !kindExists {
		if r.apiManager.IsPrometheusRulesEnabled() {
			errToLog := fmt.Errorf("Error creating prometheusrule object '%s'. Install prometheus-operator in your cluster to create prometheusrule objects", desired.Name)
			r.EventRecorder().Eventf(r.apiManager, v1.EventTypeWarning, "ReconcileError", errToLog.Error())
			r.logger.Error(errToLog, "ReconcileError")
		}
		return nil
	}

	if !r.apiManager.IsPrometheusRulesEnabled() {
		common.TagObjectToDelete(desired)
	}
	return r.ReconcileResource(&monitoringv1.PrometheusRule{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileServiceMonitor(desired *monitoringv1.ServiceMonitor, mutateFn reconcilers.MutateFn) error {
	kindExists, err := r.HasServiceMonitors()
	if err != nil {
		return err
	}

	if !kindExists {
		if r.apiManager.IsMonitoringEnabled() {
			errToLog := fmt.Errorf("Error creating servicemonitor object '%s'. Install prometheus-operator in your cluster to create servicemonitor objects", desired.Name)
			r.EventRecorder().Eventf(r.apiManager, v1.EventTypeWarning, "ReconcileError", errToLog.Error())
			r.logger.Error(errToLog, "ReconcileError")
		}
		return nil
	}

	if !r.apiManager.IsMonitoringEnabled() {
		common.TagObjectToDelete(desired)
	}
	return r.ReconcileResource(&monitoringv1.ServiceMonitor{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileHpa(desired *hpa.HorizontalPodAutoscaler, mutateFn reconcilers.MutateFn) error {
	// Only allow to create Apicast HPA regardless of running async mode or not.
	if desired.Spec.ScaleTargetRef.Name == component.ApicastProductionName && r.apiManager.Spec.Apicast.ProductionSpec.Hpa {
		return r.ReconcileResource(&hpa.HorizontalPodAutoscaler{}, desired, mutateFn)
	}
	// Only create the HPA objects if async is not disabled
	if !r.apiManager.IsAsyncDisableAnnotationPresent() {
		if desired.Spec.ScaleTargetRef.Name == component.BackendListenerName && r.apiManager.Spec.Backend.ListenerSpec.Hpa {
			return r.ReconcileResource(&hpa.HorizontalPodAutoscaler{}, desired, mutateFn)
		}
		if desired.Spec.ScaleTargetRef.Name == component.BackendWorkerName && r.apiManager.Spec.Backend.WorkerSpec.Hpa {
			return r.ReconcileResource(&hpa.HorizontalPodAutoscaler{}, desired, mutateFn)
		}
	}
	// Delete HPA in all other cases
	err := r.DeleteResource(desired)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

func (r *BaseAPIManagerLogicReconciler) ReconcilePodMonitor(desired *monitoringv1.PodMonitor, mutateFn reconcilers.MutateFn) error {
	kindExists, err := r.HasPodMonitors()
	if err != nil {
		return err
	}

	if !kindExists {
		if r.apiManager.IsMonitoringEnabled() {
			errToLog := fmt.Errorf("Error creating podmonitor object '%s'. Install prometheus-operator in your cluster to create podmonitor objects", desired.Name)
			r.EventRecorder().Eventf(r.apiManager, v1.EventTypeWarning, "ReconcileError", errToLog.Error())
			r.logger.Error(errToLog, "ReconcileError")
		}
		return nil
	}

	if !r.apiManager.IsMonitoringEnabled() {
		common.TagObjectToDelete(desired)
	}
	return r.ReconcileResource(&monitoringv1.PodMonitor{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileResource(obj, desired common.KubernetesObject, mutatefn reconcilers.MutateFn) error {
	desired.SetNamespace(r.apiManager.GetNamespace())

	// Secrets are managed by users so they do not get APIManager-based
	// owned references. In case we want to react to changes to secrets
	// in the future we will need to implement an alternative mechanism to
	// controller-based OwnerReferences due to user-managed secrets might
	// already have controller-based OwnerReferences and K8s objects
	// can only be owned by a single controller-based OwnerReference.
	if desired.GetObjectKind().GroupVersionKind().Kind != "Secret" {
		if err := r.SetControllerOwnerReference(r.apiManager, desired); err != nil {
			return err
		}
	}

	return r.BaseReconciler.ReconcileResource(obj, desired, r.APIManagerMutator(mutatefn))
}

// APIManagerMutator wraps mutator into APIManger mutator
// All resources managed by APIManager are processed by this wrapped mutator
func (r *BaseAPIManagerLogicReconciler) APIManagerMutator(mutateFn reconcilers.MutateFn) reconcilers.MutateFn {
	return func(existing, desired common.KubernetesObject) (bool, error) {
		// Metadata
		updated := helper.EnsureObjectMeta(existing, desired)

		// Secrets are managed by users so they do not get APIManager-based
		// owned references. In case we want to react to changes to secrets
		// in the future we will need to implement an alternative mechanism to
		// controller-based OwnerReferences due to user-managed secrets might
		// already have controller-based OwnerReferences and K8s objects
		// can only be owned by a single controller-based OwnerReference.
		if existing.GetObjectKind().GroupVersionKind().Kind != "Secret" {
			// OwnerRefenrence
			updatedTmp, err := r.EnsureOwnerReference(r.apiManager, existing)
			if err != nil {
				return false, err
			}
			updated = updated || updatedTmp
		}

		updatedTmp, err := mutateFn(existing, desired)
		if err != nil {
			return false, err
		}
		updated = updated || updatedTmp

		return updated, nil
	}
}

func (r *BaseAPIManagerLogicReconciler) Logger() logr.Logger {
	return r.logger
}

func (b *BaseAPIManagerLogicReconciler) HasGrafanaDashboards() (bool, error) {
	// Check if we need to update the cache
	if b.crdAvailabilityCache.grafanaDashboardCRDV4Available == nil || b.crdAvailabilityCache.grafanaDashboardCRDV5Available == nil {
		var err error
		// Check for Grafana V4 Dashboards
		v4Available, err := b.BaseReconciler.HasGrafanaV4Dashboards()
		if err != nil {
			return false, err
		}
		b.crdAvailabilityCache.grafanaDashboardCRDV4Available = &v4Available

		// Check for Grafana V5 Dashboards
		v5Available, err := b.BaseReconciler.HasGrafanaV5Dashboards()
		if err != nil {
			return false, err
		}
		b.crdAvailabilityCache.grafanaDashboardCRDV5Available = &v5Available
	}

	// Combine the results for both versions
	dashboardsAvailable := *b.crdAvailabilityCache.grafanaDashboardCRDV4Available || *b.crdAvailabilityCache.grafanaDashboardCRDV5Available
	return dashboardsAvailable, nil
}

// HasPrometheusRules checks if the PrometheusRules CRD is supported in current cluster
func (b *BaseAPIManagerLogicReconciler) HasPrometheusRules() (bool, error) {
	if b.crdAvailabilityCache.prometheusRuleCRDAvailable == nil {
		res, err := b.BaseReconciler.HasPrometheusRules()
		if err != nil {
			return res, err
		}
		b.crdAvailabilityCache.prometheusRuleCRDAvailable = &res
		return res, err
	}
	return *b.crdAvailabilityCache.prometheusRuleCRDAvailable, nil
}

func (b *BaseAPIManagerLogicReconciler) HasServiceMonitors() (bool, error) {
	if b.crdAvailabilityCache.serviceMonitorCRDAvailable == nil {
		res, err := b.BaseReconciler.HasServiceMonitors()
		if err != nil {
			return res, err
		}
		b.crdAvailabilityCache.serviceMonitorCRDAvailable = &res
		return res, err
	}
	return *b.crdAvailabilityCache.serviceMonitorCRDAvailable, nil
}

func (b *BaseAPIManagerLogicReconciler) HasPodMonitors() (bool, error) {
	if b.crdAvailabilityCache.podMonitorCRDAvailable == nil {
		res, err := b.BaseReconciler.HasPodMonitors()
		if err != nil {
			return res, err
		}
		b.crdAvailabilityCache.podMonitorCRDAvailable = &res
		return res, err
	}
	return *b.crdAvailabilityCache.podMonitorCRDAvailable, nil
}
