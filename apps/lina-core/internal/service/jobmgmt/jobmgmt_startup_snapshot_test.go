// This file verifies scheduled-job startup snapshot reuse.

package jobmgmt

import (
	"context"
	"testing"

	"lina-core/internal/service/startupstats"
)

// TestWithStartupDataSnapshotReusesJobSnapshot verifies one startup context
// does not rebuild equivalent scheduled-job snapshots repeatedly.
func TestWithStartupDataSnapshotReusesJobSnapshot(t *testing.T) {
	ctx := startupstats.WithCollector(context.Background(), startupstats.New())
	svc := newTestService(t)

	startupCtx, err := svc.WithStartupDataSnapshot(ctx)
	if err != nil {
		t.Fatalf("build first job startup snapshot: %v", err)
	}
	startupCtx, err = svc.WithStartupDataSnapshot(startupCtx)
	if err != nil {
		t.Fatalf("reuse second job startup snapshot: %v", err)
	}

	snapshot := startupstats.FromContext(startupCtx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterJobSnapshotBuilds); got != 1 {
		t.Fatalf("expected one job snapshot build, got %d", got)
	}
}
