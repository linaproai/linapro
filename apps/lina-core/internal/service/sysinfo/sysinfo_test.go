// This file verifies system-info diagnostic projections.

package sysinfo

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/service/cachecoord"
)

const (
	// testRuntimeConfigDomain is the sysinfo test projection domain.
	testRuntimeConfigDomain cachecoord.Domain = "runtime-config"
	// sqliteSysInfoChildEnv marks the isolated child process that owns
	// GoFrame's global SQLite database configuration.
	sqliteSysInfoChildEnv = "LINA_SQLITE_SYSINFO_CHILD"
	// sqliteSysInfoDBEnv stores the temporary SQLite path for the child test.
	sqliteSysInfoDBEnv = "LINA_SQLITE_SYSINFO_DB"
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

// MarkTenantChanged is unused by sysinfo diagnostics.
func (f *fakeCacheCoordService) MarkTenantChanged(
	_ context.Context,
	_ cachecoord.Domain,
	_ cachecoord.Scope,
	_ cachecoord.InvalidationScope,
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

// TestGetDbVersionSupportsSQLite verifies sysinfo does not use MySQL-only
// VERSION() diagnostics against SQLite databases.
func TestGetDbVersionSupportsSQLite(t *testing.T) {
	if os.Getenv(sqliteSysInfoChildEnv) == "1" {
		t.Skip("parent test only launches the isolated SQLite child process")
	}

	dbPath := filepath.Join(t.TempDir(), "sysinfo.db")
	cmd := exec.Command(os.Args[0], "-test.run=^TestGetDbVersionSupportsSQLiteChild$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(),
		sqliteSysInfoChildEnv+"=1",
		sqliteSysInfoDBEnv+"="+dbPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("SQLite sysinfo child test failed: %v\n%s", err, string(output))
	}
}

// TestGetDbVersionSupportsSQLiteChild runs the actual SQLite sysinfo
// diagnostic check in an isolated process because GoFrame database config is
// global.
func TestGetDbVersionSupportsSQLiteChild(t *testing.T) {
	if os.Getenv(sqliteSysInfoChildEnv) != "1" {
		t.Skip("SQLite sysinfo child test is executed by TestGetDbVersionSupportsSQLite")
	}

	ctx := context.Background()
	dbPath := os.Getenv(sqliteSysInfoDBEnv)
	if dbPath == "" {
		t.Fatalf("%s must be set", sqliteSysInfoDBEnv)
	}
	link := "sqlite::@file(" + dbPath + ")"
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: link}},
	}); err != nil {
		t.Fatalf("configure SQLite sysinfo database failed: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := g.DB().Close(ctx); closeErr != nil {
			t.Errorf("close SQLite sysinfo database failed: %v", closeErr)
		}
	})

	service := &serviceImpl{}
	version, err := service.getDbVersion(ctx)
	if err != nil {
		t.Fatalf("get SQLite sysinfo database version failed: %v", err)
	}
	if !strings.HasPrefix(version, "SQLite ") {
		t.Fatalf("expected SQLite version label, got %q", version)
	}
	if strings.TrimSpace(strings.TrimPrefix(version, "SQLite ")) == "" {
		t.Fatalf("expected SQLite version number to be non-empty, got %q", version)
	}
}
