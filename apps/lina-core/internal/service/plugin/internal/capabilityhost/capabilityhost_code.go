// This file defines caller-visible plugin host service error codes used by
// source-plugin adapters when runtime-scoped host dependencies are unavailable.

package capabilityhost

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodePluginHostAuthTokenStateUnavailable reports that shared host auth
	// token state cannot be read or written through plugin host services.
	CodePluginHostAuthTokenStateUnavailable = bizerr.MustDefine(
		"PLUGIN_HOST_AUTH_TOKEN_STATE_UNAVAILABLE",
		"Plugin host authentication token state is temporarily unavailable",
		gcode.CodeInternalError,
	)
	// CodePluginHostAuthTokenInvalid reports an invalid or missing host auth token
	// passed through plugin host services.
	CodePluginHostAuthTokenInvalid = bizerr.MustDefine(
		"PLUGIN_HOST_AUTH_TOKEN_INVALID",
		"Plugin host authentication token is invalid",
		gcode.CodeNotAuthorized,
	)
	// CodePluginSourceCacheServiceUnavailable reports that the cache facade has
	// no shared host cache service and cannot safely serve source plugins.
	CodePluginSourceCacheServiceUnavailable = bizerr.MustDefine(
		"PLUGIN_SOURCE_CACHE_SERVICE_UNAVAILABLE",
		"Source plugin cache service is not configured",
		gcode.CodeInvalidOperation,
	)
	// CodePluginSourceCachePluginIDRequired reports that the cache facade was
	// used before the host bound it to one source-plugin identity.
	CodePluginSourceCachePluginIDRequired = bizerr.MustDefine(
		"PLUGIN_SOURCE_CACHE_PLUGIN_ID_REQUIRED",
		"Source plugin cache service requires a plugin ID",
		gcode.CodeInvalidParameter,
	)
	// CodePluginSourceCacheKeyRequired reports an empty plugin cache key.
	CodePluginSourceCacheKeyRequired = bizerr.MustDefine(
		"PLUGIN_SOURCE_CACHE_KEY_REQUIRED",
		"Source plugin cache key cannot be empty",
		gcode.CodeInvalidParameter,
	)
)
