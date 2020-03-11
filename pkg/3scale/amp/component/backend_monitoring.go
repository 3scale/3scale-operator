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

func BackendWorkerMonitoringService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "backend-worker-monitoring",
			Labels: map[string]string{
				"app":                          appsv1alpha1.Default3scaleAppLabel,
				"threescale_component":         "backend",
				"threescale_component_element": "worker",
				"monitoring-key":               common.MonitoringKey,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "metrics",
					Protocol:   v1.ProtocolTCP,
					Port:       9421,
					TargetPort: intstr.FromInt(9421),
				},
			},
			Selector: map[string]string{"deploymentConfig": "backend-worker"},
		},
	}
}

func BackendWorkerServiceMonitor() *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name: "backend-worker",
			Labels: map[string]string{
				// TODO from options
				"monitoring-key":               common.MonitoringKey,
				"threescale_component":         "backend",
				"threescale_component_element": "worker",
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
					"threescale_component":         "backend",
					"threescale_component_element": "worker",
					"monitoring-key":               common.MonitoringKey,
				},
			},
		},
	}
}

func BackendWorkerGrafanaDashboard(ns string) *grafanav1alpha1.GrafanaDashboard {
	data := &struct {
		Namespace string
	}{
		ns,
	}
	return &grafanav1alpha1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name: "backend-worker",
			Labels: map[string]string{
				"monitoring-key": common.MonitoringKey,
			},
		},
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Json: assets.TemplateAsset("monitoring/backend-worker-grafana-dashboard-1.json.tpl", data),
			Name: fmt.Sprintf("%s/backend-worker-grafana-dashboard-1.json", ns),
		},
	}
}

func BackendWorkerPrometheusRules(ns string) *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		ObjectMeta: v12.ObjectMeta{
			Name: "backend-worker",
			Labels: map[string]string{
				"monitoring-key": common.MonitoringKey,
				"prometheus":     "application-monitoring",
				"role":           "alert-rules",
			},
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: fmt.Sprintf("%s/backend-worker.rules", ns),
					Rules: []monitoringv1.Rule{
						{
							Alert: "WorkersJobsCountRunningHigh",
							Annotations: map[string]string{
								"summary":     "{{$labels.container_name}} replica controller on {{$labels.namespace}}: Has more than 10000 jobs processed in the last 5 minutes",
								"description": "{{$labels.container_name}} replica controller on {{$labels.namespace}} project: Has more than 1000 jobs processed in the last 5 minutes",
							},
							Expr: intstr.FromString(fmt.Sprintf(`sum(avg_over_time(apisonator_worker_job_count{job="backend-worker-monitoring",namespace="%s"} [5m])) > 10000`, ns)),
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
