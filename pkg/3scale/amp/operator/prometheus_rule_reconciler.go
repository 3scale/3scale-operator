package operator

import (
	"context"
	"errors"

	"github.com/3scale/3scale-operator/pkg/helper"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
)

// ErrPrometheusRulesNotPresent custom error type
var ErrPrometheusRulesNotPresent = errors.New("no PrometheusRules registered with the API")

type PrometheusRuleReconciler interface {
	IsUpdateNeeded(desired, existing *monitoringv1.PrometheusRule) bool
}

type PrometheusRuleBaseReconciler struct {
	BaseAPIManagerLogicReconciler
	reconciler PrometheusRuleReconciler
}

func NewPrometheusRuleBaseReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler, reconciler PrometheusRuleReconciler) *PrometheusRuleBaseReconciler {
	return &PrometheusRuleBaseReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
		reconciler:                    reconciler,
	}
}

func (r *PrometheusRuleBaseReconciler) Reconcile(desired *monitoringv1.PrometheusRule) error {
	objectInfo := ObjectInfo(desired)

	kindExists, err := r.hasPrometheusRules()
	if err != nil {
		return err
	}
	if !kindExists {
		r.Logger().Info("Install grafana-operator in your cluster to create grafanadashboards objects", "Error creating object", objectInfo)
		return nil
	}

	existing := &monitoringv1.PrometheusRule{}
	err = r.Client().Get(
		context.TODO(),
		types.NamespacedName{Name: desired.Name, Namespace: r.apiManager.GetNamespace()},
		existing)

	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	if IsObjectTaggedTorDelete(desired) {
		if !apierrors.IsNotFound(err) {
			return r.deleteResource(existing)
		}
		// if not found, nothing else to do
		return nil
	}

	if apierrors.IsNotFound(err) {
		return r.createResource(desired)
	}

	update, err := r.isUpdateNeeded(desired, existing)
	if err != nil {
		return err
	}

	if update {
		return r.updateResource(existing)
	}

	return nil
}

// hasPrometheusRules checks if PrometheusRule is registered in the cluster.
func (r *PrometheusRuleBaseReconciler) hasPrometheusRules() (bool, error) {
	dc := discovery.NewDiscoveryClientForConfigOrDie(r.cfg)

	return k8sutil.ResourceExists(dc,
		monitoringv1.SchemeGroupVersion.String(),
		monitoringv1.PrometheusRuleKind)
}

func (r *PrometheusRuleBaseReconciler) isUpdateNeeded(desired, existing *monitoringv1.PrometheusRule) (bool, error) {
	updated := helper.EnsureObjectMeta(&existing.ObjectMeta, &desired.ObjectMeta)

	updatedTmp, err := r.ensureOwnerReference(existing)
	if err != nil {
		return false, nil
	}

	updated = updated || updatedTmp

	updatedTmp = r.reconciler.IsUpdateNeeded(desired, existing)
	updated = updated || updatedTmp

	return updated, nil
}

type CreateOnlyPrometheusRuleReconciler struct {
}

func NewCreateOnlyPrometheusRuleReconciler() *CreateOnlyPrometheusRuleReconciler {
	return &CreateOnlyPrometheusRuleReconciler{}
}

func (r *CreateOnlyPrometheusRuleReconciler) IsUpdateNeeded(desired, existing *monitoringv1.PrometheusRule) bool {
	return false
}
