package prometheusrules

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

type PrometheusRuleFactory interface {
	PrometheusRule(compatPre49 bool, ns string) *monitoringv1.PrometheusRule
	Type() string
}

type PrometheusRuleFactoryBuilder = func() PrometheusRuleFactory

// PrometheusRuleFactories is a list of prometheusrule factories
var PrometheusRuleFactories []PrometheusRuleFactoryBuilder
