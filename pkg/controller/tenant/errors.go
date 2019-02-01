package tenant

import (
	"fmt"
	"k8s.io/apimachinery/pkg/types"
)

// AccessTokenNotAvailableError type
type AccessTokenNotAvailableError struct {
	msg string
}

func (error *AccessTokenNotAvailableError) Error() string {
	return error.msg
}

// NewAccessTokenNotAvailableError constructor
func NewAccessTokenNotAvailableError(errorMsg, accountID string, namespacedName types.NamespacedName) error {
	return &AccessTokenNotAvailableError{fmt.Sprintf("%s: tenant admin's access token not available."+
		" Possible cause: tenant (%s) already exists and admin access token secret (ns: %s, name: %s) "+
		"does not exist", errorMsg, accountID, namespacedName.Namespace, namespacedName.Name)}
}
