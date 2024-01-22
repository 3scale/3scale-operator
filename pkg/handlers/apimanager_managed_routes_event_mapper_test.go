package handlers

import (
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appscommon "github.com/3scale/3scale-operator/apis/apps"
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	k8sappsv1 "k8s.io/api/apps/v1"
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

	zyncQue := &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       reconcilers.DeploymentKind,
			APIVersion: reconcilers.DeploymentAPIVersion,
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
	err := k8sappsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	err = routev1.Install(s)
	if err != nil {
		t.Fatal(err)
	}

	// Create a fake client to mock API calls.
	//cl := fake.NewFakeClient(objs...)
	cl := fake.NewFakeClientWithScheme(s, objs...)

	apimanagerRoutesEventMapper := APIManagerRoutesEventMapper{
		K8sClient: cl,
		Logger:    logr.Discard(),
	}

	cases := []struct {
		testName string
		input    client.Object
		expected []reconcile.Request
	}{
		{
			testName: "Event with route directly owned by APIManager is converted to an APIManager event",
			input: &routev1.Route{
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
			},
			expected: []reconcile.Request{
				reconcile.Request{NamespacedName: types.NamespacedName{Namespace: apimanagerNamespace, Name: apimanagerName}},
			},
		},
		{
			testName: "Event with route owned by zync-que managed by APIManager is converted to an APIManager event",
			input: &routev1.Route{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Route",
					APIVersion: "route.openshift.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "routeManagedByZyncQue",
					Namespace: apimanagerNamespace,
					OwnerReferences: []metav1.OwnerReference{
						metav1.OwnerReference{
							APIVersion: "v1",
							Kind:       "Secret",
							Name:       "asecret",
						},
						metav1.OwnerReference{
							APIVersion: reconcilers.DeploymentAPIVersion,
							Kind:       reconcilers.DeploymentKind,
							Name:       component.ZyncQueDeploymentName,
						},
					},
				},
			},
			expected: []reconcile.Request{
				reconcile.Request{NamespacedName: types.NamespacedName{Namespace: apimanagerNamespace, Name: apimanagerName}},
			},
		},
		{
			testName: "Event with route without OwnerReferences is discarded",
			input: &routev1.Route{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Route",
					APIVersion: "route.openshift.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aroute",
					Namespace: apimanagerNamespace,
				},
			},
			expected: nil,
		},
		{
			testName: "Event with route with non-APIManager OwnerReference (directly or indirectly) is discarded",
			input: &routev1.Route{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Route",
					APIVersion: "route.openshift.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aroute",
					Namespace: apimanagerNamespace,
					OwnerReferences: []metav1.OwnerReference{
						metav1.OwnerReference{
							APIVersion: "v1",
							Kind:       "Secret",
							Name:       "asecret",
						},
					},
				},
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			res := apimanagerRoutesEventMapper.Map(tc.input)
			if !reflect.DeepEqual(res, tc.expected) {
				subT.Errorf("Unexpected result: %v. Expected: %v", res, tc.expected)
			}
		})
	}
}
