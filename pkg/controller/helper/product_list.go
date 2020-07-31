package helper

import (
	"context"
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ProductList returns a list of product custom resources where all elements:
// - Sync state (ensure remote product exist and in sync)
// - Same 3scale provider Account
func ProductList(ns string, cl client.Client, providerAccount *ProviderAccount, logger logr.Logger) ([]capabilitiesv1beta1.Product, error) {
	productList := &capabilitiesv1beta1.ProductList{}
	opts := []controllerclient.ListOption{
		controllerclient.InNamespace(ns),
	}
	err := cl.List(context.TODO(), productList, opts...)
	logger.V(1).Info("Get list of Product resources.", "Err", err)
	if err != nil {
		return nil, fmt.Errorf("ProductList: %w", err)
	}
	logger.V(1).Info("Product resources", "total", len(productList.Items))

	validProducts := make([]capabilitiesv1beta1.Product, 0)
	for idx := range productList.Items {
		// Filter by synchronized
		if !productList.Items[idx].IsSynced() {
			continue
		}

		productProviderAccount, err := LookupProviderAccount(cl, ns, productList.Items[idx].Spec.ProviderAccountRef, logger)
		if err != nil {
			return nil, fmt.Errorf("BackendList: %w", err)
		}

		// Filter by provider account
		if providerAccount.AdminURLStr != productProviderAccount.AdminURLStr {
			continue
		}
		validProducts = append(validProducts, productList.Items[idx])
	}

	logger.V(1).Info("Product valid resources", "total", len(validProducts))
	return validProducts, nil
}
