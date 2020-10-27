package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	apps "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	capabilitiesv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1alpha1"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
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
)

func TestSampleCustomResources(t *testing.T) {
	root := "../../deploy/crds"
	crdCrMap := map[string]string{
		"apps.3scale.net_apimanagers_crd.yaml":        "apps.3scale.net_v1alpha1_apimanager_cr",
		"apps.3scale.net_apimanagerbackups_crd.yaml":  "apps.3scale.net_v1alpha1_apimanagerbackup_cr.yaml",
		"apps.3scale.net_apimanagerrestores_crd.yaml": "apps.3scale.net_v1alpha1_apimanagerrestore_cr.yaml",
		"capabilities.3scale.net_tenants_crd.yaml":    "capabilities.3scale.net_v1alpha1_tenant_cr",
		"capabilities.3scale.net_backends_crd.yaml":   "capabilities.3scale.net_v1beta1_backend_cr",
		"capabilities.3scale.net_products_crd.yaml":   "capabilities.3scale.net_v1beta1_product_cr",
		"capabilities.3scale.net_openapis_crd.yaml":   "capabilities.3scale.net_v1beta1_openapi_cr",
	}
	for crd, prefix := range crdCrMap {
		validateCustomResources(t, root, crd, prefix)
	}
}

func validateCustomResources(t *testing.T, root string, crd string, prefix string) {
	schema := getSchema(t, fmt.Sprintf("%s/%s", root, crd))
	assert.NotNil(t, schema)
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), "crd.yaml") {
			//Ignore CRD
			return nil
		}
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
	err := filepath.Walk(root, walkFunc)
	assert.NoError(t, err, "Error reading CR yaml files from ", root)
}

func TestCompleteCRD(t *testing.T) {
	root := "../../deploy/crds"
	crdStructMap := map[string]interface{}{
		"apps.3scale.net_apimanagers_crd.yaml":        &apps.APIManager{},
		"apps.3scale.net_apimanagerbackups_crd.yaml":  &apps.APIManagerBackup{},
		"apps.3scale.net_apimanagerrestores_crd.yaml": &apps.APIManagerRestore{},
		"capabilities.3scale.net_tenants_crd.yaml":    &capabilitiesv1alpha1.Tenant{},
		"capabilities.3scale.net_backends_crd.yaml":   &capabilitiesv1beta1.Backend{},
		"capabilities.3scale.net_products_crd.yaml":   &capabilitiesv1beta1.Product{},
		"capabilities.3scale.net_openapis_crd.yaml":   &capabilitiesv1beta1.Openapi{},
	}

	pathOmissions := []string{
		backupDestinationPVCResourceRequestsPath,
		startTimePath,
		completionTimePath,
		lastTransitionTimePath,
		systemSharedPVCResourceRequestsPath,
		systemMySQLPVCResourceRequestsPath,
		systemPostgreSQLPVCResourceRequestsPath,
	}

	for crd, obj := range crdStructMap {
		t.Run(crd, func(subT *testing.T) {
			schema := getSchema(subT, fmt.Sprintf("%s/%s", root, crd))
			missingEntries := schema.GetMissingEntries(obj)
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

func missingFieldPathInPathOmissions(path string, omissions []string) bool {
	for _, omit := range omissions {
		if strings.HasPrefix(path, omit) {
			return true
		}
	}
	return false
}
