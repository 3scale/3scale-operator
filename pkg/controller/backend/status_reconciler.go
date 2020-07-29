package backend

import (
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/common"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

type StatusReconciler struct {
	*reconcilers.BaseReconciler
	backendResource     *capabilitiesv1beta1.Backend
	backendAPIEntity    *controllerhelper.BackendAPIEntity
	providerAccountHost string
	syncError           error
	logger              logr.Logger
}

func NewStatusReconciler(b *reconcilers.BaseReconciler, backendResource *capabilitiesv1beta1.Backend, backendAPIEntity *controllerhelper.BackendAPIEntity, providerAccountHost string, syncError error) *StatusReconciler {
	return &StatusReconciler{
		BaseReconciler:      b,
		backendResource:     backendResource,
		backendAPIEntity:    backendAPIEntity,
		providerAccountHost: providerAccountHost,
		syncError:           syncError,
		logger:              b.Logger().WithValues("Status Reconciler", backendResource.Name),
	}
}

func (s *StatusReconciler) Reconcile() (reconcile.Result, error) {
	s.logger.V(1).Info("START")

	newStatus := s.calculateStatus()

	equalStatus := s.backendResource.Status.Equals(newStatus, s.logger)
	s.logger.V(1).Info("Status", "status is different", !equalStatus)
	s.logger.V(1).Info("Status", "generation is different", s.backendResource.Generation != s.backendResource.Status.ObservedGeneration)
	if equalStatus && s.backendResource.Generation == s.backendResource.Status.ObservedGeneration {
		// Steady state
		s.logger.V(1).Info("Status was not updated")
		return reconcile.Result{}, nil
	}

	// Save the generation number we acted on, otherwise we might wrongfully indicate
	// that we've seen a spec update when we retry.
	// TODO: This can clobber an update if we allow multiple agents to write to the
	// same status.
	newStatus.ObservedGeneration = s.backendResource.Generation

	s.logger.V(1).Info("Updating Status", "sequence no:", fmt.Sprintf("sequence No: %v->%v", s.backendResource.Status.ObservedGeneration, newStatus.ObservedGeneration))

	s.backendResource.Status = *newStatus
	updateErr := s.Client().Status().Update(s.Context(), s.backendResource)
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

func (s *StatusReconciler) calculateStatus() *capabilitiesv1beta1.BackendStatus {
	newStatus := &capabilitiesv1beta1.BackendStatus{}
	if s.backendAPIEntity != nil {
		tmp := s.backendAPIEntity.ID()
		newStatus.ID = &tmp
	}

	newStatus.ProviderAccountHost = s.providerAccountHost

	newStatus.ObservedGeneration = s.backendResource.Status.ObservedGeneration

	newStatus.Conditions = s.backendResource.Status.Conditions.Copy()
	newStatus.Conditions.SetCondition(s.syncCondition())
	newStatus.Conditions.SetCondition(s.invalidCondition())
	newStatus.Conditions.SetCondition(s.failedCondition())

	return newStatus
}

func (s *StatusReconciler) syncCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.BackendSyncedConditionType,
		Status: corev1.ConditionFalse,
	}

	if s.syncError == nil {
		condition.Status = corev1.ConditionTrue
	}

	return condition
}

func (s *StatusReconciler) invalidCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.BackendInvalidConditionType,
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
		Type:   capabilitiesv1beta1.BackendFailedConditionType,
		Status: corev1.ConditionFalse,
	}

	// This condition could be activated together with other conditions
	if s.syncError != nil {
		condition.Status = corev1.ConditionTrue
		condition.Message = s.syncError.Error()
	}

	return condition
}
