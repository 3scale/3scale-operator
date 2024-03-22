package controllers

import (
	"fmt"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/apispkg/common"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"reflect"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func TestApplicationAuthStatusReconciler_Reconcile(t *testing.T) {
	type fields struct {
		BaseReconciler *reconcilers.BaseReconciler
		resource       *capabilitiesv1beta1.ApplicationAuth
		reconcileError error
		logger         logr.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    reconcile.Result
		wantErr bool
	}{
		{
			name: "Test ApplicationAuth StatusReconciler",
			fields: fields{
				BaseReconciler: getBaseReconciler(getApplicationAuthStatus()),
				resource:       getApplicationAuthStatus(),
				reconcileError: fmt.Errorf("test"),
				logger:         logf.Log.WithName("applicationAuth status reconciler test"),
			},
			want:    reconcile.Result{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ApplicationAuthStatusReconciler{
				BaseReconciler: tt.fields.BaseReconciler,
				resource:       tt.fields.resource,
				reconcileError: tt.fields.reconcileError,
				logger:         tt.fields.logger,
			}
			got, err := s.Reconcile()
			if (err != nil) != tt.wantErr {
				t.Errorf("Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reconcile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func getApplicationAuthStatus() (CR *capabilitiesv1beta1.ApplicationAuth) {
	CR = &capabilitiesv1beta1.ApplicationAuth{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: capabilitiesv1beta1.ApplicationAuthSpec{
			ApplicationCRName: "test",
			GenerateSecret:    pointer.Bool(false),
			AuthSecretRef: &corev1.LocalObjectReference{
				Name: "test",
			},
			ProviderAccountRef: nil,
		},
		Status: capabilitiesv1beta1.ApplicationAuthStatus{
			Conditions: common.Conditions{
				common.Condition{
					Type:   capabilitiesv1beta1.ApplicationReadyConditionType,
					Status: corev1.ConditionFalse,
				},
				common.Condition{
					Type:    capabilitiesv1beta1.ApplicationAuthFailedConditionType,
					Status:  corev1.ConditionTrue,
					Message: "test",
				},
			},
		},
	}
	return CR

}

func TestApplicationAuthStatusReconciler_calculateStatus(t *testing.T) {
	type fields struct {
		BaseReconciler *reconcilers.BaseReconciler
		resource       *capabilitiesv1beta1.ApplicationAuth
		reconcileError error
		logger         logr.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    *capabilitiesv1beta1.ApplicationAuthStatus
		wantErr bool
	}{
		{
			name: "Test Completed status ApplicationAuth",
			fields: fields{
				BaseReconciler: getBaseReconciler(),
				resource:       getApplicationAuth(),
				reconcileError: fmt.Errorf("test"),
				logger:         logr.Discard(),
			},
			want: &capabilitiesv1beta1.ApplicationAuthStatus{
				Conditions: common.Conditions{
					common.Condition{
						Type:   capabilitiesv1beta1.ApplicationReadyConditionType,
						Status: corev1.ConditionTrue,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Test Completed status ApplicationAuth generate secret",
			fields: fields{
				BaseReconciler: getBaseReconciler(),
				resource:       getApplicationAuthGenerateSecret(),
				reconcileError: fmt.Errorf("test"),
				logger:         logr.Discard(),
			},
			want: &capabilitiesv1beta1.ApplicationAuthStatus{
				Conditions: common.Conditions{
					common.Condition{
						Type:   capabilitiesv1beta1.ApplicationReadyConditionType,
						Status: corev1.ConditionTrue,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ApplicationAuthStatusReconciler{
				BaseReconciler: tt.fields.BaseReconciler,
				resource:       tt.fields.resource,
				reconcileError: tt.fields.reconcileError,
				logger:         tt.fields.logger,
			}
			got, err := s.calculateStatus()
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Conditions.GetCondition(capabilitiesv1beta1.ApplicationAuthReadyConditionType) == tt.want.Conditions.GetCondition(capabilitiesv1beta1.ApplicationAuthReadyConditionType) {
				if !reflect.DeepEqual(got.Conditions.IsTrueFor(capabilitiesv1beta1.ApplicationAuthReadyConditionType), tt.want.Conditions.IsTrueFor(capabilitiesv1beta1.ApplicationAuthReadyConditionType)) {
					t.Errorf("calculateStatus() got = %v, want %v", got.Conditions.IsTrueFor(capabilitiesv1beta1.ApplicationAuthReadyConditionType), tt.want.Conditions.IsTrueFor(capabilitiesv1beta1.ApplicationAuthReadyConditionType))
				}
				if !reflect.DeepEqual(got.Conditions.IsFalseFor(capabilitiesv1beta1.ApplicationAuthReadyConditionType), tt.want.Conditions.IsFalseFor(capabilitiesv1beta1.ApplicationAuthReadyConditionType)) {
					t.Errorf("calculateStatus() got = %v, want %v", got.Conditions.IsFalseFor(capabilitiesv1beta1.ProxyPromoteConfigReadyConditionType), tt.want.Conditions.IsFalseFor(capabilitiesv1beta1.ProxyPromoteConfigReadyConditionType))
				}
			}
		})
	}
}
