// This file keeps root-package test bootstrap and shared helpers for plugin facade tests.

package plugin

import (
	"context"
	"encoding/base64"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
	"lina-core/internal/service/hostlock"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/locker"
	notifysvc "lina-core/internal/service/notify"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/session"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability"
	capabilityai "lina-core/pkg/plugin/capability/aicap"
	aitextsvc "lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap"
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
	orgcapsvc "lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/plugincap"
	capabilityconfig "lina-core/pkg/plugin/capability/plugincap"
	capabilityplugincap "lina-core/pkg/plugin/capability/plugincap"
	capabilitypluginlifecycle "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/tenantcap"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/plugin/pluginhost"
)

// newTestService constructs the root plugin facade with default single-node topology.
func newTestService() *serviceImpl {
	return newTestServiceWithTopology(nil)
}

// newTestServiceWithTopology constructs the root plugin facade with one explicit topology.
func newTestServiceWithTopology(topology Topology) *serviceImpl {
	var (
		configProvider = configsvc.New()
		bizCtxProvider = bizctx.New()
		cacheCoordSvc  = cachecoord.Default(cachecoord.NewStaticTopology(false))
	)
	if topology != nil && topology.IsEnabled() {
		coordSvc := coordination.NewMemory(nil)
		lockerSvc := locker.New()
		cachecoord.DefaultWithCoordination(topology, coordSvc)
		cacheCoordSvc = cachecoord.Default(topology)
		i18nSvc := i18nsvc.New(bizCtxProvider, configProvider, cacheCoordSvc)
		service, err := New(topology, configProvider, bizCtxProvider, cacheCoordSvc, i18nSvc, session.NewDBStore(), lockerSvc, coordSvc.Lock())
		if err != nil {
			panic(err)
		}
		serviceImpl := service.(*serviceImpl)
		tenantSvc := tenantcapsvc.New(serviceImpl, bizCtxProvider)
		capabilities := newRootTestCapabilities(bizCtxProvider, serviceImpl)
		serviceImpl.SetCapabilities(capabilities)
		serviceImpl.SetTenantStartupCapability(tenantSvc)
		serviceImpl.SetTenantProvisioningCapability(tenantSvc)
		configureRootWasmHostServicesForTest(configProvider, bizCtxProvider, capabilities, lockerSvc)
		return serviceImpl
	}
	lockerSvc := locker.New()
	i18nSvc := i18nsvc.New(bizCtxProvider, configProvider, cacheCoordSvc)
	service, err := New(topology, configProvider, bizCtxProvider, cacheCoordSvc, i18nSvc, session.NewDBStore(), lockerSvc, nil)
	if err != nil {
		panic(err)
	}
	serviceImpl := service.(*serviceImpl)
	tenantSvc := tenantcapsvc.New(serviceImpl, bizCtxProvider)
	capabilities := newRootTestCapabilities(bizCtxProvider, serviceImpl)
	serviceImpl.SetCapabilities(capabilities)
	serviceImpl.SetTenantStartupCapability(tenantSvc)
	serviceImpl.SetTenantProvisioningCapability(tenantSvc)
	configureRootWasmHostServicesForTest(configProvider, bizCtxProvider, capabilities, lockerSvc)
	return serviceImpl
}

// configureRootWasmHostServicesForTest mirrors HTTP startup host-service wiring
// for root plugin facade tests that construct plugin.Service directly.
func configureRootWasmHostServicesForTest(
	configProvider configsvc.Service,
	bizCtxProvider bizctx.Service,
	capabilities capability.Services,
	lockerSvc locker.Service,
) {
	hostLockSvc, err := hostlock.New(lockerSvc)
	if err != nil {
		panic(err)
	}
	if err = ConfigureWasmHostServices(
		kvcache.New(),
		hostLockSvc,
		notifysvc.New(tenantcapsvc.New(nil, bizCtxProvider)),
		configProvider,
		capabilities,
		capabilityconfig.NewConfigFactory("", ""),
		capabilityhostconfig.New(mustHostConfigRawReader(configProvider)),
		capabilitymanifest.NewFactory(""),
	); err != nil {
		panic(err)
	}
}

// mustHostConfigRawReader returns the raw host-config reader implemented by
// the root test config service and panics if the fixture wiring regresses.
func mustHostConfigRawReader(configProvider configsvc.Service) capabilityhostconfig.RawConfigReader {
	reader, ok := configProvider.(capabilityhostconfig.RawConfigReader)
	if !ok {
		panic("test config service does not support raw host config reads")
	}
	return reader
}

// rootTestCapabilities publishes the minimal host service directory required
// by root-package plugin facade tests. It mirrors the production capability
// wiring only for services used by provider construction and leaves unrelated
// capability surfaces at nil or neutral fallback values.
type rootTestCapabilities struct {
	// bizCtx exposes the request business-context projection to provider plugins.
	bizCtx bizctxcap.Service
	// pluginLifecycle exposes nil-tolerant lifecycle hooks to tenant provider code.
	pluginLifecycle plugincap.LifecycleService
	// admin exposes provider-construction management slices used by source plugins.
	admin capability.AdminServices
	// users exposes a registration-safe user-domain capability for providers.
	users capabilityusercap.Service
	// plugins exposes a registration-safe plugin-governance capability for providers.
	plugins capabilityplugincap.Service
}

// Ensure rootTestCapabilities satisfies the source-plugin host service directory.
var _ pluginhost.Services = (*rootTestCapabilities)(nil)

// Ensure rootTestCapabilities can return plugin-scoped capability views.
var _ capability.ScopedServicesFactory = (*rootTestCapabilities)(nil)

// newRootTestCapabilities creates the minimal capability directory used by root tests.
func newRootTestCapabilities(
	bizCtxProvider bizctx.Service,
	lifecycleRunner plugincap.LifecycleRunner,
) capability.Services {
	return &rootTestCapabilities{
		bizCtx:          bizCtxProvider,
		pluginLifecycle: capabilitypluginlifecycle.NewLifecycle(lifecycleRunner),
		admin:           rootNoopAdminCapabilities{},
		users:           rootNoopUsers{},
		plugins:         rootNoopPlugins{},
	}
}

// APIDoc returns no API-documentation service for root plugin facade tests.
func (s *rootTestCapabilities) APIDoc() apidoccap.Service { return nil }

// Auth returns no auth namespace for root plugin facade tests.
func (s *rootTestCapabilities) Auth() authcap.Service { return nil }

// Admin returns provider-construction management slices for root plugin facade tests.
func (s *rootTestCapabilities) Admin() capability.AdminServices {
	if s == nil {
		return nil
	}
	return s.admin
}

// AI returns the default AI capability fallback namespace.
func (s *rootTestCapabilities) AI() capabilityai.Service {
	return capabilityai.New(aitextsvc.New(nil))
}

// Users returns a registration-safe user-domain service for root plugin facade tests.
func (s *rootTestCapabilities) Users() capabilityusercap.Service {
	if s == nil {
		return nil
	}
	return s.users
}

// BizCtx returns the host business-context projection used by provider construction.
func (s *rootTestCapabilities) BizCtx() bizctxcap.Service {
	if s == nil {
		return nil
	}
	return s.bizCtx
}

// Cache returns no cache service for root plugin facade tests.
func (s *rootTestCapabilities) Cache() cachecap.Service { return nil }

// PluginConfig returns no plugin configuration service for root plugin facade tests.
func (s *rootTestCapabilities) PluginConfig() plugincap.ConfigService { return nil }

// Config returns no runtime-config domain service for root plugin facade tests.
func (s *rootTestCapabilities) Config() capabilityconfigcap.Service { return nil }

// Dict returns no dictionary-domain service for root plugin facade tests.
func (s *rootTestCapabilities) Dict() capabilitydictcap.Service { return nil }

// Files returns no file-domain service for root plugin facade tests.
func (s *rootTestCapabilities) Files() capabilityfilecap.Service { return nil }

// ForPlugin returns a plugin-bound capability view for provider construction.
func (s *rootTestCapabilities) ForPlugin(_ string) capability.Services {
	if s == nil {
		return nil
	}
	return &rootTestCapabilities{
		bizCtx:          s.bizCtx,
		pluginLifecycle: s.pluginLifecycle,
		admin:           s.admin,
		users:           s.users,
		plugins:         s.plugins,
	}
}

// HostConfig returns no host configuration service for root plugin facade tests.
func (s *rootTestCapabilities) HostConfig() hostconfigcap.Service { return nil }

// I18n returns no translation service for root plugin facade tests.
func (s *rootTestCapabilities) I18n() i18ncap.Service { return nil }

// Infra returns no infrastructure-domain service for root plugin facade tests.
func (s *rootTestCapabilities) Infra() capabilityinfracap.Service { return nil }

// Jobs returns no scheduled-job domain service for root plugin facade tests.
func (s *rootTestCapabilities) Jobs() capabilityjobcap.Service { return nil }

// Manifest returns no manifest resource service for root plugin facade tests.
func (s *rootTestCapabilities) Manifest() manifestcap.Service { return nil }

// Notifications returns no notification-domain service for root plugin facade tests.
func (s *rootTestCapabilities) Notifications() capabilitynotifycap.Service { return nil }

// Org returns the default organization capability fallback service.
func (s *rootTestCapabilities) Org() orgcapsvc.Service {
	return orgcapsvc.New(nil)
}

// Plugins returns a registration-safe plugin-governance service for root plugin facade tests.
func (s *rootTestCapabilities) Plugins() capabilityplugincap.Service {
	if s == nil {
		return nil
	}
	return s.plugins
}

// PluginLifecycle returns nil-tolerant lifecycle operations for tenant provider code.
func (s *rootTestCapabilities) PluginLifecycle() plugincap.LifecycleService {
	if s == nil {
		return nil
	}
	return s.pluginLifecycle
}

// PluginState returns no plugin-state service for root plugin facade tests.
func (s *rootTestCapabilities) PluginState() plugincap.StateService { return nil }

// Route returns no dynamic-route metadata service for root plugin facade tests.
func (s *rootTestCapabilities) Route() routecap.Service { return nil }

// Sessions returns no online-session domain service for root plugin facade tests.
func (s *rootTestCapabilities) Sessions() capabilitysessioncap.Service { return nil }

// Tenant returns the default tenant capability fallback service.
func (s *rootTestCapabilities) Tenant() tenantcapsvc.Service {
	if s == nil {
		return tenantcapsvc.New(nil, nil)
	}
	return tenantcapsvc.New(nil, s.bizCtx)
}

// TenantFilter returns no tenant-filter service for root plugin facade tests.
func (s *rootTestCapabilities) TenantFilter() tenantcap.PluginTableFilterService { return nil }

// rootNoopAdminCapabilities exposes the admin slices required by provider
// construction without mutating plugin, user, or authorization state.
type rootNoopAdminCapabilities struct{}

// Users returns no-op user management commands for provider construction.
func (rootNoopAdminCapabilities) Users() capabilityusercap.AdminService { return rootNoopUsers{} }

// Auth returns no authentication or authorization management commands for root facade tests.
func (rootNoopAdminCapabilities) Auth() authcap.AdminService { return nil }

// Dict returns no dictionary management commands for root facade tests.
func (rootNoopAdminCapabilities) Dict() capabilitydictcap.AdminService { return nil }

// Files returns no file management commands for root facade tests.
func (rootNoopAdminCapabilities) Files() capabilityfilecap.AdminService { return nil }

// Sessions returns no online-session management commands for root facade tests.
func (rootNoopAdminCapabilities) Sessions() capabilitysessioncap.AdminService { return nil }

// Config returns no runtime-config management commands for root facade tests.
func (rootNoopAdminCapabilities) Config() capabilityconfigcap.AdminService { return nil }

// Notifications returns no notification management commands for root facade tests.
func (rootNoopAdminCapabilities) Notifications() capabilitynotifycap.AdminService { return nil }

// Plugins returns no-op plugin management commands for provider construction.
func (rootNoopAdminCapabilities) Plugins() capabilityplugincap.AdminService { return rootNoopPlugins{} }

// Jobs returns no scheduled-job management commands for root facade tests.
func (rootNoopAdminCapabilities) Jobs() capabilityjobcap.AdminService { return nil }

// Infra returns no infrastructure management commands for root facade tests.
func (rootNoopAdminCapabilities) Infra() capabilityinfracap.AdminService { return nil }

// rootNoopUsers is a registration-safe user-domain fixture for root facade tests.
type rootNoopUsers struct{}

// BatchGetUsers reports all requested IDs as missing without querying storage.
func (rootNoopUsers) BatchGetUsers(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityusercap.UserID) (*capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID], error) {
	return &capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID]{
		Items:      map[capabilityusercap.UserID]*capabilityusercap.UserProjection{},
		MissingIDs: append([]capabilityusercap.UserID(nil), ids...),
	}, nil
}

// SearchUsers returns an empty bounded page for provider-construction paths.
func (rootNoopUsers) SearchUsers(context.Context, capmodel.CapabilityContext, capabilityusercap.SearchInput) (*capmodel.PageResult[*capabilityusercap.UserProjection], error) {
	return &capmodel.PageResult[*capabilityusercap.UserProjection]{Items: []*capabilityusercap.UserProjection{}}, nil
}

// EnsureUsersVisible accepts checks because root facade tests do not execute user business paths.
func (rootNoopUsers) EnsureUsersVisible(context.Context, capmodel.CapabilityContext, []capabilityusercap.UserID) error {
	return nil
}

// SetUserStatus accepts status changes without mutating shared test state.
func (rootNoopUsers) SetUserStatus(context.Context, capmodel.CapabilityContext, capabilityusercap.UserID, string) error {
	return nil
}

// rootNoopPlugins is a registration-safe plugin-governance fixture for root facade tests.
type rootNoopPlugins struct{}

// BatchGetPlugins reports all requested plugin IDs as missing projections.
func (rootNoopPlugins) BatchGetPlugins(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityplugincap.PluginID) (*capmodel.BatchResult[*capabilityplugincap.Projection, capabilityplugincap.PluginID], error) {
	return &capmodel.BatchResult[*capabilityplugincap.Projection, capabilityplugincap.PluginID]{
		Items:      map[capabilityplugincap.PluginID]*capabilityplugincap.Projection{},
		MissingIDs: append([]capabilityplugincap.PluginID(nil), ids...),
	}, nil
}

// ListTenantPlugins returns an empty tenant plugin page for construction-only tests.
func (rootNoopPlugins) ListTenantPlugins(context.Context, capmodel.CapabilityContext) (*capmodel.PageResult[*capabilityplugincap.TenantProjection], error) {
	return &capmodel.PageResult[*capabilityplugincap.TenantProjection]{Items: []*capabilityplugincap.TenantProjection{}}, nil
}

// Config returns a no-op plugin configuration reader for construction-only tests.
func (rootNoopPlugins) Config() capabilityplugincap.ConfigService {
	return rootNoopPluginConfig{}
}

// State returns a nil-backed plugin state reader for construction-only tests.
func (rootNoopPlugins) State() capabilityplugincap.StateService {
	return capabilityplugincap.NewState(nil)
}

// Lifecycle returns a nil-backed plugin lifecycle service for construction-only tests.
func (rootNoopPlugins) Lifecycle() capabilityplugincap.LifecycleService {
	return capabilityplugincap.NewLifecycle(nil)
}

// Registry returns the no-op plugin registry projection.
func (s rootNoopPlugins) Registry() capabilityplugincap.RegistryService {
	return s
}

// rootNoopPluginConfig is a no-op plugin config reader for root facade tests.
type rootNoopPluginConfig struct{}

// Get reports that the requested plugin config key does not exist.
func (rootNoopPluginConfig) Get(context.Context, string) (*gvar.Var, error) { return nil, nil }

// Exists reports that no plugin config keys exist.
func (rootNoopPluginConfig) Exists(context.Context, string) (bool, error) { return false, nil }

// Scan leaves target unchanged because root facade tests do not read plugin config sections.
func (rootNoopPluginConfig) Scan(context.Context, string, any) error { return nil }

// String returns the supplied default value.
func (rootNoopPluginConfig) String(_ context.Context, _ string, defaultValue string) (string, error) {
	return defaultValue, nil
}

// Bool returns the supplied default value.
func (rootNoopPluginConfig) Bool(_ context.Context, _ string, defaultValue bool) (bool, error) {
	return defaultValue, nil
}

// Int returns the supplied default value.
func (rootNoopPluginConfig) Int(_ context.Context, _ string, defaultValue int) (int, error) {
	return defaultValue, nil
}

// Duration returns the supplied default value.
func (rootNoopPluginConfig) Duration(_ context.Context, _ string, defaultValue time.Duration) (time.Duration, error) {
	return defaultValue, nil
}

// SetPluginEnabled accepts enablement changes without mutating shared test state.
func (rootNoopPlugins) SetPluginEnabled(context.Context, capmodel.CapabilityContext, capabilityplugincap.PluginID, bool) error {
	return nil
}

// ProvisionTenantDefaults accepts default provisioning without mutating test state.
func (rootNoopPlugins) ProvisionTenantDefaults(context.Context, capmodel.CapabilityContext, capmodel.DomainID) error {
	return nil
}

// TestNewRequiresExplicitRuntimeDependencies verifies the root plugin service
// returns a construction error when callers omit critical runtime dependencies.
func TestNewRequiresExplicitRuntimeDependencies(t *testing.T) {
	if _, err := New(nil, nil, nil, nil, nil, nil, nil, nil); err == nil {
		t.Fatal("expected plugin service construction to return an error without explicit dependencies")
	}
}

// getPluginRegistry loads one plugin registry row for assertions in root-package tests.
func (s *serviceImpl) getPluginRegistry(ctx context.Context, pluginID string) (*entity.SysPlugin, error) {
	return s.catalogSvc.GetRegistry(ctx, pluginID)
}

// getPluginRelease loads one persisted release row for assertions in root-package tests.
func (s *serviceImpl) getPluginRelease(ctx context.Context, pluginID string, version string) (*entity.SysPluginRelease, error) {
	return s.catalogSvc.GetRelease(ctx, pluginID, version)
}

// getActivePluginManifest resolves the currently active manifest for assertions in runtime tests.
func (s *serviceImpl) getActivePluginManifest(ctx context.Context, pluginID string) (*catalog.Manifest, error) {
	return s.catalogSvc.GetActiveManifest(ctx, pluginID)
}

// buildPluginGovernanceSnapshot delegates snapshot generation so tests can
// assert governance projection behavior through the facade wiring.
func (s *serviceImpl) buildPluginGovernanceSnapshot(
	ctx context.Context,
	pluginID string,
	version string,
	pluginType string,
	installed int,
	enabled int,
) (*catalog.GovernanceSnapshot, error) {
	return s.catalogSvc.BuildGovernanceSnapshot(ctx, pluginID, version, pluginType, installed, enabled)
}

// loadRuntimePluginManifestFromArtifact parses one runtime artifact into a manifest for tests.
func (s *serviceImpl) loadRuntimePluginManifestFromArtifact(artifactPath string) (*catalog.Manifest, error) {
	return s.catalogSvc.LoadManifestFromArtifactPath(artifactPath)
}

// syncPluginManifest persists one manifest into plugin governance storage for tests.
func (s *serviceImpl) syncPluginManifest(ctx context.Context, manifest *catalog.Manifest) (*entity.SysPlugin, error) {
	return s.catalogSvc.SyncManifest(ctx, manifest)
}

// setPluginInstalled updates the installed flag directly for test setup helpers.
func (s *serviceImpl) setPluginInstalled(ctx context.Context, pluginID string, installed int) error {
	return s.catalogSvc.SetPluginInstalled(ctx, pluginID, installed)
}

// setPluginStatus updates the enabled flag directly for test setup helpers.
func (s *serviceImpl) setPluginStatus(ctx context.Context, pluginID string, status int) error {
	return s.catalogSvc.SetPluginStatus(ctx, pluginID, status)
}

// executeDynamicRoute forwards one prepared bridge request to the runtime executor for tests.
func (s *serviceImpl) executeDynamicRoute(ctx context.Context, manifest *catalog.Manifest, request *protocol.BridgeRequestEnvelopeV1) (*protocol.BridgeResponseEnvelopeV1, error) {
	return s.runtimeSvc.ExecuteDynamicRoute(ctx, manifest, request)
}

// testTopology lets root-package tests simulate clustered primary/follower behavior.
type testTopology struct {
	mu      sync.RWMutex
	enabled bool
	primary bool
	nodeID  string
}

// IsEnabled reports whether the simulated topology should behave as clustered.
func (t *testTopology) IsEnabled() bool {
	if t == nil {
		return false
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.enabled
}

// IsPrimary reports whether the simulated node owns primary reconciliation duties.
func (t *testTopology) IsPrimary() bool {
	if t == nil {
		return true
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.primary
}

// NodeID returns the simulated node identifier used in cluster-state assertions.
func (t *testTopology) NodeID() string {
	if t == nil {
		return "test-node"
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.nodeID == "" {
		return "test-node"
	}
	return t.nodeID
}

// SetPrimary updates the simulated primary flag used by cluster-aware tests.
func (t *testTopology) SetPrimary(primary bool) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.primary = primary
}

// buildVersionedRuntimeFrontendAssets creates one marker-bearing asset set so
// upgrade tests can distinguish frontend content by release version.
func buildVersionedRuntimeFrontendAssets(marker string) []*catalog.ArtifactFrontendAsset {
	return []*catalog.ArtifactFrontendAsset{
		{
			Path:          "frontend/pages/index.html",
			ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>" + marker + "</body></html>")),
			ContentType:   "text/html; charset=utf-8",
		},
		{
			Path:          "frontend/pages/mount.js",
			ContentBase64: base64.StdEncoding.EncodeToString([]byte("export function mount() { return " + strconv.Quote(marker) + "; }")),
			ContentType:   "application/javascript",
		},
	}
}
