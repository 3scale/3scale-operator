package helper

import (
	capabilitiesv1alpha1 "github.com/3scale/3scale-operator/apis/capabilities/v1alpha1"
	porta_client_pkg "github.com/3scale/3scale-porta-go-client/client"
)

func FetchTenant(tenant *capabilitiesv1alpha1.Tenant, portaClient *porta_client_pkg.ThreeScaleClient) (*porta_client_pkg.Tenant, error) {
	if tenant.Status.TenantId == 0 {
		// tenantId not in status field
		// Tenant has to be created
		return nil, nil
	}

	tenantDef, err := portaClient.ShowTenant(tenant.Status.TenantId)
	if err != nil && porta_client_pkg.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return tenantDef, nil
}

func DeleteTenant(tenant *capabilitiesv1alpha1.Tenant, portaClient *porta_client_pkg.ThreeScaleClient) error {
	err := portaClient.DeleteTenant(tenant.Status.TenantId)
	if err != nil {
		return err
	}

	return nil
}

func ConfirmTenantDeleted(tenant *capabilitiesv1alpha1.Tenant, portaClient *porta_client_pkg.ThreeScaleClient) (bool, error) {
	// fetch tenant
	tenantDef, err := FetchTenant(tenant, portaClient)
	if err != nil {
		return false, err
	}

	// confirm tenant status
	if tenantDef.Signup.Account.State == "scheduled_for_deletion" {
		return true, nil
	} else {
		return false, nil
	}
}
