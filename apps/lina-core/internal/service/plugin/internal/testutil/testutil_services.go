// This file wires plugin sub-services and request-context adapters for tests.

package testutil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/container/gvar"

	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	configsvc "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/plugin/internal/capabilityowner"
	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/frontend"
	hostconfigadapter "lina-core/internal/service/plugin/internal/hostconfig"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/manifestresource"
	"lina-core/internal/service/plugin/internal/migration"
	"lina-core/internal/service/plugin/internal/openapi"
	"lina-core/internal/service/plugin/internal/pluginconfig"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/internal/service/plugin/internal/upgrade"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap"
	capabilityauthz "lina-core/pkg/plugin/capability/authcap/authz"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/capregistry"
	capabilitydictcap "lina-core/pkg/plugin/capability/dictcap"
	capabilityfilecap "lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/i18ncap"
	capabilityjobcap "lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	capabilityplugincap "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/plugin/pluginhost"
	"lina-core/pkg/statusflag"
)

// Services groups the wired plugin sub-services used by package-level tests.
type Services struct {
	// Catalog provides manifest discovery, registry, and release access.
	Catalog catalog.Service
	// Store owns plugin governance persistence and stable projections.
	Store store.Service
	// Lifecycle provides install and uninstall orchestration.
	Lifecycle lifecycle.Service
	// Migration executes plugin SQL lifecycle phases.
	Migration migration.Service
	// Runtime provides artifact parsing, reconcile, and route execution.
	Runtime runtime.Service
	// Frontend provides in-memory frontend bundle management.
	Frontend frontend.Service
	// Integration provides menu, hook, and resource-ref integration.
	Integration integration.Service
	// OpenAPI provides dynamic route OpenAPI projection.
	OpenAPI openapi.Service
	// Upgrade provides unified plugin upgrade status and preview orchestration.
	Upgrade upgrade.Service
}

// singleNodeTopology provides the default non-clustered topology for plugin tests.
type singleNodeTopology struct{}

// testConfigService wraps the real config service while routing dynamic plugin
// artifact discovery into the process-local test storage directory.
type testConfigService struct {
	configsvc.Service
	dynamicStoragePath string
}

// newTestConfigService creates the default plugin test config provider.
func newTestConfigService() configsvc.Service {
	return testConfigService{
		Service:            configsvc.New(),
		dynamicStoragePath: testDynamicStorageDir,
	}
}

// GetPlugin returns plugin config with the isolated dynamic storage path.
func (s testConfigService) GetPlugin(ctx context.Context) *configsvc.PluginConfig {
	cfg := s.Service.GetPlugin(ctx)
	if cfg == nil {
		cfg = &configsvc.PluginConfig{}
	}
	cfg.Dynamic.StoragePath = s.dynamicStoragePath
	return cfg
}

// GetPluginDynamicStoragePath returns the isolated dynamic storage path.
func (s testConfigService) GetPluginDynamicStoragePath(context.Context) string {
	return s.dynamicStoragePath
}

// Start accepts cluster startup calls without spawning coordination workers.
func (singleNodeTopology) Start(context.Context) {}

// Stop accepts cluster shutdown calls without owning external resources.
func (singleNodeTopology) Stop(context.Context) {}

// IsEnabled reports that package tests run in single-node mode.
func (singleNodeTopology) IsEnabled() bool {
	return false
}

// IsPrimary reports that the local test node owns primary-only work.
func (singleNodeTopology) IsPrimary() bool {
	return true
}

// NodeID returns the fixed node identifier used by package tests.
func (singleNodeTopology) NodeID() string {
	return "test-node"
}

// NewServices creates a fully wired plugin sub-service set for tests.
func NewServices() *Services {
	return newServicesWithInjected(nil, nil, nil)
}

// NewServicesWithCapabilities creates a test service graph with explicit source-plugin capabilities.
func NewServicesWithCapabilities(capabilities capability.Services) *Services {
	return newServicesWithInjected(capabilities, nil, nil)
}

// NewServicesWithRuntimeService creates a test service graph with an explicit dynamic runtime service.
func NewServicesWithRuntimeService(runtimeSvc runtime.Service) *Services {
	return newServicesWithInjected(nil, runtimeSvc, nil)
}

// NewServicesWithConfigService creates a test graph with an explicit runtime config service.
func NewServicesWithConfigService(configProvider configsvc.Service) *Services {
	return newServicesWithInjected(nil, nil, configProvider)
}

func newServicesWithInjected(
	injectedCapabilities capability.Services,
	injectedRuntimeSvc runtime.Service,
	injectedConfig configsvc.Service,
) *Services {
	configProvider := injectedConfig
	if configProvider == nil {
		configProvider = newTestConfigService()
	}
	var (
		bizCtxProvider  = bizctx.New()
		cacheCoordSvc   = cachecoord.Default(cachecoord.NewStaticTopology(false))
		i18nService     = i18nsvc.New(bizCtxProvider, configProvider, cacheCoordSvc)
		catalogSvc      = catalog.New(configProvider)
		topology        = singleNodeTopology{}
		storeSvc        = store.New(catalogSvc, topology)
		migrationSvc    = migration.New(catalogSvc, storeSvc)
		frontendSvc     = frontend.New(catalogSvc, storeSvc)
		openapiSvc      = openapi.New(catalogSvc, storeSvc, nil, i18nService)
		lockerSvc       = locker.New()
		sessionStore    = session.NewDBStore()
		capabilitySvc   = injectedCapabilities
		integrationHook = &testIntegrationDelegateProvider{}
		dependencySvc   = plugindep.New()
	)
	var (
		orgSvc    = orgspi.New(nil, nil, nil)
		tenantSvc = tenantspi.New(nil, nil, nil, bizCtxProvider)
		roleSvc   = role.New(integrationHook, bizCtxProvider, configProvider, i18nService, orgSvc, tenantSvc)
	)
	if capabilitySvc == nil {
		capabilitySvc = newTestCapabilities(bizCtxProvider)
	}
	wasmRuntime, err := wasm.NewRuntime(
		capabilitySvc,
		capregistry.NewRegistry(),
		pluginconfig.NewFactory("", ""),
		hostconfigadapter.NewStaticCapabilityAdapter(configProvider),
		manifestresource.NewFactory(""),
	)
	if err != nil {
		panic(fmt.Sprintf("create test wasm runtime: %v", err))
	}
	runtimeSvc := runtime.New(
		catalogSvc,
		storeSvc,
		migrationSvc,
		frontendSvc,
		i18nService,
		lockerSvc,
		topology,
		integrationHook,
		configProvider,
		bizCtxProvider,
		sessionStore,
		roleSvc,
		runtimeCacheChangeNotifier{},
		runtimeDependencyValidator{},
		capabilitySvc,
		wasmRuntime,
	)
	integrationRuntimeSvc := injectedRuntimeSvc
	if integrationRuntimeSvc == nil {
		integrationRuntimeSvc = runtimeSvc
	}
	integrationSvc := integration.New(
		catalogSvc,
		storeSvc,
		bizCtxProvider,
		topology,
		capabilitySvc,
		orgServiceForTestCapabilities(capabilitySvc, orgSvc),
		integrationRuntimeSvc,
	)
	integrationHook.service = integrationSvc
	lifecycleSvc := lifecycle.New(
		catalogSvc,
		storeSvc,
		runtimeSvc,
		integrationSvc,
		migrationSvc,
		dependencySvc,
		i18nService,
		runtimeCacheChangeNotifier{},
		topology,
		nil,
		capabilitySvc,
	)
	upgradeSvc, err := upgrade.New(
		catalogSvc,
		storeSvc,
		runtimeSvc,
		integrationSvc,
		migrationSvc,
		dependencySvc,
		i18nService,
		nil,
		runtimeCacheChangeNotifier{},
		runtimeCacheFreshener{},
		topology,
		configProvider,
	)
	if err != nil {
		panic(fmt.Sprintf("create test upgrade service: %v", err))
	}
	if err = lifecycleSvc.BindUpgrade(upgradeSvc); err != nil {
		panic(fmt.Sprintf("bind test upgrade service: %v", err))
	}

	return &Services{
		Catalog:     catalogSvc,
		Store:       storeSvc,
		Lifecycle:   lifecycleSvc,
		Migration:   migrationSvc,
		Runtime:     runtimeSvc,
		Frontend:    frontendSvc,
		Integration: integrationSvc,
		OpenAPI:     openapiSvc,
		Upgrade:     upgradeSvc,
	}
}

// runtimeCacheChangeNotifier is a no-op cache revision publisher for isolated
// plugin runtime tests.
type runtimeCacheChangeNotifier struct{}

// MarkRuntimeCacheChanged accepts one runtime cache change without publishing
// cross-node notifications in package tests.
func (runtimeCacheChangeNotifier) MarkRuntimeCacheChanged(context.Context, string) error {
	return nil
}

// PublishPluginChange accepts one plugin-scoped runtime cache change without
// publishing cross-node notifications in package tests.
func (runtimeCacheChangeNotifier) PublishPluginChange(context.Context, string, string, string) error {
	return nil
}

// SyncEnabledSnapshotAndPublishRuntimeChange accepts one local snapshot refresh
// request without publishing cross-node notifications in package tests.
func (runtimeCacheChangeNotifier) SyncEnabledSnapshotAndPublishRuntimeChange(context.Context, string, string) error {
	return nil
}

// runtimeCacheFreshener is a no-op cache freshness checker for isolated plugin tests.
type runtimeCacheFreshener struct{}

// EnsureRuntimeCacheFresh accepts runtime cache freshness checks in package tests.
func (runtimeCacheFreshener) EnsureRuntimeCacheFresh(context.Context) error {
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

// ValidateSourcePluginUpgradeCandidate accepts source upgrade candidates without
// consulting the root plugin facade in package tests.
func (runtimeDependencyValidator) ValidateSourcePluginUpgradeCandidate(context.Context, *catalog.Manifest) error {
	return nil
}

// testIntegrationDelegateProvider bridges runtime tests to integration side
// effects without duplicating production setter wiring.
type testIntegrationDelegateProvider struct {
	service integration.Service
}

// SyncPluginMenusAndPermissions delegates full menu and permission sync.
func (p *testIntegrationDelegateProvider) SyncPluginMenusAndPermissions(ctx context.Context, manifest *catalog.Manifest) error {
	if p == nil || p.service == nil {
		return nil
	}
	return p.service.SyncPluginMenusAndPermissions(ctx, manifest)
}

// SyncPluginMenus delegates manifest menu sync.
func (p *testIntegrationDelegateProvider) SyncPluginMenus(ctx context.Context, manifest *catalog.Manifest) error {
	if p == nil || p.service == nil {
		return nil
	}
	return p.service.SyncPluginMenus(ctx, manifest)
}

// DeletePluginMenusByManifest delegates plugin menu deletion.
func (p *testIntegrationDelegateProvider) DeletePluginMenusByManifest(ctx context.Context, manifest *catalog.Manifest) error {
	if p == nil || p.service == nil {
		return nil
	}
	return p.service.DeletePluginMenusByManifest(ctx, manifest)
}

// SyncPluginResourceReferences delegates resource reference synchronization.
func (p *testIntegrationDelegateProvider) SyncPluginResourceReferences(ctx context.Context, manifest *catalog.Manifest) error {
	if p == nil || p.service == nil {
		return nil
	}
	return p.service.SyncPluginResourceReferences(ctx, manifest)
}

// DispatchPluginHookEvent delegates lifecycle hook dispatch.
func (p *testIntegrationDelegateProvider) DispatchPluginHookEvent(
	ctx context.Context,
	event pluginhost.ExtensionPoint,
	values map[string]interface{},
) error {
	if p == nil || p.service == nil {
		return nil
	}
	return p.service.DispatchPluginHookEvent(ctx, event, values)
}

// FilterPermissionMenus delegates permission menu filtering.
func (p *testIntegrationDelegateProvider) FilterPermissionMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	if p == nil || p.service == nil {
		return menus
	}
	return p.service.FilterPermissionMenus(ctx, menus)
}

// CanExposeBusinessEntries delegates business entry visibility checks.
func (p *testIntegrationDelegateProvider) CanExposeBusinessEntries(ctx context.Context, pluginID string) bool {
	return p == nil || p.service == nil || p.service.CanExposeBusinessEntries(ctx, pluginID)
}

// orgServiceForTestCapabilities returns the organization service used by plugin
// resource data-scope tests while preserving fallback behavior for capability
// directories that intentionally omit Org().
func orgServiceForTestCapabilities(
	capabilities capability.Services,
	fallback capabilityorgcap.Service,
) capabilityorgcap.Service {
	var orgSvc capabilityorgcap.Service
	if capabilities != nil {
		orgSvc = capabilities.Org()
	}
	if orgSvc == nil {
		orgSvc = fallback
	}
	if orgSvc == nil {
		return nil
	}
	return orgSvc
}

// testCapabilities publishes the minimal capability services needed by
// source-plugin callbacks exercised in plugin service tests.
type testCapabilities struct {
	// authz exposes registration-safe authorization projections.
	authz capabilityauthz.Service
	// configFactory creates plugin-scoped configuration views.
	configFactory pluginconfig.Factory
	// dict exposes registration-safe dictionary projections.
	dict capabilitydictcap.Service
	// manifestFactory creates plugin-scoped manifest resource views.
	manifestFactory manifestresource.Factory
	// plugins exposes registration-safe plugin-governance projections.
	plugins capabilityplugincap.Service
	// bizCtx exposes request business-context projection to source plugins.
	bizCtx bizctxcap.Service
	// cache exposes a registration-safe no-op cache service to source plugins.
	cache cachecap.Service
	// lock exposes a registration-safe no-op lock service to source plugins.
	lock lockcap.Service
	// hostConfig exposes registration-safe host configuration defaults.
	hostConfig hostconfigcap.Service
	// tenantFilter exposes a registration-safe tenant filter.
	tenantFilter tenantcap.FilterService
	// users exposes a registration-safe user-domain capability.
	users capabilityusercap.Service
	// storage exposes a registration-safe no-op storage service to source plugins.
	storage storagecap.Service
	// pluginID scopes source-plugin capabilities when non-empty.
	pluginID string
}

// Ensure testCapabilities satisfies the ordinary capability services.
var _ capability.Services = (*testCapabilities)(nil)

// Ensure testCapabilities can return plugin-scoped capability views.
var _ capabilityowner.ScopedServicesFactory = (*testCapabilities)(nil)

// newTestCapabilities creates capability services for integration tests.
func newTestCapabilities(bizCtxSvc bizctxcap.Service) capability.Services {
	return &testCapabilities{
		authz:           testAuthzService{},
		configFactory:   pluginconfig.NewFactory("", ""),
		dict:            testDictService{},
		manifestFactory: manifestresource.NewFactory(""),
		plugins:         testPluginsService{},
		bizCtx:          bizCtxSvc,
		cache:           testCacheService{},
		lock:            testLockService{},
		hostConfig:      testHostConfigService{},
		tenantFilter:    testTenantFilterService{},
		users:           testUsersService{},
		storage:         newTestStorageService(),
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
		authz:           s.authz,
		configFactory:   s.configFactory,
		dict:            s.dict,
		manifestFactory: s.manifestFactory,
		plugins:         s.plugins,
		bizCtx:          s.bizCtx,
		cache:           s.cache,
		lock:            s.lock,
		hostConfig:      s.hostConfig,
		tenantFilter:    s.tenantFilter,
		users:           s.users,
		storage:         s.storage,
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

// Jobs returns an empty scheduled-job domain service for plugin integration tests.
func (s *testCapabilities) Jobs() capabilityjobcap.Service { return testNoopJobs{} }

// Lock returns the registration-safe lock service for plugin integration tests.
func (s *testCapabilities) Lock() lockcap.Service {
	if s == nil {
		return nil
	}
	return s.lock
}

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
	return orgspi.New(nil, nil, nil)
}

// Plugins returns the registration-safe plugin-governance domain service.
func (s *testCapabilities) Plugins() capabilityplugincap.Service {
	if s == nil {
		return nil
	}
	return s.plugins
}

// Route returns an empty dynamic-route metadata service for plugin integration tests.
func (s *testCapabilities) Route() routecap.Service { return testNoopRoute{} }

// Sessions returns an empty online-session domain service for plugin integration tests.
func (s *testCapabilities) Sessions() capabilitysessioncap.Service { return testNoopSessions{} }

// Storage returns the registration-safe storage service for plugin integration tests.
func (s *testCapabilities) Storage() storagecap.Service {
	if s == nil {
		return nil
	}
	return s.storage
}

// Tenant returns the default tenant capability fallback service.
func (s *testCapabilities) Tenant() tenantcap.Service {
	if s == nil {
		return tenantspi.New(nil, nil, nil, nil)
	}
	return testTenantService{
		Service: tenantspi.New(nil, nil, nil, nil),
		plugins: s.plugins,
		filter:  s.tenantFilter,
	}
}

// testTenantService attaches test tenant sub-capabilities to the fallback tenant service.
type testTenantService struct {
	tenantcap.Service
	plugins capabilityplugincap.Service
	filter  tenantcap.FilterService
}

// Plugins returns tenant-plugin governance through the plugin-domain fixture.
func (s testTenantService) Plugins() tenantcap.PluginService {
	if governance, ok := s.plugins.(tenantcap.PluginService); ok {
		return governance
	}
	return nil
}

// Filter returns the registration-safe tenant filter fixture.
func (s testTenantService) Filter() tenantcap.FilterService {
	return s.filter
}

// testHostConfigService returns deterministic host configuration values needed
// by registration-only source plugin callbacks.
type testHostConfigService struct{}

// Get returns a test log-retention value and nil for unrelated host keys.
func (testHostConfigService) Get(_ context.Context, key string, defaultValue any) (*gvar.Var, error) {
	if key == configsvc.RuntimeParamKeyLogRetentionDays {
		return gvar.New("30"), nil
	}
	if defaultValue != nil {
		return gvar.New(defaultValue), nil
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

// SysConfig returns a registration-only sys_config subresource.
func (testHostConfigService) SysConfig() hostconfigcap.SysConfigService {
	return testNoopSysConfig{}
}

// testAuthzService is an empty authorization fixture for registration-only tests.
type testAuthzService struct{}

// BatchGetPermissions returns label projections for non-empty permission keys.
func (testAuthzService) BatchGetPermissions(_ context.Context, keys []capabilityauthz.PermissionKey) (*capmodel.BatchResult[*capabilityauthz.PermissionInfo, capabilityauthz.PermissionKey], error) {
	result := &capmodel.BatchResult[*capabilityauthz.PermissionInfo, capabilityauthz.PermissionKey]{
		Items:      make(map[capabilityauthz.PermissionKey]*capabilityauthz.PermissionInfo, len(keys)),
		MissingIDs: []capabilityauthz.PermissionKey{},
	}
	for _, key := range keys {
		if key == "" {
			result.MissingIDs = append(result.MissingIDs, key)
			continue
		}
		result.Items[key] = &capabilityauthz.PermissionInfo{Key: key}
	}
	return result, nil
}

// BatchHasPermissions reports false for all permissions in registration-only tests.
func (testAuthzService) BatchHasPermissions(_ context.Context, keys []capabilityauthz.PermissionKey) (map[capabilityauthz.PermissionKey]bool, error) {
	result := make(map[capabilityauthz.PermissionKey]bool, len(keys))
	for _, key := range keys {
		result[key] = false
	}
	return result, nil
}

// HasPermission reports false because registration-only tests never authorize requests.
func (testAuthzService) HasPermission(context.Context, capabilityauthz.PermissionKey) (bool, error) {
	return false, nil
}

// IsPlatformAdmin reports false because registration-only tests never check admin status.
func (testAuthzService) IsPlatformAdmin(context.Context, capabilityauthz.UserID) (bool, error) {
	return false, nil
}

// ReplaceRolePermissions accepts permission replacement without mutating test state.
func (testAuthzService) ReplaceRolePermissions(context.Context, capabilityauthz.RoleID, []capabilityauthz.PermissionKey) error {
	return nil
}

// testCacheService is a no-op cache fixture for registration-only tests.
type testCacheService struct{}

// Get reports a cache miss because plugin integration tests do not persist cache data.
func (testCacheService) Get(context.Context, string, string) (*cachecap.CacheItem, bool, error) {
	return nil, false, nil
}

// GetMany reports all cache keys as missing.
func (testCacheService) GetMany(_ context.Context, in cachecap.GetManyInput) (*cachecap.GetManyOutput, error) {
	return &cachecap.GetManyOutput{
		Items:       map[string]*cachecap.CacheItem{},
		MissingKeys: append([]string(nil), in.Keys...),
	}, nil
}

// Set returns the stored projection without mutating shared cache state.
func (testCacheService) Set(_ context.Context, namespace string, key string, value string, _ time.Duration) (*cachecap.CacheItem, error) {
	return &cachecap.CacheItem{Key: namespace + ":" + key, ValueKind: cachecap.CacheValueKindString, Value: value}, nil
}

// SetMany returns stored projections without mutating shared cache state.
func (testCacheService) SetMany(_ context.Context, in cachecap.SetManyInput) (*cachecap.SetManyOutput, error) {
	output := &cachecap.SetManyOutput{Items: map[string]*cachecap.CacheItem{}}
	for _, item := range in.Items {
		output.Items[item.Key] = &cachecap.CacheItem{Key: in.Namespace + ":" + item.Key, ValueKind: cachecap.CacheValueKindString, Value: item.Value}
	}
	return output, nil
}

// Delete accepts cache deletion without touching shared state.
func (testCacheService) Delete(context.Context, string, string) error {
	return nil
}

// DeleteMany accepts cache deletion without touching shared state.
func (testCacheService) DeleteMany(context.Context, cachecap.DeleteManyInput) error {
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

// testLockService is a no-op lock fixture for registration-only tests.
type testLockService struct{}

// Acquire returns a successful isolated ticket without touching shared lock state.
func (testLockService) Acquire(_ context.Context, in lockcap.AcquireInput) (*lockcap.AcquireOutput, error) {
	expireAt := time.Now().Add(lockcap.DefaultLease)
	return &lockcap.AcquireOutput{Acquired: true, Ticket: "test-ticket:" + in.Name, ExpireAt: &expireAt}, nil
}

// Renew extends the isolated test ticket without touching shared lock state.
func (testLockService) Renew(context.Context, lockcap.RenewInput) (*lockcap.RenewOutput, error) {
	expireAt := time.Now().Add(lockcap.DefaultLease)
	return &lockcap.RenewOutput{ExpireAt: &expireAt}, nil
}

// Release accepts the isolated test ticket without touching shared lock state.
func (testLockService) Release(context.Context, lockcap.ReleaseInput) error {
	return nil
}

// testStorageService is an in-memory storage fixture used by dynamic runtime tests.
type testStorageService struct {
	mu      sync.Mutex
	objects map[string]*testStorageObject
}

type testStorageObject struct {
	body        []byte
	contentType string
}

func newTestStorageService() *testStorageService {
	return &testStorageService{objects: make(map[string]*testStorageObject)}
}

// Put stores one object in memory and returns plugin-visible metadata.
func (s *testStorageService) Put(_ context.Context, in storagecap.PutInput) (*storagecap.PutOutput, error) {
	body, err := io.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureObjects()
	s.objects[in.Path] = &testStorageObject{
		body:        append([]byte(nil), body...),
		contentType: strings.TrimSpace(in.ContentType),
	}
	return &storagecap.PutOutput{Object: s.objectLocked(in.Path)}, nil
}

// Get reads one object from memory.
func (s *testStorageService) Get(_ context.Context, in storagecap.GetInput) (*storagecap.GetOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureObjects()
	object, ok := s.objects[in.Path]
	if !ok {
		return &storagecap.GetOutput{Found: false}, nil
	}
	return &storagecap.GetOutput{
		Object: s.objectLocked(in.Path),
		Body:   io.NopCloser(bytes.NewReader(append([]byte(nil), object.body...))),
		Found:  true,
	}, nil
}

// Delete removes one object from memory.
func (s *testStorageService) Delete(_ context.Context, in storagecap.DeleteInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureObjects()
	delete(s.objects, in.Path)
	return nil
}

// DeleteMany removes explicit objects from memory.
func (s *testStorageService) DeleteMany(ctx context.Context, in storagecap.DeleteManyInput) error {
	for _, path := range in.Paths {
		if err := s.Delete(ctx, storagecap.DeleteInput{Path: path}); err != nil {
			return err
		}
	}
	return nil
}

// List returns a bounded in-memory object list.
func (s *testStorageService) List(_ context.Context, in storagecap.ListInput) (*storagecap.ListOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureObjects()
	limit := in.Limit
	if limit <= 0 {
		limit = storagecap.DefaultListLimit
	}
	if limit > storagecap.MaxListLimit {
		limit = storagecap.MaxListLimit
	}
	prefix := strings.TrimSuffix(strings.TrimSpace(in.Prefix), "/")
	keys := make([]string, 0, len(s.objects))
	for key := range s.objects {
		if key == prefix || strings.HasPrefix(key, prefix+"/") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if len(keys) > limit {
		keys = keys[:limit]
	}
	objects := make([]*storagecap.Object, 0, len(keys))
	for _, key := range keys {
		objects = append(objects, s.objectLocked(key))
	}
	return &storagecap.ListOutput{Objects: objects, Limit: limit}, nil
}

// ListCursor returns a bounded in-memory object list with cursor pagination.
func (s *testStorageService) ListCursor(_ context.Context, in storagecap.ListCursorInput) (*storagecap.ListCursorOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureObjects()
	limit := in.Limit
	if limit <= 0 {
		limit = storagecap.DefaultListLimit
	}
	if limit > storagecap.MaxListLimit {
		limit = storagecap.MaxListLimit
	}
	prefix := strings.TrimSuffix(strings.TrimSpace(in.Prefix), "/")
	keys := make([]string, 0, len(s.objects))
	for key := range s.objects {
		if key == prefix || strings.HasPrefix(key, prefix+"/") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	cursor := strings.TrimSpace(in.Cursor)
	filtered := keys[:0]
	for _, key := range keys {
		if cursor == "" || key > cursor {
			filtered = append(filtered, key)
		}
	}
	nextCursor := ""
	if len(filtered) > limit {
		nextCursor = filtered[limit-1]
		filtered = filtered[:limit]
	}
	objects := make([]*storagecap.Object, 0, len(filtered))
	for _, key := range filtered {
		objects = append(objects, s.objectLocked(key))
	}
	return &storagecap.ListCursorOutput{Objects: objects, NextCursor: nextCursor, Limit: limit}, nil
}

// Stat reads one in-memory object metadata snapshot.
func (s *testStorageService) Stat(_ context.Context, in storagecap.StatInput) (*storagecap.StatOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureObjects()
	if _, ok := s.objects[in.Path]; !ok {
		return &storagecap.StatOutput{Found: false}, nil
	}
	return &storagecap.StatOutput{Object: s.objectLocked(in.Path), Found: true}, nil
}

// BatchStat reads in-memory object metadata for explicit paths.
func (s *testStorageService) BatchStat(_ context.Context, in storagecap.BatchStatInput) (*storagecap.BatchStatOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureObjects()
	output := &storagecap.BatchStatOutput{Objects: []*storagecap.Object{}}
	for _, path := range in.Paths {
		if _, ok := s.objects[path]; !ok {
			output.MissingPaths = append(output.MissingPaths, path)
			continue
		}
		output.Objects = append(output.Objects, s.objectLocked(path))
	}
	return output, nil
}

// ProviderStatuses returns no provider diagnostics for registration-only tests.
func (*testStorageService) ProviderStatuses(context.Context) ([]*storagecap.ProviderStatus, error) {
	return []*storagecap.ProviderStatus{}, nil
}

func (s *testStorageService) ensureObjects() {
	if s.objects == nil {
		s.objects = make(map[string]*testStorageObject)
	}
}

func (s *testStorageService) objectLocked(path string) *storagecap.Object {
	object := s.objects[path]
	if object == nil {
		return nil
	}
	return &storagecap.Object{
		Path:        path,
		Size:        int64(len(object.body)),
		ContentType: object.contentType,
		Visibility:  storagecap.VisibilityPrivate,
	}
}

// testDictService is an empty dictionary fixture for registration-only tests.
type testDictService struct{}

// Type returns dictionary type methods for registration-only tests.
func (s testDictService) Type() capabilitydictcap.TypeService { return testDictTypeService{} }

// Value returns dictionary value methods for registration-only tests.
func (s testDictService) Value() capabilitydictcap.ValueService { return testDictValueService{} }

// Refresh accepts dictionary refresh requests without mutating cache state.
func (testDictService) Refresh(context.Context, capabilitydictcap.Type) error {
	return nil
}

type testDictTypeService struct{}

func (testDictTypeService) Get(context.Context, int) (*capabilitydictcap.TypeInfo, error) {
	return nil, nil
}

func (testDictTypeService) BatchGet(context.Context, []int) (*capmodel.BatchResult[*capabilitydictcap.TypeInfo, int], error) {
	return &capmodel.BatchResult[*capabilitydictcap.TypeInfo, int]{Items: map[int]*capabilitydictcap.TypeInfo{}}, nil
}

func (testDictTypeService) List(context.Context, capabilitydictcap.ListTypesInput) (*capmodel.PageResult[*capabilitydictcap.TypeInfo], error) {
	return &capmodel.PageResult[*capabilitydictcap.TypeInfo]{Items: []*capabilitydictcap.TypeInfo{}}, nil
}

func (testDictTypeService) EnsureVisible(context.Context, []int) error {
	return nil
}

func (testDictTypeService) EnsureKeysVisible(context.Context, []capabilitydictcap.Type) error {
	return nil
}

func (testDictTypeService) Create(context.Context, capabilitydictcap.CreateTypeInput) (int, error) {
	return 0, nil
}

func (testDictTypeService) Update(context.Context, capabilitydictcap.UpdateTypeInput) error {
	return nil
}

func (testDictTypeService) Delete(context.Context, int) error { return nil }

type testDictValueService struct{}

func (testDictValueService) Get(context.Context, int) (*capabilitydictcap.ValueInfo, error) {
	return nil, nil
}

func (testDictValueService) BatchGet(context.Context, capabilitydictcap.BatchGetValuesInput) (*capmodel.BatchResult[*capabilitydictcap.ValueInfo, capabilitydictcap.Value], error) {
	return &capmodel.BatchResult[*capabilitydictcap.ValueInfo, capabilitydictcap.Value]{Items: map[capabilitydictcap.Value]*capabilitydictcap.ValueInfo{}}, nil
}

// ResolveLabels returns deterministic label projections for requested values.
func (testDictValueService) ResolveLabels(_ context.Context, input capabilitydictcap.ResolveInput) (*capmodel.BatchResult[*capabilitydictcap.LabelInfo, capabilitydictcap.Value], error) {
	result := &capmodel.BatchResult[*capabilitydictcap.LabelInfo, capabilitydictcap.Value]{
		Items:      make(map[capabilitydictcap.Value]*capabilitydictcap.LabelInfo, len(input.Values)),
		MissingIDs: []capabilitydictcap.Value{},
	}
	for _, value := range input.Values {
		if value == "" {
			result.MissingIDs = append(result.MissingIDs, value)
			continue
		}
		result.Items[value] = &capabilitydictcap.LabelInfo{
			Type:  input.Type,
			Value: value,
			Label: string(value),
		}
	}
	return result, nil
}

// List returns an empty dictionary page for registration-only tests.
func (testDictValueService) List(context.Context, capabilitydictcap.ListValuesInput) (*capmodel.PageResult[*capabilitydictcap.ValueInfo], error) {
	return &capmodel.PageResult[*capabilitydictcap.ValueInfo]{Items: []*capabilitydictcap.ValueInfo{}}, nil
}

func (testDictValueService) EnsureVisible(context.Context, []int) error {
	return nil
}

// EnsureValuesVisible accepts values because registration-only tests never persist dictionary data.
func (testDictValueService) EnsureValuesVisible(context.Context, capabilitydictcap.ResolveInput) error {
	return nil
}

func (testDictValueService) Create(context.Context, capabilitydictcap.CreateValueInput) (int, error) {
	return 0, nil
}

func (testDictValueService) Update(context.Context, capabilitydictcap.UpdateValueInput) error {
	return nil
}

func (testDictValueService) Delete(context.Context, int) error {
	return nil
}

func (testDictValueService) DeleteByType(context.Context, capabilitydictcap.Type) error {
	return nil
}

// testTenantFilterService is a no-op tenant filter for registration-only tests.
type testTenantFilterService struct{}

// Context returns a platform-bypass tenant context for registration-only tests.
func (testTenantFilterService) Context(context.Context) tenantspi.TenantFilterContext {
	return tenantspi.TenantFilterContext{PlatformBypass: true}
}

// testPluginsService is an empty plugin-governance fixture for registration-only tests.
type testPluginsService struct{}

// BatchGet returns all requested plugin IDs as opaque missing records.
func (testPluginsService) BatchGet(_ context.Context, ids []capabilityplugincap.PluginID) (*capmodel.BatchResult[*capabilityplugincap.PluginInfo, capabilityplugincap.PluginID], error) {
	return &capmodel.BatchResult[*capabilityplugincap.PluginInfo, capabilityplugincap.PluginID]{
		Items:      map[capabilityplugincap.PluginID]*capabilityplugincap.PluginInfo{},
		MissingIDs: append([]capabilityplugincap.PluginID(nil), ids...),
	}, nil
}

// Current returns no current plugin projection for registration-only tests.
func (testPluginsService) Current(context.Context) (*capabilityplugincap.PluginInfo, error) {
	return nil, nil
}

// Get returns no plugin projection for registration-only tests.
func (s testPluginsService) Get(ctx context.Context, id capabilityplugincap.PluginID) (*capabilityplugincap.PluginInfo, error) {
	result, err := s.BatchGet(ctx, []capabilityplugincap.PluginID{id})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[id], nil
}

// List returns an empty plugin-governance page for registration-only tests.
func (testPluginsService) List(context.Context, capabilityplugincap.ListInput) (*capmodel.PageResult[*capabilityplugincap.PluginInfo], error) {
	return &capmodel.PageResult[*capabilityplugincap.PluginInfo]{Items: []*capabilityplugincap.PluginInfo{}}, nil
}

// ListTenantPlugins returns an empty page for registration-only tests.
func (testPluginsService) ListTenantPlugins(context.Context, capabilityplugincap.TenantListInput) (*capmodel.PageResult[*capabilityplugincap.TenantPluginInfo], error) {
	return &capmodel.PageResult[*capabilityplugincap.TenantPluginInfo]{Items: []*capabilityplugincap.TenantPluginInfo{}}, nil
}

// IsEnabled reports disabled plugin state for registration-only tests.
func (testPluginsService) IsEnabled(context.Context, capabilityplugincap.PluginID) (bool, error) {
	return false, nil
}

// IsProviderEnabled reports disabled provider state for registration-only tests.
func (testPluginsService) IsProviderEnabled(context.Context, capabilityplugincap.PluginID) (bool, error) {
	return false, nil
}

// IsEnabledAuthoritative reports disabled authoritative state for registration-only tests.
func (testPluginsService) IsEnabledAuthoritative(context.Context, capabilityplugincap.PluginID) (bool, error) {
	return false, nil
}

// Config returns a blank plugin configuration reader for registration-only tests.
func (testPluginsService) Config() capabilityplugincap.ConfigService {
	return testPluginConfigService{}
}

// Registry returns the test registry projection service.
func (s testPluginsService) Registry() capabilityplugincap.RegistryService {
	return s
}

// State returns the test plugin enablement lookup projection.
func (s testPluginsService) State() capabilityplugincap.StateService {
	return s
}

// Lifecycle returns no-op lifecycle operations for registration-only tests.
func (testPluginsService) Lifecycle() capabilityplugincap.LifecycleService {
	return testNoopPluginLifecycle{}
}

// testPluginConfigService is a no-op plugin configuration reader for registration-only tests.
type testPluginConfigService struct{}

// Get reports that the requested plugin config key does not exist.
func (testPluginConfigService) Get(_ context.Context, _ string, defaultValue any) (*gvar.Var, error) {
	if defaultValue != nil {
		return gvar.New(defaultValue), nil
	}
	return nil, nil
}

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

// SetTenantPluginEnabled accepts enablement changes without mutating test state.
func (testPluginsService) SetTenantPluginEnabled(context.Context, capabilityplugincap.PluginID, bool) error {
	return nil
}

// ProvisionTenantPluginDefaults accepts tenant default provisioning without mutating test state.
func (testPluginsService) ProvisionTenantPluginDefaults(context.Context, capmodel.DomainID) error {
	return nil
}

// testUsersService is an empty user-domain fixture for registration-only tests.
type testUsersService struct{}

// Current returns no user because registration-only tests never authenticate.
func (testUsersService) Current(context.Context) (*capabilityusercap.UserInfo, error) {
	return nil, nil
}

// BatchGet returns all requested user IDs as opaque missing records.
func (testUsersService) BatchGet(_ context.Context, ids []capabilityusercap.UserID) (*capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.UserID], error) {
	return &capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.UserID]{
		Items:      map[capabilityusercap.UserID]*capabilityusercap.UserInfo{},
		MissingIDs: append([]capabilityusercap.UserID(nil), ids...),
	}, nil
}

// Get returns no user projection for registration-only tests.
func (s testUsersService) Get(ctx context.Context, id capabilityusercap.UserID) (*capabilityusercap.UserInfo, error) {
	result, err := s.BatchGet(ctx, []capabilityusercap.UserID{id})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[id], nil
}

// BatchResolve returns all requested user identifiers as opaque missing records.
func (testUsersService) BatchResolve(_ context.Context, input capabilityusercap.BatchResolveInput) (*capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.ResolveKey], error) {
	result := &capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.ResolveKey]{
		Items:      map[capabilityusercap.ResolveKey]*capabilityusercap.UserInfo{},
		MissingIDs: []capabilityusercap.ResolveKey{},
	}
	for _, id := range input.IDs {
		result.MissingIDs = append(result.MissingIDs, capabilityusercap.ResolveKey("id:"+string(id)))
	}
	for _, username := range input.Usernames {
		result.MissingIDs = append(result.MissingIDs, capabilityusercap.ResolveKey("username:"+username))
	}
	for _, contact := range input.Contacts {
		result.MissingIDs = append(result.MissingIDs, capabilityusercap.ResolveKey("contact:"+contact))
	}
	return result, nil
}

// List returns an empty page because registration-only tests never query users.
func (testUsersService) List(context.Context, capabilityusercap.ListInput) (*capmodel.PageResult[*capabilityusercap.UserInfo], error) {
	return &capmodel.PageResult[*capabilityusercap.UserInfo]{Items: []*capabilityusercap.UserInfo{}}, nil
}

// EnsureVisible accepts all users because registration-only tests never execute route handlers.
func (testUsersService) EnsureVisible(context.Context, []capabilityusercap.UserID) error {
	return nil
}

// Create accepts user creation without mutating test state.
func (testUsersService) Create(context.Context, capabilityusercap.CreateInput) (capabilityusercap.UserID, error) {
	return "", nil
}

// Update accepts user updates without mutating test state.
func (testUsersService) Update(context.Context, capabilityusercap.UpdateInput) error {
	return nil
}

// Delete accepts user deletion without mutating test state.
func (testUsersService) Delete(context.Context, capabilityusercap.UserID) error {
	return nil
}

// SetStatus accepts user status changes without mutating test state.
func (testUsersService) SetStatus(context.Context, capabilityusercap.UserID, statusflag.Enabled) error {
	return nil
}

// ResetPassword accepts password reset without mutating test state.
func (testUsersService) ResetPassword(context.Context, capabilityusercap.UserID, string) error {
	return nil
}

// Assignment returns user-role assignment operations.
func (testUsersService) Assignment() capabilityusercap.AssignmentService {
	return testUserAssignments{}
}

// testUserAssignments accepts role replacement without mutating test state.
type testUserAssignments struct{}

// ReplaceRoles accepts role replacement without mutating test state.
func (testUserAssignments) ReplaceRoles(context.Context, capabilityusercap.UserID, []int) error {
	return nil
}
