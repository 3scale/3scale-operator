package operator

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/helper"
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
		"com.redhat.component-version": helper.ParseVersion(ApicastImageURL()),
		"com.redhat.product-name":      "3scale",
		"com.redhat.product-version":   product.ThreescaleRelease,
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
		"com.redhat.component-version": helper.ParseVersion(ApicastImageURL()),
		"com.redhat.product-name":      "3scale",
		"com.redhat.product-version":   product.ThreescaleRelease,
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

func testApicastStagingCustomResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("123m"),
			v1.ResourceMemory: resource.MustParse("456Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("789m"),
			v1.ResourceMemory: resource.MustParse("111Mi"),
		},
	}
}

func testApicastProductionCustomResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("222m"),
			v1.ResourceMemory: resource.MustParse("333Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("444m"),
			v1.ResourceMemory: resource.MustParse("555Mi"),
		},
	}
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
		Namespace:                      namespace,
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
		{"WithGlobalResourceRequirementsDisabled",
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
		{"WithAPIcastCustomResourceRequirements",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerTestApicastOptions()
				apimanager.Spec.Apicast.ProductionSpec.Resources = testApicastProductionCustomResourceRequirements()
				apimanager.Spec.Apicast.StagingSpec.Resources = testApicastStagingCustomResourceRequirements()
				return apimanager
			},
			func() *component.ApicastOptions {
				opts := defaultApicastOptions()
				opts.ProductionResourceRequirements = *testApicastProductionCustomResourceRequirements()
				opts.StagingResourceRequirements = *testApicastStagingCustomResourceRequirements()
				return opts
			},
		},
		{"WithAPIcastCustomResourceRequirementsAndGlobalResourceRequirementsDisabled",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerTestApicastOptions()
				apimanager.Spec.ResourceRequirementsEnabled = &falseValue
				apimanager.Spec.Apicast.ProductionSpec.Resources = testApicastProductionCustomResourceRequirements()
				apimanager.Spec.Apicast.StagingSpec.Resources = testApicastStagingCustomResourceRequirements()
				return apimanager
			},
			func() *component.ApicastOptions {
				opts := defaultApicastOptions()
				opts.ProductionResourceRequirements = *testApicastProductionCustomResourceRequirements()
				opts.StagingResourceRequirements = *testApicastStagingCustomResourceRequirements()
				return opts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(subT *testing.T) {
			objs := []runtime.Object{}
			cl := fake.NewFakeClient(objs...)
			optsProvider := NewApicastOptionsProvider(tc.apimanagerFactory(), cl)
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
