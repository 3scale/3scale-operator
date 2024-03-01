package component

import (
	"github.com/3scale/3scale-operator/pkg/common"
	k8sappsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type Evaluation struct {
}

func NewEvaluation() *Evaluation {
	return &Evaluation{}
}

func (evaluation *Evaluation) RemoveContainersResourceRequestsAndLimits(objects []common.KubernetesObject) {
	for _, obj := range objects {
		deployment, ok := obj.(*k8sappsv1.Deployment)
		if ok {
			for containerIdx := range deployment.Spec.Template.Spec.Containers {
				container := &deployment.Spec.Template.Spec.Containers[containerIdx]
				container.Resources = v1.ResourceRequirements{}
			}
		}
	}
}
