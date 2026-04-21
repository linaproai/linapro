// Package loginlog exposes a narrowed host login-log persistence contract to
// source plugins.
package loginlog

import (
	"context"

	internalloginlog "lina-core/internal/service/loginlog"
)

// CreateInput aliases the host login-log creation input.
type CreateInput = internalloginlog.CreateInput

// Service defines the login-log operations published to source plugins.
type Service interface {
	// Create persists one login-log record.
	Create(ctx context.Context, in CreateInput) error
}

// serviceAdapter bridges the internal login-log service into the published plugin contract.
type serviceAdapter struct {
	service internalloginlog.Service
}

// New creates and returns the published login-log service adapter.
func New() Service {
	return &serviceAdapter{service: internalloginlog.New()}
}

// Create persists one login-log record.
func (s *serviceAdapter) Create(ctx context.Context, in CreateInput) error {
	if s == nil || s.service == nil {
		return nil
	}
	return s.service.Create(ctx, in)
}
