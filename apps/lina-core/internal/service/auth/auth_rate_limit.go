// This file implements lightweight IP/email rate limiting for public auth actions.

package auth

import (
	"context"
	"time"

	"lina-core/internal/service/kvcache"
)

const (
	rateLimitStoreNamespace    = "auth-rate-limit"
	rateLimitWindow            = time.Hour
	registerRateLimitMax       = 10
	forgetPasswordRateLimitMax = 10
)

// rateLimitStore counts attempts in a fixed TTL window.
type rateLimitStore interface {
	// Allow increments the counter for key and reports whether the request is allowed.
	// limit is the maximum number of attempts permitted within the window.
	Allow(ctx context.Context, key string, limit int) (bool, error)
}

// kvRateLimitStore stores counters in the host KV cache.
type kvRateLimitStore struct {
	cache kvcache.Service
}

// newKVRateLimitStore creates a kvcache-backed rate-limit store.
func newKVRateLimitStore(cache kvcache.Service) rateLimitStore {
	return &kvRateLimitStore{cache: cache}
}

// Allow increments the counter for key and returns false when the limit is exceeded.
func (s *kvRateLimitStore) Allow(ctx context.Context, key string, limit int) (bool, error) {
	if limit <= 0 {
		return true, nil
	}
	item, err := s.cache.Incr(ctx, kvcache.OwnerTypeModule, rateLimitCacheKey(key), 1, rateLimitWindow)
	if err != nil {
		return false, err
	}
	return item.IntValue <= int64(limit), nil
}

// rateLimitCacheKey builds the scoped kvcache key for one rate-limit counter.
func rateLimitCacheKey(key string) string {
	return kvcache.BuildCacheKey(authTokenStoreOwner, rateLimitStoreNamespace, key)
}
