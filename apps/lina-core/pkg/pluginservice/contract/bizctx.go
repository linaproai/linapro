// This file defines the source-plugin visible business-context contract.

package contract

import "context"

// BizCtxService defines the business-context operations published to source plugins.
type BizCtxService interface {
	// Current returns a read-only snapshot of the request context fields.
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

// ContextProvider defines an optional source for plugin-visible business context.
type ContextProvider interface {
	// Current returns a read-only snapshot of the request context fields.
	Current(ctx context.Context) CurrentContext
}

type currentContextKey struct{}

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

// CurrentFromContext returns a plugin-visible snapshot injected with WithCurrentContext.
func CurrentFromContext(ctx context.Context) CurrentContext {
	if ctx == nil {
		return CurrentContext{}
	}
	if current, ok := ctx.Value(currentContextKey{}).(CurrentContext); ok {
		return current
	}
	return CurrentContext{}
}
