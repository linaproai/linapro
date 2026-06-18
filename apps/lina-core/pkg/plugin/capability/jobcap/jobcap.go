// Package jobcap defines scheduled-job capability contracts for plugins without
// exposing host job, cron, or task-log storage.
package jobcap

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"
)

const (
	// MaxSearchPageSize limits one scheduled-job candidate search page.
	MaxSearchPageSize = 200
	// MaxEnsureVisible limits one scheduled-job visibility check.
	MaxEnsureVisible = 200
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

// SearchInput constrains scheduled-job candidate search.
type SearchInput struct {
	// Keyword filters by job name or handler reference.
	Keyword string
	// Group filters by job group identifier.
	Group string
	// Status filters by job lifecycle status.
	Status string
	// Page constrains result size and stable sorting.
	Page capmodel.PageRequest
}

// Service defines read-oriented job capability methods.
type Service interface {
	// BatchGet returns visible job projections and opaque missing IDs.
	BatchGet(ctx context.Context, capCtx capmodel.CapabilityContext, ids []JobID) (*capmodel.BatchResult[*Projection, JobID], error)
	// Search returns one bounded page of visible scheduled-job projections.
	Search(ctx context.Context, capCtx capmodel.CapabilityContext, input SearchInput) (*capmodel.PageResult[*Projection], error)
	// EnsureVisible rejects when any requested job is absent or invisible.
	EnsureVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []JobID) error
}

// AdminService defines governed job execution commands.
type AdminService interface {
	// Run triggers one visible job after state and target checks.
	Run(ctx context.Context, capCtx capmodel.CapabilityContext, id JobID) error
	// SetStatus changes one job lifecycle status.
	SetStatus(ctx context.Context, capCtx capmodel.CapabilityContext, id JobID, status string) error
}

// ScopeService defines host-internal job visibility helpers.
type ScopeService interface {
	// EnsureVisible rejects when any job is outside caller scope.
	EnsureVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []JobID) error
}
