package helper

import (
	"testing"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/common"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestProductList(t *testing.T) {
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
		product  *capabilitiesv1beta1.Product
		expected bool
	}{
		{
			"sync'ed product and same providerAccount",
			&capabilitiesv1beta1.Product{
				ObjectMeta: metav1.ObjectMeta{Name: "somename", Namespace: ns},
				Spec:       capabilitiesv1beta1.ProductSpec{},
				Status: capabilitiesv1beta1.ProductStatus{
					Conditions: common.Conditions{
						common.Condition{
							Type:   capabilitiesv1beta1.ProductSyncedConditionType,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			true,
		},
		{
			"Not sync'ed product",
			&capabilitiesv1beta1.Product{
				ObjectMeta: metav1.ObjectMeta{Name: "somename", Namespace: ns},
				Spec:       capabilitiesv1beta1.ProductSpec{},
				Status: capabilitiesv1beta1.ProductStatus{
					Conditions: common.Conditions{
						common.Condition{
							Type:   capabilitiesv1beta1.ProductSyncedConditionType,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
			false,
		},
		{
			"provider not matching product",
			&capabilitiesv1beta1.Product{
				ObjectMeta: metav1.ObjectMeta{Name: "somename", Namespace: ns},
				Spec: capabilitiesv1beta1.ProductSpec{
					ProviderAccountRef: &corev1.LocalObjectReference{
						Name: anotherProviderSecretName,
					},
				},
				Status: capabilitiesv1beta1.ProductStatus{
					Conditions: common.Conditions{
						common.Condition{
							Type:   capabilitiesv1beta1.ProductSyncedConditionType,
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
			cl := fake.NewFakeClient(anotherProviderSecret, providerSecret, tc.product)
			productList, err := ProductList(ns, cl, providerAccountURLStr, logr.Discard())
			if err != nil {
				subT.Fatal(err)
			}

			if (len(productList) == 0) == tc.expected {
				subT.Errorf("product included: %t, expected: %t", len(productList) != 0, tc.expected)
			}
		})
	}
}

func TestFindProductBySystemName(t *testing.T) {
	ns := "somenamespace"
	productList := []capabilitiesv1beta1.Product{
		capabilitiesv1beta1.Product{
			ObjectMeta: metav1.ObjectMeta{Name: "A", Namespace: ns},
			Spec:       capabilitiesv1beta1.ProductSpec{SystemName: "A"},
		},
		capabilitiesv1beta1.Product{
			ObjectMeta: metav1.ObjectMeta{Name: "B", Namespace: ns},
			Spec:       capabilitiesv1beta1.ProductSpec{SystemName: "B"},
		},
		capabilitiesv1beta1.Product{
			ObjectMeta: metav1.ObjectMeta{Name: "C", Namespace: ns},
			Spec:       capabilitiesv1beta1.ProductSpec{SystemName: "C"},
		},
	}

	cases := []struct {
		testName    string
		systemName  string
		expectedRes int
	}{
		{"Looking for non existing", "not existing", -1},
		{"Looking for A", "A", 0},
		{"Looking for B", "B", 1},
		{"Looking for C", "C", 2},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			res := FindProductBySystemName(productList, tc.systemName)
			equals(subT, tc.expectedRes, res)
		})
	}
}
