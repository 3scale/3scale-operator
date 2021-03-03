package helper

import (
	"testing"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"

	logrtesting "github.com/go-logr/logr/testing"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestFindDeveloperUserList(t *testing.T) {
	ns := "some_namespace"
	adminRole := "admin"
	memberRole := "member"
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

	devUser1 := &capabilitiesv1beta1.DeveloperUser{
		ObjectMeta: metav1.ObjectMeta{Name: "devUser1", Namespace: ns},
		Spec: capabilitiesv1beta1.DeveloperUserSpec{
			Username: "devUser1", Email: "devUser1@example.com", Role: &adminRole,
		},
	}

	devUser2 := &capabilitiesv1beta1.DeveloperUser{
		ObjectMeta: metav1.ObjectMeta{Name: "devUser2", Namespace: ns},
		Spec: capabilitiesv1beta1.DeveloperUserSpec{
			Username: "devUser2", Email: "devUser2@example.com", Role: &memberRole,
		},
	}

	devUser3 := &capabilitiesv1beta1.DeveloperUser{
		ObjectMeta: metav1.ObjectMeta{Name: "devUser3", Namespace: ns},
		Spec: capabilitiesv1beta1.DeveloperUserSpec{
			Username: "devUser3", Email: "devUser3@example.com", Role: &adminRole,
			ProviderAccountRef: &corev1.LocalObjectReference{Name: anotherProviderSecretName},
		},
	}

	cl := fake.NewFakeClient(anotherProviderSecret, providerSecret, devUser1, devUser2, devUser3)

	// Find only not admin users from the same provider account
	adminRoleFilter := func(developerUser *capabilitiesv1beta1.DeveloperUser) (bool, error) {
		return developerUser.Spec.Role != nil && *developerUser.Spec.Role == "admin", nil
	}

	providerAccountFilter := DeveloperUserProviderAccountFilter(cl, ns, providerAccountURLStr, logrtesting.NullLogger{})

	adminUserList, err := FindDeveloperUserList(logrtesting.NullLogger{}, cl, nil, adminRoleFilter, providerAccountFilter)
	ok(t, err)
	equals(t, 1, len(adminUserList))
	equals(t, "devUser1", adminUserList[0].Spec.Username)
	equals(t, "admin", *adminUserList[0].Spec.Role)
	assert(t, adminUserList[0].Spec.ProviderAccountRef == nil, "wrong provider account")
}
