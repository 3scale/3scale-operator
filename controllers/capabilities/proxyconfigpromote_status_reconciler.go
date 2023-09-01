package controllers

import (
	"fmt"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/apispkg/common"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ProxyConfigPromoteStatusReconciler struct {
	*reconcilers.BaseReconciler
	resource                *capabilitiesv1beta1.ProxyConfigPromote
	state                   string
	productID               string
	latestProductionVersion int
	latestStagingVersion    int
	reconcileError          error
	logger                  logr.Logger
}

func NewProxyConfigPromoteStatusReconciler(b *reconcilers.BaseReconciler, resource *capabilitiesv1beta1.ProxyConfigPromote, state string, productID string, latestProductionVersion int, latestStagingVersion int, reconcileError error) *ProxyConfigPromoteStatusReconciler {
	return &ProxyConfigPromoteStatusReconciler{
		BaseReconciler:          b,
		resource:                resource,
		state:                   state,
		productID:               productID,
		latestProductionVersion: latestProductionVersion,
		latestStagingVersion:    latestStagingVersion,
		reconcileError:          reconcileError,
		logger:                  b.Logger().WithValues("Status Reconciler", resource.Name),
	}
}

func (s *ProxyConfigPromoteStatusReconciler) Reconcile() (reconcile.Result, error) {
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

		return reconcile.Result{}, fmt.Errorf("Failed to update status: %w", updateErr)
	}
	return reconcile.Result{}, nil
}

func (s *ProxyConfigPromoteStatusReconciler) calculateStatus() (*capabilitiesv1beta1.ProxyConfigPromoteStatus, error) {
	newStatus := &capabilitiesv1beta1.ProxyConfigPromoteStatus{}

	newStatus.ProductId = s.productID
	newStatus.LatestProductionVersion = s.latestProductionVersion
	newStatus.LatestStagingVersion = s.latestStagingVersion

	newStatus.Conditions = s.resource.Status.Conditions.Copy()
	newStatus.Conditions.SetCondition(s.readyCondition(s.state))

	return newStatus, nil
}

func (s *ProxyConfigPromoteStatusReconciler) readyCondition(state string) common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.ProxyPromoteConfigReadyConditionType,
		Status: corev1.ConditionFalse,
	}

	if state == "Completed" {
		condition.Status = corev1.ConditionTrue
	}

	return condition
}
