package controllers

import (
	"fmt"

	capabilitiesv1beta2 "github.com/3scale/3scale-operator/apis/capabilities/v1beta2"
	"github.com/3scale/3scale-operator/pkg/common"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

type ProductStatusReconciler struct {
	*reconcilers.BaseReconciler
	resource            *capabilitiesv1beta2.Product
	entity              *controllerhelper.ProductEntity
	providerAccountHost string
	syncError           error
	logger              logr.Logger
}

func NewProductStatusReconciler(b *reconcilers.BaseReconciler, resource *capabilitiesv1beta2.Product, entity *controllerhelper.ProductEntity, providerAccountHost string, syncError error) *ProductStatusReconciler {
	return &ProductStatusReconciler{
		BaseReconciler:      b,
		resource:            resource,
		entity:              entity,
		providerAccountHost: providerAccountHost,
		syncError:           syncError,
		logger:              b.Logger().WithValues("Status Reconciler", resource.Name),
	}
}

func (s *ProductStatusReconciler) Reconcile() (reconcile.Result, error) {
	s.logger.V(1).Info("START")

	newStatus := s.calculateStatus()

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

func (s *ProductStatusReconciler) calculateStatus() *capabilitiesv1beta2.ProductStatus {
	newStatus := &capabilitiesv1beta2.ProductStatus{}
	if s.entity != nil {
		tmpID := s.entity.ID()
		newStatus.ID = &tmpID
	}

	newStatus.ProviderAccountHost = s.providerAccountHost

	newStatus.ObservedGeneration = s.resource.Status.ObservedGeneration

	newStatus.Conditions = s.resource.Status.Conditions.Copy()
	newStatus.Conditions.SetCondition(s.syncCondition())
	newStatus.Conditions.SetCondition(s.orphanCondition())
	newStatus.Conditions.SetCondition(s.invalidCondition())
	newStatus.Conditions.SetCondition(s.failedCondition())

	return newStatus
}

func (s *ProductStatusReconciler) syncCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta2.ProductSyncedConditionType,
		Status: corev1.ConditionFalse,
	}

	if s.syncError == nil {
		condition.Status = corev1.ConditionTrue
	}

	return condition
}

func (s *ProductStatusReconciler) orphanCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta2.ProductOrphanConditionType,
		Status: corev1.ConditionFalse,
	}

	if helper.IsOrphanSpecError(s.syncError) {
		condition.Status = corev1.ConditionTrue
		condition.Message = s.syncError.Error()
	}

	return condition
}

func (s *ProductStatusReconciler) invalidCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta2.ProductInvalidConditionType,
		Status: corev1.ConditionFalse,
	}

	if helper.IsInvalidSpecError(s.syncError) {
		condition.Status = corev1.ConditionTrue
		condition.Message = s.syncError.Error()
	}

	return condition
}

func (s *ProductStatusReconciler) failedCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta2.ProductFailedConditionType,
		Status: corev1.ConditionFalse,
	}

	// This condition could be activated together with other conditions
	if s.syncError != nil {
		condition.Status = corev1.ConditionTrue
		condition.Message = s.syncError.Error()
	}

	return condition
}
