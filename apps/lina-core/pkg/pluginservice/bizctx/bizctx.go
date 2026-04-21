// Package bizctx exposes a narrowed view of the host business context to source
// plugins so they can read the current request's authenticated user without
// depending on host-internal service packages.
package bizctx

import (
	"context"

	internalbizctx "lina-core/internal/service/bizctx"
)

// Service defines the bizctx operations published to source plugins.
type Service interface {
	// CurrentUserID returns the authenticated user identifier bound to the request
	// context, or zero when no user context is attached.
	CurrentUserID(ctx context.Context) int
	// CurrentUsername returns the authenticated username bound to the request
	// context, or the empty string when no user context is attached.
	CurrentUsername(ctx context.Context) string
}

// serviceAdapter bridges the internal bizctx service into the published plugin contract.
type serviceAdapter struct {
	service internalbizctx.Service
}

// New creates and returns the published bizctx service adapter.
func New() Service {
	return &serviceAdapter{service: internalbizctx.New()}
}

// CurrentUserID returns the authenticated user identifier bound to the request context.
func (s *serviceAdapter) CurrentUserID(ctx context.Context) int {
	if s == nil || s.service == nil {
		return 0
	}
	if c := s.service.Get(ctx); c != nil {
		return c.UserId
	}
	return 0
}

// CurrentUsername returns the authenticated username bound to the request context.
func (s *serviceAdapter) CurrentUsername(ctx context.Context) string {
	if s == nil || s.service == nil {
		return ""
	}
	if c := s.service.Get(ctx); c != nil {
		return c.Username
	}
	return ""
}
