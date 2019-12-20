package operator

import (
	"strings"
	"testing"

	"github.com/3scale/3scale-operator/pkg/helper"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func getTestSecret(namespace, secretName string, data map[string]string) *v1.Secret {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		StringData: data,
		Type:       v1.SecretTypeOpaque,
	}
	secret.Data = helper.GetSecretDataFromStringData(secret.StringData)
	return secret
}

func TestGetS3Options(t *testing.T) {
	wildcardDomain := "test.3scale.net"
	appLabel := "someLabel"
	name := "example-apimanager"
	namespace := "someNS"
	tenantName := "someTenant"
	trueValue := true
	s3SecretName := "s3secretname"

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
				ResourceRequirementsEnabled:  &trueValue,
			},
			System: &appsv1alpha1.SystemSpec{
				FileStorageSpec: &appsv1alpha1.SystemFileStorageSpec{
					S3: &appsv1alpha1.SystemS3Spec{
						AWSBucket: "mybucket.example.com",
						AWSRegion: "awsRegion",
						AWSCredentials: v1.LocalObjectReference{
							Name: s3SecretName,
						},
					},
				},
			},
		},
	}

	s3SecretData := map[string]string{
		component.S3SecretAWSAccessKeyIdFieldName:     "1234",
		component.S3SecretAWSSecretAccessKeyFieldName: "1234",
	}
	s3Secret := getTestSecret(namespace, s3SecretName, s3SecretData)
	objs := []runtime.Object{apimanager, s3Secret}
	cl := fake.NewFakeClient(objs...)
	optsProvider := OperatorS3OptionsProvider{
		APIManagerSpec: &apimanager.Spec,
		Namespace:      namespace,
		Client:         cl,
	}
	_, err := optsProvider.GetS3Options()
	if err != nil {
		t.Fatal(err)
	}
	// created "opts" cannot be tested  here, it only has set methods
	// and cannot assert on setted values from a different package
	// TODO: refactor options provider structure
	// then validate setted resources
}

func TestGetS3OptionsInvalid(t *testing.T) {
	wildcardDomain := "test.3scale.net"
	appLabel := "someLabel"
	name := "example-apimanager"
	namespace := "someNS"
	tenantName := "someTenant"
	trueValue := true
	s3SecretName := "s3secretname"

	cases := []struct {
		testName   string
		secretData map[string]string
		errSubstr  string
	}{
		{"NoSecretFound", nil, "not found"},
		{"S3AccessKeyIdMissing", map[string]string{component.S3SecretAWSSecretAccessKeyFieldName: "1234"}, component.S3SecretAWSAccessKeyIdFieldName},
		{"S3SecretAccessKeyMissing", map[string]string{component.S3SecretAWSAccessKeyIdFieldName: "1234"}, component.S3SecretAWSSecretAccessKeyFieldName},
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
						ResourceRequirementsEnabled:  &trueValue,
					},
					System: &appsv1alpha1.SystemSpec{
						FileStorageSpec: &appsv1alpha1.SystemFileStorageSpec{
							S3: &appsv1alpha1.SystemS3Spec{
								AWSBucket: "mybucket.example.com",
								AWSRegion: "awsRegion",
								AWSCredentials: v1.LocalObjectReference{
									Name: s3SecretName,
								},
							},
						},
					},
				},
			}

			objs := []runtime.Object{apimanager}
			if tc.secretData != nil {
				objs = append(objs, getTestSecret(namespace, s3SecretName, tc.secretData))
			}
			cl := fake.NewFakeClient(objs...)
			optsProvider := OperatorS3OptionsProvider{
				APIManagerSpec: &apimanager.Spec,
				Namespace:      namespace,
				Client:         cl,
			}
			_, err := optsProvider.GetS3Options()
			if err == nil {
				subT.Fatal("expected to fail for invalid s3 secret")
			}
			if !strings.Contains(err.Error(), tc.errSubstr) {
				subT.Fatalf("expected error regexp: %s, got: (%v)", tc.errSubstr, err)
			}
		})
	}
}
