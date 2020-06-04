package helper

import (
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
)

func MarshalObjectToYAML(object runtime.Object) ([]byte, error) {
	serializedResult, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(serializedResult)
}
