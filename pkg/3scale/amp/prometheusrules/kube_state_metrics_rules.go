package prometheusrules

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func init() {
	PrometheusRuleFactories = append(PrometheusRuleFactories, NewKubeStateMetricsPrometheusRuleFactory)
}

type KubeStateMetricsPrometheusRuleFactory struct {
}

func NewKubeStateMetricsPrometheusRuleFactory() PrometheusRuleFactory {
	return &KubeStateMetricsPrometheusRuleFactory{}
}

func (s *KubeStateMetricsPrometheusRuleFactory) Type() string {
	return "threescale-kube-state-metrics"
}

func (s *KubeStateMetricsPrometheusRuleFactory) PrometheusRule(ns string) *monitoringv1.PrometheusRule {
	appLabel := appsv1alpha1.Default3scaleAppLabel
	return component.KubeStateMetricsPrometheusRules(ns, appLabel)
}
