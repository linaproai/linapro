// This file covers root plugin facade topology behaviors that remain in the package root.

package plugin

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
)

// TestSingleNodeModeSkipsPluginNodeProjection verifies that single-node mode
// does not materialize per-node runtime projections for dynamic plugins.
func TestSingleNodeModeSkipsPluginNodeProjection(t *testing.T) {
	service := newTestService()
	ctx := context.Background()

	var (
		pluginID   = "plugin-dynamic-single-node"
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
		catalog.TypeDynamic.String(),
		catalog.InstalledYes,
		catalog.StatusEnabled,
	)
	if err != nil {
		t.Fatalf("expected governance snapshot build to succeed, got error: %v", err)
	}
	if snapshot == nil {
		t.Fatal("expected governance snapshot to exist")
	}
	if snapshot.NodeState != catalog.NodeStateEnabled.String() {
		t.Fatalf("expected governance snapshot to derive enabled node state, got %s", snapshot.NodeState)
	}
}

// TestClusterStartupManifestNoopSkipsNodeStateWrite verifies repeated startup
// manifest sync avoids rewriting the current-node projection when nothing changed.
func TestClusterStartupManifestNoopSkipsNodeStateWrite(t *testing.T) {
	var (
		ctx      = context.Background()
		pluginID = "plugin-source-cluster-node-noop"
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
	if _, err := service.catalogSvc.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected initial manifest sync to succeed, got error: %v", err)
	}

	startupCtx, err := service.WithStartupDataSnapshot(ctx)
	if err != nil {
		t.Fatalf("expected startup snapshot build to succeed, got error: %v", err)
	}
	sqls, logs, err := captureSQLDuringStartupTopologyTest(t, startupCtx, func(ctx context.Context) error {
		_, syncErr := service.catalogSvc.SyncManifest(ctx, manifest)
		return syncErr
	})
	if err != nil {
		t.Fatalf("expected no-op manifest sync to succeed, got error: %v", err)
	}
	assertNoNodeStateMutationSQL(t, sqls)
	assertNoNodeStateMutationSQL(t, logs)
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
