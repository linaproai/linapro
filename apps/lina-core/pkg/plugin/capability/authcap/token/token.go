// Package token defines authentication token contracts and handoff operations
// published through authcap without exposing host JWT signing, session storage,
// or tenant membership internals to plugins.
package token

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// Kind names the intended use of one signed JWT carried in the tokenType claim.
// Use these constants instead of hard-coding claim string literals.
const (
	// KindAccess marks JWTs accepted by protected API middleware and dynamic
	// plugin route dispatch.
	KindAccess = "access"
	// KindRefresh marks JWTs accepted only by the refresh-token endpoint.
	KindRefresh = "refresh"
)

// ClientType identifies the user-facing client that created an authenticated
// user session. It deliberately excludes service, plugin, and machine actors.
type ClientType string

const (
	// ClientTypeWeb represents browser Web apps and the default LinaPro workbench.
	ClientTypeWeb ClientType = "web"
	// ClientTypeMobile represents mobile app user sessions.
	ClientTypeMobile ClientType = "mobile"
	// ClientTypeDesktop represents desktop client user sessions.
	ClientTypeDesktop ClientType = "desktop"
	// ClientTypeCLI represents user sessions created through command-line clients.
	ClientTypeCLI ClientType = "cli"
)

// ParseClientType validates one user-session client type value.
func ParseClientType(value string) (ClientType, bool) {
	switch clientType := ClientType(strings.TrimSpace(value)); clientType {
	case ClientTypeWeb, ClientTypeMobile, ClientTypeDesktop, ClientTypeCLI:
		return clientType, true
	default:
		return "", false
	}
}

// String returns the wire value for one client type.
func (t ClientType) String() string {
	return string(t)
}

// Service defines tenant token operations published to source plugins.
type Service interface {
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
