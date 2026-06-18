// This file verifies host-service adapters keep source-plugin contracts
// independent from concrete host service implementations.

package capabilityhost

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/model"
	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	internalbizctx "lina-core/internal/service/bizctx"
	"lina-core/internal/service/hostlock"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/locker"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap/token"
	"lina-core/pkg/plugin/pluginhost"
)

// TestAPIDocAdapterConvertsRouteTextDTOs verifies plugin apidoc calls are
// translated to host apidoc DTOs inside the capabilityhost boundary.
func TestAPIDocAdapterConvertsRouteTextDTOs(t *testing.T) {
	ctx := context.Background()
	resolver := &fakeAPIDocResolver{}
	svc := newAPIDocAdapter(resolver)

	single := svc.ResolveRouteText(ctx, apidoccap.RouteTextInput{
		OperationKey:    "api.plugin.route",
		Method:          "GET",
		Path:            "/plugin/routes",
		FallbackTitle:   "Plugin",
		FallbackSummary: "List routes",
	})
	if single.Title != "localized-api.plugin.route" || single.Summary != "summary-/plugin/routes" {
		t.Fatalf("unexpected single route text: %#v", single)
	}
	if resolver.singleInput.OperationKey != "api.plugin.route" || resolver.singleInput.Method != "GET" || resolver.singleInput.Path != "/plugin/routes" {
		t.Fatalf("expected host apidoc input to be recorded, got %#v", resolver.singleInput)
	}

	batch := svc.ResolveRouteTexts(ctx, []apidoccap.RouteTextInput{
		{OperationKey: "api.plugin.one", Method: "GET", Path: "/plugin/one", FallbackTitle: "One", FallbackSummary: "List one"},
		{OperationKey: "api.plugin.two", Method: "POST", Path: "/plugin/two", FallbackTitle: "Two", FallbackSummary: "Create two"},
	})
	if len(batch) != 2 || batch[0].Title != "localized-api.plugin.one" || batch[1].Summary != "summary-/plugin/two" {
		t.Fatalf("unexpected batch route text: %#v", batch)
	}
	if len(resolver.batchInputs) != 2 || resolver.batchInputs[1].Method != "POST" {
		t.Fatalf("expected host apidoc batch inputs to be recorded, got %#v", resolver.batchInputs)
	}

	keys := svc.FindRouteTitleOperationKeys(ctx, "plugin")
	if len(keys) != 2 || keys[0] != "api.plugin.one" || keys[1] != "api.plugin.two" {
		t.Fatalf("unexpected operation key matches: %#v", keys)
	}
}

// TestAuthAdapterUsesTenantTokenIssuer verifies plugin auth calls depend on the narrowed token issuer.
func TestAuthAdapterUsesTenantTokenIssuer(t *testing.T) {
	ctx := context.Background()
	issuer := &fakeTenantTokenIssuer{}
	svc := &authAdapter{tokenIssuer: issuer}

	selected, err := svc.SelectTenant(ctx, token.SelectTenantInput{PreToken: "pre-token", TenantID: 11})
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

	switched, err := svc.SwitchTenant(ctx, token.SwitchTenantInput{BearerToken: "bearer-token", TenantID: 22})
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

	impersonated, err := svc.IssueImpersonationToken(ctx, token.ImpersonationTokenIssueInput{ActingUserID: 1, TenantID: 33})
	if err != nil {
		t.Fatalf("issue impersonation token: %v", err)
	}
	if impersonated.AccessToken != "impersonation-token" ||
		impersonated.TokenID != "impersonation-token-id" ||
		impersonated.TenantID != 33 ||
		impersonated.ActingUserID != 1 ||
		issuer.impersonationActingUserID != 1 ||
		issuer.impersonationTenantID != 33 {
		t.Fatalf("expected impersonation issue call, out=%#v issuer=%#v", impersonated, issuer)
	}

	if err = svc.RevokeImpersonationToken(ctx, token.ImpersonationTokenRevokeInput{BearerToken: "Bearer impersonation-token", TenantID: 33}); err != nil {
		t.Fatalf("revoke impersonation token: %v", err)
	}
	if issuer.revokedImpersonationBearer != "Bearer impersonation-token" || issuer.revokedImpersonationTenantID != 33 {
		t.Fatalf("expected impersonation revoke call, issuer=%#v", issuer)
	}
}

// TestI18nAdapterFindMessageKeysUsesHostExport verifies runtime message export
// stays behind the capabilityhost adapter instead of leaking into startup.
func TestI18nAdapterFindMessageKeysUsesHostExport(t *testing.T) {
	ctx := context.Background()
	runtimeI18n := &fakeRuntimeI18nService{
		locale: "en-US",
		messages: map[string]string{
			"plugin.menu.analytics": "Analytics",
			"plugin.menu.audit":     "Audit Log",
			"plugin.form.title":     "Plugin Form",
		},
	}
	svc := newI18nAdapter(runtimeI18n)

	if locale := svc.GetLocale(ctx); locale != "en-US" {
		t.Fatalf("expected locale en-US, got %q", locale)
	}
	if translated := svc.Translate(ctx, "plugin.menu.analytics", "fallback"); translated != "translated-plugin.menu.analytics" {
		t.Fatalf("unexpected translation: %q", translated)
	}

	keys := svc.FindMessageKeys(ctx, "plugin.menu.", "aud")
	if len(keys) != 1 || keys[0] != "plugin.menu.audit" {
		t.Fatalf("unexpected i18n key matches: %#v", keys)
	}
	if runtimeI18n.exportedLocale != "en-US" {
		t.Fatalf("expected export locale en-US, got %q", runtimeI18n.exportedLocale)
	}
}

// TestBizCtxAdapterPlatformBypassRequiresAllDataPlatformContext verifies
// source plugins receive the same strict platform-bypass semantics used by host
// tenantcap instead of a tenant-id-only shortcut.
func TestBizCtxAdapterPlatformBypassRequiresAllDataPlatformContext(t *testing.T) {
	adapter := newBizCtxAdapter(internalbizctx.New())
	testCases := []struct {
		name     string
		ctx      *model.Context
		expected bool
	}{
		{name: "platform all data", ctx: &model.Context{TenantId: 0, DataScope: 1}, expected: true},
		{name: "platform tenant scope", ctx: &model.Context{TenantId: 0, DataScope: 2}, expected: false},
		{name: "impersonation", ctx: &model.Context{TenantId: 0, DataScope: 1, ActingAsTenant: true, IsImpersonation: true}, expected: false},
		{name: "tenant context", ctx: &model.Context{TenantId: 1001, DataScope: 1}, expected: false},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), testCase.ctx)
			current := adapter.Current(ctx)
			if current.PlatformBypass != testCase.expected {
				t.Fatalf("expected PlatformBypass=%t, got %#v", testCase.expected, current)
			}
		})
	}
}

// fakeAPIDocResolver records route-text DTOs passed to the host apidoc boundary.
type fakeAPIDocResolver struct {
	singleInput apidoc.RouteTextInput
	batchInputs []apidoc.RouteTextInput
}

// ResolveRouteText records and localizes one host apidoc route-text input.
func (f *fakeAPIDocResolver) ResolveRouteText(_ context.Context, input apidoc.RouteTextInput) apidoc.RouteTextOutput {
	f.singleInput = input
	return apidoc.RouteTextOutput{
		Title:   "localized-" + input.OperationKey,
		Summary: "summary-" + input.Path,
	}
}

// ResolveRouteTexts records and localizes host apidoc route-text inputs.
func (f *fakeAPIDocResolver) ResolveRouteTexts(_ context.Context, inputs []apidoc.RouteTextInput) []apidoc.RouteTextOutput {
	f.batchInputs = append([]apidoc.RouteTextInput(nil), inputs...)
	outputs := make([]apidoc.RouteTextOutput, 0, len(inputs))
	for _, input := range inputs {
		outputs = append(outputs, apidoc.RouteTextOutput{
			Title:   "localized-" + input.OperationKey,
			Summary: "summary-" + input.Path,
		})
	}
	return outputs
}

// FindRouteTitleOperationKeys returns deterministic route-title matches.
func (f *fakeAPIDocResolver) FindRouteTitleOperationKeys(_ context.Context, _ string) []string {
	return []string{"api.plugin.one", "api.plugin.two"}
}

// fakeTenantTokenIssuer records plugin adapter calls for contract tests.
type fakeTenantTokenIssuer struct {
	issuedPreToken               string
	issuedTenantID               int
	reissuedBearer               string
	reissuedTenantID             int
	impersonationActingUserID    int
	impersonationTenantID        int
	revokedImpersonationBearer   string
	revokedImpersonationTenantID int
}

// IssueTenantToken records one pre-login token exchange.
func (f *fakeTenantTokenIssuer) IssueTenantToken(
	_ context.Context,
	in auth.TenantTokenIssueInput,
) (*auth.TenantTokenOutput, error) {
	f.issuedPreToken = in.PreToken
	f.issuedTenantID = in.TenantID
	return &auth.TenantTokenOutput{AccessToken: "issued-token", RefreshToken: "issued-refresh-token"}, nil
}

// ReissueTenantTokenFromBearer records one bearer-token tenant switch.
func (f *fakeTenantTokenIssuer) ReissueTenantTokenFromBearer(
	_ context.Context,
	tokenString string,
	tenantID int,
) (*auth.TenantTokenOutput, error) {
	f.reissuedBearer = tokenString
	f.reissuedTenantID = tenantID
	return &auth.TenantTokenOutput{AccessToken: "reissued-token", RefreshToken: "reissued-refresh-token"}, nil
}

// IssueImpersonationToken records one host-owned impersonation token request.
func (f *fakeTenantTokenIssuer) IssueImpersonationToken(
	_ context.Context,
	in auth.ImpersonationTokenIssueInput,
) (*auth.ImpersonationTokenOutput, error) {
	f.impersonationActingUserID = in.ActingUserID
	f.impersonationTenantID = in.TenantID
	return &auth.ImpersonationTokenOutput{
		AccessToken:  "impersonation-token",
		TokenID:      "impersonation-token-id",
		TenantID:     in.TenantID,
		ActingUserID: in.ActingUserID,
	}, nil
}

// RevokeImpersonationToken records one host-owned impersonation revoke request.
func (f *fakeTenantTokenIssuer) RevokeImpersonationToken(_ context.Context, tokenString string, tenantID int) error {
	f.revokedImpersonationBearer = tokenString
	f.revokedImpersonationTenantID = tenantID
	return nil
}

// fakeRuntimeI18nService records runtime i18n calls used by capabilityhost tests.
type fakeRuntimeI18nService struct {
	locale         string
	exportedLocale string
	messages       map[string]string
}

// GetLocale returns the configured request locale.
func (f *fakeRuntimeI18nService) GetLocale(context.Context) string {
	return f.locale
}

// Translate returns a deterministic translation marker for assertions.
func (f *fakeRuntimeI18nService) Translate(_ context.Context, key string, _ string) string {
	return "translated-" + key
}

// ExportMessages returns the configured flat runtime message catalog.
func (f *fakeRuntimeI18nService) ExportMessages(_ context.Context, locale string) i18nsvc.MessageExportOutput {
	f.exportedLocale = locale
	return i18nsvc.MessageExportOutput{
		Locale:   locale,
		Mode:     "effective",
		Messages: f.messages,
	}
}

// TestNewWiresCompleteAdminDirectory verifies source plugins receive every
// typed management domain advertised by capability.AdminServices.
func TestNewWiresCompleteAdminDirectory(t *testing.T) {
	lockSvc, err := hostlock.New(locker.New())
	if err != nil {
		t.Fatalf("create lock service: %v", err)
	}
	services, err := New(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		kvcache.New(),
		lockSvc,
		nil,
		NewLocalStorageProvider(t.TempDir()),
	)
	if err != nil {
		t.Fatalf("create host services: %v", err)
	}
	hostServices, ok := services.(pluginhost.Services)
	if !ok {
		t.Fatalf("expected New to return pluginhost.Services, got %T", services)
	}
	admin := hostServices.Admin()
	if admin == nil {
		t.Fatal("expected admin directory")
	}
	if admin.Users() == nil {
		t.Fatal("expected user admin service")
	}
	if admin.Auth() == nil || admin.Auth().Authz() == nil {
		t.Fatal("expected authz admin service")
	}
	if admin.Dict() == nil {
		t.Fatal("expected dict admin service")
	}
	if admin.Files() == nil {
		t.Fatal("expected file admin service")
	}
	if admin.Sessions() == nil {
		t.Fatal("expected session admin service")
	}
	if admin.HostConfig() == nil {
		t.Fatal("expected host config admin service")
	}
	if admin.Notifications() == nil {
		t.Fatal("expected notification admin service")
	}
	if admin.Plugins() == nil {
		t.Fatal("expected plugin admin service")
	}
	if admin.Jobs() == nil {
		t.Fatal("expected job admin service")
	}
	if admin.Infra() == nil {
		t.Fatal("expected infra admin service")
	}
}
