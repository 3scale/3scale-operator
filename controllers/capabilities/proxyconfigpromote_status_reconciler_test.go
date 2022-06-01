package controllers

import (
	"context"
	"fmt"
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func getApiManger() (apimanager *appsv1alpha1.APIManager) {
	var (
		name                 = "test"
		namespace            = "test"
		wildcardDomain       = "test.3scale.net"
		appLabel             = "someLabel"
		tenantName           = "someTenant"
		trueValue            = true
		oneValue       int64 = 1
	)

	apimanager = &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				AppLabel:                     &appLabel,
				ImageStreamTagImportInsecure: &trueValue,
				WildcardDomain:               wildcardDomain,
				TenantName:                   &tenantName,
				ResourceRequirementsEnabled:  &trueValue,
			},
			Backend: &appsv1alpha1.BackendSpec{
				ListenerSpec: &appsv1alpha1.BackendListenerSpec{Replicas: &oneValue},
				WorkerSpec:   &appsv1alpha1.BackendWorkerSpec{Replicas: &oneValue},
				CronSpec:     &appsv1alpha1.BackendCronSpec{Replicas: &oneValue},
			},
			PodDisruptionBudget: &appsv1alpha1.PodDisruptionBudgetSpec{Enabled: true},
		},
	}
	return apimanager
}
func getProxyConfigPromoteCR() (CR *capabilitiesv1beta1.ProxyConfigPromote) {
	CR = &capabilitiesv1beta1.ProxyConfigPromote{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: capabilitiesv1beta1.ProxyConfigPromoteSpec{
			ProductCRName: "test",
		},
	}
	return CR
}

func getBaseReconciler() (baseReconciler *reconcilers.BaseReconciler) {
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.GroupVersion, getProxyConfigPromoteCR())
	s.AddKnownTypes(appsv1alpha1.GroupVersion, getProductCR())
	s.AddKnownTypes(appsv1alpha1.GroupVersion, &appsv1alpha1.APIManagerList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items:    nil,
	})
	s.AddKnownTypes(appsv1alpha1.GroupVersion, getApiManger())
	log := logf.Log.WithName("proxyPromoteConfig status reconciler test")
	ctx := context.TODO()
	// Objects to track in the fake client.
	objs := []runtime.Object{getProxyConfigPromoteCR(), getProviderAccount(), getApiManger(), getProductList()}
	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)
	clientAPIReader := fake.NewFakeClientWithScheme(s, objs...)
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)
	baseReconciler = reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	return baseReconciler
}

func TestProxyConfigPromoteStatusReconciler_calculateStatus(t *testing.T) {
	type fields struct {
		BaseReconciler          *reconcilers.BaseReconciler
		resource                *capabilitiesv1beta1.ProxyConfigPromote
		state                   string
		productID               string
		latestProductionVersion int
		latestStagingVersion    int
		reconcileError          error
		logger                  logr.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    *capabilitiesv1beta1.ProxyConfigPromoteStatus
		wantErr bool
	}{
		{
			name: "Test Completed status ProxyPromoteConfig",
			fields: fields{
				BaseReconciler:          getBaseReconciler(),
				resource:                getProxyConfigPromoteCR(),
				state:                   "Completed",
				productID:               "3",
				latestProductionVersion: 1,
				latestStagingVersion:    1,
				reconcileError:          fmt.Errorf("test"),
				logger:                  nil,
			},
			want: &capabilitiesv1beta1.ProxyConfigPromoteStatus{
				ProductId:               "3",
				LatestProductionVersion: 1,
				LatestStagingVersion:    1,
				Conditions: common.Conditions{
					common.Condition{
						Type:               "Ready",
						Status:             "true",
						Reason:             "",
						Message:            "",
						LastTransitionTime: metav1.Time{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Test Failed status ProxyPromoteConfig",
			fields: fields{
				BaseReconciler:          getBaseReconciler(),
				resource:                getProxyConfigPromoteCR(),
				state:                   "Failed",
				productID:               "3",
				latestProductionVersion: 1,
				latestStagingVersion:    1,
				reconcileError:          fmt.Errorf("test"),
				logger:                  nil,
			},
			want: &capabilitiesv1beta1.ProxyConfigPromoteStatus{
				ProductId:               "3",
				LatestProductionVersion: 1,
				LatestStagingVersion:    1,
				Conditions: common.Conditions{
					common.Condition{
						Type:               "Failed",
						Status:             "true",
						Reason:             "",
						Message:            "",
						LastTransitionTime: metav1.Time{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Test Invalid status ProxyPromoteConfig",
			fields: fields{
				BaseReconciler:          getBaseReconciler(),
				resource:                getProxyConfigPromoteCR(),
				state:                   "Invalid",
				productID:               "3",
				latestProductionVersion: 1,
				latestStagingVersion:    1,
				reconcileError:          fmt.Errorf("test"),
				logger:                  nil,
			},
			want: &capabilitiesv1beta1.ProxyConfigPromoteStatus{
				ProductId:               "3",
				LatestProductionVersion: 1,
				LatestStagingVersion:    1,
				Conditions: common.Conditions{
					common.Condition{
						Type:               "Invalid",
						Status:             "true",
						Reason:             "",
						Message:            "",
						LastTransitionTime: metav1.Time{},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ProxyConfigPromoteStatusReconciler{
				BaseReconciler:          tt.fields.BaseReconciler,
				resource:                tt.fields.resource,
				state:                   tt.fields.state,
				productID:               tt.fields.productID,
				latestProductionVersion: tt.fields.latestProductionVersion,
				latestStagingVersion:    tt.fields.latestStagingVersion,
				reconcileError:          tt.fields.reconcileError,
				logger:                  tt.fields.logger,
			}
			got, err := s.calculateStatus()
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.ProductId, tt.want.ProductId) {
				t.Errorf("calculateStatus() got = %v, want %v", got.ProductId, tt.want.ProductId)
			}
			if !reflect.DeepEqual(got.LatestProductionVersion, tt.want.LatestProductionVersion) {
				t.Errorf("calculateStatus() got = %v, want %v", got.LatestProductionVersion, tt.want.LatestProductionVersion)
			}
			if !reflect.DeepEqual(got.LatestStagingVersion, tt.want.LatestStagingVersion) {
				t.Errorf("calculateStatus() got = %v, want %v", got.LatestStagingVersion, tt.want.LatestStagingVersion)
			}
			if got.Conditions.GetCondition("Ready") == tt.want.Conditions.GetCondition("Ready") {
				if !reflect.DeepEqual(got.Conditions.IsTrueFor("Ready"), tt.want.Conditions.IsTrueFor("Ready")) {
					t.Errorf("calculateStatus() got = %v, want %v", got.Conditions.IsTrueFor("Ready"), tt.want.Conditions.IsTrueFor("Ready"))
				}
			}
			if got.Conditions.GetCondition("Failed") == tt.want.Conditions.GetCondition("Failed") {
				if !reflect.DeepEqual(got.Conditions.IsTrueFor("Failed"), tt.want.Conditions.IsTrueFor("Failed")) {
					t.Errorf("calculateStatus() got = %v, want %v", got.Conditions.IsTrueFor("Failed"), tt.want.Conditions.IsTrueFor("Failed"))
				}
			}
			if got.Conditions.GetCondition("Invalid") == tt.want.Conditions.GetCondition("Invalid") {
				if !reflect.DeepEqual(got.Conditions.IsTrueFor("Invalid"), tt.want.Conditions.IsTrueFor("Invalid")) {
					t.Errorf("calculateStatus() got = %v, want %v", got.Conditions.IsTrueFor("Invalid"), tt.want.Conditions.IsTrueFor("Invalid"))
				}
			}
		})
	}
}

func TestProxyConfigPromoteStatusReconciler_Reconcile(t *testing.T) {
	type fields struct {
		BaseReconciler          *reconcilers.BaseReconciler
		resource                *capabilitiesv1beta1.ProxyConfigPromote
		state                   string
		productID               string
		latestProductionVersion int
		latestStagingVersion    int
		reconcileError          error
		logger                  logr.Logger
	}
	var tests = []struct {
		name    string
		fields  fields
		want    reconcile.Result
		wantErr bool
	}{
		{
			name: "Test StatusReconciler",
			fields: fields{
				BaseReconciler:          getBaseReconciler(),
				resource:                getProxyConfigPromoteCR(),
				state:                   "Completed",
				productID:               "3",
				latestProductionVersion: 1,
				latestStagingVersion:    1,
				reconcileError:          fmt.Errorf("test"),
				logger:                  logf.Log.WithName("proxyPromoteConfig status reconciler test"),
			},
			want:    reconcile.Result{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ProxyConfigPromoteStatusReconciler{
				BaseReconciler:          tt.fields.BaseReconciler,
				resource:                tt.fields.resource,
				state:                   tt.fields.state,
				productID:               tt.fields.productID,
				latestProductionVersion: tt.fields.latestProductionVersion,
				latestStagingVersion:    tt.fields.latestStagingVersion,
				reconcileError:          tt.fields.reconcileError,
				logger:                  tt.fields.logger,
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
