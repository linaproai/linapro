// Package auth exposes a narrowed tenant authentication contract to source
// plugins so multi-tenant login policy can stay plugin-owned while JWT signing,
// revocation, and online-session persistence remain host-owned.
package auth

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	internalauth "lina-core/internal/service/auth"
	"lina-core/pkg/bizerr"
)

// Service defines tenant token operations published to source plugins.
type Service interface {
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

// serviceAdapter bridges the internal auth service into the published plugin contract.
type serviceAdapter struct {
	tokenIssuer internalauth.TenantTokenIssuer
}

// New creates and returns the published auth service adapter.
func New() Service {
	return &serviceAdapter{tokenIssuer: internalauth.Instance().(internalauth.TenantTokenIssuer)}
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

// SelectTenant consumes a pre-login token and issues a tenant-bound token.
func (s *serviceAdapter) SelectTenant(ctx context.Context, in SelectTenantInput) (*TenantTokenOutput, error) {
	if s == nil || s.tokenIssuer == nil {
		return nil, bizerr.NewCode(internalauth.CodeAuthTokenStateUnavailable)
	}
	out, err := s.tokenIssuer.IssueTenantToken(ctx, internalauth.TenantTokenIssueInput{
		PreToken: in.PreToken,
		TenantID: in.TenantID,
	})
	if err != nil {
		return nil, err
	}
	return &TenantTokenOutput{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken}, nil
}

// SwitchTenant validates membership, revokes the current token, and issues a new token.
func (s *serviceAdapter) SwitchTenant(ctx context.Context, in SwitchTenantInput) (*TenantTokenOutput, error) {
	if s == nil || s.tokenIssuer == nil {
		return nil, bizerr.NewCode(internalauth.CodeAuthTokenStateUnavailable)
	}
	if strings.TrimSpace(in.BearerToken) == "" {
		return nil, bizerr.NewCode(internalauth.CodeAuthTokenInvalid)
	}
	out, err := s.tokenIssuer.ReissueTenantTokenFromBearer(ctx, in.BearerToken, in.TenantID)
	if err != nil {
		return nil, err
	}
	return &TenantTokenOutput{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken}, nil
}
