package component

import (
	templatev1 "github.com/openshift/api/template/v1"
)

type AmpHATemplate struct {
	options    []string
	components []Component
}

type AmpHATemplateOptions struct {
	ampTemplateOptions      AmpTemplateOptions
	highAvailabilityOptions HighAvailabilityOptions
}

func NewAmpHATemplate(options []string) *AmpHATemplate {
	components := []Component{
		NewAmpImages(options),
		NewWildcardRouterImage(options),
		NewRedis(options),
		NewBackend(options),
		NewMemcached(options),
		NewSystem(options),
		NewZync(options),
		NewApicast(options),
		NewWildcardRouter(options),
		NewHighAvailability(options),
	}

	ampHATemplate := &AmpHATemplate{
		options:    options,
		components: components,
	}
	return ampHATemplate
}

func (ampHATemplate *AmpHATemplate) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	for componentIdx := range ampHATemplate.components {
		ampHATemplate.components[componentIdx].AssembleIntoTemplate(template, otherComponents)
	}
	ampHATemplate.setTemplateFields(template)
}

func (ampHATemplate *AmpHATemplate) PostProcess(template *templatev1.Template, otherComponents []Component) {
	for componentIdx := range ampHATemplate.components {
		ampHATemplate.components[componentIdx].PostProcess(template, otherComponents)
	}
}

func (ampHATemplate *AmpHATemplate) setTemplateFields(template *templatev1.Template) {
	template.Name = "3scale-api-management-ha"
	template.ObjectMeta.Annotations["description"] = "3scale API Management main system (High Availability)"
}
