package helper

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestIsSecretWatchedBy3scale(t *testing.T) {
	labeledSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "labeled-secret",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"apimanager.apps.3scale.net/watched-by": "apimanager",
			},
		},
	}
	unlabeledSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "unlabeled-secret",
			Namespace: "test-namespace",
		},
	}

	type args struct {
		secret *v1.Secret
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Secret doesn't have watched-by label",
			args: args{
				secret: unlabeledSecret,
			},
			want: false,
		},
		{
			name: "Secret has watched-by label",
			args: args{
				secret: labeledSecret,
			},
			want: true,
		},
		{
			name: "Secret doesn't exist",
			args: args{
				secret: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSecretWatchedBy3scale(tt.args.secret); got != tt.want {
				t.Errorf("IsSecretWatchedBy3scale() = %v, want %v", got, tt.want)
			}
		})
	}
}
