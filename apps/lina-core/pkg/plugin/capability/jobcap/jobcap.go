// Package jobcap defines scheduled-job capability contracts for plugins without
// exposing host job, cron, or task-log storage.
package jobcap

import (
	"context"
	"time"

	jobv1 "lina-core/api/job/v1"
	"lina-core/pkg/plugin/capability/capmodel"
)

// Service defines governed scheduled-job capability methods. Reads use bounded
// info queries and tenant/data-scope filtering; execution and status
// changes validate target visibility, scheduler state, audit source, and
// scheduler side effects. Dynamic built-in job registration is declaration-time
// only and is intentionally not part of this runtime service.
type Service interface {
	// Get returns one visible scheduled-job info record.
	Get(ctx context.Context, id JobID) (*JobInfo, error)
	// BatchGet returns visible job info records and opaque missing IDs.
	BatchGet(ctx context.Context, ids []JobID) (*capmodel.BatchResult[*JobInfo, JobID], error)
	// List returns one bounded page of visible scheduled-job info records.
	List(ctx context.Context, input ListInput) (*capmodel.PageResult[*JobInfo], error)
	// EnsureVisible rejects when any requested job is absent or invisible.
	EnsureVisible(ctx context.Context, ids []JobID) error
	// Create creates one governed scheduled job through the job owner.
	Create(ctx context.Context, input SaveInput) (JobID, error)
	// Update mutates one visible scheduled job through the job owner.
	Update(ctx context.Context, input UpdateInput) error
	// Delete deletes one visible scheduled job through the job owner.
	Delete(ctx context.Context, id JobID) error
	// Run triggers one visible job after state, target, tenant/data-scope, audit,
	// scheduler, and execution-boundary checks.
	Run(ctx context.Context, id JobID) error
	// SetStatus changes one job lifecycle status after target, state-machine,
	// tenant/data-scope, audit, and scheduler-registration checks.
	SetStatus(ctx context.Context, id JobID, status jobv1.Status) error
}

const (
	// MaxListPageSize limits one scheduled-job candidate list page.
	MaxListPageSize = 200
	// MaxEnsureVisible limits one scheduled-job visibility check.
	MaxEnsureVisible = 200
)

// JobID identifies one governed job.
type JobID string

// JobInfo describes one job visible to a plugin.
type JobInfo struct {
	// ID is the job identifier.
	ID JobID
	// Name is the display name.
	Name string
	// Group is the job group.
	Group string
	// Status is the lifecycle status.
	Status jobv1.Status
	// LogRetentionOverride optionally overrides the system log cleanup policy.
	LogRetentionOverride *LogRetentionOption `json:"logRetentionOverride,omitempty"`
}

// ListInput constrains scheduled-job candidate listing.
type ListInput struct {
	// Keyword filters by job name or handler reference.
	Keyword string
	// Group filters by job group identifier.
	Group string
	// Status filters by job lifecycle status.
	Status jobv1.Status
	// Page constrains result size and stable sorting.
	Page capmodel.PageRequest
}

// LogRetentionOption stores one optional task-level log cleanup policy.
type LogRetentionOption struct {
	// Mode selects the retention strategy: days, count, or none.
	Mode jobv1.RetentionMode `json:"mode"`
	// Value stores the positive threshold for days/count; none uses zero.
	Value int64 `json:"value"`
}

// SaveInput describes mutable scheduled-job fields exposed to plugins.
type SaveInput struct {
	// GroupID identifies the owning job group.
	GroupID string
	// Name is the display name.
	Name string
	// Description explains the job purpose.
	Description string
	// Timeout bounds each execution; zero uses the host default.
	Timeout time.Duration
	// ShellCmd stores the shell script content for runtime-created jobs.
	ShellCmd string
	// WorkDir stores the optional shell working directory.
	WorkDir string
	// Env stores shell environment overrides.
	Env map[string]string
	// CronExpr stores the cron expression.
	CronExpr string
	// Timezone stores the cron timezone.
	Timezone string
	// Scope selects master-only or all-node execution; empty uses the host default.
	Scope jobv1.Scope
	// Concurrency selects singleton or parallel execution; empty uses the host default.
	Concurrency jobv1.Concurrency
	// MaxConcurrency caps parallel overlap per node; zero uses the host default.
	MaxConcurrency int
	// MaxExecutions caps cron-triggered runs; zero means unlimited.
	MaxExecutions int
	// Status is the lifecycle status.
	Status jobv1.Status
	// LogRetentionOverride optionally overrides the system log cleanup policy.
	LogRetentionOverride *LogRetentionOption `json:"logRetentionOverride,omitempty"`
}

// UpdateInput describes one scheduled-job update request.
type UpdateInput struct {
	// ID identifies the target job.
	ID JobID
	// SaveInput stores mutable fields.
	SaveInput
}
