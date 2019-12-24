package operator

import (
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
)

func TestGetAmpImagesOptions(t *testing.T) {
	appLabel := "someLabel"
	apicastImage := "quay.io/3scale/apicast:mytag"
	backendImage := "quay.io/3scale/backend:mytag"
	systemImage := "quay.io/3scale/backend:mytag"
	zyncImage := "quay.io/3scale/zync:mytag"
	zyncPostgresqlImage := "postgresql-10:mytag"
	systemMemcachedImage := "memcached:mytag"
	trueValue := true

	cases := []struct {
		name       string
		apimanager *appsv1alpha1.APIManagerSpec
	}{
		{
			"apicastImage", &appsv1alpha1.APIManagerSpec{
				APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
					AppLabel:                     &appLabel,
					ImageStreamTagImportInsecure: &trueValue,
				},
				Apicast: &appsv1alpha1.ApicastSpec{Image: &apicastImage},
			},
		},
		{
			"backendImage", &appsv1alpha1.APIManagerSpec{
				APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
					AppLabel:                     &appLabel,
					ImageStreamTagImportInsecure: &trueValue,
				},
				Backend: &appsv1alpha1.BackendSpec{Image: &backendImage},
			},
		},
		{
			"systemImage", &appsv1alpha1.APIManagerSpec{
				APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
					AppLabel:                     &appLabel,
					ImageStreamTagImportInsecure: &trueValue,
				},
				System: &appsv1alpha1.SystemSpec{Image: &systemImage},
			},
		},
		{
			"zyncImage", &appsv1alpha1.APIManagerSpec{
				APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
					AppLabel:                     &appLabel,
					ImageStreamTagImportInsecure: &trueValue,
				},
				Zync: &appsv1alpha1.ZyncSpec{Image: &zyncImage},
			},
		},
		{
			"zyncPostgresqlImage", &appsv1alpha1.APIManagerSpec{
				APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
					AppLabel:                     &appLabel,
					ImageStreamTagImportInsecure: &trueValue,
				},
				Zync: &appsv1alpha1.ZyncSpec{PostgreSQLImage: &zyncPostgresqlImage},
			},
		},
		{
			"systemMemcachedImage", &appsv1alpha1.APIManagerSpec{
				APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
					AppLabel:                     &appLabel,
					ImageStreamTagImportInsecure: &trueValue,
				},
				System: &appsv1alpha1.SystemSpec{MemcachedImage: &systemMemcachedImage},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(subT *testing.T) {
			optsProvider := NewAmpImagesOptionsProvider(tc.apimanager)
			_, err := optsProvider.GetAmpImagesOptions()
			if err != nil {
				subT.Error(err)
			}
			// created "opts" cannot be tested  here, it only has set methods
			// and cannot assert on setted values from a different package
			// TODO: refactor options provider structure
			// then validate setted images are being used
		})
	}
}
