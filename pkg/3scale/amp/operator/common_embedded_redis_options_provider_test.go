package operator

import (
	"reflect"
	"testing"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/apimachinery/pkg/api/resource"
)

func testCommonEmbeddedRedisConfigMapLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "redis",
	}
}

func defaultCommonEmbeddedRedisOptions() *component.CommonEmbeddedRedisOptions {
	return &component.CommonEmbeddedRedisOptions{
		ConfigMapLabels: testCommonEmbeddedRedisConfigMapLabels(),
	}
}

func TestGetCommonEmbeddedRedisOptionsProvider(t *testing.T) {

	cases := []struct {
		testName               string
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func() *component.CommonEmbeddedRedisOptions
	}{
		{"Default", basicApimanager, defaultCommonEmbeddedRedisOptions},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			optsProvider := NewCommonEmbeddedRedisOptionProvider(tc.apimanagerFactory())
			opts, err := optsProvider.GetCommonEmbeddedRedisOptions()
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
