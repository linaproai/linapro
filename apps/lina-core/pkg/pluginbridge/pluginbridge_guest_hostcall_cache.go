//go:build wasip1

// This file provides guest-side helpers for the governed distributed cache host service.

package pluginbridge

// CacheHostService exposes guest-side helpers for the governed distributed cache host service.
type CacheHostService interface {
	// Get reads one governed cache value from the authorized namespace.
	Get(namespace string, key string) (*HostServiceCacheValue, bool, error)
	// Set writes one governed cache value into the authorized namespace.
	Set(namespace string, key string, value string, expireSeconds int64) (*HostServiceCacheValue, error)
	// Delete removes one governed cache value from the authorized namespace.
	Delete(namespace string, key string) error
	// Incr increments one governed cache integer value inside the authorized namespace.
	Incr(namespace string, key string, delta int64, expireSeconds int64) (*HostServiceCacheValue, error)
	// Expire updates one governed cache expiration policy inside the authorized namespace.
	Expire(namespace string, key string, expireSeconds int64) (bool, string, error)
}

type cacheHostService struct{}

var defaultCacheHostService CacheHostService = &cacheHostService{}

// Cache returns the distributed cache host service guest client.
func Cache() CacheHostService {
	return defaultCacheHostService
}

// Get reads one governed cache value from the authorized namespace.
func (s *cacheHostService) Get(namespace string, key string) (*HostServiceCacheValue, bool, error) {
	payload, err := invokeHostService(
		HostServiceCache,
		HostServiceMethodCacheGet,
		namespace,
		"",
		MarshalHostServiceCacheGetRequest(&HostServiceCacheGetRequest{Key: key}),
	)
	if err != nil {
		return nil, false, err
	}
	response, err := UnmarshalHostServiceCacheGetResponse(payload)
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
) (*HostServiceCacheValue, error) {
	payload, err := invokeHostService(
		HostServiceCache,
		HostServiceMethodCacheSet,
		namespace,
		"",
		MarshalHostServiceCacheSetRequest(&HostServiceCacheSetRequest{
			Key:           key,
			Value:         value,
			ExpireSeconds: expireSeconds,
		}),
	)
	if err != nil {
		return nil, err
	}
	response, err := UnmarshalHostServiceCacheSetResponse(payload)
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
	_, err := invokeHostService(
		HostServiceCache,
		HostServiceMethodCacheDelete,
		namespace,
		"",
		MarshalHostServiceCacheDeleteRequest(&HostServiceCacheDeleteRequest{Key: key}),
	)
	return err
}

// Incr increments one governed cache integer value inside the authorized namespace.
func (s *cacheHostService) Incr(
	namespace string,
	key string,
	delta int64,
	expireSeconds int64,
) (*HostServiceCacheValue, error) {
	payload, err := invokeHostService(
		HostServiceCache,
		HostServiceMethodCacheIncr,
		namespace,
		"",
		MarshalHostServiceCacheIncrRequest(&HostServiceCacheIncrRequest{
			Key:           key,
			Delta:         delta,
			ExpireSeconds: expireSeconds,
		}),
	)
	if err != nil {
		return nil, err
	}
	response, err := UnmarshalHostServiceCacheIncrResponse(payload)
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
	payload, err := invokeHostService(
		HostServiceCache,
		HostServiceMethodCacheExpire,
		namespace,
		"",
		MarshalHostServiceCacheExpireRequest(&HostServiceCacheExpireRequest{
			Key:           key,
			ExpireSeconds: expireSeconds,
		}),
	)
	if err != nil {
		return false, "", err
	}
	response, err := UnmarshalHostServiceCacheExpireResponse(payload)
	if err != nil {
		return false, "", err
	}
	if response == nil {
		return false, "", nil
	}
	return response.Found, response.ExpireAt, nil
}
