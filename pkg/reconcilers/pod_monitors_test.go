package reconcilers

import (
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func podMonitoringTestFactory(port string) *monitoringv1.PodMonitor {
	return &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-app",
		},
		Spec: monitoringv1.PodMonitorSpec{
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
				{
					Port:   port,
					Path:   "/metrics",
					Scheme: "http",
				},
			},
		},
	}
}

func TestGenericPodMonitorMutator(t *testing.T) {
	existingPort := "8080"
	desiredPort := "9090"

	existing := podMonitoringTestFactory(existingPort)
	desired := podMonitoringTestFactory(desiredPort)

	update, err := GenericPodMonitorMutator(existing, desired)
	if err != nil {
		t.Fatal(err)
	}
	if !update {
		t.Fatal("when defaults can be applied, reconciler reported no update needed")
	}

	if existing.Spec.PodMetricsEndpoints[0].Port != desiredPort {
		t.Fatalf("PodMonitor not reconciled. Expected: %s, got: %s", desiredPort, existing.Spec.PodMetricsEndpoints[0].Port)
	}
}
