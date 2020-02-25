package operator

import (
	"reflect"
	"testing"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	zyncReplica           int64 = 3
	zyncQueReplica        int64 = 4
	zyncSecretKeyBasename       = "someKeyBase"
	zyncDatabasePasswd          = "somePass3424"
	zyncAuthToken               = "someToken5252"
)

func getZyncSecret() *v1.Secret {
	data := map[string]string{
		component.ZyncSecretKeyBaseFieldName:             zyncSecretKeyBasename,
		component.ZyncSecretDatabasePasswordFieldName:    zyncDatabasePasswd,
		component.ZyncSecretAuthenticationTokenFieldName: zyncAuthToken,
	}
	return GetTestSecret(namespace, component.ZyncSecretName, data)
}

func basicApimanagerSpecTestZyncOptions() *appsv1alpha1.APIManager {
	tmpZyncReplicas := zyncReplica
	tmpZyncQueReplicas := zyncQueReplica

	apimanager := basicApimanager()
	apimanager.Spec.Zync = &appsv1alpha1.ZyncSpec{
		AppSpec: &appsv1alpha1.ZyncAppSpec{Replicas: &tmpZyncReplicas},
		QueSpec: &appsv1alpha1.ZyncQueSpec{Replicas: &tmpZyncQueReplicas},
	}
	return apimanager
}

func defaultZyncOptions(opts *component.ZyncOptions) *component.ZyncOptions {
	expectedOpts := &component.ZyncOptions{
		AppLabel:                              appLabel,
		ContainerResourceRequirements:         component.DefaultZyncContainerResourceRequirements(),
		QueContainerResourceRequirements:      component.DefaultZyncQueContainerResourceRequirements(),
		DatabaseContainerResourceRequirements: component.DefaultZyncDatabaseContainerResourceRequirements(),
		AuthenticationToken:                   opts.AuthenticationToken,
		DatabasePassword:                      opts.DatabasePassword,
		SecretKeyBase:                         opts.SecretKeyBase,
		ZyncReplicas:                          int32(zyncReplica),
		ZyncQueReplicas:                       int32(zyncQueReplica),
	}

	expectedOpts.DatabaseURL = component.DefaultZyncDatabaseURL(expectedOpts.DatabasePassword)

	return expectedOpts
}

func TestGetZyncOptionsProvider(t *testing.T) {
	falseValue := false

	cases := []struct {
		testName               string
		zyncSecret             *v1.Secret
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func(*component.ZyncOptions) *component.ZyncOptions
	}{
		{"Default", nil, basicApimanagerSpecTestZyncOptions,
			func(opts *component.ZyncOptions) *component.ZyncOptions {
				return defaultZyncOptions(opts)
			},
		},
		{"WithoutResourceRequirements", nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestZyncOptions()
				apimanager.Spec.ResourceRequirementsEnabled = &falseValue
				return apimanager
			},
			func(opts *component.ZyncOptions) *component.ZyncOptions {
				expectedOpts := defaultZyncOptions(opts)
				expectedOpts.ContainerResourceRequirements = v1.ResourceRequirements{}
				expectedOpts.QueContainerResourceRequirements = v1.ResourceRequirements{}
				expectedOpts.DatabaseContainerResourceRequirements = v1.ResourceRequirements{}
				return expectedOpts
			},
		},
		{"ZincSecret", getZyncSecret(), basicApimanagerSpecTestZyncOptions,
			func(opts *component.ZyncOptions) *component.ZyncOptions {
				expectedOpts := defaultZyncOptions(opts)
				expectedOpts.SecretKeyBase = zyncSecretKeyBasename
				expectedOpts.DatabasePassword = zyncDatabasePasswd
				expectedOpts.AuthenticationToken = zyncAuthToken
				expectedOpts.DatabaseURL = component.DefaultZyncDatabaseURL(zyncDatabasePasswd)
				return opts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			objs := []runtime.Object{}
			if tc.zyncSecret != nil {
				objs = append(objs, tc.zyncSecret)
			}
			cl := fake.NewFakeClient(objs...)
			optsProvider := NewZyncOptionsProvider(tc.apimanagerFactory(), namespace, cl)
			opts, err := optsProvider.GetZyncOptions()
			if err != nil {
				t.Error(err)
			}
			expectedOptions := tc.expectedOptionsFactory(opts)
			if !reflect.DeepEqual(expectedOptions, opts) {
				subT.Errorf("Resulting expected options differ: %s", cmp.Diff(expectedOptions, opts, cmpopts.IgnoreUnexported(resource.Quantity{})))
			}
		})
	}
}
