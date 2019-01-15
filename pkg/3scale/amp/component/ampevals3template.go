package component

import (
	templatev1 "github.com/openshift/api/template/v1"
)

type AmpEvalS3Template struct {
	options    []string
	components []Component
}

func NewAmpEvalS3Template(options []string) *AmpEvalS3Template {
	components := []Component{
		NewAmpTemplate(options),
		NewEvaluation(options),
		NewS3(options),
	}

	ampEvalS3Template := &AmpEvalS3Template{
		options:    options,
		components: components,
	}
	return ampEvalS3Template
}

func (ampEvalS3Template *AmpEvalS3Template) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	for componentIdx := range ampEvalS3Template.components {
		ampEvalS3Template.components[componentIdx].AssembleIntoTemplate(template, otherComponents)
	}
	ampEvalS3Template.setTemplateFields(template)
}

func (ampEvalS3Template *AmpEvalS3Template) PostProcess(template *templatev1.Template, otherComponents []Component) {
	for componentIdx := range ampEvalS3Template.components {
		ampEvalS3Template.components[componentIdx].PostProcess(template, otherComponents)
	}
}

func (ampEvalS3Template *AmpEvalS3Template) setTemplateFields(template *templatev1.Template) {
	template.Name = "3scale-api-management-eval-s3"
	template.ObjectMeta.Annotations["description"] = "3scale API Management main system (Evaluation) with shared file storage in AWS S3."
}
