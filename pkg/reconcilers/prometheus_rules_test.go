package reconcilers

import (
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestRemovePrometheusRulesMutator(t *testing.T) {
	type args struct {
		existing client.Object
		desired  client.Object
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "ThreescaleApicastRequestTime alert exists",
			args: args{
				existing: &monitoringv1.PrometheusRule{
					Spec: monitoringv1.PrometheusRuleSpec{
						Groups: []monitoringv1.RuleGroup{
							{
								Name: "test-group",
								Rules: []monitoringv1.Rule{
									{
										Alert: "ThreescaleApicastRequestTime",
									},
									{
										Alert: "TestAlert",
									},
								},
							},
						},
					},
				},
				desired: &monitoringv1.PrometheusRule{
					Spec: monitoringv1.PrometheusRuleSpec{
						Groups: []monitoringv1.RuleGroup{
							{
								Name: "test-group",
								Rules: []monitoringv1.Rule{
									{
										Alert: "TestAlert",
									},
								},
							},
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "ThreescaleApicastRequestTime alert does not exist",
			args: args{
				existing: &monitoringv1.PrometheusRule{
					Spec: monitoringv1.PrometheusRuleSpec{
						Groups: []monitoringv1.RuleGroup{
							{
								Name: "test-group",
								Rules: []monitoringv1.Rule{
									{
										Alert: "TestAlert",
									},
								},
							},
						},
					},
				},
				desired: &monitoringv1.PrometheusRule{
					Spec: monitoringv1.PrometheusRuleSpec{
						Groups: []monitoringv1.RuleGroup{
							{
								Name: "test-group",
								Rules: []monitoringv1.Rule{
									{
										Alert: "TestAlert",
									},
								},
							},
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RemovePrometheusRulesMutator(tt.args.existing, tt.args.desired)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemovePrometheusRulesMutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RemovePrometheusRulesMutator() = %v, want %v", got, tt.want)
			}
		})
	}
}
