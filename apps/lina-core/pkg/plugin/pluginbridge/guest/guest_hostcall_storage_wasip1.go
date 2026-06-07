//go:build wasip1

// This file provides guest-side helpers for the governed storage host service.

package guest

import "lina-core/pkg/plugin/pluginbridge/protocol"

// storageHostService is the default guest-side storage host-service client.
type storageHostService struct{}

// defaultStorageHostService stores the singleton storage host-service client
// used by package-level helpers.
var defaultStorageHostService StorageHostService = &storageHostService{}

// Storage returns the storage host service guest client.
func Storage() StorageHostService {
	return defaultStorageHostService
}

// Put writes one governed storage object under the given logical path.
func (s *storageHostService) Put(
	objectPath string,
	body []byte,
	contentType string,
	overwrite bool,
) (*protocol.HostServiceStorageObject, error) {
	request := &protocol.HostServiceStoragePutRequest{
		Path:        objectPath,
		Body:        body,
		ContentType: contentType,
		Overwrite:   overwrite,
	}
	payload, err := invokeGuestHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStoragePut,
		objectPath,
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
	return response.Object, nil
}

// PutText writes one UTF-8 text object under the given logical path.
func (s *storageHostService) PutText(
	objectPath string,
	content string,
	contentType string,
	overwrite bool,
) (*protocol.HostServiceStorageObject, error) {
	return s.Put(objectPath, []byte(content), contentType, overwrite)
}

// Get reads one governed storage object under the given logical path.
func (s *storageHostService) Get(
	objectPath string,
) ([]byte, *protocol.HostServiceStorageObject, bool, error) {
	request := &protocol.HostServiceStorageGetRequest{Path: objectPath}
	payload, err := invokeGuestHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStorageGet,
		objectPath,
		"",
		protocol.MarshalHostServiceStorageGetRequest(request),
	)
	if err != nil {
		return nil, nil, false, err
	}
	response, err := protocol.UnmarshalHostServiceStorageGetResponse(payload)
	if err != nil {
		return nil, nil, false, err
	}
	if response == nil || !response.Found {
		return nil, nil, false, nil
	}
	return response.Body, response.Object, true, nil
}

// GetText reads one UTF-8 text object under the given logical path.
func (s *storageHostService) GetText(
	objectPath string,
) (string, *protocol.HostServiceStorageObject, bool, error) {
	body, object, found, err := s.Get(objectPath)
	if err != nil || !found {
		return "", object, found, err
	}
	return string(body), object, true, nil
}

// Delete removes one governed storage object under the given logical path.
func (s *storageHostService) Delete(objectPath string) error {
	request := &protocol.HostServiceStorageDeleteRequest{Path: objectPath}
	_, err := invokeGuestHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStorageDelete,
		objectPath,
		"",
		protocol.MarshalHostServiceStorageDeleteRequest(request),
	)
	return err
}

// List lists governed storage objects under one logical path prefix.
func (s *storageHostService) List(
	prefix string,
	limit uint32,
) ([]*protocol.HostServiceStorageObject, error) {
	request := &protocol.HostServiceStorageListRequest{
		Prefix: prefix,
		Limit:  limit,
	}
	payload, err := invokeGuestHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStorageList,
		prefix,
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
	if response == nil {
		return []*protocol.HostServiceStorageObject{}, nil
	}
	return response.Objects, nil
}

// Stat reads metadata for one governed storage object under the given logical path.
func (s *storageHostService) Stat(
	objectPath string,
) (*protocol.HostServiceStorageObject, bool, error) {
	request := &protocol.HostServiceStorageStatRequest{Path: objectPath}
	payload, err := invokeGuestHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStorageStat,
		objectPath,
		"",
		protocol.MarshalHostServiceStorageStatRequest(request),
	)
	if err != nil {
		return nil, false, err
	}
	response, err := protocol.UnmarshalHostServiceStorageStatResponse(payload)
	if err != nil {
		return nil, false, err
	}
	if response == nil || !response.Found {
		return nil, false, nil
	}
	return response.Object, true, nil
}
