// This file verifies host-service adapters keep source-plugin contracts
// independent from concrete host service implementations.

package hostservices

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/model"
	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	internalbizctx "lina-core/internal/service/bizctx"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/notify"
	internalsession "lina-core/internal/service/session"
	plugincontract "lina-core/pkg/plugin/capability/contract"
)

// TestAPIDocAdapterConvertsRouteTextDTOs verifies plugin apidoc calls are
// translated to host apidoc DTOs inside the hostservices boundary.
func TestAPIDocAdapterConvertsRouteTextDTOs(t *testing.T) {
	ctx := context.Background()
	resolver := &fakeAPIDocResolver{}
	svc := newAPIDocAdapter(resolver)

	single := svc.ResolveRouteText(ctx, plugincontract.RouteTextInput{
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

	batch := svc.ResolveRouteTexts(ctx, []plugincontract.RouteTextInput{
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

	impersonated, err := svc.IssueImpersonationToken(ctx, plugincontract.ImpersonationTokenIssueInput{ActingUserID: 1, TenantID: 33})
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

	if err = svc.RevokeImpersonationToken(ctx, plugincontract.ImpersonationTokenRevokeInput{BearerToken: "Bearer impersonation-token", TenantID: 33}); err != nil {
		t.Fatalf("revoke impersonation token: %v", err)
	}
	if issuer.revokedImpersonationBearer != "Bearer impersonation-token" || issuer.revokedImpersonationTenantID != 33 {
		t.Fatalf("expected impersonation revoke call, issuer=%#v", issuer)
	}
}

// TestI18nAdapterFindMessageKeysUsesHostExport verifies runtime message export
// stays behind the hostservices adapter instead of leaking into startup.
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

// TestNotifyAdapterConvertsNoticeDTOs verifies source-plugin notify DTOs are
// converted to host notify DTOs inside the hostservices boundary.
func TestNotifyAdapterConvertsNoticeDTOs(t *testing.T) {
	ctx := context.Background()
	publisher := &fakeNotifyPublisher{}
	svc := newNotifyAdapter(publisher)

	output, err := svc.SendNoticePublication(ctx, plugincontract.NoticePublishInput{
		NoticeID:     1001,
		Title:        "Release",
		Content:      "Published",
		CategoryCode: plugincontract.CategoryCodeOther,
		SenderUserID: 42,
	})
	if err != nil {
		t.Fatalf("send notice publication: %v", err)
	}
	if output == nil || output.MessageID != 9001 || output.DeliveryCount != 3 {
		t.Fatalf("unexpected notify output: %#v", output)
	}
	if publisher.noticeInput.NoticeID != 1001 ||
		publisher.noticeInput.CategoryCode != notify.CategoryCodeOther ||
		publisher.noticeInput.SenderUserID != 42 {
		t.Fatalf("expected host notify input to be recorded, got %#v", publisher.noticeInput)
	}

	if err = svc.DeleteBySource(ctx, plugincontract.SourceTypeNotice, []string{"1001"}); err != nil {
		t.Fatalf("delete by source: %v", err)
	}
	if publisher.deletedSourceType != notify.SourceTypeNotice || len(publisher.deletedSourceIDs) != 1 || publisher.deletedSourceIDs[0] != "1001" {
		t.Fatalf("expected host notify delete input to be recorded, got %q %#v", publisher.deletedSourceType, publisher.deletedSourceIDs)
	}
}

// TestToInternalSessionFilter verifies the published filter contract is converted explicitly.
func TestToInternalSessionFilter(t *testing.T) {
	if result := toInternalSessionFilter(nil); result != nil {
		t.Fatalf("expected nil filter, got %#v", result)
	}

	filter := &plugincontract.ListFilter{
		Username: "admin",
		Ip:       "127.0.0.1",
	}
	result := toInternalSessionFilter(filter)
	if result == nil {
		t.Fatal("expected converted filter, got nil")
	}
	if result.Username != "admin" || result.Ip != "127.0.0.1" {
		t.Fatalf("unexpected converted filter: %#v", result)
	}
}

// TestFromInternalSession verifies host-internal session projections are copied into plugin DTOs.
func TestFromInternalSession(t *testing.T) {
	loginTime := time.Now()
	sessionItem := &internalsession.Session{
		TokenId:        "token-1",
		UserId:         100,
		Username:       "admin",
		ClientType:     "desktop",
		DeptName:       "Engineering",
		Ip:             "127.0.0.1",
		Browser:        "Chrome",
		Os:             "macOS",
		LoginTime:      &loginTime,
		LastActiveTime: &loginTime,
	}

	result := fromInternalSession(sessionItem)
	if result == nil {
		t.Fatal("expected converted session, got nil")
	}
	if result.TokenId != sessionItem.TokenId ||
		result.UserId != sessionItem.UserId ||
		result.Username != sessionItem.Username ||
		result.ClientType != sessionItem.ClientType ||
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

	loginTime := time.Now()
	result := fromInternalSessionListResult(&internalsession.ListResult{
		Items: []*internalsession.Session{
			{
				TokenId:        "token-2",
				UserId:         101,
				Username:       "demo",
				ClientType:     "mobile",
				DeptName:       "QA",
				Ip:             "10.0.0.1",
				Browser:        "Firefox",
				Os:             "Linux",
				LoginTime:      &loginTime,
				LastActiveTime: &loginTime,
			},
		},
		Total: 1,
	})
	if result.Total != 1 || len(result.Items) != 1 {
		t.Fatalf("unexpected converted list result: %#v", result)
	}
	if result.Items[0] == nil || result.Items[0].TokenId != "token-2" || result.Items[0].ClientType != "mobile" {
		t.Fatalf("unexpected converted item: %#v", result.Items[0])
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

// fakeRuntimeI18nService records runtime i18n calls used by hostservices tests.
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

// fakeNotifyPublisher records notify DTOs passed to the host notify boundary.
type fakeNotifyPublisher struct {
	noticeInput       notify.NoticePublishInput
	deletedSourceType notify.SourceType
	deletedSourceIDs  []string
}

// SendNoticePublication records one host notify publication input.
func (f *fakeNotifyPublisher) SendNoticePublication(_ context.Context, in notify.NoticePublishInput) (*notify.SendOutput, error) {
	f.noticeInput = in
	return &notify.SendOutput{MessageID: 9001, DeliveryCount: 3}, nil
}

// DeleteBySource records one host notify delete request.
func (f *fakeNotifyPublisher) DeleteBySource(_ context.Context, sourceType notify.SourceType, sourceIDs []string) error {
	f.deletedSourceType = sourceType
	f.deletedSourceIDs = append([]string(nil), sourceIDs...)
	return nil
}
