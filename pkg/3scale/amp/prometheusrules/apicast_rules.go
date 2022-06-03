package prometheusrules

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

	appsv1beta1 "github.com/3scale/3scale-operator/apis/apps/v1beta1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func init() {
	PrometheusRuleFactories = append(PrometheusRuleFactories, NewApicastPrometheusRuleFactory)
}

type ApicastPrometheusRuleFactory struct {
}

func NewApicastPrometheusRuleFactory() PrometheusRuleFactory {
	return &ApicastPrometheusRuleFactory{}
}

func (b *ApicastPrometheusRuleFactory) Type() string {
	return "apicast"
}

func (b *ApicastPrometheusRuleFactory) PrometheusRule(ns string) *monitoringv1.PrometheusRule {
	options, err := apicastOptions(ns)
	if err != nil {
		panic(err)
	}
	return component.NewApicast(options).ApicastPrometheusRules()
}

func apicastOptions(ns string) (*component.ApicastOptions, error) {
	o := component.NewApicastOptions()

	// Required options for generating PrometheusRules
	o.CommonLabels = commonApicastLabels()
	o.Namespace = ns

	// Required options for passing validation, but not needed for generating the prometheus rules
	// To fix this, more granularity at options level.
	o.ManagementAPI = "_"
	o.OpenSSLVerify = "_"
	o.ResponseCodes = "_"
	o.ImageTag = "_"

	o.CommonStagingLabels = map[string]string{}
	o.CommonProductionLabels = map[string]string{}
	o.StagingPodTemplateLabels = map[string]string{}
	o.ProductionPodTemplateLabels = map[string]string{}

	o.StagingTracingConfig = &component.APIcastTracingConfig{TracingLibrary: component.APIcastDefaultTracingLibrary}
	o.ProductionTracingConfig = &component.APIcastTracingConfig{TracingLibrary: component.APIcastDefaultTracingLibrary}

	o.AdditionalPodAnnotations = map[string]string{}

	return o, o.Validate()
}

func commonApicastLabels() map[string]string {
	return map[string]string{
		"app":                  appsv1beta1.Default3scaleAppLabel,
		"threescale_component": "apicast",
	}
}
