package controllers

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-logr/logr"
)

func TestOpenAPIProductReconciler_desiredMappingRules(t *testing.T) {
	trueVal := true

	type fields struct {
		BaseReconciler  *reconcilers.BaseReconciler
		openapiCR       *capabilitiesv1beta1.OpenAPI
		openapiObj      *openapi3.T
		providerAccount *controllerhelper.ProviderAccount
		logger          logr.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    []capabilitiesv1beta1.MappingRuleSpec
		wantErr bool
	}{
		{
			name: "valid OAS",
			fields: fields{
				BaseReconciler: getTestBaseReconciler(getOpenAPICR(), getValidOpenAPISecret()),
				openapiCR:      getOpenAPICR(),
				openapiObj:     getOpenAPIObj(getValidOpenAPISecret()),
				logger:         getOpenAPITestLogger(),
			},
			want: []capabilitiesv1beta1.MappingRuleSpec{
				{
					HTTPMethod:      "GET",
					Pattern:         "/v1/pets$",
					MetricMethodRef: "metric01",
					Increment:       2,
					Last:            &trueVal,
				},
			},
			wantErr: false,
		},
		{
			name: "unextended OAS",
			fields: fields{
				BaseReconciler: getTestBaseReconciler(getOpenAPICR(), getUnextendedOpenAPISecret()),
				openapiCR:      getOpenAPICR(),
				openapiObj:     getOpenAPIObj(getUnextendedOpenAPISecret()),
				logger:         getOpenAPITestLogger(),
			},
			want: []capabilitiesv1beta1.MappingRuleSpec{
				{
					HTTPMethod:      "GET",
					Pattern:         "/v1/pets$",
					MetricMethodRef: "listpets", // Defaults to OAS.paths[/pets].get.operationId when not extended
					Increment:       1,          // Defaults to 1 when not extended
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &OpenAPIProductReconciler{
				BaseReconciler:  tt.fields.BaseReconciler,
				openapiCR:       tt.fields.openapiCR,
				openapiObj:      tt.fields.openapiObj,
				providerAccount: tt.fields.providerAccount,
				logger:          tt.fields.logger,
			}
			got, err := p.desiredMappingRules()
			if (err != nil) != tt.wantErr {
				t.Errorf("desiredMappingRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("desiredMappingRules() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpenAPIProductReconciler_desiredMetrics(t *testing.T) {
	type fields struct {
		BaseReconciler  *reconcilers.BaseReconciler
		openapiCR       *capabilitiesv1beta1.OpenAPI
		openapiObj      *openapi3.T
		providerAccount *controllerhelper.ProviderAccount
		logger          logr.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]capabilitiesv1beta1.MetricSpec
		wantErr bool
	}{
		{
			name: "valid OAS",
			fields: fields{
				BaseReconciler: getTestBaseReconciler(getOpenAPICR(), getValidOpenAPISecret()),
				openapiCR:      getOpenAPICR(),
				openapiObj:     getOpenAPIObj(getValidOpenAPISecret()),
				logger:         getOpenAPITestLogger(),
			},
			want: map[string]capabilitiesv1beta1.MetricSpec{
				"metric01": {
					Name:        "My Metric 01",
					Unit:        "hits",
					Description: "This is a custom metric",
				},
			},
			wantErr: false,
		},
		{
			name: "unextended OAS",
			fields: fields{
				BaseReconciler: getTestBaseReconciler(getOpenAPICR(), getUnextendedOpenAPISecret()),
				openapiCR:      getOpenAPICR(),
				openapiObj:     getOpenAPIObj(getUnextendedOpenAPISecret()),
				logger:         getOpenAPITestLogger(),
			},
			want:    make(map[string]capabilitiesv1beta1.MetricSpec), // Expect empty map
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &OpenAPIProductReconciler{
				BaseReconciler:  tt.fields.BaseReconciler,
				openapiCR:       tt.fields.openapiCR,
				openapiObj:      tt.fields.openapiObj,
				providerAccount: tt.fields.providerAccount,
				logger:          tt.fields.logger,
			}
			got, err := p.desiredMetrics()
			if (err != nil) != tt.wantErr {
				t.Errorf("desiredMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("desiredMetrics() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpenAPIProductReconciler_desiredPolicies(t *testing.T) {
	type fields struct {
		BaseReconciler  *reconcilers.BaseReconciler
		openapiCR       *capabilitiesv1beta1.OpenAPI
		openapiObj      *openapi3.T
		providerAccount *controllerhelper.ProviderAccount
		logger          logr.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    []capabilitiesv1beta1.PolicyConfig
		wantErr bool
	}{
		{
			name: "valid OAS",
			fields: fields{
				BaseReconciler: getTestBaseReconciler(getOpenAPICR(), getValidOpenAPISecret()),
				openapiCR:      getOpenAPICR(),
				openapiObj:     getOpenAPIObj(getValidOpenAPISecret()),
				logger:         getOpenAPITestLogger(),
			},
			want: []capabilitiesv1beta1.PolicyConfig{
				{
					Name:    "myPolicy1",
					Version: "0.1",
					Enabled: true,
					Configuration: runtime.RawExtension{
						Raw: []byte("{\"http_proxy\":\"http://example.com\"}"),
					},
				},
				{
					Name:    "myPolicy2",
					Version: "2.0",
					Enabled: true,
					ConfigurationRef: corev1.SecretReference{
						Name:      "my-config-policy-secret",
						Namespace: "testNamespace",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "unextended OAS",
			fields: fields{
				BaseReconciler: getTestBaseReconciler(getOpenAPICR(), getUnextendedOpenAPISecret()),
				openapiCR:      getOpenAPICR(),
				openapiObj:     getOpenAPIObj(getUnextendedOpenAPISecret()),
				logger:         getOpenAPITestLogger(),
			},
			want:    nil, // Expect empty list
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &OpenAPIProductReconciler{
				BaseReconciler:  tt.fields.BaseReconciler,
				openapiCR:       tt.fields.openapiCR,
				openapiObj:      tt.fields.openapiObj,
				providerAccount: tt.fields.providerAccount,
				logger:          tt.fields.logger,
			}
			got, err := p.desiredPolicies()
			if (err != nil) != tt.wantErr {
				t.Errorf("desiredPolicies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("desiredPolicies() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpenAPIProductReconciler_desiredApplicationPlans(t *testing.T) {
	planName := "My Plan 01"
	planAppsRequireApproval := false
	planTrialPeriod := 1
	planSetupPrice := "1.00"
	planBackendSystemName := "Swagger_Petstore"

	type fields struct {
		BaseReconciler  *reconcilers.BaseReconciler
		openapiCR       *capabilitiesv1beta1.OpenAPI
		openapiObj      *openapi3.T
		providerAccount *controllerhelper.ProviderAccount
		logger          logr.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]capabilitiesv1beta1.ApplicationPlanSpec
		wantErr bool
	}{
		{
			name: "valid OAS",
			fields: fields{
				BaseReconciler: getTestBaseReconciler(getOpenAPICR(), getValidOpenAPISecret()),
				openapiCR:      getOpenAPICR(),
				openapiObj:     getOpenAPIObj(getValidOpenAPISecret()),
				logger:         getOpenAPITestLogger(),
			},
			want: map[string]capabilitiesv1beta1.ApplicationPlanSpec{
				"plan01": {
					Name:                &planName,
					AppsRequireApproval: &planAppsRequireApproval,
					TrialPeriod:         &planTrialPeriod,
					SetupFee:            &planSetupPrice,
					CostMonth:           &planSetupPrice,
					PricingRules: []capabilitiesv1beta1.PricingRuleSpec{
						{
							From: 1,
							To:   100,
							MetricMethodRef: capabilitiesv1beta1.MetricMethodRefSpec{
								SystemName: "metric01",
							},
							PricePerUnit: planSetupPrice,
						},
					},
					Limits: []capabilitiesv1beta1.LimitSpec{
						{
							Period: "week",
							Value:  100,
							MetricMethodRef: capabilitiesv1beta1.MetricMethodRefSpec{
								SystemName:        "hits",
								BackendSystemName: &planBackendSystemName,
							},
						},
					},
					Published: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "unextended OAS",
			fields: fields{
				BaseReconciler: getTestBaseReconciler(getOpenAPICR(), getUnextendedOpenAPISecret()),
				openapiCR:      getOpenAPICR(),
				openapiObj:     getOpenAPIObj(getUnextendedOpenAPISecret()),
				logger:         getOpenAPITestLogger(),
			},
			want:    make(map[string]capabilitiesv1beta1.ApplicationPlanSpec), // Expect empty map
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &OpenAPIProductReconciler{
				BaseReconciler:  tt.fields.BaseReconciler,
				openapiCR:       tt.fields.openapiCR,
				openapiObj:      tt.fields.openapiObj,
				providerAccount: tt.fields.providerAccount,
				logger:          tt.fields.logger,
			}
			got, err := p.desiredApplicationPlans()
			if (err != nil) != tt.wantErr {
				t.Errorf("desiredApplicationPlans() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("desiredApplicationPlans() got = %v, want %v", got, tt.want)
			}
		})
	}
}
