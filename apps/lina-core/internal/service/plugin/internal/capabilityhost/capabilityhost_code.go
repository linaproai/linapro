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
	// CodePluginHostExternalLoginPluginRequired reports that the external-login
	// facade was used before the host bound it to one source-plugin identity.
	// It is a fail-closed guard: an unbound facade never issues a session.
	CodePluginHostExternalLoginPluginRequired = bizerr.MustDefine(
		"PLUGIN_HOST_EXTERNAL_LOGIN_PLUGIN_ID_REQUIRED",
		"External login service requires a plugin ID",
		gcode.CodeInvalidParameter,
	)
	// CodePluginHostExternalLoginProviderForbidden reports that the calling
	// plugin requested an external-login provider it did not declare through
	// ProvideExternalIdentity. This prevents one plugin from minting sessions
	// through another plugin's provider.
	CodePluginHostExternalLoginProviderForbidden = bizerr.MustDefine(
		"PLUGIN_HOST_EXTERNAL_LOGIN_PROVIDER_FORBIDDEN",
		"External login provider is not owned by the calling plugin",
		gcode.CodeNotAuthorized,
	)
	// CodePluginHostExternalLoginPluginDisabled reports that the calling plugin
	// is not currently enabled and must not issue external-login sessions.
	CodePluginHostExternalLoginPluginDisabled = bizerr.MustDefine(
		"PLUGIN_HOST_EXTERNAL_LOGIN_PLUGIN_DISABLED",
		"External login is unavailable because the plugin is disabled",
		gcode.CodeNotAuthorized,
	)
)
