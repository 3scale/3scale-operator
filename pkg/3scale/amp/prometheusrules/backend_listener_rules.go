package prometheusrules

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func init() {
	PrometheusRuleFactories = append(PrometheusRuleFactories, NewBackendListenerPrometheusRuleFactory)
}

type BackendListenerPrometheusRuleFactory struct {
}

func NewBackendListenerPrometheusRuleFactory() PrometheusRuleFactory {
	return &BackendListenerPrometheusRuleFactory{}
}

func (b *BackendListenerPrometheusRuleFactory) Type() string {
	return "backend-listener"
}

func (b *BackendListenerPrometheusRuleFactory) PrometheusRule(_ bool, ns string) *monitoringv1.PrometheusRule {
	options, err := backendOptions(ns)
	if err != nil {
		panic(err)
	}
	return component.NewBackend(options).BackendListenerPrometheusRules()
}
