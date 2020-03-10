package operator

import (
	"fmt"
	"reflect"
	"strings"
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
	systemMysqlUsername = "user"
	systemMysqlPassword = "password1"
)

func defaultSystemMysqlOptions(opts *component.SystemMysqlOptions) *component.SystemMysqlOptions {
	return &component.SystemMysqlOptions{
		AppLabel:                      appLabel,
		DatabaseName:                  component.DefaultSystemMysqlDatabaseName(),
		User:                          component.DefaultSystemMysqlUser(),
		Password:                      opts.Password,
		RootPassword:                  opts.RootPassword,
		DatabaseURL:                   component.DefaultSystemMysqlDatabaseURL(opts.RootPassword, component.DefaultSystemMysqlDatabaseName()),
		ContainerResourceRequirements: component.DefaultSystemMysqlResourceRequirements(),
	}
}

func TestGetMysqlOptionsProvider(t *testing.T) {
	tmpFalseValue := false
	systemMysqlRootPassword := "rootPassw1"
	systemMysqlDatabaseName := "myDatabaseName"
	databaseURL := fmt.Sprintf("mysql2://root:%s@system-mysql/%s", systemMysqlRootPassword, systemMysqlDatabaseName)

	cases := []struct {
		testName               string
		systemDatabaseSecret   *v1.Secret
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func(*component.SystemMysqlOptions) *component.SystemMysqlOptions
	}{
		{"Default", nil, basicApimanager,
			func(opts *component.SystemMysqlOptions) *component.SystemMysqlOptions {
				return defaultSystemMysqlOptions(opts)
			},
		},
		{"WithoutResourceRequirements", nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &tmpFalseValue
				return apimanager
			},
			func(opts *component.SystemMysqlOptions) *component.SystemMysqlOptions {
				expecteOpts := defaultSystemMysqlOptions(opts)
				expecteOpts.ContainerResourceRequirements = v1.ResourceRequirements{}
				return expecteOpts
			},
		},
		{"SystemDBSecret", getSystemDBSecret(databaseURL, systemMysqlUsername, systemMysqlPassword), basicApimanager,
			func(opts *component.SystemMysqlOptions) *component.SystemMysqlOptions {
				expecteOpts := defaultSystemMysqlOptions(opts)
				expecteOpts.User = systemMysqlUsername
				expecteOpts.Password = systemMysqlPassword
				expecteOpts.DatabaseURL = databaseURL
				expecteOpts.DatabaseName = systemMysqlDatabaseName
				expecteOpts.RootPassword = systemMysqlRootPassword
				return expecteOpts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			objs := []runtime.Object{}
			if tc.systemDatabaseSecret != nil {
				objs = append(objs, tc.systemDatabaseSecret)
			}
			cl := fake.NewFakeClient(objs...)
			optsProvider := NewSystemMysqlOptionsProvider(tc.apimanagerFactory(), namespace, cl)
			opts, err := optsProvider.GetMysqlOptions()
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

func TestGetMysqlOptionsInvalidURL(t *testing.T) {
	cases := []struct {
		testName    string
		databaseURL string
		errSubstr   string
	}{
		// prefix must be mysql2
		{"invalidURL01", "mysql://root:password1@system-mysql/system", "'mysql2'"},
		// missing user
		{"invalidURL02", "mysql2://system-mysql/system", "authentication information"},
		// user not root
		{"invalidURL03", "mysql2://user:password1@system-mysql/system", "'root'"},
		// passwd missing
		{"invalidURL04", "mysql2://root@system-mysql/system", "secret must contain a password"},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			secret := getSystemDBSecret(tc.databaseURL, systemMysqlUsername, systemMysqlPassword)
			objs := []runtime.Object{secret}
			cl := fake.NewFakeClient(objs...)
			optsProvider := NewSystemMysqlOptionsProvider(basicApimanager(), namespace, cl)
			_, err := optsProvider.GetMysqlOptions()
			if err == nil {
				subT.Fatal("expected to fail for invalid URL")
			}
			if !strings.Contains(err.Error(), tc.errSubstr) {
				subT.Fatalf("expected error regexp: %s, got: (%v)", tc.errSubstr, err)
			}
		})
	}
}
