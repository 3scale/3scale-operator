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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
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
			PlanID:                  1,
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
			ID:                      3,
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
			AppName:                 "test",
			Description:             "test",
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
		accountResource     *capabilitiesv1beta1.DeveloperAccount
		productResource     *capabilitiesv1beta1.Product
		httpHandler         http.Handler
		logger              logr.Logger
	}
	tests := []struct {
		name                      string
		fields                    fields
		expectedApplicationEntity *controllerhelper.ApplicationEntity
		wantErr                   bool
	}{
		{
			name: "Application CR no change ",
			fields: fields{
				BaseReconciler:      getBaseReconciler(),
				applicationResource: getApplicationCR(),
				accountResource:     getApplicationDeveloperAccount(),
				productResource:     getApplicationProductCR(),
				httpHandler:         mock.NewApplicationMockServer(3, 3, getApplicationPlanListByProductJson(), *getApplicationJson("live")),
				// logger:              nil,
			},
			expectedApplicationEntity: getApplicationEntity(),
			wantErr:                   false,
		},
		{
			name: "Application CR setting state to suspend",
			fields: fields{
				BaseReconciler:      getBaseReconciler(),
				applicationResource: getApplicationCRSuspend(),
				accountResource:     getApplicationDeveloperAccount(),
				productResource:     getApplicationProductCR(),
				httpHandler:         mock.NewApplicationMockServer(3, 3, getApplicationPlanListByProductJson(), *getApplicationJson("suspended")),
				// logger:              nil,
			},
			expectedApplicationEntity: getApplicationEntitySuspended(),
			wantErr:                   false,
		},
		{
			name: "Application CR setting state to live",
			fields: fields{
				BaseReconciler:      getBaseReconciler(),
				applicationResource: getApplicationCR(),
				accountResource:     getApplicationDeveloperAccount(),
				productResource:     getApplicationProductCR(),
				httpHandler:         mock.NewApplicationMockServer(3, 3, getApplicationPlanListByProductJson(), *getApplicationJson("suspended")),
				// logger:              nil,
			},
			expectedApplicationEntity: getApplicationEntity(),
			wantErr:                   false,
		},
		{
			name: "Application CR change name",
			fields: fields{
				BaseReconciler: getBaseReconciler(),
				applicationResource: &capabilitiesv1beta1.Application{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: capabilitiesv1beta1.ApplicationSpec{
						AccountCR: &corev1.LocalObjectReference{
							Name: "test",
						},
						ProductCR: &corev1.LocalObjectReference{
							Name: "test",
						},
						ApplicationPlanName: "test",
						Name:                "foo",
						Description:         "test",
						Suspend:             false,
					},
					Status: capabilitiesv1beta1.ApplicationStatus{
						ID: ptr.To(int64(3)),
					},
				},
				accountResource: getApplicationDeveloperAccount(),
				productResource: getApplicationProductCR(),
				httpHandler:     mock.NewApplicationMockServer(3, 3, getApplicationPlanListByProductJson(), *getApplicationJson("live")),
			},
			expectedApplicationEntity: &controllerhelper.ApplicationEntity{
				ApplicationObj: &threescaleapi.Application{
					ID:            3,
					State:         "live",
					UserAccountID: 3,
					ServiceID:     0,
					PlanID:        1,
					AppName:       "foo",
					Description:   "test",
				},
			},
			wantErr: false,
		},
		{
			name: "Application CR change description",
			fields: fields{
				BaseReconciler: getBaseReconciler(),
				applicationResource: &capabilitiesv1beta1.Application{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: capabilitiesv1beta1.ApplicationSpec{
						AccountCR: &corev1.LocalObjectReference{
							Name: "test",
						},
						ProductCR: &corev1.LocalObjectReference{
							Name: "test",
						},
						ApplicationPlanName: "test",
						Name:                "test",
						Description:         "foo",
						Suspend:             false,
					},
					Status: capabilitiesv1beta1.ApplicationStatus{
						ID: ptr.To(int64(3)),
					},
				},
				accountResource: getApplicationDeveloperAccount(),
				productResource: getApplicationProductCR(),
				httpHandler:     mock.NewApplicationMockServer(3, 3, getApplicationPlanListByProductJson(), *getApplicationJson("live")),
				// logger:              nil,
			},
			expectedApplicationEntity: &controllerhelper.ApplicationEntity{
				ApplicationObj: &threescaleapi.Application{
					ID:            3,
					State:         "live",
					UserAccountID: 3,
					ServiceID:     0,
					PlanID:        1,
					AppName:       "test",
					Description:   "foo",
				},
			},
			wantErr: false,
		},
		{
			name: "Application CR change application Plan",
			fields: fields{
				BaseReconciler: getBaseReconciler(),
				applicationResource: &capabilitiesv1beta1.Application{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: capabilitiesv1beta1.ApplicationSpec{
						AccountCR: &corev1.LocalObjectReference{
							Name: "test",
						},
						ProductCR: &corev1.LocalObjectReference{
							Name: "test",
						},
						ApplicationPlanName: "test2",
						Name:                "test",
						Description:         "test",
						Suspend:             false,
					},
					Status: capabilitiesv1beta1.ApplicationStatus{
						ID: ptr.To(int64(3)),
					},
				},
				accountResource: getApplicationDeveloperAccount(),
				productResource: getApplicationProductCR(),
				httpHandler:     mock.NewApplicationMockServer(3, 3, getApplicationPlanListByProductJson(), *getApplicationJson("live")),
				// logger:              nil,
			},
			expectedApplicationEntity: &controllerhelper.ApplicationEntity{
				ApplicationObj: &threescaleapi.Application{
					ID:            3,
					State:         "live",
					UserAccountID: 3,
					ServiceID:     0,
					PlanID:        2,
					AppName:       "test",
					Description:   "test",
				},
			},
			wantErr: false,
		},
		{
			name: "Application CR invalid application Plan",
			fields: fields{
				BaseReconciler: getBaseReconciler(),
				applicationResource: &capabilitiesv1beta1.Application{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: capabilitiesv1beta1.ApplicationSpec{
						AccountCR: &corev1.LocalObjectReference{
							Name: "test",
						},
						ProductCR: &corev1.LocalObjectReference{
							Name: "test",
						},
						ApplicationPlanName: "unknown",
						Name:                "test",
						Description:         "test",
						Suspend:             false,
					},
					Status: capabilitiesv1beta1.ApplicationStatus{
						ID: ptr.To(int64(3)),
					},
				},
				accountResource: getApplicationDeveloperAccount(),
				productResource: getApplicationProductCR(),
				httpHandler:     mock.NewApplicationMockServer(3, 3, getApplicationPlanListByProductJson(), *getApplicationJson("live")),
				// logger:              nil,
			},
			wantErr: true,
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
				accountID:           *tt.fields.accountResource.Status.ID,
				productID:           *tt.fields.productResource.Status.ID,
				threescaleAPIClient: threescaleAPIClient,
				logger:              tt.fields.logger,
			}
			entity, err := t.Reconcile()
			if (err != nil) != tt.wantErr {
				t1.Fatalf("syncApplication() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if entity.ID() != tt.expectedApplicationEntity.ID() {
					t1.Errorf("syncApplication() expected ID = %v, got %v", tt.expectedApplicationEntity.ID(), entity.ID())
				}
				if entity.AppName() != tt.expectedApplicationEntity.AppName() {
					t1.Errorf("syncApplication() expected AppName = %v, got %v", tt.expectedApplicationEntity.AppName(), entity.AppName())
				}
				if entity.Description() != tt.expectedApplicationEntity.Description() {
					t1.Errorf("syncApplication() expected Description = %v, got %v", tt.expectedApplicationEntity.Description(), entity.Description())
				}
				if entity.ApplicationState() != tt.expectedApplicationEntity.ApplicationState() {
					t1.Errorf("syncApplication() expected ApplicationState = %v, got %v", tt.expectedApplicationEntity.ApplicationState(), entity.ApplicationState())
				}
				if entity.PlanID() != tt.expectedApplicationEntity.PlanID() {
					t1.Errorf("syncApplication() expected PlanID = %v, got %v", tt.expectedApplicationEntity.PlanID(), entity.PlanID())
				}
			}
		})
	}
}
