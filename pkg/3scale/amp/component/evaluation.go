package component

import (
	k8sappsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Evaluation struct{}

func NewEvaluation() *Evaluation {
	return &Evaluation{}
}

func (evaluation *Evaluation) RemoveContainersResourceRequestsAndLimits(objects []client.Object) {
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
