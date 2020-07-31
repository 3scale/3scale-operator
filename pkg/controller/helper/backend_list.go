package helper

import (
	"context"
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// BackendList returns a list of backend custom resources where all elements:
// - Sync state (ensure remote backend exist and in sync)
// - Same 3scale provider Account
func BackendList(ns string, cl client.Client, providerAccount *ProviderAccount, logger logr.Logger) ([]capabilitiesv1beta1.Backend, error) {
	backendList := &capabilitiesv1beta1.BackendList{}
	opts := []controllerclient.ListOption{
		controllerclient.InNamespace(ns),
	}
	err := cl.List(context.TODO(), backendList, opts...)
	logger.V(1).Info("Get list of Backend resources.", "Err", err)
	if err != nil {
		return nil, fmt.Errorf("BackendList: %w", err)
	}
	logger.V(1).Info("Backend resources", "total", len(backendList.Items))

	validBackends := make([]capabilitiesv1beta1.Backend, 0)
	for idx := range backendList.Items {
		// Filter by synchronized
		if !backendList.Items[idx].IsSynced() {
			continue
		}

		backendProviderAccount, err := LookupProviderAccount(cl, ns, backendList.Items[idx].Spec.ProviderAccountRef, logger)
		if err != nil {
			return nil, fmt.Errorf("BackendList: %w", err)
		}

		// Filter by provider account
		if providerAccount.AdminURLStr != backendProviderAccount.AdminURLStr {
			continue
		}
		validBackends = append(validBackends, backendList.Items[idx])
	}

	logger.V(1).Info("Backend valid resources", "total", len(validBackends))
	return validBackends, nil
}
