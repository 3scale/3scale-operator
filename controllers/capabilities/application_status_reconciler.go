package controllers

import (
	"fmt"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/apispkg/common"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ApplicationStatusReconciler struct {
	*reconcilers.BaseReconciler
	applicationResource *capabilitiesv1beta1.Application
	entity              *controllerhelper.ApplicationEntity
	providerAccountHost string
	syncError           error
	logger              logr.Logger
}

func NewApplicationStatusReconciler(b *reconcilers.BaseReconciler, applicationResource *capabilitiesv1beta1.Application, entity *controllerhelper.ApplicationEntity, providerAccountHost string, syncError error) *ApplicationStatusReconciler {
	return &ApplicationStatusReconciler{
		BaseReconciler:      b,
		applicationResource: applicationResource,
		entity:              entity,
		providerAccountHost: providerAccountHost,
		syncError:           syncError,
		logger:              b.Logger().WithValues("Status Reconciler", applicationResource.Name),
	}
}

func (s *ApplicationStatusReconciler) Reconcile() (reconcile.Result, error) {
	s.logger.V(1).Info("START")

	newStatus := s.calculateStatus()

	// Need to extract the application CR's applicationID annotation value to compare with the new .status
	annotationId, _ := s.applicationResource.Annotations[applicationIdAnnotation]

	equalStatus := s.applicationResource.Status.Equals(annotationId, newStatus, s.logger)
	s.logger.V(1).Info("Status", "status is different", !equalStatus)
	s.logger.V(1).Info("Status", "generation is different", s.applicationResource.Generation != s.applicationResource.Status.ObservedGeneration)
	if equalStatus && s.applicationResource.Generation == s.applicationResource.Status.ObservedGeneration {
		// Steady state
		s.logger.V(1).Info("Status was not updated")
		return reconcile.Result{}, nil
	}

	// Save the generation number we acted on, otherwise we might wrongfully indicate
	// that we've seen a spec update when we retry.
	// TODO: This can clobber an update if we allow multiple agents to write to the
	// same status.
	newStatus.ObservedGeneration = s.applicationResource.Generation

	s.logger.V(1).Info("Updating Status", "sequence no:", fmt.Sprintf("sequence No: %v->%v", s.applicationResource.Status.ObservedGeneration, newStatus.ObservedGeneration))

	s.applicationResource.Status = *newStatus
	updateErr := s.Client().Status().Update(s.Context(), s.applicationResource)
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

func (s *ApplicationStatusReconciler) calculateStatus() *capabilitiesv1beta1.ApplicationStatus {
	newStatus := &capabilitiesv1beta1.ApplicationStatus{}

	if s.entity != nil {
		tmpID := s.entity.ID()
		newStatus.ID = &tmpID
	}

	if s.entity != nil && s.entity.ApplicationState() != "" {
		newStatus.State = s.entity.ApplicationState()
	}

	newStatus.ProviderAccountHost = s.providerAccountHost

	newStatus.ObservedGeneration = s.applicationResource.Status.ObservedGeneration

	newStatus.Conditions = s.applicationResource.Status.Conditions.Copy()
	newStatus.Conditions.SetCondition(s.ReadyCondition())

	return newStatus
}

func (s *ApplicationStatusReconciler) ReadyCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.ApplicationReadyConditionType,
		Status: corev1.ConditionFalse,
	}

	if s.syncError == nil {
		condition.Status = corev1.ConditionTrue
	}

	if s.syncError != nil {
		condition.Status = corev1.ConditionFalse
		condition.Message = s.syncError.Error()
	}

	return condition
}
