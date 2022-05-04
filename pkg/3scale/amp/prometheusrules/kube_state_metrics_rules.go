package prometheusrules

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

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

func (s *KubeStateMetricsPrometheusRuleFactory) PrometheusRule(compatPre49 bool, ns string) *monitoringv1.PrometheusRule {
	sumRate := "sum_irate"
	if compatPre49 {
		sumRate = "sum_rate"
	}

	appLabel := appsv1alpha1.Default3scaleAppLabel
	return component.KubeStateMetricsPrometheusRules(sumRate, ns, appLabel)
}
