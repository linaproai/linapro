// This file keeps root-package test bootstrap and shared helpers for plugin facade tests.

package plugin

import (
	"context"
	"encoding/base64"
	"go/ast"
	"go/parser"
	"go/token"
	pluginv1 "lina-core/api/plugin/v1"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"

	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cluster"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/plugin/internal/capabilityowner"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	plugintestutil "lina-core/internal/service/plugin/internal/testutil"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilitydictcap "lina-core/pkg/plugin/capability/dictcap"
	capabilityfilecap "lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/i18ncap"
	capabilityjobcap "lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
	orgcapsvc "lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/plugincap"
	capabilityplugincap "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/storagecap"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/statusflag"
)

// rootTestConfigService wraps the real config service with plugin-package test
// overrides for dynamic storage, startup auto-enable, and force uninstall.
type rootTestConfigService struct {
	configsvc.Service
}

var rootTestConfigOverrides = struct {
	sync.Mutex
	autoEnableSet       bool
	autoEnableEntries   []configsvc.PluginAutoEnableEntry
	allowForceUninstall *bool
}{}

// newRootTestConfigService creates the config provider used by root plugin tests.
func newRootTestConfigService() configsvc.Service {
	return rootTestConfigService{Service: configsvc.New()}
}

// GetPlugin returns plugin config with package-local test overrides applied.
func (s rootTestConfigService) GetPlugin(ctx context.Context) *configsvc.PluginConfig {
	cfg := s.Service.GetPlugin(ctx)
	if cfg == nil {
		cfg = &configsvc.PluginConfig{}
	}
	cfg.Dynamic.StoragePath = plugintestutil.TestDynamicStorageDir()
	if entries, ok := rootTestPluginAutoEnableOverride(); ok {
		cfg.AutoEnable = entries
	}
	if allowForceUninstall, ok := rootTestPluginAllowForceUninstallOverride(); ok {
		cfg.AllowForceUninstall = allowForceUninstall
	}
	return cfg
}

// GetPluginAutoEnable returns the startup auto-enable IDs from the overridden plugin config.
func (s rootTestConfigService) GetPluginAutoEnable(ctx context.Context) []string {
	entries := s.GetPluginAutoEnableEntries(ctx)
	if len(entries) == 0 {
		return nil
	}
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.ID)
	}
	return ids
}

// GetPluginAutoEnableEntries returns detached startup auto-enable entries.
func (s rootTestConfigService) GetPluginAutoEnableEntries(ctx context.Context) []configsvc.PluginAutoEnableEntry {
	cfg := s.GetPlugin(ctx)
	if cfg == nil || len(cfg.AutoEnable) == 0 {
		return nil
	}
	out := make([]configsvc.PluginAutoEnableEntry, len(cfg.AutoEnable))
	copy(out, cfg.AutoEnable)
	return out
}

// GetPluginDynamicStoragePath returns the isolated dynamic plugin storage path.
func (s rootTestConfigService) GetPluginDynamicStoragePath(context.Context) string {
	return plugintestutil.TestDynamicStorageDir()
}

// setTestPluginAutoEnableOverride sets ID-only startup auto-enable entries for one test.
func setTestPluginAutoEnableOverride(pluginIDs []string) {
	if len(pluginIDs) == 0 {
		setTestPluginAutoEnableEntriesOverride(nil)
		return
	}
	entries := make([]configsvc.PluginAutoEnableEntry, 0, len(pluginIDs))
	for _, pluginID := range pluginIDs {
		entries = append(entries, configsvc.PluginAutoEnableEntry{ID: pluginID})
	}
	setTestPluginAutoEnableEntriesOverride(entries)
}

// setTestPluginAutoEnableEntriesOverride sets full startup auto-enable entries for one test.
func setTestPluginAutoEnableEntriesOverride(entries []configsvc.PluginAutoEnableEntry) {
	rootTestConfigOverrides.Lock()
	defer rootTestConfigOverrides.Unlock()

	normalized := normalizeRootTestPluginAutoEnableEntries(entries)
	rootTestConfigOverrides.autoEnableSet = len(normalized) > 0
	rootTestConfigOverrides.autoEnableEntries = normalized
}

// setTestPluginAllowForceUninstallOverride sets the force-uninstall policy for one test.
func setTestPluginAllowForceUninstallOverride(value *bool) {
	rootTestConfigOverrides.Lock()
	defer rootTestConfigOverrides.Unlock()

	if value == nil {
		rootTestConfigOverrides.allowForceUninstall = nil
		return
	}
	copied := *value
	rootTestConfigOverrides.allowForceUninstall = &copied
}

// rootTestPluginAutoEnableOverride returns a detached startup auto-enable override.
func rootTestPluginAutoEnableOverride() ([]configsvc.PluginAutoEnableEntry, bool) {
	rootTestConfigOverrides.Lock()
	defer rootTestConfigOverrides.Unlock()

	if !rootTestConfigOverrides.autoEnableSet {
		return nil, false
	}
	out := make([]configsvc.PluginAutoEnableEntry, len(rootTestConfigOverrides.autoEnableEntries))
	copy(out, rootTestConfigOverrides.autoEnableEntries)
	return out, true
}

// rootTestPluginAllowForceUninstallOverride returns the configured force-uninstall override.
func rootTestPluginAllowForceUninstallOverride() (bool, bool) {
	rootTestConfigOverrides.Lock()
	defer rootTestConfigOverrides.Unlock()

	if rootTestConfigOverrides.allowForceUninstall == nil {
		return false, false
	}
	return *rootTestConfigOverrides.allowForceUninstall, true
}

// normalizeRootTestPluginAutoEnableEntries mirrors the production config shape
// enough for package tests without exporting config test hooks.
func normalizeRootTestPluginAutoEnableEntries(entries []configsvc.PluginAutoEnableEntry) []configsvc.PluginAutoEnableEntry {
	if len(entries) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(entries))
	out := make([]configsvc.PluginAutoEnableEntry, 0, len(entries))
	for _, entry := range entries {
		id := strings.TrimSpace(entry.ID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, configsvc.PluginAutoEnableEntry{
			ID:           id,
			WithMockData: entry.WithMockData,
		})
	}
	return out
}

// newTestService constructs the root plugin facade with default single-node topology.
func newTestService() *serviceImpl {
	return newTestServiceWithTopology(nil)
}

// newTestServiceWithTopology constructs the root plugin facade with one explicit topology.
func newTestServiceWithTopology(topology cluster.Service) *serviceImpl {
	service, err := newTestServiceWithTopologyAndTenantDeps(topology, nil)
	if err != nil {
		panic(err)
	}
	return service
}

// newTestServiceWithTopologyAndTenantDeps constructs the root plugin facade
// with explicit tenant governance dependencies for tests that need to replace
// the startup-owned tenant service.
func newTestServiceWithTopologyAndTenantDeps(
	topology cluster.Service,
	tenantSvc tenantspi.Service,
) (*serviceImpl, error) {
	var (
		configProvider = newRootTestConfigService()
		bizCtxProvider = bizctx.New()
		cacheCoordSvc  = cachecoord.Default(cachecoord.NewStaticTopology(false))
		pluginRuntime  = NewRuntimeDelegate()
	)
	orgSvc := orgspi.New(nil, pluginRuntime, pluginRuntime.OrgProviderEnv)
	defaultTenantSvc := tenantspi.New(nil, pluginRuntime, pluginRuntime.TenantProviderEnv, bizCtxProvider)
	if tenantSvc == nil {
		tenantSvc = defaultTenantSvc
	} else {
		switch fake := tenantSvc.(type) {
		case *startupConsistencyTenantCapability:
			if fake.Service == nil {
				fake.Service = defaultTenantSvc
			}
		case *autoEnableTenantProvisioningService:
			if fake.Service == nil {
				fake.Service = defaultTenantSvc
			}
		}
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
			tenantSvc,
			NewPluginConfigFactory("", ""),
			NewHostConfigService(configProvider),
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
	var (
		lockerSvc = locker.New()
		i18nSvc   = i18nsvc.New(bizCtxProvider, configProvider, cacheCoordSvc)
		roleSvc   = role.New(pluginRuntime, bizCtxProvider, configProvider, i18nSvc, orgSvc, tenantSvc)
	)
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
		tenantSvc,
		NewPluginConfigFactory("", ""),
		NewHostConfigService(configProvider),
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

// rootTestCapabilities publishes the minimal host service directory required
// by root-package plugin facade tests. It mirrors the production capability
// wiring only for services used by provider construction and leaves unrelated
// capability surfaces at nil or neutral fallback values.
type rootTestCapabilities struct {
	// bizCtx exposes the request business-context projection to provider plugins.
	bizCtx bizctxcap.Service
	// users exposes a registration-safe user-domain capability for providers.
	users capabilityusercap.Service
	// plugins exposes a registration-safe plugin-governance capability for providers.
	plugins capabilityplugincap.Service
	// storage exposes a registration-safe no-op storage service for runtime cleanup.
	storage storagecap.Service
}

// Ensure rootTestCapabilities can return plugin-scoped capability views.
var _ capabilityowner.ScopedServicesFactory = (*rootTestCapabilities)(nil)

// newRootTestCapabilities creates the minimal capability directory used by root tests.
func newRootTestCapabilities(
	bizCtxProvider bizctx.Service,
	lifecycleSvc plugincap.LifecycleService,
) capability.Services {
	return &rootTestCapabilities{
		bizCtx:  bizCtxProvider,
		users:   rootNoopUsers{},
		plugins: rootNoopPlugins{lifecycle: plugincap.NewLifecycle(lifecycleSvc)},
		storage: rootNoopStorage{},
	}
}

// APIDoc returns no API-documentation service for root plugin facade tests.
func (s *rootTestCapabilities) APIDoc() apidoccap.Service { return nil }

// Auth returns no auth namespace for root plugin facade tests.
func (s *rootTestCapabilities) Auth() authcap.Service { return nil }

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
		bizCtx:  s.bizCtx,
		users:   s.users,
		plugins: s.plugins,
		storage: s.storage,
	}
}

// HostConfig returns no host configuration service for root plugin facade tests.
func (s *rootTestCapabilities) HostConfig() hostconfigcap.Service { return nil }

// I18n returns no translation service for root plugin facade tests.
func (s *rootTestCapabilities) I18n() i18ncap.Service { return nil }

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
	return orgspi.New(nil, nil, nil)
}

// Plugins returns a registration-safe plugin-governance service for root plugin facade tests.
func (s *rootTestCapabilities) Plugins() capabilityplugincap.Service {
	if s == nil {
		return nil
	}
	return s.plugins
}

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
		return tenantspi.New(nil, nil, nil, nil)
	}
	filter, _ := tenantspi.NewPluginTableFilter(s.bizCtx)
	return rootTestTenantService{
		Service: tenantspi.New(nil, nil, nil, s.bizCtx),
		plugins: s.plugins,
		filter:  filter,
	}
}

// rootNoopStorage is a registration-safe object-storage fixture for root facade tests.
type rootNoopStorage struct{}

// rootTestTenantService attaches tenant-plugin governance to the fallback tenant service.
type rootTestTenantService struct {
	tenantcapsvc.Service
	plugins capabilityplugincap.Service
	filter  tenantcapsvc.FilterService
}

// Plugins returns tenant-plugin governance through the plugin-domain test fixture.
func (s rootTestTenantService) Plugins() tenantcapsvc.PluginService {
	if governance, ok := s.plugins.(tenantcapsvc.PluginService); ok {
		return governance
	}
	return nil
}

// Filter returns source-plugin tenant filter context from the root test fixture.
func (s rootTestTenantService) Filter() tenantcapsvc.FilterService {
	if s.filter != nil {
		return s.filter
	}
	if s.Service == nil {
		return nil
	}
	return s.Service.Filter()
}

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

// rootNoopUsers is a registration-safe user-domain fixture for root facade tests.
type rootNoopUsers struct{}

// Current returns no current user for provider-construction paths.
func (rootNoopUsers) Current(context.Context) (*capabilityusercap.UserInfo, error) {
	return nil, nil
}

// Get reports the requested user as absent without querying storage.
func (rootNoopUsers) Get(context.Context, capabilityusercap.UserID) (*capabilityusercap.UserInfo, error) {
	return nil, nil
}

// BatchGet reports all requested IDs as missing without querying storage.
func (rootNoopUsers) BatchGet(_ context.Context, ids []capabilityusercap.UserID) (*capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.UserID], error) {
	return &capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.UserID]{
		Items:      map[capabilityusercap.UserID]*capabilityusercap.UserInfo{},
		MissingIDs: append([]capabilityusercap.UserID(nil), ids...),
	}, nil
}

// BatchResolve reports all requested identifiers as missing without querying storage.
func (rootNoopUsers) BatchResolve(_ context.Context, input capabilityusercap.BatchResolveInput) (*capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.ResolveKey], error) {
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

// List returns an empty bounded page for provider-construction paths.
func (rootNoopUsers) List(context.Context, capabilityusercap.ListInput) (*capmodel.PageResult[*capabilityusercap.UserInfo], error) {
	return &capmodel.PageResult[*capabilityusercap.UserInfo]{Items: []*capabilityusercap.UserInfo{}}, nil
}

// EnsureVisible accepts checks because root facade tests do not execute user business paths.
func (rootNoopUsers) EnsureVisible(context.Context, []capabilityusercap.UserID) error {
	return nil
}

// Create accepts user creation without mutating shared test state.
func (rootNoopUsers) Create(context.Context, capabilityusercap.CreateInput) (capabilityusercap.UserID, error) {
	return "", nil
}

// Update accepts user updates without mutating shared test state.
func (rootNoopUsers) Update(context.Context, capabilityusercap.UpdateInput) error {
	return nil
}

// Delete accepts user deletion without mutating shared test state.
func (rootNoopUsers) Delete(context.Context, capabilityusercap.UserID) error {
	return nil
}

// SetStatus accepts status changes without mutating shared test state.
func (rootNoopUsers) SetStatus(context.Context, capabilityusercap.UserID, statusflag.Enabled) error {
	return nil
}

// ResetPassword accepts password resets without mutating shared test state.
func (rootNoopUsers) ResetPassword(context.Context, capabilityusercap.UserID, string) error {
	return nil
}

// Assignment returns role assignment operations.
func (rootNoopUsers) Assignment() capabilityusercap.AssignmentService {
	return rootNoopUserAssignments{}
}

// rootNoopUserAssignments accepts user-role changes without mutating shared test state.
type rootNoopUserAssignments struct{}

// ReplaceRoles accepts role replacement without mutating shared test state.
func (rootNoopUserAssignments) ReplaceRoles(context.Context, capabilityusercap.UserID, []int) error {
	return nil
}

// rootNoopPlugins is a registration-safe plugin-governance fixture for root facade tests.
type rootNoopPlugins struct {
	lifecycle capabilityplugincap.LifecycleService
}

// Current returns no current plugin projection for construction-only tests.
func (rootNoopPlugins) Current(context.Context) (*capabilityplugincap.PluginInfo, error) {
	return nil, nil
}

// Get reports the requested plugin as absent without querying storage.
func (rootNoopPlugins) Get(context.Context, capabilityplugincap.PluginID) (*capabilityplugincap.PluginInfo, error) {
	return nil, nil
}

// BatchGet reports all requested plugin IDs as missing projections.
func (rootNoopPlugins) BatchGet(_ context.Context, ids []capabilityplugincap.PluginID) (*capmodel.BatchResult[*capabilityplugincap.PluginInfo, capabilityplugincap.PluginID], error) {
	return &capmodel.BatchResult[*capabilityplugincap.PluginInfo, capabilityplugincap.PluginID]{
		Items:      map[capabilityplugincap.PluginID]*capabilityplugincap.PluginInfo{},
		MissingIDs: append([]capabilityplugincap.PluginID(nil), ids...),
	}, nil
}

// List returns an empty plugin-governance page for construction-only tests.
func (rootNoopPlugins) List(context.Context, capabilityplugincap.ListInput) (*capmodel.PageResult[*capabilityplugincap.PluginInfo], error) {
	return &capmodel.PageResult[*capabilityplugincap.PluginInfo]{Items: []*capabilityplugincap.PluginInfo{}}, nil
}

// ListTenantPlugins returns an empty tenant plugin page for construction-only tests.
func (rootNoopPlugins) ListTenantPlugins(context.Context, capabilityplugincap.TenantListInput) (*capmodel.PageResult[*capabilityplugincap.TenantPluginInfo], error) {
	return &capmodel.PageResult[*capabilityplugincap.TenantPluginInfo]{Items: []*capabilityplugincap.TenantPluginInfo{}}, nil
}

// IsEnabled reports plugins as disabled in construction-only tests.
func (rootNoopPlugins) IsEnabled(context.Context, capabilityplugincap.PluginID) (bool, error) {
	return false, nil
}

// IsProviderEnabled reports providers as disabled in construction-only tests.
func (rootNoopPlugins) IsProviderEnabled(context.Context, capabilityplugincap.PluginID) (bool, error) {
	return false, nil
}

// IsEnabledAuthoritative reports persisted plugin enablement as disabled in construction-only tests.
func (rootNoopPlugins) IsEnabledAuthoritative(context.Context, capabilityplugincap.PluginID) (bool, error) {
	return false, nil
}

// Config returns a no-op plugin configuration reader for construction-only tests.
func (rootNoopPlugins) Config() capabilityplugincap.ConfigService {
	return rootNoopPluginConfig{}
}

// Registry returns the no-op plugin registry projection.
func (s rootNoopPlugins) Registry() capabilityplugincap.RegistryService {
	return s
}

// State returns the no-op plugin enablement lookup projection.
func (s rootNoopPlugins) State() capabilityplugincap.StateService {
	return s
}

// Lifecycle returns lifecycle operations for construction-only tests.
func (s rootNoopPlugins) Lifecycle() capabilityplugincap.LifecycleService {
	return s.lifecycle
}

// rootNoopPluginConfig is a no-op plugin config reader for root facade tests.
type rootNoopPluginConfig struct{}

// Get reports that the requested plugin config key does not exist.
func (rootNoopPluginConfig) Get(_ context.Context, _ string, defaultValue any) (*gvar.Var, error) {
	if defaultValue != nil {
		return gvar.New(defaultValue), nil
	}
	return nil, nil
}

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

// SetTenantPluginEnabled accepts enablement changes without mutating shared test state.
func (rootNoopPlugins) SetTenantPluginEnabled(context.Context, capabilityplugincap.PluginID, bool) error {
	return nil
}

// ProvisionTenantPluginDefaults accepts default provisioning without mutating test state.
func (rootNoopPlugins) ProvisionTenantPluginDefaults(context.Context, capmodel.DomainID) error {
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
		plugintypes.NormalizeType(registry.Type) == pluginv1.PluginTypeDynamic &&
		registry.Installed == statusflag.Installed.Int() &&
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

// testTopology lets root-package tests simulate clustered primary/follower behavior.
type testTopology struct {
	mu      sync.RWMutex
	enabled bool
	primary bool
	nodeID  string
}

// Start records no behavior for the in-memory test topology.
func (t *testTopology) Start(context.Context) {}

// Stop records no behavior for the in-memory test topology.
func (t *testTopology) Stop(context.Context) {}

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

// TestPluginInternalImportBoundaries verifies leaf and manifest packages do not
// grow forbidden sibling or generated-model imports.
func TestPluginInternalImportBoundaries(t *testing.T) {
	checkForbiddenImports(t, "internal/plugintypes", map[string]string{
		"lina-core/internal/service/plugin/internal/catalog":     "plugintypes must remain independent from catalog",
		"lina-core/internal/service/plugin/internal/store":       "plugintypes must remain independent from store",
		"lina-core/internal/service/plugin/internal/runtime":     "plugintypes must remain independent from runtime",
		"lina-core/internal/service/plugin/internal/integration": "plugintypes must remain independent from integration",
		"lina-core/internal/dao":                                 "plugintypes must not depend on generated DAO models",
		"lina-core/internal/model/do":                            "plugintypes must not depend on generated DO models",
		"lina-core/internal/model/entity":                        "plugintypes must not depend on generated entity models",
	})
	checkForbiddenImports(t, "internal/catalog", map[string]string{
		"lina-core/internal/service/plugin/internal/runtime":     "catalog must not depend on runtime implementation",
		"lina-core/internal/service/plugin/internal/integration": "catalog must not depend on integration implementation",
		"lina-core/internal/dao":                                 "catalog must not own governance persistence",
		"lina-core/internal/model/do":                            "catalog must not own generated DO writes",
		"lina-core/internal/model/entity":                        "catalog must not expose generated entity reads",
	})
}

// TestPluginStoreExportedSurfaceDoesNotLeakGeneratedModels verifies store owns
// generated model access internally without returning those types as API.
func TestPluginStoreExportedSurfaceDoesNotLeakGeneratedModels(t *testing.T) {
	files := parseGoFiles(t, "internal/store")
	for path, file := range files {
		for _, decl := range file.Decls {
			switch typed := decl.(type) {
			case *ast.GenDecl:
				if typed.Tok != token.TYPE {
					continue
				}
				for _, spec := range typed.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok || !typeSpec.Name.IsExported() {
						continue
					}
					if usesGeneratedModel(typeSpec.Type) {
						t.Fatalf("%s: exported store type %s leaks generated DAO/DO/entity model types", path, typeSpec.Name.Name)
					}
				}
			case *ast.FuncDecl:
				if !exportedFuncDecl(typed) {
					continue
				}
				if usesGeneratedModel(typed.Type) {
					t.Fatalf("%s: exported store function %s leaks generated DAO/DO/entity model types", path, typed.Name.Name)
				}
			}
		}
	}
}

// TestCatalogSetterWiringRemoved verifies catalog no longer exposes or stores
// the old runtime/integration callback wiring.
func TestCatalogSetterWiringRemoved(t *testing.T) {
	files := parseGoFiles(t, "internal/catalog")
	for path, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch typed := node.(type) {
			case *ast.FuncDecl:
				if strings.HasPrefix(typed.Name.Name, "Set") {
					t.Fatalf("%s: catalog must not define Set* wiring method/function %s", path, typed.Name.Name)
				}
			case *ast.TypeSpec:
				if typed.Name.Name != "serviceImpl" {
					return true
				}
				structType, ok := typed.Type.(*ast.StructType)
				if !ok {
					return true
				}
				for _, field := range structType.Fields.List {
					for _, name := range field.Names {
						if forbiddenCatalogCallbackField(name.Name) {
							t.Fatalf("%s: catalog serviceImpl still stores callback field %s", path, name.Name)
						}
					}
				}
			}
			return true
		})
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "plugin.go", nil, 0)
	if err != nil {
		t.Fatalf("parse plugin.go: %v", err)
	}
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !strings.HasPrefix(selector.Sel.Name, "Set") {
			return true
		}
		ident, ok := selector.X.(*ast.Ident)
		if ok && ident.Name == "catalogSvc" {
			position := fileSet.Position(selector.Pos())
			t.Fatalf("%s: plugin.New must not call catalog Set* wiring method %s", position, selector.Sel.Name)
		}
		return true
	})
}

// TestPluginRuntimeCacheOldImportRemoved verifies plugin runtime cache
// coordination is imported from the cache coordination boundary, not the old
// plugin-owned package path.
func TestPluginRuntimeCacheOldImportRemoved(t *testing.T) {
	checkForbiddenImports(t, ".", map[string]string{
		"lina-core/internal/service/plugin/runtimecache": "plugin runtime cache revision control must live under cachecoord/revisionctrl",
	})
	checkForbiddenImports(t, "../i18n", map[string]string{
		"lina-core/internal/service/plugin/runtimecache": "i18n must use cachecoord/revisionctrl instead of the plugin package tree",
	})
	checkForbiddenImports(t, "../../cmd", map[string]string{
		"lina-core/internal/service/plugin/runtimecache": "startup tests must use cachecoord/revisionctrl instead of the old plugin package tree",
	})
}

// TestPluginWiringStateStaticBoundaries verifies the B-stage wiring and
// mutable-state cleanup cannot regress silently.
func TestPluginWiringStateStaticBoundaries(t *testing.T) {
	forbiddenFiles := []string{
		filepath.Join("internal", "integration", "integration_wiring.go"),
		filepath.Join("internal", "lifecycle", "lifecycle_wiring.go"),
	}
	for _, path := range forbiddenFiles {
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("%s: old production wiring file must not be restored", path)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", path, err)
		}
	}

	checkForbiddenText(t, "internal/runtime", map[string]string{
		"ValidateRequiredDependencies": "runtime dependencies must be provided by constructor parameters",
		"SetDependencyValidator":       "runtime dependency validator must be provided by constructor parameters",
		"SetMenuManager":               "runtime menu manager must be provided by constructor parameters",
		"SetHookDispatcher":            "runtime hook dispatcher must be provided by constructor parameters",
		"SetStorageCleanupServices":    "runtime storage cleanup services must be provided by constructor parameters",
	})
	checkForbiddenText(t, "internal/integration", map[string]string{
		"defaultSharedState":     "integration shared state must stay owned by the integration service",
		"SetDynamicJobExecutor":  "integration dynamic job runtime must be provided by constructor parameters",
		"SetOrganizationService": "integration organization capability must be provided by constructor parameters",
	})
	checkForbiddenText(t, ".", map[string]string{
		"lifecycleObserverByID":              "lifecycle observers must stay on service instances",
		"pluginRuntimeCacheObservedRevision": "runtime cache observed revision must stay in revision controllers",
	})
	checkForbiddenText(t, "internal/wasm", map[string]string{
		"atomic.Pointer": "WASM host service dependencies must be held by runtime instances",
		"func Configure": "WASM host service production Configure* entrypoints must not be restored",
	})
	checkForbiddenText(t, "internal/testutil/testutil_services.go", map[string]string{
		"SetDependencyValidator": "testutil must not replicate old runtime setter wiring",
		"SetMenuManager":         "testutil must not replicate old runtime/integration setter wiring",
		"SetHookDispatcher":      "testutil must not replicate old runtime/integration setter wiring",
		"SetReconciler":          "testutil must not replicate old lifecycle setter wiring",
	})
	checkForbiddenPluginRootWiringMethods(t)
	checkForbiddenHTTPStartupPluginSetterCalls(t)
	checkForbiddenImports(t, "../middleware", map[string]string{
		"lina-core/internal/service/plugin": "middleware must not depend on the complete plugin facade when publishing route middleware",
	})
}

// TestCapabilityHostInternalMicroPackagesRemoved verifies host capability
// domain adapters stay inside the capabilityhost package instead of re-growing
// single-file internal micro packages.
func TestCapabilityHostInternalMicroPackagesRemoved(t *testing.T) {
	internalDir := filepath.Join("internal", "capabilityhost", "internal")
	if _, err := os.Stat(internalDir); os.IsNotExist(err) {
		return
	} else if err != nil {
		t.Fatalf("stat %s: %v", internalDir, err)
	}
	if err := filepath.WalkDir(internalDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			t.Fatalf("%s: capabilityhost internal micro package file must be merged into the capabilityhost package", path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk %s: %v", internalDir, err)
	}
}

// TestPluginLifecycleOrchestrationStaticBoundaries captures C-stage ownership
// boundaries while lifecycle orchestration is being migrated in batches.
func TestPluginLifecycleOrchestrationStaticBoundaries(t *testing.T) {
	checkForbiddenText(t, "internal/lifecycle", map[string]string{
		"ResolveSQLAssets":       "SQL asset resolution belongs to internal/migration",
		"ResolvePluginSQLAssets": "SQL asset resolution belongs to internal/migration",
		"SysPluginMigration":     "migration ledger writes belong to internal/migration",
		"ReconcileProvider":      "lifecycle must receive the complete orchestration runtime seam, not the old dynamic-only reconciler seam",
	})
	if _, err := os.Stat(filepath.Join("internal", "migration", "migration.go")); err != nil {
		t.Fatalf("internal/migration/migration.go must own SQL migration execution: %v", err)
	}
	checkForbiddenText(t, ".", map[string]string{
		"lifecycleReconcilerProvider": "lifecycle must be constructed after runtime with explicit orchestration dependencies",
	})
	checkLifecycleConstructorWiring(t)

	transitionalRootDAO := map[string]map[string]string{
		"plugin.go": {
			"lina-core/internal/model/entity": "root facade still exposes route projection contracts that use generated menu entity until route/presentation cleanup",
		},
		"plugin_integration.go": {
			"lina-core/internal/model/entity": "integration projection facade still exposes menu entity until projection contracts are narrowed",
		},
		"plugin_runtime_delegates.go": {
			"lina-core/internal/model/entity": "runtime delegate still bridges permission menu filtering until projection contracts are narrowed",
		},
	}
	checkPluginRootGeneratedModelImports(t, transitionalRootDAO)
	checkPluginProjectionBuilderBoundary(t)
	checkPluginChangePublisherBoundary(t)
}

// TestPluginUpgradeOrchestrationStaticBoundaries captures the D-stage upgrade
// ownership boundaries now that source and dynamic upgrade orchestration live
// under the unified upgrade owner.
func TestPluginUpgradeOrchestrationStaticBoundaries(t *testing.T) {
	var (
		upgradeDir        = filepath.Join("internal", "upgrade")
		sourceUpgradeDir  = filepath.Join("internal", "sourceupgrade")
		runtimeUpgradeDir = filepath.Join("internal", "runtimeupgrade")
	)

	if _, err := os.Stat(upgradeDir); err != nil {
		t.Fatalf("internal/upgrade must own plugin upgrade orchestration: %v", err)
	}
	checkPackageName(t, upgradeDir, "upgrade")

	for _, legacyDir := range []string{sourceUpgradeDir, runtimeUpgradeDir} {
		if _, err := os.Stat(legacyDir); err == nil {
			t.Fatalf("%s: old upgrade package must not be restored", legacyDir)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", legacyDir, err)
		}
	}

	checkForbiddenImports(t, ".", map[string]string{
		"lina-core/internal/service/plugin/internal/sourceupgrade":  "source upgrade orchestration must live under internal/upgrade",
		"lina-core/internal/service/plugin/internal/runtimeupgrade": "runtime upgrade planning must live under internal/upgrade",
	})
	checkForbiddenText(t, "internal/upgrade", map[string]string{
		"recordSourceUpgradeFailureMigration": "runtime upgrade failure ledger writes must use the unified owner helper",
	})
	checkRootRuntimeUpgradeDoesNotReenterSourcePublicMethod(t)
}

// checkPluginProjectionBuilderBoundary keeps management list, summary, detail,
// and dependency snapshot paths on the same projection builder.
func checkPluginProjectionBuilderBoundary(t *testing.T) {
	t.Helper()

	projectionCallers := map[string][]string{
		"plugin_list.go": {
			"syncAndList",
			"buildManagementList",
			"buildManagementSummaryList",
			"buildManagementDetail",
		},
		"plugin_dependency.go": {
			"buildDependencySnapshots",
		},
	}
	for path, functions := range projectionCallers {
		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, name := range functions {
			decl := findFuncDecl(file, name)
			if decl == nil {
				t.Fatalf("%s: expected function %s", path, name)
			}
			if !funcDeclCalls(decl, "buildPluginProjection") {
				t.Fatalf("%s: %s must use buildPluginProjection", path, name)
			}
			if path == "plugin_list.go" && funcDeclCalls(decl, "ScanManifests") {
				t.Fatalf("%s: %s must not own a separate manifest scan pipeline", path, name)
			}
		}
	}
}

// checkPluginChangePublisherBoundary keeps runtime revision publication,
// management read-model invalidation, and derived cache invalidation behind the
// same root plugin change publisher.
func checkPluginChangePublisherBoundary(t *testing.T) {
	t.Helper()

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "plugin_runtime_cache.go", nil, 0)
	if err != nil {
		t.Fatalf("parse plugin_runtime_cache.go: %v", err)
	}
	publisher := findFuncDecl(file, "publishPluginChange")
	if publisher == nil {
		t.Fatalf("plugin_runtime_cache.go: expected publishPluginChange")
	}
	for _, name := range []string{
		"invalidateRuntimeUpgradeCaches",
		"InvalidateManagementListCache",
		"MarkChanged",
	} {
		if !funcDeclCalls(publisher, name) {
			t.Fatalf("plugin_runtime_cache.go: publishPluginChange must call %s", name)
		}
	}
	for _, name := range []string{
		"MarkRuntimeCacheChanged",
		"PublishPluginChange",
		"syncEnabledSnapshotAndPublishRuntimeChange",
	} {
		decl := findFuncDecl(file, name)
		if decl == nil {
			t.Fatalf("plugin_runtime_cache.go: expected function %s", name)
		}
		if !funcDeclCalls(decl, "publishPluginChange") {
			t.Fatalf("plugin_runtime_cache.go: %s must delegate to publishPluginChange", name)
		}
	}

	listFile := parseSingleFile(t, "plugin_list.go")
	for _, name := range []string{"SyncSourcePluginsStrict", "SyncAndList"} {
		decl := findFuncDecl(listFile, name)
		if decl == nil {
			t.Fatalf("plugin_list.go: expected function %s", name)
		}
		if !funcDeclCalls(decl, "publishPluginChange") {
			t.Fatalf("plugin_list.go: %s must publish through publishPluginChange", name)
		}
		if funcDeclCalls(decl, "markRuntimeCacheChanged") || funcDeclCalls(decl, "Invalidate") {
			t.Fatalf("plugin_list.go: %s must not bypass publishPluginChange for cache invalidation", name)
		}
	}

	lifecycleFile := parseSingleFile(t, filepath.Join("internal", "lifecycle", "lifecycle_cache.go"))
	lifecyclePublisher := findFuncDecl(lifecycleFile, "syncEnabledSnapshotAndPublishRuntimeChange")
	if lifecyclePublisher == nil {
		t.Fatalf("internal/lifecycle/lifecycle_cache.go: expected syncEnabledSnapshotAndPublishRuntimeChange")
	}
	if !funcDeclCalls(lifecyclePublisher, "PublishPluginChange") {
		t.Fatalf("internal/lifecycle/lifecycle_cache.go: syncEnabledSnapshotAndPublishRuntimeChange must use scoped PublishPluginChange")
	}
	if funcDeclCalls(lifecyclePublisher, "MarkRuntimeCacheChanged") {
		t.Fatalf("internal/lifecycle/lifecycle_cache.go: syncEnabledSnapshotAndPublishRuntimeChange must not use legacy MarkRuntimeCacheChanged")
	}

	runtimeFile := parseSingleFile(t, filepath.Join("internal", "runtime", "runtime_wiring.go"))
	runtimePublisher := findFuncDecl(runtimeFile, "notifyRuntimeCacheChanged")
	if runtimePublisher == nil {
		t.Fatalf("internal/runtime/runtime_wiring.go: expected notifyRuntimeCacheChanged")
	}
	if !funcDeclCalls(runtimePublisher, "PublishPluginChange") {
		t.Fatalf("internal/runtime/runtime_wiring.go: notifyRuntimeCacheChanged must use scoped PublishPluginChange")
	}
	if funcDeclCalls(runtimePublisher, "MarkRuntimeCacheChanged") {
		t.Fatalf("internal/runtime/runtime_wiring.go: notifyRuntimeCacheChanged must not use legacy MarkRuntimeCacheChanged")
	}
}

// checkLifecycleConstructorWiring ensures C-stage lifecycle orchestration gets
// explicit owner dependencies through its constructor instead of a dynamic-only
// post-construction bridge.
func checkLifecycleConstructorWiring(t *testing.T) {
	t.Helper()

	fileSet := token.NewFileSet()
	lifecycleFile, err := parser.ParseFile(fileSet, filepath.Join("internal", "lifecycle", "lifecycle.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse lifecycle.go: %v", err)
	}
	newArgCount := -1
	for _, decl := range lifecycleFile.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "New" || fn.Recv != nil {
			continue
		}
		newArgCount = fn.Type.Params.NumFields()
	}
	if newArgCount != 11 {
		t.Fatalf("internal/lifecycle.New must explicitly receive 11 orchestration dependencies, got %d", newArgCount)
	}

	pluginFile, err := parser.ParseFile(fileSet, "plugin.go", nil, 0)
	if err != nil {
		t.Fatalf("parse plugin.go: %v", err)
	}
	callArgCount := -1
	ast.Inspect(pluginFile, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "New" {
			return true
		}
		ident, ok := selector.X.(*ast.Ident)
		if !ok || ident.Name != "lifecycle" {
			return true
		}
		callArgCount = len(call.Args)
		return true
	})
	if callArgCount != 11 {
		t.Fatalf("plugin.New must construct lifecycle with 11 explicit dependencies, got %d", callArgCount)
	}
}

// findFuncDecl returns the named package-level function declaration.
func findFuncDecl(file *ast.File, name string) *ast.FuncDecl {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Name.Name == name {
			return fn
		}
	}
	return nil
}

// parseSingleFile parses one Go source file from the plugin package tree.
func parseSingleFile(t *testing.T, path string) *ast.File {
	t.Helper()
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return file
}

// funcDeclCalls reports whether function body calls a selector or function name.
func funcDeclCalls(decl *ast.FuncDecl, name string) bool {
	if decl == nil || decl.Body == nil {
		return false
	}
	found := false
	ast.Inspect(decl.Body, func(node ast.Node) bool {
		if found {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			found = fun.Name == name
		case *ast.SelectorExpr:
			found = fun.Sel.Name == name
		}
		return !found
	})
	return found
}

func checkForbiddenImports(t *testing.T, dir string, forbidden map[string]string) {
	t.Helper()
	for path, file := range parseGoFiles(t, dir) {
		for _, importSpec := range file.Imports {
			importPath := strings.Trim(importSpec.Path.Value, `"`)
			if reason, ok := forbidden[importPath]; ok {
				t.Fatalf("%s imports %s: %s", path, importPath, reason)
			}
		}
	}
}

// checkPackageName verifies all production Go files in dir use the expected package.
func checkPackageName(t *testing.T, dir string, expected string) {
	t.Helper()
	for path, file := range parseGoFiles(t, dir) {
		if file.Name == nil || file.Name.Name != expected {
			t.Fatalf("%s: expected package %s, got %v", path, expected, file.Name)
		}
	}
}

// checkRootRuntimeUpgradeDoesNotReenterSourcePublicMethod blocks the old source
// branch where runtime upgrade execution called the public source upgrade entry.
func checkRootRuntimeUpgradeDoesNotReenterSourcePublicMethod(t *testing.T) {
	t.Helper()
	file := parseSingleFile(t, "plugin_runtime_upgrade.go")
	decl := findFuncDecl(file, "executeRuntimeUpgradeByType")
	if decl == nil {
		return
	}
	if funcDeclCalls(decl, "UpgradeSourcePlugin") {
		t.Fatalf("plugin_runtime_upgrade.go: executeRuntimeUpgradeByType must call the upgrade owner directly, not public UpgradeSourcePlugin")
	}
}

// checkPluginRootGeneratedModelImports blocks new root-facade DAO/DO/Entity
// imports while documenting the C-stage files that are intentionally migrated
// later in this change.
func checkPluginRootGeneratedModelImports(t *testing.T, allowlist map[string]map[string]string) {
	t.Helper()
	for path, file := range parseGoFiles(t, ".") {
		if strings.HasPrefix(path, "internal"+string(filepath.Separator)) ||
			strings.Contains(path, string(filepath.Separator)+"internal"+string(filepath.Separator)) {
			continue
		}
		base := filepath.Base(path)
		for _, importSpec := range file.Imports {
			importPath := strings.Trim(importSpec.Path.Value, `"`)
			if importPath != "lina-core/internal/dao" &&
				importPath != "lina-core/internal/model/do" &&
				importPath != "lina-core/internal/model/entity" {
				continue
			}
			if allowedForFile, ok := allowlist[base]; ok {
				if _, allowed := allowedForFile[importPath]; allowed {
					continue
				}
			}
			t.Fatalf("%s imports %s: root plugin facade must not gain new generated model access during C migration", path, importPath)
		}
	}
}

// checkForbiddenPluginRootWiringMethods verifies the root plugin facade does
// not reintroduce startup-only wiring methods.
func checkForbiddenPluginRootWiringMethods(t *testing.T) {
	t.Helper()
	forbidden := map[string]string{
		"SetCapabilities":                       "capability services must be passed into plugin.New",
		"SetOrganizationCapability":             "organization capability must be passed into plugin.New",
		"SetTenantStartupCapability":            "tenant startup capability must be passed into plugin.New",
		"SetTenantProvisioningCapability":       "tenant provisioning capability must be passed into plugin.New",
		"SetTenantPlatformGovernanceCapability": "tenant governance capability must be passed into plugin.New",
		"ConfigureWasmHostServices":             "WASM host service runtime must be constructed during plugin.New",
	}
	allowedFacets := map[string]struct{}{
		"managementService":      {},
		"startupService":         {},
		"runtimeHTTPService":     {},
		"integrationService":     {},
		"jobService":             {},
		"stateService":           {},
		"capabilityEnvService":   {},
		"tenantLifecycleService": {},
	}
	for path, file := range parseGoFiles(t, ".") {
		if filepath.Base(path) != "plugin.go" {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch typed := node.(type) {
			case *ast.FuncDecl:
				if reason, ok := forbidden[typed.Name.Name]; ok {
					t.Fatalf("%s: root plugin production method %s is forbidden: %s", path, typed.Name.Name, reason)
				}
			case *ast.TypeSpec:
				interfaceType, ok := typed.Type.(*ast.InterfaceType)
				if !ok || typed.Name.Name != "Service" {
					return true
				}
				for _, method := range interfaceType.Methods.List {
					if len(method.Names) == 0 {
						ident, ok := method.Type.(*ast.Ident)
						if !ok {
							t.Fatalf("%s: plugin Service embeds unsupported interface expression", path)
						}
						if _, allowed := allowedFacets[ident.Name]; !allowed {
							t.Fatalf("%s: plugin Service embeds unsupported facet %s", path, ident.Name)
						}
						continue
					}
					for _, name := range method.Names {
						if reason, ok := forbidden[name.Name]; ok {
							t.Fatalf("%s: plugin interface %s exposes forbidden method %s: %s", path, typed.Name.Name, name.Name, reason)
						}
					}
				}
			}
			return true
		})
	}
}

// checkForbiddenHTTPStartupPluginSetterCalls verifies the root composition
// passes startup dependencies through constructors instead of post-construction
// plugin service setters.
func checkForbiddenHTTPStartupPluginSetterCalls(t *testing.T) {
	t.Helper()
	forbidden := map[string]string{
		"SetCapabilities":                       "capability services must be passed into plugin.New",
		"SetOrganizationCapability":             "organization capability must be passed into plugin.New",
		"SetTenantStartupCapability":            "tenant startup capability must be passed into plugin.New",
		"SetTenantProvisioningCapability":       "tenant provisioning capability must be passed into plugin.New",
		"SetTenantPlatformGovernanceCapability": "tenant governance capability must be passed into plugin.New",
		"ConfigureWasmHostServices":             "WASM host service runtime must be constructed during plugin.New",
	}
	fileSet := token.NewFileSet()
	path := filepath.Join("..", "..", "cmd", "internal", "httpstartup", "httpstartup_runtime.go")
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		reason, forbiddenCall := forbidden[selector.Sel.Name]
		if !forbiddenCall {
			return true
		}
		position := fileSet.Position(selector.Pos())
		t.Fatalf("%s: http startup must not call plugin service wiring method %s: %s", position, selector.Sel.Name, reason)
		return true
	})
}

// checkForbiddenText scans Go source files under path for forbidden text.
func checkForbiddenText(t *testing.T, path string, forbidden map[string]string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if !info.IsDir() {
		checkForbiddenTextFile(t, path, forbidden)
		return
	}
	if err = filepath.WalkDir(path, func(filePath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(filePath, ".go") && !strings.HasSuffix(filePath, "_test.go") {
			checkForbiddenTextFile(t, filePath, forbidden)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk %s: %v", path, err)
	}
}

// checkForbiddenTextFile scans one file for forbidden text.
func checkForbiddenTextFile(t *testing.T, path string, forbidden map[string]string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	text := string(content)
	for pattern, reason := range forbidden {
		if strings.Contains(text, pattern) {
			t.Fatalf("%s contains %q: %s", path, pattern, reason)
		}
	}
}

// parseGoFiles parses all non-test Go source files under dir.
func parseGoFiles(t *testing.T, dir string) map[string]*ast.File {
	t.Helper()
	files := make(map[string]*ast.File)
	fileSet := token.NewFileSet()
	if err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		files[path] = file
		return nil
	}); err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	if len(files) == 0 {
		t.Fatalf("expected Go files under %s", dir)
	}
	return files
}

// exportedFuncDecl reports whether decl contributes to the package API surface.
func exportedFuncDecl(decl *ast.FuncDecl) bool {
	if decl.Recv == nil {
		return decl.Name.IsExported()
	}
	if len(decl.Recv.List) == 0 {
		return false
	}
	return decl.Name.IsExported() && exportedReceiverType(decl.Recv.List[0].Type)
}

// exportedReceiverType reports whether a method receiver is externally named.
func exportedReceiverType(expr ast.Expr) bool {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.IsExported()
	case *ast.StarExpr:
		return exportedReceiverType(typed.X)
	default:
		return false
	}
}

// usesGeneratedModel reports whether node references generated DAO/DO/Entity selectors.
func usesGeneratedModel(node ast.Node) bool {
	found := false
	ast.Inspect(node, func(child ast.Node) bool {
		if found {
			return false
		}
		selector, ok := child.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := selector.X.(*ast.Ident)
		if !ok {
			return true
		}
		switch ident.Name {
		case "dao", "do", "entity":
			found = true
			return false
		default:
			return true
		}
	})
	return found
}

// forbiddenCatalogCallbackField reports whether name is an old catalog callback slot.
func forbiddenCatalogCallbackField(name string) bool {
	switch name {
	case "backendLoader", "artifactParser", "dynamicManifestLoader",
		"nodeStateSyncer", "menuSyncer", "resourceRefSyncer",
		"releaseStateSyncer", "hookDispatcher":
		return true
	default:
		return strings.HasSuffix(name, "Syncer") || strings.HasSuffix(name, "Dispatcher")
	}
}
