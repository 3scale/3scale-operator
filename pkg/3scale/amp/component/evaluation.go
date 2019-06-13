package component

import (
	"github.com/3scale/3scale-operator/pkg/apis/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	appsv1 "github.com/openshift/api/apps/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
)

type Evaluation struct {
	options []string
}

type EvaluationOptions struct {
}

func NewEvaluation(options []string) *Evaluation {
	evaluation := &Evaluation{
		options: options,
	}
	return evaluation
}

func (evaluation *Evaluation) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
}

func (evaluation *Evaluation) PostProcess(template *templatev1.Template, otherComponents []Component) {
	objects := helper.UnwrapRawExtensions(template.Objects)
	evaluation.removeContainersResourceRequestsAndLimits(objects)
}

func (evaluation *Evaluation) PostProcessObjects(objects []common.KubernetesObject) {
	evaluation.removeContainersResourceRequestsAndLimits(objects)
}

func (evaluation *Evaluation) removeContainersResourceRequestsAndLimits(objects []common.KubernetesObject) {
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
