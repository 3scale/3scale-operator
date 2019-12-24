package operator

import (
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetZyncOptions(t *testing.T) {
	wildcardDomain := "test.3scale.net"
	appLabel := "someLabel"
	name := "example-apimanager"
	namespace := "someNS"
	tenantName := "someTenant"
	trueValue := true
	var oneValue int64 = 1

	cases := []struct {
		testName                    string
		resourceRequirementsEnabled bool
		zyncSecret                  *v1.Secret
	}{
		{"WithResourceRequirements", true, nil},
		{"ZincSecret", false, getRedisSecret(namespace)},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			apimanager := &appsv1alpha1.APIManager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appsv1alpha1.APIManagerSpec{
					APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
						AppLabel:                     &appLabel,
						ImageStreamTagImportInsecure: &trueValue,
						WildcardDomain:               wildcardDomain,
						TenantName:                   &tenantName,
						ResourceRequirementsEnabled:  &tc.resourceRequirementsEnabled,
					},
					Zync: &appsv1alpha1.ZyncSpec{
						AppSpec: &appsv1alpha1.ZyncAppSpec{Replicas: &oneValue},
						QueSpec: &appsv1alpha1.ZyncQueSpec{Replicas: &oneValue},
					},
				},
			}
			objs := []runtime.Object{apimanager}
			if tc.zyncSecret != nil {
				objs = append(objs, tc.zyncSecret)
			}

			cl := fake.NewFakeClient(objs...)
			optsProvider := NewZyncOptionsProvider(&apimanager.Spec, namespace, cl)
			_, err := optsProvider.GetZyncOptions()
			if err != nil {
				t.Error(err)
			}
			// created "opts" cannot be tested  here, it only has set methods
			// and cannot assert on setted values from a different package
			// TODO: refactor options provider structure
			// then validate setted resources
		})
	}
}
