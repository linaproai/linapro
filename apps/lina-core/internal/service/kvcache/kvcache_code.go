// This file re-exports kvcache business error codes from the internal contract
// package so existing callers can continue using the public service package.

package kvcache

import (
	"lina-core/internal/service/kvcache/internal/contract"
)

var (
	// CodeKVCacheKeyInvalid reports that a public cache key is not in the expected encoded format.
	CodeKVCacheKeyInvalid = contract.CodeKVCacheKeyInvalid
	// CodeKVCacheValueNotInteger reports that a cache entry cannot be read as an integer.
	CodeKVCacheValueNotInteger = contract.CodeKVCacheValueNotInteger
	// CodeKVCacheIncrementValueNotInteger reports that a cache entry cannot be incremented as an integer.
	CodeKVCacheIncrementValueNotInteger = contract.CodeKVCacheIncrementValueNotInteger
	// CodeKVCacheExpireSecondsNegative reports that cache expiration seconds cannot be negative.
	CodeKVCacheExpireSecondsNegative = contract.CodeKVCacheExpireSecondsNegative
	// CodeKVCacheExpireSecondsRequired reports that cache writes and expiration updates require a positive TTL.
	CodeKVCacheExpireSecondsRequired = contract.CodeKVCacheExpireSecondsRequired
	// CodeKVCacheFieldRequired reports that one decoded cache identity field is required.
	CodeKVCacheFieldRequired = contract.CodeKVCacheFieldRequired
	// CodeKVCacheFieldTooLong reports that one decoded cache identity field exceeds its byte limit.
	CodeKVCacheFieldTooLong = contract.CodeKVCacheFieldTooLong
	// CodeKVCacheValueTooLong reports that a cache string value exceeds its byte limit.
	CodeKVCacheValueTooLong = contract.CodeKVCacheValueTooLong
)
