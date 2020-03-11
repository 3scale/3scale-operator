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

func ZyncMonitoringService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync-monitoring",
			Labels: map[string]string{
				"app":                          appsv1alpha1.Default3scaleAppLabel,
				"threescale_component":         "zync",
				"threescale_component_element": "zync",
				"monitoring-key":               common.MonitoringKey,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "metrics",
					Protocol:   v1.ProtocolTCP,
					Port:       9393,
					TargetPort: intstr.FromInt(9393),
				},
			},
			Selector: map[string]string{"deploymentConfig": "zync"},
		},
	}
}

func ZyncQueMonitoringService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync-que-monitoring",
			Labels: map[string]string{
				"app":                          appsv1alpha1.Default3scaleAppLabel,
				"threescale_component":         "zync",
				"threescale_component_element": "zync-que",
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
			Selector: map[string]string{"deploymentConfig": "zync-que"},
		},
	}
}

func ZyncServiceMonitor() *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync",
			Labels: map[string]string{
				// TODO from options
				"monitoring-key":               common.MonitoringKey,
				"threescale_component":         "zync",
				"threescale_component_element": "zync",
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
					"threescale_component":         "zync",
					"threescale_component_element": "zync",
					"monitoring-key":               common.MonitoringKey,
				},
			},
		},
	}
}

func ZyncQueServiceMonitor() *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync-que",
			Labels: map[string]string{
				// TODO from options
				"monitoring-key":               common.MonitoringKey,
				"threescale_component":         "zync",
				"threescale_component_element": "zync-que",
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
					"threescale_component":         "zync",
					"threescale_component_element": "zync-que",
					"monitoring-key":               common.MonitoringKey,
				},
			},
		},
	}
}

func ZyncGrafanaDashboard(ns string) *grafanav1alpha1.GrafanaDashboard {
	data := &struct {
		Namespace string
	}{
		ns,
	}
	return &grafanav1alpha1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync",
			Labels: map[string]string{
				"monitoring-key": common.MonitoringKey,
			},
		},
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Json: assets.TemplateAsset("monitoring/zync-grafana-dashboard-1.json.tpl", data),
			Name: fmt.Sprintf("%s/zync-grafana-dashboard-1.json", ns),
		},
	}
}

func ZyncQueGrafanaDashboard(ns string) *grafanav1alpha1.GrafanaDashboard {
	data := &struct {
		Namespace string
	}{
		ns,
	}
	return &grafanav1alpha1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync-que",
			Labels: map[string]string{
				"monitoring-key": common.MonitoringKey,
			},
		},
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Json: assets.TemplateAsset("monitoring/zync-que-grafana-dashboard-1.json.tpl", data),
			Name: fmt.Sprintf("%s/zync-que-grafana-dashboard-1.json", ns),
		},
	}
}

func ZyncPrometheusRules(ns string) *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		ObjectMeta: v12.ObjectMeta{
			Name: "zync",
			Labels: map[string]string{
				"monitoring-key": common.MonitoringKey,
				"prometheus":     "application-monitoring",
				"role":           "alert-rules",
			},
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: fmt.Sprintf("%s/zync.rules", ns),
					Rules: []monitoringv1.Rule{
						{
							Alert: "PumaWorkersRunningLow",
							Annotations: map[string]string{
								"summary":     "{{$labels.container_name}} replica controller on {{$labels.namespace}}: Has less than 5 puma workers in the last 5 minutes",
								"description": "{{$labels.container_name}} replica controller on {{$labels.namespace}} project: Has less than 5 puma workers in the last 5 minutes",
							},
							Expr: intstr.FromString(fmt.Sprintf(`avg_over_time(puma_running{job="zync-monitoring",namespace="%s"} [5m]) < 5`, ns)),
							For:  "30m",
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

func ZyncQuePrometheusRules(ns string) *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		ObjectMeta: v12.ObjectMeta{
			Name: "zync-que",
			Labels: map[string]string{
				"monitoring-key": common.MonitoringKey,
				"prometheus":     "application-monitoring",
				"role":           "alert-rules",
			},
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: fmt.Sprintf("%s/zync-que.rules", ns),
					Rules: []monitoringv1.Rule{
						{
							Alert: "QueWorkersRunningLow",
							Annotations: map[string]string{
								"summary":     "{{$labels.container_name}} replica controller on {{$labels.namespace}}: Has less than 5 que workers in the last 5 minutes",
								"description": "{{$labels.container_name}} replica controller on {{$labels.namespace}} project: Has less than 5 que workers in the last 5 minutes",
							},
							Expr: intstr.FromString(fmt.Sprintf(`avg_over_time(que_workers_total{job="zync-que-monitoring",namespace="%s"} [5m]) < 5`, ns)),
							For:  "30m",
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
