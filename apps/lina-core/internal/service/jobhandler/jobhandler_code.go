// This file defines scheduled-job handler business error codes and their i18n
// metadata.

package jobhandler

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeJobHandlerRegistryRequired reports that a handler registry dependency is missing.
	CodeJobHandlerRegistryRequired = bizerr.MustDefine(
		"JOB_HANDLER_REGISTRY_REQUIRED",
		"Scheduled-job handler registry cannot be empty",
		gcode.CodeInternalError,
	)
	// CodeJobHandlerLogCleanerRequired reports that the host log cleaner dependency is missing.
	CodeJobHandlerLogCleanerRequired = bizerr.MustDefine(
		"JOB_HANDLER_LOG_CLEANER_REQUIRED",
		"Scheduled-job log cleaner cannot be empty",
		gcode.CodeInternalError,
	)
	// CodeJobHandlerLifecycleBridgeRequired reports that the plugin lifecycle bridge dependency is missing.
	CodeJobHandlerLifecycleBridgeRequired = bizerr.MustDefine(
		"JOB_HANDLER_LIFECYCLE_BRIDGE_REQUIRED",
		"Plugin lifecycle bridge cannot be empty",
		gcode.CodeInternalError,
	)
	// CodeJobHandlerRefRequired reports that a handler reference is required.
	CodeJobHandlerRefRequired = bizerr.MustDefine(
		"JOB_HANDLER_REF_REQUIRED",
		"Scheduled-job handler reference cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerDisplayNameRequired reports that a handler display name is required.
	CodeJobHandlerDisplayNameRequired = bizerr.MustDefine(
		"JOB_HANDLER_DISPLAY_NAME_REQUIRED",
		"Scheduled-job handler display name cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerCallbackRequired reports that a handler callback is required.
	CodeJobHandlerCallbackRequired = bizerr.MustDefine(
		"JOB_HANDLER_CALLBACK_REQUIRED",
		"Scheduled-job handler callback cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerSourceUnsupported reports that a handler source is unsupported.
	CodeJobHandlerSourceUnsupported = bizerr.MustDefine(
		"JOB_HANDLER_SOURCE_UNSUPPORTED",
		"Scheduled-job handler source is not supported",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerPluginIDRequired reports that a plugin handler has no plugin ID.
	CodeJobHandlerPluginIDRequired = bizerr.MustDefine(
		"JOB_HANDLER_PLUGIN_ID_REQUIRED",
		"Plugin scheduled-job handler must declare plugin ID",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerExists reports that a handler reference is already registered.
	CodeJobHandlerExists = bizerr.MustDefine(
		"JOB_HANDLER_EXISTS",
		"Scheduled-job handler {ref} already exists",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerNotFound reports that the requested scheduled-job handler does not exist.
	CodeJobHandlerNotFound = bizerr.MustDefine(
		"JOB_HANDLER_NOT_FOUND",
		"Scheduled job handler does not exist",
		gcode.CodeNotFound,
	)
	// CodeJobHandlerSchemaParseFailed reports that a handler parameter schema cannot be parsed.
	CodeJobHandlerSchemaParseFailed = bizerr.MustDefine(
		"JOB_HANDLER_SCHEMA_PARSE_FAILED",
		"Failed to parse scheduled-job handler parameter schema",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerSchemaSingleObjectRequired reports that a schema has trailing JSON content.
	CodeJobHandlerSchemaSingleObjectRequired = bizerr.MustDefine(
		"JOB_HANDLER_SCHEMA_SINGLE_OBJECT_REQUIRED",
		"Scheduled-job handler parameter schema must contain exactly one JSON object",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerSchemaRootTypeInvalid reports that a schema root type is not object.
	CodeJobHandlerSchemaRootTypeInvalid = bizerr.MustDefine(
		"JOB_HANDLER_SCHEMA_ROOT_TYPE_INVALID",
		"Scheduled-job handler parameter schema root must declare type=object",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerSchemaEmptyPropertyName reports that a schema contains an empty property name.
	CodeJobHandlerSchemaEmptyPropertyName = bizerr.MustDefine(
		"JOB_HANDLER_SCHEMA_EMPTY_PROPERTY_NAME",
		"Scheduled-job handler parameter schema cannot contain empty property names",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerSchemaRequiredFieldUnknown reports that required references a missing property.
	CodeJobHandlerSchemaRequiredFieldUnknown = bizerr.MustDefine(
		"JOB_HANDLER_SCHEMA_REQUIRED_FIELD_UNKNOWN",
		"Scheduled-job handler parameter schema required field {field} is not declared in properties",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerSchemaFieldTypeUnsupported reports that a schema field type is unsupported.
	CodeJobHandlerSchemaFieldTypeUnsupported = bizerr.MustDefine(
		"JOB_HANDLER_SCHEMA_FIELD_TYPE_UNSUPPORTED",
		"Scheduled-job handler parameter {field} type {fieldType} is not supported",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerSchemaFieldFormatUnsupported reports that a schema field format is unsupported.
	CodeJobHandlerSchemaFieldFormatUnsupported = bizerr.MustDefine(
		"JOB_HANDLER_SCHEMA_FIELD_FORMAT_UNSUPPORTED",
		"Scheduled-job handler parameter {field} format {format} is not supported",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerSchemaFieldFormatTypeInvalid reports that a non-string schema field declares format.
	CodeJobHandlerSchemaFieldFormatTypeInvalid = bizerr.MustDefine(
		"JOB_HANDLER_SCHEMA_FIELD_FORMAT_TYPE_INVALID",
		"Scheduled-job handler parameter {field} can declare format only when type is string",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerSchemaEnumValueInvalid reports that a schema enum value does not match its type.
	CodeJobHandlerSchemaEnumValueInvalid = bizerr.MustDefine(
		"JOB_HANDLER_SCHEMA_ENUM_VALUE_INVALID",
		"Scheduled-job handler parameter {field} enum value is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerParamRequiredMissing reports that a required handler parameter is missing.
	CodeJobHandlerParamRequiredMissing = bizerr.MustDefine(
		"JOB_HANDLER_PARAM_REQUIRED_MISSING",
		"Scheduled-job handler parameter is missing required field {field}",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerParamsParseFailed reports that handler parameter JSON cannot be parsed.
	CodeJobHandlerParamsParseFailed = bizerr.MustDefine(
		"JOB_HANDLER_PARAMS_PARSE_FAILED",
		"Failed to parse scheduled-job handler parameters",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerParamsSingleObjectRequired reports that params have trailing JSON content.
	CodeJobHandlerParamsSingleObjectRequired = bizerr.MustDefine(
		"JOB_HANDLER_PARAMS_SINGLE_OBJECT_REQUIRED",
		"Scheduled-job handler parameters must contain exactly one JSON object",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerParamTypeMismatch reports that a handler parameter value has the wrong type.
	CodeJobHandlerParamTypeMismatch = bizerr.MustDefine(
		"JOB_HANDLER_PARAM_TYPE_MISMATCH",
		"Scheduled-job handler parameter {field} type does not match the schema",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerParamEnumInvalid reports that a handler parameter value is outside its enum.
	CodeJobHandlerParamEnumInvalid = bizerr.MustDefine(
		"JOB_HANDLER_PARAM_ENUM_INVALID",
		"Scheduled-job handler parameter {field} is not in the allowed enum values",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerParamDateFormatInvalid reports that a handler parameter is not a YYYY-MM-DD date.
	CodeJobHandlerParamDateFormatInvalid = bizerr.MustDefine(
		"JOB_HANDLER_PARAM_DATE_FORMAT_INVALID",
		"Scheduled-job handler parameter {field} must use YYYY-MM-DD date format",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerParamDateTimeFormatInvalid reports that a handler parameter is not RFC3339 date-time.
	CodeJobHandlerParamDateTimeFormatInvalid = bizerr.MustDefine(
		"JOB_HANDLER_PARAM_DATE_TIME_FORMAT_INVALID",
		"Scheduled-job handler parameter {field} must use RFC3339 date-time format",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerValueUnsupported reports that a dynamic JSON value is unsupported for its schema type.
	CodeJobHandlerValueUnsupported = bizerr.MustDefine(
		"JOB_HANDLER_VALUE_UNSUPPORTED",
		"Unsupported scheduled-job handler parameter value: {value}",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerJSONTrailingContent reports that a JSON decoder found extra content.
	CodeJobHandlerJSONTrailingContent = bizerr.MustDefine(
		"JOB_HANDLER_JSON_TRAILING_CONTENT",
		"Detected trailing JSON content",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerWaitParamsParseFailed reports that built-in wait handler params cannot be parsed.
	CodeJobHandlerWaitParamsParseFailed = bizerr.MustDefine(
		"JOB_HANDLER_WAIT_PARAMS_PARSE_FAILED",
		"Failed to parse wait handler parameters",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerWaitSecondsInvalid reports that the built-in wait handler seconds value is invalid.
	CodeJobHandlerWaitSecondsInvalid = bizerr.MustDefine(
		"JOB_HANDLER_WAIT_SECONDS_INVALID",
		"Wait seconds must be greater than 0",
		gcode.CodeInvalidParameter,
	)
)
