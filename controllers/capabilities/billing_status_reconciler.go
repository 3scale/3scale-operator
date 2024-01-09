package controllers

import (
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/apispkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type BillingStatusReconciler struct {
	*reconcilers.BaseReconciler
	resource            *capabilitiesv1beta1.Billing
	providerAccountHost string
	reconcileError      error
	logger              logr.Logger
}

func NewBillingStatusReconciler(b *reconcilers.BaseReconciler, resource *capabilitiesv1beta1.Billing, providerAccountHost string, reconcileError error) *BillingStatusReconciler {
	return &BillingStatusReconciler{
		BaseReconciler:      b,
		resource:            resource,
		providerAccountHost: providerAccountHost,
		reconcileError:      reconcileError,
		logger:              b.Logger().WithValues("Status Reconciler", resource.Name),
	}
}

func (s *BillingStatusReconciler) Reconcile() (reconcile.Result, error) {
	s.logger.V(1).Info("START")

	newStatus, err := s.calculateStatus()
	if err != nil {
		return reconcile.Result{}, err
	}

	equalStatus := s.resource.Status.Equals(newStatus, s.logger)
	s.logger.V(1).Info("Status", "status is different", !equalStatus)
	s.logger.V(1).Info("Status", "generation is different", s.resource.Generation != s.resource.Status.ObservedGeneration)
	if equalStatus && s.resource.Generation == s.resource.Status.ObservedGeneration {
		// Steady state
		s.logger.V(1).Info("Status steady state, status was not updated")
		return reconcile.Result{}, nil
	}

	// Save the generation number we acted on, otherwise we might wrongfully indicate
	// that we've seen a spec update when we retry.
	// TODO: This can clobber an update if we allow multiple agents to write to the
	// same status.
	newStatus.ObservedGeneration = s.resource.Generation

	s.logger.V(1).Info("Updating Status", "sequence no:", fmt.Sprintf("sequence No: %v->%v", s.resource.Status.ObservedGeneration, newStatus.ObservedGeneration))

	s.resource.Status = *newStatus
	updateErr := s.Client().Status().Update(s.Context(), s.resource)
	if updateErr != nil {
		// Ignore conflicts, resource might just be outdated.
		if errors.IsConflict(updateErr) {
			s.logger.Info("Failed to update status: resource might just be outdated")
			return reconcile.Result{Requeue: true}, nil
		}

		return reconcile.Result{}, fmt.Errorf("Failed to update status: %w", updateErr)
	}
	return reconcile.Result{}, nil
}

func (s *BillingStatusReconciler) calculateStatus() (*capabilitiesv1beta1.BillingStatus, error) {
	// Initialize with existing data for data coming from 3scale
	// just in case in this reconciliation loop something goes wrong and avoid replacing right data with nil
	newStatus := &capabilitiesv1beta1.BillingStatus{
		ID:                 s.resource.Status.ID,
		TenantAccountID:    s.resource.Status.TenantAccountID,
		BillDate:           s.resource.Status.BillDate,
		Conditions:         s.resource.Status.Conditions.Copy(),
		ObservedGeneration: s.resource.Status.ObservedGeneration,
	}

	newStatus.Conditions.SetCondition(s.invalidCondition())
	newStatus.Conditions.SetCondition(s.readyCondition())
	newStatus.Conditions.SetCondition(s.waitingCondition())
	newStatus.Conditions.SetCondition(s.failedCondition())

	return newStatus, nil
}

func (s *BillingStatusReconciler) readyCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.BillingReadyConditionType,
		Status: corev1.ConditionFalse,
	}

	if s.reconcileError == nil {
		condition.Status = corev1.ConditionTrue
	}

	return condition
}

func (s *BillingStatusReconciler) invalidCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.BillingInvalidConditionType,
		Status: corev1.ConditionFalse,
	}

	if helper.IsInvalidSpecError(s.reconcileError) {
		condition.Status = corev1.ConditionTrue
		condition.Message = s.reconcileError.Error()
	}

	return condition
}

func (s *BillingStatusReconciler) failedCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.BillingFailedConditionType,
		Status: corev1.ConditionFalse,
	}

	if s.reconcileError != nil {
		// only activate this condition when others are false and still there is an error

		otherConditionsFalse := []bool{
			s.invalidCondition().IsFalse(),
			s.waitingCondition().IsFalse(),
		}

		if helper.All(otherConditionsFalse) {
			condition.Status = corev1.ConditionTrue
			condition.Message = s.reconcileError.Error()
		}
	}

	return condition
}

func (s *BillingStatusReconciler) waitingCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.BillingWaitingConditionType,
		Status: corev1.ConditionFalse,
	}

	if helper.IsWaitError(s.reconcileError) {
		condition.Status = corev1.ConditionTrue
		condition.Message = s.reconcileError.Error()
	}

	return condition
}
