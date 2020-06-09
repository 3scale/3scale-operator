package operator

import (
	"reflect"
	"testing"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func testRedisBackendCommonLabels() map[string]string {
	return map[string]string{
		"app":                  appLabel,
		"threescale_component": "backend",
	}
}

func testRedisBackendRedisLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "backend",
		"threescale_component_element": "redis",
	}
}

func testRedisBackendRedisPodTemplateLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "backend",
		"threescale_component_element": "redis",
		"com.redhat.component-name":    "backend-redis",
		"com.redhat.component-type":    "application",
		"com.redhat.component-version": "32",
		"com.redhat.product-name":      "3scale",
		"com.redhat.product-version":   "master",
		"deploymentConfig":             "backend-redis",
	}
}

func defaultBackendRedisOptions() *component.BackendRedisOptions {
	tmpInsecure := insecureImportPolicy
	return &component.BackendRedisOptions{
		AmpRelease:                    product.ThreescaleRelease,
		ImageTag:                      product.ThreescaleRelease,
		Image:                         component.BackendRedisImageURL(),
		ContainerResourceRequirements: component.DefaultBackendRedisContainerResourceRequirements(),
		InsecureImportPolicy:          &tmpInsecure,
		BackendCommonLabels:           testRedisBackendCommonLabels(),
		RedisLabels:                   testRedisBackendRedisLabels(),
		PodTemplateLabels:             testRedisBackendRedisPodTemplateLabels(),
	}
}

func TestGetBackendRedisOptionsProvider(t *testing.T) {
	tmpFalseValue := false
	backendRedisImageURL := "redis:backendCustomVersion"

	cases := []struct {
		testName               string
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func() *component.BackendRedisOptions
	}{
		{"Default", basicApimanager, defaultBackendRedisOptions},
		{"WithoutResourceRequirements",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &tmpFalseValue
				return apimanager
			},
			func() *component.BackendRedisOptions {
				opts := defaultBackendRedisOptions()
				opts.ContainerResourceRequirements = &v1.ResourceRequirements{}
				return opts
			},
		},
		{"BackendRedisImageSet",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.Backend = &appsv1alpha1.BackendSpec{
					RedisImage: &backendRedisImageURL,
				}
				return apimanager
			},
			func() *component.BackendRedisOptions {
				opts := defaultBackendRedisOptions()
				opts.Image = backendRedisImageURL
				opts.PodTemplateLabels["com.redhat.component-version"] = "backendCustomVersion"
				return opts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			optsProvider := NewBackendRedisOptionsProvider(tc.apimanagerFactory())
			opts, err := optsProvider.GetBackendRedisOptions()
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
