// This file defines shared scheduled-job domain business error codes and their
// i18n metadata.

package jobmeta

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeJobNotFound reports that the requested scheduled job does not exist.
	CodeJobNotFound = bizerr.MustDefine(
		"JOB_NOT_FOUND",
		"Scheduled job does not exist",
		gcode.CodeNotFound,
	)
	// CodeJobTaskTypeUnsupported reports that a scheduled-job task type is unsupported.
	CodeJobTaskTypeUnsupported = bizerr.MustDefine(
		"JOB_TASK_TYPE_UNSUPPORTED",
		"Scheduled job task type is not supported",
		gcode.CodeInvalidParameter,
	)
	// CodeJobCronFieldCountUnsupported reports that a cron expression cannot be normalized for the scheduler.
	CodeJobCronFieldCountUnsupported = bizerr.MustDefine(
		"JOB_CRON_FIELD_COUNT_UNSUPPORTED",
		"Cron expression field count is not supported",
		gcode.CodeInvalidParameter,
	)
	// CodeJobHandlerUnavailable reports that a plugin-owned handler is currently unavailable.
	CodeJobHandlerUnavailable = bizerr.MustDefine(
		"JOB_HANDLER_UNAVAILABLE",
		"Plugin handler is currently unavailable and cannot be triggered manually",
		gcode.CodeInvalidParameter,
	)
	// CodeJobSnapshotMarshalFailed reports that a job snapshot cannot be serialized for execution logging.
	CodeJobSnapshotMarshalFailed = bizerr.MustDefine(
		"JOB_SNAPSHOT_MARSHAL_FAILED",
		"Failed to serialize scheduled-job snapshot",
		gcode.CodeInternalError,
	)
	// CodeJobLogNotRunning reports that the requested scheduled-job execution log is not running.
	CodeJobLogNotRunning = bizerr.MustDefine(
		"JOB_LOG_NOT_RUNNING",
		"Scheduled-job execution log is not running",
		gcode.CodeInvalidParameter,
	)
	// CodeJobShellExecutorUninitialized reports that shell execution was requested without an executor.
	CodeJobShellExecutorUninitialized = bizerr.MustDefine(
		"JOB_SHELL_EXECUTOR_UNINITIALIZED",
		"Shell executor is not initialized",
		gcode.CodeInternalError,
	)
	// CodeJobShellDisabled reports that shell scheduled jobs are disabled in the current environment.
	CodeJobShellDisabled = bizerr.MustDefine(
		"JOB_SHELL_DISABLED",
		"Shell scheduled jobs are disabled in the current environment",
		gcode.CodeInvalidParameter,
	)
	// CodeJobShellCommandRequired reports that a shell scheduled job has no command text.
	CodeJobShellCommandRequired = bizerr.MustDefine(
		"JOB_SHELL_COMMAND_REQUIRED",
		"Shell command cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeJobShellTimeoutInvalid reports that a shell scheduled-job timeout is invalid.
	CodeJobShellTimeoutInvalid = bizerr.MustDefine(
		"JOB_SHELL_TIMEOUT_INVALID",
		"Shell scheduled-job timeout must be greater than 0",
		gcode.CodeInvalidParameter,
	)
	// CodeJobShellStartFailed reports that the host failed to start a shell scheduled job.
	CodeJobShellStartFailed = bizerr.MustDefine(
		"JOB_SHELL_START_FAILED",
		"Failed to start shell scheduled job",
		gcode.CodeInternalError,
	)
	// CodeJobShellExecutionFailed reports that a shell scheduled job exited with an execution error.
	CodeJobShellExecutionFailed = bizerr.MustDefine(
		"JOB_SHELL_EXECUTION_FAILED",
		"Failed to execute shell scheduled job",
		gcode.CodeInternalError,
	)
	// CodeJobShellWorkdirRootDenied reports that the shell work directory cannot be the filesystem root.
	CodeJobShellWorkdirRootDenied = bizerr.MustDefine(
		"JOB_SHELL_WORKDIR_ROOT_DENIED",
		"Shell working directory cannot be the filesystem root",
		gcode.CodeInvalidParameter,
	)
	// CodeJobShellWorkdirValidateFailed reports that the shell work directory cannot be inspected.
	CodeJobShellWorkdirValidateFailed = bizerr.MustDefine(
		"JOB_SHELL_WORKDIR_VALIDATE_FAILED",
		"Failed to validate shell working directory",
		gcode.CodeInvalidParameter,
	)
	// CodeJobShellWorkdirNotDirectory reports that the shell work directory path is not a directory.
	CodeJobShellWorkdirNotDirectory = bizerr.MustDefine(
		"JOB_SHELL_WORKDIR_NOT_DIRECTORY",
		"Shell working directory must be a directory",
		gcode.CodeInvalidParameter,
	)
	// CodeJobRetentionParseFailed reports that a retention policy JSON payload cannot be parsed.
	CodeJobRetentionParseFailed = bizerr.MustDefine(
		"JOB_RETENTION_PARSE_FAILED",
		"Failed to parse scheduled-job log retention policy",
		gcode.CodeInvalidParameter,
	)
	// CodeJobRetentionModeUnsupported reports that a retention policy mode is unsupported.
	CodeJobRetentionModeUnsupported = bizerr.MustDefine(
		"JOB_RETENTION_MODE_UNSUPPORTED",
		"Scheduled-job log retention policy mode is not supported",
		gcode.CodeInvalidParameter,
	)
	// CodeJobRetentionValueInvalid reports that a retention policy threshold is invalid.
	CodeJobRetentionValueInvalid = bizerr.MustDefine(
		"JOB_RETENTION_VALUE_INVALID",
		"Scheduled-job log retention policy threshold must be greater than 0",
		gcode.CodeInvalidParameter,
	)
	// CodeJobRetentionMarshalFailed reports that a retention policy cannot be serialized.
	CodeJobRetentionMarshalFailed = bizerr.MustDefine(
		"JOB_RETENTION_MARSHAL_FAILED",
		"Failed to serialize scheduled-job log retention policy",
		gcode.CodeInternalError,
	)
)
