// This file verifies the single-node memory backend for kvcache.

package memory_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"lina-core/internal/service/kvcache"
	"lina-core/pkg/bizerr"
)

// TestMemoryBackendStoresStringWithNativeTTL verifies single-node memory string
// values honor TTL and do not need external cleanup.
func TestMemoryBackendStoresStringWithNativeTTL(t *testing.T) {
	ctx := context.Background()
	service := kvcache.New(kvcache.WithProvider(kvcache.NewMemoryProvider()))
	cacheKey := kvcache.BuildCacheKey("unit-owner", "memory", "string")

	if service.BackendName() != kvcache.BackendMemory {
		t.Fatalf("expected memory backend, got %q", service.BackendName())
	}
	if service.RequiresExpiredCleanup() {
		t.Fatal("expected memory backend to skip expired cleanup")
	}

	item, err := service.Set(ctx, kvcache.OwnerTypePlugin, cacheKey, "value", 20*time.Millisecond)
	if err != nil {
		t.Fatalf("set memory value: %v", err)
	}
	if item.Value != "value" || item.ValueKind != kvcache.ValueKindString || item.ExpireAt == nil {
		t.Fatalf("unexpected set item: %#v", item)
	}
	read, ok, err := service.Get(ctx, kvcache.OwnerTypePlugin, cacheKey)
	if err != nil || !ok || read.Value != "value" {
		t.Fatalf("expected memory value, item=%#v ok=%t err=%v", read, ok, err)
	}
	time.Sleep(40 * time.Millisecond)
	if _, ok, err = service.Get(ctx, kvcache.OwnerTypePlugin, cacheKey); err != nil || ok {
		t.Fatalf("expected expired memory value miss, ok=%t err=%v", ok, err)
	}
}

// TestMemoryBackendRejectsMissingTTL verifies cache writes require an explicit
// positive expiration instead of accepting missing TTL values.
func TestMemoryBackendRejectsMissingTTL(t *testing.T) {
	ctx := context.Background()
	service := kvcache.New(kvcache.WithProvider(kvcache.NewMemoryProvider()))
	cacheKey := kvcache.BuildCacheKey("unit-owner", "memory", "persistent")

	if _, err := service.Set(ctx, kvcache.OwnerTypePlugin, cacheKey, "value", 0); !bizerr.Is(err, kvcache.CodeKVCacheExpireSecondsRequired) {
		t.Fatalf("expected required TTL error for set, got %v", err)
	}
	if _, err := service.Incr(ctx, kvcache.OwnerTypePlugin, cacheKey, 1, 0); !bizerr.Is(err, kvcache.CodeKVCacheExpireSecondsRequired) {
		t.Fatalf("expected required TTL error for incr, got %v", err)
	}
	if _, _, err := service.Expire(ctx, kvcache.OwnerTypePlugin, cacheKey, 0); !bizerr.Is(err, kvcache.CodeKVCacheExpireSecondsRequired) {
		t.Fatalf("expected required TTL error for expire, got %v", err)
	}
}

// TestMemoryBackendIncrIsAtomicWithinProcess verifies increments are linear
// within one host process.
func TestMemoryBackendIncrIsAtomicWithinProcess(t *testing.T) {
	ctx := context.Background()
	service := kvcache.New(kvcache.WithProvider(kvcache.NewMemoryProvider()))
	cacheKey := kvcache.BuildCacheKey("unit-owner", "memory", "counter")

	const workers = 32
	values := make(chan int64, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			item, err := service.Incr(ctx, kvcache.OwnerTypePlugin, cacheKey, 1, time.Second)
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
			t.Fatalf("increment failed: %v", err)
		}
	}
	seen := make(map[int64]struct{}, workers)
	for value := range values {
		seen[value] = struct{}{}
	}
	if len(seen) != workers {
		t.Fatalf("expected %d unique increments, got %d: %#v", workers, len(seen), seen)
	}
	finalValue, ok, err := service.GetInt(ctx, kvcache.OwnerTypePlugin, cacheKey)
	if err != nil || !ok || finalValue != workers {
		t.Fatalf("expected final value %d, value=%d ok=%t err=%v", workers, finalValue, ok, err)
	}
}

// TestMemoryBackendIncrFirstDeltaAndRefreshesTTL verifies missing counters use
// delta as the first value and each increment applies an explicit new TTL.
func TestMemoryBackendIncrFirstDeltaAndRefreshesTTL(t *testing.T) {
	ctx := context.Background()
	service := kvcache.New(kvcache.WithProvider(kvcache.NewMemoryProvider()))
	cacheKey := kvcache.BuildCacheKey("unit-owner", "memory", "ttl-counter")

	item, err := service.Incr(ctx, kvcache.OwnerTypePlugin, cacheKey, 5, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("first incr: %v", err)
	}
	if item.IntValue != 5 || item.ExpireAt == nil {
		t.Fatalf("unexpected first incr item: %#v", item)
	}
	item, err = service.Incr(ctx, kvcache.OwnerTypePlugin, cacheKey, 2, 40*time.Millisecond)
	if err != nil {
		t.Fatalf("second incr: %v", err)
	}
	if item.IntValue != 7 || item.ExpireAt == nil {
		t.Fatalf("unexpected second incr item: %#v", item)
	}
	time.Sleep(80 * time.Millisecond)
	if _, ok, err := service.GetInt(ctx, kvcache.OwnerTypePlugin, cacheKey); err != nil || ok {
		t.Fatalf("expected refreshed TTL to expire counter, ok=%t err=%v", ok, err)
	}
}

// TestMemoryBackendRejectsInvalidOperations verifies validation mirrors the
// backend-neutral kvcache contract.
func TestMemoryBackendRejectsInvalidOperations(t *testing.T) {
	ctx := context.Background()
	service := kvcache.New(kvcache.WithProvider(kvcache.NewMemoryProvider()))
	cacheKey := kvcache.BuildCacheKey("unit-owner", "memory", "string")

	if _, err := service.Set(ctx, kvcache.OwnerTypePlugin, cacheKey, strings.Repeat("x", 4097), time.Second); !bizerr.Is(err, kvcache.CodeKVCacheValueTooLong) {
		t.Fatalf("expected value-too-long error, got %v", err)
	}
	if _, err := service.Set(ctx, kvcache.OwnerTypePlugin, cacheKey, "value", -time.Second); !bizerr.Is(err, kvcache.CodeKVCacheExpireSecondsNegative) {
		t.Fatalf("expected negative TTL error, got %v", err)
	}
	if _, err := service.Set(ctx, kvcache.OwnerTypePlugin, cacheKey, "value", time.Second); err != nil {
		t.Fatalf("set string for incr test: %v", err)
	}
	if _, err := service.Incr(ctx, kvcache.OwnerTypePlugin, cacheKey, 1, time.Second); !bizerr.Is(err, kvcache.CodeKVCacheIncrementValueNotInteger) {
		t.Fatalf("expected increment type error, got %v", err)
	}
	if _, _, err := service.Expire(ctx, kvcache.OwnerTypePlugin, cacheKey, -time.Second); !bizerr.Is(err, kvcache.CodeKVCacheExpireSecondsNegative) {
		t.Fatalf("expected expire negative TTL error, got %v", err)
	}
}

// TestMemoryBackendExpireUpdatesPolicy verifies Expire changes cache lifetime
// without changing the stored value.
func TestMemoryBackendExpireUpdatesPolicy(t *testing.T) {
	ctx := context.Background()
	service := kvcache.New(kvcache.WithProvider(kvcache.NewMemoryProvider()))
	cacheKey := kvcache.BuildCacheKey("unit-owner", "memory", "expire")

	if _, err := service.Set(ctx, kvcache.OwnerTypePlugin, cacheKey, "value", time.Second); err != nil {
		t.Fatalf("set value: %v", err)
	}
	ok, expireAt, err := service.Expire(ctx, kvcache.OwnerTypePlugin, cacheKey, 20*time.Millisecond)
	if err != nil || !ok || expireAt == nil {
		t.Fatalf("expected expire update, ok=%t expireAt=%v err=%v", ok, expireAt, err)
	}
	read, hit, err := service.Get(ctx, kvcache.OwnerTypePlugin, cacheKey)
	if err != nil || !hit || read.Value != "value" || read.ExpireAt == nil {
		t.Fatalf("expected expiring value, item=%#v hit=%t err=%v", read, hit, err)
	}
	time.Sleep(40 * time.Millisecond)
	if _, hit, err = service.Get(ctx, kvcache.OwnerTypePlugin, cacheKey); err != nil || hit {
		t.Fatalf("expected expired value miss, hit=%t err=%v", hit, err)
	}
}

// TestMemoryBackendDeleteAndProcessLossAreMisses verifies cache deletion and a
// fresh backend instance both behave as cache misses.
func TestMemoryBackendDeleteAndProcessLossAreMisses(t *testing.T) {
	ctx := context.Background()
	cacheKey := kvcache.BuildCacheKey("unit-owner", "memory", "lossy")
	service := kvcache.New(kvcache.WithProvider(kvcache.NewMemoryProvider()))

	if _, err := service.Set(ctx, kvcache.OwnerTypePlugin, cacheKey, "value", time.Second); err != nil {
		t.Fatalf("set value: %v", err)
	}
	if err := service.Delete(ctx, kvcache.OwnerTypePlugin, cacheKey); err != nil {
		t.Fatalf("delete value: %v", err)
	}
	if _, ok, err := service.Get(ctx, kvcache.OwnerTypePlugin, cacheKey); err != nil || ok {
		t.Fatalf("expected deleted value miss, ok=%t err=%v", ok, err)
	}
	if _, err := service.Set(ctx, kvcache.OwnerTypePlugin, cacheKey, "value", time.Second); err != nil {
		t.Fatalf("set value before process-loss simulation: %v", err)
	}
	restartedService := kvcache.New(kvcache.WithProvider(kvcache.NewMemoryProvider()))
	if _, ok, err := restartedService.Get(ctx, kvcache.OwnerTypePlugin, cacheKey); err != nil || ok {
		t.Fatalf("expected fresh memory backend miss, ok=%t err=%v", ok, err)
	}
}
