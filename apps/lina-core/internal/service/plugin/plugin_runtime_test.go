// This file covers root-facade runtime methods defined in plugin_runtime.go,
// including reconciliation and dynamic route execution scenarios.

package plugin

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/plugin/pluginhost"
)

// TestSingleNodeModeSkipsPluginNodeProjection verifies that single-node mode
// does not materialize per-node runtime projections for dynamic plugins.
func TestSingleNodeModeSkipsPluginNodeProjection(t *testing.T) {
	service := newTestService()
	ctx := context.Background()

	var (
		pluginID   = "plugin-dev-dynamic-single-node"
		pluginName = "Dynamic Single Node Plugin"
		version    = "v0.1.0"
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	testutil.CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		pluginName,
		version,
		buildVersionedRuntimeFrontendAssets("single-node"),
		nil,
		nil,
	)

	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected single-node install to succeed, got error: %v", err)
	}
	if err := service.Enable(ctx, pluginID); err != nil {
		t.Fatalf("expected single-node enable to succeed, got error: %v", err)
	}

	nodeStateCount, err := dao.SysPluginNodeState.Ctx(ctx).
		Where(do.SysPluginNodeState{PluginId: pluginID}).
		Count()
	if err != nil {
		t.Fatalf("expected plugin node-state count query to succeed, got error: %v", err)
	}
	if nodeStateCount != 0 {
		t.Fatalf("expected single-node mode to skip node-state projection rows, got %d", nodeStateCount)
	}

	snapshot, err := service.buildPluginGovernanceSnapshot(
		ctx,
		pluginID,
		version,
		plugintypes.TypeDynamic.String(),
		plugintypes.InstalledYes,
		plugintypes.StatusEnabled,
	)
	if err != nil {
		t.Fatalf("expected governance snapshot build to succeed, got error: %v", err)
	}
	if snapshot == nil {
		t.Fatal("expected governance snapshot to exist")
	}
	if snapshot.NodeState != plugintypes.NodeStateEnabled.String() {
		t.Fatalf("expected governance snapshot to derive enabled node state, got %s", snapshot.NodeState)
	}
}

// TestClusterStartupManifestNoopSkipsNodeStateWrite verifies repeated startup
// manifest sync avoids rewriting the current-node projection when nothing changed.
func TestClusterStartupManifestNoopSkipsNodeStateWrite(t *testing.T) {
	var (
		ctx      = context.Background()
		pluginID = "plugin-dev-source-cluster-node-noop"
		version  = "v0.1.0"
		topology = &testTopology{
			enabled: true,
			primary: true,
			nodeID:  "startup-node-noop",
		}
		service = newTestServiceWithTopology(topology)
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	testutil.WriteTestFile(
		t,
		manifestPath,
		"id: "+pluginID+"\n"+
			"name: Source Cluster Node Noop Plugin\n"+
			"version: "+version+"\n"+
			"type: source\n"+
			"scope_nature: tenant_aware\n"+
			"supports_multi_tenant: false\n"+
			"default_install_mode: global\n",
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest := &catalog.Manifest{}
	if err := service.catalogSvc.LoadManifestFromYAML(manifestPath, manifest); err != nil {
		t.Fatalf("expected source manifest load to succeed, got error: %v", err)
	}
	if _, err := service.storeSvc.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected initial manifest sync to succeed, got error: %v", err)
	}

	startupCtx, err := service.WithStartupDataSnapshot(ctx)
	if err != nil {
		t.Fatalf("expected startup snapshot build to succeed, got error: %v", err)
	}
	sqls, logs, err := captureSQLDuringStartupTopologyTest(t, startupCtx, func(ctx context.Context) error {
		_, syncErr := service.storeSvc.SyncManifest(ctx, manifest)
		return syncErr
	})
	if err != nil {
		t.Fatalf("expected no-op manifest sync to succeed, got error: %v", err)
	}
	assertNoNodeStateMutationSQL(t, sqls)
	assertNoNodeStateMutationSQL(t, logs)
}

// TestSourceProviderAvailabilityFollowsEnabledSnapshot verifies provider
// declarations remain inert until their owning source plugin is platform-enabled.
func TestSourceProviderAvailabilityFollowsEnabledSnapshot(t *testing.T) {
	var (
		ctx      = bizctxcap.WithCurrentContext(context.Background(), bizctxcap.CurrentContext{TenantID: 0, PlatformBypass: true})
		pluginID = "plugin-dev-source-capability-revision"
		service  = newTestServiceWithTopology(&testTopology{
			enabled: true,
			primary: true,
			nodeID:  "capability-revision-node",
		})
	)
	cleanupTestPluginIDs(t, ctx, pluginID)

	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(fstest.MapFS{
		"plugin.yaml": &fstest.MapFile{Data: []byte(
			"id: " + pluginID + "\n" +
				"name: Runtime Revision Provider\n" +
				"version: v0.1.0\n" +
				"type: source\n" +
				"scope_nature: tenant_aware\n" +
				"supports_multi_tenant: false\n" +
				"default_install_mode: global\n",
		)},
	})
	cleanup, err := pluginhost.RegisterSourcePluginForTest(plugin)
	if err != nil {
		t.Fatalf("register source plugin fixture failed: %v", err)
	}
	t.Cleanup(cleanup)

	tenantManager := tenantspi.NewManager()
	if err = tenantManager.RegisterFactory(pluginID, func(
		context.Context,
		tenantspi.ProviderEnv,
	) (tenantspi.Provider, error) {
		return capabilityRevisionProvider{}, nil
	}); err != nil {
		t.Fatalf("register tenant provider factory failed: %v", err)
	}

	if _, err = service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("install source provider plugin failed: %v", err)
	}
	tenantSvc := tenantspi.New(tenantManager, capabilityRevisionRuntime{service: service, pluginID: pluginID}, nil)
	status := tenantSvc.Status(ctx)
	if status.Available || status.ActiveProvider == pluginID {
		t.Fatalf("expected installed-but-disabled provider unavailable, got %#v", status)
	}

	if err = service.Enable(ctx, pluginID); err != nil {
		t.Fatalf("enable source provider plugin failed: %v", err)
	}
	status = tenantSvc.Status(ctx)
	if !status.Available || status.ActiveProvider != pluginID {
		t.Fatalf("expected tenant provider active for %s, got %#v", pluginID, status)
	}

	if err = service.Disable(ctx, pluginID); err != nil {
		t.Fatalf("disable source provider plugin failed: %v", err)
	}
	status = tenantSvc.Status(ctx)
	if status.Available || status.ActiveProvider == pluginID {
		t.Fatalf("expected disabled provider unavailable, got %#v", status)
	}
}

// TestDynamicPluginRuntimeUpgradeKeepsPreviousReleaseFrontendAssets verifies
// explicit runtime upgrade keeps archived frontend bundles available for drain
// and rollback.
func TestDynamicPluginRuntimeUpgradeKeepsPreviousReleaseFrontendAssets(t *testing.T) {
	service := newTestService()
	ctx := context.Background()

	pluginID := "plugin-dev-dynamic-upgrade"
	pluginName := "Dynamic Upgrade Plugin"
	versionOne := "v0.1.0"
	versionTwo := "v0.2.0"

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	testutil.CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		pluginName,
		versionOne,
		buildVersionedRuntimeFrontendAssets("version-one"),
		nil,
		nil,
	)

	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected initial install to succeed, got error: %v", err)
	}
	if err := service.Enable(ctx, pluginID); err != nil {
		t.Fatalf("expected initial enable to succeed, got error: %v", err)
	}

	registryBeforeUpgrade, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	if registryBeforeUpgrade == nil {
		t.Fatal("expected registry row to exist after initial enable")
	}

	testutil.CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		pluginName,
		versionTwo,
		buildVersionedRuntimeFrontendAssets("version-two"),
		nil,
		nil,
	)
	targetManifest, err := service.loadRuntimePluginManifestFromArtifact(filepath.Join(testutil.TestDynamicStorageDir(), pluginID+".wasm"))
	if err != nil {
		t.Fatalf("expected target dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, targetManifest); err != nil {
		t.Fatalf("expected target manifest sync to succeed, got error: %v", err)
	}

	if _, err = service.ExecuteRuntimeUpgrade(ctx, pluginID, RuntimeUpgradeOptions{Confirmed: true}); err != nil {
		t.Fatalf("expected explicit runtime upgrade to succeed, got error: %v", err)
	}

	registryAfterUpgrade, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected upgraded registry lookup to succeed, got error: %v", err)
	}
	if registryAfterUpgrade == nil {
		t.Fatal("expected upgraded registry row to exist")
	}
	if registryAfterUpgrade.Version != versionTwo {
		t.Fatalf("expected active version %s after upgrade, got %s", versionTwo, registryAfterUpgrade.Version)
	}
	if registryAfterUpgrade.Generation <= registryBeforeUpgrade.Generation {
		t.Fatalf("expected generation to advance after upgrade, before=%d after=%d", registryBeforeUpgrade.Generation, registryAfterUpgrade.Generation)
	}
	if registryAfterUpgrade.ReleaseId == registryBeforeUpgrade.ReleaseId {
		t.Fatalf("expected active release id to change after upgrade, got %d", registryAfterUpgrade.ReleaseId)
	}

	oldAsset, err := service.ResolveRuntimeFrontendAsset(ctx, pluginID, versionOne, "index.html")
	if err != nil {
		t.Fatalf("expected previous release asset to stay resolvable, got error: %v", err)
	}
	if !strings.Contains(string(oldAsset.Content), "version-one") {
		t.Fatalf("expected previous release asset content to contain version-one marker, got %s", string(oldAsset.Content))
	}

	newAsset, err := service.ResolveRuntimeFrontendAsset(ctx, pluginID, versionTwo, "index.html")
	if err != nil {
		t.Fatalf("expected new release asset to be resolvable, got error: %v", err)
	}
	if !strings.Contains(string(newAsset.Content), "version-two") {
		t.Fatalf("expected new release asset content to contain version-two marker, got %s", string(newAsset.Content))
	}

	releaseOne, err := service.getPluginRelease(ctx, pluginID, versionOne)
	if err != nil {
		t.Fatalf("expected previous release lookup to succeed, got error: %v", err)
	}
	releaseTwo, err := service.getPluginRelease(ctx, pluginID, versionTwo)
	if err != nil {
		t.Fatalf("expected new release lookup to succeed, got error: %v", err)
	}
	if releaseOne == nil || releaseOne.Status != plugintypes.ReleaseStatusInstalled.String() {
		t.Fatalf("expected previous release to remain installed for drain/rollback, got %#v", releaseOne)
	}
	if releaseTwo == nil || releaseTwo.Status != plugintypes.ReleaseStatusActive.String() {
		t.Fatalf("expected new release to become active, got %#v", releaseTwo)
	}
}

// TestDynamicPluginRuntimeUpgradeFailureRollsBackStableRelease verifies that a
// failed explicit runtime upgrade restores the previous active release and its
// governance projection.
func TestDynamicPluginRuntimeUpgradeFailureRollsBackStableRelease(t *testing.T) {
	service := newTestService()
	ctx := context.Background()

	pluginID := "plugin-dev-dynamic-upgrade-failed"
	pluginName := "Dynamic Upgrade Failure Plugin"
	versionOne := "v0.1.0"
	versionTwo := "v0.2.0"
	permissionOne := pluginID + ":review:view"
	permissionTwo := pluginID + ":review:inspect"

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsMenusAndBackendContracts(
		t,
		pluginID,
		pluginName,
		versionOne,
		buildVersionedRuntimeFrontendAssets("stable-version"),
		runtimeRoutePermissionMenus(pluginID, pluginName, versionOne),
		nil,
		nil,
		[]*protocol.RouteContract{
			{
				Path:        "/api/v1/review-summary",
				Method:      http.MethodGet,
				Access:      protocol.AccessLogin,
				Permission:  permissionOne,
				RequestType: "ReviewSummaryReq",
			},
		},
		&protocol.BridgeSpec{
			ABIVersion:     protocol.ABIVersionV1,
			RuntimeKind:    protocol.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   protocol.CodecProtobuf,
			ResponseCodec:  protocol.CodecProtobuf,
		},
	)

	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected initial install to succeed, got error: %v", err)
	}
	if err := service.Enable(ctx, pluginID); err != nil {
		t.Fatalf("expected initial enable to succeed, got error: %v", err)
	}

	registryBeforeFailure, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	if registryBeforeFailure == nil {
		t.Fatal("expected registry row before failed upgrade")
	}

	testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsMenusAndBackendContracts(
		t,
		pluginID,
		pluginName,
		versionTwo,
		buildVersionedRuntimeFrontendAssets("broken-version"),
		runtimeRoutePermissionMenus(pluginID, pluginName, versionTwo),
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-dev-dynamic-upgrade-failed.sql",
				Content: "THIS IS NOT VALID SQL;",
			},
		},
		nil,
		[]*protocol.RouteContract{
			{
				Path:        "/api/v1/review-summary",
				Method:      http.MethodGet,
				Access:      protocol.AccessLogin,
				Permission:  permissionTwo,
				RequestType: "ReviewSummaryReq",
			},
		},
		&protocol.BridgeSpec{
			ABIVersion:     protocol.ABIVersionV1,
			RuntimeKind:    protocol.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   protocol.CodecProtobuf,
			ResponseCodec:  protocol.CodecProtobuf,
		},
	)
	targetManifest, err := service.loadRuntimePluginManifestFromArtifact(filepath.Join(testutil.TestDynamicStorageDir(), pluginID+".wasm"))
	if err != nil {
		t.Fatalf("expected failed target dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, targetManifest); err != nil {
		t.Fatalf("expected failed target manifest sync to succeed, got error: %v", err)
	}

	if _, err = service.ExecuteRuntimeUpgrade(ctx, pluginID, RuntimeUpgradeOptions{Confirmed: true}); err == nil {
		t.Fatal("expected failed explicit runtime upgrade to return an error")
	}

	registryAfterFailure, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after failed upgrade to succeed, got error: %v", err)
	}
	if registryAfterFailure == nil {
		t.Fatal("expected registry row after failed upgrade")
	}
	if registryAfterFailure.Version != versionOne {
		t.Fatalf("expected active version to stay at %s after rollback, got %s", versionOne, registryAfterFailure.Version)
	}
	if registryAfterFailure.ReleaseId != registryBeforeFailure.ReleaseId {
		t.Fatalf("expected active release id to stay unchanged after rollback, before=%d after=%d", registryBeforeFailure.ReleaseId, registryAfterFailure.ReleaseId)
	}
	if registryAfterFailure.Generation != registryBeforeFailure.Generation {
		t.Fatalf("expected generation to stay unchanged after rollback, before=%d after=%d", registryBeforeFailure.Generation, registryAfterFailure.Generation)
	}
	if registryAfterFailure.DesiredState != plugintypes.HostStateEnabled.String() || registryAfterFailure.CurrentState != plugintypes.HostStateEnabled.String() {
		t.Fatalf("expected registry to restore enabled stable state after rollback, got desired=%s current=%s", registryAfterFailure.DesiredState, registryAfterFailure.CurrentState)
	}

	stableAsset, err := service.ResolveRuntimeFrontendAsset(ctx, pluginID, versionOne, "index.html")
	if err != nil {
		t.Fatalf("expected stable release asset to remain resolvable after rollback, got error: %v", err)
	}
	if !strings.Contains(string(stableAsset.Content), "stable-version") {
		t.Fatalf("expected stable release asset content to be preserved, got %s", string(stableAsset.Content))
	}

	stablePermissionMenu, err := testutil.QueryMenuByKey(ctx, integration.BuildDynamicRoutePermissionMenuKey(pluginID, permissionOne))
	if err != nil {
		t.Fatalf("expected stable permission menu query to succeed after rollback, got error: %v", err)
	}
	if stablePermissionMenu == nil {
		t.Fatal("expected stable permission menu to be restored after rollback")
	}
	failedPermissionMenu, err := testutil.QueryMenuByKey(ctx, integration.BuildDynamicRoutePermissionMenuKey(pluginID, permissionTwo))
	if err != nil {
		t.Fatalf("expected failed permission menu query to succeed after rollback, got error: %v", err)
	}
	if failedPermissionMenu != nil {
		t.Fatal("expected failed release permission menu to be cleaned up after rollback")
	}

	failedRelease, err := service.getPluginRelease(ctx, pluginID, versionTwo)
	if err != nil {
		t.Fatalf("expected failed release lookup to succeed, got error: %v", err)
	}
	if failedRelease == nil || failedRelease.Status != plugintypes.ReleaseStatusFailed.String() {
		t.Fatalf("expected failed release status to be marked failed, got %#v", failedRelease)
	}
	if _, err = service.ResolveRuntimeFrontendAsset(ctx, pluginID, versionTwo, "index.html"); err == nil {
		t.Fatal("expected failed release asset to stay hidden from runtime frontend resolution")
	}
}

// TestDynamicPluginUninstallFailureRestoresStableRegistryFlags verifies that
// uninstall rollback restores the previously active registry flags.
func TestDynamicPluginUninstallFailureRestoresStableRegistryFlags(t *testing.T) {
	service := newTestService()
	ctx := context.Background()

	pluginID := "plugin-dev-dynamic-uninstall-failed"
	pluginName := "Dynamic Uninstall Failure Plugin"
	version := "v0.1.0"

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	testutil.CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		pluginName,
		version,
		buildVersionedRuntimeFrontendAssets("stable-version"),
		nil,
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-dev-dynamic-uninstall-failed.sql",
				Content: "THIS IS NOT VALID SQL;",
			},
		},
	)

	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected initial install to succeed, got error: %v", err)
	}
	if err := service.Enable(ctx, pluginID); err != nil {
		t.Fatalf("expected initial enable to succeed, got error: %v", err)
	}

	registryBeforeFailure, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup before failed uninstall to succeed, got error: %v", err)
	}
	if registryBeforeFailure == nil {
		t.Fatal("expected registry row before failed uninstall")
	}

	if err = service.Uninstall(ctx, pluginID, UninstallOptions{PurgeStorageData: true}); err == nil {
		t.Fatal("expected failed uninstall to return an error")
	}

	registryAfterFailure, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after failed uninstall to succeed, got error: %v", err)
	}
	if registryAfterFailure == nil {
		t.Fatal("expected registry row after failed uninstall")
	}
	if registryAfterFailure.Installed != registryBeforeFailure.Installed {
		t.Fatalf("expected installed flag to be restored after rollback, before=%d after=%d", registryBeforeFailure.Installed, registryAfterFailure.Installed)
	}
	if registryAfterFailure.Status != registryBeforeFailure.Status {
		t.Fatalf("expected status flag to be restored after rollback, before=%d after=%d", registryBeforeFailure.Status, registryAfterFailure.Status)
	}
	if registryAfterFailure.ReleaseId != registryBeforeFailure.ReleaseId {
		t.Fatalf("expected release id to stay unchanged after uninstall rollback, before=%d after=%d", registryBeforeFailure.ReleaseId, registryAfterFailure.ReleaseId)
	}
	if registryAfterFailure.DesiredState != plugintypes.HostStateEnabled.String() || registryAfterFailure.CurrentState != plugintypes.HostStateEnabled.String() {
		t.Fatalf("expected registry to restore enabled stable state after uninstall rollback, got desired=%s current=%s", registryAfterFailure.DesiredState, registryAfterFailure.CurrentState)
	}
}

// TestDynamicPluginFollowerDefersUntilPrimaryReconciles verifies that follower
// nodes only persist desired state until the primary reconciles the runtime.
func TestDynamicPluginFollowerDefersUntilPrimaryReconciles(t *testing.T) {
	topology := &testTopology{
		enabled: true,
		primary: false,
		nodeID:  "follower-node",
	}
	service := newTestServiceWithTopology(topology)
	ctx := context.Background()

	pluginID := "plugin-dev-dynamic-follower"
	pluginName := "Dynamic Follower Plugin"
	versionOne := "v0.1.0"

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	testutil.CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		pluginName,
		versionOne,
		buildVersionedRuntimeFrontendAssets("follower-version"),
		nil,
		nil,
	)

	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected follower-side install request to persist desired state, got error: %v", err)
	}

	registryBeforePrimary, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected follower registry lookup to succeed, got error: %v", err)
	}
	if registryBeforePrimary == nil {
		t.Fatal("expected registry row to exist on follower")
	}
	if registryBeforePrimary.Installed != plugintypes.InstalledNo {
		t.Fatalf("expected follower request to keep current install state unchanged, got installed=%d", registryBeforePrimary.Installed)
	}
	if registryBeforePrimary.DesiredState != plugintypes.HostStateInstalled.String() {
		t.Fatalf("expected follower request to persist desired installed state, got %s", registryBeforePrimary.DesiredState)
	}
	if registryBeforePrimary.CurrentState != plugintypes.HostStateUninstalled.String() {
		t.Fatalf("expected follower current state to remain uninstalled before primary reconciliation, got %s", registryBeforePrimary.CurrentState)
	}

	topology.SetPrimary(true)
	if err = service.ReconcileRuntimePlugins(ctx); err != nil {
		t.Fatalf("expected primary reconciliation to succeed, got error: %v", err)
	}

	registryAfterPrimary, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected primary registry lookup to succeed, got error: %v", err)
	}
	if registryAfterPrimary == nil {
		t.Fatal("expected registry row after primary reconciliation")
	}
	if registryAfterPrimary.Installed != plugintypes.InstalledYes {
		t.Fatalf("expected primary reconciliation to install plugin, got installed=%d", registryAfterPrimary.Installed)
	}
	if registryAfterPrimary.CurrentState != plugintypes.HostStateInstalled.String() {
		t.Fatalf("expected current state to converge to installed on primary, got %s", registryAfterPrimary.CurrentState)
	}
	if registryAfterPrimary.ReleaseId <= 0 {
		t.Fatalf("expected primary reconciliation to persist active release id, got %d", registryAfterPrimary.ReleaseId)
	}
}

// TestInstallSameVersionDynamicPluginRefreshesArchivedReleaseArtifact verifies
// that reinstalling the same version refreshes archived release content in place.
func TestInstallSameVersionDynamicPluginRefreshesArchivedReleaseArtifact(t *testing.T) {
	service := newTestService()
	ctx := context.Background()

	pluginID := "plugin-dev-dynamic-same-version-refresh"
	pluginName := "Dynamic Same Version Refresh Plugin"
	version := "v0.1.0"

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	initialRoutes := []*protocol.RouteContract{
		{
			Path:        "/api/v1/review-summary",
			Method:      http.MethodGet,
			Access:      protocol.AccessLogin,
			Permission:  pluginID + ":review:view",
			RequestType: "ReviewSummaryReq",
			Meta: map[string]string{
				"x-route-purpose": "review",
			},
		},
	}
	initialBridge := &protocol.BridgeSpec{
		ABIVersion:     protocol.ABIVersionV1,
		RuntimeKind:    protocol.RuntimeKindWasm,
		RouteExecution: true,
		RequestCodec:   protocol.CodecProtobuf,
		ResponseCodec:  protocol.CodecProtobuf,
		AllocExport:    protocol.DefaultGuestAllocExport,
		ExecuteExport:  protocol.DefaultGuestExecuteExport,
	}
	testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsMenusAndBackendContracts(
		t,
		pluginID,
		pluginName,
		version,
		buildVersionedRuntimeFrontendAssets("version-one"),
		runtimeRoutePermissionMenus(pluginID, pluginName, version),
		nil,
		nil,
		initialRoutes,
		initialBridge,
	)

	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected initial install to succeed, got error: %v", err)
	}
	if err := service.Enable(ctx, pluginID); err != nil {
		t.Fatalf("expected initial enable to succeed, got error: %v", err)
	}

	registryBeforeRefresh, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup before refresh to succeed, got error: %v", err)
	}
	if registryBeforeRefresh == nil {
		t.Fatal("expected registry row before same-version refresh")
	}
	releaseBeforeRefresh, err := service.getPluginRelease(ctx, pluginID, version)
	if err != nil {
		t.Fatalf("expected release lookup before refresh to succeed, got error: %v", err)
	}
	if releaseBeforeRefresh == nil {
		t.Fatal("expected release row before same-version refresh")
	}
	initialPackagePath := filepath.ToSlash(releaseBeforeRefresh.PackagePath)
	if initialPackagePath == "" {
		t.Fatal("expected initial same-version release to store an archived package path")
	}

	refreshedRoutes := []*protocol.RouteContract{
		{
			Path:        "/api/v1/review-summary",
			Method:      http.MethodGet,
			Access:      protocol.AccessLogin,
			Permission:  pluginID + ":review:inspect",
			RequestType: "ReviewSummaryReq",
			Meta: map[string]string{
				"x-route-purpose": "review",
			},
		},
	}
	testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsMenusAndBackendContracts(
		t,
		pluginID,
		pluginName,
		version,
		buildVersionedRuntimeFrontendAssets("version-two"),
		runtimeRoutePermissionMenus(pluginID, pluginName, version),
		nil,
		nil,
		refreshedRoutes,
		initialBridge,
	)
	service.catalogSvc.InvalidateManifestCache(pluginID)

	if _, err = service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected same-version refresh install to succeed, got error: %v", err)
	}

	registryAfterRefresh, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after refresh to succeed, got error: %v", err)
	}
	if registryAfterRefresh == nil {
		t.Fatal("expected registry row after same-version refresh")
	}
	if registryAfterRefresh.ReleaseId != registryBeforeRefresh.ReleaseId {
		t.Fatalf("expected same-version refresh to reuse active release id, before=%d after=%d", registryBeforeRefresh.ReleaseId, registryAfterRefresh.ReleaseId)
	}
	if registryAfterRefresh.Generation <= registryBeforeRefresh.Generation {
		t.Fatalf("expected same-version refresh to advance generation, before=%d after=%d", registryBeforeRefresh.Generation, registryAfterRefresh.Generation)
	}
	releaseAfterRefresh, err := service.getPluginRelease(ctx, pluginID, version)
	if err != nil {
		t.Fatalf("expected release lookup after refresh to succeed, got error: %v", err)
	}
	if releaseAfterRefresh == nil {
		t.Fatal("expected release row after same-version refresh")
	}
	refreshedPackagePath := filepath.ToSlash(releaseAfterRefresh.PackagePath)
	if refreshedPackagePath == initialPackagePath {
		t.Fatalf("expected same-version refresh to move archive path by checksum, still got %s", refreshedPackagePath)
	}
	if !strings.Contains(refreshedPackagePath, releaseAfterRefresh.Checksum) {
		t.Fatalf("expected refreshed package path %q to include checksum %q", refreshedPackagePath, releaseAfterRefresh.Checksum)
	}

	activeManifest, err := service.getActivePluginManifest(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected active manifest after refresh to load, got error: %v", err)
	}
	if activeManifest == nil || activeManifest.RuntimeArtifact == nil {
		t.Fatalf("expected active manifest runtime artifact after refresh, got %#v", activeManifest)
	}
	if !strings.Contains(refreshedPackagePath, activeManifest.RuntimeArtifact.Checksum) {
		t.Fatalf("expected refreshed package path %q to include active checksum %q", refreshedPackagePath, activeManifest.RuntimeArtifact.Checksum)
	}
	if activeManifest.RuntimeArtifact.Checksum != releaseAfterRefresh.Checksum {
		t.Fatalf("expected active manifest checksum %s to match release checksum %s", activeManifest.RuntimeArtifact.Checksum, releaseAfterRefresh.Checksum)
	}
	if len(activeManifest.Routes) != 1 || activeManifest.Routes[0].Permission != pluginID+":review:inspect" {
		t.Fatalf("expected active manifest routes to refresh with new permission, got %#v", activeManifest.Routes)
	}
	oldArchivePath := filepath.Join(testutil.TestDynamicStorageDir(), filepath.FromSlash(initialPackagePath))
	if _, statErr := os.Stat(oldArchivePath); !os.IsNotExist(statErr) {
		t.Fatalf("expected old same-version archive to be cleaned path=%s err=%v", oldArchivePath, statErr)
	}

	asset, err := service.ResolveRuntimeFrontendAsset(ctx, pluginID, version, "index.html")
	if err != nil {
		t.Fatalf("expected refreshed frontend asset to resolve, got error: %v", err)
	}
	if !strings.Contains(string(asset.Content), "version-two") {
		t.Fatalf("expected refreshed frontend asset to contain version-two marker, got %s", string(asset.Content))
	}
}

// runtimeRoutePermissionMenus returns the current plugin entry menu required
// for artifacts that declare dynamic route permissions.
func runtimeRoutePermissionMenus(pluginID string, pluginName string, version string) []*catalog.MenuSpec {
	return []*catalog.MenuSpec{
		{
			Key:       "plugin:" + pluginID + ":main-entry",
			Name:      pluginName,
			Path:      "/x-assets/" + pluginID + "/" + version + "/mount.js",
			Perms:     pluginID + ":view",
			Icon:      "ant-design:deployment-unit-outlined",
			Type:      catalog.MenuTypePage.String(),
			Sort:      -1,
			Component: "system/plugin/dynamic-page",
			Query:     map[string]interface{}{"pluginAccessMode": "embedded-mount"},
		},
	}
}

// assertNoNodeStateMutationSQL fails when captured SQL rewrites plugin node state.
func assertNoNodeStateMutationSQL(t *testing.T, sqls []string) {
	t.Helper()

	for _, sql := range sqls {
		normalized := strings.ToUpper(strings.TrimSpace(sql))
		if !strings.Contains(normalized, "SYS_PLUGIN_NODE_STATE") {
			continue
		}
		for _, keyword := range []string{"INSERT ", "UPDATE ", "DELETE "} {
			if strings.Contains(normalized, keyword) {
				t.Fatalf("expected no sys_plugin_node_state mutation SQL, got %q from %#v", sql, sqls)
			}
		}
	}
}

// captureSQLDuringStartupTopologyTest captures GoFrame SQL and debug log lines
// emitted by fn so no-op startup paths can assert write avoidance.
func captureSQLDuringStartupTopologyTest(
	t *testing.T,
	ctx context.Context,
	fn func(context.Context) error,
) ([]string, []string, error) {
	t.Helper()

	db := g.DB()
	previousDebug := db.GetDebug()
	previousLogger := db.GetLogger()
	captureLogger := glog.New()
	captureLogger.SetStdoutPrint(false)

	db.SetDebug(true)
	db.SetLogger(captureLogger)
	defer func() {
		db.SetLogger(previousLogger)
		db.SetDebug(previousDebug)
	}()

	var logs []string
	captureLogger.SetHandlers(func(ctx context.Context, in *glog.HandlerInput) {
		logs = append(logs, in.ValuesContent())
	})

	sqls, err := gdb.CatchSQL(ctx, fn)
	return sqls, logs, err
}

// capabilityRevisionProvider is a no-op tenant provider used by the
// lifecycle/runtime-revision integration test.
type capabilityRevisionProvider struct{}

// capabilityRevisionRuntime exposes only the provider plugin owned by this
// test so shared linapro-tenant-core state from broader Go runs cannot affect
// the provider activation assertions.
type capabilityRevisionRuntime struct {
	service  *serviceImpl
	pluginID string
}

// IsProviderEnabled delegates enablement for the test provider and hides every
// unrelated tenant provider registered in the process.
func (r capabilityRevisionRuntime) IsProviderEnabled(ctx context.Context, pluginID string) bool {
	return r.service != nil &&
		pluginID == r.pluginID &&
		r.service.IsProviderEnabled(ctx, pluginID)
}

// TenantProviderEnv returns the minimal environment required by the no-op test provider.
func (capabilityRevisionRuntime) TenantProviderEnv(string) tenantspi.ProviderEnv {
	return tenantspi.ProviderEnv{}
}

// ResolveTenant returns the platform tenant.
func (capabilityRevisionProvider) ResolveTenant(
	context.Context,
	*ghttp.Request,
) (*tenantcap.ResolverResult, error) {
	return &tenantcap.ResolverResult{
		TenantID: tenantcap.PLATFORM,
		Matched:  true,
	}, nil
}

// ValidateUserInTenant accepts every tenant validation request.
func (capabilityRevisionProvider) ValidateUserInTenant(
	context.Context,
	int,
	tenantcap.TenantID,
) error {
	return nil
}

// ListUserTenants returns no tenant memberships.
func (capabilityRevisionProvider) ListUserTenants(
	context.Context,
	int,
) ([]tenantcap.TenantInfo, error) {
	return []tenantcap.TenantInfo{}, nil
}

// SwitchTenant accepts every tenant switch request.
func (capabilityRevisionProvider) SwitchTenant(
	context.Context,
	int,
	tenantcap.TenantID,
) error {
	return nil
}
