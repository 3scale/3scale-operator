package controllers

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"testing"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-porta-go-client/client"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func mockHttpClientApplication(listAapplicationPlanByProductJson *client.ApplicationPlanJSONList, applicationJson *client.Application) *http.Client {
	// override httpClient
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		// ListApplicationPlanByProduct(productID)
		if req.Method == "GET" && req.URL.Path == "/admin/api/services/3/application_plans.json" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBuffer(responseBody(listAapplicationPlanByProductJson))),
			}
		}
		// Get Application
		if req.Method == "GET" && req.URL.Path == "/admin/api/accounts/3/applications.json" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBuffer(responseBody(listAapplicationPlanByProductJson))),
			}
		}
		// create application
		if req.Method == "POST" && req.URL.Path == "/admin/api/accounts/3/applications.json" {
			mockResponse := struct {
				Application *client.Application `json:"application"`
			}{
				Application: applicationJson,
			}
			return &http.Response{
				StatusCode: http.StatusCreated,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBuffer(responseBody(mockResponse))),
			}
		}
		// ApplicationResume
		if req.Method == "PUT" && req.URL.Path == "/admin/api/accounts/3/applications/0/resume.json" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBuffer(responseBody(applicationJson))),
			}
		}
		// ApplicationSuspend
		if req.Method == "PUT" && req.URL.Path == "/admin/api/accounts/3/applications/3/suspend.json" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBuffer(responseBody(applicationJson))),
			}
		}
		if req.Method == "PUT" && req.URL.Path == "/admin/api/accounts/3/applications/3.json" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBuffer(responseBody(applicationJson))),
			}
		}
		// delete application
		if req.Method == "DELETE" && req.URL.Path == "/admin/api/accounts/3/applications/3.json" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBuffer(responseBody(applicationJson))),
			}
		}
		// ChangeApplicationPlan(*t.accountResource.Status.ID, *t.applicationResource.Status.ID, planID)
		if req.Method == "PUT" && req.URL.Path == "/admin/api/accounts/3/applications/3/change_plan.json" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBuffer(responseBody(applicationJson))),
			}
		}
		return nil
	})
	return httpClient
}

func getApplicationPlanListByProductJson() *client.ApplicationPlanJSONList {
	applicationPlanListByProductJson := &client.ApplicationPlanJSONList{
		Plans: []client.ApplicationPlan{
			{
				Element: client.ApplicationPlanItem{
					ID:         0,
					Name:       "test",
					SystemName: "test",
				},
			},
			{
				Element: client.ApplicationPlanItem{
					ID:         0,
					Name:       "test",
					SystemName: "test",
				},
			},
		},
	}
	return applicationPlanListByProductJson
}

func getApplicationJson(state string) *client.Application {
	applicationJson := &client.Application{
		ID:                      3,
		CreatedAt:               "",
		UpdatedAt:               "",
		State:                   state,
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
	}
	return applicationJson
}

func TestApplicationReconciler_applicationReconciler(t *testing.T) {
	// admin portal
	ap, _ := client.NewAdminPortalFromStr("https://3scale-admin.test.3scale.net")
	type fields struct {
		BaseReconciler *reconcilers.BaseReconciler
	}
	type args struct {
		applicationResource     *capabilitiesv1beta1.Application
		req                     controllerruntime.Request
		threescaleApiClient     *client.ThreeScaleClient
		providerAccountAdminURL string
		accountResource         *capabilitiesv1beta1.DeveloperAccount
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ApplicationStatusReconciler
		wantErr bool
	}{
		{
			name: "Create application successful",
			fields: fields{
				BaseReconciler: getBaseReconciler(getApplicationCR(), getProductList()),
			},
			args: args{
				applicationResource: getApplicationCR(),
				req: controllerruntime.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test",
						Namespace: "test",
					},
				},
				threescaleApiClient:     client.NewThreeScale(ap, "test", mockHttpClientApplication(getApplicationPlanListByProductJson(), getApplicationJson("live"))),
				providerAccountAdminURL: "https://3scale-admin.test.3scale.net",
				accountResource:         getApplicationDeveloperAccount(),
			},
			want: NewApplicationStatusReconciler(
				getBaseReconciler(getApplicationCR()),
				getApplicationCR(),
				getApplicationEntity(),
				"https://3scale-admin.test.3scale.net",
				nil),
			wantErr: false,
		},
		{
			name: "Attempt to create application with unknown Product and Account CR",
			fields: fields{
				BaseReconciler: getBaseReconciler(),
			},
			args: args{
				applicationResource: getFailedApplicationCR(),
				req: controllerruntime.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test",
						Namespace: "test",
					},
				},
				threescaleApiClient:     client.NewThreeScale(ap, "test", mockHttpClientApplication(getApplicationPlanListByProductJson(), getApplicationJson("live"))),
				providerAccountAdminURL: "https://3scale-admin.test.3scale.net",
				accountResource:         getApplicationDeveloperAccount(),
			},
			wantErr: true,
		},
		{
			name: "Attempt to create application with unknown Account CR name",
			fields: fields{
				BaseReconciler: getBaseReconciler(),
			},
			args: args{
				applicationResource: unknowAccountApplicationCR(),
				req: controllerruntime.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test",
						Namespace: "test",
					},
				},
				threescaleApiClient:     client.NewThreeScale(ap, "test", mockHttpClientApplication(getApplicationPlanListByProductJson(), getApplicationJson("live"))),
				providerAccountAdminURL: "https://3scale-admin.test.3scale.net",
				accountResource:         getApplicationDeveloperAccount(),
			},
			wantErr: true,
		},
		{
			name: "Attempt to create application with invalid applicacationPlanName",
			fields: fields{
				BaseReconciler: getBaseReconciler(),
			},
			args: args{
				applicationResource: getUnknownPlanApplicationCR(),
				req: controllerruntime.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test",
						Namespace: "test",
					},
				},
				threescaleApiClient:     client.NewThreeScale(ap, "test", mockHttpClientApplication(getApplicationPlanListByProductJson(), getApplicationJson("live"))),
				providerAccountAdminURL: "https://3scale-admin.test.3scale.net",
				accountResource:         getApplicationDeveloperAccount(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ApplicationReconciler{
				BaseReconciler: tt.fields.BaseReconciler,
			}
			got, err := r.applicationReconciler(tt.args.applicationResource, tt.args.req, tt.args.threescaleApiClient, tt.args.providerAccountAdminURL, tt.args.accountResource)
			if (err != nil) != tt.wantErr {
				t.Errorf("applicationReconciler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				if !reflect.DeepEqual(got.applicationResource, tt.want.applicationResource) {
					t.Errorf("applicationReconciler() got = %v, want %v", got, tt.want)
				}
				if !reflect.DeepEqual(got.entity.ApplicationObj, tt.want.entity.ApplicationObj) {
					t.Errorf("applicationReconciler() got = %v, want %v", got.entity.ApplicationObj, tt.want.entity.ApplicationObj)
				}
			}
		})
	}
}

func TestApplicationReconciler_removeApplicationFrom3scale(t *testing.T) {
	// admin portal
	ap, _ := client.NewAdminPortalFromStr("https://3scale-admin.test.3scale.net")
	type fields struct {
		BaseReconciler *reconcilers.BaseReconciler
	}
	type args struct {
		application         *capabilitiesv1beta1.Application
		req                 controllerruntime.Request
		threescaleAPIClient *client.ThreeScaleClient
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Delete Application successfully",
			fields: fields{
				BaseReconciler: getBaseReconciler(getApplicationCR(), getProviderAccount(), getApiManger(), getApplicationProductList(), getApplicationDeveloperAccount(), getProviderAccountRefSecret()),
			},
			args: args{
				application: getApplicationDeleteCR(),
				req: controllerruntime.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test",
						Namespace: "test",
					},
				},
				threescaleAPIClient: client.NewThreeScale(ap, "test", mockHttpClientApplication(getApplicationPlanListByProductJson(), getApplicationJson("live"))),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ApplicationReconciler{
				BaseReconciler: tt.fields.BaseReconciler,
			}
			if err := r.removeApplicationFrom3scale(tt.args.application, tt.args.req, *tt.args.threescaleAPIClient); (err != nil) != tt.wantErr {
				t.Errorf("removeApplicationFrom3scale() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
