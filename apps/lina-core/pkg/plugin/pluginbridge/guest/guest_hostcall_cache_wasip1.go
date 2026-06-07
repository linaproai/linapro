//go:build wasip1

// This file provides guest-side helpers for the governed distributed cache host service.

package guest

import "lina-core/pkg/plugin/pluginbridge/protocol"

// cacheHostService is the default guest-side distributed cache host-service
// client.
type cacheHostService struct{}

// defaultCacheHostService stores the singleton cache host-service client used
// by package-level helpers.
var defaultCacheHostService CacheHostService = &cacheHostService{}

// Cache returns the distributed cache host service guest client.
func Cache() CacheHostService {
	return defaultCacheHostService
}

// Get reads one governed cache value from the authorized namespace.
func (s *cacheHostService) Get(namespace string, key string) (*protocol.HostServiceCacheValue, bool, error) {
	payload, err := invokeGuestHostService(
		protocol.HostServiceCache,
		protocol.HostServiceMethodCacheGet,
		namespace,
		"",
		protocol.MarshalHostServiceCacheGetRequest(&protocol.HostServiceCacheGetRequest{Key: key}),
	)
	if err != nil {
		return nil, false, err
	}
	response, err := protocol.UnmarshalHostServiceCacheGetResponse(payload)
	if err != nil {
		return nil, false, err
	}
	if response == nil || !response.Found {
		return nil, false, nil
	}
	return response.Value, true, nil
}

// Set writes one governed cache value into the authorized namespace.
func (s *cacheHostService) Set(
	namespace string,
	key string,
	value string,
	expireSeconds int64,
) (*protocol.HostServiceCacheValue, error) {
	payload, err := invokeGuestHostService(
		protocol.HostServiceCache,
		protocol.HostServiceMethodCacheSet,
		namespace,
		"",
		protocol.MarshalHostServiceCacheSetRequest(&protocol.HostServiceCacheSetRequest{
			Key:           key,
			Value:         value,
			ExpireSeconds: expireSeconds,
		}),
	)
	if err != nil {
		return nil, err
	}
	response, err := protocol.UnmarshalHostServiceCacheSetResponse(payload)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil
	}
	return response.Value, nil
}

// Delete removes one governed cache value from the authorized namespace.
func (s *cacheHostService) Delete(namespace string, key string) error {
	_, err := invokeGuestHostService(
		protocol.HostServiceCache,
		protocol.HostServiceMethodCacheDelete,
		namespace,
		"",
		protocol.MarshalHostServiceCacheDeleteRequest(&protocol.HostServiceCacheDeleteRequest{Key: key}),
	)
	return err
}

// Incr increments one governed cache integer value inside the authorized namespace.
func (s *cacheHostService) Incr(
	namespace string,
	key string,
	delta int64,
	expireSeconds int64,
) (*protocol.HostServiceCacheValue, error) {
	payload, err := invokeGuestHostService(
		protocol.HostServiceCache,
		protocol.HostServiceMethodCacheIncr,
		namespace,
		"",
		protocol.MarshalHostServiceCacheIncrRequest(&protocol.HostServiceCacheIncrRequest{
			Key:           key,
			Delta:         delta,
			ExpireSeconds: expireSeconds,
		}),
	)
	if err != nil {
		return nil, err
	}
	response, err := protocol.UnmarshalHostServiceCacheIncrResponse(payload)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil
	}
	return response.Value, nil
}

// Expire updates one governed cache expiration policy inside the authorized namespace.
func (s *cacheHostService) Expire(namespace string, key string, expireSeconds int64) (bool, string, error) {
	payload, err := invokeGuestHostService(
		protocol.HostServiceCache,
		protocol.HostServiceMethodCacheExpire,
		namespace,
		"",
		protocol.MarshalHostServiceCacheExpireRequest(&protocol.HostServiceCacheExpireRequest{
			Key:           key,
			ExpireSeconds: expireSeconds,
		}),
	)
	if err != nil {
		return false, "", err
	}
	response, err := protocol.UnmarshalHostServiceCacheExpireResponse(payload)
	if err != nil {
		return false, "", err
	}
	if response == nil {
		return false, "", nil
	}
	return response.Found, response.ExpireAt, nil
}
