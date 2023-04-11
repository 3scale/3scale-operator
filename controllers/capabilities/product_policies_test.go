package controllers

import (
	"context"
	"reflect"
	"testing"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestProductThreescaleReconciler_convertPolicyConfiguration(t *testing.T) {
	type fields struct {
		BaseReconciler      *reconcilers.BaseReconciler
		resource            *capabilitiesv1beta1.Product
		productEntity       *controllerhelper.ProductEntity
		backendRemoteIndex  *controllerhelper.BackendAPIRemoteIndex
		threescaleAPIClient *threescaleapi.ThreeScaleClient
		logger              logr.Logger
	}
	type args struct {
		crdPolicy capabilitiesv1beta1.PolicyConfig
	}

	scheme := runtime.NewScheme()
	if err := capabilitiesv1beta1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	const productNamespace = "productNs"
	const configSecretNamespace = "configNs"
	const configSecretName = "configSecret"

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "test configuration from empty value",
			fields: fields{
				BaseReconciler: reconcilers.NewBaseReconciler(context.Background(), fake.NewClientBuilder().WithScheme(scheme).Build(), scheme, nil, logr.Logger{}, nil, nil),
			},
			args: args{
				crdPolicy: capabilitiesv1beta1.PolicyConfig{},
			},
			want: map[string]interface{}{},
		},
		{
			name: "test configuration from value",
			fields: fields{
				BaseReconciler: reconcilers.NewBaseReconciler(context.Background(), fake.NewClientBuilder().WithScheme(scheme).Build(), scheme, nil, logr.Logger{}, nil, nil),
			},
			args: args{
				crdPolicy: capabilitiesv1beta1.PolicyConfig{
					Configuration: runtime.RawExtension{
						Raw: []byte(`{"test": "test"}`),
					},
				},
			},
			want: map[string]interface{}{
				"test": "test",
			},
		},
		{
			name: "test configuration from secretRef in same namespace as product",
			fields: fields{
				BaseReconciler: reconcilers.NewBaseReconciler(context.Background(), fake.NewClientBuilder().WithScheme(scheme).WithObjects(&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      configSecretName,
						Namespace: productNamespace,
					},
					Data: map[string][]byte{
						capabilitiesv1beta1.ProductPolicyConfigurationPasswordSecretField: []byte(`{"testSecret": "testSecret"}`),
					},
				}).Build(), scheme, nil, logr.Logger{}, nil, nil),
				resource: &capabilitiesv1beta1.Product{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: productNamespace,
					},
				},
			},
			args: args{
				crdPolicy: capabilitiesv1beta1.PolicyConfig{
					ConfigurationRef: corev1.SecretReference{
						Name:      configSecretName,
						Namespace: productNamespace,
					},
				},
			},
			want: map[string]interface{}{
				"testSecret": "testSecret",
			},
		},
		{
			name: "test configuration from secretRef in different namespace compared to product",
			fields: fields{
				BaseReconciler: reconcilers.NewBaseReconciler(context.Background(), fake.NewClientBuilder().WithScheme(scheme).WithObjects(&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      configSecretName,
						Namespace: configSecretNamespace,
					},
					Data: map[string][]byte{
						capabilitiesv1beta1.ProductPolicyConfigurationPasswordSecretField: []byte(`{"testSecret": "testSecret"}`),
					},
				}).Build(), scheme, nil, logr.Logger{}, nil, nil),
				resource: &capabilitiesv1beta1.Product{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: productNamespace,
					},
				},
			},
			args: args{
				crdPolicy: capabilitiesv1beta1.PolicyConfig{
					ConfigurationRef: corev1.SecretReference{
						Name:      configSecretName,
						Namespace: configSecretNamespace,
					},
				},
			},
			want: map[string]interface{}{
				"testSecret": "testSecret",
			},
		},
		{
			name: "test secret not found in secretRef",
			fields: fields{
				BaseReconciler: reconcilers.NewBaseReconciler(context.Background(), fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build(), scheme, nil, logr.Logger{}, nil, nil),
				resource: &capabilitiesv1beta1.Product{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: productNamespace,
					},
				},
			},
			args: args{
				crdPolicy: capabilitiesv1beta1.PolicyConfig{
					ConfigurationRef: corev1.SecretReference{
						Name:      configSecretName,
						Namespace: configSecretNamespace,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test missing secret key in secret from secretRef",
			fields: fields{
				BaseReconciler: reconcilers.NewBaseReconciler(context.Background(), fake.NewClientBuilder().WithScheme(scheme).WithObjects(&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      configSecretName,
						Namespace: configSecretNamespace,
					},
					Data: map[string][]byte{},
				}).Build(), scheme, nil, logr.Logger{}, nil, nil),
				resource: &capabilitiesv1beta1.Product{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: productNamespace,
					},
				},
			},
			args: args{
				crdPolicy: capabilitiesv1beta1.PolicyConfig{
					ConfigurationRef: corev1.SecretReference{
						Name:      configSecretName,
						Namespace: configSecretNamespace,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test unmarshal error from secret in secretRef",
			fields: fields{
				BaseReconciler: reconcilers.NewBaseReconciler(context.Background(), fake.NewClientBuilder().WithScheme(scheme).WithObjects(&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      configSecretName,
						Namespace: configSecretNamespace,
					},
					Data: map[string][]byte{
						capabilitiesv1beta1.ProductPolicyConfigurationPasswordSecretField: []byte(`notJson`),
					},
				}).Build(), scheme, nil, logr.Logger{}, nil, nil),
				resource: &capabilitiesv1beta1.Product{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: productNamespace,
					},
				},
			},
			args: args{
				crdPolicy: capabilitiesv1beta1.PolicyConfig{
					ConfigurationRef: corev1.SecretReference{
						Name:      configSecretName,
						Namespace: configSecretNamespace,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			t := &ProductThreescaleReconciler{
				BaseReconciler:      tt.fields.BaseReconciler,
				resource:            tt.fields.resource,
				productEntity:       tt.fields.productEntity,
				backendRemoteIndex:  tt.fields.backendRemoteIndex,
				threescaleAPIClient: tt.fields.threescaleAPIClient,
				logger:              tt.fields.logger,
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
