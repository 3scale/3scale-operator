package operator

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/helper"
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

func testSystemPostgreSQLCommonLabels() map[string]string {
	return map[string]string{
		"app":                  appLabel,
		"threescale_component": "system",
	}
}

func testSystemPostgreSQLDeploymentLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "postgresql",
	}
}

func testSystemPostgreSQLPodTemplateLabels() map[string]string {
	labels := map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "postgresql",
		"deploymentConfig":             "system-postgresql",
	}
	addExpectedMeteringLabels(labels, "system-postgresql", helper.ApplicationType)

	return labels
}

func testSystemPostgreSQLAffinity() *v1.Affinity {
	return getTestAffinity("system-postgresql")
}

func testSystemPostgreSQLTolerations() []v1.Toleration {
	return getTestTolerations("system-postgresql")
}

func testSystemPostgreSQLCustomResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("111m"),
			v1.ResourceMemory: resource.MustParse("222Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("333m"),
			v1.ResourceMemory: resource.MustParse("444Mi"),
		},
	}
}

func defaultSystemPostgreSQLOptions(opts *component.SystemPostgreSQLOptions) *component.SystemPostgreSQLOptions {
	return &component.SystemPostgreSQLOptions{
		ImageTag:                      product.ThreescaleRelease,
		DatabaseName:                  component.DefaultSystemPostgresqlDatabaseName(),
		User:                          component.DefaultSystemPostgresqlUser(),
		Password:                      opts.Password,
		DatabaseURL:                   component.DefaultSystemPostgresqlDatabaseURL(component.DefaultSystemPostgresqlUser(), opts.Password, component.DefaultSystemPostgresqlDatabaseName()),
		ContainerResourceRequirements: component.DefaultSystemPostgresqlResourceRequirements(),
		CommonLabels:                  testSystemPostgreSQLCommonLabels(),
		DeploymentLabels:              testSystemPostgreSQLDeploymentLabels(),
		PodTemplateLabels:             testSystemPostgreSQLPodTemplateLabels(),
		PVCStorageRequests:            component.DefaultSystemPostgresqlStorageResources(),
	}
}

func TestGetSystemPostgreSQLOptionsProvider(t *testing.T) {
	falseValue := false
	databaseURL := fmt.Sprintf("postgresql://%s:%s@postgresql.example.com/%s", systemPostgreSQLUsername, systemPostgreSQLPassword, systemPostgreSQLDatabaseName)
	customStorageClass := "custompostgresqlstorageclass"

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
		{"PVCSpecSet", nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					DatabaseSpec: &appsv1alpha1.SystemDatabaseSpec{
						PostgreSQL: &appsv1alpha1.SystemPostgreSQLSpec{
							PersistentVolumeClaimSpec: &appsv1alpha1.SystemPostgreSQLPVCSpec{},
						},
					},
				}
				return apimanager
			},
			func(opts *component.SystemPostgreSQLOptions) *component.SystemPostgreSQLOptions {
				expecteOpts := defaultSystemPostgreSQLOptions(opts)
				return expecteOpts
			},
		},
		{"PVCSettingsSet", nil,
			func() *appsv1alpha1.APIManager {
				tmpVolumeName := "myvolume"
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					DatabaseSpec: &appsv1alpha1.SystemDatabaseSpec{
						PostgreSQL: &appsv1alpha1.SystemPostgreSQLSpec{
							PersistentVolumeClaimSpec: &appsv1alpha1.SystemPostgreSQLPVCSpec{
								StorageClassName: &customStorageClass,
								Resources: &appsv1alpha1.PersistentVolumeClaimResources{
									Requests: resource.MustParse("456Mi"),
								},
								VolumeName: &tmpVolumeName,
							},
						},
					},
				}
				return apimanager
			},
			func(opts *component.SystemPostgreSQLOptions) *component.SystemPostgreSQLOptions {
				tmpVolumeName := "myvolume"
				expecteOpts := defaultSystemPostgreSQLOptions(opts)
				expecteOpts.PVCStorageClass = &customStorageClass
				expecteOpts.PVCVolumeName = &tmpVolumeName
				expecteOpts.PVCStorageRequests = resource.MustParse("456Mi")
				return expecteOpts
			},
		},
		{"WithAffinity", nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					DatabaseSpec: &appsv1alpha1.SystemDatabaseSpec{
						PostgreSQL: &appsv1alpha1.SystemPostgreSQLSpec{
							Affinity: testSystemPostgreSQLAffinity(),
						},
					},
				}
				return apimanager
			},
			func(opts *component.SystemPostgreSQLOptions) *component.SystemPostgreSQLOptions {
				expecteOpts := defaultSystemPostgreSQLOptions(opts)
				expecteOpts.Affinity = testSystemPostgreSQLAffinity()
				return expecteOpts
			},
		},
		{"WithTolerations", nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					DatabaseSpec: &appsv1alpha1.SystemDatabaseSpec{
						PostgreSQL: &appsv1alpha1.SystemPostgreSQLSpec{
							Tolerations: testSystemPostgreSQLTolerations(),
						},
					},
				}
				return apimanager
			},
			func(opts *component.SystemPostgreSQLOptions) *component.SystemPostgreSQLOptions {
				expecteOpts := defaultSystemPostgreSQLOptions(opts)
				expecteOpts.Tolerations = testSystemPostgreSQLTolerations()
				return expecteOpts
			},
		},
		{"WithSystemPostgreSQLCustomResourceRequirements", nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					DatabaseSpec: &appsv1alpha1.SystemDatabaseSpec{
						PostgreSQL: &appsv1alpha1.SystemPostgreSQLSpec{
							Resources: testSystemPostgreSQLCustomResourceRequirements(),
						},
					},
				}
				return apimanager
			},
			func(opts *component.SystemPostgreSQLOptions) *component.SystemPostgreSQLOptions {
				expecteOpts := defaultSystemPostgreSQLOptions(opts)
				expecteOpts.ContainerResourceRequirements = *testSystemPostgreSQLCustomResourceRequirements()
				return expecteOpts
			},
		},
		{"WithPostgreSQLCustomResourceRequirementsAndGlobalResourceRequirementsDisabled", nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &falseValue
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					DatabaseSpec: &appsv1alpha1.SystemDatabaseSpec{
						PostgreSQL: &appsv1alpha1.SystemPostgreSQLSpec{
							Resources: testSystemPostgreSQLCustomResourceRequirements(),
						},
					},
				}
				return apimanager
			},
			func(opts *component.SystemPostgreSQLOptions) *component.SystemPostgreSQLOptions {
				expecteOpts := defaultSystemPostgreSQLOptions(opts)
				expecteOpts.ContainerResourceRequirements = *testSystemPostgreSQLCustomResourceRequirements()
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
