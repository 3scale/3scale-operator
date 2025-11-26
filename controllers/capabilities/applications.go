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

	plan, err := t.findPlan(*t.productResource.Status.ID)
	if err != nil {
		return fmt.Errorf("error finding plan ID for plan : [%s]", t.applicationResource.Spec.ApplicationPlanName)
	}

	if t.applicationEntity.PlanID() != plan.Element.ID {
		_, err := t.threescaleAPIClient.ChangeApplicationPlan(*t.accountResource.Status.ID, *t.applicationResource.Status.ID, plan.Element.ID)
		if err != nil {
			return fmt.Errorf("error sync application [%s;%d]: %w", t.applicationResource.Spec.Name, t.applicationEntity.ID(), err)
		}
	}

	if t.applicationResource.Spec.Suspend && t.applicationEntity.ApplicationState() == "live" {
		_, err := t.threescaleAPIClient.ApplicationSuspend(*t.accountResource.Status.ID, t.applicationEntity.ID())
		if err != nil {
			return fmt.Errorf("error sync application [%s;%d]: %w", t.applicationResource.Spec.Name, t.applicationEntity.ID(), err)
		}
	}

	if !t.applicationResource.Spec.Suspend && t.applicationEntity.ApplicationState() == "suspended" {
		_, err := t.threescaleAPIClient.ApplicationResume(*t.accountResource.Status.ID, t.applicationEntity.ID())
		if err != nil {
			return fmt.Errorf("error sync application [%s;%d]: %w", t.applicationResource.Spec.Name, t.applicationEntity.ID(), err)
		}
	}

	if len(params) > 0 {
		_, err := t.threescaleAPIClient.UpdateApplication(*t.accountResource.Status.ID, *t.applicationResource.Status.ID, params)
		if err != nil {
			return fmt.Errorf("error sync application [%s;%d]: %w", t.applicationResource.Spec.Name, t.applicationEntity.ID(), err)
		}
	}

	return nil
}
