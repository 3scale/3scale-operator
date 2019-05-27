package component

import (
	"github.com/3scale/3scale-operator/pkg/common"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type Evaluation struct {
}

func NewEvaluation() *Evaluation {
	return &Evaluation{}
}

func (evaluation *Evaluation) RemoveContainersResourceRequestsAndLimits(objects []common.KubernetesObject) {
	for _, obj := range objects {
		dc, ok := obj.(*appsv1.DeploymentConfig)
		if ok {
			for containerIdx := range dc.Spec.Template.Spec.Containers {
				container := &dc.Spec.Template.Spec.Containers[containerIdx]
				container.Resources = v1.ResourceRequirements{}
			}
		}
	}
}
