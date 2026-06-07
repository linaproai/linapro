// This file adapts the host KV cache service into the source-plugin visible
// cache contract while binding every operation to one plugin and tenant scope.

package hostservices

import (
	"context"
	"strings"
	"time"

	"lina-core/internal/service/kvcache"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/tenantcap"
)

const (
	// sourcePluginCacheTenantScope isolates tenant-aware source-plugin cache
	// keys from other tenant-scoped runtime cache content.
	sourcePluginCacheTenantScope = "plugin-cache"
)

// cacheAdapter binds the shared host cache service to one source plugin.
type cacheAdapter struct {
	service  kvcache.Service
	bizCtx   bizctxcap.Service
	pluginID string
}

// newCacheAdapter creates one source-plugin cache adapter bound to pluginID.
func newCacheAdapter(
	service kvcache.Service,
	bizCtx bizctxcap.Service,
	pluginID string,
) cachecap.Service {
	return &cacheAdapter{
		service:  service,
		bizCtx:   bizCtx,
		pluginID: strings.TrimSpace(pluginID),
	}
}

// Get returns one unexpired plugin cache item.
func (s *cacheAdapter) Get(ctx context.Context, namespace string, key string) (*cachecap.CacheItem, bool, error) {
	cacheKey, err := s.buildCacheKey(ctx, namespace, key)
	if err != nil {
		return nil, false, err
	}
	item, found, err := s.service.Get(ctx, kvcache.OwnerTypePlugin, cacheKey)
	if err != nil || !found {
		return nil, found, err
	}
	return fromKVCacheItem(item, key), true, nil
}

// Set stores a string plugin cache item.
func (s *cacheAdapter) Set(
	ctx context.Context,
	namespace string,
	key string,
	value string,
	ttl time.Duration,
) (*cachecap.CacheItem, error) {
	cacheKey, err := s.buildCacheKey(ctx, namespace, key)
	if err != nil {
		return nil, err
	}
	item, err := s.service.Set(ctx, kvcache.OwnerTypePlugin, cacheKey, value, ttl)
	return fromKVCacheItem(item, key), err
}

// Delete removes one plugin cache item.
func (s *cacheAdapter) Delete(ctx context.Context, namespace string, key string) error {
	cacheKey, err := s.buildCacheKey(ctx, namespace, key)
	if err != nil {
		return err
	}
	return s.service.Delete(ctx, kvcache.OwnerTypePlugin, cacheKey)
}

// Incr increments one integer plugin cache item.
func (s *cacheAdapter) Incr(
	ctx context.Context,
	namespace string,
	key string,
	delta int64,
	ttl time.Duration,
) (*cachecap.CacheItem, error) {
	cacheKey, err := s.buildCacheKey(ctx, namespace, key)
	if err != nil {
		return nil, err
	}
	item, err := s.service.Incr(ctx, kvcache.OwnerTypePlugin, cacheKey, delta, ttl)
	return fromKVCacheItem(item, key), err
}

// Expire updates one plugin cache item's expiration policy.
func (s *cacheAdapter) Expire(
	ctx context.Context,
	namespace string,
	key string,
	ttl time.Duration,
) (bool, *time.Time, error) {
	cacheKey, err := s.buildCacheKey(ctx, namespace, key)
	if err != nil {
		return false, nil, err
	}
	return s.service.Expire(ctx, kvcache.OwnerTypePlugin, cacheKey, ttl)
}

// buildCacheKey maps one plugin-visible cache identity to the host kvcache key.
func (s *cacheAdapter) buildCacheKey(ctx context.Context, namespace string, key string) (string, error) {
	if s == nil || s.service == nil {
		return "", bizerr.NewCode(CodePluginSourceCacheServiceUnavailable)
	}
	if s.pluginID == "" {
		return "", bizerr.NewCode(CodePluginSourceCachePluginIDRequired)
	}
	tenantID := s.currentTenantID(ctx)
	if tenantID > 0 {
		return kvcache.BuildTenantCacheKey(
			tenantcap.TenantID(tenantID),
			sourcePluginCacheTenantScope,
			s.pluginID,
			namespace,
			key,
		), nil
	}
	return kvcache.BuildCacheKey(s.pluginID, namespace, key), nil
}

// currentTenantID returns the plugin-visible tenant scope for the current call.
func (s *cacheAdapter) currentTenantID(ctx context.Context) int {
	if s != nil && s.bizCtx != nil {
		return s.bizCtx.Current(ctx).TenantID
	}
	return bizctxcap.CurrentFromContext(ctx).TenantID
}

// fromKVCacheItem maps one internal cache item into the source-plugin contract
// without exposing host-internal encoded cache keys.
func fromKVCacheItem(item *kvcache.Item, logicalKey string) *cachecap.CacheItem {
	if item == nil {
		return nil
	}
	return &cachecap.CacheItem{
		Key:       strings.TrimSpace(logicalKey),
		ValueKind: item.ValueKind,
		Value:     item.Value,
		IntValue:  item.IntValue,
		ExpireAt:  item.ExpireAt,
	}
}
