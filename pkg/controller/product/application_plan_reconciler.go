package product

import (
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
)

type applicationPlanReconciler struct {
	*reconcilers.BaseReconciler
	systemName          string
	resource            capabilitiesv1beta1.ApplicationPlanSpec
	productEntity       *helper.ProductEntity
	planEntity          *helper.ApplicationPlanEntity
	threescaleAPIClient *threescaleapi.ThreeScaleClient
	logger              logr.Logger
}

func newApplicationPlanReconciler(b *reconcilers.BaseReconciler,
	systemName string,
	resource capabilitiesv1beta1.ApplicationPlanSpec,
	threescaleAPIClient *threescaleapi.ThreeScaleClient,
	productEntity *helper.ProductEntity,
	planEntity *helper.ApplicationPlanEntity,
	logger logr.Logger,
) *applicationPlanReconciler {
	return &applicationPlanReconciler{
		BaseReconciler:      b,
		systemName:          systemName,
		resource:            resource,
		threescaleAPIClient: threescaleAPIClient,
		productEntity:       productEntity,
		planEntity:          planEntity,
		logger:              logger.WithValues("Plan", systemName),
	}
}

func (a *applicationPlanReconciler) Reconcile() error {
	// TODO

	return nil
}
