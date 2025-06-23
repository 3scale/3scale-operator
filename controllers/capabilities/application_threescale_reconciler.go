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
	accountResource     *capabilitiesv1beta1.DeveloperAccount
	productResource     *capabilitiesv1beta1.Product
	threescaleAPIClient *threescaleapi.ThreeScaleClient
	logger              logr.Logger
}

func NewApplicationReconciler(b *reconcilers.BaseReconciler, applicationResource *capabilitiesv1beta1.Application, accountResource *capabilitiesv1beta1.DeveloperAccount, productResource *capabilitiesv1beta1.Product, threescaleAPIClient *threescaleapi.ThreeScaleClient) *ApplicationThreescaleReconciler {
	return &ApplicationThreescaleReconciler{
		BaseReconciler:      b,
		applicationResource: applicationResource,
		accountResource:     accountResource,
		productResource:     productResource,
		threescaleAPIClient: threescaleAPIClient,
		logger:              b.Logger().WithValues("3scale Reconciler", applicationResource.Name),
	}
}

func (t *ApplicationThreescaleReconciler) Reconcile() (*controllerhelper.ApplicationEntity, error) {
	applicationEntity, err := t.reconcile3scaleApplication()
	if err != nil {
		return nil, err
	}
	t.applicationEntity = applicationEntity
	taskRunner := helper.NewTaskRunner(nil, t.logger)
	taskRunner.AddTask("SyncApplication", t.syncApplication)

	err = taskRunner.Run()
	if err != nil {
		return nil, err
	}

	return t.applicationEntity, nil
}

func (t *ApplicationThreescaleReconciler) reconcile3scaleApplication() (*controllerhelper.ApplicationEntity, error) {
	planObj, err := t.findPlan()
	if err != nil {
		return nil, fmt.Errorf("reconcile3scaleApplication application [%s]: %w", t.applicationResource.Spec.Name, err)
	}

	application, err := t.findApplication()
	if err != nil {
		return nil, err
	}

	if application == nil {
		// Application doesn't exist yet - create it
		a, err := t.threescaleAPIClient.CreateApp(strconv.FormatInt(*t.accountResource.Status.ID, 10), strconv.FormatInt(planObj.Element.ID, 10), t.applicationResource.Spec.Name, t.applicationResource.Spec.Description)
		if err != nil {
			return nil, fmt.Errorf("reconcile3scaleApplication application [%s]: %w", t.applicationResource.Spec.Name, err)
		}
		application = &a
		t.applicationResource.Status.ID = &application.ID
	}

	return controllerhelper.NewApplicationEntity(application, t.threescaleAPIClient, t.logger), nil
}

func (t *ApplicationThreescaleReconciler) findApplication() (*threescaleapi.Application, error) {
	applicationList, err := t.threescaleAPIClient.ListApplications(*t.accountResource.Status.ID)
	if err != nil {
		return nil, fmt.Errorf("reconcile3scaleApplication application [%s]: %w", t.applicationResource.Spec.Name, err)
	}

	var applicationID int64 = 0
	// If the application.Status.ID is nil, check if application.annotations.applicationID is present and use it instead
	if t.applicationResource.Status.ID == nil {
		if value, found := t.applicationResource.ObjectMeta.Annotations[applicationIdAnnotation]; found {
			// If the applicationID annotation is found, convert it to int64
			applicationIdConvertedFromString, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to convert applicationID annotation value %s to int64: %w", value, err)
			}

			applicationID = applicationIdConvertedFromString
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

func (t *ApplicationThreescaleReconciler) findPlan() (*threescaleapi.ApplicationPlan, error) {
	planList, err := t.threescaleAPIClient.ListApplicationPlansByProduct(*t.productResource.Status.ID)
	if err != nil {
		return nil, fmt.Errorf("reconcile3scaleApplications application [%s]: %w", t.applicationResource.Spec.ApplicationPlanName, err)
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
		return planObj, nil
	}
	return nil, fmt.Errorf("plan [%s] doesnt exist in product [%s] ", t.applicationResource.Spec.ApplicationPlanName, t.productResource.Spec.Name)
}
