// This file keeps root-package test bootstrap and shared helpers for plugin facade tests.

package plugin

import (
	"context"
	"encoding/base64"
	"strconv"
	"sync"
	"testing"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/session"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/contract"
	orgcapsvc "lina-core/pkg/plugin/capability/orgcap"
	capabilitypluginlifecycle "lina-core/pkg/plugin/capability/pluginlifecycle"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
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
		cachecoord.DefaultWithCoordination(topology, coordSvc)
		cacheCoordSvc = cachecoord.Default(topology)
		i18nSvc := i18nsvc.New(bizCtxProvider, configProvider, cacheCoordSvc)
		service, err := New(topology, configProvider, bizCtxProvider, cacheCoordSvc, i18nSvc, session.NewDBStore(), locker.New(), coordSvc.Lock())
		if err != nil {
			panic(err)
		}
		serviceImpl := service.(*serviceImpl)
		tenantSvc := tenantcapsvc.New(serviceImpl, bizCtxProvider)
		serviceImpl.SetCapabilities(newRootTestCapabilities(bizCtxProvider, serviceImpl))
		serviceImpl.SetTenantStartupCapability(tenantSvc)
		serviceImpl.SetTenantProvisioningCapability(tenantSvc)
		serviceImpl.SetTenantPlatformGovernanceCapability(tenantSvc)
		return serviceImpl
	}
	i18nSvc := i18nsvc.New(bizCtxProvider, configProvider, cacheCoordSvc)
	service, err := New(topology, configProvider, bizCtxProvider, cacheCoordSvc, i18nSvc, session.NewDBStore(), locker.New(), nil)
	if err != nil {
		panic(err)
	}
	serviceImpl := service.(*serviceImpl)
	tenantSvc := tenantcapsvc.New(serviceImpl, bizCtxProvider)
	serviceImpl.SetCapabilities(newRootTestCapabilities(bizCtxProvider, serviceImpl))
	serviceImpl.SetTenantStartupCapability(tenantSvc)
	serviceImpl.SetTenantProvisioningCapability(tenantSvc)
	serviceImpl.SetTenantPlatformGovernanceCapability(tenantSvc)
	return serviceImpl
}

// rootTestCapabilities publishes the minimal host service directory required
// by root-package plugin facade tests. It mirrors the production capability
// wiring only for services used by provider construction and leaves unrelated
// capability surfaces at nil or neutral fallback values.
type rootTestCapabilities struct {
	// bizCtx exposes the request business-context projection to provider plugins.
	bizCtx contract.BizCtxService
	// pluginLifecycle exposes nil-tolerant lifecycle hooks to tenant provider code.
	pluginLifecycle contract.PluginLifecycleService
}

// Ensure rootTestCapabilities satisfies the source-plugin host service directory.
var _ pluginhost.Services = (*rootTestCapabilities)(nil)

// Ensure rootTestCapabilities can return plugin-scoped capability views.
var _ capability.ScopedServicesFactory = (*rootTestCapabilities)(nil)

// newRootTestCapabilities creates the minimal capability directory used by root tests.
func newRootTestCapabilities(
	bizCtxProvider bizctx.Service,
	lifecycleRunner contract.PluginLifecycleRunner,
) capability.Services {
	return &rootTestCapabilities{
		bizCtx:          bizCtxProvider,
		pluginLifecycle: capabilitypluginlifecycle.New(lifecycleRunner),
	}
}

// APIDoc returns no API-documentation service for root plugin facade tests.
func (s *rootTestCapabilities) APIDoc() contract.APIDocService { return nil }

// Auth returns no auth service for root plugin facade tests.
func (s *rootTestCapabilities) Auth() contract.AuthService { return nil }

// BizCtx returns the host business-context projection used by provider construction.
func (s *rootTestCapabilities) BizCtx() contract.BizCtxService {
	if s == nil {
		return nil
	}
	return s.bizCtx
}

// Cache returns no cache service for root plugin facade tests.
func (s *rootTestCapabilities) Cache() contract.CacheService { return nil }

// Config returns no plugin configuration service for root plugin facade tests.
func (s *rootTestCapabilities) Config() contract.ConfigService { return nil }

// ForPlugin returns a plugin-bound capability view for provider construction.
func (s *rootTestCapabilities) ForPlugin(_ string) capability.Services {
	if s == nil {
		return nil
	}
	return &rootTestCapabilities{
		bizCtx:          s.bizCtx,
		pluginLifecycle: s.pluginLifecycle,
	}
}

// HostConfig returns no public host configuration service for root plugin facade tests.
func (s *rootTestCapabilities) HostConfig() contract.HostConfigService { return nil }

// I18n returns no translation service for root plugin facade tests.
func (s *rootTestCapabilities) I18n() contract.I18nService { return nil }

// Manifest returns no manifest resource service for root plugin facade tests.
func (s *rootTestCapabilities) Manifest() contract.ManifestService { return nil }

// Notify returns no notification service for root plugin facade tests.
func (s *rootTestCapabilities) Notify() contract.NotifyService { return nil }

// Org returns the default organization capability fallback service.
func (s *rootTestCapabilities) Org() orgcapsvc.Service {
	return orgcapsvc.New(nil)
}

// PluginLifecycle returns nil-tolerant lifecycle operations for tenant provider code.
func (s *rootTestCapabilities) PluginLifecycle() contract.PluginLifecycleService {
	if s == nil {
		return nil
	}
	return s.pluginLifecycle
}

// PluginState returns no plugin-state service for root plugin facade tests.
func (s *rootTestCapabilities) PluginState() contract.PluginStateService { return nil }

// Route returns no dynamic-route metadata service for root plugin facade tests.
func (s *rootTestCapabilities) Route() contract.RouteService { return nil }

// Session returns no online-session service for root plugin facade tests.
func (s *rootTestCapabilities) Session() contract.SessionService { return nil }

// Tenant returns the default tenant capability fallback service.
func (s *rootTestCapabilities) Tenant() tenantcapsvc.Service {
	if s == nil {
		return tenantcapsvc.New(nil, nil)
	}
	return tenantcapsvc.New(nil, s.bizCtx)
}

// TenantFilter returns no tenant-filter service for root plugin facade tests.
func (s *rootTestCapabilities) TenantFilter() contract.TenantFilterService { return nil }

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
