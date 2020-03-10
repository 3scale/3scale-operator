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

func defaultRedisOptions() *component.RedisOptions {
	tmpInsecure := insecureImportPolicy
	return &component.RedisOptions{
		AppLabel:     appLabel,
		AmpRelease:   product.ThreescaleRelease,
		BackendImage: component.BackendRedisImageURL(),
		SystemImage:  component.SystemRedisImageURL(),
		BackendRedisContainerResourceRequirements: component.DefaultBackendRedisContainerResourceRequirements(),
		SystemRedisContainerResourceRequirements:  component.DefaultSystemRedisContainerResourceRequirements(),
		InsecureImportPolicy:                      &tmpInsecure,
	}
}

func TestGetRedisOptionsProvider(t *testing.T) {
	tmpFalseValue := false
	backendRedisImageURL := "redis:backend"
	systemRedisImageURL := "redis:system"

	cases := []struct {
		testName               string
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func() *component.RedisOptions
	}{
		{"Default", basicApimanager, defaultRedisOptions},
		{"WithoutResourceRequirements",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &tmpFalseValue
				return apimanager
			},
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				opts.BackendRedisContainerResourceRequirements = &v1.ResourceRequirements{}
				opts.SystemRedisContainerResourceRequirements = &v1.ResourceRequirements{}
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
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				opts.BackendImage = backendRedisImageURL
				return opts
			},
		},
		{"SystemRedisImageSet",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					RedisImage: &systemRedisImageURL,
				}
				return apimanager
			},
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				opts.SystemImage = systemRedisImageURL
				return opts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			optsProvider := NewRedisOptionsProvider(tc.apimanagerFactory())
			opts, err := optsProvider.GetRedisOptions()
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
