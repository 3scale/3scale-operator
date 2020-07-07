package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/go-logr/logr"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type BaseAPIManagerLogicReconciler struct {
	*reconcilers.BaseReconciler
	apiManager *appsv1alpha1.APIManager
	logger     logr.Logger
}

func NewBaseAPIManagerLogicReconciler(b *reconcilers.BaseReconciler, apiManager *appsv1alpha1.APIManager) *BaseAPIManagerLogicReconciler {
	return &BaseAPIManagerLogicReconciler{
		BaseReconciler: b,
		apiManager:     apiManager,
		logger:         b.Logger().WithValues("APIManager Controller", apiManager.Name),
	}
}

func (r *BaseAPIManagerLogicReconciler) NamespacedNameWithAPIManagerNamespace(obj metav1.Object) types.NamespacedName {
	return types.NamespacedName{Namespace: r.apiManager.GetNamespace(), Name: obj.GetName()}
}

func (r *BaseAPIManagerLogicReconciler) setOwnerReference(obj common.KubernetesObject) error {
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

func (r *BaseAPIManagerLogicReconciler) ensureOwnerReference(obj common.KubernetesObject) (bool, error) {
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

func (r *BaseAPIManagerLogicReconciler) ReconcilePodDisruptionBudget(desired *v1beta1.PodDisruptionBudget, mutatefn reconcilers.MutateFn) error {
	if !r.apiManager.IsPDBEnabled() {
		common.TagObjectToDelete(desired)
	}
	return r.ReconcileResource(&v1beta1.PodDisruptionBudget{}, desired, mutatefn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileImagestream(desired *imagev1.ImageStream, mutatefn reconcilers.MutateFn) error {
	return r.ReconcileResource(&imagev1.ImageStream{}, desired, mutatefn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileDeploymentConfig(desired *appsv1.DeploymentConfig, mutatefn reconcilers.MutateFn) error {
	return r.ReconcileResource(&appsv1.DeploymentConfig{}, desired, mutatefn)
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

func (r *BaseAPIManagerLogicReconciler) ReconcileGrafanaDashboard(desired *grafanav1alpha1.GrafanaDashboard, mutateFn reconcilers.MutateFn) error {
	kindExists, err := r.hasGrafanaDashboards()
	if err != nil {
		return err
	}

	if !kindExists {
		r.Logger().Info("Install grafana-operator in your cluster to create grafanadashboards objects", "Error creating grafanadashboard object", desired.Name)
		return nil
	}

	if !r.apiManager.IsMonitoringEnabled() {
		common.TagObjectToDelete(desired)
	}
	return r.ReconcileResource(&grafanav1alpha1.GrafanaDashboard{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) hasGrafanaDashboards() (bool, error) {
	return k8sutil.ResourceExists(r.DiscoveryClient(),
		grafanav1alpha1.SchemeGroupVersion.String(),
		grafanav1alpha1.GrafanaDashboardKind)
}

func (r *BaseAPIManagerLogicReconciler) ReconcilePrometheusRules(desired *monitoringv1.PrometheusRule, mutateFn reconcilers.MutateFn) error {
	kindExists, err := r.hasPrometheusRules()
	if err != nil {
		return err
	}
	if !kindExists {
		r.Logger().Info("Install prometheus-operator in your cluster to create prometheusrules objects", "Error creating prometheusrule object", desired.Name)
		return nil
	}

	if !r.apiManager.IsMonitoringEnabled() {
		common.TagObjectToDelete(desired)
	}
	return r.ReconcileResource(&monitoringv1.PrometheusRule{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) hasPrometheusRules() (bool, error) {
	return k8sutil.ResourceExists(r.DiscoveryClient(),
		monitoringv1.SchemeGroupVersion.String(),
		monitoringv1.PrometheusRuleKind)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileServiceMonitor(desired *monitoringv1.ServiceMonitor, mutateFn reconcilers.MutateFn) error {
	kindExists, err := r.hasServiceMonitors()
	if err != nil {
		return err
	}

	if !kindExists {
		r.Logger().Info("Install prometheus-operator in your cluster to create servicemonitor objects", "Error creating servicemonitor object", desired.Name)
		return nil
	}

	if !r.apiManager.IsMonitoringEnabled() {
		common.TagObjectToDelete(desired)
	}
	return r.ReconcileResource(&monitoringv1.ServiceMonitor{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) hasServiceMonitors() (bool, error) {
	return k8sutil.ResourceExists(r.DiscoveryClient(),
		monitoringv1.SchemeGroupVersion.String(),
		monitoringv1.ServiceMonitorsKind)
}

func (r *BaseAPIManagerLogicReconciler) ReconcilePodMonitor(desired *monitoringv1.PodMonitor, mutateFn reconcilers.MutateFn) error {
	kindExists, err := r.hasPodMonitors()
	if err != nil {
		return err
	}

	if !kindExists {
		r.Logger().Info("Install prometheus-operator in your cluster to create podmonitor objects", "Error creating podmonitor object", desired.Name)
		return nil
	}

	if !r.apiManager.IsMonitoringEnabled() {
		common.TagObjectToDelete(desired)
	}
	return r.ReconcileResource(&monitoringv1.PodMonitor{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) hasPodMonitors() (bool, error) {
	return k8sutil.ResourceExists(r.DiscoveryClient(),
		monitoringv1.SchemeGroupVersion.String(),
		monitoringv1.PodMonitorsKind)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileResource(obj, desired common.KubernetesObject, mutatefn reconcilers.MutateFn) error {
	desired.SetNamespace(r.apiManager.GetNamespace())
	if err := r.setOwnerReference(desired); err != nil {
		return err
	}

	return r.BaseReconciler.ReconcileResource(obj, desired, r.APIManagerMutator(mutatefn))
}

// APIManagerMutator wraps mutator into APIManger mutator
// All resources managed by APIManager are processed by this wrapped mutator
func (r *BaseAPIManagerLogicReconciler) APIManagerMutator(mutateFn reconcilers.MutateFn) reconcilers.MutateFn {
	return func(existing, desired common.KubernetesObject) (bool, error) {
		// Metadata
		updated := helper.EnsureObjectMeta(existing, desired)

		// OwnerRefenrence
		updatedTmp, err := r.ensureOwnerReference(existing)
		if err != nil {
			return false, err
		}
		updated = updated || updatedTmp

		updatedTmp, err = mutateFn(existing, desired)
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
