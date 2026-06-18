// This file defines reusable business errors shared by plugin-domain
// capability adapters.

package capmodel

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeCapabilityContextRequired reports a missing plugin capability context.
	CodeCapabilityContextRequired = bizerr.MustDefine(
		"CAPABILITY_CONTEXT_REQUIRED",
		"Capability context is required",
		gcode.CodeInvalidParameter,
	)
	// CodeCapabilityActorRequired reports a sensitive call without an actor.
	CodeCapabilityActorRequired = bizerr.MustDefine(
		"CAPABILITY_ACTOR_REQUIRED",
		"Capability actor is required",
		gcode.CodeNotAuthorized,
	)
	// CodeCapabilityDenied reports a denied capability method or target.
	CodeCapabilityDenied = bizerr.MustDefine(
		"CAPABILITY_DENIED",
		"Capability access is denied",
		gcode.CodeNotAuthorized,
	)
	// CodeCapabilityUnavailable reports a currently unavailable domain capability.
	CodeCapabilityUnavailable = bizerr.MustDefine(
		"CAPABILITY_UNAVAILABLE",
		"Capability {capability} is unavailable",
		gcode.CodeInternalError,
	)
	// CodeCapabilityProviderConflict reports that multiple provider plugins can serve one singleton capability.
	CodeCapabilityProviderConflict = bizerr.MustDefine(
		"CAPABILITY_PROVIDER_CONFLICT",
		"Multiple providers are available for capability {capability}: {providerIds}",
		gcode.CodeInvalidOperation,
	)
	// CodeCapabilityLimitExceeded reports a request that exceeds a domain limit.
	CodeCapabilityLimitExceeded = bizerr.MustDefine(
		"CAPABILITY_LIMIT_EXCEEDED",
		"Capability request exceeds limit {limit}",
		gcode.CodeInvalidParameter,
	)
)
