package helper

import (
	"context"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	pkghelper "github.com/3scale/3scale-operator/pkg/helper"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeveloperUserListFilter defines the function signature to implement filter
type DeveloperUserListFilter func(developerUser *capabilitiesv1beta1.DeveloperUser) (bool, error)

// FindDeveloperUserList returns a list of developerusers custom resources:
// - Provide []ListOption to modify options for a list request.
// - Filter results with a set of DeveloperUserListFilter methods
func FindDeveloperUserList(logger logr.Logger, cl client.Client, queryOpts []client.ListOption, filters ...DeveloperUserListFilter) ([]capabilitiesv1beta1.DeveloperUser, error) {
	logger.V(1).Info("FindDeveloperUserList: start")
	list := &capabilitiesv1beta1.DeveloperUserList{}

	err := cl.List(context.TODO(), list, queryOpts...)
	logger.V(1).Info("FindDeveloperUserList", "err", err)
	if err != nil {
		return nil, err
	}
	logger.V(1).Info("FindDeveloperUserList", "total", len(list.Items))

	filteredList := make([]capabilitiesv1beta1.DeveloperUser, 0)

	for idx := range list.Items {
		// Loop through each filter
		filterResult := make([]bool, len(filters))
		for filterIdx, filter := range filters {
			valid, err := filter(&list.Items[idx])
			if err != nil {
				return nil, err
			}
			filterResult[filterIdx] = valid
		}

		if pkghelper.All(filterResult) {
			filteredList = append(filteredList, list.Items[idx])
		}
	}

	logger.V(1).Info("FindDeveloperUserList", "valid total", len(filteredList))
	return filteredList, nil
}

// DeveloperUserProviderAccountFilter implements a response filter by providerAccount
func DeveloperUserProviderAccountFilter(cl client.Client, ns, providerAccountURLStr string, logger logr.Logger) DeveloperUserListFilter {
	return func(developerUser *capabilitiesv1beta1.DeveloperUser) (bool, error) {
		providerAccount, err := LookupProviderAccount(cl, ns, developerUser.Spec.ProviderAccountRef, logger)
		if err != nil {
			return false, err
		}

		return providerAccount.AdminURLStr == providerAccountURLStr, nil
	}
}
