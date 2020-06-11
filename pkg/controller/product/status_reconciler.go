package product

import (
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
)

type StatusReconciler struct {
	*reconcilers.BaseReconciler
	resource  *capabilitiesv1beta1.Product
	entity    *helper.ProductEntity
	syncError error
	logger    logr.Logger
}

func NewStatusReconciler(b *reconcilers.BaseReconciler, resource *capabilitiesv1beta1.Product, entity *helper.ProductEntity, syncError error) *StatusReconciler {
	return &StatusReconciler{
		BaseReconciler: b,
		resource:       resource,
		entity:         entity,
		syncError:      syncError,
		logger:         b.Logger().WithValues("Status Reconciler", resource.Name),
	}
}

func (s *StatusReconciler) Reconcile() error {
	s.logger.V(1).Info("START")

	newStatus := s.calculateStatus()

	equalStatus := s.resource.Status.Equals(newStatus, s.logger)
	s.logger.V(1).Info("Status", "status is different", !equalStatus)
	s.logger.V(1).Info("Status", "generation is different", s.resource.Generation != s.resource.Status.ObservedGeneration)
	if equalStatus && s.resource.Generation == s.resource.Status.ObservedGeneration {
		// Steady state
		s.logger.V(1).Info("Status steady state, status was not updated")
		return nil
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
		return fmt.Errorf("Failed to update status: %w", updateErr)
	}
	return nil
}

func (s *StatusReconciler) calculateStatus() *capabilitiesv1beta1.ProductStatus {
	newStatus := &capabilitiesv1beta1.ProductStatus{}
	if s.entity != nil {
		tmpID := s.entity.ID()
		newStatus.ID = &tmpID
		tmpState := s.entity.State()
		newStatus.State = &tmpState
	}

	newStatus.ObservedGeneration = s.resource.Status.ObservedGeneration

	newStatus.Conditions = s.resource.Status.Conditions.Copy()
	newStatus.Conditions.SetCondition(s.syncCondition())
	newStatus.Conditions.SetCondition(s.orphanCondition())
	newStatus.Conditions.SetCondition(s.invalidCondition())
	newStatus.Conditions.SetCondition(s.failedCondition())

	// terminal problems
	if helper.IsInvalidSpecError(s.syncError) {
		message := s.syncError.Error()
		newStatus.ErrorMessage = &message
	}

	return newStatus
}

func (s *StatusReconciler) syncCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.ProductSyncedConditionType,
		Status: corev1.ConditionFalse,
	}

	if s.syncError == nil {
		condition.Status = corev1.ConditionTrue
	}

	return condition
}

func (s *StatusReconciler) orphanCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.ProductOrphanConditionType,
		Status: corev1.ConditionFalse,
	}

	if helper.IsOrphanSpecError(s.syncError) {
		condition.Status = corev1.ConditionTrue
		condition.Message = s.syncError.Error()
	}

	return condition
}

func (s *StatusReconciler) invalidCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.ProductInvalidConditionType,
		Status: corev1.ConditionFalse,
	}

	if helper.IsInvalidSpecError(s.syncError) {
		condition.Status = corev1.ConditionTrue
		condition.Message = s.syncError.Error()
	}

	return condition
}

func (s *StatusReconciler) failedCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.ProductFailedConditionType,
		Status: corev1.ConditionFalse,
	}

	// This condition could be activated together with other conditions
	if s.syncError != nil {
		condition.Status = corev1.ConditionTrue
		condition.Message = s.syncError.Error()
	}

	return condition
}
