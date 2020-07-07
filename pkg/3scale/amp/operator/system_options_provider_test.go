package operator

import (
	"reflect"
	"testing"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	systemAppReplicas     int64 = 3
	systemSidekiqReplicas int64 = 4
	apicastRegistryURL          = "http://otherapicast:8090/policies"
)

func testSystemCommonLabels() map[string]string {
	return map[string]string{
		"app":                  appLabel,
		"threescale_component": "system",
	}
}

func testSystemCommonAppLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "app",
	}
}

func testSystemAppPodTemplateLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "app",
		"com.redhat.component-name":    "system-app",
		"com.redhat.component-type":    "application",
		"com.redhat.component-version": "nightly",
		"com.redhat.product-name":      "3scale",
		"com.redhat.product-version":   "master",
		"deploymentConfig":             "system-app",
	}
}

func testSystemCommonSidekiqLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "sidekiq",
	}
}

func testSystemSidekiqPodTemplateLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "sidekiq",
		"com.redhat.component-name":    "system-sidekiq",
		"com.redhat.component-type":    "application",
		"com.redhat.component-version": "nightly",
		"com.redhat.product-name":      "3scale",
		"com.redhat.product-version":   "master",
		"deploymentConfig":             "system-sidekiq",
	}
}

func testSystemProviderUILabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "provider-ui",
	}
}

func testSystemMasterUILabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "master-ui",
	}
}

func testSystemDeveloperUILabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "developer-ui",
	}
}

func testSystemSphinxLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "sphinx",
	}
}

func testSystemSphinxPodTemplateLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "sphinx",
		"com.redhat.component-name":    "system-sphinx",
		"com.redhat.component-type":    "application",
		"com.redhat.component-version": "nightly",
		"com.redhat.product-name":      "3scale",
		"com.redhat.product-version":   "master",
		"deploymentConfig":             "system-sphinx",
	}
}

func testSystemMemcachedLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "memcache",
	}
}

func testSystemSMTPLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "smtp",
	}
}

func testSystemAppAffinity() *v1.Affinity {
	return getTestAffinity("system-app")
}

func testSystemSidekiqAffinity() *v1.Affinity {
	return getTestAffinity("system-sidekiq")
}

func testSystemSphinxAffinity() *v1.Affinity {
	return getTestAffinity("system-sphinx")
}

func testSystemAppTolerations() []v1.Toleration {
	return getTestTolerations("system-app")
}

func testSystemSidekiqTolerations() []v1.Toleration {
	return getTestTolerations("system-sidekiq")
}

func testSystemSphinxTolerations() []v1.Toleration {
	return getTestTolerations("system-sphinx")
}

func basicApimanagerSpecTestSystemOptions() *appsv1alpha1.APIManager {
	tmpSystemAppReplicas := systemAppReplicas
	tmpSystemSideKiqReplicas := systemSidekiqReplicas
	tmpApicastRegistryURL := apicastRegistryURL

	apimanager := basicApimanager()
	apimanager.Spec.Apicast = &appsv1alpha1.ApicastSpec{RegistryURL: &tmpApicastRegistryURL}
	apimanager.Spec.System = &appsv1alpha1.SystemSpec{
		FileStorageSpec: &appsv1alpha1.SystemFileStorageSpec{},
		AppSpec:         &appsv1alpha1.SystemAppSpec{Replicas: &tmpSystemAppReplicas},
		SidekiqSpec:     &appsv1alpha1.SystemSidekiqSpec{Replicas: &tmpSystemSideKiqReplicas},
		SphinxSpec:      &appsv1alpha1.SystemSphinxSpec{},
	}
	apimanager.Spec.PodDisruptionBudget = &appsv1alpha1.PodDisruptionBudgetSpec{Enabled: true}
	return apimanager
}

func getMemcachedSecret() *v1.Secret {
	data := map[string]string{
		component.SystemSecretSystemMemcachedServersFieldName: "mymemcache:11211",
	}
	return GetTestSecret(namespace, component.SystemSecretSystemMemcachedSecretName, data)
}

func getRecaptchaSecret() *v1.Secret {
	data := map[string]string{
		component.SystemSecretSystemRecaptchaPublicKeyFieldName:  "someCaptchaPK",
		component.SystemSecretSystemRecaptchaPrivateKeyFieldName: "someCaptchaPrivate",
	}
	return GetTestSecret(namespace, component.SystemSecretSystemRecaptchaSecretName, data)
}

func getEventHookSecret() *v1.Secret {
	data := map[string]string{
		component.SystemSecretSystemEventsHookPasswordFieldName: "somePassword",
		component.SystemSecretSystemEventsHookURLFieldName:      "http://mymaster:5000/somepath",
	}
	return GetTestSecret(namespace, component.SystemSecretSystemEventsHookSecretName, data)
}

func getSystemRedisSecret() *v1.Secret {
	data := map[string]string{
		component.SystemSecretSystemRedisNamespace:                   "systemRedis",
		component.SystemSecretSystemRedisURLFieldName:                "redis://system1:6379",
		component.SystemSecretSystemRedisSentinelHosts:               "someHosts1",
		component.SystemSecretSystemRedisSentinelRole:                "someRole1",
		component.SystemSecretSystemRedisMessageBusRedisNamespace:    "mbus",
		component.SystemSecretSystemRedisMessageBusSentinelHosts:     "someHosts2",
		component.SystemSecretSystemRedisMessageBusSentinelRole:      "someRole2",
		component.SystemSecretSystemRedisMessageBusRedisURLFieldName: "redis://system2:6379",
	}
	return GetTestSecret(namespace, component.SystemSecretSystemRedisSecretName, data)
}

func getSystemAppSecret() *v1.Secret {
	data := map[string]string{
		component.SystemSecretSystemAppSecretKeyBaseFieldName: "somePassword234",
	}
	return GetTestSecret(namespace, component.SystemSecretSystemAppSecretName, data)
}

func getSystemSeedSecret() *v1.Secret {
	data := map[string]string{
		component.SystemSecretSystemSeedMasterDomainFieldName:      "masterDomainName",
		component.SystemSecretSystemSeedMasterUserFieldName:        "masterUsername",
		component.SystemSecretSystemSeedMasterPasswordFieldName:    "masterPasswd",
		component.SystemSecretSystemSeedMasterAccessTokenFieldName: "masterAccessToken",
		component.SystemSecretSystemSeedAdminUserFieldName:         "adminUsername",
		component.SystemSecretSystemSeedAdminPasswordFieldName:     "adminPasswd",
		component.SystemSecretSystemSeedAdminAccessTokenFieldName:  "adminAccessToken",
		component.SystemSecretSystemSeedAdminEmailFieldName:        "myemail@example.com",
	}
	return GetTestSecret(namespace, component.SystemSecretSystemSeedSecretName, data)
}

func getSystemMasterApicastSecret() *v1.Secret {
	data := map[string]string{
		component.SystemSecretSystemMasterApicastAccessToken: "apicastAccessToken",
	}
	return GetTestSecret(namespace, component.SystemSecretSystemMasterApicastSecretName, data)
}

func defaultSystemOptions(opts *component.SystemOptions) *component.SystemOptions {
	recaptchaPublicKey := component.DefaultRecaptchaPublickey()
	recaptchaPrivateKey := component.DefaultRecaptchaPrivatekey()
	redisSentinelHosts := component.DefaultSystemRedisSentinelHosts()
	redisSentinelRole := component.DefaultSystemRedisSentinelRole()
	messageBusRedisURL := component.DefaultSystemMessageBusRedisURL()
	messageBusRedisSentinelHosts := component.DefaultSystemMessageBusRedisSentinelHosts()
	messageBusRedisSentinelRole := component.DefaultSystemMessageBusRedisSentinelRole()
	messageBusRedisNamespace := component.DefaultSystemMessageBusRedisNamespace()
	redisNamespace := component.DefaultSystemRedisNamespace()
	tmpSystemAppReplicas := int32(systemAppReplicas)
	tmpSystemSideKiqReplicas := int32(systemSidekiqReplicas)
	tmpSystemAdminEmail := component.DefaultSystemAdminEmail()
	tmpSMTPAddress := component.DefaultSystemSMTPAddress()
	tmpSMTPAuthentication := component.DefaultSystemSMTPAuthentication()
	tmpSMTPDomain := component.DefaultSystemSMTPDomain()
	tmpSMTPOpenSSLVerifyMode := component.DefaultSystemSMTPOpenSSLVerifyMode()
	tmpSMTPPassword := opts.SmtpSecretOptions.Password
	tmpSMTPPort := component.DefaultSystemSMTPPort()
	tmpSMTPUsername := component.DefaultSystemSMTPUsername()

	expectedOpts := &component.SystemOptions{
		TenantName:                               tenantName,
		WildcardDomain:                           wildcardDomain,
		AmpRelease:                               product.ThreescaleRelease,
		ImageTag:                                 product.ThreescaleRelease,
		ApicastRegistryURL:                       apicastRegistryURL,
		AppMasterContainerResourceRequirements:   component.DefaultAppMasterContainerResourceRequirements(),
		AppProviderContainerResourceRequirements: component.DefaultAppProviderContainerResourceRequirements(),
		AppDeveloperContainerResourceRequirements: component.DefaultAppDeveloperContainerResourceRequirements(),
		SidekiqContainerResourceRequirements:      component.DefaultSidekiqContainerResourceRequirements(),
		SphinxContainerResourceRequirements:       component.DefaultSphinxContainerResourceRequirements(),
		MemcachedServers:                          component.DefaultMemcachedServers(),
		RecaptchaPublicKey:                        &recaptchaPublicKey,
		RecaptchaPrivateKey:                       &recaptchaPrivateKey,
		BackendSharedSecret:                       opts.BackendSharedSecret,
		EventHooksURL:                             component.DefaultEventHooksURL(),
		AppSecretKeyBase:                          opts.AppSecretKeyBase,
		MasterName:                                component.DefaultSystemMasterName(),
		MasterUsername:                            component.DefaultSystemMasterUsername(),
		MasterPassword:                            opts.MasterPassword,
		AdminUsername:                             component.DefaultSystemAdminUsername(),
		AdminPassword:                             opts.AdminPassword,
		AdminAccessToken:                          opts.AdminAccessToken,
		MasterAccessToken:                         opts.MasterAccessToken,
		ApicastAccessToken:                        opts.ApicastAccessToken,
		AppReplicas:                               &tmpSystemAppReplicas,
		SidekiqReplicas:                           &tmpSystemSideKiqReplicas,
		RedisURL:                                  component.DefaultSystemRedisURL(),
		RedisSentinelHosts:                        &redisSentinelHosts,
		RedisSentinelRole:                         &redisSentinelRole,
		MessageBusRedisURL:                        &messageBusRedisURL,
		MessageBusRedisSentinelHosts:              &messageBusRedisSentinelHosts,
		MessageBusRedisSentinelRole:               &messageBusRedisSentinelRole,
		MessageBusRedisNamespace:                  &messageBusRedisNamespace,
		RedisNamespace:                            &redisNamespace,
		AdminEmail:                                &tmpSystemAdminEmail,
		PvcFileStorageOptions:                     &component.PVCFileStorageOptions{},
		SmtpSecretOptions: component.SystemSMTPSecretOptions{
			Address:           &tmpSMTPAddress,
			Authentication:    &tmpSMTPAuthentication,
			Domain:            &tmpSMTPDomain,
			OpenSSLVerifyMode: &tmpSMTPOpenSSLVerifyMode,
			Password:          tmpSMTPPassword,
			Port:              &tmpSMTPPort,
			Username:          &tmpSMTPUsername,
		},
		CommonLabels:             testSystemCommonLabels(),
		CommonAppLabels:          testSystemCommonAppLabels(),
		AppPodTemplateLabels:     testSystemAppPodTemplateLabels(),
		CommonSidekiqLabels:      testSystemCommonSidekiqLabels(),
		SidekiqPodTemplateLabels: testSystemSidekiqPodTemplateLabels(),
		ProviderUILabels:         testSystemProviderUILabels(),
		MasterUILabels:           testSystemMasterUILabels(),
		DeveloperUILabels:        testSystemDeveloperUILabels(),
		SphinxLabels:             testSystemSphinxLabels(),
		SphinxPodTemplateLabels:  testSystemSphinxPodTemplateLabels(),
		MemcachedLabels:          testSystemMemcachedLabels(),
		SMTPLabels:               testSystemSMTPLabels(),
		SideKiqMetrics:           true,
	}

	expectedOpts.ApicastSystemMasterProxyConfigEndpoint = component.DefaultApicastSystemMasterProxyConfigEndpoint(opts.ApicastAccessToken)
	return expectedOpts
}

func TestGetSystemOptionsProvider(t *testing.T) {
	falseValue := false

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
		expectedOptionsFactory    func(*component.SystemOptions) *component.SystemOptions
	}{
		{"Default", basicApimanagerSpecTestSystemOptions,
			nil, nil, nil, nil, nil, nil, nil,
			func(opts *component.SystemOptions) *component.SystemOptions {
				return defaultSystemOptions(opts)
			},
		},
		{"WithoutResourceRequirements",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestSystemOptions()
				apimanager.Spec.ResourceRequirementsEnabled = &falseValue
				return apimanager
			}, nil, nil, nil, nil, nil, nil, nil,
			func(opts *component.SystemOptions) *component.SystemOptions {
				expectedOpts := defaultSystemOptions(opts)
				expectedOpts.AppMasterContainerResourceRequirements = &v1.ResourceRequirements{}
				expectedOpts.AppProviderContainerResourceRequirements = &v1.ResourceRequirements{}
				expectedOpts.AppDeveloperContainerResourceRequirements = &v1.ResourceRequirements{}
				expectedOpts.SidekiqContainerResourceRequirements = &v1.ResourceRequirements{}
				expectedOpts.SphinxContainerResourceRequirements = &v1.ResourceRequirements{}
				return expectedOpts
			},
		},
		{"WithMemcachedSecret", basicApimanagerSpecTestSystemOptions,
			getMemcachedSecret(), nil, nil, nil, nil, nil, nil,
			func(opts *component.SystemOptions) *component.SystemOptions {
				expectedOpts := defaultSystemOptions(opts)
				expectedOpts.MemcachedServers = "mymemcache:11211"
				return expectedOpts
			},
		},
		{"WithRecaptchaSecret", basicApimanagerSpecTestSystemOptions,
			nil, getRecaptchaSecret(), nil, nil, nil, nil, nil,
			func(opts *component.SystemOptions) *component.SystemOptions {
				expectedOpts := defaultSystemOptions(opts)
				recaptchaPublicKey := "someCaptchaPK"
				expectedOpts.RecaptchaPublicKey = &recaptchaPublicKey
				recaptchaPrivateKey := "someCaptchaPrivate"
				expectedOpts.RecaptchaPrivateKey = &recaptchaPrivateKey
				return expectedOpts
			},
		},
		{"WithEventsHookSecret", basicApimanagerSpecTestSystemOptions,
			nil, nil, getEventHookSecret(), nil, nil, nil, nil,
			func(opts *component.SystemOptions) *component.SystemOptions {
				expectedOpts := defaultSystemOptions(opts)
				expectedOpts.BackendSharedSecret = "somePassword"
				expectedOpts.EventHooksURL = "http://mymaster:5000/somepath"
				return expectedOpts
			},
		},
		{"WithRedisSecret", basicApimanagerSpecTestSystemOptions,
			nil, nil, nil, getSystemRedisSecret(), nil, nil, nil,
			func(opts *component.SystemOptions) *component.SystemOptions {
				expectedOpts := defaultSystemOptions(opts)
				expectedOpts.RedisURL = "redis://system1:6379"
				redisSentinelHosts := "someHosts1"
				expectedOpts.RedisSentinelHosts = &redisSentinelHosts
				redisSentinelRole := "someRole1"
				expectedOpts.RedisSentinelRole = &redisSentinelRole
				messageBusRedisURL := "redis://system2:6379"
				expectedOpts.MessageBusRedisURL = &messageBusRedisURL
				messageBusRedisSentinelHosts := "someHosts2"
				expectedOpts.MessageBusRedisSentinelHosts = &messageBusRedisSentinelHosts
				messageBusRedisSentinelRole := "someRole2"
				expectedOpts.MessageBusRedisSentinelRole = &messageBusRedisSentinelRole
				redisNamespace := "systemRedis"
				expectedOpts.RedisNamespace = &redisNamespace
				messageBusRedisNamespace := "mbus"
				expectedOpts.MessageBusRedisNamespace = &messageBusRedisNamespace
				return expectedOpts
			},
		},
		{"WithAppSecret", basicApimanagerSpecTestSystemOptions,
			nil, nil, nil, nil, getSystemAppSecret(), nil, nil,
			func(opts *component.SystemOptions) *component.SystemOptions {
				expectedOpts := defaultSystemOptions(opts)
				expectedOpts.AppSecretKeyBase = "somePassword234"
				return expectedOpts
			},
		},
		{"WithSeedSecret", basicApimanagerSpecTestSystemOptions,
			nil, nil, nil, nil, nil, getSystemSeedSecret(), nil,
			func(opts *component.SystemOptions) *component.SystemOptions {
				expectedOpts := defaultSystemOptions(opts)
				expectedOpts.MasterName = "masterDomainName"
				expectedOpts.MasterUsername = "masterUsername"
				expectedOpts.MasterPassword = "masterPasswd"
				expectedOpts.AdminUsername = "adminUsername"
				expectedOpts.AdminPassword = "adminPasswd"
				expectedOpts.AdminAccessToken = "adminAccessToken"
				expectedOpts.MasterAccessToken = "masterAccessToken"
				adminEmail := "myemail@example.com"
				expectedOpts.AdminEmail = &adminEmail
				return expectedOpts
			},
		},
		{"WithMasterApicastSecret", basicApimanagerSpecTestSystemOptions,
			nil, nil, nil, nil, nil, nil, getSystemMasterApicastSecret(),
			func(opts *component.SystemOptions) *component.SystemOptions {
				expectedOpts := defaultSystemOptions(opts)
				expectedOpts.ApicastAccessToken = "apicastAccessToken"
				expectedOpts.ApicastSystemMasterProxyConfigEndpoint = component.DefaultApicastSystemMasterProxyConfigEndpoint("apicastAccessToken")
				return opts
			},
		},
		{"WithS3",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestSystemOptions()
				apimanager.Spec.System.FileStorageSpec.PVC = nil
				apimanager.Spec.System.FileStorageSpec.S3 = &appsv1alpha1.SystemS3Spec{
					ConfigurationSecretRef: v1.LocalObjectReference{Name: "myawsauth"},
				}
				return apimanager
			},
			nil, nil, nil, nil, nil, nil, nil,
			func(opts *component.SystemOptions) *component.SystemOptions {
				expectedOpts := defaultSystemOptions(opts)
				expectedOpts.S3FileStorageOptions = &component.S3FileStorageOptions{
					ConfigurationSecretName: "myawsauth",
				}
				expectedOpts.PvcFileStorageOptions = nil
				return expectedOpts
			},
		},
		{"WithPVC",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestSystemOptions()
				tmp := "mystorageclassname"
				apimanager.Spec.System.FileStorageSpec.PVC = &appsv1alpha1.SystemPVCSpec{StorageClassName: &tmp}
				return apimanager
			}, nil, nil, nil, nil, nil, nil, nil,
			func(opts *component.SystemOptions) *component.SystemOptions {
				expectedOpts := defaultSystemOptions(opts)
				tmp := "mystorageclassname"
				expectedOpts.PvcFileStorageOptions = &component.PVCFileStorageOptions{StorageClass: &tmp}
				expectedOpts.S3FileStorageOptions = nil
				return expectedOpts
			},
		},
		{"WithAffinity",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestSystemOptions()
				apimanager.Spec.System.AppSpec.Affinity = testSystemAppAffinity()
				apimanager.Spec.System.SidekiqSpec.Affinity = testSystemSidekiqAffinity()
				apimanager.Spec.System.SphinxSpec.Affinity = testSystemSphinxAffinity()
				return apimanager
			}, nil, nil, nil, nil, nil, nil, nil,
			func(opts *component.SystemOptions) *component.SystemOptions {
				expectedOpts := defaultSystemOptions(opts)
				expectedOpts.AppAffinity = testSystemAppAffinity()
				expectedOpts.SidekiqAffinity = testSystemSidekiqAffinity()
				expectedOpts.SphinxAffinity = testSystemSphinxAffinity()
				return expectedOpts
			},
		},
		{"WithTolerations",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestSystemOptions()
				apimanager.Spec.System.AppSpec.Tolerations = testSystemAppTolerations()
				apimanager.Spec.System.SidekiqSpec.Tolerations = testSystemSidekiqTolerations()
				apimanager.Spec.System.SphinxSpec.Tolerations = testSystemSphinxTolerations()
				return apimanager
			}, nil, nil, nil, nil, nil, nil, nil,
			func(opts *component.SystemOptions) *component.SystemOptions {
				expectedOpts := defaultSystemOptions(opts)
				expectedOpts.AppTolerations = testSystemAppTolerations()
				expectedOpts.SidekiqTolerations = testSystemSidekiqTolerations()
				expectedOpts.SphinxTolerations = testSystemSphinxTolerations()
				return expectedOpts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			objs := []runtime.Object{}
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
			cl := fake.NewFakeClient(objs...)
			optsProvider := NewSystemOptionsProvider(tc.apimanagerFactory(), namespace, cl)
			opts, err := optsProvider.GetSystemOptions()
			if err != nil {
				subT.Error(err)
			}
			expectedOptions := tc.expectedOptionsFactory(opts)
			if !reflect.DeepEqual(expectedOptions, opts) {
				subT.Errorf("Resulting expected options differ: %s", cmp.Diff(expectedOptions, opts, cmpopts.IgnoreUnexported(resource.Quantity{})))
			}
		})
	}
}
