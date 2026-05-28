// This file defines the source-plugin visible authentication contract.

package contract

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// AuthService defines tenant token operations published to source plugins.
type AuthService interface {
	// SelectTenant consumes a pre-login token and issues a tenant-bound token.
	SelectTenant(ctx context.Context, in SelectTenantInput) (*TenantTokenOutput, error)
	// SwitchTenant validates membership, revokes the current token, and issues a new token.
	SwitchTenant(ctx context.Context, in SwitchTenantInput) (*TenantTokenOutput, error)
	// IssueImpersonationToken asks the host auth service to sign and register a
	// tenant impersonation access token for the current platform administrator.
	// Plugins remain responsible for their own business authorization and audit
	// checks before calling this method; the host owns JWT signing, token IDs,
	// online-session state, and permission-cache priming.
	IssueImpersonationToken(ctx context.Context, in ImpersonationTokenIssueInput) (*ImpersonationTokenOutput, error)
	// RevokeImpersonationToken validates that bearerToken is an impersonation
	// access token for the optional tenant boundary and revokes the host session.
	RevokeImpersonationToken(ctx context.Context, in ImpersonationTokenRevokeInput) error
	// LoginByExternal hands off a verified external provider identity to the
	// host login flow and returns either a token pair (single-tenant users)
	// or a pre-login handoff token plus tenant candidates (multi-tenant
	// users). Source-plugin OAuth callbacks must validate the provider
	// exchange before calling this method; the host enforces login policy,
	// user lookup by email, tenant resolution, JWT issuance, online session
	// creation, and auth lifecycle hooks identically to password login.
	LoginByExternal(ctx context.Context, in ExternalLoginInput) (*ExternalLoginOutput, error)
	// OAuthHandoffURL returns the full URL path that source-plugin OAuth
	// callbacks should redirect the browser to so the workbench can consume
	// the host login outcome. The returned value already accounts for the
	// admin workspace base path and the configured vue-router history
	// backend (hash or history). Callers supply the URL-decoded query
	// parameters (tokens, pre-login handoff, error code, redirect target)
	// as a map; the helper percent-encodes them and appends them to the URL
	// using the right delimiter for the active router backend so callers do
	// not need to know whether the workspace uses hash or history routing.
	OAuthHandoffURL(ctx context.Context, queryParams map[string]string) string
}

// ExternalLoginInput describes a verified external provider identity passed
// from a source-plugin OAuth callback to the host login handoff.
type ExternalLoginInput struct {
	// ProviderID is the stable provider identifier (e.g. "google").
	ProviderID string
	// PluginID is the owning source-plugin identifier (e.g. "linapro-oidc-google").
	PluginID string
	// ExternalUserID is the provider-side user identifier, used for logging.
	ExternalUserID string
	// Email is the email address returned by the provider; required.
	Email string
	// DisplayName is the human-readable name returned by the provider; optional.
	DisplayName string
	// ClientIP is the original client IP forwarded by the OAuth callback.
	// When empty the host falls back to the current request IP.
	ClientIP string
}

// ExternalLoginTenant describes one tenant candidate returned to source
// plugins when the resolved user is eligible for multiple tenants.
type ExternalLoginTenant struct {
	// ID is the tenant database identifier.
	ID int
	// Code is the tenant code shown in tenant selection UI.
	Code string
	// Name is the tenant display name shown in tenant selection UI.
	Name string
	// Status is the tenant lifecycle status string published by the host.
	Status string
}

// ExternalLoginOutput is the host login outcome returned to source-plugin
// OAuth callbacks. Exactly one of AccessToken/RefreshToken or PreToken/Tenants
// is populated, mirroring the password login outcome shape.
type ExternalLoginOutput struct {
	// AccessToken is the issued host JWT access token; populated for
	// single-tenant users.
	AccessToken string
	// RefreshToken is the issued host JWT refresh token; populated together
	// with AccessToken.
	RefreshToken string
	// PreToken is the short-lived pre-login handoff token used by frontend
	// tenant selection; populated for multi-tenant users.
	PreToken string
	// Tenants are the tenant candidates eligible for selection; populated
	// together with PreToken.
	Tenants []ExternalLoginTenant
}

// SelectTenantInput defines input for a pre-token tenant selection.
type SelectTenantInput struct {
	// PreToken is the short-lived pre-login token produced by host login.
	PreToken string
	// TenantID is the requested target tenant.
	TenantID int
}

// SwitchTenantInput defines input for authenticated tenant switching.
type SwitchTenantInput struct {
	// BearerToken is the current Authorization bearer token.
	BearerToken string
	// TenantID is the requested target tenant.
	TenantID int
}

// TenantTokenOutput contains one newly signed tenant-bound access token.
type TenantTokenOutput struct {
	// AccessToken is the host-compatible JWT.
	AccessToken string
	// RefreshToken is the host-compatible refresh JWT for the same session.
	RefreshToken string
}

// ImpersonationTokenIssueInput defines input for host-owned impersonation
// token issuance.
type ImpersonationTokenIssueInput struct {
	// ActingUserID is the platform administrator user ID that owns the token.
	ActingUserID int
	// TenantID is the target tenant that the administrator enters.
	TenantID int
}

// ImpersonationTokenRevokeInput defines input for host-owned impersonation
// token revocation.
type ImpersonationTokenRevokeInput struct {
	// BearerToken is the impersonation access token or Authorization header value.
	BearerToken string
	// TenantID is the expected tenant. A zero value skips tenant matching.
	TenantID int
}

// ImpersonationTokenOutput contains one host-signed impersonation access token
// and the session metadata registered with host auth state.
type ImpersonationTokenOutput struct {
	// AccessToken is the host-compatible impersonation JWT.
	AccessToken string
	// TokenID is the generated token/session identifier.
	TokenID string
	// TenantID is the target tenant stored in token claims and session state.
	TenantID int
	// ActingUserID is the platform administrator user ID stored in token claims.
	ActingUserID int
}

// BearerTokenFromContext extracts the bearer token from the current HTTP request.
func BearerTokenFromContext(ctx context.Context) (string, bool) {
	request := g.RequestFromCtx(ctx)
	if request == nil {
		return "", false
	}
	header := request.GetHeader("Authorization")
	token := strings.TrimPrefix(header, "Bearer ")
	return token, token != "" && token != header
}
