// This file tests distributed KV cache mutation and expiration behavior.

package sqltable

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/model/do"
	"lina-core/pkg/dialect"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// currentSQLTableKVCacheDDL keeps package tests aligned with the delivered
// persistent sys_kv_cache SQL table definition.
const currentSQLTableKVCacheDDL = `
CREATE TABLE IF NOT EXISTS sys_kv_cache (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   INT NOT NULL DEFAULT 0,
    owner_type  VARCHAR(16) NOT NULL DEFAULT '',
    owner_key   VARCHAR(64) NOT NULL DEFAULT '',
    namespace   VARCHAR(64) NOT NULL DEFAULT '',
    cache_key   VARCHAR(128) NOT NULL DEFAULT '',
    value_kind  SMALLINT NOT NULL DEFAULT 1,
    value_bytes BYTEA NOT NULL,
    value_int   BIGINT NOT NULL DEFAULT 0,
    expire_at   TIMESTAMP NULL DEFAULT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_kv_cache_tenant_owner_namespace_key ON sys_kv_cache (tenant_id, owner_type, owner_key, namespace, cache_key);
CREATE INDEX IF NOT EXISTS idx_sys_kv_cache_expire_at ON sys_kv_cache (expire_at);
`

// TestIncrConcurrentCallsAreAtomic verifies concurrent increments on one cache
// key do not lose successful updates while the SQL table is alive.
func TestIncrConcurrentCallsAreAtomic(t *testing.T) {
	ctx := context.Background()
	service := newTestSQLTableBackend(t, ctx)
	cacheKey := BuildCacheKey("unit-plugin", "counter", "atomic")
	cleanupKVCacheKey(t, ctx, service, OwnerTypePlugin, cacheKey)

	const workers = 16
	values := make(chan int64, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			item, err := service.Incr(ctx, OwnerTypePlugin, cacheKey, 1, 0)
			if err != nil {
				errs <- err
				return
			}
			values <- item.IntValue
		}()
	}
	wg.Wait()
	close(values)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent increment failed: %v", err)
		}
	}

	seen := make(map[int64]struct{}, workers)
	for value := range values {
		seen[value] = struct{}{}
	}
	if len(seen) != workers {
		t.Fatalf("expected %d unique increment results, got %d: %#v", workers, len(seen), seen)
	}

	value, ok, err := service.GetInt(ctx, OwnerTypePlugin, cacheKey)
	if err != nil {
		t.Fatalf("read final increment value failed: %v", err)
	}
	if !ok || value != workers {
		t.Fatalf("expected final value %d, got value=%d ok=%t", workers, value, ok)
	}
}

// TestIncrSQLiteCallsAreAtomic verifies the increment algorithm against the
// SQLite dialect, including concurrent first writes to a missing cache key.
func TestIncrSQLiteCallsAreAtomic(t *testing.T) {
	ctx := context.Background()
	service := newSQLiteTestSQLTableBackend(t, ctx)
	cacheKey := BuildCacheKey("unit-plugin", "counter", "sqlite-atomic")

	const workers = 12
	values := make(chan int64, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			item, err := service.Incr(ctx, OwnerTypePlugin, cacheKey, 1, 0)
			if err != nil {
				errs <- err
				return
			}
			values <- item.IntValue
		}()
	}
	wg.Wait()
	close(values)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("SQLite concurrent increment failed: %v", err)
		}
	}

	seen := make(map[int64]struct{}, workers)
	for value := range values {
		seen[value] = struct{}{}
	}
	if len(seen) != workers {
		t.Fatalf("expected %d unique SQLite increment results, got %d: %#v", workers, len(seen), seen)
	}

	value, ok, err := service.GetInt(ctx, OwnerTypePlugin, cacheKey)
	if err != nil {
		t.Fatalf("read final SQLite increment value failed: %v", err)
	}
	if !ok || value != workers {
		t.Fatalf("expected final SQLite value %d, got value=%d ok=%t", workers, value, ok)
	}
}

// TestIncrMissingKeyStartsFromDelta verifies first increment preserves the
// public delta-as-initial-value contract.
func TestIncrMissingKeyStartsFromDelta(t *testing.T) {
	ctx := context.Background()
	service := newTestSQLTableBackend(t, ctx)
	cacheKey := BuildCacheKey("unit-plugin", "counter", "initial-delta")
	cleanupKVCacheKey(t, ctx, service, OwnerTypePlugin, cacheKey)

	item, err := service.Incr(ctx, OwnerTypePlugin, cacheKey, 5, 0)
	if err != nil {
		t.Fatalf("first increment failed: %v", err)
	}
	if item.IntValue != 5 {
		t.Fatalf("expected first increment value 5, got %d", item.IntValue)
	}

	item, err = service.Incr(ctx, OwnerTypePlugin, cacheKey, 2, 0)
	if err != nil {
		t.Fatalf("second increment failed: %v", err)
	}
	if item.IntValue != 7 {
		t.Fatalf("expected second increment value 7, got %d", item.IntValue)
	}
}

// TestIncrZeroDeltaIsStable verifies zero-delta increments do not depend on
// database affected-row behavior for no-op updates.
func TestIncrZeroDeltaIsStable(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		name       string
		newService func(*testing.T, context.Context) *SQLTableBackend
	}{
		{name: "postgres", newService: newTestSQLTableBackend},
		{name: "sqlite", newService: newSQLiteTestSQLTableBackend},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			service := testCase.newService(t, ctx)
			cacheKey := BuildCacheKey("unit-plugin", "counter", "zero-delta-"+testCase.name)
			cleanupKVCacheKey(t, ctx, service, OwnerTypePlugin, cacheKey)

			item, err := service.Incr(ctx, OwnerTypePlugin, cacheKey, 0, 0)
			if err != nil {
				t.Fatalf("zero-delta first increment failed: %v", err)
			}
			if item.IntValue != 0 {
				t.Fatalf("expected zero-delta first value 0, got %d", item.IntValue)
			}

			item, err = service.Incr(ctx, OwnerTypePlugin, cacheKey, 0, 0)
			if err != nil {
				t.Fatalf("zero-delta existing increment failed: %v", err)
			}
			if item.IntValue != 0 {
				t.Fatalf("expected zero-delta existing value 0, got %d", item.IntValue)
			}

			item, err = service.Incr(ctx, OwnerTypePlugin, cacheKey, 3, 0)
			if err != nil {
				t.Fatalf("post-zero increment failed: %v", err)
			}
			if item.IntValue != 3 {
				t.Fatalf("expected post-zero increment value 3, got %d", item.IntValue)
			}
		})
	}
}

// TestIncrRejectsExistingStringWithoutMutation verifies non-integer values are
// rejected and preserved.
func TestIncrRejectsExistingStringWithoutMutation(t *testing.T) {
	ctx := context.Background()
	service := newTestSQLTableBackend(t, ctx)
	cacheKey := BuildCacheKey("unit-plugin", "counter", "string-value")
	cleanupKVCacheKey(t, ctx, service, OwnerTypePlugin, cacheKey)

	if _, err := service.Set(ctx, OwnerTypePlugin, cacheKey, "not-an-int", 0); err != nil {
		t.Fatalf("seed string value failed: %v", err)
	}
	if _, err := service.Incr(ctx, OwnerTypePlugin, cacheKey, 1, 0); err == nil {
		t.Fatal("expected incrementing a string value to fail")
	}
	item, ok, err := service.Get(ctx, OwnerTypePlugin, cacheKey)
	if err != nil {
		t.Fatalf("read string value after failed increment: %v", err)
	}
	if !ok || item.Value != "not-an-int" || item.ValueKind != ValueKindString {
		t.Fatalf("expected original string value to remain, got item=%#v ok=%t", item, ok)
	}
}

// TestGetExpiredKeyIsReadOnlyMiss verifies request-path expiration handling
// returns a miss without deleting touched or unrelated expired rows.
func TestGetExpiredKeyIsReadOnlyMiss(t *testing.T) {
	ctx := context.Background()
	service := newSQLitePersistentTestSQLTableBackend(t, ctx)
	targetKey := BuildCacheKey("unit-plugin", "ttl", "target")
	otherKey := BuildCacheKey("unit-plugin", "ttl", "other")
	cleanupKVCacheKey(t, ctx, service, OwnerTypePlugin, targetKey)
	cleanupKVCacheKey(t, ctx, service, OwnerTypePlugin, otherKey)

	targetIdentity, err := parseCacheKey(targetKey)
	if err != nil {
		t.Fatalf("parse target key failed: %v", err)
	}
	otherIdentity, err := parseCacheKey(otherKey)
	if err != nil {
		t.Fatalf("parse other key failed: %v", err)
	}
	expiredAt := gtime.Now().Add(-time.Minute)
	insertExpiredKVRow(t, ctx, service, OwnerTypePlugin, targetIdentity, expiredAt)
	insertExpiredKVRow(t, ctx, service, OwnerTypePlugin, otherIdentity, expiredAt)

	if _, ok, err := service.Get(ctx, OwnerTypePlugin, targetKey); err != nil {
		t.Fatalf("get expired target failed: %v", err)
	} else if ok {
		t.Fatal("expected expired target key to be treated as a miss")
	}

	targetCount := countKVCacheKey(t, ctx, service, OwnerTypePlugin, targetIdentity)
	otherCount := countKVCacheKey(t, ctx, service, OwnerTypePlugin, otherIdentity)
	if targetCount != 1 {
		t.Fatalf("expected touched expired row to remain for background cleanup, got %d", targetCount)
	}
	if otherCount != 1 {
		t.Fatalf("expected unrelated expired row to remain for background cleanup, got %d", otherCount)
	}
}

// TestSetReplacesExpiredRowAsFreshValue verifies write paths inspect
// expiration before reusing a persistent row left behind by an earlier process.
func TestSetReplacesExpiredRowAsFreshValue(t *testing.T) {
	ctx := context.Background()
	service := newSQLitePersistentTestSQLTableBackend(t, ctx)
	cacheKey := BuildCacheKey("unit-plugin", "ttl", "set-expired")
	cleanupKVCacheKey(t, ctx, service, OwnerTypePlugin, cacheKey)

	identity, err := parseCacheKey(cacheKey)
	if err != nil {
		t.Fatalf("parse cache key failed: %v", err)
	}
	insertExpiredKVRow(t, ctx, service, OwnerTypePlugin, identity, gtime.Now().Add(-time.Minute))

	item, err := service.Set(ctx, OwnerTypePlugin, cacheKey, "fresh", 0)
	if err != nil {
		t.Fatalf("set fresh value over expired row failed: %v", err)
	}
	if item.Value != "fresh" || item.ExpireAt != nil {
		t.Fatalf("expected fresh persistent value, got %#v", item)
	}
	if count := countKVCacheKey(t, ctx, service, OwnerTypePlugin, identity); count != 1 {
		t.Fatalf("expected exactly one fresh cache row, got %d", count)
	}
}

// TestUnexpiredRowSurvivesBackendRecreation verifies valid SQL-table cache
// entries are not cleared when a new backend instance is created after restart.
func TestUnexpiredRowSurvivesBackendRecreation(t *testing.T) {
	ctx := context.Background()
	firstService := newTestSQLTableBackend(t, ctx)
	cacheKey := BuildCacheKey("unit-plugin", "restart", "survives")
	cleanupKVCacheKey(t, ctx, firstService, OwnerTypePlugin, cacheKey)

	if _, err := firstService.Set(ctx, OwnerTypePlugin, cacheKey, "warm", time.Hour); err != nil {
		t.Fatalf("seed cache value before backend recreation failed: %v", err)
	}

	secondService := NewSQLTableBackend()
	item, ok, err := secondService.Get(ctx, OwnerTypePlugin, cacheKey)
	if err != nil {
		t.Fatalf("read cache value after backend recreation failed: %v", err)
	}
	if !ok || item == nil || item.Value != "warm" {
		t.Fatalf("expected cache value to survive backend recreation, got item=%#v ok=%t", item, ok)
	}

	cleanupKVCacheKey(t, ctx, firstService, OwnerTypePlugin, cacheKey)
}

// TestIncrExpiredIntegerStartsFromDelta verifies increment paths do not add to
// stale integer values after the persistent row has expired.
func TestIncrExpiredIntegerStartsFromDelta(t *testing.T) {
	ctx := context.Background()
	service := newSQLitePersistentTestSQLTableBackend(t, ctx)
	cacheKey := BuildCacheKey("unit-plugin", "ttl", "incr-expired")
	cleanupKVCacheKey(t, ctx, service, OwnerTypePlugin, cacheKey)

	identity, err := parseCacheKey(cacheKey)
	if err != nil {
		t.Fatalf("parse cache key failed: %v", err)
	}
	insertExpiredKVIntRow(t, ctx, service, OwnerTypePlugin, identity, 40, gtime.Now().Add(-time.Minute))

	item, err := service.Incr(ctx, OwnerTypePlugin, cacheKey, 2, 0)
	if err != nil {
		t.Fatalf("increment expired integer row failed: %v", err)
	}
	if item.IntValue != 2 {
		t.Fatalf("expected expired integer row to restart at delta 2, got %d", item.IntValue)
	}
}

// TestExpireAndSetCanClearExpiration verifies zero-expiration operations remove
// an existing TTL instead of leaving stale expire_at metadata behind.
func TestExpireAndSetCanClearExpiration(t *testing.T) {
	ctx := context.Background()
	service := newTestSQLTableBackend(t, ctx)
	cacheKey := BuildCacheKey("unit-plugin", "ttl", "clear")
	cleanupKVCacheKey(t, ctx, service, OwnerTypePlugin, cacheKey)

	if _, err := service.Set(ctx, OwnerTypePlugin, cacheKey, "temporary", 60*time.Second); err != nil {
		t.Fatalf("seed expiring value failed: %v", err)
	}
	if found, expireAt, err := service.Expire(ctx, OwnerTypePlugin, cacheKey, 0); err != nil {
		t.Fatalf("clear expiration failed: %v", err)
	} else if !found || expireAt != nil {
		t.Fatalf("expected expiration to clear, got found=%t expireAt=%v", found, expireAt)
	}
	identity, err := parseCacheKey(cacheKey)
	if err != nil {
		t.Fatalf("parse cache key failed: %v", err)
	}
	if expireAt := readKVExpireAt(t, ctx, service, OwnerTypePlugin, identity); expireAt != nil {
		t.Fatalf("expected database expire_at to be NULL after Expire(0), got %v", expireAt)
	}

	if _, err := service.Set(ctx, OwnerTypePlugin, cacheKey, "temporary-again", 60*time.Second); err != nil {
		t.Fatalf("reset expiring value failed: %v", err)
	}
	if _, err := service.Set(ctx, OwnerTypePlugin, cacheKey, "persistent", 0); err != nil {
		t.Fatalf("set persistent value failed: %v", err)
	}
	if expireAt := readKVExpireAt(t, ctx, service, OwnerTypePlugin, identity); expireAt != nil {
		t.Fatalf("expected database expire_at to be NULL after Set(..., 0), got %v", expireAt)
	}
}

// TestCleanupExpiredRemovesExpiredRowsAsMisses verifies the global cleanup path
// is idempotent and later reads treat removed SQL-table cache rows as misses.
func TestCleanupExpiredRemovesExpiredRowsAsMisses(t *testing.T) {
	ctx := context.Background()
	service := newTestSQLTableBackend(t, ctx)
	cacheKey := BuildCacheKey("unit-plugin", "ttl", "cleanup")
	cleanupKVCacheKey(t, ctx, service, OwnerTypePlugin, cacheKey)

	identity, err := parseCacheKey(cacheKey)
	if err != nil {
		t.Fatalf("parse cache key failed: %v", err)
	}
	insertExpiredKVRow(t, ctx, service, OwnerTypePlugin, identity, gtime.Now().Add(-time.Minute))

	if err = service.CleanupExpired(ctx); err != nil {
		t.Fatalf("cleanup expired rows failed: %v", err)
	}
	if err = service.CleanupExpired(ctx); err != nil {
		t.Fatalf("repeat cleanup expired rows failed: %v", err)
	}
	if item, ok, err := service.Get(ctx, OwnerTypePlugin, cacheKey); err != nil {
		t.Fatalf("read cleaned cache key failed: %v", err)
	} else if ok || item != nil {
		t.Fatalf("expected cleaned cache key to behave as cache miss, got item=%#v ok=%t", item, ok)
	}
}

// TestSetRejectsOversizedInputsWithoutWriting verifies bounded cache identity
// and payload fields fail before any truncated value can be persisted.
func TestSetRejectsOversizedInputsWithoutWriting(t *testing.T) {
	ctx := context.Background()
	service := newTestSQLTableBackend(t, ctx)
	validKey := BuildCacheKey("unit-plugin", "oversized", "value")
	cleanupKVCacheKey(t, ctx, service, OwnerTypePlugin, validKey)

	testCases := []struct {
		name     string
		cacheKey string
		value    string
	}{
		{
			name:     "namespace too long",
			cacheKey: BuildCacheKey("unit-plugin", strings.Repeat("n", maxNamespaceBytes+1), "logical"),
			value:    "value",
		},
		{
			name:     "cache key too long",
			cacheKey: BuildCacheKey("unit-plugin", "oversized", strings.Repeat("k", maxCacheKeyBytes+1)),
			value:    "value",
		},
		{
			name:     "value too long",
			cacheKey: validKey,
			value:    strings.Repeat("v", maxValueBytes+1),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := service.Set(ctx, OwnerTypePlugin, testCase.cacheKey, testCase.value, 0); err == nil {
				t.Fatal("expected oversized cache input to fail")
			}
		})
	}

	if item, ok, err := service.Get(ctx, OwnerTypePlugin, validKey); err != nil {
		t.Fatalf("read valid key after rejected oversized value failed: %v", err)
	} else if ok || item != nil {
		t.Fatalf("expected rejected oversized value to leave cache missing, got item=%#v ok=%t", item, ok)
	}
}

// TestDeletedSQLTableCacheRowBehavesAsMiss verifies callers recover from SQL
// table row loss as a normal cache miss.
func TestDeletedSQLTableCacheRowBehavesAsMiss(t *testing.T) {
	ctx := context.Background()
	service := newTestSQLTableBackend(t, ctx)
	cacheKey := BuildCacheKey("unit-plugin", "restart", "lost-row")
	cleanupKVCacheKey(t, ctx, service, OwnerTypePlugin, cacheKey)

	if _, err := service.Set(ctx, OwnerTypePlugin, cacheKey, "warm", 0); err != nil {
		t.Fatalf("seed cache value failed: %v", err)
	}
	identity, err := parseCacheKey(cacheKey)
	if err != nil {
		t.Fatalf("parse cache key failed: %v", err)
	}
	if _, err = service.model(ctx).Where(do.SysKvCache{
		TenantId:  identity.tenantID,
		OwnerType: OwnerTypePlugin.String(),
		OwnerKey:  identity.ownerKey,
		Namespace: identity.namespace,
		CacheKey:  identity.cacheKey,
	}).Delete(); err != nil {
		t.Fatalf("delete simulated lost SQL table row failed: %v", err)
	}

	if item, ok, err := service.Get(ctx, OwnerTypePlugin, cacheKey); err != nil {
		t.Fatalf("read lost SQL table cache row failed: %v", err)
	} else if ok || item != nil {
		t.Fatalf("expected lost SQL table row to behave as cache miss, got item=%#v ok=%t", item, ok)
	}
}

// TestTenantCacheKeysAreIsolated verifies equal logical cache keys stay isolated
// when the owner key carries a tenant discriminator.
func TestTenantCacheKeysAreIsolated(t *testing.T) {
	ctx := context.Background()
	service := newTestSQLTableBackend(t, ctx)
	tenantOneKey := BuildCacheKey(
		pkgtenantcap.CacheKey(1, "dict", "sys"),
		"runtime",
		"user_status",
	)
	tenantTwoKey := BuildCacheKey(
		pkgtenantcap.CacheKey(2, "dict", "sys"),
		"runtime",
		"user_status",
	)
	cleanupKVCacheKey(t, ctx, service, OwnerTypeModule, tenantOneKey)
	cleanupKVCacheKey(t, ctx, service, OwnerTypeModule, tenantTwoKey)

	if _, err := service.Set(ctx, OwnerTypeModule, tenantOneKey, "tenant-one", 0); err != nil {
		t.Fatalf("set tenant one cache value failed: %v", err)
	}
	if _, err := service.Set(ctx, OwnerTypeModule, tenantTwoKey, "tenant-two", 0); err != nil {
		t.Fatalf("set tenant two cache value failed: %v", err)
	}

	itemOne, ok, err := service.Get(ctx, OwnerTypeModule, tenantOneKey)
	if err != nil {
		t.Fatalf("read tenant one cache value failed: %v", err)
	}
	if !ok || itemOne.Value != "tenant-one" {
		t.Fatalf("expected tenant one value to be isolated, got item=%#v ok=%t", itemOne, ok)
	}
	itemTwo, ok, err := service.Get(ctx, OwnerTypeModule, tenantTwoKey)
	if err != nil {
		t.Fatalf("read tenant two cache value failed: %v", err)
	}
	if !ok || itemTwo.Value != "tenant-two" {
		t.Fatalf("expected tenant two value to be isolated, got item=%#v ok=%t", itemTwo, ok)
	}
}

// newTestSQLTableBackend creates one backend on the process default database,
// preserving the existing PostgreSQL-backed package test coverage.
func newTestSQLTableBackend(t *testing.T, ctx context.Context) *SQLTableBackend {
	t.Helper()

	ensureCurrentSQLTableKVCacheTable(t, ctx)
	return NewSQLTableBackend()
}

// newSQLiteTestSQLTableBackend creates one isolated SQLite backend after
// translating the delivered sys_kv_cache table DDL through the public dialect.
func newSQLiteTestSQLTableBackend(t *testing.T, ctx context.Context) *SQLTableBackend {
	t.Helper()

	db := newSQLiteKVCacheDB(t, ctx)
	dbDialect, err := dialect.From("sqlite::@file(./temp/sqlite/kv-cache.db)")
	if err != nil {
		t.Fatalf("resolve SQLite dialect failed: %v", err)
	}
	translated, err := dbDialect.TranslateDDL(ctx, "sys_kv_cache.sql", currentSQLTableKVCacheDDL)
	if err != nil {
		t.Fatalf("translate sys_kv_cache SQLite DDL failed: %v", err)
	}
	for _, statement := range dialect.SplitSQLStatements(translated) {
		if _, err = db.Exec(ctx, statement); err != nil {
			t.Fatalf("execute sys_kv_cache SQLite DDL failed: %v\nSQL:\n%s", err, statement)
		}
	}
	return NewSQLTableBackendWithDB(db)
}

// newSQLitePersistentTestSQLTableBackend creates an isolated SQLite backend
// with direct test DDL for expiration behavior that must not depend on the
// process default database.
func newSQLitePersistentTestSQLTableBackend(t *testing.T, ctx context.Context) *SQLTableBackend {
	t.Helper()

	db := newSQLiteKVCacheDB(t, ctx)
	statements := []string{
		`CREATE TABLE IF NOT EXISTS sys_kv_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			owner_type TEXT NOT NULL DEFAULT '',
			owner_key TEXT NOT NULL DEFAULT '',
			namespace TEXT NOT NULL DEFAULT '',
			cache_key TEXT NOT NULL DEFAULT '',
			value_kind INTEGER NOT NULL DEFAULT 1,
			value_bytes BLOB NOT NULL,
			value_int INTEGER NOT NULL DEFAULT 0,
			expire_at DATETIME NULL DEFAULT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (tenant_id, owner_type, owner_key, namespace, cache_key)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sys_kv_cache_expire_at ON sys_kv_cache (expire_at)`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(ctx, statement); err != nil {
			t.Fatalf("execute SQLite persistent kv-cache test DDL failed: %v\nSQL:\n%s", err, statement)
		}
	}
	return NewSQLTableBackendWithDB(db)
}

// newSQLiteKVCacheDB opens one temporary SQLite database and closes it with the test.
func newSQLiteKVCacheDB(t *testing.T, ctx context.Context) gdb.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "kv-cache.db")
	db, err := gdb.New(gdb.ConfigNode{Link: "sqlite::@file(" + dbPath + ")"})
	if err != nil {
		t.Fatalf("create SQLite kv-cache test DB failed: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(ctx); closeErr != nil {
			t.Fatalf("close SQLite kv-cache test DB failed: %v", closeErr)
		}
	})
	return db
}

// ensureCurrentSQLTableKVCacheTable creates sys_kv_cache for package tests.
func ensureCurrentSQLTableKVCacheTable(t *testing.T, ctx context.Context) {
	t.Helper()

	for _, statement := range dialect.SplitSQLStatements(currentSQLTableKVCacheDDL) {
		if _, err := g.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("ensure sys_kv_cache table failed: %v\nSQL:\n%s", err, statement)
		}
	}
}

// insertExpiredKVRow inserts one expired string cache row for lazy cleanup tests.
func insertExpiredKVRow(
	t *testing.T,
	ctx context.Context,
	service *SQLTableBackend,
	ownerType OwnerType,
	identity *cacheIdentity,
	expiredAt *gtime.Time,
) {
	t.Helper()

	_, err := service.model(ctx).Data(do.SysKvCache{
		TenantId:   identity.tenantID,
		OwnerType:  ownerType.String(),
		OwnerKey:   identity.ownerKey,
		Namespace:  identity.namespace,
		CacheKey:   identity.cacheKey,
		ValueKind:  ValueKindString,
		ValueBytes: []byte("expired"),
		ValueInt:   0,
		ExpireAt:   expiredAt,
	}).InsertIgnore()
	if err != nil {
		t.Fatalf("insert expired kv row failed: %v", err)
	}
}

// insertExpiredKVIntRow inserts one expired integer cache row for stale-value
// increment tests.
func insertExpiredKVIntRow(
	t *testing.T,
	ctx context.Context,
	service *SQLTableBackend,
	ownerType OwnerType,
	identity *cacheIdentity,
	value int64,
	expiredAt *gtime.Time,
) {
	t.Helper()

	_, err := service.model(ctx).Data(do.SysKvCache{
		TenantId:   identity.tenantID,
		OwnerType:  ownerType.String(),
		OwnerKey:   identity.ownerKey,
		Namespace:  identity.namespace,
		CacheKey:   identity.cacheKey,
		ValueKind:  ValueKindInt,
		ValueBytes: []byte{},
		ValueInt:   value,
		ExpireAt:   expiredAt,
	}).InsertIgnore()
	if err != nil {
		t.Fatalf("insert expired integer kv row failed: %v", err)
	}
}

// cleanupKVCacheKey removes one cache key before and after a test.
func cleanupKVCacheKey(
	t *testing.T,
	ctx context.Context,
	service *SQLTableBackend,
	ownerType OwnerType,
	cacheKey string,
) {
	t.Helper()

	identity, err := parseCacheKey(cacheKey)
	if err != nil {
		t.Fatalf("parse cache key failed: %v", err)
	}
	cleanup := func() {
		if _, err = service.model(ctx).Where(do.SysKvCache{
			TenantId:  identity.tenantID,
			OwnerType: ownerType.String(),
			OwnerKey:  identity.ownerKey,
			Namespace: identity.namespace,
			CacheKey:  identity.cacheKey,
		}).Delete(); err != nil {
			t.Fatalf("cleanup kv cache key failed: %v", err)
		}
	}
	cleanup()
	t.Cleanup(cleanup)
}

// countKVCacheKey returns the number of rows matching one cache identity.
func countKVCacheKey(
	t *testing.T,
	ctx context.Context,
	service *SQLTableBackend,
	ownerType OwnerType,
	identity *cacheIdentity,
) int {
	t.Helper()

	count, err := service.model(ctx).Where(do.SysKvCache{
		TenantId:  identity.tenantID,
		OwnerType: ownerType.String(),
		OwnerKey:  identity.ownerKey,
		Namespace: identity.namespace,
		CacheKey:  identity.cacheKey,
	}).Count()
	if err != nil {
		t.Fatalf("count kv cache key failed: %v", err)
	}
	return count
}

// readKVExpireAt returns the stored expiration timestamp for one cache identity.
func readKVExpireAt(
	t *testing.T,
	ctx context.Context,
	service *SQLTableBackend,
	ownerType OwnerType,
	identity *cacheIdentity,
) *gtime.Time {
	t.Helper()

	var row struct {
		ExpireAt *gtime.Time
	}
	err := service.model(ctx).Where(do.SysKvCache{
		TenantId:  identity.tenantID,
		OwnerType: ownerType.String(),
		OwnerKey:  identity.ownerKey,
		Namespace: identity.namespace,
		CacheKey:  identity.cacheKey,
	}).Scan(&row)
	if err != nil {
		t.Fatalf("read kv expire_at failed: %v", err)
	}
	return row.ExpireAt
}
