package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	templatev1 "github.com/openshift/api/template/v1"
)

type EvalAdapter struct {
}

func NewEvalAdapter() Adapter {
	return &EvalAdapter{}
}

func (e *EvalAdapter) Adapt(template *templatev1.Template) {
	objects := helper.UnwrapRawExtensions(template.Objects)
	evalComponent := component.NewEvaluation()
	evalComponent.RemoveContainersResourceRequestsAndLimits(objects)
}
