// This file defines kvcache business error codes shared by the public facade
// and internal backend implementations.

package contract

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeKVCacheKeyInvalid reports that a public cache key is not in the expected encoded format.
	CodeKVCacheKeyInvalid = bizerr.MustDefine(
		"KV_CACHE_KEY_INVALID",
		"Cache key format is invalid; use kvcache.BuildCacheKey to construct it",
		gcode.CodeInvalidParameter,
	)
	// CodeKVCacheValueNotInteger reports that a cache entry cannot be read as an integer.
	CodeKVCacheValueNotInteger = bizerr.MustDefine(
		"KV_CACHE_VALUE_NOT_INTEGER",
		"Cache value is not an integer",
		gcode.CodeInvalidParameter,
	)
	// CodeKVCacheIncrementValueNotInteger reports that a cache entry cannot be incremented as an integer.
	CodeKVCacheIncrementValueNotInteger = bizerr.MustDefine(
		"KV_CACHE_INCREMENT_VALUE_NOT_INTEGER",
		"Cache value is not an integer and cannot be incremented",
		gcode.CodeInvalidParameter,
	)
	// CodeKVCacheExpireSecondsNegative reports that cache expiration seconds cannot be negative.
	CodeKVCacheExpireSecondsNegative = bizerr.MustDefine(
		"KV_CACHE_EXPIRE_SECONDS_NEGATIVE",
		"Cache expiration seconds cannot be negative",
		gcode.CodeInvalidParameter,
	)
	// CodeKVCacheExpireSecondsRequired reports that cache writes and expiration updates require a positive TTL.
	CodeKVCacheExpireSecondsRequired = bizerr.MustDefine(
		"KV_CACHE_EXPIRE_SECONDS_REQUIRED",
		"Cache expiration seconds must be greater than 0",
		gcode.CodeInvalidParameter,
	)
	// CodeKVCacheFieldRequired reports that one decoded cache identity field is required.
	CodeKVCacheFieldRequired = bizerr.MustDefine(
		"KV_CACHE_FIELD_REQUIRED",
		"Cache field {field} cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeKVCacheFieldTooLong reports that one decoded cache identity field exceeds its byte limit.
	CodeKVCacheFieldTooLong = bizerr.MustDefine(
		"KV_CACHE_FIELD_TOO_LONG",
		"Cache field {field} exceeds the limit of {maxBytes} bytes",
		gcode.CodeInvalidParameter,
	)
	// CodeKVCacheValueTooLong reports that a cache string value exceeds its byte limit.
	CodeKVCacheValueTooLong = bizerr.MustDefine(
		"KV_CACHE_VALUE_TOO_LONG",
		"Cache value exceeds the limit of {maxBytes} bytes",
		gcode.CodeInvalidParameter,
	)
)
