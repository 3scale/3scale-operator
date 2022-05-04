package prometheusrules

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func init() {
	PrometheusRuleFactories = append(PrometheusRuleFactories, NewZyncQuePrometheusRuleFactory)
}

type ZyncQuePrometheusRuleFactory struct {
}

func NewZyncQuePrometheusRuleFactory() PrometheusRuleFactory {
	return &ZyncQuePrometheusRuleFactory{}
}

func (s *ZyncQuePrometheusRuleFactory) Type() string {
	return "zync-que"
}

func (s *ZyncQuePrometheusRuleFactory) PrometheusRule(_ bool, ns string) *monitoringv1.PrometheusRule {
	options, err := zyncOptions(ns)
	if err != nil {
		panic(err)
	}
	return component.NewZync(options).ZyncQuePrometheusRules()
}
