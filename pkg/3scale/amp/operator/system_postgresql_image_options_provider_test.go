package operator

import (
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetSystemPostgreSQLImageOptions(t *testing.T) {
	appLabel := "someLabel"
	name := "example-apimanager"
	namespace := "someNS"
	trueValue := true
	imageUrl := "postgresql:test"

	cases := []struct {
		testName   string
		apimanager *appsv1alpha1.APIManager
	}{
		{
			"ImageSet",
			&appsv1alpha1.APIManager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appsv1alpha1.APIManagerSpec{
					APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
						AppLabel:                     &appLabel,
						ImageStreamTagImportInsecure: &trueValue,
					},
					System: &appsv1alpha1.SystemSpec{
						DatabaseSpec: &appsv1alpha1.SystemDatabaseSpec{
							PostgreSQL: &appsv1alpha1.SystemPostgreSQLSpec{
								Image: &imageUrl,
							},
						},
					},
				},
			},
		},
		{
			"ImageNotSet",
			&appsv1alpha1.APIManager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appsv1alpha1.APIManagerSpec{
					APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
						AppLabel:                     &appLabel,
						ImageStreamTagImportInsecure: &trueValue,
					},
					System: &appsv1alpha1.SystemSpec{
						DatabaseSpec: &appsv1alpha1.SystemDatabaseSpec{
							PostgreSQL: &appsv1alpha1.SystemPostgreSQLSpec{},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			optsProvider := NewSystemPostgreSQLImageOptionsProvider(tc.apimanager)
			_, err := optsProvider.GetSystemPostgreSQLImageOptions()
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
