package operator

import (
	"context"
	"fmt"
	"sort"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type PersistentVolumeClaimReconciler struct {
	BaseReconciler
	ObjectMetaMerger ObjectMetaMerger
}

func NewPersistentVolumeClaimReconciler(baseReconciler BaseReconciler, objectMetaMerger ObjectMetaMerger) PersistentVolumeClaimReconciler {
	return PersistentVolumeClaimReconciler{
		BaseReconciler:   baseReconciler,
		ObjectMetaMerger: objectMetaMerger,
	}
}

func (r *PersistentVolumeClaimReconciler) Reconcile(desiredPVC *v1.PersistentVolumeClaim) error {
	objectInfo := ObjectInfo(desiredPVC)
	existingPVC := &v1.PersistentVolumeClaim{}
	err := r.Client().Get(context.TODO(), types.NamespacedName{Name: desiredPVC.Name, Namespace: desiredPVC.Namespace}, existingPVC)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.Client().Create(context.TODO(), desiredPVC)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			r.Logger().Info(fmt.Sprintf("Created object %s", objectInfo))
			return nil
		}
		return err
	}

	needsUpdate, err := r.ensurePersistentVolumeClaim(existingPVC, desiredPVC)
	if err != nil {
		return err
	}

	if needsUpdate {
		r.Logger().Info(fmt.Sprintf("Updating PersistentVolumeClaim %s", objectInfo))
		err := r.Client().Update(context.TODO(), existingPVC)
		if err != nil {
			r.Logger().Error(err, fmt.Sprintf("error updating PersistentVolumeClaim %s", objectInfo))
			return err
		}
	}

	return nil
}

func (r *PersistentVolumeClaimReconciler) ensurePersistentVolumeClaim(updated, desired *v1.PersistentVolumeClaim) (bool, error) {
	changed := false

	objectMetaChanged, err := r.ObjectMetaMerger.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta)
	if err != nil {
		return false, err
	}
	if objectMetaChanged {
		changed = true
	}

	r.ensureResourceRequirements(&updated.Spec.Resources, &desired.Spec.Resources)
	r.ensureAccessModes(updated.Spec.AccessModes, desired.Spec.AccessModes)
	return changed, nil
}

// TODO this is also used for DeploymentConfigReconciler so we should refactor
// it in some way
func (r *PersistentVolumeClaimReconciler) ensureResourceRequirements(updated, desired *v1.ResourceRequirements) {
	r.ensureResourceList(updated.Limits, desired.Limits)
	r.ensureResourceList(updated.Requests, desired.Requests)
}

// This method assigns the value of desired to updated when the resource quantities
// are not nil in both sides and are equal. The reason for this is
// because although they are equal internally are stored differently (the units
// might be expressed in a different way) and executing DeepEqual returns
// different even though they are "logically" equal
func (r *PersistentVolumeClaimReconciler) ensureResourceList(updated, desired v1.ResourceList) {
	if !desired.Cpu().IsZero() && !updated.Cpu().IsZero() &&
		desired.Cpu().Cmp(*updated.Cpu()) == 0 {
		desired[v1.ResourceCPU] = *updated.Cpu()
	}
	if !desired.Memory().IsZero() && !updated.Memory().IsZero() &&
		desired.Memory().Cmp(*updated.Memory()) == 0 {
		desired[v1.ResourceMemory] = *updated.Memory()
	}
	if !desired.Pods().IsZero() && !updated.Pods().IsZero() &&
		desired.Pods().Cmp(*updated.Pods()) == 0 {
		desired[v1.ResourcePods] = *updated.Pods()
	}
	if !desired.StorageEphemeral().IsZero() && !updated.StorageEphemeral().IsZero() &&
		desired.StorageEphemeral().Cmp(*updated.StorageEphemeral()) == 0 {
		desired[v1.ResourceEphemeralStorage] = *updated.StorageEphemeral()
	}
}

func (r *PersistentVolumeClaimReconciler) ensureAccessModes(updated, desired []v1.PersistentVolumeAccessMode) {
	sort.Slice(updated, func(i, j int) bool { return updated[i] < updated[j] })
	sort.Slice(desired, func(i, j int) bool { return desired[i] < desired[j] })
}
