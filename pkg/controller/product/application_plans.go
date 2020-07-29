package product

import (
	"fmt"

	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
)

func (t *ThreescaleReconciler) syncApplicationPlans(_ interface{}) error {
	desiredKeys := make([]string, 0, len(t.resource.Spec.ApplicationPlans))
	for systemName := range t.resource.Spec.ApplicationPlans {
		desiredKeys = append(desiredKeys, systemName)
	}

	existingList, err := t.productEntity.ApplicationPlans()
	if err != nil {
		return fmt.Errorf("Error sync product [%s] plans: %w", t.resource.Spec.SystemName, err)
	}

	existingKeys := make([]string, 0, len(existingList.Plans))
	existingMap := map[string]threescaleapi.ApplicationPlanItem{}
	for _, existing := range existingList.Plans {
		systemName := existing.Element.SystemName
		existingKeys = append(existingKeys, systemName)
		existingMap[systemName] = existing.Element
	}

	//
	// Deleted existing and not desired
	//

	notDesiredExistingKeys := helper.ArrayStringDifference(existingKeys, desiredKeys)
	t.logger.V(1).Info("syncApplicationPlans", "notDesiredExistingKeys", notDesiredExistingKeys)
	for _, systemName := range notDesiredExistingKeys {
		// key is expected to exist
		// notDesiredExistingKeys is a subset of the existingMap key set
		err := t.productEntity.DeleteApplicationPlan(existingMap[systemName].ID)
		if err != nil {
			return fmt.Errorf("Error sync product [%s] plans: %w", t.resource.Spec.SystemName, err)
		}
	}

	//
	// Reconcile existing
	//
	matchedKeys := helper.ArrayStringIntersection(existingKeys, desiredKeys)
	t.logger.V(1).Info("syncApplicationPlans", "matchedKeys", matchedKeys)
	for _, systemName := range matchedKeys {
		// interface to remote entity
		planEntity := controllerhelper.NewApplicationPlanEntity(t.productEntity.ID(), existingMap[systemName], t.threescaleAPIClient, t.logger)
		// desired spec
		planSpec := t.resource.Spec.ApplicationPlans[systemName]
		reconciler := newApplicationPlanReconciler(t.BaseReconciler, systemName, planSpec, t.threescaleAPIClient, t.productEntity, t.backendRemoteIndex, planEntity, t.logger)
		err := reconciler.Reconcile()
		if err != nil {
			return fmt.Errorf("Error sync product [%s] plan [%s]: %w", t.resource.Spec.SystemName, systemName, err)
		}
	}

	//
	// Create not existing and desired
	//

	desiredNewKeys := helper.ArrayStringDifference(desiredKeys, existingKeys)
	t.logger.V(1).Info("syncApplicationPlans", "desiredNewKeys", desiredNewKeys)
	for _, systemName := range desiredNewKeys {
		// key is expected to exist
		// desiredNewKeys is a subset of the Spec.ApplicationPlans map key set
		planSpec := t.resource.Spec.ApplicationPlans[systemName]

		// Create Application Plan using system_name.
		// it cannot be modified later
		params := threescaleapi.Params{"system_name": systemName, "name": systemName}
		obj, err := t.productEntity.CreateApplicationPlan(params)
		if err != nil {
			return fmt.Errorf("Error sync product [%s] plan [%s]: %w", t.resource.Spec.SystemName, systemName, err)
		}
		// interface to remote entity
		planEntity := controllerhelper.NewApplicationPlanEntity(t.productEntity.ID(), obj.Element, t.threescaleAPIClient, t.logger)

		reconciler := newApplicationPlanReconciler(t.BaseReconciler, systemName, planSpec, t.threescaleAPIClient, t.productEntity, t.backendRemoteIndex, planEntity, t.logger)
		err = reconciler.Reconcile()
		if err != nil {
			return fmt.Errorf("Error sync product [%s] plan [%s]: %w", t.resource.Spec.SystemName, systemName, err)
		}
	}

	return nil
}
