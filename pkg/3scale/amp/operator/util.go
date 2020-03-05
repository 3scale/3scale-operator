package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/common"
)

const (
	DeleteTagAnnotation = "apps.3scale.net/delete"
)

func ObjectInfo(obj common.KubernetesObject) string {
	return fmt.Sprintf("%s/%s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
}

func TagObjectToDelete(obj common.KubernetesObject) {
	// Add custom annotation
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
		obj.SetAnnotations(annotations)
	}
	annotations[DeleteTagAnnotation] = "true"
}

func IsObjectTaggedTorDelete(obj common.KubernetesObject) bool {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return false
	}

	annotation, ok := annotations[DeleteTagAnnotation]
	return ok && annotation == "true"
}
