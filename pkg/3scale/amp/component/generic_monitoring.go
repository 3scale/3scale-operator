package component

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/assets"
	"github.com/3scale/3scale-operator/pkg/common"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func KubernetesResourcesByNamespaceGrafanaDashboard(ns string) *grafanav1alpha1.GrafanaDashboard {
	data := &struct {
		Namespace string
	}{
		ns,
	}
	return &grafanav1alpha1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubernetes-resources-by-namespace",
			Labels: map[string]string{
				"monitoring-key": common.MonitoringKey,
			},
		},
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Json: assets.TemplateAsset("monitoring/kubernetes-resources-by-namespace-grafana-dashboard-1.json.tpl", data),
			Name: fmt.Sprintf("%s/kubernetes-resources-by-namespace-grafana-dashboard-1.json", ns),
		},
	}
}

func KubernetesResourcesByPodGrafanaDashboard(ns string) *grafanav1alpha1.GrafanaDashboard {
	data := &struct {
		Namespace string
	}{
		ns,
	}
	return &grafanav1alpha1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubernetes-resources-by-pod",
			Labels: map[string]string{
				"monitoring-key": common.MonitoringKey,
			},
		},
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Json: assets.TemplateAsset("monitoring/kubernetes-resources-by-pod-grafana-dashboard-1.json.tpl", data),
			Name: fmt.Sprintf("%s/kubernetes-resources-by-pod-grafana-dashboard-1.json", ns),
		},
	}
}

func KubeStateMetricsPrometheusRules(ns string) *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		ObjectMeta: v12.ObjectMeta{
			Name: "threescale-kube-state-metrics",
			Labels: map[string]string{
				"monitoring-key": common.MonitoringKey,
				"prometheus":     "application-monitoring",
				"role":           "alert-rules",
			},
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: fmt.Sprintf("%s/threescale-kube-state-metrics.rules", ns),
					Rules: []monitoringv1.Rule{
						{
							Alert: "ThreescalePodCrashLooping",
							Annotations: map[string]string{
								"message": `Pod {{ $labels.namespace }}/{{ $labels.pod }} ({{ $labels.container }}) is restarting {{ printf "%.2f" $value }} times / 5 minutes.`,
							},
							Expr: intstr.FromString(fmt.Sprintf(`rate(kube_pod_container_status_restarts_total{namespace="%s",pod=~"(apicast-.*|backend-.*|system-.*|zync-.*)"}[15m]) * 60 * 5 > 0`, ns)),
							For:  "5m",
							Labels: map[string]string{
								"severity": "critical",
							},
						},
						{
							Alert: "ThreescalePodNotReady",
							Annotations: map[string]string{
								"message": `Pod {{ $labels.namespace }}/{{ $labels.pod }} has been in a non-ready state for longer than 5 minutes.`,
							},
							Expr: intstr.FromString(fmt.Sprintf(`sum by (namespace, pod) (max by(namespace, pod) (kube_pod_status_phase{namespace="%s",pod=~"(apicast-.*|backend-.*|system-.*|zync-.*)", phase=~"Pending|Unknown"}) * on(namespace, pod) group_left(owner_kind) max by(namespace, pod, owner_kind) (kube_pod_owner{namespace="%s",owner_kind!="Job"})) > 0`, ns, ns)),
							For:  "5m",
							Labels: map[string]string{
								"severity": "critical",
							},
						},
						{
							Alert: "ThreescaleReplicationControllerReplicasMismatch",
							Annotations: map[string]string{
								"message": `ReplicationController {{ $labels.namespace }}/{{ $labels.replicationcontroller }} has not matched the expected number of replicas for longer than 5 minutes.`,
							},
							Expr: intstr.FromString(fmt.Sprintf(`kube_replicationcontroller_spec_replicas {namespace="%s",replicationcontroller=~"(apicast-.*|backend-.*|system-.*|zync-.*)"} != kube_replicationcontroller_status_ready_replicas {namespace="%s",replicationcontroller=~"(apicast-.*|backend-.*|system-.*|zync-.*)"}`, ns, ns)),
							For:  "5m",
							Labels: map[string]string{
								"severity": "critical",
							},
						},
						{
							Alert: "ThreescaleContainerWaiting",
							Annotations: map[string]string{
								"message": `Pod {{ $labels.namespace }}/{{ $labels.pod }} container {{ $labels.container }} has been in waiting state for longer than 1 hour.`,
							},
							Expr: intstr.FromString(fmt.Sprintf(`sum by (namespace, pod, container) (kube_pod_container_status_waiting_reason{namespace="%s",pod=~"(apicast-.*|backend-.*|system-.*|zync-.*)"}) > 0`, ns)),
							For:  "1h",
							Labels: map[string]string{
								"severity": "warning",
							},
						},
						{
							Alert: "ThreescaleContainerCPUHigh",
							Annotations: map[string]string{
								"message": `Pod {{ $labels.namespace }}/{{ $labels.pod }} container {{ $labels.container }} has High CPU usage for longer than 15 minutes.`,
							},
							Expr: intstr.FromString(fmt.Sprintf(`sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{namespace="%s",pod=~"(apicast-.*|backend-.*|system-.*|zync-.*)"}) by (namespace, container, pod) / sum(kube_pod_container_resource_limits_cpu_cores{namespace="%s",pod=~"(apicast-.*|backend-.*|system-.*|zync-.*)"}) by (namespace, container, pod) * 100 > 90`, ns, ns)),
							For:  "15m",
							Labels: map[string]string{
								"severity": "warning",
							},
						},
						{
							Alert: "ThreescaleContainerMemoryHigh",
							Annotations: map[string]string{
								"message": `Pod {{ $labels.namespace }}/{{ $labels.pod }} container {{ $labels.container }} has High Memory usage for longer than 15 minutes.`,
							},
							Expr: intstr.FromString(fmt.Sprintf(`sum(container_memory_usage_bytes{namespace="%s",container!="",pod=~"(apicast-.*|backend-.*|system-.*|zync-.*)"}) by(namespace, container, pod) / sum(kube_pod_container_resource_limits_memory_bytes{namespace="%s",pod=~"(apicast-.*|backend-.*|system-.*|zync-.*)"}) by(namespace, container, pod) * 100 > 90`, ns, ns)),
							For:  "15m",
							Labels: map[string]string{
								"severity": "warning",
							},
						},
						{
							Alert: "ThreescaleContainerCPUThrottlingHigh",
							Annotations: map[string]string{
								"message": `{{ $value | humanizePercentage }} throttling of CPU in namespace {{ $labels.namespace }} for container {{ $labels.container }} in pod {{ $labels.pod }}.`,
							},
							Expr: intstr.FromString(fmt.Sprintf(`sum(increase(container_cpu_cfs_throttled_periods_total{namespace="%s",container!="",pod=~"(apicast-.*|backend-.*|system-.*|zync-.*)" }[5m])) by (container, pod, namespace) / sum(increase(container_cpu_cfs_periods_total{namespace="%s",pod=~"(apicast-.*|backend-.*|system-.*|zync-.*)"}[5m])) by (container, pod, namespace) > ( 25 / 100 )`, ns, ns)),
							For:  "15m",
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
