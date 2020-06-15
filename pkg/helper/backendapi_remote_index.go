package helper

import (
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"

	"github.com/go-logr/logr"
)

type BackendAPIRemoteIndex struct {
	client                 *threescaleapi.ThreeScaleClient
	logger                 logr.Logger
	backendIDIndex         map[int64]*BackendAPIEntity
	backendSystemNameIndex map[string]*BackendAPIEntity
}

func NewBackendAPIRemoteIndex(client *threescaleapi.ThreeScaleClient, logger logr.Logger) (*BackendAPIRemoteIndex, error) {
	backendAPIs, err := client.ListBackendApis()
	if err != nil {
		return nil, err
	}

	backendIDIndex := map[int64]*BackendAPIEntity{}
	backendSystemNameIndex := map[string]*BackendAPIEntity{}

	for _, backendObj := range backendAPIs.Backends {
		backendAPIEntity := NewBackendAPIEntity(&backendObj, client, logger)
		backendIDIndex[backendAPIEntity.ID()] = backendAPIEntity
		backendSystemNameIndex[backendAPIEntity.SystemName()] = backendAPIEntity
	}

	reader := &BackendAPIRemoteIndex{
		client:                 client,
		logger:                 logger,
		backendIDIndex:         backendIDIndex,
		backendSystemNameIndex: backendSystemNameIndex,
	}

	return reader, nil
}

// FindByID finds remote backendAPI item by ID
func (b *BackendAPIRemoteIndex) FindByID(id int64) (*BackendAPIEntity, bool) {
	item, ok := b.backendIDIndex[id]
	return item, ok
}

// FindBySystemName finds remote backendAPI item by SystenName
func (b *BackendAPIRemoteIndex) FindBySystemName(systemName string) (*BackendAPIEntity, bool) {
	item, ok := b.backendSystemNameIndex[systemName]
	return item, ok
}

// CreateBackendAPI create remote backendAPI
func (b *BackendAPIRemoteIndex) CreateBackendAPI(params threescaleapi.Params) (*BackendAPIEntity, error) {
	backendObj, err := b.client.CreateBackendApi(params)
	if err != nil {
		return nil, err
	}

	backendAPIEntity := NewBackendAPIEntity(backendObj, b.client, b.logger)
	b.backendIDIndex[backendAPIEntity.ID()] = backendAPIEntity
	b.backendSystemNameIndex[backendAPIEntity.SystemName()] = backendAPIEntity

	return backendAPIEntity, nil
}
