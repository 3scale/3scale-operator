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

func (zync *Zync) ZyncPodMonitor() *monitoringv1.PodMonitor {
	return &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "zync",
			Labels: zync.Options.CommonZyncLabels,
		},
		Spec: monitoringv1.PodMonitorSpec{
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{{
				Port:   "metrics",
				Path:   "/metrics",
				Scheme: "http",
			}},
			Selector: metav1.LabelSelector{
				MatchLabels: zync.Options.CommonZyncLabels,
			},
		},
	}
}

func (zync *Zync) ZyncQuePodMonitor() *monitoringv1.PodMonitor {
	return &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "zync-que",
			Labels: zync.Options.CommonZyncQueLabels,
		},
		Spec: monitoringv1.PodMonitorSpec{
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{{
				Port:   "metrics",
				Path:   "/metrics",
				Scheme: "http",
			}},
			Selector: metav1.LabelSelector{
				MatchLabels: zync.Options.CommonZyncQueLabels,
			},
		},
	}
}

func (zync *Zync) ZyncGrafanaV5Dashboard(sumRate string) *grafanav1beta1.GrafanaDashboard {
	data := &struct {
		Namespace, SumRate string
	}{
		zync.Options.Namespace, sumRate,
	}
	return &grafanav1beta1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "zync",
			Labels: zync.monitoringLabels(),
		},
		Spec: grafanav1beta1.GrafanaDashboardSpec{
			InstanceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"apim-management": "grafana",
				},
			},
			Json: assets.TemplateAsset("monitoring/zync-grafana-dashboard-1.json.tpl", data),
		},
	}
}

func (zync *Zync) ZyncGrafanaV4Dashboard(sumRate string) *grafanav1alpha1.GrafanaDashboard {
	data := &struct {
		Namespace, SumRate string
	}{
		zync.Options.Namespace, sumRate,
	}
	return &grafanav1alpha1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "zync",
			Labels: zync.monitoringLabels(),
		},
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Json: assets.TemplateAsset("monitoring/zync-grafana-dashboard-1.json.tpl", data),
		},
	}
}

func (zync *Zync) ZyncPrometheusRules() *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PrometheusRuleKind,
			APIVersion: fmt.Sprintf("%s/%s", monitoring.GroupName, monitoringv1.Version),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "zync",
			Labels: zync.prometheusRulesMonitoringLabels(),
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: fmt.Sprintf("%s/zync.rules", zync.Options.Namespace),
					Rules: []monitoringv1.Rule{
						{
							Alert: "ThreescaleZyncJobDown",
							Annotations: map[string]string{
								"sop_url":     ThreescalePrometheusJobDownURL,
								"summary":     "Job {{ $labels.job }} on {{ $labels.namespace }} is DOWN",
								"description": "Job {{ $labels.job }} on {{ $labels.namespace }} is DOWN",
							},
							Expr: intstr.FromString(fmt.Sprintf(`up{job=~".*/zync",namespace="%s"} == 0`, zync.Options.Namespace)),
							For:  "1m",
							Labels: map[string]string{
								"severity": "critical",
							},
						},
						{
							Alert: "ThreescaleZync5XXRequestsHigh",
							Annotations: map[string]string{
								"sop_url":     ThreescaleZync5XXRequestsHighURL,
								"summary":     "Job {{ $labels.job }} on {{ $labels.namespace }} has more than 50 HTTP 5xx requests in the last minute",
								"description": "Job {{ $labels.job }} on {{ $labels.namespace }} has more than 50 HTTP 5xx requests in the last minute",
							},
							Expr: intstr.FromString(fmt.Sprintf(`sum(rate(rails_requests_total{namespace="%s",pod=~"zync-[a-z0-9]+-[a-z0-9]+",status=~"5[0-9]*"}[1m])) by (namespace,job) > 50`, zync.Options.Namespace)),
							For:  "1m",
							Labels: map[string]string{
								"severity": "warning",
							},
						},
					},
				},
			},
		},
	}
}

func (zync *Zync) ZyncQuePrometheusRules() *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PrometheusRuleKind,
			APIVersion: fmt.Sprintf("%s/%s", monitoring.GroupName, monitoringv1.Version),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "zync-que",
			Labels: zync.prometheusRulesMonitoringLabels(),
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: fmt.Sprintf("%s/zync-que.rules", zync.Options.Namespace),
					Rules: []monitoringv1.Rule{
						{
							Alert: "ThreescaleZyncQueJobDown",
							Annotations: map[string]string{
								"sop_url":     ThreescalePrometheusJobDownURL,
								"summary":     "Job {{ $labels.job }} on {{ $labels.namespace }} is DOWN",
								"description": "Job {{ $labels.job }} on {{ $labels.namespace }} is DOWN",
							},
							Expr: intstr.FromString(fmt.Sprintf(`up{job=~".*/zync-que",namespace="%s"} == 0`, zync.Options.Namespace)),
							For:  "1m",
							Labels: map[string]string{
								"severity": "critical",
							},
						},
						{
							Alert: "ThreescaleZyncQueScheduledJobCountHigh",
							Annotations: map[string]string{
								"sop_url":     ThreescaleZyncQueScheduledJobCountHighURL,
								"summary":     "Job {{ $labels.job }} on {{ $labels.namespace }} has scheduled job count over 100",
								"description": "Job {{ $labels.job }} on {{ $labels.namespace }} has scheduled job count over 100",
							},
							Expr: intstr.FromString(fmt.Sprintf(`max(que_jobs_scheduled_total{pod=~'zync-que.*',type='scheduled',namespace="%s"}) by (namespace,job,exported_job) > 250`, zync.Options.Namespace)),
							For:  "1m",
							Labels: map[string]string{
								"severity": "warning",
							},
						},
						{
							Alert: "ThreescaleZyncQueFailedJobCountHigh",
							Annotations: map[string]string{
								"sop_url":     ThreescaleZyncQueFailedJobCountHighURL,
								"summary":     "Job {{ $labels.job }} on {{ $labels.namespace }} has failed job count over 100",
								"description": "Job {{ $labels.job }} on {{ $labels.namespace }} has failed job count over 100",
							},
							Expr: intstr.FromString(fmt.Sprintf(`max(que_jobs_scheduled_total{pod=~'zync-que.*',type='failed',namespace="%s"}) by (namespace,job,exported_job) > 250`, zync.Options.Namespace)),
							For:  "1m",
							Labels: map[string]string{
								"severity": "warning",
							},
						},
						{
							Alert: "ThreescaleZyncQueReadyJobCountHigh",
							Annotations: map[string]string{
								"sop_url":     ThreescaleZyncQueReadyJobCountHighURL,
								"summary":     "Job {{ $labels.job }} on {{ $labels.namespace }} has ready job count over 100",
								"description": "Job {{ $labels.job }} on {{ $labels.namespace }} has ready job count over 100",
							},
							Expr: intstr.FromString(fmt.Sprintf(`max(que_jobs_scheduled_total{pod=~'zync-que.*',type='ready',namespace="%s"}) by (namespace,job,exported_job) > 250`, zync.Options.Namespace)),
							For:  "1m",
							Labels: map[string]string{
								"severity": "warning",
							},
						},
					},
				},
			},
		},
	}
}

func (zync *Zync) monitoringLabels() map[string]string {
	labels := make(map[string]string)

	for key, value := range zync.Options.CommonLabels {
		labels[key] = value
	}

	labels["monitoring-key"] = MonitoringKey
	return labels
}

func (zync *Zync) prometheusRulesMonitoringLabels() map[string]string {
	labels := make(map[string]string)

	for key, value := range zync.Options.CommonLabels {
		labels[key] = value
	}

	labels["prometheus"] = "application-monitoring"
	labels["role"] = "alert-rules"
	return labels
}
