package controllers

import (
	"context"
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/common"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/go-logr/logr"
	v1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
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
	"time"
)

func getApplicationCR() (CR *capabilitiesv1beta1.Application) {
	statusID := int64(3)
	CR = &capabilitiesv1beta1.Application{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: capabilitiesv1beta1.ApplicationSpec{
			AccountCRName: &corev1.LocalObjectReference{
				Name: "test",
			},
			ProductCRName: &corev1.LocalObjectReference{
				Name: "test",
			},
			ApplicationPlanName: "test",
			Name:                "test",
			Description:         "test",
			Suspend:             false,
		},
		Status: capabilitiesv1beta1.ApplicationStatus{
			ID: &statusID,
		},
	}
	return CR
}
func getApplicationCRSuspend() (CR *capabilitiesv1beta1.Application) {
	statusID := int64(3)
	CR = &capabilitiesv1beta1.Application{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: capabilitiesv1beta1.ApplicationSpec{
			AccountCRName: &corev1.LocalObjectReference{
				Name: "test",
			},
			ProductCRName: &corev1.LocalObjectReference{
				Name: "test",
			},
			ApplicationPlanName: "test",
			Name:                "test",
			Description:         "test",
			Suspend:             true,
		},
		Status: capabilitiesv1beta1.ApplicationStatus{
			ID: &statusID,
		},
	}
	return CR
}

func getUnknownPlanApplicationCR() (CR *capabilitiesv1beta1.Application) {
	statusID := int64(3)
	CR = &capabilitiesv1beta1.Application{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: capabilitiesv1beta1.ApplicationSpec{
			AccountCRName: &corev1.LocalObjectReference{
				Name: "test",
			},
			ProductCRName: &corev1.LocalObjectReference{
				Name: "test",
			},
			ApplicationPlanName: "unknown",
			Name:                "test",
			Description:         "test",
			Suspend:             false,
		},
		Status: capabilitiesv1beta1.ApplicationStatus{
			ID: &statusID,
		},
	}
	return CR
}

func getApplicationDeleteCR() (CR *capabilitiesv1beta1.Application) {
	statusID := int64(3)
	timestamp2 := time.Now()
	timestamp := metav1.Time{Time: timestamp2}
	CR = &capabilitiesv1beta1.Application{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:                       "test",
			GenerateName:               "",
			Namespace:                  "test",
			SelfLink:                   "",
			UID:                        "",
			ResourceVersion:            "",
			Generation:                 0,
			CreationTimestamp:          metav1.Time{},
			DeletionTimestamp:          &timestamp,
			DeletionGracePeriodSeconds: nil,
			Labels:                     nil,
			Annotations:                nil,
			OwnerReferences:            nil,
			Finalizers:                 nil,
			ClusterName:                "",
			ManagedFields:              nil,
		},
		Spec: capabilitiesv1beta1.ApplicationSpec{
			AccountCRName: &corev1.LocalObjectReference{
				Name: "test",
			},
			ProductCRName: &corev1.LocalObjectReference{
				Name: "test",
			},
			ApplicationPlanName: "test",
			Name:                "test",
			Description:         "test",
			Suspend:             false,
		},
		Status: capabilitiesv1beta1.ApplicationStatus{
			ID: &statusID,
		},
	}
	return CR
}

func getFailedApplicationCR() (CR *capabilitiesv1beta1.Application) {
	CR = &capabilitiesv1beta1.Application{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: capabilitiesv1beta1.ApplicationSpec{
			AccountCRName: &corev1.LocalObjectReference{
				Name: "unknown",
			},
			ProductCRName: &corev1.LocalObjectReference{
				Name: "unknown",
			},
			ApplicationPlanName: "",
			Name:                "",
			Description:         "",
			Suspend:             false,
		},
	}
	return CR
}
func unknowAccountApplicationCR() (CR *capabilitiesv1beta1.Application) {
	CR = &capabilitiesv1beta1.Application{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: capabilitiesv1beta1.ApplicationSpec{
			AccountCRName: &corev1.LocalObjectReference{
				Name: "unknown",
			},
			ProductCRName: &corev1.LocalObjectReference{
				Name: "test",
			},
			ApplicationPlanName: "",
			Name:                "",
			Description:         "",
			Suspend:             false,
		},
	}
	return CR
}
func getApplicationProductCR() (CR *capabilitiesv1beta1.Product) {
	// used for string pointer
	test := "test"

	CR = &capabilitiesv1beta1.Product{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: capabilitiesv1beta1.ProductSpec{
			Name:        "test",
			SystemName:  "test",
			Description: "test",
			ApplicationPlans: map[string]capabilitiesv1beta1.ApplicationPlanSpec{
				"test": {
					Name: &test,
					Limits: []capabilitiesv1beta1.LimitSpec{
						{
							Period: "month",
							Value:  300,
							MetricMethodRef: capabilitiesv1beta1.MetricMethodRefSpec{
								SystemName:        "test",
								BackendSystemName: &test,
							},
						},
					},
					Published: nil,
				},
			},
		},
		Status: capabilitiesv1beta1.ProductStatus{
			ID:                  create(3),
			ProviderAccountHost: "some string",
			ObservedGeneration:  1,
			Conditions: common.Conditions{common.Condition{
				Type:   capabilitiesv1beta1.ProductSyncedConditionType,
				Status: corev1.ConditionStatus(v1.ConditionTrue),
			}},
		},
	}
	return CR
}

func getApplicationProductList() (productList *capabilitiesv1beta1.ProductList) {
	// used for string pointer
	test := "test"

	productList = &capabilitiesv1beta1.ProductList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items: []capabilitiesv1beta1.Product{
			{TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: capabilitiesv1beta1.ProductSpec{
					Name:        "test",
					SystemName:  "test",
					Description: "test",
					ApplicationPlans: map[string]capabilitiesv1beta1.ApplicationPlanSpec{
						"test": {
							Name: &test,
							Limits: []capabilitiesv1beta1.LimitSpec{
								{
									Period: "month",
									Value:  300,
									MetricMethodRef: capabilitiesv1beta1.MetricMethodRefSpec{
										SystemName:        "test",
										BackendSystemName: &test,
									},
								},
							},
							Published: nil,
						},
					},
				},
				Status: capabilitiesv1beta1.ProductStatus{
					ID:                  create(3),
					ProviderAccountHost: "some string",
					ObservedGeneration:  1,
					Conditions:          nil,
				},
			},
		},
	}
	return productList
}

func getProviderAccountRefSecret() (secret *corev1.Secret) {
	secret = &corev1.Secret{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Immutable: nil,
		Data: map[string][]byte{
			"adminURL": []byte("https://3scale-admin.test.3scale.net"),
			"token":    []byte("token"),
		},
		StringData: nil,
		Type:       "",
	}
	return secret
}

func getApplicationDeveloperAccount() (CR *capabilitiesv1beta1.DeveloperAccount) {

	CR = &capabilitiesv1beta1.DeveloperAccount{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: capabilitiesv1beta1.DeveloperAccountSpec{
			OrgName:                "test",
			MonthlyBillingEnabled:  nil,
			MonthlyChargingEnabled: nil,
			ProviderAccountRef: &corev1.LocalObjectReference{
				Name: "test",
			},
		},
		Status: capabilitiesv1beta1.DeveloperAccountStatus{
			ID:                  create(3),
			ProviderAccountHost: "some string",
			Conditions: common.Conditions{
				common.Condition{
					Type:   capabilitiesv1beta1.DeveloperAccountInvalidConditionType,
					Status: corev1.ConditionStatus(v1.ConditionFalse),
				},
			},
		},
	}
	return CR
}

func getApplicationBaseReconciler() (baseReconciler *reconcilers.BaseReconciler) {
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.GroupVersion, getApplicationCR())
	s.AddKnownTypes(appsv1alpha1.GroupVersion, getApplicationProductCR())
	s.AddKnownTypes(appsv1alpha1.GroupVersion, getApiManger())
	s.AddKnownTypes(appsv1alpha1.GroupVersion, getApplicationDeveloperAccount())
	log := logf.Log.WithName("Application status reconciler test")
	ctx := context.TODO()
	// Objects to track in the fake client.
	objs := []runtime.Object{getApplicationCR(), getProviderAccount(), getApiManger(), getApplicationProductList(), getApplicationDeveloperAccount(), getProviderAccountRefSecret()}
	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)
	clientAPIReader := fake.NewFakeClientWithScheme(s, objs...)
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)
	baseReconciler = reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	return baseReconciler
}

func TestApplicationStatusReconciler_Reconcile(t *testing.T) {
	type fields struct {
		BaseReconciler      *reconcilers.BaseReconciler
		applicationResource *capabilitiesv1beta1.Application
		entity              *controllerhelper.ApplicationEntity
		providerAccountHost string
		syncError           error
		logger              logr.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    reconcile.Result
		wantErr bool
	}{
		{
			name: "Test Completed status reconciler for Application",
			fields: fields{
				BaseReconciler:      getApplicationBaseReconciler(),
				applicationResource: getApplicationCR(),
				entity:              nil,
				providerAccountHost: "",
				syncError:           nil,
				logger:              logf.Log.WithName("status reconciler"),
			},
			want:    reconcile.Result{},
			wantErr: false,
		},
		{
			name: "Test Failed status reconciler for Application",
			fields: fields{
				BaseReconciler:      getBaseReconciler(),
				applicationResource: getApplicationCR(),
				entity:              nil,
				providerAccountHost: "",
				syncError:           nil,
				logger:              logf.Log.WithName("status reconciler"),
			},
			want:    reconcile.Result{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ApplicationStatusReconciler{
				BaseReconciler:      tt.fields.BaseReconciler,
				applicationResource: tt.fields.applicationResource,
				entity:              tt.fields.entity,
				providerAccountHost: tt.fields.providerAccountHost,
				syncError:           tt.fields.syncError,
				logger:              tt.fields.logger,
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

func TestApplicationStatusReconciler_calculateStatus(t *testing.T) {
	//ID := int64(3)
	type fields struct {
		BaseReconciler      *reconcilers.BaseReconciler
		applicationResource *capabilitiesv1beta1.Application
		entity              *controllerhelper.ApplicationEntity
		providerAccountHost string
		syncError           error
		logger              logr.Logger
	}
	tests := []struct {
		name   string
		fields fields
		want   *capabilitiesv1beta1.ApplicationStatus
	}{
		{
			name: "Test Ready status Application true",
			fields: fields{
				BaseReconciler:      getApplicationBaseReconciler(),
				applicationResource: getApplicationCR(),
				entity:              nil,
				providerAccountHost: "",
				syncError:           nil,
				logger:              logf.Log.WithName("status reconciler"),
			},
			want: &capabilitiesv1beta1.ApplicationStatus{
				Conditions: common.Conditions{
					common.Condition{
						Type:   capabilitiesv1beta1.ApplicationReadyConditionType,
						Status: corev1.ConditionStatus(v1.ConditionTrue),
					},
				},
			},
		},
		{
			name: "Test Ready status Application true",
			fields: fields{
				BaseReconciler:      getBaseReconciler(),
				applicationResource: getFailedApplicationCR(),
				entity:              nil,
				providerAccountHost: "",
				syncError:           nil,
				logger:              logf.Log.WithName("status reconciler"),
			},
			want: &capabilitiesv1beta1.ApplicationStatus{
				Conditions: common.Conditions{
					common.Condition{
						Type:   capabilitiesv1beta1.ApplicationReadyConditionType,
						Status: corev1.ConditionStatus(v1.ConditionFalse),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ApplicationStatusReconciler{
				BaseReconciler:      tt.fields.BaseReconciler,
				applicationResource: tt.fields.applicationResource,
				entity:              tt.fields.entity,
				providerAccountHost: tt.fields.providerAccountHost,
				syncError:           tt.fields.syncError,
				logger:              tt.fields.logger,
			}
			got := s.calculateStatus()
			if got.Conditions.GetCondition(capabilitiesv1beta1.ApplicationReadyConditionType) == tt.want.Conditions.GetCondition(capabilitiesv1beta1.ApplicationReadyConditionType) {
				if !reflect.DeepEqual(got.Conditions.IsTrueFor(capabilitiesv1beta1.ApplicationReadyConditionType), tt.want.Conditions.IsTrueFor(capabilitiesv1beta1.ApplicationReadyConditionType)) {
					t.Errorf("calculateStatus() = %v, want %v", got.Conditions.GetCondition(capabilitiesv1beta1.ApplicationReadyConditionType), tt.want.Conditions.GetCondition(capabilitiesv1beta1.ApplicationReadyConditionType))
				}
				if !reflect.DeepEqual(got.Conditions.IsFalseFor(capabilitiesv1beta1.ApplicationReadyConditionType), tt.want.Conditions.IsFalseFor(capabilitiesv1beta1.ApplicationReadyConditionType)) {
					t.Errorf("calculateStatus() = %v, want %v", got.Conditions.GetCondition(capabilitiesv1beta1.ApplicationReadyConditionType), tt.want.Conditions.GetCondition(capabilitiesv1beta1.ApplicationReadyConditionType))
				}
			}
		})
	}
}
