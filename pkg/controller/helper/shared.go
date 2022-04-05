package helper

import (
	"context"

	capabilitiesv1alpha1 "github.com/3scale/3scale-operator/apis/capabilities/v1alpha1"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

/*
RetrieveTenantCR retrieves tenantCR of a tenant that matches the provider account org name
- providerAccount
- k8client
- logger
If the tenantList is empty it will return nil, nil
If tenantCR for given providerAccount org is not present, it will return nil, nil
*/
func RetrieveTenantCR(providerAccount *ProviderAccount, client k8sclient.Client, logger logr.Logger, namespace string) (*capabilitiesv1alpha1.Tenant, error) {
	// Retrieve all product CRs that are under the same ns as the backend CR
	opts := k8sclient.ListOptions{
		Namespace: namespace,
	}

	tenantList := &capabilitiesv1alpha1.TenantList{}
	err := client.List(context.TODO(), tenantList, &opts)
	if err != nil {
		return nil, err
	}

	for _, tenant := range tenantList.Items {
		tenantSecret, err := retrieveTenantSecret(client, &tenant)
		if err != nil {
			return nil, err
		}

		adminURL, ok := tenantSecret.Data[string("adminURL")]
		if !ok {
			return nil, err
		}

		if string(adminURL) == providerAccount.AdminURLStr {
			return &tenant, nil
		}
	}

	return nil, nil
}

/*
SetOwnersReference sets ownersReference in given object
- object
- k8client
- tenantCR
*/
func SetOwnersReference(object controllerutil.Object, client k8sclient.Client, tenantCR *capabilitiesv1alpha1.Tenant) (bool, error) {
	ownerReference := []metav1.OwnerReference{
		{
			APIVersion: tenantCR.APIVersion,
			Kind:       tenantCR.Kind,
			Name:       tenantCR.Name,
			UID:        tenantCR.UID,
		},
	}

	originalSize := len(object.GetOwnerReferences())
	object.SetOwnerReferences(ownerReference)
	newSize := len(object.GetOwnerReferences())

	if originalSize != newSize {
		err := client.Update(context.TODO(), object)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

/*
retrieveTenantSecret retrieves tenants secret
- k8client
- tenantCR
*/
func retrieveTenantSecret(client k8sclient.Client, tenantCR *capabilitiesv1alpha1.Tenant) (*corev1.Secret, error) {
	secret := &corev1.Secret{}

	err := client.Get(context.TODO(), k8sclient.ObjectKey{Name: tenantCR.Spec.TenantSecretRef.Name, Namespace: tenantCR.Spec.TenantSecretRef.Namespace}, secret)
	if err != nil {
		return secret, err
	}

	return secret, nil
}
