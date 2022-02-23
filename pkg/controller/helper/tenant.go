package helper

import (
	capabilitiesv1alpha1 "github.com/3scale/3scale-operator/apis/capabilities/v1alpha1"
	porta_client_pkg "github.com/3scale/3scale-porta-go-client/client"
)

/*
FetchTenant fetches tenant from 3scale
- tenant
- portaClient
*/
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
