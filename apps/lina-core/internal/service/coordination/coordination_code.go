// This file re-exports coordination business error codes from the internal
// core contract package.

package coordination

import "lina-core/internal/service/coordination/internal/core"

var (
	// CodeCoordinationKeyInvalid reports an invalid coordination key.
	CodeCoordinationKeyInvalid = core.CodeCoordinationKeyInvalid
	// CodeCoordinationTTLInvalid reports an invalid coordination TTL.
	CodeCoordinationTTLInvalid = core.CodeCoordinationTTLInvalid
	// CodeCoordinationLockNotHeld reports that a lock handle no longer owns the lock.
	CodeCoordinationLockNotHeld = core.CodeCoordinationLockNotHeld
	// CodeCoordinationRedisUnavailable reports Redis connectivity failures.
	CodeCoordinationRedisUnavailable = core.CodeCoordinationRedisUnavailable
	// CodeCoordinationKVOperationFailed reports coordination KV failures.
	CodeCoordinationKVOperationFailed = core.CodeCoordinationKVOperationFailed
	// CodeCoordinationRevisionUnavailable reports revision read or write failures.
	CodeCoordinationRevisionUnavailable = core.CodeCoordinationRevisionUnavailable
	// CodeCoordinationEventPublishFailed reports event publish failures.
	CodeCoordinationEventPublishFailed = core.CodeCoordinationEventPublishFailed
	// CodeCoordinationEventSubscribeFailed reports event subscription failures.
	CodeCoordinationEventSubscribeFailed = core.CodeCoordinationEventSubscribeFailed
)
