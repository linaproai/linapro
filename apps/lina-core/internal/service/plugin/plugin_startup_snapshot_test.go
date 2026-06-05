// This file verifies root plugin startup snapshot reuse.

package plugin

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/model"
	i18nsvc "lina-core/internal/service/i18n"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/startupstats"
)

// TestWithStartupDataSnapshotReusesCatalogAndIntegrationSnapshots verifies one
// startup context does not rebuild equivalent plugin snapshots repeatedly.
func TestWithStartupDataSnapshotReusesCatalogAndIntegrationSnapshots(t *testing.T) {
	ctx := startupstats.WithCollector(context.Background(), startupstats.New())
	service := newTestService()

	startupCtx, err := service.WithStartupDataSnapshot(ctx)
	if err != nil {
		t.Fatalf("build first startup snapshot: %v", err)
	}
	startupCtx, err = service.WithStartupDataSnapshot(startupCtx)
	if err != nil {
		t.Fatalf("reuse second startup snapshot: %v", err)
	}
	if _, err = service.ReadOnlyList(startupCtx); err != nil {
		t.Fatalf("read plugin list with startup snapshots: %v", err)
	}

	snapshot := startupstats.FromContext(startupCtx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterCatalogSnapshotBuilds); got != 1 {
		t.Fatalf("expected one catalog snapshot build, got %d", got)
	}
	if got := snapshot.CounterValue(startupstats.CounterIntegrationSnapshotBuilds); got != 1 {
		t.Fatalf("expected one integration snapshot build, got %d", got)
	}
}

// TestReadOnlyListOnlyBuildsCatalogSnapshot verifies management read paths do
// not load integration snapshots that are only needed by startup sync.
func TestReadOnlyListOnlyBuildsCatalogSnapshot(t *testing.T) {
	ctx := startupstats.WithCollector(context.Background(), startupstats.New())
	service := newTestService()

	if _, err := service.ReadOnlyList(ctx); err != nil {
		t.Fatalf("read plugin list with catalog startup snapshot: %v", err)
	}

	snapshot := startupstats.FromContext(ctx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterCatalogSnapshotBuilds); got != 1 {
		t.Fatalf("expected one catalog snapshot build, got %d", got)
	}
	if got := snapshot.CounterValue(startupstats.CounterIntegrationSnapshotBuilds); got != 0 {
		t.Fatalf("expected no integration snapshot build, got %d", got)
	}
}

// TestDependencySnapshotCacheReusesCatalogSnapshot verifies repeated dependency
// checks in one list projection reuse request-local dependency snapshots.
func TestDependencySnapshotCacheReusesCatalogSnapshot(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = startupstats.WithCollector(context.Background(), startupstats.New())
		pluginID = "plugin-dev-source-dependency-cache"
	)

	createTestSourceDependencyPlugin(t, pluginID, "Source Dependency Cache", "v0.1.0", "")
	cleanupTestPluginIDs(t, context.Background(), pluginID)

	readCtx := service.WithDependencySnapshotCache(ctx)
	first, err := service.buildDependencySnapshots(readCtx, nil)
	if err != nil {
		t.Fatalf("build first dependency snapshots: %v", err)
	}
	firstSnapshot := findDependencySnapshotForTest(first, pluginID)
	if firstSnapshot == nil {
		t.Fatalf("expected first dependency snapshot for %s", pluginID)
	}
	firstSnapshot.Name = "mutated by caller"

	second, err := service.buildDependencySnapshots(readCtx, nil)
	if err != nil {
		t.Fatalf("build second dependency snapshots: %v", err)
	}
	secondSnapshot := findDependencySnapshotForTest(second, pluginID)
	if secondSnapshot == nil {
		t.Fatalf("expected second dependency snapshot for %s", pluginID)
	}
	if secondSnapshot.Name == firstSnapshot.Name {
		t.Fatalf("expected cached dependency snapshots to be cloned, got mutated name %q", secondSnapshot.Name)
	}

	snapshot := startupstats.FromContext(ctx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterCatalogSnapshotBuilds); got != 1 {
		t.Fatalf("expected one catalog snapshot build with dependency cache, got %d", got)
	}
}

// TestManagementListCacheAvoidsRepeatedManifestScans verifies the management
// list read model is reused until an explicit plugin-runtime invalidation.
func TestManagementListCacheAvoidsRepeatedManifestScans(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = startupstats.WithCollector(context.Background(), startupstats.New())
		pluginID = "plugin-dev-source-management-list-cache"
	)

	createTestSourceDependencyPlugin(t, pluginID, "Source Management List Cache", "v0.1.0", "")
	cleanupTestPluginIDs(t, context.Background(), pluginID)

	first, err := service.List(ctx, ListInput{})
	if err != nil {
		t.Fatalf("build first management list: %v", err)
	}
	if findPluginItem(first, pluginID) == nil {
		t.Fatalf("expected first management list to include %s", pluginID)
	}

	second, err := service.List(ctx, ListInput{})
	if err != nil {
		t.Fatalf("read cached management list: %v", err)
	}
	if findPluginItem(second, pluginID) == nil {
		t.Fatalf("expected cached management list to include %s", pluginID)
	}

	snapshot := startupstats.FromContext(ctx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterPluginScans); got != 1 {
		t.Fatalf("expected cached list to avoid repeated scans, got %d", got)
	}

	service.InvalidateManagementListCache(ctx, "test")
	third, err := service.List(ctx, ListInput{})
	if err != nil {
		t.Fatalf("rebuild invalidated management list: %v", err)
	}
	if findPluginItem(third, pluginID) == nil {
		t.Fatalf("expected rebuilt management list to include %s", pluginID)
	}

	snapshot = startupstats.FromContext(ctx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterPluginScans); got != 2 {
		t.Fatalf("expected invalidated list to rescan once, got %d", got)
	}
}

// TestRuntimeCacheChangeInvalidatesManagementList verifies lifecycle cache
// publications clear the plugin-management list read model.
func TestRuntimeCacheChangeInvalidatesManagementList(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = startupstats.WithCollector(context.Background(), startupstats.New())
		pluginID = "plugin-dev-source-management-list-runtime-invalidate"
	)

	createTestSourceDependencyPlugin(t, pluginID, "Source Management List Runtime Invalidate", "v0.1.0", "")
	cleanupTestPluginIDs(t, context.Background(), pluginID)

	if _, err := service.List(ctx, ListInput{}); err != nil {
		t.Fatalf("build management list: %v", err)
	}
	if _, err := service.List(ctx, ListInput{}); err != nil {
		t.Fatalf("read cached management list: %v", err)
	}
	if _, err := service.markRuntimeCacheChanged(ctx, "test_runtime_cache_changed"); err != nil {
		t.Fatalf("mark runtime cache changed: %v", err)
	}
	if _, err := service.List(ctx, ListInput{}); err != nil {
		t.Fatalf("rebuild after runtime cache change: %v", err)
	}

	snapshot := startupstats.FromContext(ctx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterPluginScans); got != 2 {
		t.Fatalf("expected runtime cache change to invalidate list, got %d scans", got)
	}
}

// TestPrewarmManagementListPopulatesCache verifies startup prewarm fills the
// same complete read model later consumed by management list requests.
func TestPrewarmManagementListPopulatesCache(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = startupstats.WithCollector(context.Background(), startupstats.New())
		pluginID = "plugin-dev-source-management-list-prewarm"
	)

	createTestSourceDependencyPlugin(t, pluginID, "Source Management List Prewarm", "v0.1.0", "")
	cleanupTestPluginIDs(t, context.Background(), pluginID)

	if err := service.PrewarmManagementList(ctx); err != nil {
		t.Fatalf("prewarm management list: %v", err)
	}
	out, err := service.List(ctx, ListInput{ID: pluginID})
	if err != nil {
		t.Fatalf("read prewarmed management list: %v", err)
	}
	if len(out.List) != 1 || out.List[0] == nil || out.List[0].Id != pluginID {
		t.Fatalf("expected prewarmed filtered list for %s, got %#v", pluginID, out)
	}

	snapshot := startupstats.FromContext(ctx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterPluginScans); got != 1 {
		t.Fatalf("expected prewarm plus list to scan once, got %d", got)
	}
}

// TestManagementListCacheIsLocaleScoped verifies localized plugin metadata
// cannot leak from startup prewarm or another request locale.
func TestManagementListCacheIsLocaleScoped(t *testing.T) {
	var (
		service   = newTestService()
		baseCtx   = context.Background()
		defaultID = "plugin-dev-source-management-list-default-locale"
		englishID = "plugin-dev-source-management-list-english-locale"
	)

	createTestSourceDependencyPlugin(t, defaultID, "Source Management List Default Locale", "v0.1.0", "")
	createTestSourceDependencyPlugin(t, englishID, "Source Management List English Locale", "v0.1.0", "")
	cleanupTestPluginIDs(t, context.Background(), defaultID, englishID)

	if _, err := service.List(baseCtx, ListInput{ID: defaultID}); err != nil {
		t.Fatalf("build default-locale management list: %v", err)
	}

	englishCtx := context.WithValue(
		context.Background(),
		gctx.StrKey("BizCtx"),
		&model.Context{Locale: i18nsvc.EnglishLocale},
	)
	if _, err := service.List(englishCtx, ListInput{ID: englishID}); err != nil {
		t.Fatalf("build english-locale management list: %v", err)
	}
	baseKey, err := service.managementListCacheKey(baseCtx)
	if err != nil {
		t.Fatalf("build default-locale cache key: %v", err)
	}
	if _, ok := service.managementListCache.Get(baseKey); !ok {
		t.Fatalf("expected default-locale management list cache")
	}
	englishKey, err := service.managementListCacheKey(englishCtx)
	if err != nil {
		t.Fatalf("build english-locale cache key: %v", err)
	}
	if _, ok := service.managementListCache.Get(englishKey); !ok {
		t.Fatalf("expected english-locale management list cache")
	}
}

// findDependencySnapshotForTest returns one dependency snapshot by plugin ID.
func findDependencySnapshotForTest(items []*plugindep.PluginSnapshot, pluginID string) *plugindep.PluginSnapshot {
	for _, item := range items {
		if item != nil && item.ID == pluginID {
			return item
		}
	}
	return nil
}
