package reconcilers

import (
	"testing"

	"github.com/3scale/3scale-operator/pkg/helper"
	hpa "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func hpaTestFactory(maxPods int32) *hpa.HorizontalPodAutoscaler {
	return &hpa.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "someNs",
		},
		Spec: hpa.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: hpa.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       "test",
				APIVersion: "apps/v1",
			},
			MinReplicas: helper.Int32Ptr(1),
			MaxReplicas: maxPods,
			Metrics: []hpa.MetricSpec{
				{
					Type: hpa.ResourceMetricSourceType,
					Resource: &hpa.ResourceMetricSource{
						Name: "memory",
						Target: hpa.MetricTarget{
							Type:               hpa.UtilizationMetricType,
							AverageUtilization: helper.Int32Ptr(90),
						},
					},
				},
				{
					Type: hpa.ResourceMetricSourceType,
					Resource: &hpa.ResourceMetricSource{
						Name: "cpu",
						Target: hpa.MetricTarget{
							Type:               hpa.UtilizationMetricType,
							AverageUtilization: helper.Int32Ptr(90),
						},
					},
				},
			},
		},
	}
}

func TestGenericHPAMutator(t *testing.T) {
	var existingMaxPods int32 = 1
	var desiredMaxPods int32 = 2

	existing := hpaTestFactory(existingMaxPods)
	desired := hpaTestFactory(desiredMaxPods)

	update, err := GenericHPAMutator(existing, desired)
	if err != nil {
		t.Fatal(err)
	}
	if !update {
		t.Fatal("when defaults can be applied, reconciler reported no update needed")
	}

	if existing.Spec.MaxReplicas != desiredMaxPods {
		t.Fatalf("MaxReplicas not reconciled. Expected: %d, got: %d", desiredMaxPods, existing.Spec.MaxReplicas)
	}
}
