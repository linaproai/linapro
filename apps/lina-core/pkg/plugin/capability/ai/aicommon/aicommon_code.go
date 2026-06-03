// This file defines shared multimodal AI capability business error codes used
// by fallback services and host-service boundary checks.

package aicommon

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeProviderUnavailable reports that no usable AI provider is available.
	CodeProviderUnavailable = bizerr.MustDefine(
		"AI_PROVIDER_UNAVAILABLE",
		"AI provider is unavailable for {capabilityType}.{capabilityMethod}",
		gcode.CodeInvalidOperation,
	)
	// CodePurposeRequired reports that a governed AI purpose was omitted.
	CodePurposeRequired = bizerr.MustDefine(
		"AI_PURPOSE_REQUIRED",
		"AI purpose is required",
		gcode.CodeInvalidParameter,
	)
	// CodeTierInvalid reports that a caller supplied an unsupported tier.
	CodeTierInvalid = bizerr.MustDefine(
		"AI_TIER_INVALID",
		"AI tier {tier} is not supported",
		gcode.CodeInvalidParameter,
	)
	// CodeAssetRefRequired reports that an asset reference is required.
	CodeAssetRefRequired = bizerr.MustDefine(
		"AI_ASSET_REF_REQUIRED",
		"AI asset reference is required",
		gcode.CodeInvalidParameter,
	)
	// CodeOperationRefRequired reports that a provider operation reference is required.
	CodeOperationRefRequired = bizerr.MustDefine(
		"AI_OPERATION_REF_REQUIRED",
		"AI provider operation reference is required",
		gcode.CodeInvalidParameter,
	)
	// CodeUnsupportedMethod reports that a method is intentionally outside the AI boundary.
	CodeUnsupportedMethod = bizerr.MustDefine(
		"AI_METHOD_UNSUPPORTED",
		"AI method {method} is not supported by the typed multimodal capability boundary",
		gcode.CodeInvalidParameter,
	)
)
