package controllers

import (
	"fmt"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
)

func (t *ApplicationThreescaleReconciler) syncApplication(_ interface{}) error {
	params := threescaleapi.Params{}

	if t.applicationEntity.AppName() != t.applicationResource.Spec.Name {
		params["name"] = t.applicationResource.Spec.Name
	}

	if t.applicationEntity.Description() != t.applicationResource.Spec.Description {
		params["description"] = t.applicationResource.Spec.Description
	}

	planID, err := t.findPlanId(*t.productResource.Status.ID)
	if err != nil {
		return fmt.Errorf("error finding plan ID for plan : [%s]", t.applicationResource.Spec.ApplicationPlanName)
	}
	if t.applicationEntity.PlanID() != planID {
		_, err := t.threescaleAPIClient.ChangeApplicationPlan(*t.accountResource.Status.ID, *t.applicationResource.Status.ID, planID)
		if err != nil {
			return fmt.Errorf("error sync applicaiton [%s;%d]: %w", t.applicationResource.Spec.Name, t.applicationEntity.ID(), err)
		}
		params["applicationPlanName"] = t.applicationResource.Spec.ApplicationPlanName
	}

	if t.applicationResource.Spec.Suspend == true && t.applicationEntity.ApplicationState() == "live" {
		_, err := t.threescaleAPIClient.ApplicationSuspend(*t.accountResource.Status.ID, t.applicationEntity.ID())
		if err != nil {
			return fmt.Errorf("error sync applicaiton [%s;%d]: %w", t.applicationResource.Spec.Name, t.applicationEntity.ID(), err)
		}
		params["state"] = "suspended"
	}

	if t.applicationResource.Spec.Suspend == false && t.applicationEntity.ApplicationState() == "suspended" {
		_, err := t.threescaleAPIClient.ApplicationResume(*t.accountResource.Status.ID, t.applicationEntity.ID())
		if err != nil {
			return fmt.Errorf("error sync applicaiton [%s;%d]: %w", t.applicationResource.Spec.Name, t.applicationEntity.ID(), err)
		}
		params["state"] = "live"
	}

	if len(params) > 0 {
		_, err := t.threescaleAPIClient.UpdateApplication(*t.accountResource.Status.ID, *t.applicationResource.Status.ID, params)
		if err != nil {
			return fmt.Errorf("error sync applicaiton [%s;%d]: %w", t.applicationResource.Spec.Name, t.applicationEntity.ID(), err)
		}
	}

	return nil
}

func (t *ApplicationThreescaleReconciler) findPlanId(productID int64) (int64, error) {
	planList, err := t.threescaleAPIClient.ListApplicationPlansByProduct(productID)
	if err != nil {
		return -1, fmt.Errorf("reconcile3scaleApplication application [%s]: %w", t.applicationResource.Spec.ApplicationPlanName, err)
	}

	planID, planExists := func(pList []threescaleapi.ApplicationPlan) (int, bool) {
		for i, item := range pList {
			if item.Element.SystemName == t.applicationResource.Spec.ApplicationPlanName {
				return i, true
			}
		}
		return -1, false
	}(planList.Plans)
	var planObj *threescaleapi.ApplicationPlan
	if planExists {
		planObj = &planList.Plans[planID]
		return planObj.Element.ID, nil
	}
	return -1, fmt.Errorf("plan no longer exists")
}
