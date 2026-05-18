// This file defines runtime-configuration business error codes and their i18n
// metadata.

package config

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeConfigParamRequired reports that a runtime configuration value is required.
	CodeConfigParamRequired = bizerr.MustDefine(
		"CONFIG_PARAM_REQUIRED",
		"Parameter {key} cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeConfigParamDurationInvalid reports that a runtime configuration value is not a valid duration.
	CodeConfigParamDurationInvalid = bizerr.MustDefine(
		"CONFIG_PARAM_DURATION_INVALID",
		"Parameter {key} must be a valid duration",
		gcode.CodeInvalidParameter,
	)
	// CodeConfigParamPositiveRequired reports that a runtime configuration value must be positive.
	CodeConfigParamPositiveRequired = bizerr.MustDefine(
		"CONFIG_PARAM_POSITIVE_REQUIRED",
		"Parameter {key} must be greater than 0",
		gcode.CodeInvalidParameter,
	)
	// CodeConfigParamIntegerInvalid reports that a runtime configuration value must be an integer.
	CodeConfigParamIntegerInvalid = bizerr.MustDefine(
		"CONFIG_PARAM_INTEGER_INVALID",
		"Parameter {key} must be an integer",
		gcode.CodeInvalidParameter,
	)
	// CodeConfigParamIPCIDRInvalid reports that a runtime configuration value contains an invalid IP or CIDR entry.
	CodeConfigParamIPCIDRInvalid = bizerr.MustDefine(
		"CONFIG_PARAM_IP_CIDR_INVALID",
		"Parameter {key} contains invalid IP or CIDR value: {value}",
		gcode.CodeInvalidParameter,
	)
	// CodeConfigParamBoolInvalid reports that a runtime configuration value must be true or false.
	CodeConfigParamBoolInvalid = bizerr.MustDefine(
		"CONFIG_PARAM_BOOL_INVALID",
		"Parameter {key} must be true or false",
		gcode.CodeInvalidParameter,
	)
	// CodeConfigParamAllowedValueInvalid reports that a runtime configuration value is outside its whitelist.
	CodeConfigParamAllowedValueInvalid = bizerr.MustDefine(
		"CONFIG_PARAM_ALLOWED_VALUE_INVALID",
		"Parameter {key} is not in the supported range",
		gcode.CodeInvalidParameter,
	)
	// CodeConfigParamTextTooLong reports that a runtime configuration text value exceeds its character limit.
	CodeConfigParamTextTooLong = bizerr.MustDefine(
		"CONFIG_PARAM_TEXT_TOO_LONG",
		"Parameter {key} cannot exceed {maxLen} characters",
		gcode.CodeInvalidParameter,
	)
	// CodeConfigParamJSONObjectInvalid reports that a runtime configuration value is not a valid JSON object.
	CodeConfigParamJSONObjectInvalid = bizerr.MustDefine(
		"CONFIG_PARAM_JSON_OBJECT_INVALID",
		"Parameter {key} must be a valid JSON object",
		gcode.CodeInvalidParameter,
	)
	// CodeConfigParamJSONSingleObjectRequired reports that a runtime configuration JSON value has trailing content.
	CodeConfigParamJSONSingleObjectRequired = bizerr.MustDefine(
		"CONFIG_PARAM_JSON_SINGLE_OBJECT_REQUIRED",
		"Parameter {key} must contain exactly one JSON object",
		gcode.CodeInvalidParameter,
	)
	// CodeConfigParamJSONTrailingContent reports that a runtime configuration JSON value has unreadable trailing content.
	CodeConfigParamJSONTrailingContent = bizerr.MustDefine(
		"CONFIG_PARAM_JSON_TRAILING_CONTENT",
		"Parameter {key} contains trailing JSON content",
		gcode.CodeInvalidParameter,
	)
	// CodeConfigCronRetentionValuePositiveRequired reports that cron retention value must be positive.
	CodeConfigCronRetentionValuePositiveRequired = bizerr.MustDefine(
		"CONFIG_CRON_RETENTION_VALUE_POSITIVE_REQUIRED",
		"Parameter {key} value must be greater than 0",
		gcode.CodeInvalidParameter,
	)
	// CodeConfigCronRetentionValueNonNegativeRequired reports that cron retention none-mode value cannot be negative.
	CodeConfigCronRetentionValueNonNegativeRequired = bizerr.MustDefine(
		"CONFIG_CRON_RETENTION_VALUE_NON_NEGATIVE_REQUIRED",
		"Parameter {key} value cannot be negative",
		gcode.CodeInvalidParameter,
	)
	// CodeConfigCronRetentionModeUnsupported reports that cron retention mode is unsupported.
	CodeConfigCronRetentionModeUnsupported = bizerr.MustDefine(
		"CONFIG_CRON_RETENTION_MODE_UNSUPPORTED",
		"Parameter {key} mode is not supported",
		gcode.CodeInvalidParameter,
	)
	// CodeConfigRuntimeParamRevisionUnavailable reports that runtime-configuration freshness cannot be checked.
	CodeConfigRuntimeParamRevisionUnavailable = bizerr.MustDefine(
		"CONFIG_RUNTIME_PARAM_REVISION_UNAVAILABLE",
		"Runtime configuration revision is unavailable",
		gcode.CodeInternalError,
	)
)
