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
	systemPostgreSQLUsername     = "postgresqluser"
	systemPostgreSQLPassword     = "password1435"
	systemPostgreSQLDatabaseName = "system"
)

func defaultSystemPostgreSQLOptions(opts *component.SystemPostgreSQLOptions) *component.SystemPostgreSQLOptions {
	return &component.SystemPostgreSQLOptions{
		AppLabel:                      appLabel,
		DatabaseName:                  component.DefaultSystemPostgresqlDatabaseName(),
		User:                          component.DefaultSystemPostgresqlUser(),
		Password:                      opts.Password,
		DatabaseURL:                   component.DefaultSystemPostgresqlDatabaseURL(component.DefaultSystemPostgresqlUser(), opts.Password, component.DefaultSystemPostgresqlDatabaseName()),
		ContainerResourceRequirements: component.DefaultSystemPostgresqlResourceRequirements(),
	}
}

func TestGetSystemPostgreSQLOptionsProvider(t *testing.T) {
	falseValue := false
	databaseURL := fmt.Sprintf("postgresql://%s:%s@postgresql.example.com/%s", systemPostgreSQLUsername, systemPostgreSQLPassword, systemPostgreSQLDatabaseName)

	cases := []struct {
		testName               string
		systemDatabaseSecret   *v1.Secret
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func(*component.SystemPostgreSQLOptions) *component.SystemPostgreSQLOptions
	}{
		{"Default", nil, basicApimanager, defaultSystemPostgreSQLOptions},
		{"WithoutResourceRequirements", nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &falseValue
				return apimanager
			},
			func(opts *component.SystemPostgreSQLOptions) *component.SystemPostgreSQLOptions {
				expecteOpts := defaultSystemPostgreSQLOptions(opts)
				expecteOpts.ContainerResourceRequirements = v1.ResourceRequirements{}
				return expecteOpts
			},
		},
		{"SystemDBSecret", getSystemDBSecret(databaseURL, systemPostgreSQLUsername, systemPostgreSQLPassword),
			basicApimanager, func(opts *component.SystemPostgreSQLOptions) *component.SystemPostgreSQLOptions {
				expecteOpts := defaultSystemPostgreSQLOptions(opts)
				expecteOpts.User = systemPostgreSQLUsername
				expecteOpts.Password = systemPostgreSQLPassword
				expecteOpts.DatabaseURL = databaseURL
				expecteOpts.DatabaseName = systemPostgreSQLDatabaseName
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
			optsProvider := NewSystemPostgresqlOptionsProvider(tc.apimanagerFactory(), namespace, cl)
			opts, err := optsProvider.GetSystemPostgreSQLOptions()
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

func TestGetPostgreSQLOptionsInvalidURL(t *testing.T) {
	cases := []struct {
		testName    string
		databaseURL string
		errSubstr   string
	}{
		// prefix must be postgresql
		{"invalidURL01", "mysql://root:password1@system-postgresql/system", "'postgresql'"},
		// missing user
		{"invalidURL02", "postgresql://system-postgresql/system", "authentication information"},
		// missing user
		{"invalidURL03", "postgresql://:password1@system-postgresql/system", "secret must contain a username"},
		// missing passwd
		{"invalidURL04", "postgresql://user@system-postgresql/system", "secret must contain a password"},
		// path missing
		{"invalidURL05", "postgresql://user:password1@system-postgresql", "database name"},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			secret := getSystemDBSecret(tc.databaseURL, systemPostgreSQLUsername, systemPostgreSQLPassword)
			objs := []runtime.Object{secret}
			cl := fake.NewFakeClient(objs...)
			optsProvider := NewSystemPostgresqlOptionsProvider(basicApimanager(), namespace, cl)
			_, err := optsProvider.GetSystemPostgreSQLOptions()
			if err == nil {
				subT.Fatal("expected to fail for invalid URL")
			}
			if !strings.Contains(err.Error(), tc.errSubstr) {
				subT.Fatalf("expected error regexp: %s, got: (%v)", tc.errSubstr, err)
			}
		})
	}
}
