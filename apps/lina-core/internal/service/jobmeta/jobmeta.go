// Package jobmeta defines shared scheduled-job domain enums and JSON payload
// helpers used by job management, scheduling, and handler execution.
package jobmeta

import (
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// TaskType identifies the scheduled-job execution mode.
type TaskType string

// Supported scheduled-job execution modes.
const (
	TaskTypeHandler TaskType = "handler" // Handler-based host or plugin task.
	TaskTypeShell   TaskType = "shell"   // Shell command task.
)

// JobScope identifies where one scheduled job is allowed to execute.
type JobScope string

// Supported scheduled-job scopes.
const (
	JobScopeMasterOnly JobScope = "master_only" // Execute only on the primary node.
	JobScopeAllNode    JobScope = "all_node"    // Execute on every node.
)

// JobConcurrency identifies the in-node overlap policy for one job.
type JobConcurrency string

// Supported job concurrency strategies.
const (
	JobConcurrencySingleton JobConcurrency = "singleton" // Skip overlapping executions.
	JobConcurrencyParallel  JobConcurrency = "parallel"  // Allow overlaps up to maxConcurrency.
)

// JobStatus identifies the persistent lifecycle status of one job definition.
type JobStatus string

// Supported job lifecycle statuses.
const (
	JobStatusEnabled         JobStatus = "enabled"          // The job should stay scheduled.
	JobStatusDisabled        JobStatus = "disabled"         // The job is stored but not scheduled.
	JobStatusPausedByPlugin  JobStatus = "paused_by_plugin" // The job is blocked because its handler is unavailable.
)

// TriggerType identifies how one execution was started.
type TriggerType string

// Supported execution trigger types.
const (
	TriggerTypeCron   TriggerType = "cron"   // The execution was started by cron.
	TriggerTypeManual TriggerType = "manual" // The execution was started manually.
)

// LogStatus identifies the recorded execution outcome.
type LogStatus string

// Supported job-log statuses.
const (
	LogStatusRunning               LogStatus = "running"                  // The execution is still in progress.
	LogStatusSuccess               LogStatus = "success"                  // The execution completed successfully.
	LogStatusFailed                LogStatus = "failed"                   // The execution completed with a non-timeout error.
	LogStatusCancelled             LogStatus = "cancelled"                // The execution was cancelled manually.
	LogStatusTimeout               LogStatus = "timeout"                  // The execution exceeded the configured timeout.
	LogStatusSkippedNotPrimary     LogStatus = "skipped_not_primary"      // The current node is not primary for a master-only job.
	LogStatusSkippedSingleton      LogStatus = "skipped_singleton"        // The singleton guard blocked the execution.
	LogStatusSkippedMaxConcurrency LogStatus = "skipped_max_concurrency"  // The parallelism guard blocked the execution.
)

// StopReason identifies why one job stopped scheduling new executions.
type StopReason string

// Supported stop reasons persisted on sys_job.stop_reason.
const (
	StopReasonManual               StopReason = "manual"                  // The operator disabled the job manually.
	StopReasonPluginUnavailable    StopReason = "plugin_unavailable"      // The handler registry no longer exposes the target handler.
	StopReasonMaxExecutionsReached StopReason = "max_executions_reached"  // The job exhausted its execution quota.
)

// HandlerSource identifies the registry owner of one handler definition.
type HandlerSource string

// Supported handler source types.
const (
	HandlerSourceHost   HandlerSource = "host"   // Host-provided handler.
	HandlerSourcePlugin HandlerSource = "plugin" // Plugin-provided handler.
)

// RetentionMode identifies one log-retention policy mode.
type RetentionMode string

// Supported log-retention policy modes.
const (
	RetentionModeDays  RetentionMode = "days"  // Remove logs older than N days.
	RetentionModeCount RetentionMode = "count" // Keep only the latest N logs.
	RetentionModeNone  RetentionMode = "none"  // Do not clean matching logs automatically.
)

// RetentionOption stores one normalized log-retention policy snapshot.
type RetentionOption struct {
	Mode  RetentionMode `json:"mode"`  // Mode selects the retention strategy.
	Value int64         `json:"value"` // Value stores the positive threshold for the selected mode.
}

// ShellResult stores one shell execution result snapshot persisted to sys_job_log.result_json.
type ShellResult struct {
	Stdout    string `json:"stdout,omitempty"`    // Stdout stores the captured standard output snippet.
	Stderr    string `json:"stderr,omitempty"`    // Stderr stores the captured standard error snippet.
	ExitCode  int    `json:"exitCode"`            // ExitCode stores the process exit code when available.
	Cancelled bool   `json:"cancelled,omitempty"` // Cancelled reports whether cancellation interrupted the command.
	TimedOut  bool   `json:"timedOut,omitempty"`  // TimedOut reports whether the command exceeded its timeout.
}

// NormalizeTaskType trims and normalizes one raw task type string.
func NormalizeTaskType(value string) TaskType {
	return TaskType(strings.TrimSpace(value))
}

// NormalizeJobScope trims and normalizes one raw job scope string.
func NormalizeJobScope(value string) JobScope {
	return JobScope(strings.TrimSpace(value))
}

// NormalizeJobConcurrency trims and normalizes one raw concurrency string.
func NormalizeJobConcurrency(value string) JobConcurrency {
	return JobConcurrency(strings.TrimSpace(value))
}

// NormalizeJobStatus trims and normalizes one raw job status string.
func NormalizeJobStatus(value string) JobStatus {
	return JobStatus(strings.TrimSpace(value))
}

// NormalizeTriggerType trims and normalizes one raw trigger type string.
func NormalizeTriggerType(value string) TriggerType {
	return TriggerType(strings.TrimSpace(value))
}

// NormalizeLogStatus trims and normalizes one raw log status string.
func NormalizeLogStatus(value string) LogStatus {
	return LogStatus(strings.TrimSpace(value))
}

// NormalizeHandlerSource trims and normalizes one raw handler source string.
func NormalizeHandlerSource(value string) HandlerSource {
	return HandlerSource(strings.TrimSpace(value))
}

// NormalizeRetentionMode trims and normalizes one raw retention mode string.
func NormalizeRetentionMode(value string) RetentionMode {
	return RetentionMode(strings.TrimSpace(value))
}

// IsValid reports whether the task type is supported.
func (t TaskType) IsValid() bool {
	switch t {
	case TaskTypeHandler, TaskTypeShell:
		return true
	}
	return false
}

// IsValid reports whether the job scope is supported.
func (s JobScope) IsValid() bool {
	switch s {
	case JobScopeMasterOnly, JobScopeAllNode:
		return true
	}
	return false
}

// IsValid reports whether the job concurrency mode is supported.
func (c JobConcurrency) IsValid() bool {
	switch c {
	case JobConcurrencySingleton, JobConcurrencyParallel:
		return true
	}
	return false
}

// IsValid reports whether the job status is supported.
func (s JobStatus) IsValid() bool {
	switch s {
	case JobStatusEnabled, JobStatusDisabled, JobStatusPausedByPlugin:
		return true
	}
	return false
}

// IsValid reports whether the trigger type is supported.
func (t TriggerType) IsValid() bool {
	switch t {
	case TriggerTypeCron, TriggerTypeManual:
		return true
	}
	return false
}

// IsValid reports whether the log status is supported.
func (s LogStatus) IsValid() bool {
	switch s {
	case LogStatusRunning,
		LogStatusSuccess,
		LogStatusFailed,
		LogStatusCancelled,
		LogStatusTimeout,
		LogStatusSkippedNotPrimary,
		LogStatusSkippedSingleton,
		LogStatusSkippedMaxConcurrency:
		return true
	}
	return false
}

// IsValid reports whether the handler source is supported.
func (s HandlerSource) IsValid() bool {
	switch s {
	case HandlerSourceHost, HandlerSourcePlugin:
		return true
	}
	return false
}

// IsValid reports whether the retention mode is supported.
func (m RetentionMode) IsValid() bool {
	switch m {
	case RetentionModeDays, RetentionModeCount, RetentionModeNone:
		return true
	}
	return false
}

// ParseRetentionOption parses one JSON retention payload and validates its range.
func ParseRetentionOption(raw string) (*RetentionOption, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	var option RetentionOption
	if err := json.Unmarshal([]byte(trimmed), &option); err != nil {
		return nil, gerror.Wrap(err, "解析日志保留策略失败")
	}
	option.Mode = NormalizeRetentionMode(string(option.Mode))
	if !option.Mode.IsValid() {
		return nil, gerror.New("日志保留策略模式不受支持")
	}
	if option.Mode == RetentionModeNone {
		option.Value = 0
		return &option, nil
	}
	if option.Value <= 0 {
		return nil, gerror.New("日志保留策略阈值必须大于0")
	}
	return &option, nil
}

// MustMarshalRetentionOption serializes one optional retention payload for persistence.
func MustMarshalRetentionOption(option *RetentionOption) (string, error) {
	if option == nil {
		return "", nil
	}
	data, err := json.Marshal(option)
	if err != nil {
		return "", gerror.Wrap(err, "序列化日志保留策略失败")
	}
	return string(data), nil
}
