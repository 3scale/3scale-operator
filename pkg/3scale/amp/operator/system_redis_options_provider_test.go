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

func testRedisSystemCommonLabels() map[string]string {
	return map[string]string{
		"app":                  appLabel,
		"threescale_component": "system",
	}
}

func testRedisSystemRedisLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "redis",
	}
}

func testRedisSystemRedisPodTemplateLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "redis",
		"com.redhat.component-name":    "system-redis",
		"com.redhat.component-type":    "application",
		"com.redhat.component-version": "32",
		"com.redhat.product-name":      "3scale",
		"com.redhat.product-version":   "master",
		"deploymentConfig":             "system-redis",
	}
}

func defaultSystemRedisOptions() *component.SystemRedisOptions {
	tmpInsecure := insecureImportPolicy
	return &component.SystemRedisOptions{
		AmpRelease:                    product.ThreescaleRelease,
		ImageTag:                      product.ThreescaleRelease,
		Image:                         component.SystemRedisImageURL(),
		ContainerResourceRequirements: component.DefaultSystemRedisContainerResourceRequirements(),
		InsecureImportPolicy:          &tmpInsecure,
		SystemCommonLabels:            testRedisSystemCommonLabels(),
		RedisLabels:                   testRedisSystemRedisLabels(),
		PodTemplateLabels:             testRedisSystemRedisPodTemplateLabels(),
	}
}

func TestGetSystemRedisOptionsProvider(t *testing.T) {
	tmpFalseValue := false
	systemRedisImageURL := "redis:systemCustomVersion"

	cases := []struct {
		testName               string
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func() *component.SystemRedisOptions
	}{
		{"Default", basicApimanager, defaultSystemRedisOptions},
		{"WithoutResourceRequirements",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &tmpFalseValue
				return apimanager
			},
			func() *component.SystemRedisOptions {
				opts := defaultSystemRedisOptions()
				opts.ContainerResourceRequirements = &v1.ResourceRequirements{}
				return opts
			},
		},
		{"SystemRedisImageSet",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					RedisDatabaseSpec: &appsv1alpha1.SystemRedisDatabaseSpec{
						SystemRedisDatabaseModeSpec: appsv1alpha1.SystemRedisDatabaseModeSpec{
							EmbeddedDatabaseSpec: &appsv1alpha1.SystemRedisDatabaseEmbeddedSpec{
								Image: &systemRedisImageURL,
							},
						},
					},
				}
				return apimanager
			},
			func() *component.SystemRedisOptions {
				opts := defaultSystemRedisOptions()
				opts.Image = systemRedisImageURL
				opts.PodTemplateLabels["com.redhat.component-version"] = "systemCustomVersion"
				return opts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			optsProvider := NewSystemRedisOptionsProvider(tc.apimanagerFactory())
			opts, err := optsProvider.GetSystemRedisOptions()
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
