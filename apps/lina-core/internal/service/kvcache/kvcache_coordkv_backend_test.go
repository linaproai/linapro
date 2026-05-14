// This file verifies the coordination KV kvcache backend adapter with the
// in-memory coordination provider.

package kvcache

import (
	"context"
	"sync"
	"testing"
	"time"

	"lina-core/internal/service/coordination"
	"lina-core/pkg/bizerr"
)

// TestCoordKVBackendStoresStringWithNativeTTL verifies coordination KV-backed
// string values honor TTL and do not require external cleanup.
func TestCoordKVBackendStoresStringWithNativeTTL(t *testing.T) {
	ctx := context.Background()
	service := New(WithProvider(NewCoordinationKVProvider(coordination.NewMemory(nil))))
	cacheKey := BuildCacheKey("unit-owner", "coordkv", "string")

	if service.BackendName() != BackendCoordinationKV {
		t.Fatalf("expected coordination KV backend, got %q", service.BackendName())
	}
	if service.RequiresExpiredCleanup() {
		t.Fatal("expected coordination KV backend to skip expired cleanup")
	}

	item, err := service.Set(ctx, OwnerTypePlugin, cacheKey, "value", 20*time.Millisecond)
	if err != nil {
		t.Fatalf("set coordination KV value: %v", err)
	}
	if item.Value != "value" || item.ValueKind != ValueKindString || item.ExpireAt == nil {
		t.Fatalf("unexpected set item: %#v", item)
	}
	read, ok, err := service.Get(ctx, OwnerTypePlugin, cacheKey)
	if err != nil || !ok || read.Value != "value" {
		t.Fatalf("expected coordination KV value, item=%#v ok=%t err=%v", read, ok, err)
	}
	time.Sleep(40 * time.Millisecond)
	if _, ok, err = service.Get(ctx, OwnerTypePlugin, cacheKey); err != nil || ok {
		t.Fatalf("expected expired coordination KV value miss, ok=%t err=%v", ok, err)
	}
}

// TestCoordKVBackendIncrIsAtomic verifies increments use coordination KV
// atomic integer operations.
func TestCoordKVBackendIncrIsAtomic(t *testing.T) {
	ctx := context.Background()
	service := New(WithProvider(NewCoordinationKVProvider(coordination.NewMemory(nil))))
	cacheKey := BuildCacheKey("unit-owner", "coordkv", "counter")

	const workers = 16
	values := make(chan int64, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			item, err := service.Incr(ctx, OwnerTypePlugin, cacheKey, 1, time.Second)
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
	finalValue, ok, err := service.GetInt(ctx, OwnerTypePlugin, cacheKey)
	if err != nil || !ok || finalValue != workers {
		t.Fatalf("expected final value %d, value=%d ok=%t err=%v", workers, finalValue, ok, err)
	}
}

// TestCoordKVBackendRejectsStringIncrement verifies string payloads cannot be
// incremented as integers.
func TestCoordKVBackendRejectsStringIncrement(t *testing.T) {
	ctx := context.Background()
	service := New(WithProvider(NewCoordinationKVProvider(coordination.NewMemory(nil))))
	cacheKey := BuildCacheKey("unit-owner", "coordkv", "typed")

	if _, err := service.Set(ctx, OwnerTypePlugin, cacheKey, "not-int", 0); err != nil {
		t.Fatalf("seed string value: %v", err)
	}
	if _, err := service.Incr(ctx, OwnerTypePlugin, cacheKey, 1, 0); !bizerr.Is(err, CodeKVCacheIncrementValueNotInteger) {
		t.Fatalf("expected string increment error, got %v", err)
	}
}

// TestCoordKVBackendTenantKeysAreIsolated verifies tenant-aware public keys
// map to isolated coordination keys.
func TestCoordKVBackendTenantKeysAreIsolated(t *testing.T) {
	ctx := context.Background()
	service := New(WithProvider(NewCoordinationKVProvider(coordination.NewMemory(nil))))
	tenantOne := BuildTenantCacheKey(1, "auth", "owner", "coordkv", "tenant")
	tenantTwo := BuildTenantCacheKey(2, "auth", "owner", "coordkv", "tenant")

	if _, err := service.Set(ctx, OwnerTypeModule, tenantOne, "one", 0); err != nil {
		t.Fatalf("set tenant one: %v", err)
	}
	if _, err := service.Set(ctx, OwnerTypeModule, tenantTwo, "two", 0); err != nil {
		t.Fatalf("set tenant two: %v", err)
	}
	itemOne, ok, err := service.Get(ctx, OwnerTypeModule, tenantOne)
	if err != nil || !ok || itemOne.Value != "one" {
		t.Fatalf("expected tenant one value, item=%#v ok=%t err=%v", itemOne, ok, err)
	}
	itemTwo, ok, err := service.Get(ctx, OwnerTypeModule, tenantTwo)
	if err != nil || !ok || itemTwo.Value != "two" {
		t.Fatalf("expected tenant two value, item=%#v ok=%t err=%v", itemTwo, ok, err)
	}
}
