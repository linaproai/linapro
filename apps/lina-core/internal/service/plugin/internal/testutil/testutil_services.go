// This file wires plugin sub-services and request-context adapters for tests.

package testutil

import (
	"context"
	"time"

	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	configsvc "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/openapi"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/pkg/pluginhost"
	pluginserviceconfig "lina-core/pkg/pluginservice/config"
	"lina-core/pkg/pluginservice/contract"
)

// Services groups the wired plugin sub-services used by package-level tests.
type Services struct {
	// Catalog provides manifest discovery, registry, and release access.
	Catalog catalog.Service
	// Lifecycle provides install and uninstall orchestration.
	Lifecycle lifecycle.Service
	// Runtime provides artifact parsing, reconcile, and route execution.
	Runtime runtime.Service
	// Frontend provides in-memory frontend bundle management.
	Frontend frontend.Service
	// Integration provides menu, hook, and resource-ref integration.
	Integration integration.Service
	// OpenAPI provides dynamic route OpenAPI projection.
	OpenAPI openapi.Service
}

// singleNodeTopology provides the default non-clustered topology for plugin tests.
type singleNodeTopology struct{}

// IsClusterModeEnabled reports that package tests run in single-node mode.
func (singleNodeTopology) IsClusterModeEnabled() bool {
	return false
}

// IsPrimaryNode reports that the local test node owns primary-only work.
func (singleNodeTopology) IsPrimaryNode() bool {
	return true
}

// CurrentNodeID returns the fixed node identifier used by package tests.
func (singleNodeTopology) CurrentNodeID() string {
	return "test-node"
}

// NewServices creates a fully wired plugin sub-service set for tests.
func NewServices() *Services {
	var (
		configProvider = configsvc.New()
		bizCtxProvider = bizctx.New()
		cacheCoordSvc  = cachecoord.Default(cachecoord.NewStaticTopology(false))
		i18nService    = i18nsvc.New(bizCtxProvider, configProvider, cacheCoordSvc)
		catalogSvc     = catalog.New(configProvider)
		lifecycleSvc   = lifecycle.New(catalogSvc)
		frontendSvc    = frontend.New(catalogSvc)
		openapiSvc     = openapi.New(catalogSvc)
		runtimeSvc     = runtime.New(catalogSvc, lifecycleSvc, frontendSvc, openapiSvc, i18nService)
		integrationSvc = integration.New(catalogSvc)
		topology       = singleNodeTopology{}
	)

	catalogSvc.SetBackendLoader(integrationSvc)
	catalogSvc.SetArtifactParser(runtimeSvc)
	catalogSvc.SetDynamicManifestLoader(runtimeSvc)
	catalogSvc.SetNodeStateSyncer(runtimeSvc)
	catalogSvc.SetMenuSyncer(integrationSvc)
	catalogSvc.SetResourceRefSyncer(integrationSvc)
	catalogSvc.SetReleaseStateSyncer(runtimeSvc)
	catalogSvc.SetHookDispatcher(integrationSvc)

	lifecycleSvc.SetReconciler(runtimeSvc)
	lifecycleSvc.SetTopology(topology)

	integrationSvc.SetBizCtxProvider(&bizCtxAdapter{svc: bizCtxProvider})
	integrationSvc.SetDynamicCronExecutor(runtimeSvc)
	integrationSvc.SetHostServices(newTestHostServices())
	integrationSvc.SetTopologyProvider(topology)

	runtimeSvc.SetMenuManager(integrationSvc)
	runtimeSvc.SetHookDispatcher(integrationSvc)
	runtimeSvc.SetPermissionMenuFilter(integrationSvc)
	runtimeSvc.SetJwtConfigProvider(&jwtConfigAdapter{svc: configProvider})
	runtimeSvc.SetUploadSizeProvider(&uploadSizeAdapter{svc: configProvider})
	runtimeSvc.SetUserContextSetter(&userCtxAdapter{svc: bizCtxProvider})
	runtimeSvc.SetTopology(topology)

	return &Services{
		Catalog:     catalogSvc,
		Lifecycle:   lifecycleSvc,
		Runtime:     runtimeSvc,
		Frontend:    frontendSvc,
		Integration: integrationSvc,
		OpenAPI:     openapiSvc,
	}
}

// testHostServices publishes the minimal host service directory needed by
// source-plugin callbacks exercised in plugin service tests.
type testHostServices struct {
	// configSvc exposes read-only GoFrame configuration to source plugins.
	configSvc contract.ConfigService
}

// Ensure testHostServices satisfies the source-plugin host service directory.
var _ pluginhost.HostServices = (*testHostServices)(nil)

// newTestHostServices creates a host service directory for integration tests.
func newTestHostServices() pluginhost.HostServices {
	return &testHostServices{
		configSvc: pluginserviceconfig.New(),
	}
}

// APIDoc returns no apidoc service for plugin integration tests.
func (s *testHostServices) APIDoc() contract.APIDocService { return nil }

// Auth returns no auth service for plugin integration tests.
func (s *testHostServices) Auth() contract.AuthService { return nil }

// BizCtx returns no bizctx service for plugin integration tests.
func (s *testHostServices) BizCtx() contract.BizCtxService { return nil }

// Config returns the test host configuration service.
func (s *testHostServices) Config() contract.ConfigService {
	if s == nil {
		return nil
	}
	return s.configSvc
}

// I18n returns no i18n service for plugin integration tests.
func (s *testHostServices) I18n() contract.I18nService { return nil }

// Notify returns no notification service for plugin integration tests.
func (s *testHostServices) Notify() contract.NotifyService { return nil }

// PluginState returns no plugin-state service for plugin integration tests.
func (s *testHostServices) PluginState() contract.PluginStateService { return nil }

// Route returns no route service for plugin integration tests.
func (s *testHostServices) Route() contract.RouteService { return nil }

// Session returns no session service for plugin integration tests.
func (s *testHostServices) Session() contract.SessionService { return nil }

// TenantFilter returns no tenant-filter service for plugin integration tests.
func (s *testHostServices) TenantFilter() contract.TenantFilterService { return nil }

// jwtConfigAdapter exposes config service JWT settings through the runtime test seam.
type jwtConfigAdapter struct {
	// svc provides JWT runtime configuration.
	svc configsvc.Service
}

// GetJwtSecret returns the configured JWT signing secret for test wiring.
func (a *jwtConfigAdapter) GetJwtSecret(ctx context.Context) string {
	return a.svc.GetJwtSecret(ctx)
}

// GetSessionTimeout returns the runtime-effective session timeout for test wiring.
func (a *jwtConfigAdapter) GetSessionTimeout(ctx context.Context) (time.Duration, error) {
	return a.svc.GetSessionTimeout(ctx)
}

// uploadSizeAdapter exposes upload-size config through the runtime test seam.
type uploadSizeAdapter struct {
	// svc provides upload-size runtime configuration.
	svc configsvc.Service
}

// GetUploadMaxSize returns the runtime-effective upload limit used in tests.
func (a *uploadSizeAdapter) GetUploadMaxSize(ctx context.Context) (int64, error) {
	return a.svc.GetUploadMaxSize(ctx)
}

// userCtxAdapter forwards authenticated user injection to the shared bizctx service.
type userCtxAdapter struct {
	// svc stores request-local user context.
	svc bizctx.Service
}

// SetUser injects authenticated user identity into the test request context.
func (a *userCtxAdapter) SetUser(ctx context.Context, tokenID string, userID int, username string, status int) {
	a.svc.SetUser(ctx, tokenID, userID, username, status)
}

// SetTenant injects the resolved tenant into the test request context.
func (a *userCtxAdapter) SetTenant(ctx context.Context, tenantID int) {
	a.svc.SetTenant(ctx, tenantID)
}

// SetUserAccess injects cached access-snapshot fields into the test request context.
func (a *userCtxAdapter) SetUserAccess(ctx context.Context, dataScope int, dataScopeUnsupported bool, unsupportedDataScope int) {
	a.svc.SetUserAccess(ctx, dataScope, dataScopeUnsupported, unsupportedDataScope)
}

// bizCtxAdapter exposes the current request user ID to integration-layer tests.
type bizCtxAdapter struct {
	// svc reads request-local user context.
	svc bizctx.Service
}

// GetUserId returns the current request user ID for integration-layer tests.
func (a *bizCtxAdapter) GetUserId(ctx context.Context) int {
	localCtx := a.svc.Get(ctx)
	if localCtx == nil {
		return 0
	}
	return localCtx.UserId
}

// GetDataScope returns the current request user's effective role data-scope.
func (a *bizCtxAdapter) GetDataScope(ctx context.Context) int {
	localCtx := a.svc.Get(ctx)
	if localCtx == nil {
		return 0
	}
	return localCtx.DataScope
}

// GetDataScopeUnsupported returns the unsupported data-scope state from the current request.
func (a *bizCtxAdapter) GetDataScopeUnsupported(ctx context.Context) (bool, int) {
	localCtx := a.svc.Get(ctx)
	if localCtx == nil {
		return false, 0
	}
	return localCtx.DataScopeUnsupported, localCtx.UnsupportedDataScope
}
