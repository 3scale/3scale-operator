package helper

import (
	"context"
	"strings"

	capabilitiesv1alpha1 "github.com/3scale/3scale-operator/apis/capabilities/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

/*
RetrieveTenantCR retrieves tenantCR of a tenant that matches the provider account org name
- providerAccount
- k8client
If the tenantList is empty it will return nil, nil
If tenantCR for given providerAccount org is not present, it will return nil, nil
*/
func RetrieveTenantCR(providerAccount *ProviderAccount, client client.Client) (*capabilitiesv1alpha1.Tenant, error) {
	owner := strings.Split(providerAccount.AdminURLStr[8:len(providerAccount.AdminURLStr)], "-")

	tenantList := &capabilitiesv1alpha1.TenantList{}

	err := client.List(context.TODO(), tenantList)
	if err != nil {
		return nil, err
	}

	for _, tenant := range tenantList.Items {
		if tenant.Spec.OrganizationName == owner[0] {
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
func SetOwnersReference(object controllerutil.Object, client client.Client, tenantCR *capabilitiesv1alpha1.Tenant) error {
	ownerReference := []metav1.OwnerReference{
		{
			APIVersion: tenantCR.APIVersion,
			Kind:       tenantCR.Kind,
			Name:       tenantCR.Name,
			UID:        tenantCR.UID,
		},
	}

	object.SetOwnerReferences(ownerReference)
	err := client.Update(context.TODO(), object)
	if err != nil {
		return err
	}

	return nil
}
