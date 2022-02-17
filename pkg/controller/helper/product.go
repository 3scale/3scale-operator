package helper

import (
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	porta_client_pkg "github.com/3scale/3scale-porta-go-client/client"
	"unsafe"
)

/*
FetchTenant fetches tenant from 3scale
- tenant
- portaClient
*/
func FetchProduct(product *capabilitiesv1beta1.Product, portaClient *porta_client_pkg.ThreeScaleClient) (*porta_client_pkg.Product, error) {
	productID := int64(uintptr(unsafe.Pointer(product.Status.ID)))

	if productID == 0 {
		// Product not in status field
		// Product has to be created
		return nil, nil
	}

	productDef, err := portaClient.Product(productID)
	if err != nil && porta_client_pkg.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return productDef, nil
}

/*
ConfirmTenantDeleted confirms that tenant has been deleted from 3scale
- tenant
- portaClient
If tenant is not marked as "scheduled_for_deletion" the function returns false
*/
func ConfirmProductDeleted(product *capabilitiesv1beta1.Product, portaClient *porta_client_pkg.ThreeScaleClient) (bool, error) {
	// fetch product
	productDef, err := FetchProduct(product, portaClient)
	if err != nil {
		return false, err
	}

	// confirm product status
	if productDef.Element.State == "scheduled_for_deletion" {
		return true, nil
	}

	return false, nil
}
