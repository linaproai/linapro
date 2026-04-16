//go:build wasip1

// This file provides guest-side helpers for the governed storage host service.

package pluginbridge

// StorageHostService exposes guest-side helpers for the governed storage host service.
type StorageHostService interface {
	// Put writes one governed storage object under the given logical path.
	Put(objectPath string, body []byte, contentType string, overwrite bool) (*HostServiceStorageObject, error)
	// PutText writes one UTF-8 text object under the given logical path.
	PutText(objectPath string, content string, contentType string, overwrite bool) (*HostServiceStorageObject, error)
	// Get reads one governed storage object under the given logical path.
	Get(objectPath string) ([]byte, *HostServiceStorageObject, bool, error)
	// GetText reads one UTF-8 text object under the given logical path.
	GetText(objectPath string) (string, *HostServiceStorageObject, bool, error)
	// Delete removes one governed storage object under the given logical path.
	Delete(objectPath string) error
	// List lists governed storage objects under one logical path prefix.
	List(prefix string, limit uint32) ([]*HostServiceStorageObject, error)
	// Stat reads metadata for one governed storage object under the given logical path.
	Stat(objectPath string) (*HostServiceStorageObject, bool, error)
}

type storageHostService struct{}

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
) (*HostServiceStorageObject, error) {
	request := &HostServiceStoragePutRequest{
		Path:        objectPath,
		Body:        body,
		ContentType: contentType,
		Overwrite:   overwrite,
	}
	payload, err := invokeHostService(
		HostServiceStorage,
		HostServiceMethodStoragePut,
		objectPath,
		"",
		MarshalHostServiceStoragePutRequest(request),
	)
	if err != nil {
		return nil, err
	}
	response, err := UnmarshalHostServiceStoragePutResponse(payload)
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
) (*HostServiceStorageObject, error) {
	return s.Put(objectPath, []byte(content), contentType, overwrite)
}

// Get reads one governed storage object under the given logical path.
func (s *storageHostService) Get(
	objectPath string,
) ([]byte, *HostServiceStorageObject, bool, error) {
	request := &HostServiceStorageGetRequest{Path: objectPath}
	payload, err := invokeHostService(
		HostServiceStorage,
		HostServiceMethodStorageGet,
		objectPath,
		"",
		MarshalHostServiceStorageGetRequest(request),
	)
	if err != nil {
		return nil, nil, false, err
	}
	response, err := UnmarshalHostServiceStorageGetResponse(payload)
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
) (string, *HostServiceStorageObject, bool, error) {
	body, object, found, err := s.Get(objectPath)
	if err != nil || !found {
		return "", object, found, err
	}
	return string(body), object, true, nil
}

// Delete removes one governed storage object under the given logical path.
func (s *storageHostService) Delete(objectPath string) error {
	request := &HostServiceStorageDeleteRequest{Path: objectPath}
	_, err := invokeHostService(
		HostServiceStorage,
		HostServiceMethodStorageDelete,
		objectPath,
		"",
		MarshalHostServiceStorageDeleteRequest(request),
	)
	return err
}

// List lists governed storage objects under one logical path prefix.
func (s *storageHostService) List(
	prefix string,
	limit uint32,
) ([]*HostServiceStorageObject, error) {
	request := &HostServiceStorageListRequest{
		Prefix: prefix,
		Limit:  limit,
	}
	payload, err := invokeHostService(
		HostServiceStorage,
		HostServiceMethodStorageList,
		prefix,
		"",
		MarshalHostServiceStorageListRequest(request),
	)
	if err != nil {
		return nil, err
	}
	response, err := UnmarshalHostServiceStorageListResponse(payload)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return []*HostServiceStorageObject{}, nil
	}
	return response.Objects, nil
}

// Stat reads metadata for one governed storage object under the given logical path.
func (s *storageHostService) Stat(
	objectPath string,
) (*HostServiceStorageObject, bool, error) {
	request := &HostServiceStorageStatRequest{Path: objectPath}
	payload, err := invokeHostService(
		HostServiceStorage,
		HostServiceMethodStorageStat,
		objectPath,
		"",
		MarshalHostServiceStorageStatRequest(request),
	)
	if err != nil {
		return nil, false, err
	}
	response, err := UnmarshalHostServiceStorageStatResponse(payload)
	if err != nil {
		return nil, false, err
	}
	if response == nil || !response.Found {
		return nil, false, nil
	}
	return response.Object, true, nil
}
