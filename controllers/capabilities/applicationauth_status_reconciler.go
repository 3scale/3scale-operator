package controllers

import (
	"fmt"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/apispkg/common"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ApplicationAuthStatusReconciler struct {
	*reconcilers.BaseReconciler
	resource       *capabilitiesv1beta1.ApplicationAuth
	reconcileError error
	logger         logr.Logger
}

func NewApplicationAuthStatusReconciler(b *reconcilers.BaseReconciler, resource *capabilitiesv1beta1.ApplicationAuth, reconcileError error) *ApplicationAuthStatusReconciler {
	return &ApplicationAuthStatusReconciler{
		BaseReconciler: b,
		resource:       resource,
		reconcileError: reconcileError,
		logger:         b.Logger().WithValues("Status Reconciler", resource.Name),
	}
}

func (s *ApplicationAuthStatusReconciler) Reconcile() (reconcile.Result, error) {
	s.logger.V(1).Info("START")

	newStatus, err := s.calculateStatus()
	if err != nil {
		return reconcile.Result{}, err
	}

	equalStatus := s.resource.Status.Equals(newStatus, s.logger)
	s.logger.V(1).Info("Status", "status is different", !equalStatus)
	if equalStatus {
		// Steady state
		s.logger.V(1).Info("Status steady state, status was not updated")
		return reconcile.Result{}, nil
	}

	s.resource.Status = *newStatus
	updateErr := s.Client().Status().Update(s.Context(), s.resource)
	if updateErr != nil {
		// Ignore conflicts, resource might just be outdated.
		if errors.IsConflict(updateErr) {
			s.logger.Info("Failed to update status: resource might just be outdated")
			return reconcile.Result{Requeue: true}, nil
		}

		return reconcile.Result{}, fmt.Errorf("failed to update status: %w", updateErr)
	}
	return reconcile.Result{}, nil
}

func (s *ApplicationAuthStatusReconciler) calculateStatus() (*capabilitiesv1beta1.ApplicationAuthStatus, error) {
	newStatus := &capabilitiesv1beta1.ApplicationAuthStatus{}

	newStatus.Conditions = s.resource.Status.Conditions.Copy()
	newStatus.Conditions.SetCondition(s.readyCondition())
	newStatus.Conditions.SetCondition(s.failedCondition())

	return newStatus, nil
}

func (s *ApplicationAuthStatusReconciler) readyCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.ApplicationReadyConditionType,
		Status: corev1.ConditionFalse,
	}

	if s.reconcileError == nil {
		condition.Status = corev1.ConditionTrue
		condition.Message = "Application authentication has been successfully pushed, any further interactions with this CR will not be applied"
	}

	return condition
}

func (s *ApplicationAuthStatusReconciler) failedCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.ApplicationAuthFailedConditionType,
		Status: corev1.ConditionFalse,
	}

	if s.reconcileError != nil {
		condition.Status = corev1.ConditionTrue
		if strings.Contains(s.reconcileError.Error(), "Limit reached") {
			condition.Message = "ApplicationKey limit reached a maximum of 5 keys allowed"
		} else if strings.Contains(s.reconcileError.Error(), "has already been taken") {
			condition.Message = "ApplicationKey or UserKey of this value already exists for this application update your authSecretRef with a new value"
		} else {
			condition.Message = s.reconcileError.Error()
		}
	}

	return condition
}
