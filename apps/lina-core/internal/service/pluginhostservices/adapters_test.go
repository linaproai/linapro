// This file verifies host-service adapters keep source-plugin contracts
// independent from host-internal DTOs.

package pluginhostservices

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/os/gtime"

	internalauth "lina-core/internal/service/auth"
	internalsession "lina-core/internal/service/session"
	plugincontract "lina-core/pkg/pluginservice/contract"
)

// TestAuthAdapterUsesTenantTokenIssuer verifies plugin auth calls depend on the narrowed token issuer.
func TestAuthAdapterUsesTenantTokenIssuer(t *testing.T) {
	ctx := context.Background()
	issuer := &fakeTenantTokenIssuer{}
	svc := &authAdapter{tokenIssuer: issuer}

	selected, err := svc.SelectTenant(ctx, plugincontract.SelectTenantInput{PreToken: "pre-token", TenantID: 11})
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

	switched, err := svc.SwitchTenant(ctx, plugincontract.SwitchTenantInput{BearerToken: "bearer-token", TenantID: 22})
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

// TestToInternalSessionFilter verifies the published filter contract is converted explicitly.
func TestToInternalSessionFilter(t *testing.T) {
	if result := toInternalSessionFilter(nil); result != nil {
		t.Fatalf("expected nil filter, got %#v", result)
	}

	result := toInternalSessionFilter(&plugincontract.ListFilter{
		Username: "admin",
		Ip:       "127.0.0.1",
	})
	if result == nil {
		t.Fatal("expected converted filter, got nil")
	}
	if result.Username != "admin" || result.Ip != "127.0.0.1" {
		t.Fatalf("unexpected converted filter: %#v", result)
	}
}

// TestFromInternalSession verifies host-internal session projections are copied into plugin DTOs.
func TestFromInternalSession(t *testing.T) {
	loginTime := gtime.Now()
	sessionItem := &internalsession.Session{
		TokenId:        "token-1",
		UserId:         100,
		Username:       "admin",
		DeptName:       "Engineering",
		Ip:             "127.0.0.1",
		Browser:        "Chrome",
		Os:             "macOS",
		LoginTime:      loginTime,
		LastActiveTime: loginTime,
	}

	result := fromInternalSession(sessionItem)
	if result == nil {
		t.Fatal("expected converted session, got nil")
	}
	if result.TokenId != sessionItem.TokenId ||
		result.UserId != sessionItem.UserId ||
		result.Username != sessionItem.Username ||
		result.DeptName != sessionItem.DeptName ||
		result.Ip != sessionItem.Ip ||
		result.Browser != sessionItem.Browser ||
		result.Os != sessionItem.Os ||
		result.LoginTime != sessionItem.LoginTime ||
		result.LastActiveTime != sessionItem.LastActiveTime {
		t.Fatalf("unexpected converted session: %#v", result)
	}
}

// TestFromInternalSessionListResult verifies nil-safe list conversion and item projection.
func TestFromInternalSessionListResult(t *testing.T) {
	empty := fromInternalSessionListResult(nil)
	if empty == nil {
		t.Fatal("expected empty result, got nil")
	}
	if empty.Total != 0 || len(empty.Items) != 0 {
		t.Fatalf("unexpected empty result: %#v", empty)
	}

	loginTime := gtime.Now()
	result := fromInternalSessionListResult(&internalsession.ListResult{
		Items: []*internalsession.Session{
			{
				TokenId:        "token-2",
				UserId:         101,
				Username:       "demo",
				DeptName:       "QA",
				Ip:             "10.0.0.1",
				Browser:        "Firefox",
				Os:             "Linux",
				LoginTime:      loginTime,
				LastActiveTime: loginTime,
			},
		},
		Total: 1,
	})
	if result.Total != 1 || len(result.Items) != 1 {
		t.Fatalf("unexpected converted list result: %#v", result)
	}
	if result.Items[0] == nil || result.Items[0].TokenId != "token-2" {
		t.Fatalf("unexpected converted item: %#v", result.Items[0])
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
