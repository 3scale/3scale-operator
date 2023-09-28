package controllers

import (
	"context"

	capabilitiesv1alpha1 "github.com/3scale/3scale-operator/apis/capabilities/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"github.com/3scale/3scale-operator/pkg/apispkg/common"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type TenantStatusReconciler struct {
	*reconcilers.BaseReconciler
	resourceUpdated  *capabilitiesv1alpha1.Tenant
	resourceOriginal *capabilitiesv1alpha1.Tenant
	reconcileError   error
	logger           logr.Logger
}

func NewTenantStatusReconciler(b *reconcilers.BaseReconciler, resourceUpdated, resourceOriginal *capabilitiesv1alpha1.Tenant, reconcileError error) *TenantStatusReconciler {
	return &TenantStatusReconciler{
		BaseReconciler:   b,
		resourceUpdated:  resourceUpdated,
		resourceOriginal: resourceOriginal,
		reconcileError:   reconcileError,
		logger:           b.Logger().WithValues("Status Reconciler", resourceUpdated.Name),
	}
}

// Compare and update the status of the tenant CR (both, object and object.status)
func (s *TenantStatusReconciler) Reconcile() (reconcile.Result, error) {
	// Check for changes to spec or metadata
	equalSpec := s.resourceOriginal.SpecEqual(s.resourceUpdated, s.logger)
	if !equalSpec {
		s.logger.Info("updating tenant object")
		err := s.Client().Update(context.TODO(), s.resourceUpdated)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{Requeue: true}, nil
	}

	// Check for changes to the status
	newStatus := s.calculateStatus()
	equalStatus := s.resourceOriginal.Status.StatusEqual(&newStatus, s.logger)

	if !equalStatus {
		s.logger.Info("updating tenant status")
		s.resourceUpdated.Status = newStatus
		err := s.UpdateResourceStatus(s.resourceUpdated)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

func (s *TenantStatusReconciler) calculateStatus() capabilitiesv1alpha1.TenantStatus {
	status := s.resourceUpdated.Status
	status.Conditions = s.resourceUpdated.Status.Conditions.Copy()
	status.Conditions.SetCondition(s.readyCondition())
	status.Conditions.SetCondition(s.failedCondition())

	return status
}

func (s *TenantStatusReconciler) readyCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1alpha1.TenantReadyConditionType,
		Status: corev1.ConditionFalse,
	}

	if s.reconcileError == nil && s.resourceUpdated.Status.TenantId != 0 && s.resourceUpdated.Status.AdminId != 0 {
		condition.Status = corev1.ConditionTrue
	}

	return condition
}

func (s *TenantStatusReconciler) failedCondition() common.Condition {
	condition := common.Condition{
		Type:    capabilitiesv1alpha1.TenantFailedConditionType,
		Status:  corev1.ConditionFalse,
		Message: "",
	}

	if s.reconcileError != nil {
		condition.Status = corev1.ConditionTrue
		condition.Message = s.reconcileError.Error()
	}

	return condition
}
