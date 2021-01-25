package controllers

import (
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type DeveloperUserStatusReconciler struct {
	*reconcilers.BaseReconciler
	userCR              *capabilitiesv1beta1.DeveloperUser
	parentAccountCR     *capabilitiesv1beta1.DeveloperAccount
	providerAccountHost string
	remoteDeveloperUser *threescaleapi.DeveloperUser
	reconcileError      error
	logger              logr.Logger
}

func NewDeveloperUserStatusReconciler(b *reconcilers.BaseReconciler,
	userCR *capabilitiesv1beta1.DeveloperUser,
	parentAccountCR *capabilitiesv1beta1.DeveloperAccount,
	providerAccountHost string,
	remoteDeveloperUser *threescaleapi.DeveloperUser,
	reconcileError error,
) *DeveloperUserStatusReconciler {
	return &DeveloperUserStatusReconciler{
		BaseReconciler:      b,
		userCR:              userCR,
		parentAccountCR:     parentAccountCR,
		providerAccountHost: providerAccountHost,
		remoteDeveloperUser: remoteDeveloperUser,
		reconcileError:      reconcileError,
		logger:              b.Logger().WithValues("Status Reconciler", userCR.Name),
	}
}

func (s *DeveloperUserStatusReconciler) Reconcile() (reconcile.Result, error) {
	s.logger.V(1).Info("START")

	newStatus, err := s.calculateStatus()
	if err != nil {
		return reconcile.Result{}, err
	}

	equalStatus := s.userCR.Status.Equals(newStatus, s.logger)
	s.logger.V(1).Info("Status", "status is different", !equalStatus)
	s.logger.V(1).Info("Status", "generation is different", s.userCR.Generation != s.userCR.Status.ObservedGeneration)
	if equalStatus && s.userCR.Generation == s.userCR.Status.ObservedGeneration {
		// Steady state
		s.logger.V(1).Info("Status steady state, status was not updated")
		return reconcile.Result{}, nil
	}

	// Save the generation number we acted on, otherwise we might wrongfully indicate
	// that we've seen a spec update when we retry.
	// TODO: This can clobber an update if we allow multiple agents to write to the
	// same status.
	newStatus.ObservedGeneration = s.userCR.Generation

	s.logger.V(1).Info("Updating Status", "sequence no:", fmt.Sprintf("sequence No: %v->%v", s.userCR.Status.ObservedGeneration, newStatus.ObservedGeneration))

	s.userCR.Status = *newStatus
	updateErr := s.Client().Status().Update(s.Context(), s.userCR)
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

func (s *DeveloperUserStatusReconciler) calculateStatus() (*capabilitiesv1beta1.DeveloperUserStatus, error) {
	// If there is an error and s.remoteDeveloperUser is nil, do not change status fields read from it
	// Initialize with existing data for data coming from 3scale
	// just in case in this reconciliation loop something goes wrong and avoid replacing right data with nil
	newStatus := &capabilitiesv1beta1.DeveloperUserStatus{
		ID:                  s.userCR.Status.ID,
		DeveloperUserState:  s.userCR.Status.DeveloperUserState,
		ProviderAccountHost: s.userCR.Status.ProviderAccountHost,
		AccountID:           s.userCR.Status.AccountID,
		Conditions:          s.userCR.Status.Conditions.Copy(),
		ObservedGeneration:  s.userCR.Status.ObservedGeneration,
	}

	if s.remoteDeveloperUser != nil {
		newStatus.ID = s.remoteDeveloperUser.Element.ID
		newStatus.DeveloperUserState = s.remoteDeveloperUser.Element.State
	}

	if s.parentAccountCR != nil {
		newStatus.AccountID = s.parentAccountCR.Status.ID
	}

	if s.providerAccountHost != "" {
		newStatus.ProviderAccountHost = s.providerAccountHost
	}

	newStatus.Conditions.SetCondition(s.invalidCondition())
	newStatus.Conditions.SetCondition(s.readyCondition())
	newStatus.Conditions.SetCondition(s.orphanCondition())
	newStatus.Conditions.SetCondition(s.failedCondition())

	return newStatus, nil
}

func (s *DeveloperUserStatusReconciler) readyCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.DeveloperUserReadyConditionType,
		Status: corev1.ConditionFalse,
	}

	if s.reconcileError == nil {
		condition.Status = corev1.ConditionTrue
	}

	return condition
}

func (s *DeveloperUserStatusReconciler) invalidCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.DeveloperUserInvalidConditionType,
		Status: corev1.ConditionFalse,
	}

	if helper.IsInvalidSpecError(s.reconcileError) {
		condition.Status = corev1.ConditionTrue
		condition.Message = s.reconcileError.Error()
	}

	return condition
}

func (s *DeveloperUserStatusReconciler) failedCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.DeveloperUserFailedConditionType,
		Status: corev1.ConditionFalse,
	}

	if s.reconcileError != nil {
		// only activate this condition when others are false and still there is an error

		otherConditionsFalse := []bool{
			s.invalidCondition().IsFalse(),
			s.orphanCondition().IsFalse(),
		}

		if helper.All(otherConditionsFalse) {
			condition.Status = corev1.ConditionTrue
			condition.Message = s.reconcileError.Error()
		}
	}

	return condition
}

func (s *DeveloperUserStatusReconciler) orphanCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.DeveloperUserOrphanConditionType,
		Status: corev1.ConditionFalse,
	}

	if helper.IsOrphanSpecError(s.reconcileError) {
		condition.Status = corev1.ConditionTrue
		condition.Message = s.reconcileError.Error()
	}

	return condition
}
