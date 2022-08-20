package common

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const (
	DeleteTagAnnotation                  = "apps.3scale.net/delete"
	DeletePropagationPolicyTagAnnotation = "apps.3scale.net/delete-propagation-policy"
)

type KubernetesObject interface {
	metav1.Object
	runtime.Object
}

func ObjectInfo(obj KubernetesObject) string {
	return fmt.Sprintf("%s/%s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
}

func ObjectKey(obj KubernetesObject) types.NamespacedName {
	return types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}
}

func TagObjectToDelete(obj KubernetesObject) {
	// Add custom annotation
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
		obj.SetAnnotations(annotations)
	}
	annotations[DeleteTagAnnotation] = "true"
}

func TagToObjectDeleteWithPropagationPolicy(obj KubernetesObject, deletionPropagationPolicy metav1.DeletionPropagation) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
		obj.SetAnnotations(annotations)

	}
	annotations[DeleteTagAnnotation] = "true"
	annotations[DeletePropagationPolicyTagAnnotation] = string(deletionPropagationPolicy)
}

func IsObjectTaggedToDelete(obj KubernetesObject) bool {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return false
	}

	annotation, ok := annotations[DeleteTagAnnotation]
	return ok && annotation == "true"
}

func GetDeletePropagationPolicyAnnotation(obj KubernetesObject) *metav1.DeletionPropagation {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return nil
	}
	annotation, ok := annotations[DeletePropagationPolicyTagAnnotation]
	var res *metav1.DeletionPropagation
	if ok {
		tmp := metav1.DeletionPropagation(annotation)
		res = &tmp
	}
	return res
}
