// This file verifies system-info diagnostic projections.

package sysinfo

import (
	"context"
	"errors"
	"testing"
	"time"

	"lina-core/internal/service/cachecoord"
)

const (
	// testRuntimeConfigDomain is the sysinfo test projection domain.
	testRuntimeConfigDomain cachecoord.Domain = "runtime-config"
)

// fakeCacheCoordService provides deterministic cachecoord snapshots for
// sysinfo diagnostics.
type fakeCacheCoordService struct {
	items []cachecoord.SnapshotItem
	err   error
}

// ConfigureDomain is unused by sysinfo diagnostics.
func (f *fakeCacheCoordService) ConfigureDomain(_ cachecoord.DomainSpec) error {
	return nil
}

// MarkChanged is unused by sysinfo diagnostics.
func (f *fakeCacheCoordService) MarkChanged(
	_ context.Context,
	_ cachecoord.Domain,
	_ cachecoord.Scope,
	_ cachecoord.ChangeReason,
) (int64, error) {
	return 0, nil
}

// EnsureFresh is unused by sysinfo diagnostics.
func (f *fakeCacheCoordService) EnsureFresh(
	_ context.Context,
	_ cachecoord.Domain,
	_ cachecoord.Scope,
	_ cachecoord.Refresher,
) (int64, error) {
	return 0, nil
}

// CurrentRevision is unused by sysinfo diagnostics.
func (f *fakeCacheCoordService) CurrentRevision(
	_ context.Context,
	_ cachecoord.Domain,
	_ cachecoord.Scope,
) (int64, error) {
	return 0, nil
}

// Snapshot returns the configured diagnostic rows.
func (f *fakeCacheCoordService) Snapshot(_ context.Context) ([]cachecoord.SnapshotItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

// TestLoadCacheCoordinationMapsSnapshot verifies cachecoord diagnostics are
// exposed by sysinfo without changing their semantic fields.
func TestLoadCacheCoordinationMapsSnapshot(t *testing.T) {
	syncedAt := time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)
	service := &serviceImpl{
		cacheCoordSvc: &fakeCacheCoordService{
			items: []cachecoord.SnapshotItem{
				{
					Domain:           testRuntimeConfigDomain,
					Scope:            cachecoord.ScopeGlobal,
					AuthoritySource:  "sys_config protected runtime parameters",
					ConsistencyModel: cachecoord.ConsistencySharedRevision,
					MaxStale:         10 * time.Second,
					FailureStrategy:  cachecoord.FailureStrategyReturnVisibleError,
					LocalRevision:    3,
					SharedRevision:   4,
					LastSyncedAt:     syncedAt,
					RecentError:      "previous read failed",
					StaleSeconds:     2,
				},
			},
		},
	}

	items := service.loadCacheCoordination(context.Background())
	if len(items) != 1 {
		t.Fatalf("expected one cache coordination diagnostic row, got %d", len(items))
	}
	item := items[0]
	if item.Domain != string(testRuntimeConfigDomain) ||
		item.Scope != string(cachecoord.ScopeGlobal) ||
		item.ConsistencyModel != string(cachecoord.ConsistencySharedRevision) ||
		item.FailureStrategy != string(cachecoord.FailureStrategyReturnVisibleError) ||
		item.MaxStale != 10*time.Second ||
		item.LocalRevision != 3 ||
		item.SharedRevision != 4 ||
		!item.LastSyncedAt.Equal(syncedAt) ||
		item.RecentError != "previous read failed" ||
		item.StaleSeconds != 2 {
		t.Fatalf("unexpected cache coordination diagnostic row: %#v", item)
	}
}

// TestLoadCacheCoordinationToleratesSnapshotFailure verifies system-info output
// remains available when cachecoord diagnostics cannot be loaded.
func TestLoadCacheCoordinationToleratesSnapshotFailure(t *testing.T) {
	service := &serviceImpl{
		cacheCoordSvc: &fakeCacheCoordService{err: errors.New("snapshot unavailable")},
	}

	if items := service.loadCacheCoordination(context.Background()); len(items) != 0 {
		t.Fatalf("expected empty diagnostics after snapshot failure, got %#v", items)
	}
}
