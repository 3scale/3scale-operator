package operator

import (
	"reflect"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func testMemcachedDeploymentLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "memcache",
	}
}

func testPodTemplateLabels() map[string]string {
	labels := map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "memcache",
		"deploymentConfig":             "system-memcache",
	}
	addExpectedMeteringLabels(labels, "system-memcache", helper.ApplicationType)

	return labels
}

func testMemcachedAffinity() *v1.Affinity {
	return getTestAffinity("memcached")
}

func testMemcachedTolerations() []v1.Toleration {
	return getTestTolerations("memcached")
}

func testSystemMemcachedCustomResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("111m"),
			v1.ResourceMemory: resource.MustParse("222Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("333m"),
			v1.ResourceMemory: resource.MustParse("444Mi"),
		},
	}
}

func defaultMemcachedOptions() *component.MemcachedOptions {
	return &component.MemcachedOptions{
		ImageTag:             product.ThreescaleRelease,
		ResourceRequirements: component.DefaultMemcachedResourceRequirements(),
		DeploymentLabels:     testMemcachedDeploymentLabels(),
		PodTemplateLabels:    testPodTemplateLabels(),
	}
}

func TestMemcachedOptionsProvider(t *testing.T) {
	falseValue := false

	cases := []struct {
		testName               string
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func() *component.MemcachedOptions
	}{
		{"Default", basicApimanager, defaultMemcachedOptions},
		{"WithoutResourceRequirements",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &falseValue
				return apimanager
			},
			func() *component.MemcachedOptions {
				opts := defaultMemcachedOptions()
				opts.ResourceRequirements = v1.ResourceRequirements{}
				return opts
			},
		},
		{"WithAffinity",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System.MemcachedAffinity = testMemcachedAffinity()
				return apimanager
			},
			func() *component.MemcachedOptions {
				opts := defaultMemcachedOptions()
				opts.Affinity = testMemcachedAffinity()
				return opts
			},
		},
		{"WithTolerations",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System.MemcachedTolerations = testMemcachedTolerations()
				return apimanager
			},
			func() *component.MemcachedOptions {
				opts := defaultMemcachedOptions()
				opts.Tolerations = testMemcachedTolerations()
				return opts
			},
		},
		{"WithSystemMemcachedCustomResourceRequirements",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System.MemcachedResources = testSystemMemcachedCustomResourceRequirements()
				return apimanager
			},
			func() *component.MemcachedOptions {
				opts := defaultMemcachedOptions()
				opts.ResourceRequirements = *testSystemMemcachedCustomResourceRequirements()
				return opts
			},
		},
		{"WithSystemMemcachedCustomResourceRequirementsAndGlobalResourceRequirementsDisabled",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &falseValue
				apimanager.Spec.System.MemcachedResources = testSystemMemcachedCustomResourceRequirements()
				return apimanager
			},
			func() *component.MemcachedOptions {
				opts := defaultMemcachedOptions()
				opts.ResourceRequirements = *testSystemMemcachedCustomResourceRequirements()
				return opts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			optsProvider := NewMemcachedOptionsProvider(tc.apimanagerFactory())
			opts, err := optsProvider.GetMemcachedOptions()
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
