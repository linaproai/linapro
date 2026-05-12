// This file defines coordination business error codes.

package core

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeCoordinationKeyInvalid reports an invalid coordination key.
	CodeCoordinationKeyInvalid = bizerr.MustDefine(
		"COORDINATION_KEY_INVALID",
		"Coordination key is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeCoordinationTTLInvalid reports an invalid coordination TTL.
	CodeCoordinationTTLInvalid = bizerr.MustDefine(
		"COORDINATION_TTL_INVALID",
		"Coordination TTL is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeCoordinationLockNotHeld reports that a lock handle no longer owns the lock.
	CodeCoordinationLockNotHeld = bizerr.MustDefine(
		"COORDINATION_LOCK_NOT_HELD",
		"Coordination lock is not held by the current owner",
		gcode.CodeInternalError,
	)
	// CodeCoordinationRedisUnavailable reports Redis connectivity failures.
	CodeCoordinationRedisUnavailable = bizerr.MustDefine(
		"COORDINATION_REDIS_UNAVAILABLE",
		"Redis coordination backend is unavailable",
		gcode.CodeInternalError,
	)
	// CodeCoordinationKVOperationFailed reports coordination KV failures.
	CodeCoordinationKVOperationFailed = bizerr.MustDefine(
		"COORDINATION_KV_OPERATION_FAILED",
		"Coordination KV operation failed",
		gcode.CodeInternalError,
	)
	// CodeCoordinationRevisionUnavailable reports revision read or write failures.
	CodeCoordinationRevisionUnavailable = bizerr.MustDefine(
		"COORDINATION_REVISION_UNAVAILABLE",
		"Coordination revision is unavailable",
		gcode.CodeInternalError,
	)
	// CodeCoordinationEventPublishFailed reports event publish failures.
	CodeCoordinationEventPublishFailed = bizerr.MustDefine(
		"COORDINATION_EVENT_PUBLISH_FAILED",
		"Coordination event publish failed",
		gcode.CodeInternalError,
	)
	// CodeCoordinationEventSubscribeFailed reports event subscription failures.
	CodeCoordinationEventSubscribeFailed = bizerr.MustDefine(
		"COORDINATION_EVENT_SUBSCRIBE_FAILED",
		"Coordination event subscription failed",
		gcode.CodeInternalError,
	)
)
