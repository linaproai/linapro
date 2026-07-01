// Package jobmeta defines shared scheduled-job domain types, input value
// objects, and JSON payload helpers used by job management, scheduling, and
// handler execution.
package jobmeta

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	jobv1 "lina-core/api/job/v1"
	jobhandlerv1 "lina-core/api/jobhandler/v1"
	joblogv1 "lina-core/api/joblog/v1"
	"lina-core/pkg/bizerr"
)

// TaskType identifies the scheduled-job execution mode.
type TaskType = jobv1.TaskType

// JobScope identifies where one scheduled job is allowed to execute.
type JobScope = jobv1.Scope

// JobConcurrency identifies the in-node overlap policy for one job.
type JobConcurrency = jobv1.Concurrency

// TriggerType identifies how one execution was started.
type TriggerType = joblogv1.Trigger

// LogStatus identifies the recorded execution outcome.
type LogStatus = joblogv1.Status

// StopReason identifies why one job stopped scheduling new executions.
type StopReason string

// Supported stop reasons persisted on sys_job.stop_reason.
const (
	StopReasonManual               StopReason = "manual"                 // The operator disabled the job manually.
	StopReasonPluginUnavailable    StopReason = "plugin_unavailable"     // The handler registry no longer exposes the target handler.
	StopReasonMaxExecutionsReached StopReason = "max_executions_reached" // The job exhausted its execution quota.
)

// HandlerSource identifies the registry owner of one handler definition.
type HandlerSource = jobhandlerv1.Source

const (
	// HandlerI18nKeyPrefix is the runtime i18n prefix for scheduled-job
	// handler and built-in job display metadata.
	HandlerI18nKeyPrefix = "job.handler"
)

// RetentionMode identifies one log-retention policy mode.
type RetentionMode = jobv1.RetentionMode

// RetentionOption stores one normalized log-retention policy snapshot.
type RetentionOption struct {
	Mode  RetentionMode `json:"mode"`  // Mode selects the retention strategy.
	Value int64         `json:"value"` // Value stores the positive threshold for the selected mode.
}

// SaveJobInput stores mutable scheduled-job fields shared by job management
// and host capability adapters.
type SaveJobInput struct {
	GroupID              int64             // GroupID identifies the owning group.
	Name                 string            // Name is unique within the group.
	Description          string            // Description explains the job purpose.
	TaskType             TaskType          // TaskType selects handler or shell execution.
	HandlerRef           string            // HandlerRef selects the registered handler for handler jobs.
	Params               map[string]any    // Params stores handler parameters.
	Timeout              time.Duration     // Timeout bounds each execution.
	ShellCmd             string            // ShellCmd stores the shell script for shell jobs.
	WorkDir              string            // WorkDir stores the optional shell working directory.
	Env                  map[string]string // Env stores shell environment overrides.
	CronExpr             string            // CronExpr stores the cron expression.
	Timezone             string            // Timezone stores the cron timezone identifier.
	Scope                JobScope          // Scope selects master-only or all-node execution.
	Concurrency          JobConcurrency    // Concurrency selects singleton or parallel execution.
	MaxConcurrency       int               // MaxConcurrency caps parallel overlap per node.
	MaxExecutions        int               // MaxExecutions caps cron-triggered runs.
	Status               jobv1.Status      // Status selects enabled or disabled persistence state.
	LogRetentionOverride *RetentionOption  // LogRetentionOverride stores the optional per-job policy.
}

// UpdateJobInput stores one job update request shared by job management and
// host capability adapters.
type UpdateJobInput struct {
	ID int64 // ID identifies the target job.
	SaveJobInput
}

// Owner defines governed scheduled-job mutation and execution operations
// implemented by jobmgmt.Service and consumed by host capability adapters.
type Owner interface {
	// CreateJob persists one governed scheduled job and updates scheduler state.
	CreateJob(ctx context.Context, in SaveJobInput) (int64, error)
	// UpdateJob mutates one governed scheduled job and updates scheduler state.
	UpdateJob(ctx context.Context, in UpdateJobInput) error
	// DeleteJobs removes governed scheduled jobs and unregisters scheduler entries.
	DeleteJobs(ctx context.Context, ids string) error
	// UpdateJobStatus toggles one scheduled job and updates scheduler state.
	UpdateJobStatus(ctx context.Context, id int64, status jobv1.Status) error
	// TriggerJob starts one manual execution through the scheduler.
	TriggerJob(ctx context.Context, id int64) (int64, error)
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
	return jobv1.TaskType(strings.TrimSpace(value))
}

// NormalizeJobScope trims and normalizes one raw job scope string.
func NormalizeJobScope(value string) JobScope {
	return jobv1.Scope(strings.TrimSpace(value))
}

// NormalizeJobConcurrency trims and normalizes one raw concurrency string.
func NormalizeJobConcurrency(value string) JobConcurrency {
	return jobv1.Concurrency(strings.TrimSpace(value))
}

// NormalizeJobStatus trims and normalizes one raw job status string.
func NormalizeJobStatus(value string) jobv1.Status {
	return jobv1.Status(strings.TrimSpace(value))
}

// NormalizeTriggerType trims and normalizes one raw trigger type string.
func NormalizeTriggerType(value string) TriggerType {
	return joblogv1.Trigger(strings.TrimSpace(value))
}

// NormalizeLogStatus trims and normalizes one raw log status string.
func NormalizeLogStatus(value string) LogStatus {
	return joblogv1.Status(strings.TrimSpace(value))
}

// NormalizeHandlerSource trims and normalizes one raw handler source string.
func NormalizeHandlerSource(value string) HandlerSource {
	return jobhandlerv1.Source(strings.TrimSpace(value))
}

// NormalizeRetentionMode trims and normalizes one raw retention mode string.
func NormalizeRetentionMode(value string) RetentionMode {
	return jobv1.RetentionMode(strings.TrimSpace(value))
}

// HandlerI18nKey builds one stable runtime i18n key for handler-owned display
// metadata, using the handler ref as the backend anchor.
func HandlerI18nKey(ref string, field string) string {
	refPath := messageKeyPath(ref)
	fieldPath := messageKeyPath(field)
	if refPath == "" || fieldPath == "" {
		return ""
	}
	return HandlerI18nKeyPrefix + "." + refPath + "." + fieldPath
}

// ParseRetentionOption parses one JSON retention payload and validates its range.
func ParseRetentionOption(raw string) (*RetentionOption, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	var option RetentionOption
	if err := json.Unmarshal([]byte(trimmed), &option); err != nil {
		return nil, bizerr.WrapCode(err, CodeJobRetentionParseFailed)
	}
	option.Mode = NormalizeRetentionMode(string(option.Mode))
	if !option.Mode.IsValid() {
		return nil, bizerr.NewCode(CodeJobRetentionModeUnsupported)
	}
	if option.Mode == jobv1.RetentionModeNone {
		option.Value = 0
		return &option, nil
	}
	if option.Value <= 0 {
		return nil, bizerr.NewCode(CodeJobRetentionValueInvalid)
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
		return "", bizerr.WrapCode(err, CodeJobRetentionMarshalFailed)
	}
	return string(data), nil
}
