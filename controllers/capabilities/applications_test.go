package controllers

import (
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	"testing"
)

func getApplicationEntity() *controllerhelper.ApplicationEntity {
	applicationEntity := &controllerhelper.ApplicationEntity{
		ApplicationObj: &threescaleapi.Application{
			ID:                      0,
			CreatedAt:               "",
			UpdatedAt:               "",
			State:                   "",
			UserAccountID:           "",
			FirstTrafficAt:          "",
			FirstDailyTrafficAt:     "",
			EndUserRequired:         false,
			ServiceID:               0,
			UserKey:                 "",
			ProviderVerificationKey: "",
			PlanID:                  0,
			AppName:                 "",
			Description:             "",
			ExtraFields:             "",
			Error:                   "",
		},
		ApplicationList:         nil,
		ApplicationPlanJSONList: nil,
	}
	return applicationEntity
}
func getApplicationEntitySuspended() *controllerhelper.ApplicationEntity {
	applicationEntity := &controllerhelper.ApplicationEntity{
		ApplicationObj: &threescaleapi.Application{
			ID:                      0,
			CreatedAt:               "",
			UpdatedAt:               "",
			State:                   "suspended",
			UserAccountID:           "",
			FirstTrafficAt:          "",
			FirstDailyTrafficAt:     "",
			EndUserRequired:         false,
			ServiceID:               0,
			UserKey:                 "",
			ProviderVerificationKey: "",
			PlanID:                  0,
			AppName:                 "",
			Description:             "",
			ExtraFields:             "",
			Error:                   "",
		},
		ApplicationList:         nil,
		ApplicationPlanJSONList: nil,
	}
	return applicationEntity
}

func getApplicationEntityPlanID() *controllerhelper.ApplicationEntity {
	applicationEntity := &controllerhelper.ApplicationEntity{
		ApplicationObj: &threescaleapi.Application{
			ID:                      0,
			CreatedAt:               "",
			UpdatedAt:               "",
			State:                   "suspended",
			UserAccountID:           "",
			FirstTrafficAt:          "",
			FirstDailyTrafficAt:     "",
			EndUserRequired:         false,
			ServiceID:               0,
			UserKey:                 "",
			ProviderVerificationKey: "",
			PlanID:                  1,
			AppName:                 "",
			Description:             "",
			ExtraFields:             "",
			Error:                   "",
		},
		ApplicationList:         nil,
		ApplicationPlanJSONList: nil,
	}
	return applicationEntity
}

func TestApplicationThreescaleReconciler_syncApplication(t1 *testing.T) {
	//admin portal
	ap, _ := threescaleapi.NewAdminPortalFromStr("https://3scale-admin.test.3scale.net")
	type fields struct {
		BaseReconciler      *reconcilers.BaseReconciler
		applicationResource *capabilitiesv1beta1.Application
		applicationEntity   *controllerhelper.ApplicationEntity
		accountResource     *capabilitiesv1beta1.DeveloperAccount
		productResource     *capabilitiesv1beta1.Product
		threescaleAPIClient *threescaleapi.ThreeScaleClient
		logger              logr.Logger
	}
	type args struct {
		in0 interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Application CR no change ",
			fields: fields{
				BaseReconciler:      getBaseReconciler(),
				applicationResource: getApplicationCR(),
				applicationEntity:   getApplicationEntity(),
				accountResource:     getApplicationDeveloperAccount(),
				productResource:     getApplicationProductCR(),
				threescaleAPIClient: threescaleapi.NewThreeScale(ap, "test", mockHttpClientApplication(getApplicationPlanListByProductJson(), getApplicationJson())),
				//logger:              nil,
			},
			wantErr: false,
		},
		{
			name: "Application CR setting state to suspend",
			fields: fields{
				BaseReconciler:      getBaseReconciler(),
				applicationResource: getApplicationCRSuspend(),
				applicationEntity:   getApplicationEntity(),
				accountResource:     getApplicationDeveloperAccount(),
				productResource:     getApplicationProductCR(),
				threescaleAPIClient: threescaleapi.NewThreeScale(ap, "test", mockHttpClientApplication(getApplicationPlanListByProductJson(), getApplicationJson())),
				//logger:              nil,
			},
			wantErr: false,
		},
		{
			name: "Application CR setting state to live",
			fields: fields{
				BaseReconciler:      getBaseReconciler(),
				applicationResource: getApplicationCR(),
				applicationEntity:   getApplicationEntitySuspended(),
				accountResource:     getApplicationDeveloperAccount(),
				productResource:     getApplicationProductCR(),
				threescaleAPIClient: threescaleapi.NewThreeScale(ap, "test", mockHttpClientApplication(getApplicationPlanListByProductJson(), getApplicationJson())),
				//logger:              nil,
			},
			wantErr: false,
		},
		{
			name: "Application CR change application Plan",
			fields: fields{
				BaseReconciler:      getBaseReconciler(),
				applicationResource: getApplicationCR(),
				applicationEntity:   getApplicationEntityPlanID(),
				accountResource:     getApplicationDeveloperAccount(),
				productResource:     getApplicationProductCR(),
				threescaleAPIClient: threescaleapi.NewThreeScale(ap, "test", mockHttpClientApplication(getApplicationPlanListByProductJson(), getApplicationJson())),
				//logger:              nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ApplicationThreescaleReconciler{
				BaseReconciler:      tt.fields.BaseReconciler,
				applicationResource: tt.fields.applicationResource,
				applicationEntity:   tt.fields.applicationEntity,
				accountResource:     tt.fields.accountResource,
				productResource:     tt.fields.productResource,
				threescaleAPIClient: tt.fields.threescaleAPIClient,
				logger:              tt.fields.logger,
			}
			if err := t.syncApplication(tt.args.in0); (err != nil) != tt.wantErr {
				t1.Errorf("syncApplication() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
