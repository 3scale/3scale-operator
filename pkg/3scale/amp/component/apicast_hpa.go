package component

import (
	"github.com/3scale/3scale-operator/pkg/helper"
	hpa "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (apicast *Apicast) ApicastProductionHpa(namespace string, minPods *int32, maxPods int32, cpuPercent *int32, memoryPercent *int32) *hpa.HorizontalPodAutoscaler {
	if minPods == nil {
		minPods = helper.Int32Ptr(1)
	}
	if maxPods == 0 {
		maxPods = 5
	}
	if cpuPercent == nil {
		cpuPercent = helper.Int32Ptr(90)
	}
	if memoryPercent == nil {
		memoryPercent = helper.Int32Ptr(90)
	}
	// currently pointing at dc will need to change to deployment
	return &hpa.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ApicastProductionName,
			Namespace: namespace,
		},
		Spec: hpa.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: hpa.CrossVersionObjectReference{
				Kind:       "DeploymentConfig",
				Name:       ApicastProductionName,
				APIVersion: "apps.openshift.io/v1",
			},
			MinReplicas: minPods,
			MaxReplicas: maxPods,
			Metrics: []hpa.MetricSpec{
				{
					Type: hpa.ResourceMetricSourceType,
					Resource: &hpa.ResourceMetricSource{
						Name: "memory",
						Target: hpa.MetricTarget{
							Type:               hpa.UtilizationMetricType,
							AverageUtilization: memoryPercent,
						},
					},
				},
				{
					Type: hpa.ResourceMetricSourceType,
					Resource: &hpa.ResourceMetricSource{
						Name: "cpu",
						Target: hpa.MetricTarget{
							Type:               hpa.UtilizationMetricType,
							AverageUtilization: cpuPercent,
						},
					},
				},
			},
		},
	}
}
