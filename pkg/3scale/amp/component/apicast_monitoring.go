package component

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/assets"
	"github.com/3scale/3scale-operator/pkg/common"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (apicast *Apicast) ApicastProductionPodMonitor() *monitoringv1.PodMonitor {
	return &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "apicast-production",
			Labels: apicast.Options.CommonProductionLabels,
		},
		Spec: monitoringv1.PodMonitorSpec{
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{{
				Port:   "metrics",
				Path:   "/metrics",
				Scheme: "http",
			}},
			Selector: metav1.LabelSelector{
				MatchLabels: apicast.Options.CommonProductionLabels,
			},
		},
	}
}

func (apicast *Apicast) ApicastStagingPodMonitor() *monitoringv1.PodMonitor {
	return &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "apicast-staging",
			Labels: apicast.Options.CommonStagingLabels,
		},
		Spec: monitoringv1.PodMonitorSpec{
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{{
				Port:   "metrics",
				Path:   "/metrics",
				Scheme: "http",
			}},
			Selector: metav1.LabelSelector{
				MatchLabels: apicast.Options.CommonStagingLabels,
			},
		},
	}
}

func (apicast *Apicast) ApicastMainAppGrafanaDashboard() *grafanav1alpha1.GrafanaDashboard {
	data := &struct {
		Namespace string
	}{
		apicast.Options.Namespace,
	}

	return &grafanav1alpha1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "apicast-mainapp",
			Labels: apicast.monitoringLabels(),
		},
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Json: assets.TemplateAsset("monitoring/apicast-grafana-dashboard-1.json.tpl", data),
			Name: fmt.Sprintf("%s/apicast-grafana-dashboard-1.json", apicast.Options.Namespace),
		},
	}
}

func (apicast *Apicast) ApicastServicesGrafanaDashboard() *grafanav1alpha1.GrafanaDashboard {
	data := &struct {
		Namespace string
	}{
		apicast.Options.Namespace,
	}
	return &grafanav1alpha1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "apicast-services",
			Labels: apicast.monitoringLabels(),
		},
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Json: assets.TemplateAsset("monitoring/apicast-grafana-dashboard-2.json.tpl", data),
			Name: fmt.Sprintf("%s/apicast-grafana-dashboard-2.json", apicast.Options.Namespace),
		},
	}
}

func (apicast *Apicast) ApicastPrometheusRules() *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "apicast",
			Labels: apicast.prometheusRulesMonitoringLabels(),
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: fmt.Sprintf("%s/apicast.rules", apicast.Options.Namespace),
					Rules: []monitoringv1.Rule{
						{
							Alert: "ThreescaleApicastJobDown",
							Annotations: map[string]string{
								"sop_url":     ThreescalePodNotReadyURL,
								"summary":     "Job {{ $labels.job }} on {{ $labels.namespace }} is DOWN",
								"description": "Job {{ $labels.job }} on {{ $labels.namespace }} is DOWN",
							},
							Expr: intstr.FromString(fmt.Sprintf(`up{job=~".*/apicast-production|.*/apicast-staging",namespace="%s"} == 0`, apicast.Options.Namespace)),
							For:  "1m",
							Labels: map[string]string{
								"severity": "critical",
							},
						},
						{
							Alert: "ThreescaleApicastRequestTime",
							Annotations: map[string]string{
								"sop_url":     ThreescaleApicastRequestTimeURL,
								"summary":     "Request on instance {{ $labels.instance }} is taking more than one second to process the requests",
								"description": "High number of request taking more than a second to be processed",
							},
							Expr: intstr.FromString(fmt.Sprintf(`sum(rate(total_response_time_seconds_bucket{namespace='%s', pod=~'apicast-production.*'}[1m])) - sum(rate(upstream_response_time_seconds_bucket{namespace='%s', pod=~'apicast-production.*'}[1m])) > 1`, apicast.Options.Namespace, apicast.Options.Namespace)),
							For:  "2m",
							Labels: map[string]string{
								"severity": "warning",
							},
						},
						{
							Alert: "ThreescaleApicastHttp4xxErrorRate",
							Annotations: map[string]string{
								"sop_url":     ThreescaleApicastHttp4xxErrorRateURL,
								"summary":     "APICast high HTTP 4XX error rate (instance {{ $labels.instance }})",
								"description": "The number of request with 4XX is bigger than the 5% of total request.",
							},
							Expr: intstr.FromString(fmt.Sprintf(`sum(rate(apicast_status{namespace='%s', status=~"^4.."}[1m])) / sum(rate(apicast_status{namespace='%s'}[1m])) * 100 > 5`, apicast.Options.Namespace, apicast.Options.Namespace)),
							For:  "5m",
							Labels: map[string]string{
								"severity": "warning",
							},
						},
						{
							Alert: "ThreescaleApicastLatencyHigh",
							Annotations: map[string]string{
								"sop_url":     ThreescaleApicastLatencyHighURL,
								"summary":     "APICast latency high (instance {{ $labels.instance }})",
								"description": "APIcast p99 latency is higher than 5 seconds\n  VALUE = {{ $value }}\n  LABELS: {{ $labels }}",
							},
							Expr: intstr.FromString(fmt.Sprintf(`histogram_quantile(0.99, sum(rate(total_response_time_seconds_bucket{namespace='%s',}[30m])) by (le)) > 5`, apicast.Options.Namespace)),
							For:  "5m",
							Labels: map[string]string{
								"severity": "warning",
							},
						},
						{
							Alert: "ThreescaleApicastWorkerRestart",
							Annotations: map[string]string{
								"sop_url":     ThreescaleApicastWorkerRestartURL,
								"summary":     "A new worker process in Nginx has been started",
								"description": "A new thread has been started. This could indicate that a worker process has died due to the memory limits being exceeded. Please investigate the memory pressure on pod (instance {{ $labels.instance }})",
							},
							Expr: intstr.FromString(fmt.Sprintf(`changes(worker_process{namespace='%s', pod=~'apicast-production.*'}[5m]) > 0`, apicast.Options.Namespace)),
							For:  "5m",
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

func (apicast *Apicast) monitoringLabels() map[string]string {
	labels := make(map[string]string)

	for key, value := range apicast.Options.CommonLabels {
		labels[key] = value
	}

	labels["monitoring-key"] = common.MonitoringKey
	return labels
}

func (apicast *Apicast) prometheusRulesMonitoringLabels() map[string]string {
	labels := make(map[string]string)

	for key, value := range apicast.Options.CommonLabels {
		labels[key] = value
	}

	labels["prometheus"] = "application-monitoring"
	labels["role"] = "alert-rules"
	return labels
}
