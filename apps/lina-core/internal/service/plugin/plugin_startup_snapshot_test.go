// This file verifies root plugin startup snapshot reuse.

package plugin

import (
	"context"
	"testing"

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

// findDependencySnapshotForTest returns one dependency snapshot by plugin ID.
func findDependencySnapshotForTest(items []*plugindep.PluginSnapshot, pluginID string) *plugindep.PluginSnapshot {
	for _, item := range items {
		if item != nil && item.ID == pluginID {
			return item
		}
	}
	return nil
}
