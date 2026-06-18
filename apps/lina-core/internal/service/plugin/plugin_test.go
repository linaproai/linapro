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

	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/internal/service/role"
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
	orgcapsvc "lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/plugincap"
	capabilityconfig "lina-core/pkg/plugin/capability/plugincap"
	capabilityplugincap "lina-core/pkg/plugin/capability/plugincap"
	capabilitypluginlifecycle "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/storagecap"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
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
	service, err := newTestServiceWithTopologyAndTenantDeps(topology, nil, nil, nil)
	if err != nil {
		panic(err)
	}
	return service
}

// newTestServiceWithTopologyAndTenantDeps constructs the root plugin facade
// with explicit tenant governance dependencies for tests that need to replace
// one startup-owned tenant slice.
func newTestServiceWithTopologyAndTenantDeps(
	topology Topology,
	tenantStartup pluginTenantStartupCapability,
	tenantProvisioning tenantspi.PluginProvisioningService,
	tenantGovernance platformGovernanceTenantCapability,
) (*serviceImpl, error) {
	var (
		configProvider = configsvc.New()
		bizCtxProvider = bizctx.New()
		cacheCoordSvc  = cachecoord.Default(cachecoord.NewStaticTopology(false))
		pluginRuntime  = NewRuntimeDelegate()
	)
	orgSvc := orgspi.New(nil, pluginRuntime)
	tenantSvc := tenantspi.New(nil, pluginRuntime, bizCtxProvider)
	if tenantStartup == nil {
		tenantStartup = tenantSvc
	}
	if tenantProvisioning == nil {
		tenantProvisioning = tenantSvc
	}
	if tenantGovernance == nil {
		tenantGovernance = tenantSvc
	}
	capabilities := newRootTestCapabilities(bizCtxProvider, pluginRuntime)
	if topology != nil && topology.IsEnabled() {
		coordSvc := coordination.NewMemory(nil)
		lockerSvc := locker.New()
		cachecoord.DefaultWithCoordination(topology, coordSvc)
		cacheCoordSvc = cachecoord.Default(topology)
		i18nSvc := i18nsvc.New(bizCtxProvider, configProvider, cacheCoordSvc)
		roleSvc := role.New(pluginRuntime, bizCtxProvider, configProvider, i18nSvc, orgSvc, tenantSvc)
		service, err := New(
			topology,
			configProvider,
			bizCtxProvider,
			cacheCoordSvc,
			i18nSvc,
			session.NewDBStore(),
			roleSvc,
			lockerSvc,
			coordSvc.Lock(),
			capabilities,
			orgSvc,
			tenantStartup,
			tenantProvisioning,
			tenantGovernance,
			capabilityconfig.NewConfigFactory("", ""),
			capabilityhostconfig.New(mustHostConfigRawReader(configProvider)),
			capabilitymanifest.NewFactory(""),
		)
		if err != nil {
			return nil, err
		}
		serviceImpl := service.(*serviceImpl)
		if err = pluginRuntime.BindService(service); err != nil {
			return nil, err
		}
		return serviceImpl, nil
	}
	lockerSvc := locker.New()
	i18nSvc := i18nsvc.New(bizCtxProvider, configProvider, cacheCoordSvc)
	roleSvc := role.New(pluginRuntime, bizCtxProvider, configProvider, i18nSvc, orgSvc, tenantSvc)
	service, err := New(
		topology,
		configProvider,
		bizCtxProvider,
		cacheCoordSvc,
		i18nSvc,
		session.NewDBStore(),
		roleSvc,
		lockerSvc,
		nil,
		capabilities,
		orgSvc,
		tenantStartup,
		tenantProvisioning,
		tenantGovernance,
		capabilityconfig.NewConfigFactory("", ""),
		capabilityhostconfig.New(mustHostConfigRawReader(configProvider)),
		capabilitymanifest.NewFactory(""),
	)
	if err != nil {
		return nil, err
	}
	serviceImpl := service.(*serviceImpl)
	if err = pluginRuntime.BindService(service); err != nil {
		return nil, err
	}
	return serviceImpl, nil
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
	// storage exposes a registration-safe no-op storage service for runtime cleanup.
	storage storagecap.Service
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
		storage:         rootNoopStorage{},
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
	return capabilityai.New(aitextsvc.New(nil, nil))
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
		storage:         s.storage,
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

// Lock returns no lock service for root plugin facade tests.
func (s *rootTestCapabilities) Lock() lockcap.Service { return nil }

// Manifest returns no manifest resource service for root plugin facade tests.
func (s *rootTestCapabilities) Manifest() manifestcap.Service { return nil }

// Notifications returns no notification-domain service for root plugin facade tests.
func (s *rootTestCapabilities) Notifications() capabilitynotifycap.Service { return nil }

// Org returns the default organization capability fallback service.
func (s *rootTestCapabilities) Org() orgcapsvc.Service {
	return orgspi.New(nil, nil)
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

// Storage returns a no-op object storage service for root plugin facade tests.
func (s *rootTestCapabilities) Storage() storagecap.Service {
	if s == nil {
		return nil
	}
	return s.storage
}

// Tenant returns the default tenant capability fallback service.
func (s *rootTestCapabilities) Tenant() tenantcapsvc.Service {
	if s == nil {
		return tenantspi.New(nil, nil, nil)
	}
	return tenantspi.New(nil, nil, s.bizCtx)
}

// TenantFilter returns no tenant-filter service for root plugin facade tests.
func (s *rootTestCapabilities) TenantFilter() tenantspi.PluginTableFilterService { return nil }

// rootNoopStorage is a registration-safe object-storage fixture for root facade tests.
type rootNoopStorage struct{}

// Put returns metadata for the requested object without storing bytes.
func (rootNoopStorage) Put(_ context.Context, in storagecap.PutInput) (*storagecap.PutOutput, error) {
	return &storagecap.PutOutput{Object: &storagecap.Object{Path: in.Path, Size: in.Size, ContentType: in.ContentType}}, nil
}

// Get reports that no root-test object exists.
func (rootNoopStorage) Get(context.Context, storagecap.GetInput) (*storagecap.GetOutput, error) {
	return &storagecap.GetOutput{Found: false}, nil
}

// Delete accepts deletion without touching shared state.
func (rootNoopStorage) Delete(context.Context, storagecap.DeleteInput) error {
	return nil
}

// DeleteMany accepts batch deletion without touching shared state.
func (rootNoopStorage) DeleteMany(context.Context, storagecap.DeleteManyInput) error {
	return nil
}

// List returns an empty bounded object list.
func (rootNoopStorage) List(_ context.Context, in storagecap.ListInput) (*storagecap.ListOutput, error) {
	return &storagecap.ListOutput{Objects: []*storagecap.Object{}, Limit: in.Limit}, nil
}

// ListCursor returns an empty bounded cursor object list.
func (rootNoopStorage) ListCursor(_ context.Context, in storagecap.ListCursorInput) (*storagecap.ListCursorOutput, error) {
	return &storagecap.ListCursorOutput{Objects: []*storagecap.Object{}, Limit: in.Limit}, nil
}

// Stat reports that no root-test object exists.
func (rootNoopStorage) Stat(context.Context, storagecap.StatInput) (*storagecap.StatOutput, error) {
	return &storagecap.StatOutput{Found: false}, nil
}

// BatchStat reports all root-test objects as missing.
func (rootNoopStorage) BatchStat(_ context.Context, in storagecap.BatchStatInput) (*storagecap.BatchStatOutput, error) {
	return &storagecap.BatchStatOutput{MissingPaths: append([]string(nil), in.Paths...)}, nil
}

// ProviderStatuses returns no provider diagnostics for root facade tests.
func (rootNoopStorage) ProviderStatuses(context.Context) ([]*storagecap.ProviderStatus, error) {
	return []*storagecap.ProviderStatus{}, nil
}

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

// HostConfig returns no runtime host-configuration management commands for root facade tests.
func (rootNoopAdminCapabilities) HostConfig() hostconfigcap.AdminService { return nil }

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

// Current returns no current user for provider-construction paths.
func (rootNoopUsers) Current(context.Context, capmodel.CapabilityContext) (*capabilityusercap.UserProjection, error) {
	return nil, nil
}

// BatchGet reports all requested IDs as missing without querying storage.
func (rootNoopUsers) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityusercap.UserID) (*capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID], error) {
	return &capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID]{
		Items:      map[capabilityusercap.UserID]*capabilityusercap.UserProjection{},
		MissingIDs: append([]capabilityusercap.UserID(nil), ids...),
	}, nil
}

// BatchResolve reports all requested identifiers as missing without querying storage.
func (rootNoopUsers) BatchResolve(_ context.Context, _ capmodel.CapabilityContext, input capabilityusercap.BatchResolveInput) (*capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.ResolveKey], error) {
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

// Search returns an empty bounded page for provider-construction paths.
func (rootNoopUsers) Search(context.Context, capmodel.CapabilityContext, capabilityusercap.SearchInput) (*capmodel.PageResult[*capabilityusercap.UserProjection], error) {
	return &capmodel.PageResult[*capabilityusercap.UserProjection]{Items: []*capabilityusercap.UserProjection{}}, nil
}

// EnsureVisible accepts checks because root facade tests do not execute user business paths.
func (rootNoopUsers) EnsureVisible(context.Context, capmodel.CapabilityContext, []capabilityusercap.UserID) error {
	return nil
}

// SetStatus accepts status changes without mutating shared test state.
func (rootNoopUsers) SetStatus(context.Context, capmodel.CapabilityContext, capabilityusercap.UserID, string) error {
	return nil
}

// rootNoopPlugins is a registration-safe plugin-governance fixture for root facade tests.
type rootNoopPlugins struct{}

// Current returns no current plugin projection for construction-only tests.
func (rootNoopPlugins) Current(context.Context, capmodel.CapabilityContext) (*capabilityplugincap.Projection, error) {
	return nil, nil
}

// BatchGet reports all requested plugin IDs as missing projections.
func (rootNoopPlugins) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityplugincap.PluginID) (*capmodel.BatchResult[*capabilityplugincap.Projection, capabilityplugincap.PluginID], error) {
	return &capmodel.BatchResult[*capabilityplugincap.Projection, capabilityplugincap.PluginID]{
		Items:      map[capabilityplugincap.PluginID]*capabilityplugincap.Projection{},
		MissingIDs: append([]capabilityplugincap.PluginID(nil), ids...),
	}, nil
}

// Search returns an empty plugin-governance page for construction-only tests.
func (rootNoopPlugins) Search(context.Context, capmodel.CapabilityContext, capabilityplugincap.SearchInput) (*capmodel.PageResult[*capabilityplugincap.Projection], error) {
	return &capmodel.PageResult[*capabilityplugincap.Projection]{Items: []*capabilityplugincap.Projection{}}, nil
}

// ListTenantPlugins returns an empty tenant plugin page for construction-only tests.
func (rootNoopPlugins) ListTenantPlugins(context.Context, capmodel.CapabilityContext, capabilityplugincap.TenantListInput) (*capmodel.PageResult[*capabilityplugincap.TenantProjection], error) {
	return &capmodel.PageResult[*capabilityplugincap.TenantProjection]{Items: []*capabilityplugincap.TenantProjection{}}, nil
}

// BatchGetCapabilityStatus returns unavailable status for requested capability keys.
func (rootNoopPlugins) BatchGetCapabilityStatus(_ context.Context, _ capmodel.CapabilityContext, keys []capabilityplugincap.CapabilityKey) (*capmodel.BatchResult[*capmodel.CapabilityStatus, capabilityplugincap.CapabilityKey], error) {
	result := &capmodel.BatchResult[*capmodel.CapabilityStatus, capabilityplugincap.CapabilityKey]{
		Items:      make(map[capabilityplugincap.CapabilityKey]*capmodel.CapabilityStatus, len(keys)),
		MissingIDs: []capabilityplugincap.CapabilityKey{},
	}
	for _, key := range keys {
		result.Items[key] = &capmodel.CapabilityStatus{Available: false, Reason: "test_no_provider"}
	}
	return result, nil
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

// SetEnabled accepts enablement changes without mutating shared test state.
func (rootNoopPlugins) SetEnabled(context.Context, capmodel.CapabilityContext, capabilityplugincap.PluginID, bool) error {
	return nil
}

// ProvisionTenantDefaults accepts default provisioning without mutating test state.
func (rootNoopPlugins) ProvisionTenantDefaults(context.Context, capmodel.CapabilityContext, capmodel.DomainID) error {
	return nil
}

// TestNewRequiresExplicitRuntimeDependencies verifies the root plugin service
// returns a construction error when callers omit critical runtime dependencies.
func TestNewRequiresExplicitRuntimeDependencies(t *testing.T) {
	if _, err := New(
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
		nil,
		nil,
		nil,
	); err == nil {
		t.Fatal("expected plugin service construction to return an error without explicit dependencies")
	}
}

// getPluginRegistry loads one plugin registry row for assertions in root-package tests.
func (s *serviceImpl) getPluginRegistry(ctx context.Context, pluginID string) (*store.PluginRecord, error) {
	return s.storeSvc.GetRegistry(ctx, pluginID)
}

// getPluginRelease loads one persisted release row for assertions in root-package tests.
func (s *serviceImpl) getPluginRelease(ctx context.Context, pluginID string, version string) (*store.ReleaseRecord, error) {
	return s.storeSvc.GetRelease(ctx, pluginID, version)
}

// getActivePluginManifest resolves the currently active manifest for assertions in runtime tests.
func (s *serviceImpl) getActivePluginManifest(ctx context.Context, pluginID string) (*catalog.Manifest, error) {
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	if registry != nil &&
		plugintypes.NormalizeType(registry.Type) == plugintypes.TypeDynamic &&
		registry.Installed == plugintypes.InstalledYes &&
		registry.ReleaseId > 0 {
		return s.runtimeSvc.LoadActiveDynamicPluginManifest(ctx, registry)
	}
	return s.catalogSvc.GetDesiredManifest(pluginID)
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
) (*store.GovernanceSnapshot, error) {
	return s.storeSvc.BuildGovernanceSnapshot(ctx, pluginID, version, pluginType, installed, enabled)
}

// loadRuntimePluginManifestFromArtifact parses one runtime artifact into a manifest for tests.
func (s *serviceImpl) loadRuntimePluginManifestFromArtifact(artifactPath string) (*catalog.Manifest, error) {
	return s.catalogSvc.LoadManifestFromArtifactPath(artifactPath)
}

// syncPluginManifest persists one manifest into plugin governance storage for tests.
func (s *serviceImpl) syncPluginManifest(ctx context.Context, manifest *catalog.Manifest) (*store.PluginRecord, error) {
	return s.storeSvc.SyncManifest(ctx, manifest)
}

// setPluginInstalled updates the installed flag directly for test setup helpers.
func (s *serviceImpl) setPluginInstalled(ctx context.Context, pluginID string, installed int) error {
	return s.storeSvc.SetPluginInstalled(ctx, pluginID, installed)
}

// setPluginStatus updates the enabled flag directly for test setup helpers.
func (s *serviceImpl) setPluginStatus(ctx context.Context, pluginID string, status int) error {
	return s.storeSvc.SetPluginStatus(ctx, pluginID, status)
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
