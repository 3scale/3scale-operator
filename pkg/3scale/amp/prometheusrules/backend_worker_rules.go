package prometheusrules

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func init() {
	PrometheusRuleFactories = append(PrometheusRuleFactories, NewBackendWorkerPrometheusRuleFactory)
}

type BackendWorkerPrometheusRuleFactory struct {
}

func NewBackendWorkerPrometheusRuleFactory() PrometheusRuleFactory {
	return &BackendWorkerPrometheusRuleFactory{}
}

func (b *BackendWorkerPrometheusRuleFactory) Type() string {
	return "backend-worker"
}

func (b *BackendWorkerPrometheusRuleFactory) PrometheusRule() *monitoringv1.PrometheusRule {
	options, err := backendOptions()
	if err != nil {
		panic(err)
	}
	return component.NewBackend(options).BackendWorkerPrometheusRules()
}

func backendOptions() (*component.BackendOptions, error) {
	bo := component.NewBackendOptions()

	// Required options for generating PrometheusRules
	bo.CommonLabels = commonBackendLabels()
	bo.Namespace = "__NAMESPACE__"

	// Required options for passing validation, but not needed for generating the prometheus rules
	// To fix this, more granularity at options level.
	bo.WildcardDomain = "_"
	bo.ServiceEndpoint = "_"
	bo.RouteEndpoint = "_"
	bo.ImageTag = "_"
	bo.SystemBackendUsername = "_"
	bo.SystemBackendPassword = "_"
	bo.TenantName = "_"
	bo.CommonListenerLabels = map[string]string{}
	bo.CommonWorkerLabels = map[string]string{}
	bo.CommonCronLabels = map[string]string{}
	bo.ListenerPodTemplateLabels = map[string]string{}
	bo.WorkerPodTemplateLabels = map[string]string{}
	bo.CronPodTemplateLabels = map[string]string{}

	return bo, bo.Validate()
}

func commonBackendLabels() map[string]string {
	return map[string]string{
		"app":                  appsv1alpha1.Default3scaleAppLabel,
		"threescale_component": "backend",
	}
}
