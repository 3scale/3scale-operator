package controllers

import (
	"fmt"
	"strconv"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"

	"github.com/go-logr/logr"
)

type ApplicationThreescaleReconciler struct {
	*reconcilers.BaseReconciler
	applicationResource *capabilitiesv1beta1.Application
	applicationEntity   *controllerhelper.ApplicationEntity
	authParams          map[string]string
	accountID           int64
	productID           int64
	threescaleAPIClient *threescaleapi.ThreeScaleClient
	logger              logr.Logger
}

func NewApplicationReconciler(b *reconcilers.BaseReconciler, applicationResource *capabilitiesv1beta1.Application, authParams map[string]string, accountID int64, productID int64, threescaleAPIClient *threescaleapi.ThreeScaleClient) *ApplicationThreescaleReconciler {
	return &ApplicationThreescaleReconciler{
		BaseReconciler:      b,
		applicationResource: applicationResource,
		authParams:          authParams,
		accountID:           accountID,
		productID:           productID,
		threescaleAPIClient: threescaleAPIClient,
		logger:              b.Logger().WithValues("3scale Reconciler", applicationResource.Name),
	}
}

func (t *ApplicationThreescaleReconciler) Reconcile() (*controllerhelper.ApplicationEntity, error) {
	taskRunner := helper.NewTaskRunner(nil, t.logger)
	taskRunner.AddTask("SyncApplication", t.syncApplication)

	err := taskRunner.Run()
	if err != nil {
		return nil, err
	}

	return t.applicationEntity, nil
}

func (t *ApplicationThreescaleReconciler) findApplication(accountID int64) (*threescaleapi.Application, error) {
	applicationList, err := t.threescaleAPIClient.ListApplications(accountID)
	if err != nil {
		return nil, fmt.Errorf("reconcile3scaleApplication application [%s]: %w", t.applicationResource.Spec.Name, err)
	}

	var applicationID int64 = 0
	// If the application.Status.ID is nil, check if application.annotations.applicationID is present and use it instead
	if t.applicationResource.Status.ID == nil {
		if value, found := t.applicationResource.ObjectMeta.Annotations[applicationIdAnnotation]; found {
			// If the applicationID annotation is found, convert it to int64
			applicationIDConvertedFromString, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to convert applicationID annotation value %s to int64: %w", value, err)
			}

			applicationID = applicationIDConvertedFromString
		}
	} else {
		applicationID = *t.applicationResource.Status.ID
	}

	if applicationID == 0 {
		return nil, nil
	}

	for _, applicationIn3scale := range applicationList.Applications {
		if applicationIn3scale.Application.ID == applicationID {
			return &applicationIn3scale.Application, nil
		}
	}

	// Unable to find application
	return nil, nil
}

func (t *ApplicationThreescaleReconciler) findPlan(productID int64) (*threescaleapi.ApplicationPlan, error) {
	planList, err := t.threescaleAPIClient.ListApplicationPlansByProduct(productID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve application plans for product [%s]: %w", t.applicationResource.Spec.ProductCR.Name, err)
	}

	for _, plan := range planList.Plans {
		if plan.Element.SystemName == t.applicationResource.Spec.ApplicationPlanName {
			return &plan, nil
		}
	}

	return nil, fmt.Errorf("plan [%s] doesnt exist in product [%s]", t.applicationResource.Spec.ApplicationPlanName, t.applicationResource.Spec.ProductCR.Name)
}

func (t *ApplicationThreescaleReconciler) syncApplication(_ any) error {
	plan, err := t.findPlan(t.productID)
	if err != nil {
		return fmt.Errorf("error sync application [%s]: %w", t.applicationResource.Spec.Name, err)
	}

	application, err := t.findApplication(t.accountID)
	if err != nil {
		return err
	}

	if application == nil {
		params := threescaleapi.Params{
			"name":        t.applicationResource.Spec.Name,
			"description": t.applicationResource.Spec.Description,
		}

		if t.authParams != nil {
			for key, value := range t.authParams {
				params.AddParam(key, value)
			}
		}
		// Application doesn't exist yet - create it
		a, err := t.threescaleAPIClient.CreateApplication(t.accountID, plan.Element.ID, t.applicationResource.Spec.Name, params)
		if err != nil {
			return fmt.Errorf("error sync application [%s]: %w", t.applicationResource.Spec.Name, err)
		}
		application = &a
		t.applicationResource.Status.ID = &application.ID
	} else if application.ServiceID != t.productID {
		// Product reference has changed
		err := t.threescaleAPIClient.DeleteApplication(t.accountID, application.ID)
		if err != nil {
			return fmt.Errorf("failed to delete application [%s] id:[%d] productID:[%d] - err: %w", t.applicationResource.Spec.Name, application.ID, application.ServiceID, err)
		}

		// Recreate the application
		params := threescaleapi.Params{
			"name":        t.applicationResource.Spec.Name,
			"description": t.applicationResource.Spec.Description,
		}

		if t.authParams != nil {
			for key, value := range t.authParams {
				params.AddParam(key, value)
			}
		}

		// Application doesn't exist yet - create it
		a, err := t.threescaleAPIClient.CreateApplication(t.accountID, plan.Element.ID, t.applicationResource.Spec.Name, params)
		if err != nil {
			return fmt.Errorf("reconcile3scaleApplication application [%s]: %w", t.applicationResource.Spec.Name, err)
		}

		application = &a
		t.applicationResource.Status.ID = &application.ID
	}

	applicationEntity := controllerhelper.NewApplicationEntity(application, t.threescaleAPIClient, t.logger)
	t.applicationEntity = applicationEntity

	params := threescaleapi.Params{}

	if t.applicationEntity.AppName() != t.applicationResource.Spec.Name {
		params["name"] = t.applicationResource.Spec.Name
	}

	if t.applicationEntity.Description() != t.applicationResource.Spec.Description {
		params["description"] = t.applicationResource.Spec.Description
	}

	if t.applicationEntity.PlanID() != plan.Element.ID {
		_, err := t.threescaleAPIClient.ChangeApplicationPlan(t.accountID, applicationEntity.ID(), plan.Element.ID)
		if err != nil {
			return fmt.Errorf("error sync application [%s;%d]: %w", t.applicationResource.Spec.Name, applicationEntity.ID(), err)
		}
	}

	if t.applicationResource.Spec.Suspend && t.applicationEntity.ApplicationState() == "live" {
		_, err := t.threescaleAPIClient.ApplicationSuspend(t.accountID, applicationEntity.ID())
		if err != nil {
			return fmt.Errorf("error sync application [%s;%d]: %w", t.applicationResource.Spec.Name, applicationEntity.ID(), err)
		}
	}

	if !t.applicationResource.Spec.Suspend && t.applicationEntity.ApplicationState() == "suspended" {
		_, err := t.threescaleAPIClient.ApplicationResume(t.accountID, applicationEntity.ID())
		if err != nil {
			return fmt.Errorf("error sync application [%s;%d]: %w", t.applicationResource.Spec.Name, applicationEntity.ID(), err)
		}
	}

	if len(params) > 0 {
		_, err := t.threescaleAPIClient.UpdateApplication(t.accountID, applicationEntity.ID(), params)
		if err != nil {
			return fmt.Errorf("error sync application [%s;%d]: %w", t.applicationResource.Spec.Name, t.applicationEntity.ID(), err)
		}
	}

	return nil
}
