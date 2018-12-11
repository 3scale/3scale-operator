package component

import (
	appsv1 "github.com/openshift/api/apps/v1"
	templatev1 "github.com/openshift/api/template/v1"
	"k8s.io/api/core/v1"
)

type Evaluation struct {
	options []string
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
	evaluation.removeContainersResourceRequestsAndLimits(template)
}

func (evaluation *Evaluation) removeContainersResourceRequestsAndLimits(template *templatev1.Template) {
	for _, rawExtension := range template.Objects {
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
