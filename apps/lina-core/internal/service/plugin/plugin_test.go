// This file keeps root-package test bootstrap and shared helpers for plugin facade tests.

package plugin

import (
	"context"
	"encoding/base64"
	"sync"
	"testing"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/session"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/pkg/pluginbridge"
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
		service, err := New(topology, configProvider, bizCtxProvider, cacheCoordSvc, i18nSvc, session.NewDBStore(), coordSvc.Lock())
		if err != nil {
			panic(err)
		}
		serviceImpl := service.(*serviceImpl)
		serviceImpl.SetTenantCapability(tenantcapsvc.New(serviceImpl, bizCtxProvider))
		return serviceImpl
	}
	i18nSvc := i18nsvc.New(bizCtxProvider, configProvider, cacheCoordSvc)
	service, err := New(topology, configProvider, bizCtxProvider, cacheCoordSvc, i18nSvc, session.NewDBStore(), nil)
	if err != nil {
		panic(err)
	}
	serviceImpl := service.(*serviceImpl)
	serviceImpl.SetTenantCapability(tenantcapsvc.New(serviceImpl, bizCtxProvider))
	return serviceImpl
}

// TestNewRequiresExplicitRuntimeDependencies verifies the root plugin service
// returns a construction error when callers omit critical runtime dependencies.
func TestNewRequiresExplicitRuntimeDependencies(t *testing.T) {
	if _, err := New(nil, nil, nil, nil, nil, nil, nil); err == nil {
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
func (s *serviceImpl) executeDynamicRoute(ctx context.Context, manifest *catalog.Manifest, request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
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
			Path:          "index.html",
			ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>" + marker + "</body></html>")),
			ContentType:   "text/html; charset=utf-8",
		},
	}
}
