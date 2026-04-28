// This file defines scheduled-job management business error codes and their
// i18n metadata.

package jobmgmt

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeJobGroupDefaultNotFound reports that the built-in default job group is missing.
	CodeJobGroupDefaultNotFound = bizerr.MustDefine(
		"JOB_GROUP_DEFAULT_NOT_FOUND",
		"Default scheduled-job group does not exist",
		gcode.CodeNotFound,
	)
	// CodeJobGroupNotFound reports that the requested scheduled-job group does not exist.
	CodeJobGroupNotFound = bizerr.MustDefine(
		"JOB_GROUP_NOT_FOUND",
		"Scheduled-job group does not exist",
		gcode.CodeNotFound,
	)
	// CodeJobGroupCodeRequired reports that a scheduled-job group code is required.
	CodeJobGroupCodeRequired = bizerr.MustDefine(
		"JOB_GROUP_CODE_REQUIRED",
		"Scheduled-job group code cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeJobGroupNameRequired reports that a scheduled-job group name is required.
	CodeJobGroupNameRequired = bizerr.MustDefine(
		"JOB_GROUP_NAME_REQUIRED",
		"Scheduled-job group name cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeJobGroupCodeExists reports that a scheduled-job group code is already used.
	CodeJobGroupCodeExists = bizerr.MustDefine(
		"JOB_GROUP_CODE_EXISTS",
		"Scheduled-job group code already exists",
		gcode.CodeInvalidParameter,
	)
	// CodeJobGroupDeleteRequired reports that no scheduled-job groups were selected for deletion.
	CodeJobGroupDeleteRequired = bizerr.MustDefine(
		"JOB_GROUP_DELETE_REQUIRED",
		"Please select scheduled-job groups to delete",
		gcode.CodeInvalidParameter,
	)
	// CodeJobGroupDefaultDeleteDenied reports that the default scheduled-job group cannot be deleted.
	CodeJobGroupDefaultDeleteDenied = bizerr.MustDefine(
		"JOB_GROUP_DEFAULT_DELETE_DENIED",
		"Default scheduled-job group cannot be deleted",
		gcode.CodeInvalidParameter,
	)
	// CodeJobGroupDeleteEmpty reports that no selected scheduled-job groups can be deleted.
	CodeJobGroupDeleteEmpty = bizerr.MustDefine(
		"JOB_GROUP_DELETE_EMPTY",
		"No scheduled-job groups can be deleted",
		gcode.CodeInvalidParameter,
	)
	// CodeJobCreateShellOnly reports that UI-created jobs must use shell execution.
	CodeJobCreateShellOnly = bizerr.MustDefine(
		"JOB_CREATE_SHELL_ONLY",
		"Only Shell scheduled jobs can be created from the UI",
		gcode.CodeInvalidParameter,
	)
	// CodeJobUpdateShellOnly reports that UI-edited jobs must use shell execution.
	CodeJobUpdateShellOnly = bizerr.MustDefine(
		"JOB_UPDATE_SHELL_ONLY",
		"Only Shell scheduled jobs can be edited from the UI",
		gcode.CodeInvalidParameter,
	)
	// CodeJobBuiltinUpdateDenied reports that source-registered scheduled jobs cannot be updated.
	CodeJobBuiltinUpdateDenied = bizerr.MustDefine(
		"JOB_BUILTIN_UPDATE_DENIED",
		"Source-registered scheduled jobs cannot be updated",
		gcode.CodeInvalidParameter,
	)
	// CodeJobBuiltinDeleteDenied reports that source-registered scheduled jobs cannot be deleted.
	CodeJobBuiltinDeleteDenied = bizerr.MustDefine(
		"JOB_BUILTIN_DELETE_DENIED",
		"Source-registered scheduled jobs cannot be deleted",
		gcode.CodeInvalidParameter,
	)
	// CodeJobBuiltinStatusUpdateDenied reports that source-registered scheduled jobs cannot change status.
	CodeJobBuiltinStatusUpdateDenied = bizerr.MustDefine(
		"JOB_BUILTIN_STATUS_UPDATE_DENIED",
		"Source-registered scheduled jobs cannot change status",
		gcode.CodeInvalidParameter,
	)
	// CodeJobBuiltinResetDenied reports that source-registered scheduled jobs cannot reset execution count.
	CodeJobBuiltinResetDenied = bizerr.MustDefine(
		"JOB_BUILTIN_RESET_DENIED",
		"Source-registered scheduled jobs cannot reset execution count",
		gcode.CodeInvalidParameter,
	)
	// CodeJobDeleteRequired reports that no scheduled jobs were selected for deletion.
	CodeJobDeleteRequired = bizerr.MustDefine(
		"JOB_DELETE_REQUIRED",
		"Please select scheduled jobs to delete",
		gcode.CodeInvalidParameter,
	)
	// CodeJobNameRequired reports that a scheduled-job name is required.
	CodeJobNameRequired = bizerr.MustDefine(
		"JOB_NAME_REQUIRED",
		"Scheduled-job name cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeJobNameTooLong reports that a scheduled-job name exceeds the supported length.
	CodeJobNameTooLong = bizerr.MustDefine(
		"JOB_NAME_TOO_LONG",
		"Scheduled-job name cannot exceed 128 characters",
		gcode.CodeInvalidParameter,
	)
	// CodeJobNameExistsInGroup reports that a scheduled-job name is already used in the group.
	CodeJobNameExistsInGroup = bizerr.MustDefine(
		"JOB_NAME_EXISTS_IN_GROUP",
		"Scheduled-job name already exists in the current group",
		gcode.CodeInvalidParameter,
	)
	// CodeJobTimeoutSecondAlignedRequired reports that scheduled-job timeout must use whole seconds.
	CodeJobTimeoutSecondAlignedRequired = bizerr.MustDefine(
		"JOB_TIMEOUT_SECOND_ALIGNED_REQUIRED",
		"Scheduled-job timeout must be configured in whole seconds",
		gcode.CodeInvalidParameter,
	)
	// CodeJobTimeoutOutOfRange reports that scheduled-job timeout is outside the allowed range.
	CodeJobTimeoutOutOfRange = bizerr.MustDefine(
		"JOB_TIMEOUT_OUT_OF_RANGE",
		"Scheduled-job timeout must be between 1 and 86400 seconds",
		gcode.CodeInvalidParameter,
	)
	// CodeJobTaskTypeInvalid reports that a scheduled-job task type is invalid for persistence.
	CodeJobTaskTypeInvalid = bizerr.MustDefine(
		"JOB_TASK_TYPE_INVALID",
		"Scheduled-job task type only supports handler or shell",
		gcode.CodeInvalidParameter,
	)
	// CodeJobScopeInvalid reports that a scheduled-job dispatch scope is invalid.
	CodeJobScopeInvalid = bizerr.MustDefine(
		"JOB_SCOPE_INVALID",
		"Scheduled-job scope only supports master_only or all_node",
		gcode.CodeInvalidParameter,
	)
	// CodeJobConcurrencyInvalid reports that a scheduled-job concurrency policy is invalid.
	CodeJobConcurrencyInvalid = bizerr.MustDefine(
		"JOB_CONCURRENCY_INVALID",
		"Scheduled-job concurrency only supports singleton or parallel",
		gcode.CodeInvalidParameter,
	)
	// CodeJobStatusInvalid reports that a scheduled-job status is invalid for persistence.
	CodeJobStatusInvalid = bizerr.MustDefine(
		"JOB_STATUS_INVALID",
		"Scheduled-job status only supports enabled or disabled",
		gcode.CodeInvalidParameter,
	)
	// CodeJobStatusToggleInvalid reports that a status update only supports enabled or disabled.
	CodeJobStatusToggleInvalid = bizerr.MustDefine(
		"JOB_STATUS_TOGGLE_INVALID",
		"Scheduled-job status only supports enabled or disabled",
		gcode.CodeInvalidParameter,
	)
	// CodeJobMaxExecutionsInvalid reports that max executions must not be negative.
	CodeJobMaxExecutionsInvalid = bizerr.MustDefine(
		"JOB_MAX_EXECUTIONS_INVALID",
		"Maximum executions must be an integer greater than or equal to 0",
		gcode.CodeInvalidParameter,
	)
	// CodeJobMaxConcurrencyInvalid reports that max concurrency is outside the allowed range.
	CodeJobMaxConcurrencyInvalid = bizerr.MustDefine(
		"JOB_MAX_CONCURRENCY_INVALID",
		"Maximum concurrency must be an integer between 1 and 100",
		gcode.CodeInvalidParameter,
	)
	// CodeJobParamsMarshalFailed reports that handler parameters cannot be serialized.
	CodeJobParamsMarshalFailed = bizerr.MustDefine(
		"JOB_PARAMS_MARSHAL_FAILED",
		"Failed to serialize scheduled-job parameters",
		gcode.CodeInternalError,
	)
	// CodeJobShellEnvMarshalFailed reports that shell environment variables cannot be serialized.
	CodeJobShellEnvMarshalFailed = bizerr.MustDefine(
		"JOB_SHELL_ENV_MARSHAL_FAILED",
		"Failed to serialize Shell environment variables",
		gcode.CodeInternalError,
	)
	// CodeJobSchedulerUninitialized reports that the scheduled-job scheduler is unavailable.
	CodeJobSchedulerUninitialized = bizerr.MustDefine(
		"JOB_SCHEDULER_UNINITIALIZED",
		"Scheduled-job scheduler is not initialized",
		gcode.CodeInternalError,
	)
	// CodeJobLogNotFound reports that the requested scheduled-job execution log does not exist.
	CodeJobLogNotFound = bizerr.MustDefine(
		"JOB_LOG_NOT_FOUND",
		"Scheduled-job execution log does not exist",
		gcode.CodeNotFound,
	)
	// CodeJobLogCurrentUserMissing reports that the current request has no authenticated user context.
	CodeJobLogCurrentUserMissing = bizerr.MustDefine(
		"JOB_LOG_CURRENT_USER_MISSING",
		"Current authenticated user is unavailable",
		gcode.CodeNotAuthorized,
	)
	// CodeJobLogShellCancelPermissionDenied reports that the current user cannot cancel shell job logs.
	CodeJobLogShellCancelPermissionDenied = bizerr.MustDefine(
		"JOB_LOG_SHELL_CANCEL_PERMISSION_DENIED",
		"Current user lacks required API permissions: {permission}",
		gcode.CodeNotAuthorized,
	)
	// CodeJobCronExpressionRequired reports that a cron expression is required.
	CodeJobCronExpressionRequired = bizerr.MustDefine(
		"JOB_CRON_EXPRESSION_REQUIRED",
		"Scheduled-job cron expression cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeJobCronExpressionTooLong reports that a cron expression exceeds the supported length.
	CodeJobCronExpressionTooLong = bizerr.MustDefine(
		"JOB_CRON_EXPRESSION_TOO_LONG",
		"Scheduled-job cron expression cannot exceed 128 characters",
		gcode.CodeInvalidParameter,
	)
	// CodeJobCronExpressionInvalid reports that a cron expression cannot be parsed.
	CodeJobCronExpressionInvalid = bizerr.MustDefine(
		"JOB_CRON_EXPRESSION_INVALID",
		"Scheduled-job cron expression is invalid: {reason}",
		gcode.CodeInvalidParameter,
	)
	// CodeJobCronSecondsRequired reports that a six-field cron expression requires a concrete seconds field.
	CodeJobCronSecondsRequired = bizerr.MustDefine(
		"JOB_CRON_SECONDS_REQUIRED",
		"The seconds field in a six-field cron expression must be concrete; use a five-field expression without #",
		gcode.CodeInvalidParameter,
	)
	// CodeJobCronFieldCountInvalid reports that a cron expression has an unsupported field count.
	CodeJobCronFieldCountInvalid = bizerr.MustDefine(
		"JOB_CRON_FIELD_COUNT_INVALID",
		"Scheduled-job cron expression only supports five or six fields",
		gcode.CodeInvalidParameter,
	)
	// CodeJobTimezoneInvalid reports that a scheduled-job timezone is invalid.
	CodeJobTimezoneInvalid = bizerr.MustDefine(
		"JOB_TIMEZONE_INVALID",
		"Scheduled-job timezone is invalid: {timezone}",
		gcode.CodeInvalidParameter,
	)
	// CodeJobBuiltinGroupNotFound reports that a source-registered job references a missing group.
	CodeJobBuiltinGroupNotFound = bizerr.MustDefine(
		"JOB_BUILTIN_GROUP_NOT_FOUND",
		"Scheduled-job group does not exist: {groupCode}",
		gcode.CodeInvalidParameter,
	)
	// CodeJobBuiltinTypeUnsupported reports that a source-registered job uses an unsupported task type.
	CodeJobBuiltinTypeUnsupported = bizerr.MustDefine(
		"JOB_BUILTIN_TYPE_UNSUPPORTED",
		"Source-registered scheduled-job task type is not supported",
		gcode.CodeInvalidParameter,
	)
	// CodeJobBuiltinNameRequired reports that a source-registered job name is required.
	CodeJobBuiltinNameRequired = bizerr.MustDefine(
		"JOB_BUILTIN_NAME_REQUIRED",
		"Source-registered scheduled-job name cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeJobBuiltinTimeoutSecondAlignedRequired reports that a source-registered job timeout must use whole seconds.
	CodeJobBuiltinTimeoutSecondAlignedRequired = bizerr.MustDefine(
		"JOB_BUILTIN_TIMEOUT_SECOND_ALIGNED_REQUIRED",
		"Source-registered scheduled-job timeout must be configured in whole seconds",
		gcode.CodeInvalidParameter,
	)
	// CodeJobBuiltinCronExpressionRequired reports that a source-registered job cron expression is required.
	CodeJobBuiltinCronExpressionRequired = bizerr.MustDefine(
		"JOB_BUILTIN_CRON_EXPRESSION_REQUIRED",
		"Source-registered scheduled-job cron expression cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeJobBuiltinCronExpressionTooLong reports that a source-registered job cron expression is too long.
	CodeJobBuiltinCronExpressionTooLong = bizerr.MustDefine(
		"JOB_BUILTIN_CRON_EXPRESSION_TOO_LONG",
		"Source-registered scheduled-job cron expression cannot exceed 128 characters",
		gcode.CodeInvalidParameter,
	)
	// CodeJobBuiltinMaxExecutionsInvalid reports that source-registered max executions must not be negative.
	CodeJobBuiltinMaxExecutionsInvalid = bizerr.MustDefine(
		"JOB_BUILTIN_MAX_EXECUTIONS_INVALID",
		"Source-registered scheduled-job maximum executions cannot be negative",
		gcode.CodeInvalidParameter,
	)
	// CodeJobBuiltinHandlerRefRequired reports that a source-registered handler job has no handler ref.
	CodeJobBuiltinHandlerRefRequired = bizerr.MustDefine(
		"JOB_BUILTIN_HANDLER_REF_REQUIRED",
		"Source-registered scheduled-job handler reference cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeJobBuiltinParamsMarshalFailed reports that source-registered job params cannot be serialized.
	CodeJobBuiltinParamsMarshalFailed = bizerr.MustDefine(
		"JOB_BUILTIN_PARAMS_MARSHAL_FAILED",
		"Failed to serialize source-registered scheduled-job parameters",
		gcode.CodeInternalError,
	)
)
