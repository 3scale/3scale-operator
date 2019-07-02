package template

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"
	templatev1 "github.com/openshift/api/template/v1"
)

type TemplateType string

type TemplateFactory interface {
	Adapters() []adapters.Adapter
	Type() TemplateType
}

type TemplateFactoryBuilder = func() TemplateFactory

// TemplateFactories is a list of template factories
var TemplateFactories []TemplateFactoryBuilder

// NewTemplate implements the main loop of adapters to generate template object
func NewTemplate(templateName string) *templatev1.Template {
	templateFactoryIndex := map[TemplateType]TemplateFactory{}
	for _, factoryBuilder := range TemplateFactories {
		factory := factoryBuilder()
		if _, ok := templateFactoryIndex[factory.Type()]; ok {
			panic(fmt.Errorf("Template %s already exists", factory.Type()))
		}
		templateFactoryIndex[factory.Type()] = factory
	}

	factory, ok := templateFactoryIndex[TemplateType(templateName)]
	if !ok {
		panic(fmt.Errorf("Template %s not found", templateName))
	}

	tpl := Basic3scaleTemplate()
	for _, adapter := range factory.Adapters() {
		adapter.Adapt(tpl)
	}
	return tpl
}
