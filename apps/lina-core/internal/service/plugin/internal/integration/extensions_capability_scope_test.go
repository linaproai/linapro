// This file verifies source-plugin callback flows receive plugin-scoped
// capability services through the public scoping mechanism.

package integration_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/plugin/capability"
	capabilityai "lina-core/pkg/plugin/capability/ai"
	"lina-core/pkg/plugin/capability/contract"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginhost"
)

// capabilityScopeRecorder records plugin-scope binding performed by
// capability.ServicesForPlugin.
type capabilityScopeRecorder struct {
	emptySourceServicesDirectory

	mu     sync.Mutex
	scopes []string
}

var _ capability.Services = (*capabilityScopeRecorder)(nil)
var _ capability.ScopedServicesFactory = (*capabilityScopeRecorder)(nil)

// ForPlugin returns a source-plugin capability view bound to pluginID.
func (r *capabilityScopeRecorder) ForPlugin(pluginID string) capability.Services {
	normalizedPluginID := strings.TrimSpace(pluginID)
	r.mu.Lock()
	r.scopes = append(r.scopes, normalizedPluginID)
	r.mu.Unlock()
	return &scopedSourceServicesDirectory{pluginID: normalizedPluginID}
}

// seenScope reports whether ServicesForPlugin bound pluginID at least once.
func (r *capabilityScopeRecorder) seenScope(pluginID string) bool {
	normalizedPluginID := strings.TrimSpace(pluginID)
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, scope := range r.scopes {
		if scope == normalizedPluginID {
			return true
		}
	}
	return false
}

// scopedSourceServicesDirectory is the plugin-bound Services view
// returned to source-plugin callback code.
type scopedSourceServicesDirectory struct {
	emptySourceServicesDirectory

	pluginID string
}

var _ pluginhost.Services = (*scopedSourceServicesDirectory)(nil)

// scopedPluginID exposes the test-only scope marker for assertions.
func (d *scopedSourceServicesDirectory) scopedPluginID() string {
	if d == nil {
		return ""
	}
	return d.pluginID
}

// APIDoc returns a fallback API-doc service required by source-plugin route registration.
func (d *scopedSourceServicesDirectory) APIDoc() contract.APIDocService {
	return scopedCapabilityAPIDoc{}
}

// Auth returns a no-op auth service required by tenant-core route registration.
func (d *scopedSourceServicesDirectory) Auth() contract.AuthService {
	return scopedCapabilityAuth{}
}

// BizCtx returns a minimal non-nil business context service required by source
// plugin route registration callbacks in this test.
func (d *scopedSourceServicesDirectory) BizCtx() contract.BizCtxService {
	return scopedCapabilityBizCtx{}
}

// Config returns a defaulting plugin configuration service required by plugin cron registration.
func (d *scopedSourceServicesDirectory) Config() contract.ConfigService {
	return scopedCapabilityConfig{}
}

// Notify returns a no-op notification service required by content-notice route registration.
func (d *scopedSourceServicesDirectory) Notify() contract.NotifyService {
	return scopedCapabilityNotify{}
}

// I18n returns a fallback translator required by source-plugin route registration.
func (d *scopedSourceServicesDirectory) I18n() contract.I18nService {
	return scopedCapabilityI18n{}
}

// PluginLifecycle returns no-op lifecycle operations required by tenant-core route registration.
func (d *scopedSourceServicesDirectory) PluginLifecycle() contract.PluginLifecycleService {
	return scopedCapabilityPluginLifecycle{}
}

// PluginState returns a disabled-state reader required by global middleware registration.
func (d *scopedSourceServicesDirectory) PluginState() contract.PluginStateService {
	return scopedCapabilityPluginState{}
}

// Route returns a no-op dynamic-route metadata reader required by audit middleware registration.
func (d *scopedSourceServicesDirectory) Route() contract.RouteService {
	return scopedCapabilityRoute{}
}

// Session returns an empty session service required by monitor-online route registration.
func (d *scopedSourceServicesDirectory) Session() contract.SessionService {
	return scopedCapabilitySession{}
}

// TenantFilter returns a no-op tenant filter service required by source-plugin
// registrations that construct tenant-aware services.
func (d *scopedSourceServicesDirectory) TenantFilter() contract.TenantFilterService {
	return scopedCapabilityTenantFilter{}
}

// scopedCapabilityAPIDoc is a fallback API-doc fixture for registration-only tests.
type scopedCapabilityAPIDoc struct{}

// ResolveRouteText returns the supplied fallback route text.
func (scopedCapabilityAPIDoc) ResolveRouteText(_ context.Context, input contract.RouteTextInput) contract.RouteTextOutput {
	return contract.RouteTextOutput{Title: input.FallbackTitle, Summary: input.FallbackSummary}
}

// ResolveRouteTexts returns fallback route text for each input.
func (scopedCapabilityAPIDoc) ResolveRouteTexts(_ context.Context, inputs []contract.RouteTextInput) []contract.RouteTextOutput {
	outputs := make([]contract.RouteTextOutput, 0, len(inputs))
	for _, input := range inputs {
		outputs = append(outputs, contract.RouteTextOutput{Title: input.FallbackTitle, Summary: input.FallbackSummary})
	}
	return outputs
}

// FindRouteTitleOperationKeys returns no matches in registration-only tests.
func (scopedCapabilityAPIDoc) FindRouteTitleOperationKeys(context.Context, string) []string {
	return nil
}

// scopedCapabilityAuth is a no-op auth fixture for registration-only tests.
type scopedCapabilityAuth struct{}

// SelectTenant returns an empty token output because registration-only tests never authenticate.
func (scopedCapabilityAuth) SelectTenant(context.Context, contract.SelectTenantInput) (*contract.TenantTokenOutput, error) {
	return &contract.TenantTokenOutput{}, nil
}

// SwitchTenant returns an empty token output because registration-only tests never authenticate.
func (scopedCapabilityAuth) SwitchTenant(context.Context, contract.SwitchTenantInput) (*contract.TenantTokenOutput, error) {
	return &contract.TenantTokenOutput{}, nil
}

// IssueImpersonationToken returns an empty token output for registration-only tests.
func (scopedCapabilityAuth) IssueImpersonationToken(context.Context, contract.ImpersonationTokenIssueInput) (*contract.ImpersonationTokenOutput, error) {
	return &contract.ImpersonationTokenOutput{}, nil
}

// RevokeImpersonationToken performs no revocation in registration-only tests.
func (scopedCapabilityAuth) RevokeImpersonationToken(context.Context, contract.ImpersonationTokenRevokeInput) error {
	return nil
}

// scopedCapabilityBizCtx is a minimal plugin-visible business-context fixture.
type scopedCapabilityBizCtx struct{}

// Current returns a platform-scoped context for registration-only tests.
func (scopedCapabilityBizCtx) Current(context.Context) contract.CurrentContext {
	return contract.CurrentContext{PlatformBypass: true}
}

// scopedCapabilityConfig is a defaulting config fixture for registration-only tests.
type scopedCapabilityConfig struct{}

// Get returns no configured value.
func (scopedCapabilityConfig) Get(context.Context, string) (*gvar.Var, error) {
	return nil, nil
}

// Exists reports that no config key exists.
func (scopedCapabilityConfig) Exists(context.Context, string) (bool, error) {
	return false, nil
}

// Scan leaves target unchanged because no test config is present.
func (scopedCapabilityConfig) Scan(context.Context, string, any) error {
	return nil
}

// String returns the supplied default value.
func (scopedCapabilityConfig) String(_ context.Context, _ string, defaultValue string) (string, error) {
	return defaultValue, nil
}

// Bool returns the supplied default value.
func (scopedCapabilityConfig) Bool(_ context.Context, _ string, defaultValue bool) (bool, error) {
	return defaultValue, nil
}

// Int returns the supplied default value.
func (scopedCapabilityConfig) Int(_ context.Context, _ string, defaultValue int) (int, error) {
	return defaultValue, nil
}

// Duration returns the supplied default value.
func (scopedCapabilityConfig) Duration(_ context.Context, _ string, defaultValue time.Duration) (time.Duration, error) {
	return defaultValue, nil
}

// scopedCapabilityNotify is a no-op notification fixture for registration-only tests.
type scopedCapabilityNotify struct{}

// SendNoticePublication records no messages in registration-only tests.
func (scopedCapabilityNotify) SendNoticePublication(context.Context, contract.NoticePublishInput) (*contract.SendOutput, error) {
	return &contract.SendOutput{}, nil
}

// DeleteBySource removes no messages in registration-only tests.
func (scopedCapabilityNotify) DeleteBySource(context.Context, contract.SourceType, []string) error {
	return nil
}

// scopedCapabilityI18n is a fallback translator fixture for registration-only tests.
type scopedCapabilityI18n struct{}

// GetLocale returns the default locale used by registration-only tests.
func (scopedCapabilityI18n) GetLocale(context.Context) string {
	return "zh-CN"
}

// Translate returns fallback text because registration-only tests do not render messages.
func (scopedCapabilityI18n) Translate(_ context.Context, _ string, fallback string) string {
	return fallback
}

// FindMessageKeys returns no keys because registration-only tests do not search messages.
func (scopedCapabilityI18n) FindMessageKeys(context.Context, string, string) []string {
	return nil
}

// scopedCapabilityPluginLifecycle is a no-op lifecycle fixture for registration-only tests.
type scopedCapabilityPluginLifecycle struct{}

// EnsureTenantPluginDisableAllowed always allows tenant plugin disable in registration-only tests.
func (scopedCapabilityPluginLifecycle) EnsureTenantPluginDisableAllowed(context.Context, string, int) error {
	return nil
}

// NotifyTenantPluginDisabled records no notification in registration-only tests.
func (scopedCapabilityPluginLifecycle) NotifyTenantPluginDisabled(context.Context, string, int) {}

// EnsureTenantDeleteAllowed always allows tenant delete in registration-only tests.
func (scopedCapabilityPluginLifecycle) EnsureTenantDeleteAllowed(context.Context, int) error {
	return nil
}

// NotifyTenantDeleted records no notification in registration-only tests.
func (scopedCapabilityPluginLifecycle) NotifyTenantDeleted(context.Context, int) {}

// scopedCapabilityPluginState is a disabled-state fixture for registration-only tests.
type scopedCapabilityPluginState struct{}

// IsEnabled reports false because registration-only tests never execute plugin branches.
func (scopedCapabilityPluginState) IsEnabled(context.Context, string) bool {
	return false
}

// IsProviderEnabled reports false because registration-only tests never activate providers.
func (scopedCapabilityPluginState) IsProviderEnabled(context.Context, string) bool {
	return false
}

// IsEnabledAuthoritative reports false for registration-only global middleware fixtures.
func (scopedCapabilityPluginState) IsEnabledAuthoritative(context.Context, string) bool {
	return false
}

// scopedCapabilityRoute is a no-op route metadata fixture for registration-only tests.
type scopedCapabilityRoute struct{}

// DynamicRouteMetadata returns no dynamic-route metadata.
func (scopedCapabilityRoute) DynamicRouteMetadata(*ghttp.Request) *contract.DynamicRouteMetadata {
	return nil
}

// scopedCapabilitySession is an empty session fixture for registration-only tests.
type scopedCapabilitySession struct{}

// ListPage returns an empty session page.
func (scopedCapabilitySession) ListPage(context.Context, *contract.ListFilter, int, int) (*contract.ListResult, error) {
	return &contract.ListResult{Items: []*contract.Session{}, Total: 0}, nil
}

// Revoke records no revocation in registration-only tests.
func (scopedCapabilitySession) Revoke(context.Context, string) error {
	return nil
}

// scopedCapabilityTenantFilter is a no-op tenant filter fixture for registration-only tests.
type scopedCapabilityTenantFilter struct{}

// Context returns a platform-bypass tenant context for registration-only tests.
func (scopedCapabilityTenantFilter) Context(context.Context) contract.TenantFilterContext {
	return contract.TenantFilterContext{PlatformBypass: true}
}

// Apply returns the model unchanged because registration-only tests never query plugin tables.
func (scopedCapabilityTenantFilter) Apply(_ context.Context, model *gdb.Model, _ string) *gdb.Model {
	return model
}

// emptySourceServicesDirectory is a minimal Services test double.
type emptySourceServicesDirectory struct{}

var _ pluginhost.Services = (*emptySourceServicesDirectory)(nil)

// APIDoc returns no API-doc service for this capability-scope test.
func (emptySourceServicesDirectory) APIDoc() contract.APIDocService { return nil }

// Auth returns no auth service for this capability-scope test.
func (emptySourceServicesDirectory) Auth() contract.AuthService { return nil }

// AI returns the default AI fallback namespace for this capability-scope test.
func (emptySourceServicesDirectory) AI() capabilityai.Service { return capabilityai.New(nil) }

// BizCtx returns no business-context service for this capability-scope test.
func (emptySourceServicesDirectory) BizCtx() contract.BizCtxService { return nil }

// Cache returns no cache service for this capability-scope test.
func (emptySourceServicesDirectory) Cache() contract.CacheService { return nil }

// Config returns no config service for this capability-scope test.
func (emptySourceServicesDirectory) Config() contract.ConfigService { return nil }

// HostConfig returns no host-config service for this capability-scope test.
func (emptySourceServicesDirectory) HostConfig() contract.HostConfigService { return nil }

// I18n returns no i18n service for this capability-scope test.
func (emptySourceServicesDirectory) I18n() contract.I18nService { return nil }

// Manifest returns no manifest service for this capability-scope test.
func (emptySourceServicesDirectory) Manifest() contract.ManifestService { return nil }

// Notify returns no notification service for this capability-scope test.
func (emptySourceServicesDirectory) Notify() contract.NotifyService { return nil }

// Org returns no organization capability for this capability-scope test.
func (emptySourceServicesDirectory) Org() capabilityorgcap.Service { return nil }

// PluginLifecycle returns no plugin lifecycle service for this capability-scope test.
func (emptySourceServicesDirectory) PluginLifecycle() contract.PluginLifecycleService { return nil }

// PluginState returns no plugin-state service for this capability-scope test.
func (emptySourceServicesDirectory) PluginState() contract.PluginStateService { return nil }

// Route returns no route service for this capability-scope test.
func (emptySourceServicesDirectory) Route() contract.RouteService { return nil }

// Session returns no session service for this capability-scope test.
func (emptySourceServicesDirectory) Session() contract.SessionService { return nil }

// Tenant returns no tenant capability for this capability-scope test.
func (emptySourceServicesDirectory) Tenant() tenantcapsvc.Service { return nil }

// TenantFilter returns no tenant-filter service for this capability-scope test.
func (emptySourceServicesDirectory) TenantFilter() contract.TenantFilterService { return nil }

// scopedCapabilityView is implemented by test doubles returned from ForPlugin.
type scopedCapabilityView interface {
	scopedPluginID() string
}

// TestSourcePluginCallbacksUsePluginScopedServices verifies route, cron, hook,
// and managed-cron integration flows all bind runtime services through
// capability.ServicesForPlugin before exposing them to a source plugin.
func TestSourcePluginCallbacksUsePluginScopedServices(t *testing.T) {
	const pluginID = "plugin-dev-source-capability-scope"

	services := testutil.NewServices()
	recorder := &capabilityScopeRecorder{}
	services.Integration.SetCapabilities(recorder)

	observed := make(map[string]string)
	currentPhase := ""
	recordServices := func(label string, services pluginhost.Services) error {
		if services == nil {
			return fmt.Errorf("%s services are nil", label)
		}
		scoped, ok := services.(scopedCapabilityView)
		if !ok {
			return fmt.Errorf("%s services were not plugin-scoped: %T", label, services)
		}
		if got := scoped.scopedPluginID(); got != pluginID {
			return fmt.Errorf("%s services scope = %q, want %q", label, got, pluginID)
		}
		observed[label] = scoped.scopedPluginID()
		return nil
	}

	sourcePlugin := pluginhost.NewSourcePlugin(pluginID)
	sourcePlugin.Assets().UseEmbeddedFiles(fstest.MapFS{
		"plugin.yaml": &fstest.MapFile{Data: []byte(
			"id: " + pluginID + "\n" +
				"name: Source Capability Scope Plugin\n" +
				"version: 0.1.0\n" +
				"type: source\n" +
				"scope_nature: tenant_aware\n" +
				"supports_multi_tenant: true\n" +
				"default_install_mode: tenant_scoped\n",
		)},
		"backend/plugin.go":             &fstest.MapFile{Data: []byte("package backend\n")},
		"frontend/pages/main-entry.vue": &fstest.MapFile{Data: []byte("<template><div /></template>\n")},
	})

	if err := sourcePlugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		func(_ context.Context, registrar pluginhost.HTTPRegistrar) error {
			return recordServices("route", registrar.Services())
		},
	); err != nil {
		t.Fatalf("failed to register source route handler: %v", err)
	}
	if err := sourcePlugin.Cron().RegisterCron(
		pluginhost.ExtensionPointCronRegister,
		pluginhost.CallbackExecutionModeBlocking,
		func(_ context.Context, registrar pluginhost.CronRegistrar) error {
			if currentPhase == "" {
				return fmt.Errorf("cron registration phase was not set")
			}
			return recordServices(currentPhase, registrar.Services())
		},
	); err != nil {
		t.Fatalf("failed to register source cron handler: %v", err)
	}
	if err := sourcePlugin.Hooks().RegisterHook(
		pluginhost.ExtensionPointPluginInstalled,
		pluginhost.CallbackExecutionModeBlocking,
		func(_ context.Context, payload pluginhost.HookPayload) error {
			return recordServices("hook", payload.Services())
		},
	); err != nil {
		t.Fatalf("failed to register source hook handler: %v", err)
	}

	cleanup, err := pluginhost.RegisterSourcePluginForTest(sourcePlugin)
	if err != nil {
		t.Fatalf("failed to register source plugin fixture: %v", err)
	}
	t.Cleanup(cleanup)

	ctx := context.Background()
	server := g.Server("integration-source-capability-scope")
	server.SetDumpRouterMap(false)
	var rootGroup *ghttp.RouterGroup
	server.Group("/", func(group *ghttp.RouterGroup) {
		rootGroup = group
	})

	noopMiddleware := func(req *ghttp.Request) {
		req.Middleware.Next()
	}
	middlewares := pluginhost.NewRouteMiddlewares(
		noopMiddleware,
		noopMiddleware,
		noopMiddleware,
		noopMiddleware,
		noopMiddleware,
		noopMiddleware,
		noopMiddleware,
		noopMiddleware,
	)

	if err = services.Integration.RegisterHTTPRoutes(ctx, server, rootGroup, middlewares); err != nil {
		t.Fatalf("expected route registration to receive scoped services, got error: %v", err)
	}

	currentPhase = "cron"
	if err = services.Integration.RegisterCrons(ctx); err != nil {
		t.Fatalf("expected cron registration to receive scoped services, got error: %v", err)
	}

	if err = services.Integration.DispatchPluginHookEvent(
		ctx,
		pluginhost.ExtensionPointPluginInstalled,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: pluginID,
			Name:     "Source Capability Scope Plugin",
			Version:  "0.1.0",
		}),
	); err != nil {
		t.Fatalf("expected hook dispatch to receive scoped services, got error: %v", err)
	}

	currentPhase = "managed-cron"
	if _, err = services.Integration.ListCronDeclarationsByPlugin(ctx, pluginID); err != nil {
		t.Fatalf("expected managed cron collection to receive scoped services, got error: %v", err)
	}

	for _, label := range []string{"route", "cron", "hook", "managed-cron"} {
		if got := observed[label]; got != pluginID {
			t.Fatalf("expected %s callback to receive plugin-scoped services for %q, got %q", label, pluginID, got)
		}
	}
	if !recorder.seenScope(pluginID) {
		t.Fatalf("expected capability.ServicesForPlugin to bind %q", pluginID)
	}
}
