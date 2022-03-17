package helper

import (
	"fmt"

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

/*
FindUser finds users by email and username
- portaClient
- tenantID
- email
- username
*/
func FindUser(portaClient *porta_client_pkg.ThreeScaleClient, tenantID int64, email, username string) (*porta_client_pkg.DeveloperUser, error) {
	// Any state
	// Any role (admin, member)
	filterParams := porta_client_pkg.Params{}
	// for the master account, the DeveloperUsers of one account are the users
	// of the provider account associated the master account
	userList, err := portaClient.ListDeveloperUsers(tenantID, filterParams)
	if err != nil {
		return nil, err
	}

	for idx, user := range userList.Items {
		if user.Element.Email != nil && *user.Element.Email == email &&
			user.Element.Username != nil && *user.Element.Username == username {
			return &userList.Items[idx], nil
		}
	}

	return nil, nil
}

/*
CreateAdminUser creates and active admin user
- portaClient
- tenantID
- email
- username
*/
func CreateAdminUser(portaClient *porta_client_pkg.ThreeScaleClient, tenantID int64, email, username string) (*porta_client_pkg.DeveloperUser, error) {
	desiredAdmin := &porta_client_pkg.DeveloperUser{
		Element: porta_client_pkg.DeveloperUserItem{
			Email:    &email,
			Username: &username,
		},
	}

	admin, err := portaClient.CreateDeveloperUser(tenantID, desiredAdmin)
	if err != nil {
		return nil, err
	}

	if admin.Element.ID == nil {
		return nil, fmt.Errorf("admin returned nil ID for tenantID %d", tenantID)
	}

	admin, err = portaClient.ChangeRoleToAdminDeveloperUser(tenantID, *admin.Element.ID)
	if err != nil {
		return nil, err
	}

	admin, err = portaClient.ActivateDeveloperUser(tenantID, *admin.Element.ID)
	if err != nil {
		return nil, err
	}

	return admin, nil
}
