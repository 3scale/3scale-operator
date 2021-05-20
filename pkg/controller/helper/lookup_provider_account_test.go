package helper

import (
	"errors"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"

	logrtesting "github.com/go-logr/logr/testing"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestLookupProviderAccountSecretReference(t *testing.T) {
	ns := "some_namespace"
	secretName := "provideraccount"
	providerAccountURLStr := "https://example.com"
	providerAccountToken := "12345"

	data := map[string]string{
		providerAccountSecretURLFieldName:   providerAccountURLStr,
		providerAccountSecretTokenFieldName: providerAccountToken,
	}
	providerSecret := GetTestSecret(ns, secretName, data)

	providerAccountRef := &corev1.LocalObjectReference{
		Name: secretName,
	}

	cl := fake.NewFakeClient(providerSecret)

	providerAccount, err := LookupProviderAccount(cl, ns, providerAccountRef, logrtesting.NullLogger{})
	ok(t, err)
	assert(t, providerAccount != nil, "provider account returned nil")
	equals(t, providerAccount.AdminURLStr, providerAccountURLStr)
	equals(t, providerAccount.Token, providerAccountToken)
}

func TestLookupProviderAccountDefaultSecret(t *testing.T) {
	ns := "some_namespace"
	providerAccountURLStr := "https://example.com"
	providerAccountToken := "12345"

	data := map[string]string{
		providerAccountSecretURLFieldName:   providerAccountURLStr,
		providerAccountSecretTokenFieldName: providerAccountToken,
	}
	providerSecret := GetTestSecret(ns, providerAccountDefaultSecretName, data)

	cl := fake.NewFakeClient(providerSecret)

	providerAccount, err := LookupProviderAccount(cl, ns, nil, logrtesting.NullLogger{})
	ok(t, err)
	assert(t, providerAccount != nil, "provider account returned nil")
	equals(t, providerAccount.AdminURLStr, providerAccountURLStr)
	equals(t, providerAccount.Token, providerAccountToken)
}

func TestLookupProviderAccountLocal3scale(t *testing.T) {
	ns := "some_namespace"
	providerAccountToken := "12345"
	tenantName := "testaccount"

	s := scheme.Scheme
	err := appsv1alpha1.AddToScheme(s)

	apimanager := &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{Name: "somename", Namespace: ns},
		Spec: appsv1alpha1.APIManagerSpec{
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				WildcardDomain: "example.com",
				TenantName:     &tenantName,
			},
		},
	}

	data := map[string]string{
		component.SystemSecretSystemSeedAdminAccessTokenFieldName: providerAccountToken,
	}
	secret := GetTestSecret(ns, component.SystemSecretSystemSeedSecretName, data)

	cl := fake.NewFakeClient(apimanager, secret)

	providerAccount, err := LookupProviderAccount(cl, ns, nil, logrtesting.NullLogger{})
	ok(t, err)
	assert(t, providerAccount != nil, "provider account returned nil")
	equals(t, providerAccount.AdminURLStr, "https://testaccount-admin.example.com")
	equals(t, providerAccount.Token, providerAccountToken)
}

func TestLookupProviderAccountNotFoundError(t *testing.T) {
	ns := "some_namespace"
	cl := fake.NewFakeClient()
	_, err := LookupProviderAccount(cl, ns, nil, logrtesting.NullLogger{})
	equals(t, errors.New("LookupProviderAccount: no provider account found"), err)
}
