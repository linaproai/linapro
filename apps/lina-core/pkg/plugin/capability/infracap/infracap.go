// Package infracap defines infrastructure capability projections for plugins.
// It exposes host-owned status and primitive metadata without leaking concrete
// runtime backends.
package infracap

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"
)

// ComponentID identifies one infrastructure component.
type ComponentID string

// StatusProjection describes one infrastructure component state.
type StatusProjection struct {
	// ID is the component identifier.
	ID ComponentID
	// Available reports whether the component can serve requests.
	Available bool
	// Status is the component-owned status value.
	Status string
	// LabelKey is the stable runtime i18n label key.
	LabelKey string
	// Label is the optional locale-resolved label.
	Label string
}

// Service defines read-oriented infrastructure capability methods.
type Service interface {
	// BatchGetStatus returns visible component status projections.
	BatchGetStatus(ctx context.Context, capCtx capmodel.CapabilityContext, ids []ComponentID) (*capmodel.BatchResult[*StatusProjection, ComponentID], error)
}

// AdminService defines governed infrastructure management commands.
type AdminService interface {
	// RefreshStatus refreshes one component status snapshot.
	RefreshStatus(ctx context.Context, capCtx capmodel.CapabilityContext, id ComponentID) error
}

// ScopeService defines host-internal infrastructure visibility helpers.
type ScopeService interface {
	// EnsureComponentsVisible rejects when any component is outside caller scope.
	EnsureComponentsVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []ComponentID) error
}
