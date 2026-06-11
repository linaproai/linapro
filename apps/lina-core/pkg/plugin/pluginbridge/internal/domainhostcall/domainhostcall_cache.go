// This file implements the guest-side cache host-service client.

package domainhostcall

import (
	"context"
	"time"

	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// cacheService adapts the cache host service to cachecap.Service.
type cacheService struct{ baseService }

// Cache creates the distributed cache domain guest client.
func Cache(invoker HostServiceInvoker) cachecap.Service {
	return &cacheService{baseService: newBaseServiceWithHostService(nil, invoker)}
}

// Get reads one governed cache value from the authorized namespace.
func (s *cacheService) Get(_ context.Context, namespace string, key string) (*cachecap.CacheItem, bool, error) {
	payload, err := s.callHostService(
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
	return cacheItemFromWire(key, response.Value), true, nil
}

// Set writes one governed cache value into the authorized namespace.
func (s *cacheService) Set(
	_ context.Context,
	namespace string,
	key string,
	value string,
	ttl time.Duration,
) (*cachecap.CacheItem, error) {
	payload, err := s.callHostService(
		protocol.HostServiceCache,
		protocol.HostServiceMethodCacheSet,
		namespace,
		"",
		protocol.MarshalHostServiceCacheSetRequest(&protocol.HostServiceCacheSetRequest{
			Key:           key,
			Value:         value,
			ExpireSeconds: durationSeconds(ttl),
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
	return cacheItemFromWire(key, response.Value), nil
}

// Delete removes one governed cache value from the authorized namespace.
func (s *cacheService) Delete(_ context.Context, namespace string, key string) error {
	_, err := s.callHostService(
		protocol.HostServiceCache,
		protocol.HostServiceMethodCacheDelete,
		namespace,
		"",
		protocol.MarshalHostServiceCacheDeleteRequest(&protocol.HostServiceCacheDeleteRequest{Key: key}),
	)
	return err
}

// Incr increments one governed cache integer value inside the authorized namespace.
func (s *cacheService) Incr(
	_ context.Context,
	namespace string,
	key string,
	delta int64,
	ttl time.Duration,
) (*cachecap.CacheItem, error) {
	payload, err := s.callHostService(
		protocol.HostServiceCache,
		protocol.HostServiceMethodCacheIncr,
		namespace,
		"",
		protocol.MarshalHostServiceCacheIncrRequest(&protocol.HostServiceCacheIncrRequest{
			Key:           key,
			Delta:         delta,
			ExpireSeconds: durationSeconds(ttl),
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
	return cacheItemFromWire(key, response.Value), nil
}

// Expire updates one governed cache expiration policy inside the authorized namespace.
func (s *cacheService) Expire(_ context.Context, namespace string, key string, ttl time.Duration) (bool, *time.Time, error) {
	payload, err := s.callHostService(
		protocol.HostServiceCache,
		protocol.HostServiceMethodCacheExpire,
		namespace,
		"",
		protocol.MarshalHostServiceCacheExpireRequest(&protocol.HostServiceCacheExpireRequest{
			Key:           key,
			ExpireSeconds: durationSeconds(ttl),
		}),
	)
	if err != nil {
		return false, nil, err
	}
	response, err := protocol.UnmarshalHostServiceCacheExpireResponse(payload)
	if err != nil {
		return false, nil, err
	}
	if response == nil {
		return false, nil, nil
	}
	return response.Found, parseWireTime(response.ExpireAt), nil
}

func cacheItemFromWire(key string, value *protocol.HostServiceCacheValue) *cachecap.CacheItem {
	if value == nil {
		return nil
	}
	return &cachecap.CacheItem{
		Key:       key,
		ValueKind: int(value.ValueKind),
		Value:     value.Value,
		IntValue:  value.IntValue,
		ExpireAt:  parseWireTime(value.ExpireAt),
	}
}

func durationSeconds(duration time.Duration) int64 {
	if duration <= 0 {
		return 0
	}
	return int64((duration + time.Second - 1) / time.Second)
}
