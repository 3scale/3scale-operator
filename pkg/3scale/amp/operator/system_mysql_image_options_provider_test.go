package operator

import (
	"reflect"
	"testing"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/google/go-cmp/cmp"
)

const (
	mysqlImage = "mysql:test"
)

func defaultSystemMySQLImageOptions() *component.SystemMySQLImageOptions {
	tmpInsecure := insecureImportPolicy
	return &component.SystemMySQLImageOptions{
		AppLabel:             appLabel,
		AmpRelease:           product.ThreescaleRelease,
		InsecureImportPolicy: &tmpInsecure,
		Image:                component.SystemMySQLImageURL(),
	}
}

func TestGetSystemMySQLImageOptions(t *testing.T) {
	tmpImageURL := mysqlImage
	cases := []struct {
		testName               string
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func() *component.SystemMySQLImageOptions
	}{
		{"Default", basicApimanager, defaultSystemMySQLImageOptions},
		{"ImageSet",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					DatabaseSpec: &appsv1alpha1.SystemDatabaseSpec{
						MySQL: &appsv1alpha1.SystemMySQLSpec{
							Image: &tmpImageURL,
						},
					},
				}
				return apimanager
			},
			func() *component.SystemMySQLImageOptions {
				opts := defaultSystemMySQLImageOptions()
				opts.Image = tmpImageURL
				return opts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			optsProvider := NewSystemMysqlImageOptionsProvider(tc.apimanagerFactory())
			opts, err := optsProvider.GetSystemMySQLImageOptions()
			if err != nil {
				subT.Error(err)
			}
			expectedOptions := tc.expectedOptionsFactory()
			if !reflect.DeepEqual(expectedOptions, opts) {
				subT.Errorf("Resulting expected options differ: %s", cmp.Diff(expectedOptions, opts))
			}
		})
	}
}
