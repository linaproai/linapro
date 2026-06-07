// This file defines shared text AI capability business error codes.

package aitext

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeTextProviderUnavailable reports that no usable text AI provider is available.
	CodeTextProviderUnavailable = bizerr.MustDefine(
		"AI_TEXT_PROVIDER_UNAVAILABLE",
		"Text AI provider is unavailable",
		gcode.CodeInvalidOperation,
	)
	// CodeTextTierInvalid reports that a caller supplied an unsupported text AI tier.
	CodeTextTierInvalid = bizerr.MustDefine(
		"AI_TEXT_TIER_INVALID",
		"Text AI tier {tier} is not supported",
		gcode.CodeInvalidParameter,
	)
	// CodeTextMessagesRequired reports that a text generation request has no messages.
	CodeTextMessagesRequired = bizerr.MustDefine(
		"AI_TEXT_MESSAGES_REQUIRED",
		"Text AI generation requires at least one message",
		gcode.CodeInvalidParameter,
	)
	// CodeTextMessageRoleInvalid reports that a message role is unsupported.
	CodeTextMessageRoleInvalid = bizerr.MustDefine(
		"AI_TEXT_MESSAGE_ROLE_INVALID",
		"Text AI message role {role} is not supported",
		gcode.CodeInvalidParameter,
	)
	// CodeTextThinkingEffortInvalid reports an unsupported thinking effort value.
	CodeTextThinkingEffortInvalid = bizerr.MustDefine(
		"AI_TEXT_THINKING_EFFORT_INVALID",
		"Text AI thinking effort {effort} is not supported",
		gcode.CodeInvalidParameter,
	)
	// CodeTextPurposeRequired reports that a governed AI purpose was omitted.
	CodeTextPurposeRequired = bizerr.MustDefine(
		"AI_TEXT_PURPOSE_REQUIRED",
		"Text AI purpose is required",
		gcode.CodeInvalidParameter,
	)
	// CodeTextMetadataTooLarge reports metadata that exceeds the capability boundary.
	CodeTextMetadataTooLarge = bizerr.MustDefine(
		"AI_TEXT_METADATA_TOO_LARGE",
		"Text AI metadata exceeds the allowed size",
		gcode.CodeInvalidParameter,
	)
	// CodeTextMaxOutputTokensInvalid reports an invalid output-token bound.
	CodeTextMaxOutputTokensInvalid = bizerr.MustDefine(
		"AI_TEXT_MAX_OUTPUT_TOKENS_INVALID",
		"Text AI max output tokens must be greater than or equal to zero",
		gcode.CodeInvalidParameter,
	)
)
