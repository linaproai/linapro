// This file defines caller-visible plugin host service error codes used by
// source-plugin adapters when runtime-scoped host dependencies are unavailable.

package pluginhostservices

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
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
)
