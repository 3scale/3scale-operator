package operator

import (
	"reflect"
	"testing"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func defaultMemcachedOptions() *component.MemcachedOptions {
	return &component.MemcachedOptions{
		AppLabel:             appLabel,
		ResourceRequirements: component.DefaultMemcachedResourceRequirements(),
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
