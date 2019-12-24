package operator

import (
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
)

func TestGetApicastOptions(t *testing.T) {
	wildcardDomain := "test.3scale.net"
	appLabel := "someLabel"
	tenantName := "someTenant"
	apicastManagementAPI := "disabled"
	trueValue := true
	var oneValue int64 = 1

	cases := []struct {
		name                        string
		resourceRequirementsEnabled bool
	}{
		{"WithResourceRequirements", true},
		{"WithoutResourceRequirements", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(subT *testing.T) {
			resourceRequirementsEnabled := tc.resourceRequirementsEnabled
			apimanager := &appsv1alpha1.APIManagerSpec{
				APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
					WildcardDomain:               wildcardDomain,
					AppLabel:                     &appLabel,
					ImageStreamTagImportInsecure: &trueValue,
					TenantName:                   &tenantName,
					ResourceRequirementsEnabled:  &resourceRequirementsEnabled,
				},
				Apicast: &appsv1alpha1.ApicastSpec{
					ApicastManagementAPI: &apicastManagementAPI,
					OpenSSLVerify:        &trueValue,
					IncludeResponseCodes: &trueValue,
					StagingSpec: &appsv1alpha1.ApicastStagingSpec{
						Replicas: &oneValue,
					},
					ProductionSpec: &appsv1alpha1.ApicastProductionSpec{
						Replicas: &oneValue,
					},
				},
			}
			optsProvider := NewApicastOptionsProvider(apimanager)
			_, err := optsProvider.GetApicastOptions()
			if err != nil {
				subT.Error(err)
			}
			// created "opts" cannot be tested  here, it only has set methods
			// and cannot assert on setted values from a different package
			// TODO: refactor options provider structure
			// then validate setted resources
		})
	}

}
