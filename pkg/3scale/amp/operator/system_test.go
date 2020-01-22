package operator

import (
	"testing"

	"github.com/3scale/3scale-operator/pkg/helper"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func basicApimanagerSpecTestSystemOptions(name, namespace string) *appsv1alpha1.APIManager {
	wildcardDomain := "test.3scale.net"
	appLabel := "someLabel"
	tenantName := "someTenant"
	trueValue := true
	var oneValue int64 = 1
	falseValue := false
	apiastRegistryURL := "http://apicast:8090/policies"

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
			Apicast: &appsv1alpha1.ApicastSpec{
				RegistryURL: &apiastRegistryURL,
			},
			System: &appsv1alpha1.SystemSpec{
				FileStorageSpec: &appsv1alpha1.SystemFileStorageSpec{
					PVC: &appsv1alpha1.SystemPVCSpec{},
				},
				AppSpec: &appsv1alpha1.SystemAppSpec{
					Replicas: &oneValue,
				},
				SidekiqSpec: &appsv1alpha1.SystemSidekiqSpec{
					Replicas: &oneValue,
				},
			},
		},
	}
}

func getMemcachedSecret(namespace string) *v1.Secret {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      component.SystemSecretSystemMemcachedSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			component.SystemSecretSystemMemcachedServersFieldName: "mymemcache:11211",
		},
		Type: v1.SecretTypeOpaque,
	}
	secret.Data = helper.GetSecretDataFromStringData(secret.StringData)
	return secret
}

func getRecaptchaSecret(namespace string) *v1.Secret {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      component.SystemSecretSystemRecaptchaSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			component.SystemSecretSystemRecaptchaPublicKeyFieldName:  "someCaptchaPK",
			component.SystemSecretSystemRecaptchaPrivateKeyFieldName: "someCaptchaPrivate",
		},
		Type: v1.SecretTypeOpaque,
	}
	secret.Data = helper.GetSecretDataFromStringData(secret.StringData)
	return secret
}

func getEventHookSecret(namespace string) *v1.Secret {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      component.SystemSecretSystemEventsHookSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			component.SystemSecretSystemEventsHookPasswordFieldName: "somePassword",
			component.SystemSecretSystemEventsHookURLFieldName:      "http://mymaster:5000/somepath",
		},
		Type: v1.SecretTypeOpaque,
	}
	secret.Data = helper.GetSecretDataFromStringData(secret.StringData)
	return secret
}

func getSystemRedisSecret(namespace string) *v1.Secret {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      component.SystemSecretSystemRedisSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			component.SystemSecretSystemRedisNamespace:                   "systemRedis",
			component.SystemSecretSystemRedisURLFieldName:                "redis://system1:6379",
			component.SystemSecretSystemRedisSentinelHosts:               "someHosts1",
			component.SystemSecretSystemRedisSentinelRole:                "someRole1",
			component.SystemSecretSystemRedisMessageBusRedisNamespace:    "mbus",
			component.SystemSecretSystemRedisMessageBusSentinelHosts:     "someHosts2",
			component.SystemSecretSystemRedisMessageBusSentinelRole:      "someRole2",
			component.SystemSecretSystemRedisMessageBusRedisURLFieldName: "redis://system2:6379",
		},
		Type: v1.SecretTypeOpaque,
	}
	secret.Data = helper.GetSecretDataFromStringData(secret.StringData)
	return secret
}

func getSystemAppSecret(namespace string) *v1.Secret {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      component.SystemSecretSystemAppSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			component.SystemSecretSystemAppSecretKeyBaseFieldName: "somePassword",
		},
		Type: v1.SecretTypeOpaque,
	}
	secret.Data = helper.GetSecretDataFromStringData(secret.StringData)
	return secret
}

func getSystemSeedSecret(namespace string) *v1.Secret {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      component.SystemSecretSystemSeedSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			component.SystemSecretSystemSeedMasterDomainFieldName:      "masterDomainName",
			component.SystemSecretSystemSeedMasterUserFieldName:        "masterUsername",
			component.SystemSecretSystemSeedMasterPasswordFieldName:    "masterPasswd",
			component.SystemSecretSystemSeedMasterAccessTokenFieldName: "masterAccessToken",
			component.SystemSecretSystemSeedAdminUserFieldName:         "adminUsername",
			component.SystemSecretSystemSeedAdminPasswordFieldName:     "adminPasswd",
			component.SystemSecretSystemSeedAdminAccessTokenFieldName:  "adminAccessToken",
			component.SystemSecretSystemSeedAdminEmailFieldName:        "adminEmail",
		},
		Type: v1.SecretTypeOpaque,
	}
	secret.Data = helper.GetSecretDataFromStringData(secret.StringData)
	return secret
}

func getSystemMasterApicastSecret(namespace string) *v1.Secret {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      component.SystemSecretSystemMasterApicastSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			component.SystemSecretSystemMasterApicastAccessToken: "apicastAccessToken",
		},
		Type: v1.SecretTypeOpaque,
	}
	secret.Data = helper.GetSecretDataFromStringData(secret.StringData)
	return secret
}

func getAWSCredentialsSecret(name, namespace string) *v1.Secret {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		StringData: map[string]string{
			component.AwsAccessKeyID:     "my_aws_access_key_id",
			component.AwsSecretAccessKey: "my_aws_secret_access_key",
			component.AwsBucket:          "someBucket",
			component.AwsRegion:          "someRegion",
		},
		Type: v1.SecretTypeOpaque,
	}
	secret.Data = helper.GetSecretDataFromStringData(secret.StringData)
	return secret
}

func TestGetSystemOptions(t *testing.T) {
	name := "example-apimanager"
	namespace := "someNS"
	trueValue := true

	cases := []struct {
		testName                  string
		apimanagerFactory         func() *appsv1alpha1.APIManager
		memcachedSecret           *v1.Secret
		recaptchadSecret          *v1.Secret
		eventHookSecret           *v1.Secret
		redisSecret               *v1.Secret
		systemAppSecret           *v1.Secret
		systemSeedSecret          *v1.Secret
		systemMasterApicastSecret *v1.Secret
		awsCredentialsSecret      *v1.Secret
	}{
		{"WithResourceRequirements",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestSystemOptions(name, namespace)
				apimanager.Spec.ResourceRequirementsEnabled = &trueValue
				return apimanager
			}, nil, nil, nil, nil, nil, nil, nil, nil,
		},
		{"WithMemcachedSecret",
			func() *appsv1alpha1.APIManager {
				return basicApimanagerSpecTestSystemOptions(name, namespace)
			}, getMemcachedSecret(namespace), nil, nil, nil, nil, nil, nil, nil,
		},
		{"WithRecaptchaSecret",
			func() *appsv1alpha1.APIManager {
				return basicApimanagerSpecTestSystemOptions(name, namespace)
			}, nil, getRecaptchaSecret(namespace), nil, nil, nil, nil, nil, nil,
		},
		{"WithEventsHookSecret",
			func() *appsv1alpha1.APIManager {
				return basicApimanagerSpecTestSystemOptions(name, namespace)
			}, nil, nil, getEventHookSecret(namespace), nil, nil, nil, nil, nil,
		},
		{"WithRedisSecret",
			func() *appsv1alpha1.APIManager {
				return basicApimanagerSpecTestSystemOptions(name, namespace)
			}, nil, nil, nil, getSystemRedisSecret(namespace), nil, nil, nil, nil,
		},
		{"WithAppSecret",
			func() *appsv1alpha1.APIManager {
				return basicApimanagerSpecTestSystemOptions(name, namespace)
			}, nil, nil, nil, nil, getSystemAppSecret(namespace), nil, nil, nil,
		},
		{"WithSeedSecret",
			func() *appsv1alpha1.APIManager {
				return basicApimanagerSpecTestSystemOptions(name, namespace)
			}, nil, nil, nil, nil, nil, getSystemSeedSecret(namespace), nil, nil,
		},
		{"WithMasterApicastSecret",
			func() *appsv1alpha1.APIManager {
				return basicApimanagerSpecTestSystemOptions(name, namespace)
			}, nil, nil, nil, nil, nil, nil, getSystemMasterApicastSecret(namespace), nil,
		},
		{"WithS3",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestSystemOptions(name, namespace)
				apimanager.Spec.System.FileStorageSpec.PVC = nil
				apimanager.Spec.System.FileStorageSpec.S3 = &appsv1alpha1.SystemS3Spec{
					ConfigurationSecretRef: v1.LocalObjectReference{Name: "myawsauth"},
				}
				return apimanager
			}, nil, nil, nil, nil, nil, nil, nil, getAWSCredentialsSecret("myawsauth", namespace),
		},
		{"WithPVC",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestSystemOptions(name, namespace)
				tmp := "mystorageclassname"
				apimanager.Spec.System.FileStorageSpec.PVC = &appsv1alpha1.SystemPVCSpec{StorageClassName: &tmp}
				return apimanager
			}, nil, nil, nil, nil, nil, nil, nil, nil,
		},
		{"WithAppReplicas",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestSystemOptions(name, namespace)
				var tmp int64 = 5
				apimanager.Spec.System.AppSpec.Replicas = &tmp
				return apimanager
			}, nil, nil, nil, nil, nil, nil, nil, nil,
		},
		{"WithSidekiqReplicas",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestSystemOptions(name, namespace)
				var tmp int64 = 8
				apimanager.Spec.System.SidekiqSpec.Replicas = &tmp
				return apimanager
			}, nil, nil, nil, nil, nil, nil, nil, nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			objs := []runtime.Object{tc.apimanagerFactory()}
			if tc.memcachedSecret != nil {
				objs = append(objs, tc.memcachedSecret)
			}
			if tc.recaptchadSecret != nil {
				objs = append(objs, tc.recaptchadSecret)
			}
			if tc.eventHookSecret != nil {
				objs = append(objs, tc.eventHookSecret)
			}
			if tc.redisSecret != nil {
				objs = append(objs, tc.redisSecret)
			}
			if tc.systemAppSecret != nil {
				objs = append(objs, tc.systemAppSecret)
			}
			if tc.systemSeedSecret != nil {
				objs = append(objs, tc.systemSeedSecret)
			}
			if tc.systemMasterApicastSecret != nil {
				objs = append(objs, tc.systemMasterApicastSecret)
			}
			if tc.awsCredentialsSecret != nil {
				objs = append(objs, tc.awsCredentialsSecret)
			}
			cl := fake.NewFakeClient(objs...)
			optsProvider := OperatorSystemOptionsProvider{
				APIManagerSpec: &tc.apimanagerFactory().Spec,
				Namespace:      namespace,
				Client:         cl,
			}
			_, err := optsProvider.GetSystemOptions()
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
