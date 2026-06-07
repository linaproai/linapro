// Package jobcap defines scheduled-job capability contracts for plugins without
// exposing host job, cron, or task-log storage.
package jobcap

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"
)

// JobID identifies one governed job.
type JobID string

// Projection describes one job visible to a plugin.
type Projection struct {
	// ID is the job identifier.
	ID JobID
	// Name is the display name.
	Name string
	// Group is the job group.
	Group string
	// Status is the lifecycle status.
	Status string
}

// Service defines read-oriented job capability methods.
type Service interface {
	// BatchGetJobs returns visible job projections and opaque missing IDs.
	BatchGetJobs(ctx context.Context, capCtx capmodel.CapabilityContext, ids []JobID) (*capmodel.BatchResult[*Projection, JobID], error)
}

// AdminService defines governed job execution commands.
type AdminService interface {
	// RunJob triggers one visible job after state and target checks.
	RunJob(ctx context.Context, capCtx capmodel.CapabilityContext, id JobID) error
	// SetJobStatus changes one job lifecycle status.
	SetJobStatus(ctx context.Context, capCtx capmodel.CapabilityContext, id JobID, status string) error
}

// ScopeService defines host-internal job visibility helpers.
type ScopeService interface {
	// EnsureJobsVisible rejects when any job is outside caller scope.
	EnsureJobsVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []JobID) error
}
