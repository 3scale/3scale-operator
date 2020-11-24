package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	apps "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	capabilitiesv1alpha1 "github.com/3scale/3scale-operator/apis/capabilities/v1alpha1"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/RHsyseng/operator-utils/pkg/validation"
	"github.com/ghodss/yaml"

	"github.com/stretchr/testify/assert"
)

// Missing fields path omissions
const (
	backupDestinationPVCResourceRequestsPath = "/spec/backupDestination/persistentVolumeClaim/resources/requests"
	startTimePath                            = "/status/startTime"
	completionTimePath                       = "/status/completionTime"
	lastTransitionTimePath                   = "/status/conditions/lastTransitionTime"
	systemSharedPVCResourceRequestsPath      = "/spec/system/fileStorage/persistentVolumeClaim/resources/requests"
	systemMySQLPVCResourceRequestsPath       = "/spec/system/database/mysql/persistentVolumeClaim/resources/requests"
	systemPostgreSQLPVCResourceRequestsPath  = "/spec/system/database/postgresql/persistentVolumeClaim/resources/requests"
	productPoliciesConfigurationPath         = "/spec/policies/configuration"
)

type testCRInfo struct {
	crPrefix   string
	apiVersion string
}

func TestSampleCustomResources(t *testing.T) {
	schemaRoot := "../../bundle/manifests"
	samplesRoot := "../../config/samples"
	// Map of CRD:CR_sample_prefix
	crdCrMap := map[string]testCRInfo{
		"apps.3scale.net_apimanagers.yaml": testCRInfo{
			crPrefix:   "apps_v1alpha1_apimanager_",
			apiVersion: apps.GroupVersion.Version,
		},
		"apps.3scale.net_apimanagerbackups.yaml": testCRInfo{
			crPrefix:   "apps_v1alpha1_apimanagerbackup.yaml",
			apiVersion: apps.GroupVersion.Version,
		},
		"apps.3scale.net_apimanagerrestores.yaml": testCRInfo{
			crPrefix:   "apps_v1alpha1_apimanagerrestore.yaml",
			apiVersion: apps.GroupVersion.Version,
		},
		"capabilities.3scale.net_tenants.yaml": testCRInfo{
			crPrefix:   "capabilities_v1alpha1_tenant",
			apiVersion: capabilitiesv1alpha1.GroupVersion.Version,
		},
		"capabilities.3scale.net_backends.yaml": testCRInfo{
			crPrefix:   "capabilities_v1beta1_backend",
			apiVersion: capabilitiesv1beta1.GroupVersion.Version,
		},
		"capabilities.3scale.net_products.yaml": testCRInfo{
			crPrefix:   "capabilities_v1beta1_product",
			apiVersion: capabilitiesv1beta1.GroupVersion.Version,
		},
		"capabilities.3scale.net_openapis.yaml": testCRInfo{
			crPrefix:   "capabilities_v1beta1_openapi",
			apiVersion: capabilitiesv1beta1.GroupVersion.Version,
		},
	}

	for crd, elem := range crdCrMap {
		validateCustomResources(t, schemaRoot, samplesRoot, crd, elem.crPrefix, elem.apiVersion)
	}
}

func validateCustomResources(t *testing.T, schemaRoot, samplesRoot, crd, prefix string, version string) {
	schema := getSchemaVersioned(t, fmt.Sprintf("%s/%s", schemaRoot, crd), version)
	assert.NotNil(t, schema)
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), prefix) {
			t.Run(info.Name(), func(subT *testing.T) {
				bytes, err := ioutil.ReadFile(path)
				assert.NoError(subT, err, "Error reading CR yaml from %v", path)
				var input map[string]interface{}
				assert.NoError(subT, yaml.Unmarshal(bytes, &input))
				assert.NoError(subT, schema.Validate(input), "File %v does not validate against the %s CRD schema", info.Name(), crd)
			})
		}
		return nil
	}
	err := filepath.Walk(samplesRoot, walkFunc)
	assert.NoError(t, err, "Error reading CR yaml files from ", samplesRoot)
}

type testCRDInfo struct {
	obj        interface{}
	apiVersion string
}

func TestCompleteCRD(t *testing.T) {
	root := "../../bundle/manifests"
	crdStructMap := map[string]testCRDInfo{
		"apps.3scale.net_apimanagers.yaml": testCRDInfo{
			obj:        &apps.APIManager{},
			apiVersion: apps.GroupVersion.Version,
		},
		"apps.3scale.net_apimanagerbackups.yaml": testCRDInfo{
			obj:        &apps.APIManagerBackup{},
			apiVersion: apps.GroupVersion.Version,
		},
		"apps.3scale.net_apimanagerrestores.yaml": testCRDInfo{
			obj:        &apps.APIManagerRestore{},
			apiVersion: apps.GroupVersion.Version,
		},
		"capabilities.3scale.net_tenants.yaml": testCRDInfo{
			obj:        &capabilitiesv1alpha1.Tenant{},
			apiVersion: capabilitiesv1alpha1.GroupVersion.Version,
		},
		"capabilities.3scale.net_backends.yaml": testCRDInfo{
			obj:        &capabilitiesv1beta1.Backend{},
			apiVersion: capabilitiesv1beta1.GroupVersion.Version,
		},
		"capabilities.3scale.net_products.yaml": testCRDInfo{
			obj:        &capabilitiesv1beta1.Product{},
			apiVersion: capabilitiesv1beta1.GroupVersion.Version,
		},
		"capabilities.3scale.net_openapis.yaml": testCRDInfo{
			obj:        &capabilitiesv1beta1.OpenAPI{},
			apiVersion: capabilitiesv1beta1.GroupVersion.Version,
		},
	}

	pathOmissions := []string{
		backupDestinationPVCResourceRequestsPath,
		startTimePath,
		completionTimePath,
		lastTransitionTimePath,
		systemSharedPVCResourceRequestsPath,
		systemMySQLPVCResourceRequestsPath,
		systemPostgreSQLPVCResourceRequestsPath,
		productPoliciesConfigurationPath,
	}

	for crd, elem := range crdStructMap {
		t.Run(crd, func(subT *testing.T) {
			schema := getSchemaVersioned(subT, fmt.Sprintf("%s/%s", root, crd), elem.apiVersion)
			missingEntries := schema.GetMissingEntries(elem.obj)
			for _, missing := range missingEntries {

				if missingFieldPathInPathOmissions(missing.Path, pathOmissions) {
					continue
				}
				assert.Fail(subT, "Discrepancy between CRD and Struct", "CRD: %s: Missing or incorrect schema validation at %s, expected type %s", crd, missing.Path, missing.Type)
			}
		})
	}
}

func getSchema(t *testing.T, crd string) validation.Schema {
	bytes, err := ioutil.ReadFile(crd)
	assert.NoError(t, err, "Error reading CRD yaml from %v", crd)
	schema, err := validation.New(bytes)
	assert.NoError(t, err)
	return schema
}

func getSchemaVersioned(t *testing.T, crd string, version string) validation.Schema {
	bytes, err := ioutil.ReadFile(crd)
	assert.NoError(t, err, "Error reading CRD yaml from %v", crd)
	schema, err := validation.NewVersioned(bytes, version)
	assert.NoError(t, err)
	return schema
}

func missingFieldPathInPathOmissions(path string, omissions []string) bool {
	for _, omit := range omissions {
		if strings.HasPrefix(path, omit) {
			return true
		}
	}
	return false
}
