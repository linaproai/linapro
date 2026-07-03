// This file defines the source-plugin visible business-context contract.

package bizctxcap

import "context"

// Service defines the business-context operations published to source plugins.
type Service interface {
	// Current returns a read-only snapshot of the request context fields.
	Current(ctx context.Context) CurrentContext
}

// CurrentContext is the plugin-visible read-only business context snapshot.
type CurrentContext struct {
	// TokenID is the authenticated token or online-session identifier bound to the request context.
	TokenID string
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
	// Permissions contains effective permission keys for the current request.
	Permissions []string
	// DataScope stores the effective role data-scope snapshot.
	DataScope int
	// DataScopeUnsupported reports whether the role snapshot contains an unsupported data scope.
	DataScopeUnsupported bool
	// UnsupportedDataScope stores the first unsupported data-scope value.
	UnsupportedDataScope int
	// IsSuperAdmin reports whether the caller bypasses normal permission checks.
	IsSuperAdmin bool
	// PlatformBypass reports whether the request runs in platform scope.
	PlatformBypass bool
}

// currentContextKey is the private key for plugin-visible context snapshots.
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
	current.Permissions = cloneStrings(current.Permissions)
	return context.WithValue(ctx, currentContextKey{}, current)
}

// CurrentFromContext returns a plugin-visible snapshot injected with WithCurrentContext.
func CurrentFromContext(ctx context.Context) CurrentContext {
	if ctx == nil {
		return CurrentContext{}
	}
	if current, ok := ctx.Value(currentContextKey{}).(CurrentContext); ok {
		current.Permissions = cloneStrings(current.Permissions)
		return current
	}
	return CurrentContext{}
}

// cloneStrings returns a detached copy of string values stored in context.
func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}
