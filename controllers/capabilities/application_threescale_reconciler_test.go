package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/mock"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
)

func getApplicationEntity() *controllerhelper.ApplicationEntity {
	applicationEntity := &controllerhelper.ApplicationEntity{
		ApplicationObj: &threescaleapi.Application{
			ID:                      3,
			CreatedAt:               "",
			UpdatedAt:               "",
			State:                   "live",
			UserAccountID:           3,
			FirstTrafficAt:          "",
			FirstDailyTrafficAt:     "",
			EndUserRequired:         false,
			ServiceID:               0,
			UserKey:                 "",
			ProviderVerificationKey: "",
			PlanID:                  0,
			AppName:                 "test",
			Description:             "test",
			ExtraFields:             "",
			Error:                   "",
		},
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
			UserAccountID:           3,
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
			UserAccountID:           3,
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
	}
	return applicationEntity
}

func TestApplicationThreescaleReconciler_syncApplication(t1 *testing.T) {
	type fields struct {
		BaseReconciler      *reconcilers.BaseReconciler
		applicationResource *capabilitiesv1beta1.Application
		applicationEntity   *controllerhelper.ApplicationEntity
		accountResource     *capabilitiesv1beta1.DeveloperAccount
		productResource     *capabilitiesv1beta1.Product
		httpHandler         http.Handler
		logger              logr.Logger
	}
	tests := []struct {
		name    string
		fields  fields
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
				httpHandler:         mock.NewApplicationMockServer(3, 3, getApplicationPlanListByProductJson(), *getApplicationJson("live")),
				// logger:              nil,
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
				httpHandler:         mock.NewApplicationMockServer(3, 3, getApplicationPlanListByProductJson(), *getApplicationJson("live")),
				// logger:              nil,
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
				httpHandler:         mock.NewApplicationMockServer(3, 3, getApplicationPlanListByProductJson(), *getApplicationJson("live")),
				// logger:              nil,
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
				httpHandler:         mock.NewApplicationMockServer(3, 3, getApplicationPlanListByProductJson(), *getApplicationJson("live")),
				// logger:              nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			srv := httptest.NewServer(tt.fields.httpHandler)
			defer srv.Close()
			// admin portal
			ap, _ := threescaleapi.NewAdminPortalFromStr(srv.URL)
			threescaleAPIClient := threescaleapi.NewThreeScale(ap, "test", srv.Client())

			t := &ApplicationThreescaleReconciler{
				BaseReconciler:      tt.fields.BaseReconciler,
				applicationResource: tt.fields.applicationResource,
				applicationEntity:   tt.fields.applicationEntity,
				accountResource:     tt.fields.accountResource,
				productResource:     tt.fields.productResource,
				threescaleAPIClient: threescaleAPIClient,
				logger:              tt.fields.logger,
			}
			_, err := t.Reconcile()
			if (err != nil) != tt.wantErr {
				t1.Errorf("syncApplication() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
