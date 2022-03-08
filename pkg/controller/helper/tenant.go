package helper

import (
	porta_client_pkg "github.com/3scale/3scale-porta-go-client/client"
)

/*
FetchTenant fetches tenant from 3scale
- tenantID
- portaClient
*/
func FetchTenant(tenantID int64, portaClient *porta_client_pkg.ThreeScaleClient) (*porta_client_pkg.Tenant, error) {
	if tenantID == 0 {
		return nil, nil
	}

	tenantDef, err := portaClient.ShowTenant(tenantID)
	if err != nil && porta_client_pkg.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return tenantDef, nil
}
