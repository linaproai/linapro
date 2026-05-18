// This file verifies root plugin startup snapshot reuse.

package plugin

import (
	"context"
	"testing"

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
