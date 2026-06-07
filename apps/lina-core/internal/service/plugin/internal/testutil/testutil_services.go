// This file wires plugin sub-services and request-context adapters for tests.

package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/hostlock"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/notify"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/openapi"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/capability"
	capabilityai "lina-core/pkg/plugin/capability/aicap"
	capabilityaitext "lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap"
	capabilityauthz "lina-core/pkg/plugin/capability/authcap/authz"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilityconfigcap "lina-core/pkg/plugin/capability/configcap"
	capabilitydictcap "lina-core/pkg/plugin/capability/dictcap"
	capabilityfilecap "lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	capabilityhostconfig "lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/i18ncap"
	capabilityinfracap "lina-core/pkg/plugin/capability/infracap"
	capabilityjobcap "lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	capabilitymanifest "lina-core/pkg/plugin/capability/manifestcap"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/plugincap"
	capabilityconfig "lina-core/pkg/plugin/capability/plugincap"
	capabilityplugincap "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/tenantcap"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/plugin/pluginhost"
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
		lockerSvc      = locker.New()
		runtimeSvc     = runtime.New(catalogSvc, lifecycleSvc, frontendSvc, openapiSvc, i18nService, lockerSvc)
		integrationSvc = integration.New(catalogSvc)
		topology       = singleNodeTopology{}
		kvCacheSvc     = kvcache.New()
		sessionStore   = session.NewDBStore()
		tenantSvc      = tenantcapsvc.New(nil, bizCtxProvider)
		notifySvc      = notify.New(tenantSvc)
		capabilitySvc  = newTestCapabilities(bizCtxProvider)
	)
	hostLockSvc := mustNewHostLockServiceForTest(lockerSvc)

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
	integrationSvc.SetCapabilities(capabilitySvc)
	integrationSvc.SetTopologyProvider(topology)

	runtimeSvc.SetMenuManager(integrationSvc)
	runtimeSvc.SetHookDispatcher(integrationSvc)
	runtimeSvc.SetPermissionMenuFilter(integrationSvc)
	runtimeSvc.SetJwtConfigProvider(&jwtConfigAdapter{svc: configProvider})
	runtimeSvc.SetUploadSizeProvider(&uploadSizeAdapter{svc: configProvider})
	runtimeSvc.SetUserContextSetter(&userCtxAdapter{svc: bizCtxProvider})
	runtimeSvc.SetTopology(topology)
	runtimeSvc.SetSessionStore(sessionStore)
	runtimeSvc.SetRuntimeCacheChangeNotifier(runtimeCacheChangeNotifier{})
	runtimeSvc.SetDependencyValidator(runtimeDependencyValidator{})

	mustConfigureWasmHostServicesForTest(
		kvCacheSvc,
		hostLockSvc,
		notifySvc,
		configProvider,
		capabilitySvc,
		capabilityconfig.NewConfigFactory("", ""),
		capabilityhostconfig.New(mustHostConfigRawReader(configProvider)),
		capabilitymanifest.NewFactory(""),
	)

	return &Services{
		Catalog:     catalogSvc,
		Lifecycle:   lifecycleSvc,
		Runtime:     runtimeSvc,
		Frontend:    frontendSvc,
		Integration: integrationSvc,
		OpenAPI:     openapiSvc,
	}
}

// mustHostConfigRawReader returns the raw host-config reader implemented by
// the shared test config service and panics if fixture wiring regresses.
func mustHostConfigRawReader(configProvider configsvc.Service) capabilityhostconfig.RawConfigReader {
	reader, ok := configProvider.(capabilityhostconfig.RawConfigReader)
	if !ok {
		panic("test config service does not support raw host config reads")
	}
	return reader
}

// runtimeCacheChangeNotifier is a no-op cache revision publisher for isolated
// plugin runtime tests.
type runtimeCacheChangeNotifier struct{}

// MarkRuntimeCacheChanged accepts one runtime cache change without publishing
// cross-node notifications in package tests.
func (runtimeCacheChangeNotifier) MarkRuntimeCacheChanged(context.Context, string) error {
	return nil
}

// runtimeDependencyValidator is a no-op dynamic dependency validator for
// isolated plugin runtime tests.
type runtimeDependencyValidator struct{}

// ValidateDynamicPluginCandidate accepts runtime test manifests without
// consulting the root plugin facade.
func (runtimeDependencyValidator) ValidateDynamicPluginCandidate(context.Context, *catalog.Manifest) error {
	return nil
}

// mustNewHostLockServiceForTest creates the host-lock dependency used by wasm
// bridge tests. A failure means the fixture wiring is invalid.
func mustNewHostLockServiceForTest(lockerSvc locker.Service) hostlock.Service {
	service, err := hostlock.New(lockerSvc)
	if err != nil {
		panic(fmt.Sprintf("configure test host lock service: %v", err))
	}
	return service
}

// mustConfigureWasmHostServicesForTest mirrors the HTTP startup host-service
// wiring so runtime package tests are self-contained and order independent.
func mustConfigureWasmHostServicesForTest(
	kvCacheSvc kvcache.Service,
	hostLockSvc hostlock.Service,
	notifySvc notify.Service,
	configProvider configsvc.Service,
	hostServices capability.Services,
	configFactory plugincap.ConfigServiceFactory,
	hostConfigSvc hostconfigcap.Service,
	manifestFactory manifestcap.ServiceFactory,
) {
	configure := []struct {
		name string
		fn   func() error
	}{
		{name: "cache", fn: func() error { return wasm.ConfigureCacheHostService(kvCacheSvc) }},
		{name: "lock", fn: func() error { return wasm.ConfigureLockHostService(hostLockSvc) }},
		{name: "notify", fn: func() error { return wasm.ConfigureNotifyHostService(notifySvc) }},
		{name: "storage", fn: func() error { return wasm.ConfigureStorageHostService(configProvider) }},
		{name: "ai", fn: func() error { return wasm.ConfigureAITextHostService(hostServices) }},
		{name: "org", fn: func() error { return wasm.ConfigureOrgHostService(hostServices) }},
		{name: "tenant", fn: func() error { return wasm.ConfigureTenantHostService(hostServices) }},
		{name: "config", fn: func() error { return wasm.ConfigureConfigHostService(configFactory) }},
		{name: "host config", fn: func() error { return wasm.ConfigureHostConfigService(hostConfigSvc) }},
		{name: "manifest", fn: func() error { return wasm.ConfigureManifestHostService(manifestFactory) }},
	}
	for _, item := range configure {
		if err := item.fn(); err != nil {
			panic(fmt.Sprintf("configure test wasm %s host service: %v", item.name, err))
		}
	}
}

// testCapabilities publishes the minimal capability services needed by
// source-plugin callbacks exercised in plugin service tests.
type testCapabilities struct {
	// admin exposes registration-safe management capability slices.
	admin capability.AdminServices
	// authz exposes registration-safe authorization projections.
	authz capabilityauthz.Service
	// configFactory creates plugin-scoped configuration views.
	configFactory plugincap.ConfigServiceFactory
	// dict exposes registration-safe dictionary projections.
	dict capabilitydictcap.Service
	// manifestFactory creates plugin-scoped manifest resource views.
	manifestFactory manifestcap.ServiceFactory
	// plugins exposes registration-safe plugin-governance projections.
	plugins capabilityplugincap.Service
	// bizCtx exposes request business-context projection to source plugins.
	bizCtx bizctxcap.Service
	// cache exposes a registration-safe no-op cache service to source plugins.
	cache cachecap.Service
	// hostConfig exposes registration-safe host configuration defaults.
	hostConfig hostconfigcap.Service
	// tenantFilter exposes a registration-safe tenant filter.
	tenantFilter tenantcap.PluginTableFilterService
	// users exposes a registration-safe user-domain capability.
	users capabilityusercap.Service
	// pluginID scopes source-plugin capabilities when non-empty.
	pluginID string
}

// Ensure testCapabilities satisfies the source-plugin capability services.
var _ pluginhost.Services = (*testCapabilities)(nil)

// Ensure testCapabilities can return plugin-scoped capability views.
var _ capability.ScopedServicesFactory = (*testCapabilities)(nil)

// newTestCapabilities creates capability services for integration tests.
func newTestCapabilities(bizCtxSvc bizctxcap.Service) capability.Services {
	return &testCapabilities{
		admin:           testAdminServices{},
		authz:           testAuthzService{},
		configFactory:   capabilityconfig.NewConfigFactory("", ""),
		dict:            testDictService{},
		manifestFactory: capabilitymanifest.NewFactory(""),
		plugins:         testPluginsService{},
		bizCtx:          bizCtxSvc,
		cache:           testCacheService{},
		hostConfig:      testHostConfigService{},
		tenantFilter:    testTenantFilterService{},
		users:           testUsersService{},
	}
}

// APIDoc returns a fallback apidoc service for plugin integration tests.
func (s *testCapabilities) APIDoc() apidoccap.Service { return testNoopAPIDoc{} }

// Auth returns a no-op auth namespace for plugin integration tests.
func (s *testCapabilities) Auth() authcap.Service {
	if s == nil {
		return authcap.New(testNoopAuth{}, nil)
	}
	return authcap.New(testNoopAuth{}, s.authz)
}

// Admin returns the registration-safe management directory for plugin integration tests.
func (s *testCapabilities) Admin() capability.AdminServices {
	if s == nil {
		return nil
	}
	return s.admin
}

// AI returns the default AI capability fallback namespace.
func (s *testCapabilities) AI() capabilityai.Service {
	return capabilityai.New(capabilityaitext.New(nil))
}

// Users returns an empty user-domain service for plugin integration tests.
func (s *testCapabilities) Users() capabilityusercap.Service {
	if s == nil {
		return nil
	}
	return s.users
}

// BizCtx returns the shared test business-context projection.
func (s *testCapabilities) BizCtx() bizctxcap.Service {
	if s == nil {
		return nil
	}
	return s.bizCtx
}

// Cache returns the registration-safe cache service for plugin integration tests.
func (s *testCapabilities) Cache() cachecap.Service {
	if s == nil {
		return nil
	}
	return s.cache
}

// PluginConfig returns the plugin-scoped test host configuration service.
func (s *testCapabilities) PluginConfig() plugincap.ConfigService {
	if s == nil || s.configFactory == nil {
		return nil
	}
	return s.configFactory.ForPlugin(s.pluginID)
}

// Config returns an empty runtime-config domain service for plugin integration tests.
func (s *testCapabilities) Config() capabilityconfigcap.Service {
	return testNoopRuntimeConfig{}
}

// Dict returns the registration-safe dictionary-domain service.
func (s *testCapabilities) Dict() capabilitydictcap.Service {
	if s == nil {
		return nil
	}
	return s.dict
}

// Files returns an empty file-domain service for plugin integration tests.
func (s *testCapabilities) Files() capabilityfilecap.Service { return testNoopFiles{} }

// ForPlugin returns a plugin-bound capability view for source-plugin callbacks.
func (s *testCapabilities) ForPlugin(pluginID string) capability.Services {
	if s == nil {
		return nil
	}
	return &testCapabilities{
		admin:           s.admin,
		authz:           s.authz,
		configFactory:   s.configFactory,
		dict:            s.dict,
		manifestFactory: s.manifestFactory,
		plugins:         s.plugins,
		bizCtx:          s.bizCtx,
		cache:           s.cache,
		hostConfig:      s.hostConfig,
		tenantFilter:    s.tenantFilter,
		users:           s.users,
		pluginID:        pluginID,
	}
}

// HostConfig returns the registration-safe host config service for plugin integration tests.
func (s *testCapabilities) HostConfig() hostconfigcap.Service {
	if s == nil {
		return nil
	}
	return s.hostConfig
}

// I18n returns a fallback i18n service for plugin integration tests.
func (s *testCapabilities) I18n() i18ncap.Service { return testNoopI18n{} }

// Infra returns an empty infrastructure-domain service for plugin integration tests.
func (s *testCapabilities) Infra() capabilityinfracap.Service { return testNoopInfra{} }

// Jobs returns an empty scheduled-job domain service for plugin integration tests.
func (s *testCapabilities) Jobs() capabilityjobcap.Service { return testNoopJobs{} }

// Manifest returns the plugin-scoped manifest service for plugin integration tests.
func (s *testCapabilities) Manifest() manifestcap.Service {
	if s == nil || s.manifestFactory == nil {
		return nil
	}
	return s.manifestFactory.ForPlugin(s.pluginID)
}

// Notifications returns an empty notification-domain service for plugin integration tests.
func (s *testCapabilities) Notifications() capabilitynotifycap.Service {
	return testNoopNotifications{}
}

// Org returns the default organization capability fallback service.
func (s *testCapabilities) Org() capabilityorgcap.Service {
	return capabilityorgcap.New(nil)
}

// Plugins returns the registration-safe plugin-governance domain service.
func (s *testCapabilities) Plugins() capabilityplugincap.Service {
	if s == nil {
		return nil
	}
	return s.plugins
}

// PluginLifecycle returns no-op lifecycle operations for plugin integration tests.
func (s *testCapabilities) PluginLifecycle() plugincap.LifecycleService {
	return testNoopPluginLifecycle{}
}

// PluginState returns a disabled-state service for plugin integration tests.
func (s *testCapabilities) PluginState() plugincap.StateService { return testNoopPluginState{} }

// Route returns an empty dynamic-route metadata service for plugin integration tests.
func (s *testCapabilities) Route() routecap.Service { return testNoopRoute{} }

// Sessions returns an empty online-session domain service for plugin integration tests.
func (s *testCapabilities) Sessions() capabilitysessioncap.Service { return testNoopSessions{} }

// TenantFilter returns the registration-safe tenant filter for plugin integration tests.
func (s *testCapabilities) TenantFilter() tenantcap.PluginTableFilterService {
	if s == nil {
		return nil
	}
	return s.tenantFilter
}

// Tenant returns the default tenant capability fallback service.
func (s *testCapabilities) Tenant() tenantcapsvc.Service {
	return tenantcapsvc.New(nil, nil)
}

// testHostConfigService returns deterministic host configuration values needed
// by registration-only source plugin callbacks.
type testHostConfigService struct{}

// Get returns a test log-retention value and nil for unrelated host keys.
func (testHostConfigService) Get(_ context.Context, key string) (*gvar.Var, error) {
	if key == configsvc.RuntimeParamKeyLogRetentionDays {
		return gvar.New("30"), nil
	}
	return nil, nil
}

// Exists reports that only the test log-retention key is available.
func (testHostConfigService) Exists(_ context.Context, key string) (bool, error) {
	return key == configsvc.RuntimeParamKeyLogRetentionDays, nil
}

// String returns the test log-retention value or the supplied default.
func (testHostConfigService) String(_ context.Context, key string, defaultValue string) (string, error) {
	if key == configsvc.RuntimeParamKeyLogRetentionDays {
		return "30", nil
	}
	return defaultValue, nil
}

// Bool returns the supplied default value for registration-only tests.
func (testHostConfigService) Bool(_ context.Context, _ string, defaultValue bool) (bool, error) {
	return defaultValue, nil
}

// Int returns the test log-retention value or the supplied default.
func (testHostConfigService) Int(_ context.Context, key string, defaultValue int) (int, error) {
	if key == configsvc.RuntimeParamKeyLogRetentionDays {
		return 30, nil
	}
	return defaultValue, nil
}

// Duration returns the supplied default value for registration-only tests.
func (testHostConfigService) Duration(_ context.Context, _ string, defaultValue time.Duration) (time.Duration, error) {
	return defaultValue, nil
}

// testAuthzService is an empty authorization fixture for registration-only tests.
type testAuthzService struct{}

// BatchGetPermissions returns label projections for non-empty permission keys.
func (testAuthzService) BatchGetPermissions(_ context.Context, _ capmodel.CapabilityContext, keys []capabilityauthz.PermissionKey) (*capmodel.BatchResult[*capabilityauthz.PermissionProjection, capabilityauthz.PermissionKey], error) {
	result := &capmodel.BatchResult[*capabilityauthz.PermissionProjection, capabilityauthz.PermissionKey]{
		Items:      make(map[capabilityauthz.PermissionKey]*capabilityauthz.PermissionProjection, len(keys)),
		MissingIDs: []capabilityauthz.PermissionKey{},
	}
	for _, key := range keys {
		if key == "" {
			result.MissingIDs = append(result.MissingIDs, key)
			continue
		}
		result.Items[key] = &capabilityauthz.PermissionProjection{Key: key}
	}
	return result, nil
}

// HasPermission reports false because registration-only tests never authorize requests.
func (testAuthzService) HasPermission(context.Context, capmodel.CapabilityContext, capabilityauthz.PermissionKey) (bool, error) {
	return false, nil
}

// IsPlatformAdmin reports false because registration-only tests never check admin status.
func (testAuthzService) IsPlatformAdmin(context.Context, capmodel.CapabilityContext, capabilityauthz.UserID) (bool, error) {
	return false, nil
}

// testCacheService is a no-op cache fixture for registration-only tests.
type testCacheService struct{}

// Get reports a cache miss because plugin integration tests do not persist cache data.
func (testCacheService) Get(context.Context, string, string) (*cachecap.CacheItem, bool, error) {
	return nil, false, nil
}

// Set returns the stored projection without mutating shared cache state.
func (testCacheService) Set(_ context.Context, namespace string, key string, value string, _ time.Duration) (*cachecap.CacheItem, error) {
	return &cachecap.CacheItem{Key: namespace + ":" + key, ValueKind: cachecap.CacheValueKindString, Value: value}, nil
}

// Delete accepts cache deletion without touching shared state.
func (testCacheService) Delete(context.Context, string, string) error {
	return nil
}

// Incr returns the requested delta as an isolated integer cache item.
func (testCacheService) Incr(_ context.Context, namespace string, key string, delta int64, _ time.Duration) (*cachecap.CacheItem, error) {
	return &cachecap.CacheItem{Key: namespace + ":" + key, ValueKind: cachecap.CacheValueKindInt, IntValue: delta}, nil
}

// Expire reports that no cache item existed to expire.
func (testCacheService) Expire(context.Context, string, string, time.Duration) (bool, *time.Time, error) {
	return false, nil, nil
}

// testDictService is an empty dictionary fixture for registration-only tests.
type testDictService struct{}

// ResolveLabels returns deterministic label projections for requested values.
func (testDictService) ResolveLabels(_ context.Context, _ capmodel.CapabilityContext, input capabilitydictcap.ResolveInput) (*capmodel.BatchResult[*capabilitydictcap.LabelProjection, capabilitydictcap.Value], error) {
	result := &capmodel.BatchResult[*capabilitydictcap.LabelProjection, capabilitydictcap.Value]{
		Items:      make(map[capabilitydictcap.Value]*capabilitydictcap.LabelProjection, len(input.Values)),
		MissingIDs: []capabilitydictcap.Value{},
	}
	for _, value := range input.Values {
		if value == "" {
			result.MissingIDs = append(result.MissingIDs, value)
			continue
		}
		result.Items[value] = &capabilitydictcap.LabelProjection{
			Type:  input.Type,
			Value: value,
			Label: string(value),
		}
	}
	return result, nil
}

// testTenantFilterService is a no-op tenant filter for registration-only tests.
type testTenantFilterService struct{}

// Context returns a platform-bypass tenant context for registration-only tests.
func (testTenantFilterService) Context(context.Context) tenantcap.TenantFilterContext {
	return tenantcap.TenantFilterContext{PlatformBypass: true}
}

// Apply returns the model unchanged because registration-only tests never query plugin tables.
func (testTenantFilterService) Apply(_ context.Context, model *gdb.Model, _ string) *gdb.Model {
	return model
}

// testPluginsService is an empty plugin-governance fixture for registration-only tests.
type testPluginsService struct{}

// BatchGetPlugins returns all requested plugin IDs as opaque missing records.
func (testPluginsService) BatchGetPlugins(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityplugincap.PluginID) (*capmodel.BatchResult[*capabilityplugincap.Projection, capabilityplugincap.PluginID], error) {
	return &capmodel.BatchResult[*capabilityplugincap.Projection, capabilityplugincap.PluginID]{
		Items:      map[capabilityplugincap.PluginID]*capabilityplugincap.Projection{},
		MissingIDs: append([]capabilityplugincap.PluginID(nil), ids...),
	}, nil
}

// ListTenantPlugins returns an empty page for registration-only tests.
func (testPluginsService) ListTenantPlugins(context.Context, capmodel.CapabilityContext) (*capmodel.PageResult[*capabilityplugincap.TenantProjection], error) {
	return &capmodel.PageResult[*capabilityplugincap.TenantProjection]{Items: []*capabilityplugincap.TenantProjection{}}, nil
}

// Config returns a blank plugin configuration reader for registration-only tests.
func (testPluginsService) Config() capabilityplugincap.ConfigService {
	return testPluginConfigService{}
}

// State returns a nil-backed plugin state reader for registration-only tests.
func (testPluginsService) State() capabilityplugincap.StateService {
	return capabilityplugincap.NewState(nil)
}

// Lifecycle returns a nil-backed lifecycle service for registration-only tests.
func (testPluginsService) Lifecycle() capabilityplugincap.LifecycleService {
	return capabilityplugincap.NewLifecycle(nil)
}

// Registry returns the test registry projection service.
func (s testPluginsService) Registry() capabilityplugincap.RegistryService {
	return s
}

// testPluginConfigService is a no-op plugin configuration reader for registration-only tests.
type testPluginConfigService struct{}

// Get reports that the requested plugin config key does not exist.
func (testPluginConfigService) Get(context.Context, string) (*gvar.Var, error) { return nil, nil }

// Exists reports that no plugin config keys exist.
func (testPluginConfigService) Exists(context.Context, string) (bool, error) { return false, nil }

// Scan leaves target unchanged because registration-only tests do not read plugin config sections.
func (testPluginConfigService) Scan(context.Context, string, any) error { return nil }

// String returns the supplied default value.
func (testPluginConfigService) String(_ context.Context, _ string, defaultValue string) (string, error) {
	return defaultValue, nil
}

// Bool returns the supplied default value.
func (testPluginConfigService) Bool(_ context.Context, _ string, defaultValue bool) (bool, error) {
	return defaultValue, nil
}

// Int returns the supplied default value.
func (testPluginConfigService) Int(_ context.Context, _ string, defaultValue int) (int, error) {
	return defaultValue, nil
}

// Duration returns the supplied default value.
func (testPluginConfigService) Duration(_ context.Context, _ string, defaultValue time.Duration) (time.Duration, error) {
	return defaultValue, nil
}

// testPluginAdminService is a no-op plugin-governance admin fixture.
type testPluginAdminService struct{}

// SetPluginEnabled accepts enablement changes without mutating test state.
func (testPluginAdminService) SetPluginEnabled(context.Context, capmodel.CapabilityContext, capabilityplugincap.PluginID, bool) error {
	return nil
}

// ProvisionTenantDefaults accepts tenant default provisioning without mutating test state.
func (testPluginAdminService) ProvisionTenantDefaults(context.Context, capmodel.CapabilityContext, capmodel.DomainID) error {
	return nil
}

// testAdminServices exposes only the admin slices needed by registration tests.
type testAdminServices struct{}

// Users returns no-op user management commands for registration-only tests.
func (testAdminServices) Users() capabilityusercap.AdminService { return testNoopUsers{} }

// Auth returns no-op authentication and authorization management commands for registration-only tests.
func (testAdminServices) Auth() authcap.AdminService { return authcap.NewAdmin(testNoopAuthz{}) }

// Dict returns no-op dictionary management commands for registration-only tests.
func (testAdminServices) Dict() capabilitydictcap.AdminService { return testNoopDict{} }

// Files returns no-op file management commands for registration-only tests.
func (testAdminServices) Files() capabilityfilecap.AdminService { return testNoopFiles{} }

// Sessions returns no-op session management commands for registration-only tests.
func (testAdminServices) Sessions() capabilitysessioncap.AdminService { return testNoopSessions{} }

// Config returns no-op runtime-config management commands for registration-only tests.
func (testAdminServices) Config() capabilityconfigcap.AdminService { return testNoopRuntimeConfig{} }

// Notifications returns no-op notification management commands for registration-only tests.
func (testAdminServices) Notifications() capabilitynotifycap.AdminService {
	return testNoopNotifications{}
}

// Plugins returns no-op plugin management commands for tenant route construction.
func (testAdminServices) Plugins() capabilityplugincap.AdminService {
	return testPluginAdminService{}
}

// Jobs returns no-op scheduled-job management commands for registration-only tests.
func (testAdminServices) Jobs() capabilityjobcap.AdminService { return testNoopJobs{} }

// Infra returns no-op infrastructure management commands for registration-only tests.
func (testAdminServices) Infra() capabilityinfracap.AdminService { return testNoopInfra{} }

// testUsersService is an empty user-domain fixture for registration-only tests.
type testUsersService struct{}

// BatchGetUsers returns all requested user IDs as opaque missing records.
func (testUsersService) BatchGetUsers(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityusercap.UserID) (*capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID], error) {
	return &capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID]{
		Items:      map[capabilityusercap.UserID]*capabilityusercap.UserProjection{},
		MissingIDs: append([]capabilityusercap.UserID(nil), ids...),
	}, nil
}

// SearchUsers returns an empty page because registration-only tests never query users.
func (testUsersService) SearchUsers(context.Context, capmodel.CapabilityContext, capabilityusercap.SearchInput) (*capmodel.PageResult[*capabilityusercap.UserProjection], error) {
	return &capmodel.PageResult[*capabilityusercap.UserProjection]{Items: []*capabilityusercap.UserProjection{}}, nil
}

// EnsureUsersVisible accepts all users because registration-only tests never execute route handlers.
func (testUsersService) EnsureUsersVisible(context.Context, capmodel.CapabilityContext, []capabilityusercap.UserID) error {
	return nil
}

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
func (a *userCtxAdapter) SetUser(ctx context.Context, tokenID string, userID int, username string, status int, clientType string) {
	a.svc.SetUser(ctx, tokenID, userID, username, status, clientType)
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
