// This file centralizes backend-neutral TTL validation and cache item time
// helpers shared by kvcache backend implementations.

package contract

import (
	"time"

	"lina-core/pkg/bizerr"
)

// ValidatePositiveTTL enforces the kvcache contract that every write or
// expiration update has an explicit finite lifetime.
func ValidatePositiveTTL(ttl time.Duration) error {
	if ttl < 0 {
		return bizerr.NewCode(CodeKVCacheExpireSecondsNegative)
	}
	if ttl == 0 {
		return bizerr.NewCode(CodeKVCacheExpireSecondsRequired)
	}
	return nil
}

// ExpireAtFromTTL converts a relative TTL into the public cache item shape.
func ExpireAtFromTTL(ttl time.Duration) *time.Time {
	if ttl <= 0 {
		return nil
	}
	expireAt := time.Now().Add(ttl)
	return &expireAt
}

// CloneTime returns a copy so callers cannot mutate cached metadata through the
// public item.
func CloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
