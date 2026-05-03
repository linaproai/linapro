// This file tests distributed KV cache mutation and expiration behavior.

package mysqlmemory

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
)

// TestIncrConcurrentCallsAreAtomic verifies concurrent increments on one cache
// key do not lose successful updates while the MEMORY table is alive.
func TestIncrConcurrentCallsAreAtomic(t *testing.T) {
	ctx := context.Background()
	service := NewMySQLMemoryBackend()
	cacheKey := BuildCacheKey("unit-plugin", "counter", "atomic")
	cleanupKVCacheKey(t, ctx, OwnerTypePlugin, cacheKey)

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

// TestIncrRejectsExistingStringWithoutMutation verifies non-integer values are
// rejected and preserved.
func TestIncrRejectsExistingStringWithoutMutation(t *testing.T) {
	ctx := context.Background()
	service := NewMySQLMemoryBackend()
	cacheKey := BuildCacheKey("unit-plugin", "counter", "string-value")
	cleanupKVCacheKey(t, ctx, OwnerTypePlugin, cacheKey)

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
	service := NewMySQLMemoryBackend()
	targetKey := BuildCacheKey("unit-plugin", "ttl", "target")
	otherKey := BuildCacheKey("unit-plugin", "ttl", "other")
	cleanupKVCacheKey(t, ctx, OwnerTypePlugin, targetKey)
	cleanupKVCacheKey(t, ctx, OwnerTypePlugin, otherKey)

	targetIdentity, err := parseCacheKey(targetKey)
	if err != nil {
		t.Fatalf("parse target key failed: %v", err)
	}
	otherIdentity, err := parseCacheKey(otherKey)
	if err != nil {
		t.Fatalf("parse other key failed: %v", err)
	}
	expiredAt := gtime.Now().Add(-time.Minute)
	insertExpiredKVRow(t, ctx, OwnerTypePlugin, targetIdentity, expiredAt)
	insertExpiredKVRow(t, ctx, OwnerTypePlugin, otherIdentity, expiredAt)

	if _, ok, err := service.Get(ctx, OwnerTypePlugin, targetKey); err != nil {
		t.Fatalf("get expired target failed: %v", err)
	} else if ok {
		t.Fatal("expected expired target key to be treated as a miss")
	}

	targetCount := countKVCacheKey(t, ctx, OwnerTypePlugin, targetIdentity)
	otherCount := countKVCacheKey(t, ctx, OwnerTypePlugin, otherIdentity)
	if targetCount != 1 {
		t.Fatalf("expected touched expired row to remain for background cleanup, got %d", targetCount)
	}
	if otherCount != 1 {
		t.Fatalf("expected unrelated expired row to remain for background cleanup, got %d", otherCount)
	}
}

// TestExpireAndSetCanClearExpiration verifies zero-expiration operations remove
// an existing TTL instead of leaving stale expire_at metadata behind.
func TestExpireAndSetCanClearExpiration(t *testing.T) {
	ctx := context.Background()
	service := NewMySQLMemoryBackend()
	cacheKey := BuildCacheKey("unit-plugin", "ttl", "clear")
	cleanupKVCacheKey(t, ctx, OwnerTypePlugin, cacheKey)

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
	if expireAt := readKVExpireAt(t, ctx, OwnerTypePlugin, identity); expireAt != nil {
		t.Fatalf("expected database expire_at to be NULL after Expire(0), got %v", expireAt)
	}

	if _, err := service.Set(ctx, OwnerTypePlugin, cacheKey, "temporary-again", 60*time.Second); err != nil {
		t.Fatalf("reset expiring value failed: %v", err)
	}
	if _, err := service.Set(ctx, OwnerTypePlugin, cacheKey, "persistent", 0); err != nil {
		t.Fatalf("set persistent value failed: %v", err)
	}
	if expireAt := readKVExpireAt(t, ctx, OwnerTypePlugin, identity); expireAt != nil {
		t.Fatalf("expected database expire_at to be NULL after Set(..., 0), got %v", expireAt)
	}
}

// TestCleanupExpiredRemovesExpiredRowsAsMisses verifies the global cleanup path
// is idempotent and later reads treat removed MEMORY-table cache rows as misses.
func TestCleanupExpiredRemovesExpiredRowsAsMisses(t *testing.T) {
	ctx := context.Background()
	service := NewMySQLMemoryBackend()
	cacheKey := BuildCacheKey("unit-plugin", "ttl", "cleanup")
	cleanupKVCacheKey(t, ctx, OwnerTypePlugin, cacheKey)

	identity, err := parseCacheKey(cacheKey)
	if err != nil {
		t.Fatalf("parse cache key failed: %v", err)
	}
	insertExpiredKVRow(t, ctx, OwnerTypePlugin, identity, gtime.Now().Add(-time.Minute))

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
	service := NewMySQLMemoryBackend()
	validKey := BuildCacheKey("unit-plugin", "oversized", "value")
	cleanupKVCacheKey(t, ctx, OwnerTypePlugin, validKey)

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

// TestDeletedMemoryCacheRowBehavesAsMiss verifies callers recover from MEMORY
// table row loss as a normal cache miss.
func TestDeletedMemoryCacheRowBehavesAsMiss(t *testing.T) {
	ctx := context.Background()
	service := NewMySQLMemoryBackend()
	cacheKey := BuildCacheKey("unit-plugin", "restart", "lost-row")
	cleanupKVCacheKey(t, ctx, OwnerTypePlugin, cacheKey)

	if _, err := service.Set(ctx, OwnerTypePlugin, cacheKey, "warm", 0); err != nil {
		t.Fatalf("seed cache value failed: %v", err)
	}
	identity, err := parseCacheKey(cacheKey)
	if err != nil {
		t.Fatalf("parse cache key failed: %v", err)
	}
	if _, err = dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
		OwnerType: OwnerTypePlugin.String(),
		OwnerKey:  identity.ownerKey,
		Namespace: identity.namespace,
		CacheKey:  identity.cacheKey,
	}).Delete(); err != nil {
		t.Fatalf("delete simulated lost MEMORY row failed: %v", err)
	}

	if item, ok, err := service.Get(ctx, OwnerTypePlugin, cacheKey); err != nil {
		t.Fatalf("read lost MEMORY cache row failed: %v", err)
	} else if ok || item != nil {
		t.Fatalf("expected lost MEMORY row to behave as cache miss, got item=%#v ok=%t", item, ok)
	}
}

// insertExpiredKVRow inserts one expired string cache row for lazy cleanup tests.
func insertExpiredKVRow(
	t *testing.T,
	ctx context.Context,
	ownerType OwnerType,
	identity *cacheIdentity,
	expiredAt *gtime.Time,
) {
	t.Helper()

	_, err := dao.SysKvCache.Ctx(ctx).Data(do.SysKvCache{
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

// cleanupKVCacheKey removes one cache key before and after a test.
func cleanupKVCacheKey(t *testing.T, ctx context.Context, ownerType OwnerType, cacheKey string) {
	t.Helper()

	identity, err := parseCacheKey(cacheKey)
	if err != nil {
		t.Fatalf("parse cache key failed: %v", err)
	}
	cleanup := func() {
		if _, err = dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
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
	ownerType OwnerType,
	identity *cacheIdentity,
) int {
	t.Helper()

	count, err := dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
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
	ownerType OwnerType,
	identity *cacheIdentity,
) *gtime.Time {
	t.Helper()

	var row struct {
		ExpireAt *gtime.Time
	}
	err := dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
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
