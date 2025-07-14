package reconcilers

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/3scale/3scale-operator/pkg/common"
	v1 "k8s.io/api/core/v1"
)

func ServiceAccountImagePullPolicyMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*v1.ServiceAccount)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.ServiceAccount", existingObj)
	}
	desired, ok := desiredObj.(*v1.ServiceAccount)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.ServiceAccount", desiredObj)
	}

	updated := false

	// Copy received ServiceAccounts due to we are going to
	// sort (modify) its ImagePullSecrets
	existingCopy := existing.DeepCopy()
	desiredCopy := desired.DeepCopy()

	existingCopyImagePullSecrets := existingCopy.ImagePullSecrets
	sort.Slice(existingCopyImagePullSecrets, func(i, j int) bool {
		return existingCopyImagePullSecrets[i].Name < existingCopyImagePullSecrets[j].Name
	})

	// Some secrets are automatically added as ImagePullPolicy secrets by K8s.
	// We try to incorporate them by detecting their usual pattern prefix.
	// This has the disadvantage that users themselves cannot create secrets with
	// those prefixes but allows us to compare
	for _, obj := range existingCopy.ImagePullSecrets {
		if strings.HasPrefix(obj.Name, fmt.Sprintf("%s-dockercfg-", existing.Name)) ||
			strings.HasPrefix(obj.Name, fmt.Sprintf("%s-token-", existing.Name)) {
			if arrayLocalObjectReferenceFind(obj.Name, desiredCopy.ImagePullSecrets) == -1 {
				desiredCopy.ImagePullSecrets = append(desiredCopy.ImagePullSecrets, obj)
			}
		}
	}

	desiredCopyImagePullSecrets := desiredCopy.ImagePullSecrets
	sort.Slice(desiredCopyImagePullSecrets, func(i, j int) bool {
		return desiredCopyImagePullSecrets[i].Name < desiredCopyImagePullSecrets[j].Name
	})

	// SET behavior
	if !reflect.DeepEqual(existingCopyImagePullSecrets, desiredCopyImagePullSecrets) {
		existing.ImagePullSecrets = desiredCopyImagePullSecrets
		updated = true
	}

	return updated, nil
}

func arrayLocalObjectReferenceFind(name string, arr []v1.LocalObjectReference) int {
	for i, obj := range arr {
		if obj.Name == name {
			return i
		}
	}

	return -1
}
