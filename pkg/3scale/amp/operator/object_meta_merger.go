package operator

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ObjectMetaMerger struct {
	BaseReconciler
	Owner metav1.Object
}

func NewObjectMetaMerger(baseReconciler BaseReconciler, owner metav1.Object) ObjectMetaMerger {
	return ObjectMetaMerger{
		BaseReconciler: baseReconciler,
		Owner:          owner,
	}
}

func (r *ObjectMetaMerger) EnsureObjectMeta(updated, desired *metav1.ObjectMeta) (bool, error) {
	changed := false
	if updated.Name != desired.Name {
		updated.Name = desired.Name
		changed = true
	}

	if updated.Namespace != desired.Namespace {
		updated.Namespace = desired.Namespace
		changed = true
	}

	if r.mergeMapStringString(updated.Labels, desired.Labels) {
		changed = true
	}

	if r.mergeMapStringString(updated.Annotations, desired.Annotations) {
		changed = true

	}

	var err error
	if r.Owner != nil {
		originalSize := len(updated.OwnerReferences)
		err = controllerutil.SetControllerReference(r.Owner, updated.GetObjectMeta(), r.scheme)
		newSize := len(updated.OwnerReferences)
		if originalSize != newSize {
			changed = true
		}
	}

	return changed, err
}

func (r *ObjectMetaMerger) mergeMapStringString(updated, desired map[string]string) bool {
	changed := false

	for k, v := range desired {
		if updatedVal, ok := updated[k]; !ok || v != updatedVal {
			updated[k] = v
			changed = true
		}
	}

	return changed
}

func (r *ObjectMetaMerger) EnsureAnnotations(updated, desired map[string]string) bool {
	return r.mergeMapStringString(updated, desired)
}

func (r *ObjectMetaMerger) EnsureLabels(updated, desired map[string]string) bool {
	return r.mergeMapStringString(updated, desired)
}
