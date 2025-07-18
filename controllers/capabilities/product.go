package controllers

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/helper"
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
)

func (t *ProductThreescaleReconciler) syncProduct(_ interface{}) error {
	params := threescaleapi.Params{}

	if t.productEntity.Name() != t.resource.Spec.Name {
		params["name"] = t.resource.Spec.Name
	}

	if t.productEntity.Description() != t.resource.Spec.Description {
		params["description"] = t.resource.Spec.Description
	}

	specDeploymentOption := t.resource.Spec.DeploymentOption()
	if specDeploymentOption != nil {
		if t.productEntity.DeploymentOption() != *specDeploymentOption {
			params["deployment_option"] = *specDeploymentOption
		}
	} // only update deployment_option when set in the CR

	specAuthMode := t.resource.Spec.AuthenticationMode()
	if specAuthMode != nil {
		if t.productEntity.BackendVersion() != *specAuthMode {
			params["backend_version"] = *specAuthMode
		}
	} // only update backend_version when set in the CR

	if !helper.ManagedByOperatorAnnotationExists(t.productEntity.Annotations()) {
		for k, v := range helper.ManagedByOperatorAnnotation() {
			params[k] = v
		}
	}

	if len(params) > 0 {
		err := t.productEntity.Update(params)
		if err != nil {
			return fmt.Errorf("error sync product [%s;%d]: %w", t.resource.Spec.SystemName, t.productEntity.ID(), err)
		}
	}

	return nil
}
