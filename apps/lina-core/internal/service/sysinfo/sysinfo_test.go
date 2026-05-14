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
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
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

// fakeCoordinationService provides deterministic coordination health snapshots
// for sysinfo diagnostics.
type fakeCoordinationService struct {
	snapshot coordination.HealthSnapshot
}

// BackendName returns the configured backend name.
func (f *fakeCoordinationService) BackendName() coordination.BackendName {
	return f.snapshot.Backend
}

// KeyBuilder returns a default key builder for interface completeness.
func (f *fakeCoordinationService) KeyBuilder() *coordination.KeyBuilder {
	return coordination.DefaultKeyBuilder()
}

// Lock is unused by sysinfo diagnostics.
func (f *fakeCoordinationService) Lock() coordination.LockStore {
	return nil
}

// KV is unused by sysinfo diagnostics.
func (f *fakeCoordinationService) KV() coordination.KVStore {
	return nil
}

// Revision is unused by sysinfo diagnostics.
func (f *fakeCoordinationService) Revision() coordination.RevisionStore {
	return nil
}

// Events is unused by sysinfo diagnostics.
func (f *fakeCoordinationService) Events() coordination.EventBus {
	return nil
}

// Health returns the deterministic health checker.
func (f *fakeCoordinationService) Health() coordination.HealthChecker {
	return fakeCoordinationHealth{snapshot: f.snapshot}
}

// Close is a no-op for the fake service.
func (f *fakeCoordinationService) Close(context.Context) error {
	return nil
}

// fakeCoordinationHealth returns deterministic health snapshots.
type fakeCoordinationHealth struct {
	snapshot coordination.HealthSnapshot
}

// Ping returns nil only when the snapshot is healthy.
func (f fakeCoordinationHealth) Ping(context.Context) error {
	if f.snapshot.Healthy {
		return nil
	}
	return errors.New("coordination backend error")
}

// Snapshot returns the deterministic health snapshot.
func (f fakeCoordinationHealth) Snapshot(context.Context) coordination.HealthSnapshot {
	return f.snapshot
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

// TestNewRejectsNilConfigService verifies sysinfo construction fails fast when
// runtime-owned config dependencies are missing.
func TestNewRejectsNilConfigService(t *testing.T) {
	assertPanic(t, "sysinfo service requires a non-nil config service", func() {
		New(nil, nil, nil, nil)
	})
}

// TestNewRejectsNilCacheCoordinationService verifies sysinfo construction
// fails fast when cache coordination diagnostics would use an isolated fallback.
func TestNewRejectsNilCacheCoordinationService(t *testing.T) {
	assertPanic(t, "sysinfo service requires a non-nil cache coordination service", func() {
		New(config.New(), nil, nil, nil)
	})
}

// TestLoadCacheCoordinationMapsSnapshot verifies cachecoord diagnostics are
// exposed by sysinfo without changing their semantic fields.
func TestLoadCacheCoordinationMapsSnapshot(t *testing.T) {
	syncedAt := time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)
	eventAt := syncedAt.Add(time.Second)
	service := &serviceImpl{
		cacheCoordSvc: &fakeCacheCoordService{
			items: []cachecoord.SnapshotItem{
				{
					Domain:                 testRuntimeConfigDomain,
					Scope:                  cachecoord.ScopeGlobal,
					AuthoritySource:        "sys_config protected runtime parameters",
					ConsistencyModel:       cachecoord.ConsistencySharedRevision,
					MaxStale:               10 * time.Second,
					FailureStrategy:        cachecoord.FailureStrategyReturnVisibleError,
					Backend:                coordination.BackendRedis,
					CoordinationHealthy:    true,
					LocalRevision:          3,
					SharedRevision:         4,
					LastSyncedAt:           syncedAt,
					EventSubscriberRunning: true,
					LastEventReceivedAt:    eventAt,
					RecentError:            "previous read failed",
					StaleSeconds:           2,
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
		item.Backend != coordination.BackendRedis ||
		!item.Healthy ||
		item.MaxStale != 10*time.Second ||
		item.LocalRevision != 3 ||
		item.SharedRevision != 4 ||
		!item.LastSyncedAt.Equal(syncedAt) ||
		!item.EventSubscriber ||
		!item.LastEventAt.Equal(eventAt) ||
		item.RecentError != "previous read failed" ||
		item.StaleSeconds != 2 {
		t.Fatalf("unexpected cache coordination diagnostic row: %#v", item)
	}
}

// TestLoadCoordinationUsesRuntimeServices verifies sysinfo reports the active
// runtime cluster topology and coordination backend health.
func TestLoadCoordinationUsesRuntimeServices(t *testing.T) {
	coordSvc := coordination.NewMemory(nil)
	service := &serviceImpl{
		configSvc:       config.New(),
		clusterSvc:      cluster.NewWithCoordination(&config.ClusterConfig{Enabled: true}, coordSvc),
		coordinationSvc: coordSvc,
	}

	info := service.loadCoordination(context.Background())
	if !info.ClusterEnabled ||
		info.Backend != coordination.BackendMemory ||
		info.NodeID == "" ||
		info.Primary ||
		info.LastSuccessAt.IsZero() ||
		info.LastError != "" {
		t.Fatalf("unexpected coordination diagnostics: %#v", info)
	}
}

// TestLoadCoordinationReportsRedisHealth verifies Redis backend health is
// exposed without leaking raw connection details.
func TestLoadCoordinationReportsRedisHealth(t *testing.T) {
	ctx := context.Background()
	successAt := time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)

	testCases := []struct {
		name          string
		healthy       bool
		lastError     string
		expectedError string
	}{
		{name: "healthy", healthy: true},
		{
			name:          "unhealthy sanitized",
			healthy:       false,
			lastError:     "redis://:secret@127.0.0.1:6379 token linapro:default:default:auth:revoke:x",
			expectedError: "coordination backend error",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			service := &serviceImpl{
				configSvc:  config.New(),
				clusterSvc: cluster.NewWithCoordination(&config.ClusterConfig{Enabled: true}, nil),
				coordinationSvc: &fakeCoordinationService{
					snapshot: coordination.HealthSnapshot{
						Backend:       coordination.BackendRedis,
						Healthy:       testCase.healthy,
						LastSuccessAt: successAt,
						LastError:     testCase.lastError,
					},
				},
			}

			info := service.loadCoordination(ctx)
			if info.Backend != coordination.BackendRedis ||
				info.RedisHealthy != testCase.healthy ||
				!info.LastSuccessAt.Equal(successAt) ||
				info.LastError != testCase.expectedError {
				t.Fatalf("unexpected redis coordination diagnostics: %#v", info)
			}
		})
	}
}

// TestLoadCoordinationSanitizesErrors verifies diagnostics avoid leaking
// connection strings, credentials, token keys, or full Redis keys.
func TestLoadCoordinationSanitizesErrors(t *testing.T) {
	if got := sanitizeCoordinationError("redis://:secret@127.0.0.1:6379 token linapro:default:default:auth:revoke:x"); got != "coordination backend error" {
		t.Fatalf("expected sensitive error to be sanitized, got %q", got)
	}
	if got := sanitizeCoordinationError("dial tcp 127.0.0.1:6379: connect: connection refused"); got != "redis coordination connection failed" {
		t.Fatalf("expected connection error category, got %q", got)
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
