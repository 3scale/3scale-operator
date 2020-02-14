package operator

import (
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
)

func TestGetSystemMySQLImageOptions(t *testing.T) {
	name := "example-apimanager"
	namespace := "someNS"
	imageUrl := "mysql:test"
	trueValue := true
	falseValue := false

	cases := []struct {
		testName          string
		apimanagerFactory func() *appsv1alpha1.APIManager
	}{
		{
			"ImageSet",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestOptions(name, namespace)
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					DatabaseSpec: &appsv1alpha1.SystemDatabaseSpec{
						MySQL: &appsv1alpha1.SystemMySQLSpec{
							Image: &imageUrl,
						},
					},
				}
				return apimanager
			},
		},
		{
			"ImageNotSet",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestOptions(name, namespace)
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					DatabaseSpec: &appsv1alpha1.SystemDatabaseSpec{
						MySQL: &appsv1alpha1.SystemMySQLSpec{},
					},
				}
				return apimanager
			},
		},
		{
			"ImageStreamInsecureFalse",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestOptions(name, namespace)
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					DatabaseSpec: &appsv1alpha1.SystemDatabaseSpec{
						MySQL: &appsv1alpha1.SystemMySQLSpec{},
					},
				}
				apimanager.Spec.ImageStreamTagImportInsecure = &falseValue
				return apimanager
			},
		},
		{
			"ImageStreamInsecureTrue",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestOptions(name, namespace)
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					DatabaseSpec: &appsv1alpha1.SystemDatabaseSpec{
						MySQL: &appsv1alpha1.SystemMySQLSpec{},
					},
				}
				apimanager.Spec.ImageStreamTagImportInsecure = &trueValue
				return apimanager
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			optsProvider := NewSystemMysqlImageOptionsProvider(tc.apimanagerFactory())
			_, err := optsProvider.GetSystemMySQLImageOptions()
			if err != nil {
				subT.Error(err)
			}
			// created "opts" cannot be tested  here, it only has set methods
			// and cannot assert on setted values from a different package
			// TODO: refactor options provider structure
			// then validate setted resources
		})
	}
}
