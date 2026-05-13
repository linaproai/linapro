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
