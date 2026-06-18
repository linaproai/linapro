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
	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	configsvc "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/migration"
	"lina-core/internal/service/plugin/internal/openapi"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/internal/service/plugin/internal/upgrade"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/internal/service/role"
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
	capabilitydictcap "lina-core/pkg/plugin/capability/dictcap"
	capabilityfilecap "lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	capabilityhostconfig "lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/i18ncap"
	capabilityinfracap "lina-core/pkg/plugin/capability/infracap"
	capabilityjobcap "lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	capabilitymanifest "lina-core/pkg/plugin/capability/manifestcap"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/plugincap"
	capabilityconfig "lina-core/pkg/plugin/capability/plugincap"
	capabilityplugincap "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/plugin/pluginhost"
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

// upgradeTopologyAdapter adapts test topology naming to upgrade.Topology.
type upgradeTopologyAdapter struct {
	topology singleNodeTopology
}

// IsEnabled reports whether clustered coordination is enabled for tests.
func (a upgradeTopologyAdapter) IsEnabled() bool {
	return a.topology.IsClusterModeEnabled()
}

// NodeID returns the stable test node identifier.
func (a upgradeTopologyAdapter) NodeID() string {
	return a.topology.CurrentNodeID()
}

// NewServices creates a fully wired plugin sub-service set for tests.
func NewServices() *Services {
	return newServicesWithInjected(nil, nil, nil)
}

// NewServicesWithCapabilities creates a test service graph with explicit source-plugin capabilities.
func NewServicesWithCapabilities(capabilities capability.Services) *Services {
	return newServicesWithInjected(capabilities, nil, nil)
}

// NewServicesWithDynamicJobExecutor creates a test service graph with an explicit dynamic job executor.
func NewServicesWithDynamicJobExecutor(executor integration.DynamicJobExecutor) *Services {
	return newServicesWithInjected(nil, executor, nil)
}

// NewServicesWithUploadSizeProvider creates a test graph with an explicit
// runtime upload-size provider.
func NewServicesWithUploadSizeProvider(uploadSize runtime.UploadSizeProvider) *Services {
	return newServicesWithInjected(nil, nil, uploadSize)
}

func newServicesWithInjected(
	injectedCapabilities capability.Services,
	injectedDynamicJobExecutor integration.DynamicJobExecutor,
	injectedUploadSize runtime.UploadSizeProvider,
) *Services {
	var (
		configProvider  = configsvc.New()
		bizCtxProvider  = bizctx.New()
		cacheCoordSvc   = cachecoord.Default(cachecoord.NewStaticTopology(false))
		i18nService     = i18nsvc.New(bizCtxProvider, configProvider, cacheCoordSvc)
		catalogSvc      = catalog.New(configProvider)
		topology        = singleNodeTopology{}
		storeSvc        = store.New(catalogSvc, topology)
		migrationSvc    = migration.New(catalogSvc, storeSvc)
		frontendSvc     = frontend.New(catalogSvc, storeSvc)
		openapiSvc      = openapi.New(catalogSvc, storeSvc, nil, openapi.DefaultLocaleBundleReader{})
		lockerSvc       = locker.New()
		sessionStore    = session.NewDBStore()
		capabilitySvc   = injectedCapabilities
		integrationHook = &testIntegrationDelegateProvider{}
		dependencySvc   = plugindep.New()
	)
	orgSvc := orgspi.New(nil, nil)
	tenantSvc := tenantspi.New(nil, nil, bizCtxProvider)
	roleSvc := role.New(integrationHook, bizCtxProvider, configProvider, i18nService, orgSvc, tenantSvc)
	if capabilitySvc == nil {
		capabilitySvc = newTestCapabilities(bizCtxProvider)
	}
	uploadSize := injectedUploadSize
	if uploadSize == nil {
		uploadSize = &uploadSizeAdapter{svc: configProvider}
	}
	wasmRuntime, err := wasm.NewRuntime(
		capabilitySvc,
		capabilityconfig.NewConfigFactory("", ""),
		capabilityhostconfig.New(mustHostConfigRawReader(configProvider)),
		capabilitymanifest.NewFactory(""),
	)
	if err != nil {
		panic(fmt.Sprintf("create test wasm runtime: %v", err))
	}
	runtimeSvc := runtime.New(
		catalogSvc,
		storeSvc,
		migrationSvc,
		frontendSvc,
		openapiSvc,
		i18nService,
		lockerSvc,
		topology,
		integrationHook,
		integrationHook,
		integrationHook,
		&jwtConfigAdapter{svc: configProvider},
		uploadSize,
		&userCtxAdapter{svc: bizCtxProvider},
		sessionStore,
		roleSvc,
		integrationHook,
		runtimeCacheChangeNotifier{},
		runtimeDependencyValidator{},
		testStorageCleanupProvider{capabilities: capabilitySvc},
		wasmRuntime,
	)
	dynamicJobExecutor := injectedDynamicJobExecutor
	if dynamicJobExecutor == nil {
		dynamicJobExecutor = runtimeSvc
	}
	integrationSvc := integration.New(
		catalogSvc,
		storeSvc,
		&bizCtxAdapter{svc: bizCtxProvider},
		topology,
		testSourceServicesProvider{capabilities: capabilitySvc},
		capabilitySvc.Org(),
		dynamicJobExecutor,
		integration.NewSharedState(),
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
		testSourceServicesProvider{capabilities: capabilitySvc},
	)
	upgradeSvc, err := upgrade.New(
		catalogSvc,
		storeSvc,
		lifecycleSvc,
		runtimeSvc,
		integrationSvc,
		migrationSvc,
		dependencySvc,
		i18nService,
		nil,
		runtimeCacheChangeNotifier{},
		runtimeCacheFreshener{},
		upgradeTopologyAdapter{topology: topology},
		configProvider,
	)
	if err != nil {
		panic(fmt.Sprintf("create test upgrade service: %v", err))
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

// testSourceServicesProvider returns source-plugin service directories scoped by plugin ID.
type testSourceServicesProvider struct {
	capabilities capability.Services
}

// SourceServicesForPlugin returns plugin-scoped services for source-plugin callbacks.
func (p testSourceServicesProvider) SourceServicesForPlugin(pluginID string) pluginhost.Services {
	if p.capabilities == nil {
		return nil
	}
	services := capability.ServicesForPlugin(p.capabilities, pluginID)
	sourceServices, _ := services.(pluginhost.Services)
	return sourceServices
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

// testStorageCleanupProvider returns the capability directory used by storage
// cleanup in runtime uninstall tests.
type testStorageCleanupProvider struct {
	capabilities capability.Services
}

// StorageCleanupServices returns the test capability directory.
func (p testStorageCleanupProvider) StorageCleanupServices() capability.Services {
	return p.capabilities
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
	// lock exposes a registration-safe no-op lock service to source plugins.
	lock lockcap.Service
	// hostConfig exposes registration-safe host configuration defaults.
	hostConfig hostconfigcap.Service
	// tenantFilter exposes a registration-safe tenant filter.
	tenantFilter tenantspi.PluginTableFilterService
	// users exposes a registration-safe user-domain capability.
	users capabilityusercap.Service
	// storage exposes a registration-safe no-op storage service to source plugins.
	storage storagecap.Service
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

// Admin returns the registration-safe management directory for plugin integration tests.
func (s *testCapabilities) Admin() capability.AdminServices {
	if s == nil {
		return nil
	}
	return s.admin
}

// AI returns the default AI capability fallback namespace.
func (s *testCapabilities) AI() capabilityai.Service {
	return capabilityai.New(capabilityaitext.New(nil, nil))
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

// Infra returns an empty infrastructure-domain service for plugin integration tests.
func (s *testCapabilities) Infra() capabilityinfracap.Service { return testNoopInfra{} }

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
	return orgspi.New(nil, nil)
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

// Storage returns the registration-safe storage service for plugin integration tests.
func (s *testCapabilities) Storage() storagecap.Service {
	if s == nil {
		return nil
	}
	return s.storage
}

// TenantFilter returns the registration-safe tenant filter for plugin integration tests.
func (s *testCapabilities) TenantFilter() tenantspi.PluginTableFilterService {
	if s == nil {
		return nil
	}
	return s.tenantFilter
}

// Tenant returns the default tenant capability fallback service.
func (s *testCapabilities) Tenant() tenantcap.Service {
	return tenantspi.New(nil, nil, nil)
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

// BatchHasPermissions reports false for all permissions in registration-only tests.
func (testAuthzService) BatchHasPermissions(_ context.Context, _ capmodel.CapabilityContext, keys []capabilityauthz.PermissionKey) (map[capabilityauthz.PermissionKey]bool, error) {
	result := make(map[capabilityauthz.PermissionKey]bool, len(keys))
	for _, key := range keys {
		result[key] = false
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

// ListValues returns an empty dictionary page for registration-only tests.
func (testDictService) ListValues(context.Context, capmodel.CapabilityContext, capabilitydictcap.ListValuesInput) (*capmodel.PageResult[*capabilitydictcap.LabelProjection], error) {
	return &capmodel.PageResult[*capabilitydictcap.LabelProjection]{Items: []*capabilitydictcap.LabelProjection{}}, nil
}

// EnsureValuesVisible accepts values because registration-only tests never persist dictionary data.
func (testDictService) EnsureValuesVisible(context.Context, capmodel.CapabilityContext, capabilitydictcap.ResolveInput) error {
	return nil
}

// testTenantFilterService is a no-op tenant filter for registration-only tests.
type testTenantFilterService struct{}

// Context returns a platform-bypass tenant context for registration-only tests.
func (testTenantFilterService) Context(context.Context) tenantspi.TenantFilterContext {
	return tenantspi.TenantFilterContext{PlatformBypass: true}
}

// Apply returns the model unchanged because registration-only tests never query plugin tables.
func (testTenantFilterService) Apply(_ context.Context, model *gdb.Model, _ string) *gdb.Model {
	return model
}

// testPluginsService is an empty plugin-governance fixture for registration-only tests.
type testPluginsService struct{}

// BatchGet returns all requested plugin IDs as opaque missing records.
func (testPluginsService) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityplugincap.PluginID) (*capmodel.BatchResult[*capabilityplugincap.Projection, capabilityplugincap.PluginID], error) {
	return &capmodel.BatchResult[*capabilityplugincap.Projection, capabilityplugincap.PluginID]{
		Items:      map[capabilityplugincap.PluginID]*capabilityplugincap.Projection{},
		MissingIDs: append([]capabilityplugincap.PluginID(nil), ids...),
	}, nil
}

// Current returns no current plugin projection for registration-only tests.
func (testPluginsService) Current(context.Context, capmodel.CapabilityContext) (*capabilityplugincap.Projection, error) {
	return nil, nil
}

// Search returns an empty plugin-governance page for registration-only tests.
func (testPluginsService) Search(context.Context, capmodel.CapabilityContext, capabilityplugincap.SearchInput) (*capmodel.PageResult[*capabilityplugincap.Projection], error) {
	return &capmodel.PageResult[*capabilityplugincap.Projection]{Items: []*capabilityplugincap.Projection{}}, nil
}

// ListTenantPlugins returns an empty page for registration-only tests.
func (testPluginsService) ListTenantPlugins(context.Context, capmodel.CapabilityContext, capabilityplugincap.TenantListInput) (*capmodel.PageResult[*capabilityplugincap.TenantProjection], error) {
	return &capmodel.PageResult[*capabilityplugincap.TenantProjection]{Items: []*capabilityplugincap.TenantProjection{}}, nil
}

// BatchGetCapabilityStatus returns all requested capability keys as unavailable.
func (testPluginsService) BatchGetCapabilityStatus(_ context.Context, _ capmodel.CapabilityContext, keys []capabilityplugincap.CapabilityKey) (*capmodel.BatchResult[*capmodel.CapabilityStatus, capabilityplugincap.CapabilityKey], error) {
	result := &capmodel.BatchResult[*capmodel.CapabilityStatus, capabilityplugincap.CapabilityKey]{
		Items:      make(map[capabilityplugincap.CapabilityKey]*capmodel.CapabilityStatus, len(keys)),
		MissingIDs: []capabilityplugincap.CapabilityKey{},
	}
	for _, key := range keys {
		result.Items[key] = &capmodel.CapabilityStatus{Available: false, Reason: "test_no_provider"}
	}
	return result, nil
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

// SetEnabled accepts enablement changes without mutating test state.
func (testPluginAdminService) SetEnabled(context.Context, capmodel.CapabilityContext, capabilityplugincap.PluginID, bool) error {
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

// HostConfig returns no-op runtime host-configuration management commands for registration-only tests.
func (testAdminServices) HostConfig() hostconfigcap.AdminService { return testNoopRuntimeConfig{} }

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

// Current returns no user because registration-only tests never authenticate.
func (testUsersService) Current(context.Context, capmodel.CapabilityContext) (*capabilityusercap.UserProjection, error) {
	return nil, nil
}

// BatchGet returns all requested user IDs as opaque missing records.
func (testUsersService) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityusercap.UserID) (*capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID], error) {
	return &capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID]{
		Items:      map[capabilityusercap.UserID]*capabilityusercap.UserProjection{},
		MissingIDs: append([]capabilityusercap.UserID(nil), ids...),
	}, nil
}

// BatchResolve returns all requested user identifiers as opaque missing records.
func (testUsersService) BatchResolve(_ context.Context, _ capmodel.CapabilityContext, input capabilityusercap.BatchResolveInput) (*capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.ResolveKey], error) {
	result := &capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.ResolveKey]{
		Items:      map[capabilityusercap.ResolveKey]*capabilityusercap.UserProjection{},
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

// Search returns an empty page because registration-only tests never query users.
func (testUsersService) Search(context.Context, capmodel.CapabilityContext, capabilityusercap.SearchInput) (*capmodel.PageResult[*capabilityusercap.UserProjection], error) {
	return &capmodel.PageResult[*capabilityusercap.UserProjection]{Items: []*capabilityusercap.UserProjection{}}, nil
}

// EnsureVisible accepts all users because registration-only tests never execute route handlers.
func (testUsersService) EnsureVisible(context.Context, capmodel.CapabilityContext, []capabilityusercap.UserID) error {
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
