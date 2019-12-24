package operator

import (
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func basicApimanagerSpecTestOptions(name, namespace string) *appsv1alpha1.APIManager {
	wildcardDomain := "test.3scale.net"
	appLabel := "someLabel"
	tenantName := "someTenant"
	trueValue := true
	falseValue := false

	return &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				WildcardDomain:               wildcardDomain,
				AppLabel:                     &appLabel,
				ImageStreamTagImportInsecure: &trueValue,
				TenantName:                   &tenantName,
				ResourceRequirementsEnabled:  &falseValue,
			},
		},
	}
}

func TestMemcachedOptions(t *testing.T) {
	name := "example-apimanager"
	namespace := "someNS"
	trueValue := true
	falseValue := false

	cases := []struct {
		testName          string
		apimanagerFactory func() *appsv1alpha1.APIManager
	}{
		{"WithResourceRequirements",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestOptions(name, namespace)
				apimanager.Spec.ResourceRequirementsEnabled = &trueValue
				return apimanager
			},
		},
		{"WithoutResourceRequirements",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestOptions(name, namespace)
				apimanager.Spec.ResourceRequirementsEnabled = &falseValue
				return apimanager
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			optsProvider := NewMemcachedOptionsProvider(&tc.apimanagerFactory().Spec)
			_, err := optsProvider.GetMemcachedOptions()
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
