package helper

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"

	"k8s.io/apimachinery/pkg/util/validation"
)

type ComponentType string

const (
	ApplicationType    ComponentType = "application"
	InfrastructureType ComponentType = "infrastructure"
)

func MeteringLabels(componentName, componentVersion string, componentType ComponentType) map[string]string {
	labels := map[string]string{
		"com.redhat.product-name":      "3scale",
		"com.redhat.component-name":    componentName,
		"com.redhat.product-version":   product.ThreescaleRelease,
		"com.redhat.component-type":    string(componentType),
		"com.redhat.component-version": componentVersion,
	}

	if len(validation.IsValidLabelValue(componentVersion)) > 0 {
		labels["com.redhat.component-version"] = ""
	}

	return labels
}
