package handlers

import (
	"reflect"
	"testing"

	logrtesting "github.com/go-logr/logr/testing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appscommon "github.com/3scale/3scale-operator/apis/apps"
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1 "github.com/openshift/api/apps/v1"
)

func TestAPIManagerRoutesEventMapperMap(t *testing.T) {
	apimanagerName := "apimanagerName"
	apimanagerNamespace := "examplenamespace"
	apimanager := &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apimanagerName,
			Namespace: apimanagerNamespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1alpha1.GroupVersion.String(),
			Kind:       appscommon.APIManagerKind,
		},
	}

	zyncQue := &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      component.ZyncQueDeploymentName,
			Namespace: apimanagerNamespace,
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					APIVersion: appsv1alpha1.GroupVersion.String(),
					Kind:       appscommon.APIManagerKind,
					Name:       apimanager.Name,
				},
			},
		},
	}

	objs := []runtime.Object{zyncQue}

	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.GroupVersion, apimanager)
	err := appsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	err = routev1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	// Create a fake client to mock API calls.
	//cl := fake.NewFakeClient(objs...)
	cl := fake.NewFakeClientWithScheme(s, objs...)

	apimanagerRoutesEventMapper := APIManagerRoutesEventMapper{
		K8sClient: cl,
		Logger:    logrtesting.NullLogger{},
	}

	cases := []struct {
		testName string
		input    func() *handler.MapObject
		expected []reconcile.Request
	}{
		{
			testName: "Event with route directly owned by APIManager is converted to an APIManager event",
			input: func() *handler.MapObject {
				routeDirectlyManagedByAPIManager := &routev1.Route{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Route",
						APIVersion: "route.openshift.io/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "routeDirectlyManagedByAPIManager",
						Namespace: apimanagerNamespace,
						OwnerReferences: []metav1.OwnerReference{
							metav1.OwnerReference{
								APIVersion: "v1",
								Kind:       "Secret",
								Name:       "asecret",
							},
							metav1.OwnerReference{
								APIVersion: appsv1alpha1.GroupVersion.String(),
								Kind:       appscommon.APIManagerKind,
								Name:       apimanager.Name,
							},
						},
					},
				}

				return &handler.MapObject{Meta: routeDirectlyManagedByAPIManager, Object: routeDirectlyManagedByAPIManager}
			},
			expected: []reconcile.Request{
				reconcile.Request{NamespacedName: types.NamespacedName{Namespace: apimanagerNamespace, Name: apimanagerName}},
			},
		},
		{
			testName: "Event with route owned by zync-que managed by APIManager is converted to an APIManager event",
			input: func() *handler.MapObject {
				zyncManagedRoute := &routev1.Route{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Route",
						APIVersion: "route.openshift.io/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "routeManagedByZyncQue",
						Namespace: apimanagerNamespace,
						OwnerReferences: []metav1.OwnerReference{
							v1.OwnerReference{
								APIVersion: "v1",
								Kind:       "Secret",
								Name:       "asecret",
							},
							metav1.OwnerReference{
								APIVersion: appsv1.GroupVersion.String(),
								Kind:       "DeploymentConfig",
								Name:       component.ZyncQueDeploymentName,
							},
						},
					},
				}
				return &handler.MapObject{Meta: zyncManagedRoute, Object: zyncManagedRoute}
			},
			expected: []reconcile.Request{
				reconcile.Request{NamespacedName: types.NamespacedName{Namespace: apimanagerNamespace, Name: apimanagerName}},
			},
		},
		{
			testName: "Event with route without OwnerReferences is discarded",
			input: func() *handler.MapObject {
				nonAPIManagerRoute := &routev1.Route{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Route",
						APIVersion: "route.openshift.io/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "aroute",
						Namespace: apimanagerNamespace,
					},
				}
				return &handler.MapObject{Meta: nonAPIManagerRoute, Object: nonAPIManagerRoute}
			},
			expected: nil,
		},
		{
			testName: "Event with route with non-APIManager OwnerReference (directly or indirectly) is discarded",
			input: func() *handler.MapObject {
				nonAPIManagerRoute := &routev1.Route{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Route",
						APIVersion: "route.openshift.io/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "aroute",
						Namespace: apimanagerNamespace,
						OwnerReferences: []v1.OwnerReference{
							metav1.OwnerReference{
								APIVersion: "v1",
								Kind:       "Secret",
								Name:       "asecret",
							},
						},
					},
				}

				return &handler.MapObject{Meta: nonAPIManagerRoute, Object: nonAPIManagerRoute}
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			res := apimanagerRoutesEventMapper.Map(*tc.input())
			if !reflect.DeepEqual(res, tc.expected) {
				subT.Errorf("Unexpected result: %v. Expected: %v", res, tc.expected)
			}
		})
	}
}
