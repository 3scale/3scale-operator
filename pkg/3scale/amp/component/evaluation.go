package component

import (
	appsv1 "github.com/openshift/api/apps/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	evaluation.removeContainersResourceRequestsAndLimits(template.Objects)
}

func (evaluation *Evaluation) PostProcessObjects(objects []runtime.RawExtension) {
	evaluation.removeContainersResourceRequestsAndLimits(objects)
}

func (evaluation *Evaluation) removeContainersResourceRequestsAndLimits(objects []runtime.RawExtension) {
	for _, rawExtension := range objects {
		obj := rawExtension.Object
		dc, ok := obj.(*appsv1.DeploymentConfig)
		if ok {
			for containerIdx := range dc.Spec.Template.Spec.Containers {
				container := &dc.Spec.Template.Spec.Containers[containerIdx]
				container.Resources = v1.ResourceRequirements{}
			}
		}
	}
}
