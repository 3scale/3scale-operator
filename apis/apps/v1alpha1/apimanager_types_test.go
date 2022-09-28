package v1alpha1

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetDefaults(t *testing.T) {
	tmpDefaultAppLabel := Default3scaleAppLabel
	tmpDefaultTenantName := defaultTenantName
	tmpDefaultImageStreamTagImportInsecure := defaultImageStreamImportInsecure
	tmpDefaultResourceRequirementsEnabled := defaultResourceRequirementsEnabled
	tmpDefaultApicastManagementAPI := defaultApicastManagementAPI
	tmpDefaultApicastOpenSSLVerify := defaultApicastOpenSSLVerify
	tmpDefaultApicastResponseCodes := defaultApicastResponseCodes
	tmpDefaultApicastRegistryURL := defaultApicastRegistryURL

	var tmpDefaultReplicas int64 = 1

	inputAPIManager := minimumAPIManagerTest()

	expectedAPIManager := APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				OperatorVersionAnnotation:   version.Version,
				ThreescaleVersionAnnotation: product.ThreescaleRelease,
			},
		},
		Spec: APIManagerSpec{
			APIManagerCommonSpec: APIManagerCommonSpec{
				WildcardDomain:               "test.3scale.com",
				AppLabel:                     &tmpDefaultAppLabel,
				TenantName:                   &tmpDefaultTenantName,
				ImageStreamTagImportInsecure: &tmpDefaultImageStreamTagImportInsecure,
				ResourceRequirementsEnabled:  &tmpDefaultResourceRequirementsEnabled,
			},
			Apicast: &ApicastSpec{
				IncludeResponseCodes: &tmpDefaultApicastResponseCodes,
				ApicastManagementAPI: &tmpDefaultApicastManagementAPI,
				OpenSSLVerify:        &tmpDefaultApicastOpenSSLVerify,
				RegistryURL:          &tmpDefaultApicastRegistryURL,
				ProductionSpec: &ApicastProductionSpec{
					Replicas: &tmpDefaultReplicas,
				},
				StagingSpec: &ApicastStagingSpec{
					Replicas: &tmpDefaultReplicas,
				},
			},
			Backend: &BackendSpec{
				ListenerSpec: &BackendListenerSpec{},
				WorkerSpec:   &BackendWorkerSpec{},
				CronSpec:     &BackendCronSpec{},
			},
			System: &SystemSpec{
				AppSpec: &SystemAppSpec{
					Replicas: &tmpDefaultReplicas,
				},
				SidekiqSpec: &SystemSidekiqSpec{
					Replicas: &tmpDefaultReplicas,
				},
				SphinxSpec: &SystemSphinxSpec{},
			},
			Zync: &ZyncSpec{
				AppSpec: &ZyncAppSpec{
					Replicas: &tmpDefaultReplicas,
				},
				QueSpec: &ZyncQueSpec{
					Replicas: &tmpDefaultReplicas,
				},
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
