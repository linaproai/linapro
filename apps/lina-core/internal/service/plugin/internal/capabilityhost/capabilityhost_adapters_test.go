// This file verifies host-service adapters keep source-plugin contracts
// independent from concrete host service implementations.

package capabilityhost

import (
	"context"
	jobv1 "lina-core/api/job/v1"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/model"
	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	internalbizctx "lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/hostlock"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/jobmeta"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/plugin/internal/pluginconfig"
	storagesvc "lina-core/internal/service/storage"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap/token"
	capabilityfilecap "lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/storagecap"
	capabilitytenantcap "lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// TestAPIDocAdapterConvertsRouteTextDTOs verifies plugin apidoc calls are
// translated to host apidoc DTOs inside the capabilityhost boundary.
func TestAPIDocAdapterConvertsRouteTextDTOs(t *testing.T) {
	var (
		ctx      = context.Background()
		resolver = &fakeAPIDocResolver{}
		svc      = newAPIDocAdapter(resolver)
	)

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

// TestScopedDirectoryFilesInjectsPluginStorage verifies file capabilities can
// copy from the same plugin-scoped Storage() view exposed by the directory.
func TestScopedDirectoryFilesInjectsPluginStorage(t *testing.T) {
	recorder := &scopedFilesRecorder{}
	base := &directory{
		files:           recorder,
		storageProvider: newCapabilityHostLocalStorageProviderForTest(t),
	}

	files := (&scopedDirectory{base: base, pluginID: "demo-plugin"}).Files()
	scopedRecorder, ok := files.(*scopedFilesRecorder)
	if !ok {
		t.Fatalf("expected scoped file recorder, got %T", files)
	}
	if scopedRecorder.storage == nil {
		t.Fatal("expected scoped file adapter to receive storage service")
	}
	output, err := scopedRecorder.storage.Put(context.Background(), storagecap.PutInput{
		Path:      "exports/report.txt",
		Body:      strings.NewReader("report"),
		Size:      int64(len("report")),
		Overwrite: true,
	})
	if err != nil {
		t.Fatalf("write through scoped storage: %v", err)
	}
	if output == nil || output.Object == nil || output.Object.Path != "exports/report.txt" {
		t.Fatalf("unexpected scoped storage output: %#v", output)
	}
}

// TestAuthAdapterUsesAuthService verifies plugin auth calls depend on auth.Service.
func TestAuthAdapterUsesAuthService(t *testing.T) {
	var (
		ctx    = context.Background()
		issuer = &fakeAuthService{}
		svc    = newAuthAdapter(issuer)
	)

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

// scopedFilesRecorder records the storage service passed to a scoped file adapter.
type scopedFilesRecorder struct {
	capabilityfilecap.Service
	storage storagecap.Service
}

// WithStorage returns a copy that records the plugin-scoped storage service.
func (s *scopedFilesRecorder) WithStorage(storage storagecap.Service) capabilityfilecap.Service {
	return &scopedFilesRecorder{Service: s.Service, storage: storage}
}

// TestI18nAdapterTranslatesThroughHostService verifies runtime translation
// stays behind the capabilityhost adapter instead of leaking into startup.
func TestI18nAdapterTranslatesThroughHostService(t *testing.T) {
	ctx := context.Background()
	runtimeI18n := &fakeRuntimeI18nService{
		locale: "en-US",
	}
	svc := newI18nAdapter(runtimeI18n)

	if locale := svc.GetLocale(ctx); locale != "en-US" {
		t.Fatalf("expected locale en-US, got %q", locale)
	}
	if translated := svc.Translate(ctx, "plugin.menu.analytics", "fallback"); translated != "translated-plugin.menu.analytics" {
		t.Fatalf("unexpected translation: %q", translated)
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
	apidoc.Service
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

// fakeAuthService records plugin auth adapter calls for contract tests.
type fakeAuthService struct {
	auth.Service

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
func (f *fakeAuthService) IssueTenantToken(
	_ context.Context,
	in auth.TenantTokenIssueInput,
) (*auth.TenantTokenOutput, error) {
	f.issuedPreToken = in.PreToken
	f.issuedTenantID = in.TenantID
	return &auth.TenantTokenOutput{AccessToken: "issued-token", RefreshToken: "issued-refresh-token"}, nil
}

// ReissueTenantTokenFromBearer records one bearer-token tenant switch.
func (f *fakeAuthService) ReissueTenantTokenFromBearer(
	_ context.Context,
	tokenString string,
	tenantID int,
) (*auth.TenantTokenOutput, error) {
	f.reissuedBearer = tokenString
	f.reissuedTenantID = tenantID
	return &auth.TenantTokenOutput{AccessToken: "reissued-token", RefreshToken: "reissued-refresh-token"}, nil
}

// IssueImpersonationToken records one host-owned impersonation token request.
func (f *fakeAuthService) IssueImpersonationToken(
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
func (f *fakeAuthService) RevokeImpersonationToken(_ context.Context, tokenString string, tenantID int) error {
	f.revokedImpersonationBearer = tokenString
	f.revokedImpersonationTenantID = tenantID
	return nil
}

// fakeRuntimeI18nService records runtime i18n calls used by capabilityhost tests.
type fakeRuntimeI18nService struct {
	i18nsvc.Service

	locale string
}

// GetLocale returns the configured request locale.
func (f *fakeRuntimeI18nService) GetLocale(context.Context) string {
	return f.locale
}

// Translate returns a deterministic translation marker for assertions.
func (f *fakeRuntimeI18nService) Translate(_ context.Context, key string, _ string) string {
	return "translated-" + key
}

// TestNewWiresCompleteUnifiedDirectory verifies source plugins receive the
// unified domain services directly from capability.Services.
func TestNewWiresCompleteUnifiedDirectory(t *testing.T) {
	lockSvc, err := hostlock.New(locker.New())
	if err != nil {
		t.Fatalf("create lock service: %v", err)
	}
	services, err := New(
		nil,
		nil,
		nil,
		nil,
		testHostConfigService{},
		nil,
		cachecoord.New(cachecoord.NewStaticTopology(false)),
		nil,
		nil,
		nil,
		nil,
		nil,
		noopJobOwner{},
		nil,
		newCapabilityHostTestTenantService(),
		nil,
		kvcache.New(),
		lockSvc,
		pluginconfig.NewFactory("", ""),
		nil,
		newCapabilityHostLocalStorageProviderForTest(t),
	)
	if err != nil {
		t.Fatalf("create host services: %v", err)
	}
	hostServices := services
	if hostServices == nil {
		t.Fatalf("expected source-plugin host services")
	}
	if hostServices.Users() == nil {
		t.Fatal("expected user service")
	}
	if hostServices.Auth() == nil || hostServices.Auth().Authz() == nil {
		t.Fatal("expected authz service")
	}
	if hostServices.Dict() == nil {
		t.Fatal("expected dict service")
	}
	if hostServices.Files() == nil {
		t.Fatal("expected file service")
	}
	if hostServices.Sessions() == nil {
		t.Fatal("expected session service")
	}
	if hostServices.HostConfig() == nil {
		t.Fatal("expected host config service")
	}
	if hostServices.Notifications() == nil {
		t.Fatal("expected notification service")
	}
	if hostServices.Plugins() == nil {
		t.Fatal("expected plugin service")
	}
	if hostServices.Jobs() == nil {
		t.Fatal("expected job service")
	}
	if hostServices.Plugins().Lifecycle() == nil {
		t.Fatal("expected source-plugin lifecycle service")
	}
	if hostServices.Plugins().State() == nil {
		t.Fatal("expected source-plugin state service")
	}
	if hostServices.Tenant() == nil || hostServices.Tenant().Plugins() == nil {
		t.Fatal("expected tenant plugin governance service")
	}
	if hostServices.Tenant().Filter() == nil {
		t.Fatal("expected tenant filter context service")
	}
}

// TestNewReusesTenantServiceFilterContract verifies capabilityhost reuses the
// tenant service filter contract instead of constructing a local duplicate.
func TestNewReusesTenantServiceFilterContract(t *testing.T) {
	lockSvc, err := hostlock.New(locker.New())
	if err != nil {
		t.Fatalf("create lock service: %v", err)
	}
	filterSvc := &testTenantFilterService{}
	tenantSvc := testTenantService{
		Service: tenantspi.New(nil, nil, nil, internalbizctx.New()),
		filter:  filterSvc,
	}
	services, err := New(
		nil,
		nil,
		nil,
		nil,
		testHostConfigService{},
		nil,
		cachecoord.New(cachecoord.NewStaticTopology(false)),
		nil,
		nil,
		nil,
		nil,
		nil,
		noopJobOwner{},
		nil,
		tenantSvc,
		nil,
		kvcache.New(),
		lockSvc,
		pluginconfig.NewFactory("", ""),
		nil,
		newCapabilityHostLocalStorageProviderForTest(t),
	)
	if err != nil {
		t.Fatalf("create host services: %v", err)
	}
	hostServices := services
	if hostServices == nil {
		t.Fatalf("expected source-plugin host services")
	}
	if got := hostServices.Tenant().Filter(); got != filterSvc {
		t.Fatalf("expected tenant service filter contract, got %T", got)
	}
}

// TestNewRequiresHostConfigService verifies missing host config dependency is
// rejected during source-plugin host service construction.
func TestNewRequiresHostConfigService(t *testing.T) {
	lockSvc, err := hostlock.New(locker.New())
	if err != nil {
		t.Fatalf("create lock service: %v", err)
	}
	_, err = New(
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
		noopJobOwner{},
		nil,
		nil,
		nil,
		kvcache.New(),
		lockSvc,
		pluginconfig.NewFactory("", ""),
		nil,
		newCapabilityHostLocalStorageProviderForTest(t),
	)
	if err == nil || !strings.Contains(err.Error(), "host config service is nil") {
		t.Fatalf("expected missing host config dependency to fail, got %v", err)
	}
}

// TestNewRequiresTenantService verifies tenant filtering is supplied by the
// startup-owned tenant service rather than created inside capabilityhost.
func TestNewRequiresTenantService(t *testing.T) {
	lockSvc, err := hostlock.New(locker.New())
	if err != nil {
		t.Fatalf("create lock service: %v", err)
	}
	_, err = New(
		nil,
		nil,
		nil,
		nil,
		testHostConfigService{},
		nil,
		cachecoord.New(cachecoord.NewStaticTopology(false)),
		nil,
		nil,
		nil,
		nil,
		nil,
		noopJobOwner{},
		nil,
		nil,
		nil,
		kvcache.New(),
		lockSvc,
		pluginconfig.NewFactory("", ""),
		nil,
		newCapabilityHostLocalStorageProviderForTest(t),
	)
	if err == nil || !strings.Contains(err.Error(), "tenant service is nil") {
		t.Fatalf("expected missing tenant service dependency to fail, got %v", err)
	}
}

func newCapabilityHostTestTenantService() tenantspi.Service {
	return tenantspi.New(nil, nil, nil, internalbizctx.New())
}

func newCapabilityHostLocalStorageProviderForTest(t *testing.T) storagecap.Provider {
	t.Helper()
	return NewLocalStorageProvider(storagesvc.New(storagesvc.Config{NamespaceRoots: map[string]string{
		storagesvc.NamespacePlugins: t.TempDir(),
	}}))
}

// testHostConfigService satisfies hostconfigcap.Service for capabilityhost
// wiring tests that only verify directory construction.
type testHostConfigService struct{}

// Get returns the supplied default host config value.
func (testHostConfigService) Get(_ context.Context, _ string, defaultValue any) (*gvar.Var, error) {
	if defaultValue == nil {
		return nil, nil
	}
	return gvar.New(defaultValue), nil
}

// Exists reports that no static host config keys are configured.
func (testHostConfigService) Exists(context.Context, string) (bool, error) {
	return false, nil
}

// String returns the supplied default string.
func (testHostConfigService) String(_ context.Context, _ string, defaultValue string) (string, error) {
	return defaultValue, nil
}

// Bool returns the supplied default boolean.
func (testHostConfigService) Bool(_ context.Context, _ string, defaultValue bool) (bool, error) {
	return defaultValue, nil
}

// Int returns the supplied default integer.
func (testHostConfigService) Int(_ context.Context, _ string, defaultValue int) (int, error) {
	return defaultValue, nil
}

// Duration returns the supplied default duration.
func (testHostConfigService) Duration(_ context.Context, _ string, defaultValue time.Duration) (time.Duration, error) {
	return defaultValue, nil
}

// SysConfig returns no sys_config subresource for directory wiring tests.
func (testHostConfigService) SysConfig() hostconfigcap.SysConfigService {
	return nil
}

// noopJobOwner satisfies the Jobs owner contract for capabilityhost wiring tests.
type noopJobOwner struct {
	jobmeta.Owner
}

// CreateJob accepts one scheduled-job create request.
func (noopJobOwner) CreateJob(_ context.Context, _ jobmeta.SaveJobInput) (int64, error) {
	return 0, nil
}

// UpdateJob accepts one scheduled-job update request.
func (noopJobOwner) UpdateJob(_ context.Context, _ jobmeta.UpdateJobInput) error {
	return nil
}

// DeleteJobs accepts one scheduled-job delete request.
func (noopJobOwner) DeleteJobs(_ context.Context, _ string) error {
	return nil
}

// UpdateJobStatus accepts one scheduled-job status request.
func (noopJobOwner) UpdateJobStatus(_ context.Context, _ int64, _ jobv1.Status) error {
	return nil
}

// TriggerJob accepts one scheduled-job manual trigger request.
func (noopJobOwner) TriggerJob(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}

// testTenantService wraps the default tenant service and overrides Filter for
// capabilityhost dependency-source assertions.
type testTenantService struct {
	tenantspi.Service
	filter capabilitytenantcap.FilterService
}

// Filter returns the injected filter service.
func (s testTenantService) Filter() capabilitytenantcap.FilterService {
	return s.filter
}

// testTenantFilterService records no state; identity is the assertion surface.
type testTenantFilterService struct{}

// Context returns an empty tenant filter context for construction tests.
func (*testTenantFilterService) Context(context.Context) capabilitytenantcap.TenantFilterContext {
	return capabilitytenantcap.TenantFilterContext{}
}
