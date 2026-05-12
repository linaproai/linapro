// Package bizctx exposes a narrowed view of the host business context to source
// plugins so they can read current request identity, tenancy, and impersonation
// metadata without depending on host-internal service packages.
package bizctx

import (
	"context"

	"lina-core/internal/model"
	internalbizctx "lina-core/internal/service/bizctx"
)

// Service defines the bizctx operations published to source plugins.
type Service interface {
	// Current returns a read-only snapshot of the request context fields published
	// to source plugins.
	Current(ctx context.Context) CurrentContext
}

// CurrentContext is the plugin-visible read-only business context snapshot.
type CurrentContext struct {
	// UserID is the authenticated user identifier bound to the request context.
	UserID int
	// Username is the authenticated username bound to the request context.
	Username string
	// TenantID is the tenant identifier bound to the request context.
	TenantID int
	// ActingUserID is the real platform user ID during impersonation.
	ActingUserID int
	// ActingAsTenant reports whether the request acts through a tenant view.
	ActingAsTenant bool
	// IsImpersonation reports whether the current token represents impersonation.
	IsImpersonation bool
	// PlatformBypass reports whether the request runs in platform scope.
	PlatformBypass bool
}

type currentContextKey struct{}

// serviceAdapter bridges the internal bizctx service into the published plugin contract.
type serviceAdapter struct {
	service internalbizctx.Service
}

// WithCurrentContext returns a child context carrying a plugin-visible business
// context snapshot without exposing host-internal context model types.
func WithCurrentContext(ctx context.Context, current CurrentContext) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if current.TenantID == 0 {
		current.PlatformBypass = true
	}
	return context.WithValue(ctx, currentContextKey{}, current)
}

// New creates and returns the published bizctx service adapter.
func New() Service {
	return &serviceAdapter{service: internalbizctx.New()}
}

// Current returns a read-only snapshot of the request context fields published
// to source plugins.
func (s *serviceAdapter) Current(ctx context.Context) CurrentContext {
	if s != nil && s.service != nil && ctx != nil {
		if c := s.service.Get(ctx); c != nil {
			return currentContextFromModel(c)
		}
	}
	if ctx == nil {
		return CurrentContext{}
	}
	if current, ok := ctx.Value(currentContextKey{}).(CurrentContext); ok {
		return current
	}
	if c, ok := ctx.Value(internalbizctx.ContextKey).(*model.Context); ok {
		return currentContextFromModel(c)
	}
	return CurrentContext{}
}

// currentContextFromModel converts the host context model to the published
// plugin read-only snapshot without exposing the mutable internal pointer.
func currentContextFromModel(c *model.Context) CurrentContext {
	if c == nil {
		return CurrentContext{}
	}
	return CurrentContext{
		UserID:          c.UserId,
		Username:        c.Username,
		TenantID:        c.TenantId,
		ActingUserID:    c.ActingUserId,
		ActingAsTenant:  c.ActingAsTenant,
		IsImpersonation: c.IsImpersonation,
		PlatformBypass:  c.TenantId == 0,
	}
}
