//go:build wasip1

// This file adapts the governed storage host service transport to storagecap.Service.

package pluginbridge

import (
	"bytes"
	"context"
	"io"

	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// storageHostService is the default guest-side storage host-service client.
type storageHostService struct{}

// defaultStorageHostService stores the singleton storage host-service client
// used by package-level helpers.
var defaultStorageHostService storagecap.Service = &storageHostService{}

// Storage returns the storage domain guest client.
func Storage() storagecap.Service {
	return defaultStorageHostService
}

// Put writes one governed storage object under the given logical path.
func (s *storageHostService) Put(_ context.Context, in storagecap.PutInput) (*storagecap.PutOutput, error) {
	body, err := io.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}
	request := &protocol.HostServiceStoragePutRequest{
		Path:        in.Path,
		Body:        body,
		ContentType: in.ContentType,
		Overwrite:   in.Overwrite,
	}
	payload, err := invokeGuestHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStoragePut,
		in.Path,
		"",
		protocol.MarshalHostServiceStoragePutRequest(request),
	)
	if err != nil {
		return nil, err
	}
	response, err := protocol.UnmarshalHostServiceStoragePutResponse(payload)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil
	}
	return &storagecap.PutOutput{Object: storageObjectFromWire(response.Object)}, nil
}

// Get reads one governed storage object under the given logical path.
func (s *storageHostService) Get(_ context.Context, in storagecap.GetInput) (*storagecap.GetOutput, error) {
	request := &protocol.HostServiceStorageGetRequest{Path: in.Path}
	payload, err := invokeGuestHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStorageGet,
		in.Path,
		"",
		protocol.MarshalHostServiceStorageGetRequest(request),
	)
	if err != nil {
		return nil, err
	}
	response, err := protocol.UnmarshalHostServiceStorageGetResponse(payload)
	if err != nil {
		return nil, err
	}
	if response == nil || !response.Found {
		return &storagecap.GetOutput{Found: false}, nil
	}
	return &storagecap.GetOutput{
		Object: storageObjectFromWire(response.Object),
		Body:   io.NopCloser(bytes.NewReader(response.Body)),
		Found:  true,
	}, nil
}

// Delete removes one governed storage object under the given logical path.
func (s *storageHostService) Delete(_ context.Context, in storagecap.DeleteInput) error {
	request := &protocol.HostServiceStorageDeleteRequest{Path: in.Path}
	_, err := invokeGuestHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStorageDelete,
		in.Path,
		"",
		protocol.MarshalHostServiceStorageDeleteRequest(request),
	)
	return err
}

// List lists governed storage objects under one logical path prefix.
func (s *storageHostService) List(_ context.Context, in storagecap.ListInput) (*storagecap.ListOutput, error) {
	request := &protocol.HostServiceStorageListRequest{
		Prefix: in.Prefix,
		Limit:  uint32(in.Limit),
	}
	payload, err := invokeGuestHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStorageList,
		in.Prefix,
		"",
		protocol.MarshalHostServiceStorageListRequest(request),
	)
	if err != nil {
		return nil, err
	}
	response, err := protocol.UnmarshalHostServiceStorageListResponse(payload)
	if err != nil {
		return nil, err
	}
	output := &storagecap.ListOutput{
		Objects: []*storagecap.Object{},
		Limit:   storageListEffectiveLimit(in.Limit),
	}
	if response == nil {
		return output, nil
	}
	for _, object := range response.Objects {
		output.Objects = append(output.Objects, storageObjectFromWire(object))
	}
	return output, nil
}

// Stat reads metadata for one governed storage object under the given logical path.
func (s *storageHostService) Stat(_ context.Context, in storagecap.StatInput) (*storagecap.StatOutput, error) {
	request := &protocol.HostServiceStorageStatRequest{Path: in.Path}
	payload, err := invokeGuestHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStorageStat,
		in.Path,
		"",
		protocol.MarshalHostServiceStorageStatRequest(request),
	)
	if err != nil {
		return nil, err
	}
	response, err := protocol.UnmarshalHostServiceStorageStatResponse(payload)
	if err != nil {
		return nil, err
	}
	if response == nil || !response.Found {
		return &storagecap.StatOutput{Found: false}, nil
	}
	return &storagecap.StatOutput{Object: storageObjectFromWire(response.Object), Found: true}, nil
}

// ProviderStatuses is not exposed through the dynamic storage host-service
// transport. Source plugins can call the host-side storagecap.Service directly.
func (s *storageHostService) ProviderStatuses(_ context.Context) ([]*storagecap.ProviderStatus, error) {
	return nil, ErrHostCallsUnavailable
}

// storageObjectFromWire maps one transport storage snapshot to a domain object.
func storageObjectFromWire(object *protocol.HostServiceStorageObject) *storagecap.Object {
	if object == nil {
		return nil
	}
	return &storagecap.Object{
		Path:        object.Path,
		Size:        object.Size,
		ContentType: object.ContentType,
		UpdatedAt:   parseWireTime(object.UpdatedAt),
		Visibility:  object.Visibility,
	}
}
