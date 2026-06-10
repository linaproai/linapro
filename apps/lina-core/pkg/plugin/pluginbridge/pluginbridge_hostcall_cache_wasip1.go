//go:build wasip1

// This file adapts the governed cache host service transport to cachecap.Service.

package pluginbridge

import (
	"context"
	"time"

	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// cacheHostService is the default guest-side distributed cache host-service
// client.
type cacheHostService struct{}

// defaultCacheHostService stores the singleton cache host-service client used
// by package-level helpers.
var defaultCacheHostService cachecap.Service = &cacheHostService{}

// Cache returns the distributed cache domain guest client.
func Cache() cachecap.Service {
	return defaultCacheHostService
}

// Get reads one governed cache value from the authorized namespace.
func (s *cacheHostService) Get(_ context.Context, namespace string, key string) (*cachecap.CacheItem, bool, error) {
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
	return cacheItemFromWire(key, response.Value), true, nil
}

// Set writes one governed cache value into the authorized namespace.
func (s *cacheHostService) Set(
	_ context.Context,
	namespace string,
	key string,
	value string,
	ttl time.Duration,
) (*cachecap.CacheItem, error) {
	payload, err := invokeGuestHostService(
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
func (s *cacheHostService) Delete(_ context.Context, namespace string, key string) error {
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
	_ context.Context,
	namespace string,
	key string,
	delta int64,
	ttl time.Duration,
) (*cachecap.CacheItem, error) {
	payload, err := invokeGuestHostService(
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
func (s *cacheHostService) Expire(_ context.Context, namespace string, key string, ttl time.Duration) (bool, *time.Time, error) {
	payload, err := invokeGuestHostService(
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

// cacheItemFromWire maps one transport cache value into a domain cache item.
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

// durationSeconds converts a domain duration to the wire seconds field.
func durationSeconds(duration time.Duration) int64 {
	if duration <= 0 {
		return 0
	}
	return int64((duration + time.Second - 1) / time.Second)
}
