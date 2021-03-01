package helper

import (
	"testing"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/common"

	logrtesting "github.com/go-logr/logr/testing"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestBackendList(t *testing.T) {
	ns := "somenamespace"
	providerAccountURLStr := "https://example.com"
	providerAccountToken := "12345"
	anotherProviderSecretName := "anotherSecretName"

	data := map[string]string{
		providerAccountSecretURLFieldName:   providerAccountURLStr,
		providerAccountSecretTokenFieldName: providerAccountToken,
	}
	providerSecret := GetTestSecret(ns, providerAccountDefaultSecretName, data)

	data = map[string]string{
		providerAccountSecretURLFieldName:   "https://other.example.com",
		providerAccountSecretTokenFieldName: providerAccountToken,
	}
	anotherProviderSecret := GetTestSecret(ns, anotherProviderSecretName, data)

	s := scheme.Scheme
	err := capabilitiesv1beta1.AddToScheme(s)
	if err != nil {
		t.Fatalf("Unable to add Apps scheme: (%v)", err)
	}

	cases := []struct {
		testName string
		backend  *capabilitiesv1beta1.Backend
		expected bool
	}{
		{
			"sync'ed backend and same providerAccount",
			&capabilitiesv1beta1.Backend{
				ObjectMeta: metav1.ObjectMeta{Name: "somename", Namespace: ns},
				Spec:       capabilitiesv1beta1.BackendSpec{},
				Status: capabilitiesv1beta1.BackendStatus{
					Conditions: common.Conditions{
						common.Condition{
							Type:   capabilitiesv1beta1.BackendSyncedConditionType,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			true,
		},
		{
			"Not sync'ed backend",
			&capabilitiesv1beta1.Backend{
				ObjectMeta: metav1.ObjectMeta{Name: "somename", Namespace: ns},
				Spec:       capabilitiesv1beta1.BackendSpec{},
				Status: capabilitiesv1beta1.BackendStatus{
					Conditions: common.Conditions{
						common.Condition{
							Type:   capabilitiesv1beta1.BackendSyncedConditionType,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
			false,
		},
		{
			"provider not matching backend",
			&capabilitiesv1beta1.Backend{
				ObjectMeta: metav1.ObjectMeta{Name: "somename", Namespace: ns},
				Spec: capabilitiesv1beta1.BackendSpec{
					ProviderAccountRef: &corev1.LocalObjectReference{
						Name: anotherProviderSecretName,
					},
				},
				Status: capabilitiesv1beta1.BackendStatus{
					Conditions: common.Conditions{
						common.Condition{
							Type:   capabilitiesv1beta1.BackendSyncedConditionType,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			objs := []runtime.Object{anotherProviderSecret, providerSecret, tc.backend}
			cl := fake.NewFakeClient(objs...)
			backendList, err := BackendList(ns, cl, providerAccountURLStr, logrtesting.NullLogger{})
			if err != nil {
				subT.Fatal(err)
			}

			if (len(backendList) == 0) == tc.expected {
				subT.Errorf("backend included: %t, expected: %t", len(backendList) != 0, tc.expected)
			}
		})
	}
}
