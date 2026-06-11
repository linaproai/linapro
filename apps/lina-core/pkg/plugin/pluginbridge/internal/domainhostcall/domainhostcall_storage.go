// This file implements the guest-side storage host-service client using the
// injected raw host-service invoker and existing protocol codecs.

package domainhostcall

import (
	"bytes"
	"context"
	"io"
	"time"

	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// storageService adapts the storage host service to storagecap.Service.
type storageService struct{ baseService }

// Storage creates the storage domain guest client.
func Storage(invoker HostServiceInvoker) storagecap.Service {
	return &storageService{baseService: newBaseServiceWithHostService(nil, invoker)}
}

// Put writes one governed storage object under the given logical path.
func (s *storageService) Put(_ context.Context, in storagecap.PutInput) (*storagecap.PutOutput, error) {
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
	payload, err := s.callHostService(
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
func (s *storageService) Get(_ context.Context, in storagecap.GetInput) (*storagecap.GetOutput, error) {
	request := &protocol.HostServiceStorageGetRequest{Path: in.Path}
	payload, err := s.callHostService(
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
func (s *storageService) Delete(_ context.Context, in storagecap.DeleteInput) error {
	request := &protocol.HostServiceStorageDeleteRequest{Path: in.Path}
	_, err := s.callHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStorageDelete,
		in.Path,
		"",
		protocol.MarshalHostServiceStorageDeleteRequest(request),
	)
	return err
}

// List lists governed storage objects under one logical path prefix.
func (s *storageService) List(_ context.Context, in storagecap.ListInput) (*storagecap.ListOutput, error) {
	request := &protocol.HostServiceStorageListRequest{
		Prefix: in.Prefix,
		Limit:  uint32(in.Limit),
	}
	payload, err := s.callHostService(
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
func (s *storageService) Stat(_ context.Context, in storagecap.StatInput) (*storagecap.StatOutput, error) {
	request := &protocol.HostServiceStorageStatRequest{Path: in.Path}
	payload, err := s.callHostService(
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
func (s *storageService) ProviderStatuses(_ context.Context) ([]*storagecap.ProviderStatus, error) {
	return nil, errHostCallsUnavailable
}

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

func storageListEffectiveLimit(limit int) int {
	if limit <= 0 {
		return storagecap.DefaultListLimit
	}
	if limit > storagecap.MaxListLimit {
		return storagecap.MaxListLimit
	}
	return limit
}

func parseWireTime(value string) *time.Time {
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil
	}
	return &parsed
}
