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
	postgresqlImageURL = "postgresql:test"
)

func defaultSystemPostgreSQLImageOptions() *component.SystemPostgreSQLImageOptions {
	tmpInsecure := insecureImportPolicy
	return &component.SystemPostgreSQLImageOptions{
		AppLabel:             appLabel,
		AmpRelease:           product.ThreescaleRelease,
		InsecureImportPolicy: &tmpInsecure,
		Image:                component.SystemPostgreSQLImageURL(),
	}
}

func TestGetSystemPostgreSQLImageOptionsProvider(t *testing.T) {
	tmpImageURL := postgresqlImageURL

	cases := []struct {
		testName               string
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func() *component.SystemPostgreSQLImageOptions
	}{
		{"Default", basicApimanager, defaultSystemPostgreSQLImageOptions},
		{"ImageSet",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					DatabaseSpec: &appsv1alpha1.SystemDatabaseSpec{
						PostgreSQL: &appsv1alpha1.SystemPostgreSQLSpec{
							Image: &tmpImageURL,
						},
					},
				}
				return apimanager
			},
			func() *component.SystemPostgreSQLImageOptions {
				opts := defaultSystemPostgreSQLImageOptions()
				opts.Image = tmpImageURL
				return opts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			optsProvider := NewSystemPostgreSQLImageOptionsProvider(tc.apimanagerFactory())
			opts, err := optsProvider.GetSystemPostgreSQLImageOptions()
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
