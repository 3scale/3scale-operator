package controllers

import (
	capabilitiesv1alpha1 "github.com/3scale/3scale-operator/apis/capabilities/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"github.com/3scale/3scale-operator/pkg/apispkg/common"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
)

type TenantStatusReconciler struct {
	*reconcilers.BaseReconciler
	tenantResource *capabilitiesv1alpha1.Tenant
	reconcileError error
	logger         logr.Logger
}

func NewTenantStatusReconciler(b *reconcilers.BaseReconciler, tenantResource *capabilitiesv1alpha1.Tenant, reconcileError error) *TenantStatusReconciler {
	return &TenantStatusReconciler{
		BaseReconciler: b,
		tenantResource: tenantResource,
		reconcileError: reconcileError,
		logger:         b.Logger().WithValues("Status Reconciler", tenantResource.Name),
	}
}

// Status reconciler returns an error if there's an error updating the resource. If resource is updated, due to watchers, a new reconcile will trigger
// therefore no manual retrigger of reconcile is required
func (s *TenantStatusReconciler) Reconcile() (bool, error) {
	// Check for changes to the status
	newStatus := s.calculateStatus()
	equalStatus := s.tenantResource.Status.StatusEqual(&newStatus, s.logger)

	if !equalStatus {
		s.logger.Info("updating tenant status")
		s.tenantResource.Status = newStatus
		err := s.UpdateResourceStatus(s.tenantResource)
		if err != nil {
			s.logger.Info("ERROR", "error in tenant status reconciler when updating the resource", err)
			return equalStatus, err
		}
	}

	return equalStatus, nil
}

func (s *TenantStatusReconciler) calculateStatus() capabilitiesv1alpha1.TenantStatus {
	status := s.tenantResource.Status
	status.Conditions = s.tenantResource.Status.Conditions.Copy()
	status.Conditions.SetCondition(s.readyCondition())

	return status
}

func (s *TenantStatusReconciler) readyCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1alpha1.TenantReadyConditionType,
		Status: corev1.ConditionFalse,
	}

	if s.reconcileError == nil && s.tenantResource.Status.TenantId != 0 && s.tenantResource.Status.AdminId != 0 {
		condition.Status = corev1.ConditionTrue
	}

	if s.reconcileError != nil {
		condition.Message = s.reconcileError.Error()
	}

	return condition
}
