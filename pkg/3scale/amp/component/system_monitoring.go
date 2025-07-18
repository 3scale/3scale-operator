package component

import (
	"fmt"

	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	grafanav1beta1 "github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/3scale/3scale-operator/pkg/assets"
)

func (system *System) SystemSidekiqPodMonitor() *monitoringv1.PodMonitor {
	return &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-sidekiq",
			Labels: system.Options.CommonSidekiqLabels,
		},
		Spec: monitoringv1.PodMonitorSpec{
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{{
				Port:   "metrics",
				Path:   "/metrics",
				Scheme: "http",
			}},
			Selector: metav1.LabelSelector{
				MatchLabels: system.Options.CommonSidekiqLabels,
			},
		},
	}
}

func (system *System) SystemAppPodMonitor() *monitoringv1.PodMonitor {
	return &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-app",
			Labels: system.Options.CommonAppLabels,
		},
		Spec: monitoringv1.PodMonitorSpec{
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
				{
					Port:   SystemAppMasterContainerMetricsPortName,
					Path:   "/metrics",
					Scheme: "http",
				},
				{
					Port:   SystemAppProviderContainerMetricsPortName,
					Path:   "/metrics",
					Scheme: "http",
				},
				{
					Port:   SystemAppDeveloperContainerMetricsPortName,
					Path:   "/metrics",
					Scheme: "http",
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: system.Options.CommonAppLabels,
			},
		},
	}
}

func (system *System) SystemGrafanaV5Dashboard(sumRate string) *grafanav1beta1.GrafanaDashboard {
	data := &struct {
		Namespace, SumRate string
	}{
		system.Options.Namespace, sumRate,
	}
	return &grafanav1beta1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system",
			Labels: system.monitoringLabels(),
		},
		Spec: grafanav1beta1.GrafanaDashboardSpec{
			InstanceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"apim-management": "grafana",
				},
			},
			Json: assets.TemplateAsset("monitoring/system-grafana-dashboard-1.json.tpl", data),
		},
	}
}

func (system *System) SystemGrafanaV4Dashboard(sumRate string) *grafanav1alpha1.GrafanaDashboard {
	data := &struct {
		Namespace, SumRate string
	}{
		system.Options.Namespace, sumRate,
	}
	return &grafanav1alpha1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system",
			Labels: system.monitoringLabels(),
		},
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Json: assets.TemplateAsset("monitoring/system-grafana-dashboard-1.json.tpl", data),
		},
	}
}

func (system *System) SystemAppPrometheusRules() *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PrometheusRuleKind,
			APIVersion: fmt.Sprintf("%s/%s", monitoring.GroupName, monitoringv1.Version),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-app",
			Labels: system.prometheusRulesMonitoringLabels(),
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: fmt.Sprintf("%s/system-app.rules", system.Options.Namespace),
					Rules: []monitoringv1.Rule{
						{
							Alert: "ThreescaleSystemApp5XXRequestsHigh",
							Annotations: map[string]string{
								"sop_url":     ThreescaleSystemApp5XXRequestsHighURL,
								"summary":     "Job {{ $labels.job }} on {{ $labels.namespace }} has more than 50 HTTP 5xx requests in the last minute",
								"description": "Job {{ $labels.job }} on {{ $labels.namespace }} has more than 50 HTTP 5xx requests in the last minute",
							},
							Expr: intstr.FromString(fmt.Sprintf(`sum(rate(rails_requests_total{namespace="%s",pod=~"system-app-[a-z0-9]+-[a-z0-9]+",status=~"5[0-9]*"}[1m])) by (namespace,job) > 50`, system.Options.Namespace)),
							For:  "1m",
							Labels: map[string]string{
								"severity": "warning",
							},
						},
						{
							Alert: "ThreescaleSystemAppJobDown",
							Annotations: map[string]string{
								"sop_url":     ThreescalePrometheusJobDownURL,
								"summary":     "Job {{ $labels.job }} on {{ $labels.namespace }} is DOWN",
								"description": "Job {{ $labels.job }} on {{ $labels.namespace }} is DOWN",
							},
							Expr: intstr.FromString(fmt.Sprintf(`up{job=~".*system-app.*",namespace="%s"} == 0`, system.Options.Namespace)),
							For:  "1m",
							Labels: map[string]string{
								"severity": "critical",
							},
						},
					},
				},
			},
		},
	}
}

func (system *System) SystemSidekiqPrometheusRules() *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PrometheusRuleKind,
			APIVersion: fmt.Sprintf("%s/%s", monitoring.GroupName, monitoringv1.Version),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-sidekiq",
			Labels: system.prometheusRulesMonitoringLabels(),
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: fmt.Sprintf("%s/system-sidekiq.rules", system.Options.Namespace),
					Rules: []monitoringv1.Rule{
						{
							Alert: "ThreescaleSystemSidekiqJobDown",
							Annotations: map[string]string{
								"sop_url":     ThreescalePrometheusJobDownURL,
								"summary":     "Job {{ $labels.job }} on {{ $labels.namespace }} is DOWN",
								"description": "Job {{ $labels.job }} on {{ $labels.namespace }} is DOWN",
							},
							Expr: intstr.FromString(fmt.Sprintf(`up{job=~".*system-sidekiq.*",namespace="%s"} == 0`, system.Options.Namespace)),
							For:  "1m",
							Labels: map[string]string{
								"severity": "critical",
							},
						},
					},
				},
			},
		},
	}
}

func (system *System) monitoringLabels() map[string]string {
	labels := make(map[string]string)

	for key, value := range system.Options.CommonLabels {
		labels[key] = value
	}

	labels["monitoring-key"] = MonitoringKey
	return labels
}

func (system *System) prometheusRulesMonitoringLabels() map[string]string {
	labels := make(map[string]string)

	for key, value := range system.Options.CommonLabels {
		labels[key] = value
	}

	labels["prometheus"] = "application-monitoring"
	labels["role"] = "alert-rules"
	return labels
}
