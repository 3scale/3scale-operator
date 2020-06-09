package component

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/assets"
	"github.com/3scale/3scale-operator/pkg/common"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (apicast *Apicast) ApicastProductionMonitoringService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "apicast-production-metrics",
			Labels: apicast.Options.ProductionMonitoringLabels,
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
			Selector: map[string]string{"deploymentConfig": "apicast-production"},
		},
	}
}

func (apicast *Apicast) ApicastStagingMonitoringService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "apicast-staging-metrics",
			Labels: apicast.Options.StagingMonitoringLabels,
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
			Selector: map[string]string{"deploymentConfig": "apicast-staging"},
		},
	}
}

func (apicast *Apicast) ApicastProductionServiceMonitor() *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "apicast-production",
			Labels: apicast.Options.ProductionMonitoringLabels,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{{
				Port:   "metrics",
				Path:   "/metrics",
				Scheme: "http",
			}},
			Selector: metav1.LabelSelector{
				MatchLabels: apicast.Options.ProductionMonitoringLabels,
			},
		},
	}
}

func (apicast *Apicast) ApicastStagingServiceMonitor() *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "apicast-staging",
			Labels: apicast.Options.StagingMonitoringLabels,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{{
				Port:   "metrics",
				Path:   "/metrics",
				Scheme: "http",
			}},
			Selector: metav1.LabelSelector{
				MatchLabels: apicast.Options.StagingMonitoringLabels,
			},
		},
	}
}

func ApicastGrafanaDashboard(ns string) *grafanav1alpha1.GrafanaDashboard {
	data := &struct {
		Namespace string
	}{
		ns,
	}
	return &grafanav1alpha1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name: "apicast",
			Labels: map[string]string{
				"monitoring-key": common.MonitoringKey,
			},
		},
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Json: assets.TemplateAsset("monitoring/apicast-grafana-dashboard-1.json.tpl", data),
			Name: fmt.Sprintf("%s/apicast-grafana-dashboard-1.json", ns),
		},
	}
}

func ApicastPrometheusRules(ns string) *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		ObjectMeta: v12.ObjectMeta{
			Name: "apicast",
			Labels: map[string]string{
				"monitoring-key": common.MonitoringKey,
				"prometheus":     "application-monitoring",
				"role":           "alert-rules",
			},
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: fmt.Sprintf("%s/apicast.rules", ns),
					Rules: []monitoringv1.Rule{
						{
							Alert: "ApicastDroppedConnections",
							Annotations: map[string]string{
								"summary":     "{{$labels.container_name}} replica controller on {{$labels.namespace}}: Has more than 10 dropped connections in the last 5 minutes",
								"description": "{{$labels.container_name}} replica controller on {{$labels.namespace}} project: Has more than 10 dropped connections in the last 5 minutes",
							},
							Expr: intstr.FromString(fmt.Sprintf(`sum (increase(nginx_http_connections{namespace="%s",job=~"apicast-(production|staging)-monitoring",state="accepted"}[5m])) - sum (increase(nginx_http_connections{namespace="%s",job=~"apicast-(production|staging)-monitoring",state="handled"}[5m])) > 10`, ns, ns)),
							For:  "2m",
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
