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

	inputAPIManager := APIManager{
		Spec: APIManagerSpec{
			APIManagerCommonSpec: APIManagerCommonSpec{
				WildcardDomain: "test.3scale.com",
			},
		},
	}
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
				ListenerSpec: &BackendListenerSpec{
					Replicas: &tmpDefaultReplicas,
				},
				WorkerSpec: &BackendWorkerSpec{
					Replicas: &tmpDefaultReplicas,
				},
				CronSpec: &BackendCronSpec{
					Replicas: &tmpDefaultReplicas,
				},
			},
			System: &SystemSpec{
				AppSpec: &SystemAppSpec{
					Replicas: &tmpDefaultReplicas,
				},
				SidekiqSpec: &SystemSidekiqSpec{
					Replicas: &tmpDefaultReplicas,
				},
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
