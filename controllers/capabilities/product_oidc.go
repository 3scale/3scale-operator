package controllers

import (
	"fmt"
)

func (t *ProductThreescaleReconciler) syncOIDCConfiguration(_ interface{}) error {
	desiredSpec := t.resource.Spec.OIDCSpec()
	if desiredSpec == nil || desiredSpec.AuthenticationFlow == nil {
		return nil
	}

	existing, err := t.productEntity.OIDCConfiguration()
	if err != nil {
		return fmt.Errorf("Error sync product [%s] oidc configuration: %w", t.resource.Spec.SystemName, err)
	}

	newOIDCConf := *existing

	updated := false

	if newOIDCConf.Element.StandardFlowEnabled != desiredSpec.AuthenticationFlow.StandardFlowEnabled {
		newOIDCConf.Element.StandardFlowEnabled = desiredSpec.AuthenticationFlow.StandardFlowEnabled
		updated = true
	}

	if newOIDCConf.Element.ImplicitFlowEnabled != desiredSpec.AuthenticationFlow.ImplicitFlowEnabled {
		newOIDCConf.Element.ImplicitFlowEnabled = desiredSpec.AuthenticationFlow.ImplicitFlowEnabled
		updated = true
	}

	if newOIDCConf.Element.ServiceAccountsEnabled != desiredSpec.AuthenticationFlow.ServiceAccountsEnabled {
		newOIDCConf.Element.ServiceAccountsEnabled = desiredSpec.AuthenticationFlow.ServiceAccountsEnabled
		updated = true
	}

	if newOIDCConf.Element.DirectAccessGrantsEnabled != desiredSpec.AuthenticationFlow.DirectAccessGrantsEnabled {
		newOIDCConf.Element.DirectAccessGrantsEnabled = desiredSpec.AuthenticationFlow.DirectAccessGrantsEnabled
		updated = true
	}

	if updated {
		err := t.productEntity.UpdateOIDCConfiguration(&newOIDCConf)
		if err != nil {
			return fmt.Errorf("Error sync product [%s] oidc configuration: %w", t.resource.Spec.SystemName, err)
		}
	}

	return nil
}
