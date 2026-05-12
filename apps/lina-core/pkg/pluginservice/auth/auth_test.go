// This file verifies the plugin-facing auth adapter contract.

package auth

import (
	"context"
	"testing"

	internalauth "lina-core/internal/service/auth"
)

// TestServiceAdapterUsesTenantTokenIssuer verifies plugin auth calls depend on
// the narrowed tenant token issuer rather than the full host auth service.
func TestServiceAdapterUsesTenantTokenIssuer(t *testing.T) {
	ctx := context.Background()
	issuer := &fakeTenantTokenIssuer{}
	svc := &serviceAdapter{tokenIssuer: issuer}

	selected, err := svc.SelectTenant(ctx, SelectTenantInput{PreToken: "pre-token", TenantID: 11})
	if err != nil {
		t.Fatalf("select tenant: %v", err)
	}
	if selected.AccessToken != "issued-token" || selected.RefreshToken != "issued-refresh-token" || issuer.issuedPreToken != "pre-token" || issuer.issuedTenantID != 11 {
		t.Fatalf(
			"expected issue call, token=%q refresh=%q preToken=%q tenant=%d",
			selected.AccessToken,
			selected.RefreshToken,
			issuer.issuedPreToken,
			issuer.issuedTenantID,
		)
	}

	switched, err := svc.SwitchTenant(ctx, SwitchTenantInput{BearerToken: "bearer-token", TenantID: 22})
	if err != nil {
		t.Fatalf("switch tenant: %v", err)
	}
	if switched.AccessToken != "reissued-token" || switched.RefreshToken != "reissued-refresh-token" || issuer.reissuedBearer != "bearer-token" || issuer.reissuedTenantID != 22 {
		t.Fatalf(
			"expected reissue call, token=%q refresh=%q bearer=%q tenant=%d",
			switched.AccessToken,
			switched.RefreshToken,
			issuer.reissuedBearer,
			issuer.reissuedTenantID,
		)
	}
}

// fakeTenantTokenIssuer records plugin adapter calls for contract tests.
type fakeTenantTokenIssuer struct {
	issuedPreToken   string
	issuedTenantID   int
	reissuedBearer   string
	reissuedTenantID int
}

// IssueTenantToken records one pre-login token exchange.
func (f *fakeTenantTokenIssuer) IssueTenantToken(
	_ context.Context,
	in internalauth.TenantTokenIssueInput,
) (*internalauth.TenantTokenOutput, error) {
	f.issuedPreToken = in.PreToken
	f.issuedTenantID = in.TenantID
	return &internalauth.TenantTokenOutput{AccessToken: "issued-token", RefreshToken: "issued-refresh-token"}, nil
}

// ReissueTenantToken records no state because the plugin adapter uses bearer-token handoff.
func (f *fakeTenantTokenIssuer) ReissueTenantToken(
	context.Context,
	internalauth.TenantTokenReissueInput,
) (*internalauth.TenantTokenOutput, error) {
	return &internalauth.TenantTokenOutput{AccessToken: ""}, nil
}

// ReissueTenantTokenFromBearer records one bearer-token tenant switch.
func (f *fakeTenantTokenIssuer) ReissueTenantTokenFromBearer(
	_ context.Context,
	tokenString string,
	tenantID int,
) (*internalauth.TenantTokenOutput, error) {
	f.reissuedBearer = tokenString
	f.reissuedTenantID = tenantID
	return &internalauth.TenantTokenOutput{AccessToken: "reissued-token", RefreshToken: "reissued-refresh-token"}, nil
}
