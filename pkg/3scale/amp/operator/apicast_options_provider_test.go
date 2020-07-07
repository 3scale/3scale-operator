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

func testApicastCommonLabels() map[string]string {
	return map[string]string{
		"app":                  appLabel,
		"threescale_component": "apicast",
	}
}

func testApicastStagingLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "apicast",
		"threescale_component_element": "staging",
	}
}

func testApicastProductionLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "apicast",
		"threescale_component_element": "production",
	}
}

func testApicastStagingPodLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "apicast",
		"threescale_component_element": "staging",
		"com.redhat.component-name":    "apicast-staging",
		"com.redhat.component-type":    "application",
		"com.redhat.component-version": "nightly",
		"com.redhat.product-name":      "3scale",
		"com.redhat.product-version":   "master",
		"deploymentConfig":             "apicast-staging",
	}
}

func testApicastProductionPodLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "apicast",
		"threescale_component_element": "production",
		"com.redhat.component-name":    "apicast-production",
		"com.redhat.component-type":    "application",
		"com.redhat.component-version": "nightly",
		"com.redhat.product-name":      "3scale",
		"com.redhat.product-version":   "master",
		"deploymentConfig":             "apicast-production",
	}
}

func testApicastStagingAffinity() *v1.Affinity {
	return getTestAffinity("apicast-staging")
}

func testApicastProductionAffinity() *v1.Affinity {
	return getTestAffinity("apicast-production")
}

func testApicastStagingTolerations() []v1.Toleration {
	return getTestTolerations("apicast-staging")
}

func testApicastProductionTolerations() []v1.Toleration {
	return getTestTolerations("apicast-production")
}

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
		ManagementAPI:                  apicastManagementAPI,
		OpenSSLVerify:                  strconv.FormatBool(openSSLVerify),
		ResponseCodes:                  strconv.FormatBool(responseCodes),
		ImageTag:                       product.ThreescaleRelease,
		ExtendedMetrics:                true,
		ProductionResourceRequirements: component.DefaultProductionResourceRequirements(),
		StagingResourceRequirements:    component.DefaultStagingResourceRequirements(),
		ProductionReplicas:             int32(productionReplicaCount),
		StagingReplicas:                int32(stagingReplicaCount),
		CommonLabels:                   testApicastCommonLabels(),
		CommonStagingLabels:            testApicastStagingLabels(),
		CommonProductionLabels:         testApicastProductionLabels(),
		StagingPodTemplateLabels:       testApicastStagingPodLabels(),
		ProductionPodTemplateLabels:    testApicastProductionPodLabels(),
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
		{"WithAffinity",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerTestApicastOptions()
				apimanager.Spec.Apicast.ProductionSpec.Affinity = testApicastProductionAffinity()
				apimanager.Spec.Apicast.StagingSpec.Affinity = testApicastStagingAffinity()
				return apimanager
			},
			func() *component.ApicastOptions {
				opts := defaultApicastOptions()
				opts.ProductionAffinity = testApicastProductionAffinity()
				opts.StagingAffinity = testApicastStagingAffinity()
				return opts
			},
		},
		{"WithTolerations",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerTestApicastOptions()
				apimanager.Spec.Apicast.ProductionSpec.Tolerations = testApicastProductionTolerations()
				apimanager.Spec.Apicast.StagingSpec.Tolerations = testApicastStagingTolerations()
				return apimanager
			},
			func() *component.ApicastOptions {
				opts := defaultApicastOptions()
				opts.ProductionTolerations = testApicastProductionTolerations()
				opts.StagingTolerations = testApicastStagingTolerations()
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
