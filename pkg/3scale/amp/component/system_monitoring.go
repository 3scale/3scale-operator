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

func SystemGrafanaDashboard(ns string) *grafanav1alpha1.GrafanaDashboard {
	data := &struct {
		Namespace string
	}{
		ns,
	}
	return &grafanav1alpha1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system",
			Labels: map[string]string{
				"monitoring-key": common.MonitoringKey,
			},
		},
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Json: assets.TemplateAsset("monitoring/system-grafana-dashboard-1.json.tpl", data),
			Name: fmt.Sprintf("%s/system-grafana-dashboard-1.json", ns),
		},
	}
}

func SystemSidekiqPrometheusRules(ns string) *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-sidekiq",
			Labels: map[string]string{
				"prometheus": "application-monitoring",
				"role":       "alert-rules",
			},
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: fmt.Sprintf("%s/system-sidekiq.rules", ns),
					Rules: []monitoringv1.Rule{
						{
							Alert: "SystemSidekiqZyncRuntime",
							Annotations: map[string]string{
								"summary":     "Rule example:  Zync runtime average more than 300 seconds",
								"description": "Rule example:  Zync runtime average more than 300 seconds",
							},
							Expr: intstr.FromString(fmt.Sprintf(`avg(sidekiq_job_runtime_seconds_sum{queue="zync",worker="ZyncWorker",namespace="%s"}) > 300`, ns)),
							For:  "10m",
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
