package operator

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	productionReplicaCount int64 = 3
	stagingReplicaCount    int64 = 4
	apicastManagementAPI         = "disabled"
	openSSLVerify                = false
	responseCodes                = true
)

func basicApimanagerTestApicastOptions() *appsv1alpha1.APIManager {
	tmpApicastManagementAPI := apicastManagementAPI
	tmpOpenSSLVerify := openSSLVerify
	tmpResponseCodes := responseCodes
	tmpProductionReplicaCount := productionReplicaCount
	tmpStagingReplicaCount := stagingReplicaCount

	apimanager := basicApimanager()
	apimanager.Spec.Apicast = &appsv1alpha1.ApicastSpec{
		ApicastManagementAPI: &tmpApicastManagementAPI,
		OpenSSLVerify:        &tmpOpenSSLVerify,
		IncludeResponseCodes: &tmpResponseCodes,
		StagingSpec: &appsv1alpha1.ApicastStagingSpec{
			Replicas: &tmpStagingReplicaCount,
		},
		ProductionSpec: &appsv1alpha1.ApicastProductionSpec{
			Replicas: &tmpProductionReplicaCount,
		},
	}
	return apimanager
}

func defaultApicastOptions() *component.ApicastOptions {
	return &component.ApicastOptions{
		AppLabel:                       appLabel,
		ManagementAPI:                  apicastManagementAPI,
		OpenSSLVerify:                  strconv.FormatBool(openSSLVerify),
		ResponseCodes:                  strconv.FormatBool(responseCodes),
		TenantName:                     tenantName,
		WildcardDomain:                 wildcardDomain,
		ImageTag:                       product.ThreescaleRelease,
		ProductionResourceRequirements: component.DefaultProductionResourceRequirements(),
		StagingResourceRequirements:    component.DefaultStagingResourceRequirements(),
		ProductionReplicas:             int32(productionReplicaCount),
		StagingReplicas:                int32(stagingReplicaCount),
	}
}

func TestGetApicastOptionsProvider(t *testing.T) {
	falseValue := false

	cases := []struct {
		name                   string
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func() *component.ApicastOptions
	}{
		{"Default", basicApimanagerTestApicastOptions, defaultApicastOptions},
		{"WithoutResourceRequirements",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerTestApicastOptions()
				apimanager.Spec.ResourceRequirementsEnabled = &falseValue
				return apimanager
			},
			func() *component.ApicastOptions {
				opts := defaultApicastOptions()
				opts.ProductionResourceRequirements = v1.ResourceRequirements{}
				opts.StagingResourceRequirements = v1.ResourceRequirements{}
				return opts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(subT *testing.T) {
			optsProvider := NewApicastOptionsProvider(tc.apimanagerFactory())
			opts, err := optsProvider.GetApicastOptions()
			if err != nil {
				subT.Error(err)
			}
			expectedOptions := tc.expectedOptionsFactory()
			if !reflect.DeepEqual(expectedOptions, opts) {
				subT.Errorf("Resulting expected options differ: %s", cmp.Diff(expectedOptions, opts, cmpopts.IgnoreUnexported(resource.Quantity{})))
			}
		})
	}
}
