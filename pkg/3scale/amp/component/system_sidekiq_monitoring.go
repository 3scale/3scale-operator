package component

import (
	"fmt"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/assets"
	"github.com/3scale/3scale-operator/pkg/common"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func SystemSidekiqMonitoringService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-sidekiq-monitoring",
			Labels: map[string]string{
				"app":                          appsv1alpha1.Default3scaleAppLabel,
				"threescale_component":         "system",
				"threescale_component_element": "sidekiq",
				"monitoring-key":               common.MonitoringKey,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "metrics",
					Protocol:   v1.ProtocolTCP,
					Port:       9394,
					TargetPort: intstr.FromInt(9394),
				},
			},
			Selector: map[string]string{"deploymentConfig": "system-sidekiq"},
		},
	}
}

func SystemSidekiqServiceMonitor() *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-sidekiq",
			Labels: map[string]string{
				// TODO from options
				"threescale_component":         "system",
				"threescale_component_element": "sidekiq",
				"monitoring-key":               common.MonitoringKey,
				"app":                          appsv1alpha1.Default3scaleAppLabel,
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{{
				Port:   "metrics",
				Path:   "/metrics",
				Scheme: "http",
			}},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					// TODO from options
					"app":                          appsv1alpha1.Default3scaleAppLabel,
					"threescale_component":         "system",
					"threescale_component_element": "sidekiq",
					"monitoring-key":               common.MonitoringKey,
				},
			},
		},
	}
}

func SystemSidekiqGrafanaDashboard(ns string) *grafanav1alpha1.GrafanaDashboard {
	data := &struct {
		Namespace string
	}{
		ns,
	}
	return &grafanav1alpha1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-sidekiq",
			Labels: map[string]string{
				"monitoring-key": common.MonitoringKey,
			},
		},
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Json: assets.TemplateAsset("monitoring/Sidekiq-grafana-dashboard-1.json.tpl", data),
			Name: fmt.Sprintf("%s/Sidekiq-grafana-dashboard-1.json", ns),
		},
	}
}

func SystemSidekiqPrometheusRules(ns string) *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		ObjectMeta: v12.ObjectMeta{
			Name: "system-sidekiq",
			Labels: map[string]string{
				"monitoring-key": common.MonitoringKey,
				"prometheus":     "application-monitoring",
				"role":           "alert-rules",
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
