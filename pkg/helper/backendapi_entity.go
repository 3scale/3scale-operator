package helper

import (
	"fmt"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"

	"github.com/go-logr/logr"
)

type BackendAPIEntity struct {
	client        *threescaleapi.ThreeScaleClient
	backendAPIObj *threescaleapi.BackendApi
	logger        logr.Logger
}

func NewBackendAPIEntity(backendAPIObj *threescaleapi.BackendApi, client *threescaleapi.ThreeScaleClient, logger logr.Logger) *BackendAPIEntity {
	return &BackendAPIEntity{
		backendAPIObj: backendAPIObj,
		client:        client,
		logger:        logger.WithValues("BackendAPI", backendAPIObj.Element.ID),
	}
}

func (b *BackendAPIEntity) ID() int64 {
	return b.backendAPIObj.Element.ID
}

func (b *BackendAPIEntity) Update(params threescaleapi.Params) error {
	b.logger.V(1).Info("Update", "params", params)
	updatedBackendAPI, err := b.client.UpdateBackendApi(b.backendAPIObj.Element.ID, params)
	if err != nil {
		return fmt.Errorf("backend [%s] update request: %w", b.backendAPIObj.Element.SystemName, err)
	}

	b.backendAPIObj = updatedBackendAPI

	return nil
}

func (b *BackendAPIEntity) Methods() (*threescaleapi.MethodList, error) {
	b.logger.V(1).Info("Methods")
	// TODO
	return nil, nil
}

func (b *BackendAPIEntity) CreateMethod(params threescaleapi.Params) error {
	b.logger.V(1).Info("CreateMethod", "params", params)
	// TODO
	return nil
}

func (b *BackendAPIEntity) DeleteMethod(id int64) error {
	b.logger.V(1).Info("DeleteMethod", "method ID", id)
	// TODO
	return nil
}

func (b *BackendAPIEntity) UpdateMethod(id int64, params threescaleapi.Params) error {
	b.logger.V(1).Info("UpdateMethod", "method ID", id, "params", params)
	// TODO
	return nil
}
