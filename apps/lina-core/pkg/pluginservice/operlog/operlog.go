// Package operlog exposes a narrowed host operation-log persistence contract to
// source plugins.
package operlog

import (
	"context"

	internaloperlog "lina-core/internal/service/operlog"
)

// CreateInput aliases the host operation-log creation input.
type CreateInput = internaloperlog.CreateInput

// Service defines the operation-log operations published to source plugins.
type Service interface {
	// Create persists one operation-log record.
	Create(ctx context.Context, in CreateInput) error
}

// serviceAdapter bridges the internal operation-log service into the published plugin contract.
type serviceAdapter struct {
	service internaloperlog.Service
}

// New creates and returns the published operation-log service adapter.
func New() Service {
	return &serviceAdapter{service: internaloperlog.New()}
}

// Create persists one operation-log record.
func (s *serviceAdapter) Create(ctx context.Context, in CreateInput) error {
	if s == nil || s.service == nil {
		return nil
	}
	return s.service.Create(ctx, in)
}
