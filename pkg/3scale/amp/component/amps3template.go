package component

import (
	templatev1 "github.com/openshift/api/template/v1"
)

type AmpS3Template struct {
	options    []string
	components []Component
}

type AmpS3TemplateOptions struct {
	ampTemplateOptions AmpTemplateOptions
	s3Options          S3Options
}

func NewAmpS3Template(options []string) *AmpS3Template {
	components := []Component{
		NewAmpTemplate(options),
		NewS3(options),
	}

	ampS3Template := &AmpS3Template{
		options:    options,
		components: components,
	}
	return ampS3Template
}

func (ampS3Template *AmpS3Template) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	for componentIdx := range ampS3Template.components {
		ampS3Template.components[componentIdx].AssembleIntoTemplate(template, otherComponents)
	}
	ampS3Template.setTemplateFields(template)
}

func (ampS3Template *AmpS3Template) PostProcess(template *templatev1.Template, otherComponents []Component) {
	for componentIdx := range ampS3Template.components {
		ampS3Template.components[componentIdx].PostProcess(template, otherComponents)
	}
}

func (ampS3Template *AmpS3Template) setTemplateFields(template *templatev1.Template) {
	template.Name = "3scale-api-management-s3"
	template.ObjectMeta.Annotations["description"] = "3scale API Management main system with shared file storage in AWS S3."
}
