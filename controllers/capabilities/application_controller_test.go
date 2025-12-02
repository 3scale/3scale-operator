package controllers

import (
	"context"
	"net/http/httptest"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/controllers/capabilities/mocks"
	"github.com/3scale/3scale-operator/pkg/apispkg/common"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-porta-go-client/client"
	v1 "github.com/openshift/api/config/v1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	controllerruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	fakectrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func getApplicationPlanListByProductJson() *client.ApplicationPlanJSONList {
	applicationPlanListByProductJson := &client.ApplicationPlanJSONList{
		Plans: []client.ApplicationPlan{
			{
				Element: client.ApplicationPlanItem{
					ID:         1,
					Name:       "test",
					SystemName: "test",
				},
			},
			{
				Element: client.ApplicationPlanItem{
					ID:         2,
					Name:       "test2",
					SystemName: "test2",
				},
			},
		},
	}
	return applicationPlanListByProductJson
}

func getApplicationJson(state string) *client.Application {
	applicationJson := &client.Application{
		ID:                      3,
		CreatedAt:               "",
		UpdatedAt:               "",
		State:                   state,
		UserAccountID:           3,
		FirstTrafficAt:          "",
		FirstDailyTrafficAt:     "",
		EndUserRequired:         false,
		ServiceID:               0,
		UserKey:                 "",
		ProviderVerificationKey: "",
		PlanID:                  1,
		AppName:                 "test",
		Description:             "test",
		ExtraFields:             "",
		Error:                   "",
	}
	return applicationJson
}

func TestApplicationReconciler_Reconcile(t *testing.T) {
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test",
			Name:      "test",
		},
	}

	testCases := []struct {
		name               string
		application        *capabilitiesv1beta1.Application
		account            *capabilitiesv1beta1.DeveloperAccount
		product            *capabilitiesv1beta1.Product
		httpHandlerOptions []mocks.ApplicationAPIHandlerOpt
		testBody           func(t *testing.T, reconciler *reconcilers.BaseReconciler, req reconcile.Request)
	}{
		{
			name: "Account not found - reconciliation should fail",
			application: &capabilitiesv1beta1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: capabilitiesv1beta1.ApplicationSpec{
					AccountCR: &corev1.LocalObjectReference{
						Name: "test",
					},
					ProductCR: &corev1.LocalObjectReference{
						Name: "test",
					},
					ApplicationPlanName: "test",
					Name:                "test",
					Description:         "test",
				},
			},
			product: getProductCR(),
			testBody: func(t *testing.T, r *reconcilers.BaseReconciler, req reconcile.Request) {
				applicationReconciler := ApplicationReconciler{BaseReconciler: r}
				_, err := applicationReconciler.Reconcile(context.Background(), req)
				// No error is returned
				require.Error(t, err)
				require.Equal(t, "developeraccounts.capabilities.3scale.net \"test\" not found", err.Error())

				var currentApplication capabilitiesv1beta1.Application
				require.NoError(t, r.Client().Get(context.Background(), req.NamespacedName, &currentApplication))
				require.Empty(t, currentApplication.Status.ID)

				condition := currentApplication.Status.Conditions.GetCondition(capabilitiesv1beta1.ApplicationReadyConditionType)
				require.Equal(t, corev1.ConditionFalse, condition.Status)
				require.Equal(t, "developeraccounts.capabilities.3scale.net \"test\" not found", condition.Message)
			},
		},
		{
			name: "Product not found - reconciliation should fail",
			application: &capabilitiesv1beta1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: capabilitiesv1beta1.ApplicationSpec{
					AccountCR: &corev1.LocalObjectReference{
						Name: "test",
					},
					ProductCR: &corev1.LocalObjectReference{
						Name: "test",
					},
					ApplicationPlanName: "test",
					Name:                "test",
					Description:         "test",
				},
			},
			account: getApplicationDeveloperAccount(),
			httpHandlerOptions: []mocks.ApplicationAPIHandlerOpt{
				mocks.WithService(3, getApplicationPlanListByProductJson()),
				mocks.WithAccount(3, &client.ApplicationList{Applications: []client.ApplicationElem{
					{Application: *getApplicationJson("live")},
				}}),
			},
			testBody: func(t *testing.T, r *reconcilers.BaseReconciler, req reconcile.Request) {
				ctx := context.Background()
				applicationReconciler := ApplicationReconciler{BaseReconciler: r}
				_, err := applicationReconciler.Reconcile(ctx, req)
				require.NoError(t, err)

				// need to trigger the Reconcile again because the first one only updated the finalizers
				_, err = applicationReconciler.Reconcile(ctx, req)
				require.Error(t, err)

				var currentApplication capabilitiesv1beta1.Application
				require.NoError(t, r.Client().Get(context.Background(), req.NamespacedName, &currentApplication))
				require.Empty(t, currentApplication.Status.ID)

				condition := currentApplication.Status.Conditions.GetCondition(capabilitiesv1beta1.ApplicationReadyConditionType)
				require.Equal(t, corev1.ConditionFalse, condition.Status)
				require.Equal(t, "products.capabilities.3scale.net \"test\" not found", condition.Message)
			},
		},
		{
			name: "Account found - but with no ProviderAccountRef",
			application: &capabilitiesv1beta1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: capabilitiesv1beta1.ApplicationSpec{
					AccountCR: &corev1.LocalObjectReference{
						Name: "test",
					},
					ProductCR: &corev1.LocalObjectReference{
						Name: "test",
					},
					ApplicationPlanName: "test",
					Name:                "test",
					Description:         "test",
				},
			},
			account: &capabilitiesv1beta1.DeveloperAccount{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: capabilitiesv1beta1.DeveloperAccountSpec{
					OrgName:                "test",
					MonthlyBillingEnabled:  nil,
					MonthlyChargingEnabled: nil,
				},
				Status: capabilitiesv1beta1.DeveloperAccountStatus{
					ID:                  ptr.To(int64(3)),
					ProviderAccountHost: "some string",
					Conditions: common.Conditions{
						common.Condition{
							Type:   capabilitiesv1beta1.DeveloperAccountInvalidConditionType,
							Status: corev1.ConditionStatus(v1.ConditionFalse),
						},
					},
				},
			},
			product: getProductCR(),
			httpHandlerOptions: []mocks.ApplicationAPIHandlerOpt{
				mocks.WithService(3, getApplicationPlanListByProductJson()),
				mocks.WithAccount(3, &client.ApplicationList{Applications: []client.ApplicationElem{
					{Application: *getApplicationJson("live")},
				}}),
			},
			testBody: func(t *testing.T, r *reconcilers.BaseReconciler, req reconcile.Request) {
				applicationReconciler := ApplicationReconciler{BaseReconciler: r}
				_, err := applicationReconciler.Reconcile(context.Background(), req)
				require.NoError(t, err)

				// need to trigger the Reconcile again because the first one only updated the finalizers
				_, err = applicationReconciler.Reconcile(context.Background(), req)
				require.Error(t, err)
				require.Equal(t, "LookupProviderAccount: no provider account found", err.Error())

				var currentApplication capabilitiesv1beta1.Application
				require.NoError(t, r.Client().Get(context.Background(), req.NamespacedName, &currentApplication))
				require.Empty(t, currentApplication.Status.ID)
			},
		},
		{
			name: "Product found - but with invalid application plan",
			application: &capabilitiesv1beta1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: capabilitiesv1beta1.ApplicationSpec{
					AccountCR: &corev1.LocalObjectReference{
						Name: "test",
					},
					ProductCR: &corev1.LocalObjectReference{
						Name: "test",
					},
					ApplicationPlanName: "unknown",
					Name:                "test",
					Description:         "test",
				},
			},
			account: getApplicationDeveloperAccount(),
			product: getProductCR(),
			httpHandlerOptions: []mocks.ApplicationAPIHandlerOpt{
				mocks.WithService(3, getApplicationPlanListByProductJson()),
				mocks.WithAccount(3, &client.ApplicationList{Applications: []client.ApplicationElem{
					{Application: *getApplicationJson("live")},
				}}),
			},
			testBody: func(t *testing.T, r *reconcilers.BaseReconciler, req reconcile.Request) {
				ctx := context.Background()
				applicationReconciler := ApplicationReconciler{BaseReconciler: r}
				_, err := applicationReconciler.Reconcile(context.Background(), req)
				// No error is returned
				require.NoError(t, err)

				// need to trigger the Reconcile again because the first one only updated the finalizers
				_, err = applicationReconciler.Reconcile(ctx, req)
				require.Error(t, err)

				var currentApplication capabilitiesv1beta1.Application
				require.NoError(t, r.Client().Get(context.Background(), req.NamespacedName, &currentApplication))
				require.Empty(t, currentApplication.Status.ID)

				condition := currentApplication.Status.Conditions.GetCondition(capabilitiesv1beta1.ApplicationReadyConditionType)
				require.Equal(t, corev1.ConditionFalse, condition.Status)
				require.Equal(t, `task failed SyncApplication: error sync application [test]: plan [unknown] doesnt exist in product [test]`, condition.Message)
			},
		},
		{
			name: "Create application successful",
			application: &capabilitiesv1beta1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: capabilitiesv1beta1.ApplicationSpec{
					AccountCR: &corev1.LocalObjectReference{
						Name: "test",
					},
					ProductCR: &corev1.LocalObjectReference{
						Name: "test",
					},
					Name:                "test",
					Description:         "test",
					ApplicationPlanName: "test",
				},
			},
			account: getApplicationDeveloperAccount(),
			product: getProductCR(),
			httpHandlerOptions: []mocks.ApplicationAPIHandlerOpt{
				mocks.WithService(3, getApplicationPlanListByProductJson()),
				mocks.WithAccount(3, &client.ApplicationList{Applications: []client.ApplicationElem{
					{Application: *getApplicationJson("live")},
				}}),
			},
			testBody: func(t *testing.T, r *reconcilers.BaseReconciler, req reconcile.Request) {
				ctx := context.Background()
				applicationReconciler := ApplicationReconciler{BaseReconciler: r}
				_, err := applicationReconciler.Reconcile(ctx, req)
				require.NoError(t, err)

				t.Log("verifying the Application gets finalizers assigned")
				var application capabilitiesv1beta1.Application
				require.NoError(t, r.Client().Get(ctx, req.NamespacedName, &application))
				require.ElementsMatch(t, application.GetFinalizers(), []string{
					applicationFinalizer,
				})

				// TODO: check owner reference

				// need to trigger the Reconcile again because the first one only updated the finalizers
				_, err = applicationReconciler.Reconcile(ctx, req)
				require.NoError(t, err, "reconciliation returned an error")
				// need to trigger the Reconcile again because the previous updated the Status
				_, err = applicationReconciler.Reconcile(ctx, req)
				require.NoError(t, err, "reconciliation returned an error")

				var currentApplication capabilitiesv1beta1.Application
				require.NoError(t, r.Client().Get(context.Background(), req.NamespacedName, &currentApplication))
				// Check status ID
				require.Equal(t, currentApplication.Status.ID, ptr.To(int64(3)))
				// check annotation
				require.Equal(t, currentApplication.Annotations[applicationIdAnnotation], "3")
				// Check condition
				condition := currentApplication.Status.Conditions.GetCondition(capabilitiesv1beta1.ApplicationReadyConditionType)
				require.Equal(t, corev1.ConditionTrue, condition.Status)
			},
		},
		{
			name: "Delete application successful",
			application: &capabilitiesv1beta1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: capabilitiesv1beta1.ApplicationSpec{
					AccountCR: &corev1.LocalObjectReference{
						Name: "test",
					},
					ProductCR: &corev1.LocalObjectReference{
						Name: "test",
					},
					Name:                "test",
					Description:         "test",
					ApplicationPlanName: "test",
				},
			},
			account: getApplicationDeveloperAccount(),
			product: getProductCR(),
			httpHandlerOptions: []mocks.ApplicationAPIHandlerOpt{
				mocks.WithService(3, getApplicationPlanListByProductJson()),
				mocks.WithAccount(3, &client.ApplicationList{Applications: []client.ApplicationElem{
					{Application: *getApplicationJson("live")},
				}}),
			},
			testBody: func(t *testing.T, r *reconcilers.BaseReconciler, req reconcile.Request) {
				ctx := context.Background()
				applicationReconciler := ApplicationReconciler{BaseReconciler: r}
				_, err := applicationReconciler.Reconcile(ctx, req)
				require.NoError(t, err)

				t.Log("verifying the Application gets finalizers assigned")
				var application capabilitiesv1beta1.Application
				require.NoError(t, r.Client().Get(ctx, req.NamespacedName, &application))
				require.ElementsMatch(t, application.GetFinalizers(), []string{
					applicationFinalizer,
				})

				// TODO: check owner reference

				// need to trigger the Reconcile again because the first one only updated the finalizers
				_, err = applicationReconciler.Reconcile(ctx, req)
				require.NoError(t, err, "reconciliation returned an error")
				// need to trigger the Reconcile again because the previous updated the Status
				_, err = applicationReconciler.Reconcile(ctx, req)
				require.NoError(t, err, "reconciliation returned an error")

				// remove the cr
				require.NoError(t, r.Client().Delete(ctx, &application))
				_, err = applicationReconciler.Reconcile(ctx, req)
				require.NoError(t, err, "reconciliation returned an error")

				var currentApplication capabilitiesv1beta1.Application
				err = r.Client().Get(context.Background(), req.NamespacedName, &currentApplication)
				require.Error(t, err)
				require.Equal(t, err.Error(), "applications.capabilities.3scale.net \"test\" not found")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var httpServer *httptest.Server
			objectsToAdd := []controllerruntimeclient.Object{
				tc.application,
			}
			if tc.account != nil {
				objectsToAdd = append(objectsToAdd, tc.account)
			}
			if tc.product != nil {
				objectsToAdd = append(objectsToAdd, tc.product)
			}

			if tc.httpHandlerOptions != nil {
				httpHandler := mocks.NewApplicationAPIHandler(tc.httpHandlerOptions...)
				httpServer = httptest.NewServer(httpHandler)
				defer httpServer.Close()

				secret := &corev1.Secret{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Immutable: nil,
					Data: map[string][]byte{
						"adminURL": []byte(httpServer.URL),
						"token":    []byte("token"),
					},
					StringData: nil,
					Type:       "",
				}

				objectsToAdd = append(objectsToAdd, secret)
			}

			s := scheme.Scheme
			_ = capabilitiesv1beta1.AddToScheme(s)
			_ = appsv1alpha1.AddToScheme(s)

			fakeClient := fakectrlruntimeclient.
				NewClientBuilder().
				WithScheme(scheme.Scheme).
				WithObjects(objectsToAdd...).
				WithStatusSubresource(objectsToAdd...).
				Build()

			log := logf.Log.WithName("Application reconciler test")
			clientset := fakeclientset.NewSimpleClientset()
			recorder := record.NewFakeRecorder(10000)
			baseReconciler := reconcilers.NewBaseReconciler(context.TODO(), fakeClient, s, fakeClient, log, clientset.Discovery(), recorder)
			tc.testBody(t, baseReconciler, req)
		})
	}
}
