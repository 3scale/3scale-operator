package openapi

import (
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OpenAPIStatusReconciler struct {
	*reconcilers.BaseReconciler
	resource            *capabilitiesv1beta1.OpenAPI
	providerAccountHost string
	reconcileError      error
	reconcileReady      bool
	logger              logr.Logger
}

func NewOpenAPIStatusReconciler(b *reconcilers.BaseReconciler, resource *capabilitiesv1beta1.OpenAPI, providerAccountHost string, reconcileError error, reconcileReady bool) *OpenAPIStatusReconciler {
	return &OpenAPIStatusReconciler{
		BaseReconciler:      b,
		resource:            resource,
		providerAccountHost: providerAccountHost,
		reconcileError:      reconcileError,
		reconcileReady:      reconcileReady,
		logger:              b.Logger().WithValues("Status Reconciler", resource.Name),
	}
}

func (s *OpenAPIStatusReconciler) Reconcile() (reconcile.Result, error) {
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

func (s *OpenAPIStatusReconciler) calculateStatus() (*capabilitiesv1beta1.OpenAPIStatus, error) {
	newStatus := &capabilitiesv1beta1.OpenAPIStatus{}

	newStatus.ProviderAccountHost = s.providerAccountHost

	productResourceName, err := s.getManagedProduct()
	if err != nil {
		return nil, err
	}
	newStatus.ProductResourceName = productResourceName

	backendResourceNames, err := s.getManagedBackends()
	if err != nil {
		return nil, err
	}
	newStatus.BackendResourceNames = backendResourceNames

	newStatus.ObservedGeneration = s.resource.Status.ObservedGeneration

	newStatus.Conditions = s.resource.Status.Conditions.Copy()
	newStatus.Conditions.SetCondition(s.readyCondition())
	newStatus.Conditions.SetCondition(s.invalidCondition())
	newStatus.Conditions.SetCondition(s.failedCondition())

	return newStatus, nil
}

func (s *OpenAPIStatusReconciler) readyCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.OpenAPIReadyConditionType,
		Status: corev1.ConditionFalse,
	}

	if s.reconcileReady {
		condition.Status = corev1.ConditionTrue
	}

	return condition
}

func (s *OpenAPIStatusReconciler) invalidCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.OpenAPIInvalidConditionType,
		Status: corev1.ConditionFalse,
	}

	if helper.IsInvalidSpecError(s.reconcileError) {
		condition.Status = corev1.ConditionTrue
		condition.Message = s.reconcileError.Error()
	}

	return condition
}

func (s *OpenAPIStatusReconciler) failedCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.OpenAPIFailedConditionType,
		Status: corev1.ConditionFalse,
	}

	// This condition could be activated together with other conditions
	if s.reconcileError != nil {
		condition.Status = corev1.ConditionTrue
		condition.Message = s.reconcileError.Error()
	}

	return condition
}

func (s *OpenAPIStatusReconciler) getManagedProduct() (*corev1.LocalObjectReference, error) {
	listOps := []client.ListOption{
		client.InNamespace(s.resource.Namespace),
	}
	productList := &capabilitiesv1beta1.ProductList{}
	err := s.Client().List(s.Context(), productList, listOps...)
	if err != nil {
		return nil, fmt.Errorf("Failed to list product: %w", err)
	}

	for _, product := range productList.Items {
		for _, ownerRef := range product.GetOwnerReferences() {
			if ownerRef.UID == s.resource.UID {
				return &corev1.LocalObjectReference{
					Name: product.Name,
				}, nil
			}
		}
	}

	return nil, nil
}

func (s *OpenAPIStatusReconciler) getManagedBackends() ([]corev1.LocalObjectReference, error) {
	listOps := []client.ListOption{
		client.InNamespace(s.resource.Namespace),
	}
	list := &capabilitiesv1beta1.BackendList{}
	err := s.Client().List(s.Context(), list, listOps...)
	if err != nil {
		return nil, fmt.Errorf("Failed to list backends: %w", err)
	}

	var managedBackends []corev1.LocalObjectReference
	for _, backend := range list.Items {
		for _, ownerRef := range backend.GetOwnerReferences() {
			if ownerRef.UID == s.resource.UID {
				managedBackends = append(managedBackends, corev1.LocalObjectReference{
					Name: backend.Name,
				})
			}
		}
	}

	return managedBackends, nil
}
