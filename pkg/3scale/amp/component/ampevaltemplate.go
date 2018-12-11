package component

import (
	templatev1 "github.com/openshift/api/template/v1"
)

type AmpEvalTemplate struct {
	options    []string
	components []Component
}

func NewAmpEvalTemplate(options []string) *AmpEvalTemplate {
	components := []Component{
		NewAmpTemplate(options),
		NewEvaluation(options),
	}

	ampEvalTemplate := &AmpEvalTemplate{
		options:    options,
		components: components,
	}
	return ampEvalTemplate
}

func (ampEvalTemplate *AmpEvalTemplate) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	for componentIdx := range ampEvalTemplate.components {
		ampEvalTemplate.components[componentIdx].AssembleIntoTemplate(template, otherComponents)
	}
	ampEvalTemplate.setTemplateFields(template)
}

func (ampEvalTemplate *AmpEvalTemplate) PostProcess(template *templatev1.Template, otherComponents []Component) {
	for componentIdx := range ampEvalTemplate.components {
		ampEvalTemplate.components[componentIdx].PostProcess(template, otherComponents)
	}
}

func (ampEvalTemplate *AmpEvalTemplate) setTemplateFields(template *templatev1.Template) {
	template.Name = "3scale-api-management-eval"
	template.ObjectMeta.Annotations["description"] = "3scale API Management main system (Evaluation)"
}
