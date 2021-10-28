package test

import (
	"bytes"
	"io/ioutil"
	"path"
	"testing"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"

	appsv1 "k8s.io/api/apps/v1"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

func TestDeploymentVersions(t *testing.T) {
	root := "../../config/manager/"
	path := path.Join(root, "manager.yaml")
	yamlBytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	bytesReader := ioutil.NopCloser(bytes.NewReader(yamlBytes))
	yamlDocumentDecoder := utilyaml.NewDocumentDecoder(bytesReader)

	// Read and discard Namespace object from the yaml file
	res := make([]byte, len(yamlBytes))
	_, err = yamlDocumentDecoder.Read(res)
	if err != nil {
		t.Fatal(err)
	}

	// Read the Deployment object from the yaml file
	n, err := yamlDocumentDecoder.Read(res)
	if err != nil {
		t.Fatal(err)
	}

	// Copy the Deployment object bytes length
	deploymentBytes := make([]byte, n)
	copy(deploymentBytes, res[:n])

	// Decode the Deployment object
	deployment := appsv1.Deployment{}
	fd := bytes.NewReader(deploymentBytes)
	yamlDecoder := utilyaml.NewYAMLOrJSONDecoder(fd, fd.Len())
	err = yamlDecoder.Decode(&deployment)
	if err != nil {
		t.Fatal(err)
	}

	if deployment.Kind != "Deployment" {
		t.Errorf("Parsed object is not a Deployment object")
	}

	if deployment.Spec.Template.Labels["rht.comp_ver"] != product.ThreescaleRelease {
		t.Errorf("com.redhat.product-version differ: expected: %s; found: %s", product.ThreescaleRelease, deployment.Spec.Template.Labels["com.redhat.product-version"])
	}
}
