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

type CustomPolicyDefinitionStatusReconciler struct {
	*reconcilers.BaseReconciler
	resource            *capabilitiesv1beta1.CustomPolicyDefinition
	providerAccountHost string
	customPolicy        *threescaleapi.APIcastPolicy
	reconcileError      error
	logger              logr.Logger
}

func NewCustomPolicyDefinitionStatusReconciler(b *reconcilers.BaseReconciler, resource *capabilitiesv1beta1.CustomPolicyDefinition, providerAccountHost string, customPolicy *threescaleapi.APIcastPolicy, reconcileError error) *CustomPolicyDefinitionStatusReconciler {
	return &CustomPolicyDefinitionStatusReconciler{
		BaseReconciler:      b,
		resource:            resource,
		providerAccountHost: providerAccountHost,
		customPolicy:        customPolicy,
		reconcileError:      reconcileError,
		logger:              b.Logger().WithValues("Status Reconciler", resource.Name),
	}
}

func (s *CustomPolicyDefinitionStatusReconciler) Reconcile() (reconcile.Result, error) {
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

func (s *CustomPolicyDefinitionStatusReconciler) calculateStatus() (*capabilitiesv1beta1.CustomPolicyDefinitionStatus, error) {
	newStatus := &capabilitiesv1beta1.CustomPolicyDefinitionStatus{}

	if s.customPolicy != nil {
		newStatus.ID = s.customPolicy.Element.ID
	}

	newStatus.ProviderAccountHost = s.providerAccountHost

	newStatus.ObservedGeneration = s.resource.Status.ObservedGeneration

	newStatus.Conditions = s.resource.Status.Conditions.Copy()
	newStatus.Conditions.SetCondition(s.readyCondition())
	newStatus.Conditions.SetCondition(s.invalidCondition())
	newStatus.Conditions.SetCondition(s.failedCondition())

	return newStatus, nil
}

func (s *CustomPolicyDefinitionStatusReconciler) readyCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.CustomPolicyDefinitionReadyConditionType,
		Status: corev1.ConditionFalse,
	}

	if s.reconcileError == nil {
		condition.Status = corev1.ConditionTrue
	}

	return condition
}

func (s *CustomPolicyDefinitionStatusReconciler) invalidCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.CustomPolicyDefinitionInvalidConditionType,
		Status: corev1.ConditionFalse,
	}

	if helper.IsInvalidSpecError(s.reconcileError) {
		condition.Status = corev1.ConditionTrue
		condition.Message = s.reconcileError.Error()
	}

	return condition
}

func (s *CustomPolicyDefinitionStatusReconciler) failedCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.CustomPolicyDefinitionFailedConditionType,
		Status: corev1.ConditionFalse,
	}

	// This condition could be activated together with other conditions
	if s.reconcileError != nil {
		condition.Status = corev1.ConditionTrue
		condition.Message = s.reconcileError.Error()
	}

	return condition
}
