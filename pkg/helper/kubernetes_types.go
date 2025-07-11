package helper

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DeleteTagAnnotation                  = "apps.3scale.net/delete"
	DeletePropagationPolicyTagAnnotation = "apps.3scale.net/delete-propagation-policy"
)

func ObjectInfo(obj client.Object) string {
	return fmt.Sprintf("%s/%s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
}

func ObjectKey(obj client.Object) types.NamespacedName {
	return types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}
}

func TagObjectToDelete(obj client.Object) {
	// Add custom annotation
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
		obj.SetAnnotations(annotations)
	}
	annotations[DeleteTagAnnotation] = "true"
}

func TagToObjectDeleteWithPropagationPolicy(obj client.Object, deletionPropagationPolicy metav1.DeletionPropagation) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
		obj.SetAnnotations(annotations)
	}
	annotations[DeleteTagAnnotation] = "true"
	annotations[DeletePropagationPolicyTagAnnotation] = string(deletionPropagationPolicy)
}

func IsObjectTaggedToDelete(obj client.Object) bool {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return false
	}

	annotation, ok := annotations[DeleteTagAnnotation]
	return ok && annotation == "true"
}

func GetDeletePropagationPolicyAnnotation(obj client.Object) *metav1.DeletionPropagation {
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
