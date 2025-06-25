package helper

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

func MarshalObjectToYAML(object runtime.Object) ([]byte, error) {
	serializedResult, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(serializedResult)
}
