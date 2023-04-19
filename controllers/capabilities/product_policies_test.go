package controllers

import (
	"context"
	"reflect"
	"testing"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestProductThreescaleReconciler_convertPolicyConfiguration(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := capabilitiesv1beta1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	type fields struct {
		BaseReconciler *reconcilers.BaseReconciler
		resource       *capabilitiesv1beta1.Product
	}
	type args struct {
		crdPolicy capabilitiesv1beta1.PolicyConfig
	}

	const (
		productNamespace      = "productNs"
		configSecretNamespace = "configNs"
		configSecretName      = "configSecret"
		plainConfigValue      = `{"test": "test"}`
		secretConfigValue     = `{"testSecret": "testSecret"}`
	)

	var (
		expectedConfigFromValue  = map[string]interface{}{"test": "test"}
		expectedConfigFromSecret = map[string]interface{}{"testSecret": "testSecret"}
		productResource          = &capabilitiesv1beta1.Product{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: productNamespace,
			},
		}
	)

	policyFactory := func(plainValue []byte, secretValue corev1.SecretReference) capabilitiesv1beta1.PolicyConfig {
		policy := capabilitiesv1beta1.PolicyConfig{
			Configuration: runtime.RawExtension{
				Raw: plainValue,
			},
			ConfigurationRef: secretValue,
		}

		return policy
	}

	reconcilerFactory := func(objects ...runtime.Object) *reconcilers.BaseReconciler {
		return reconcilers.NewBaseReconciler(context.Background(), fake.NewClientBuilder().WithRuntimeObjects(objects...).WithScheme(scheme).Build(), scheme, nil, logr.Logger{}, nil, nil)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "test configuration from value",
			fields: fields{
				BaseReconciler: reconcilerFactory(),
			},
			args: args{
				crdPolicy: policyFactory([]byte(plainConfigValue), corev1.SecretReference{}),
			},
			want: expectedConfigFromValue,
		},
		{
			name: "test configuration from value takes precedence over secret",
			fields: fields{
				BaseReconciler: reconcilerFactory(
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      configSecretName,
							Namespace: productNamespace,
						},
						Data: map[string][]byte{
							capabilitiesv1beta1.ProductPolicyConfigurationPasswordSecretField: []byte(secretConfigValue),
						},
					},
				),
				resource: productResource,
			},
			args: args{
				crdPolicy: policyFactory(
					[]byte(plainConfigValue),
					corev1.SecretReference{
						Name:      configSecretName,
						Namespace: productNamespace,
					},
				),
			},
			want: expectedConfigFromValue,
		},
		{
			name: "test configuration from secret takes precedence if plain value is using default",
			fields: fields{
				BaseReconciler: reconcilerFactory(
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      configSecretName,
							Namespace: productNamespace,
						},
						Data: map[string][]byte{
							capabilitiesv1beta1.ProductPolicyConfigurationPasswordSecretField: []byte(secretConfigValue),
						},
					},
				),
				resource: productResource,
			},
			args: args{
				crdPolicy: policyFactory(
					[]byte(capabilitiesv1beta1.ProductPolicyConfigurationDefault),
					corev1.SecretReference{
						Name:      configSecretName,
						Namespace: productNamespace,
					},
				),
			},
			want: expectedConfigFromSecret,
		},
		{
			name: "test configuration from secretRef in different namespace compared to product",
			fields: fields{
				BaseReconciler: reconcilerFactory(
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      configSecretName,
							Namespace: configSecretNamespace,
						},
						Data: map[string][]byte{
							capabilitiesv1beta1.ProductPolicyConfigurationPasswordSecretField: []byte(secretConfigValue),
						},
					},
				),
				resource: productResource,
			},
			args: args{
				crdPolicy: policyFactory(
					[]byte(capabilitiesv1beta1.ProductPolicyConfigurationDefault),
					corev1.SecretReference{
						Name:      configSecretName,
						Namespace: configSecretNamespace,
					},
				),
			},
			want: expectedConfigFromSecret,
		},
		{
			name: "test error secret not found in secretRef",
			fields: fields{
				BaseReconciler: reconcilerFactory(),
				resource:       productResource,
			},
			args: args{
				crdPolicy: policyFactory(
					[]byte(capabilitiesv1beta1.ProductPolicyConfigurationDefault),
					corev1.SecretReference{
						Name:      configSecretName,
						Namespace: productNamespace,
					},
				),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test error missing secret key in secret from secretRef",
			fields: fields{
				BaseReconciler: reconcilerFactory(
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      configSecretName,
							Namespace: configSecretNamespace,
						},
						Data: map[string][]byte{},
					},
				),
				resource: productResource,
			},
			args: args{
				crdPolicy: policyFactory(
					[]byte(capabilitiesv1beta1.ProductPolicyConfigurationDefault),
					corev1.SecretReference{
						Name:      configSecretName,
						Namespace: productNamespace,
					},
				),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test unmarshal error from secret in secretRef",
			fields: fields{
				BaseReconciler: reconcilerFactory(
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      configSecretName,
							Namespace: configSecretNamespace,
						},
						Data: map[string][]byte{
							capabilitiesv1beta1.ProductPolicyConfigurationPasswordSecretField: []byte(`notJson`),
						},
					},
				),
				resource: productResource,
			},
			args: args{
				crdPolicy: policyFactory(
					[]byte(capabilitiesv1beta1.ProductPolicyConfigurationDefault),
					corev1.SecretReference{
						Name:      configSecretName,
						Namespace: productNamespace,
					},
				),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			t := &ProductThreescaleReconciler{
				BaseReconciler: tt.fields.BaseReconciler,
				resource:       tt.fields.resource,
			}
			got, err := t.convertPolicyConfiguration(tt.args.crdPolicy)
			if (err != nil) != tt.wantErr {
				t1.Errorf("convertPolicyConfiguration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("convertPolicyConfiguration() got = %v, want %v", got, tt.want)
			}
		})
	}
}
