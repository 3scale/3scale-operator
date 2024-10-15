package v1alpha1

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/3scale/3scale-operator/version"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetDefaults(t *testing.T) {
	tmpDefaultAppLabel := Default3scaleAppLabel
	tmpDefaultTenantName := defaultTenantName
	tmpDefaultResourceRequirementsEnabled := defaultResourceRequirementsEnabled
	tmpDefaultApicastManagementAPI := defaultApicastManagementAPI
	tmpDefaultApicastOpenSSLVerify := defaultApicastOpenSSLVerify
	tmpDefaultApicastResponseCodes := defaultApicastResponseCodes
	tmpDefaultApicastRegistryURL := defaultApicastRegistryURL
	tmpDefaultZyncEnabled := true

	inputAPIManager := minimumAPIManagerTest()

	expectedAPIManager := APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				OperatorVersionAnnotation:   version.Version,
				ThreescaleVersionAnnotation: version.ThreescaleVersionMajorMinorPatch(),
			},
		},
		Spec: APIManagerSpec{
			APIManagerCommonSpec: APIManagerCommonSpec{
				WildcardDomain:              "test.3scale.com",
				AppLabel:                    &tmpDefaultAppLabel,
				TenantName:                  &tmpDefaultTenantName,
				ResourceRequirementsEnabled: &tmpDefaultResourceRequirementsEnabled,
			},
			Apicast: &ApicastSpec{
				IncludeResponseCodes: &tmpDefaultApicastResponseCodes,
				ApicastManagementAPI: &tmpDefaultApicastManagementAPI,
				OpenSSLVerify:        &tmpDefaultApicastOpenSSLVerify,
				RegistryURL:          &tmpDefaultApicastRegistryURL,
				ProductionSpec:       &ApicastProductionSpec{},
				StagingSpec:          &ApicastStagingSpec{},
			},
			Backend: &BackendSpec{
				ListenerSpec: &BackendListenerSpec{},
				WorkerSpec:   &BackendWorkerSpec{},
				CronSpec:     &BackendCronSpec{},
			},
			System: &SystemSpec{
				AppSpec:     &SystemAppSpec{},
				SidekiqSpec: &SystemSidekiqSpec{},
				SearchdSpec: &SystemSearchdSpec{},
			},
			Zync: &ZyncSpec{
				AppSpec: &ZyncAppSpec{},
				QueSpec: &ZyncQueSpec{},
				Enabled: &tmpDefaultZyncEnabled,
			},
			PodDisruptionBudget: nil,
		},
	}

	tmpInput := inputAPIManager.DeepCopy()
	t.Run("BasicDefaults", testBasicAPIManagerDefaults(tmpInput, &expectedAPIManager))
}

func testBasicAPIManagerDefaults(input, expected *APIManager) func(t *testing.T) {
	return func(t *testing.T) {
		changed, err := input.SetDefaults()
		if !changed {
			t.Errorf("Expected introduced APIManager defaults changed")
		}
		if err != nil {
			t.Errorf("Expected not an error being returned")
		}
		if !reflect.DeepEqual(input, expected) {
			t.Errorf("Resulting APIManager differs from the expected one. Differences are: (%s)", cmp.Diff(input, expected))
		}
	}
}

func TestZyncExternalDatabaseIsEnabled(t *testing.T) {
	trueVal := true
	falseVal := false

	cases := []struct {
		testName          string
		apimanagerFactory func() *APIManager
		expectedResult    bool
	}{
		{"WithDefaultAPIManager",
			func() *APIManager {
				return minimumAPIManagerTest()
			},
			false,
		},
		{"WithHighAvailabilityEnabledOnly",
			func() *APIManager {
				apimanager := minimumAPIManagerTest()
				apimanager.Spec.HighAvailability = &HighAvailabilitySpec{
					Enabled: true,
				}
				return apimanager
			},
			false,
		},
		{"WithBothHighAvailabilityAndExternalZyncDBEnabled",
			func() *APIManager {
				apimanager := minimumAPIManagerTest()
				apimanager.Spec.HighAvailability = &HighAvailabilitySpec{
					Enabled:                     true,
					ExternalZyncDatabaseEnabled: &trueVal,
				}
				apimanager.Spec.ExternalComponents = &ExternalComponentsSpec{
					Zync: &ExternalZyncComponents{Database: &trueVal},
				}
				return apimanager
			},
			true,
		},
		{"WithHADisabledAndExternalZyncDBEnabled",
			func() *APIManager {
				apimanager := minimumAPIManagerTest()
				apimanager.Spec.HighAvailability = &HighAvailabilitySpec{
					Enabled:                     false,
					ExternalZyncDatabaseEnabled: &trueVal,
				}
				return apimanager
			},
			false,
		},
		{"WithHAEnabledAndExternalZyncDBDisabled",
			func() *APIManager {
				apimanager := minimumAPIManagerTest()
				apimanager.Spec.HighAvailability = &HighAvailabilitySpec{
					Enabled:                     false,
					ExternalZyncDatabaseEnabled: &falseVal,
				}
				return apimanager
			},
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			receivedResult := tc.apimanagerFactory().IsExternal(ZyncDatabase)
			if !reflect.DeepEqual(tc.expectedResult, receivedResult) {
				subT.Errorf("Expected result differs: Expected: %t, Received: %t", tc.expectedResult, receivedResult)
			}
		})
	}
}

func minimumAPIManagerTest() *APIManager {
	return &APIManager{
		Spec: APIManagerSpec{
			APIManagerCommonSpec: APIManagerCommonSpec{
				WildcardDomain: "test.3scale.com",
			},
		},
	}
}

func TestRemoveDuplicateSecretRefs(t *testing.T) {
	type args struct {
		refs []*v1.LocalObjectReference
	}
	tests := []struct {
		name string
		args args
		want []*v1.LocalObjectReference
	}{
		{
			name: "SecretRefs is nil",
			args: args{
				refs: nil,
			},
			want: []*v1.LocalObjectReference{},
		},
		{
			name: "SecretRefs is empty",
			args: args{
				refs: []*v1.LocalObjectReference{},
			},
			want: []*v1.LocalObjectReference{},
		},
		{
			name: "SecretRefs has duplicates",
			args: args{
				refs: []*v1.LocalObjectReference{
					{
						Name: "ref1",
					},
					{
						Name: "ref1",
					},
					{
						Name: "ref2",
					},
				},
			},
			want: []*v1.LocalObjectReference{
				{
					Name: "ref1",
				},
				{
					Name: "ref2",
				},
			},
		},
		{
			name: "SecretRefs does not have duplicates",
			args: args{
				refs: []*v1.LocalObjectReference{
					{
						Name: "ref1",
					},
					{
						Name: "ref2",
					},
				},
			},
			want: []*v1.LocalObjectReference{
				{
					Name: "ref1",
				},
				{
					Name: "ref2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeDuplicateSecretRefs(tt.args.refs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveDuplicateSecretRefs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIManager_Get3scaleSecretRefs(t *testing.T) {
	type fields struct {
		Spec APIManagerSpec
	}
	tests := []struct {
		name   string
		fields fields
		want   []*v1.LocalObjectReference
	}{
		{
			name: "No secret refs to gather",
			fields: fields{
				Spec: APIManagerSpec{
					Apicast: &ApicastSpec{
						ProductionSpec: &ApicastProductionSpec{},
						StagingSpec:    &ApicastStagingSpec{},
					},
				},
			},
			want: []*v1.LocalObjectReference{
				{Name: "system-redis"},
				{Name: "backend-redis"},
				{
					Name: "system-database",
				},
				{
					Name: "zync",
				},
			},
		},
		{
			name: "Apicast has secret refs",
			fields: fields{
				Spec: APIManagerSpec{
					Apicast: &ApicastSpec{
						ProductionSpec: &ApicastProductionSpec{
							HTTPSCertificateSecretRef: &v1.LocalObjectReference{
								Name: "https-cert-secret",
							},
							OpenTelemetry: &OpenTelemetrySpec{
								TracingConfigSecretRef: &v1.LocalObjectReference{
									Name: "otel-secret",
								},
							},
							CustomEnvironments: []CustomEnvironmentSpec{
								{
									SecretRef: &v1.LocalObjectReference{
										Name: "custom-env-1-secret",
									},
								},
							},
						},
						StagingSpec: &ApicastStagingSpec{
							CustomEnvironments: []CustomEnvironmentSpec{
								{
									SecretRef: &v1.LocalObjectReference{
										Name: "custom-env-1-secret",
									},
								},
							},
							CustomPolicies: []CustomPolicySpec{
								{
									SecretRef: &v1.LocalObjectReference{
										Name: "custom-policy-1-secret",
									},
								},
							},
						},
					},
				},
			},
			want: []*v1.LocalObjectReference{
				{
					Name: "system-redis",
				},
				{
					Name: "backend-redis",
				},
				{
					Name: "https-cert-secret",
				},
				{
					Name: "otel-secret",
				},
				{
					Name: "custom-env-1-secret",
				},
				{
					Name: "custom-policy-1-secret",
				},
				{
					Name: "system-database",
				},
				{
					Name: "zync",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apimanager := &APIManager{
				Spec: tt.fields.Spec,
			}
			if got := apimanager.Get3scaleSecretRefs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get3scaleSecretRefs() = %v, want %v", got, tt.want)
			}
		})
	}
}
