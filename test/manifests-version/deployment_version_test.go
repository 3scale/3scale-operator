package test

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/version"

	"github.com/ghodss/yaml"
	appsv1 "k8s.io/api/apps/v1"
)

func TestDeploymentVersions(t *testing.T) {
	root := "../../deploy"
	path := path.Join(root, "operator.yaml")
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	deployment := appsv1.Deployment{}

	err = yaml.Unmarshal(bytes, &deployment)
	if err != nil {
		t.Fatal(err)
	}

	if deployment.Spec.Template.Labels["com.redhat.component-version"] != version.Version {
		t.Errorf("com.redhat.component-version differ: expected: %s; found: %s", version.Version, deployment.Spec.Template.Labels["com.redhat.component-version"])
	}

	if deployment.Spec.Template.Labels["com.redhat.product-version"] != product.ThreescaleRelease {
		t.Errorf("com.redhat.product-version differ: expected: %s; found: %s", product.ThreescaleRelease, deployment.Spec.Template.Labels["com.redhat.product-version"])
	}
}
