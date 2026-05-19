// Package v1 defines shared scheduled-job API DTOs and compact enum contracts.
package v1

import (
	"lina-core/pkg/statusflag"
)

// TaskType identifies the scheduled-job execution mode.
type TaskType string

// Supported scheduled-job execution modes.
const (
	TaskTypeHandler TaskType = "handler"
	TaskTypeShell   TaskType = "shell"
)

// Scope identifies where one scheduled job is allowed to execute.
type Scope string

// Supported scheduled-job scopes.
const (
	ScopeMasterOnly Scope = "master_only"
	ScopeAllNode    Scope = "all_node"
)

// Concurrency identifies the in-node overlap policy for one job.
type Concurrency string

// Supported job concurrency strategies.
const (
	ConcurrencySingleton Concurrency = "singleton"
	ConcurrencyParallel  Concurrency = "parallel"
)

// Status identifies the persistent lifecycle status of one job definition.
type Status string

// Supported job lifecycle statuses.
const (
	StatusEnabled        Status = "enabled"
	StatusDisabled       Status = "disabled"
	StatusPausedByPlugin Status = "paused_by_plugin"
)

// RetentionMode identifies one job-log retention policy mode.
type RetentionMode string

// Supported job-log retention policy modes.
const (
	RetentionModeDays  RetentionMode = "days"
	RetentionModeCount RetentionMode = "count"
	RetentionModeNone  RetentionMode = "none"
)

// JobItem exposes scheduled-job fields needed by the management UI.
type JobItem struct {
	Id                   int64            `json:"id" dc:"Job ID" eg:"1"`
	GroupId              int64            `json:"groupId" dc:"Owning group ID" eg:"1"`
	Name                 string           `json:"name" dc:"Job name" eg:"Log cleanup"`
	Description          string           `json:"description" dc:"Job description" eg:"Clean expired logs"`
	TaskType             TaskType         `json:"taskType" dc:"Job type: handler or shell" eg:"handler"`
	HandlerRef           string           `json:"handlerRef" dc:"Handler reference" eg:"host:cleanup-logs"`
	Params               string           `json:"params" dc:"Handler parameters JSON" eg:"{}"`
	TimeoutSeconds       int              `json:"timeoutSeconds" dc:"Execution timeout in seconds" eg:"30"`
	ShellCmd             string           `json:"shellCmd" dc:"Shell script content" eg:"echo hello"`
	WorkDir              string           `json:"workDir" dc:"Shell working directory" eg:"/tmp"`
	Env                  string           `json:"env" dc:"Shell environment JSON" eg:"{}"`
	CronExpr             string           `json:"cronExpr" dc:"Cron expression" eg:"0 0 * * * *"`
	Timezone             string           `json:"timezone" dc:"Timezone identifier" eg:"Asia/Shanghai"`
	Scope                Scope            `json:"scope" dc:"Scheduling scope" eg:"master_only"`
	Concurrency          Concurrency      `json:"concurrency" dc:"Concurrency policy" eg:"singleton"`
	MaxConcurrency       int              `json:"maxConcurrency" dc:"Maximum concurrency" eg:"1"`
	MaxExecutions        int              `json:"maxExecutions" dc:"Maximum executions, 0 means unlimited" eg:"0"`
	ExecutedCount        int64            `json:"executedCount" dc:"Executed count" eg:"12"`
	StopReason           string           `json:"stopReason" dc:"Stop reason" eg:""`
	LogRetentionOverride string           `json:"logRetentionOverride" dc:"Log retention override JSON" eg:"{\"mode\":\"days\",\"value\":30}"`
	Status               Status           `json:"status" dc:"Job status" eg:"enabled"`
	IsBuiltin            statusflag.YesNo `json:"isBuiltin" dc:"Built-in job flag: 1=yes 0=no" eg:"0"`
	SeedVersion          int              `json:"seedVersion" dc:"Built-in seed version" eg:"1"`
	CreatedBy            int64            `json:"createdBy" dc:"Creator user ID" eg:"1"`
	UpdatedBy            int64            `json:"updatedBy" dc:"Updater user ID" eg:"1"`
	CreatedAt            *int64           `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1778733600000"`
	UpdatedAt            *int64           `json:"updatedAt" dc:"Update time as Unix timestamp in milliseconds" eg:"1778733600000"`
}
