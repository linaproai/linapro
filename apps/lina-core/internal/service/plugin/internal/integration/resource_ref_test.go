// This file covers resource-reference synchronization behaviors owned by integration.

package integration_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/internal/service/startupstats"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// TestSyncPluginResourceReferencesRevivesSoftDeletedRows verifies that sync can
// revive previously soft-deleted governance rows instead of colliding on insert.
func TestSyncPluginResourceReferencesRevivesSoftDeletedRows(t *testing.T) {
	services := testutil.NewServices()
	service := services.Integration
	ctx := context.Background()

	pluginID := "plugin-dynamic-ref-revive"
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		pluginID,
		"Runtime Ref Revive Plugin",
		"v0.9.0",
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-plugin-dynamic-ref-revive.sql", Content: "SELECT 1;"},
		},
		nil,
	)

	manifest := &catalog.Manifest{
		ID:           pluginID,
		Name:         "Runtime Ref Revive Plugin",
		Version:      "v0.9.0",
		Type:         catalog.TypeDynamic.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}
	if err := services.Catalog.ValidateManifest(manifest, manifest.ManifestPath); err != nil {
		t.Fatalf("expected dynamic manifest to be valid, got error: %v", err)
	}

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected plugin manifest sync to succeed, got error: %v", err)
	}

	release, err := services.Catalog.GetRelease(ctx, pluginID, manifest.Version)
	if err != nil {
		t.Fatalf("expected plugin release to exist, got error: %v", err)
	}
	if release == nil {
		t.Fatalf("expected plugin release to be created")
	}

	if _, err = dao.SysPluginResourceRef.Ctx(ctx).
		Where(do.SysPluginResourceRef{
			PluginId:  pluginID,
			ReleaseId: release.Id,
		}).
		Delete(); err != nil {
		t.Fatalf("expected resource refs to be soft-deleted, got error: %v", err)
	}

	if err = service.SyncPluginResourceReferences(ctx, manifest); err != nil {
		t.Fatalf("expected sync to revive soft-deleted rows without duplicate-key errors, got error: %v", err)
	}

	activeRefs, err := service.ListPluginResourceRefs(ctx, pluginID, release.Id)
	if err != nil {
		t.Fatalf("expected resource refs to be queryable, got error: %v", err)
	}
	if len(activeRefs) == 0 {
		t.Fatalf("expected revived resource refs to exist")
	}

	for _, item := range activeRefs {
		if item == nil {
			continue
		}
		if item.DeletedAt != nil {
			t.Fatalf("expected revived resource ref %s/%s to be active, got deleted_at=%v", item.ResourceType, item.ResourceKey, item.DeletedAt)
		}
	}
}

// TestSyncPluginResourceReferencesNoopSkipsWrites verifies no-op resource
// projection sync avoids writes when the startup snapshot already matches.
func TestSyncPluginResourceReferencesNoopSkipsWrites(t *testing.T) {
	services := testutil.NewServices()
	service := services.Integration
	ctx := context.Background()

	pluginID := "plugin-resource-ref-noop"
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		pluginID,
		"Runtime Ref Noop Plugin",
		"v0.9.1",
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-plugin-resource-ref-noop.sql", Content: "SELECT 1;"},
		},
		nil,
	)

	manifest := &catalog.Manifest{
		ID:           pluginID,
		Name:         "Runtime Ref Noop Plugin",
		Version:      "v0.9.1",
		Type:         catalog.TypeDynamic.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}
	if err := services.Catalog.ValidateManifest(manifest, manifest.ManifestPath); err != nil {
		t.Fatalf("expected dynamic manifest to be valid, got error: %v", err)
	}

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected plugin manifest sync to succeed, got error: %v", err)
	}
	if err := service.SyncPluginResourceReferences(ctx, manifest); err != nil {
		t.Fatalf("expected initial resource ref sync to succeed, got error: %v", err)
	}

	collector := startupstats.New()
	startupCtx := startupstats.WithCollector(ctx, collector)
	var err error
	startupCtx, err = services.Catalog.WithStartupDataSnapshot(startupCtx)
	if err != nil {
		t.Fatalf("build catalog startup snapshot: %v", err)
	}
	startupCtx, err = services.Integration.WithStartupDataSnapshot(startupCtx)
	if err != nil {
		t.Fatalf("build integration startup snapshot: %v", err)
	}

	sqls, logs, err := captureSQLDuring(t, startupCtx, func(ctx context.Context) error {
		return service.SyncPluginResourceReferences(ctx, manifest)
	})
	if err != nil {
		t.Fatalf("expected no-op resource ref sync to succeed, got error: %v", err)
	}
	assertNoMutationSQL(t, sqls)
	assertNoMutationSQL(t, logs)

	snapshot := collector.Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterPluginResourceSyncNoop); got != 1 {
		t.Fatalf("expected one no-op resource sync, got %d", got)
	}
	if got := snapshot.CounterValue(startupstats.CounterPluginResourceSyncChanged); got != 0 {
		t.Fatalf("expected no changed resource sync, got %d", got)
	}
}

// assertNoMutationSQL fails when captured SQL contains writes or transaction
// statements that should be absent from no-op startup sync paths.
func assertNoMutationSQL(t *testing.T, sqls []string) {
	t.Helper()

	for _, sql := range sqls {
		normalized := strings.ToUpper(strings.TrimSpace(sql))
		for _, keyword := range []string{"INSERT ", "UPDATE ", "DELETE ", "BEGIN", "COMMIT", "ROLLBACK"} {
			if strings.Contains(normalized, keyword) {
				t.Fatalf("expected no mutation or transaction SQL, got %q from %#v", sql, sqls)
			}
		}
	}
}

// captureSQLDuring executes fn while capturing GoFrame SQL statements and debug
// log lines so tests can detect transactions that CatchSQL does not expose.
func captureSQLDuring(
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
