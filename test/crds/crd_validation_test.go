package test

import (
	"fmt"
	apps "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	capabilities "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1alpha1"
	"github.com/RHsyseng/operator-utils/pkg/validation"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSampleCustomResources(t *testing.T) {
	root := "../../deploy/crds"
	crdCrMap := map[string]string{
		"apps_v1alpha1_apimanager_crd.yaml": "apps_v1alpha1_apimanager_cr",
		"capabilities_v1alpha1_api_crd.yaml": "capabilities_v1alpha1_api_cr",
		"capabilities_v1alpha1_binding_crd.yaml": "capabilities_v1alpha1_binding_cr",
		"capabilities_v1alpha1_limit_crd.yaml": "capabilities_v1alpha1_limit_cr",
		"capabilities_v1alpha1_mappingrule_crd.yaml": "capabilities_v1alpha1_mappingrule_cr",
		"capabilities_v1alpha1_metric_crd.yaml": "capabilities_v1alpha1_metric_cr",
		"capabilities_v1alpha1_plan_crd.yaml": "capabilities_v1alpha1_plan_cr",
		"capabilities_v1alpha1_tenant_crd.yaml": "capabilities_v1alpha1_tenant_cr",
	}
	for crd, prefix := range crdCrMap {
		validateCustomResource(t, root, crd, prefix)
	}
}

func validateCustomResource(t *testing.T, root string, crd string, prefix string) {
	schema := getSchema(t, fmt.Sprintf("%s/%s", root, crd))
	assert.NotNil(t, schema)
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), "crd.yaml") {
			//Ignore CRD
			return nil
		}
		if strings.HasPrefix(info.Name(), prefix) {
			bytes, err := ioutil.ReadFile(path)
			assert.NoError(t, err, "Error reading CR yaml from %v", path)
			var input map[string]interface{}
			assert.NoError(t, yaml.Unmarshal(bytes, &input))
			assert.NoError(t, schema.Validate(input), "File %v does not validate against the %s CRD schema", info.Name(), crd)
		}
		return nil
	}
	err := filepath.Walk(root, walkFunc)
	assert.NoError(t, err, "Error reading CR yaml files from ", root)
}

func TestCompleteCRD(t *testing.T) {
	root := "../../deploy/crds"
	crdStructMap := map[string]interface{}{
		"apps_v1alpha1_apimanager_crd.yaml": &apps.APIManager{},
		"capabilities_v1alpha1_api_crd.yaml": &capabilities.API{},
		"capabilities_v1alpha1_binding_crd.yaml": &capabilities.Binding{},
		"capabilities_v1alpha1_limit_crd.yaml": &capabilities.Limit{},
		"capabilities_v1alpha1_mappingrule_crd.yaml": &capabilities.MappingRule{},
		"capabilities_v1alpha1_metric_crd.yaml": &capabilities.Metric{},
		"capabilities_v1alpha1_plan_crd.yaml": &capabilities.Plan{},
		"capabilities_v1alpha1_tenant_crd.yaml": &capabilities.Tenant{},
	}
	for crd, obj := range crdStructMap {
		schema := getSchema(t, fmt.Sprintf("%s/%s", root, crd))
		missingEntries := schema.GetMissingEntries(obj)
		for _, missing := range missingEntries {
			assert.Fail(t, "Discrepancy between %s CRD and Struct", "Missing or incorrect schema validation at %v, expected type %v", crd, missing.Path, missing.Type)
		}
	}
}

func getSchema(t *testing.T, crd string) validation.Schema {
	bytes, err := ioutil.ReadFile(crd)
	assert.NoError(t, err, "Error reading CRD yaml from %v", crd)
	schema, err := validation.New(bytes)
	assert.NoError(t, err)
	return schema
}
