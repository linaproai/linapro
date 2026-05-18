// This file defines cache coordination business error codes.

package cachecoord

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeCacheCoordDomainInvalid reports an invalid cache domain.
	CodeCacheCoordDomainInvalid = bizerr.MustDefine(
		"CACHE_COORD_DOMAIN_INVALID",
		"Cache coordination domain is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeCacheCoordScopeInvalid reports an invalid cache invalidation scope.
	CodeCacheCoordScopeInvalid = bizerr.MustDefine(
		"CACHE_COORD_SCOPE_INVALID",
		"Cache coordination scope is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeCacheCoordPublishFailed reports a failure to publish a critical cache revision.
	CodeCacheCoordPublishFailed = bizerr.MustDefine(
		"CACHE_COORD_PUBLISH_FAILED",
		"Cache coordination revision publish failed for {domain}/{scope}",
		gcode.CodeInternalError,
	)
	// CodeCacheCoordRevisionUnavailable reports that freshness could not be confirmed.
	CodeCacheCoordRevisionUnavailable = bizerr.MustDefine(
		"CACHE_COORD_REVISION_UNAVAILABLE",
		"Cache coordination revision is unavailable for {domain}/{scope}",
		gcode.CodeInternalError,
	)
	// CodeCacheCoordFreshnessUnavailable reports that local cache freshness exceeded its allowed window.
	CodeCacheCoordFreshnessUnavailable = bizerr.MustDefine(
		"CACHE_COORD_FRESHNESS_UNAVAILABLE",
		"Cache freshness cannot be confirmed for {domain}/{scope}",
		gcode.CodeInternalError,
	)
)
