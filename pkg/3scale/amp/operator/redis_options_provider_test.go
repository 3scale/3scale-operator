package operator

import (
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetRedisOptions(t *testing.T) {
	appLabel := "someLabel"
	name := "example-apimanager"
	namespace := "someNS"
	trueValue := true
	falseValue := false
	backendRedisImageUrl := "redis:backend"
	systemRedisImageUrl := "redis:system"
	tenantName := "someTenant"

	cases := []struct {
		testName   string
		apimanager *appsv1alpha1.APIManager
	}{
		{"WithResourceRequirements",
			&appsv1alpha1.APIManager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appsv1alpha1.APIManagerSpec{
					APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
						AppLabel:                     &appLabel,
						ImageStreamTagImportInsecure: &trueValue,
						TenantName:                   &tenantName,
						ResourceRequirementsEnabled:  &trueValue,
					},
				},
			},
		},
		{"BackendRedisImageSet",
			&appsv1alpha1.APIManager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appsv1alpha1.APIManagerSpec{
					APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
						AppLabel:                     &appLabel,
						ImageStreamTagImportInsecure: &trueValue,
						TenantName:                   &tenantName,
						ResourceRequirementsEnabled:  &falseValue,
					},
					Backend: &appsv1alpha1.BackendSpec{
						RedisImage: &backendRedisImageUrl,
					},
				},
			},
		},
		{"SystemRedisImageSet",
			&appsv1alpha1.APIManager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appsv1alpha1.APIManagerSpec{
					APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
						AppLabel:                     &appLabel,
						ImageStreamTagImportInsecure: &trueValue,
						TenantName:                   &tenantName,
						ResourceRequirementsEnabled:  &falseValue,
					},
					System: &appsv1alpha1.SystemSpec{
						RedisImage: &systemRedisImageUrl,
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			optsProvider := NewRedisOptionsProvider(tc.apimanager)
			_, err := optsProvider.GetRedisOptions()
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
