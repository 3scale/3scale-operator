package resourcemerge

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func EnsureObjectMeta(updated, desired *metav1.ObjectMeta, owner metav1.Object, scheme *runtime.Scheme) (bool, error) {
	changed := false
	if updated.Name != desired.Name {
		updated.Name = desired.Name
		changed = true
	}

	if updated.Namespace != desired.Namespace {
		updated.Namespace = desired.Namespace
		changed = true
	}

	if mergeMapStringString(updated.Labels, desired.Labels) {
		changed = true
	}

	if mergeMapStringString(updated.Annotations, desired.Annotations) {
		changed = true

	}

	var err error
	originalSize := len(updated.OwnerReferences)
	err = controllerutil.SetControllerReference(owner, updated.GetObjectMeta(), scheme)
	newSize := len(updated.OwnerReferences)
	if originalSize != newSize {
		changed = true
	}

	return changed, err
}

func mergeMapStringString(updated, desired map[string]string) bool {
	changed := false

	for k, v := range desired {
		if updatedVal, ok := updated[k]; !ok || v != updatedVal {
			updated[k] = v
			changed = true
		}
	}

	return changed
}

func EnsureAnnotations(updated, desired map[string]string) bool {
	return mergeMapStringString(updated, desired)
}

func EnsureLabels(updated, desired map[string]string) bool {
	return mergeMapStringString(updated, desired)
}
