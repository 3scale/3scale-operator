package helper

import (
	"fmt"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
)

type BackendAPIEntity struct {
	client        *threescaleapi.ThreeScaleClient
	backendAPIObj *threescaleapi.BackendApi
}

func NewBackendAPIEntity(backendAPIObj *threescaleapi.BackendApi, client *threescaleapi.ThreeScaleClient) *BackendAPIEntity {
	return &BackendAPIEntity{
		backendAPIObj: backendAPIObj,
		client:        client,
	}
}

func (b *BackendAPIEntity) ID() int64 {
	return b.backendAPIObj.Element.ID
}

func (b *BackendAPIEntity) Update(params threescaleapi.Params) error {
	updatedBackendAPI, err := b.client.UpdateBackendApi(b.backendAPIObj.Element.ID, params)
	if err != nil {
		return fmt.Errorf("backend [%s] update request: %w", b.backendAPIObj.Element.SystemName, err)
	}

	b.backendAPIObj = updatedBackendAPI

	return nil
}

func (b *BackendAPIEntity) Methods() (*threescaleapi.MethodList, error) {
	// TODO
	return nil, nil
}
