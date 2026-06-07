// This file verifies the source-plugin cache adapter keeps plugin-visible
// cache calls scoped to plugin identity, tenant identity, and the shared host
// kvcache service.

package hostservices

import (
	"context"
	"errors"
	"testing"
	"time"

	"lina-core/internal/service/coordination"
	"lina-core/internal/service/kvcache"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// TestCacheAdapterUsesSharedKVCache verifies common cache operations delegate
// to the injected host kvcache service.
func TestCacheAdapterUsesSharedKVCache(t *testing.T) {
	ctx := context.Background()
	cacheSvc := kvcache.New(kvcache.WithProvider(kvcache.NewCoordinationKVProvider(coordination.NewMemory(nil))))
	adapter := newCacheAdapter(cacheSvc, nil, "source-plugin-a")

	item, err := adapter.Set(ctx, "profiles", "current", "value", time.Minute)
	if err != nil {
		t.Fatalf("set source plugin cache: %v", err)
	}
	if item == nil || item.ValueKind != cachecap.CacheValueKindString || item.Value != "value" {
		t.Fatalf("unexpected set item: %#v", item)
	}
	if item.Key != "current" {
		t.Fatalf("expected plugin-visible logical key, got %q", item.Key)
	}

	read, found, err := adapter.Get(ctx, "profiles", "current")
	if err != nil || !found || read == nil || read.Value != "value" {
		t.Fatalf("expected cached value, item=%#v found=%t err=%v", read, found, err)
	}
	if read.Key != "current" {
		t.Fatalf("expected plugin-visible logical key on read, got %q", read.Key)
	}

	counter, err := adapter.Incr(ctx, "profiles", "counter", 2, time.Minute)
	if err != nil {
		t.Fatalf("increment source plugin cache: %v", err)
	}
	if counter == nil || counter.ValueKind != cachecap.CacheValueKindInt || counter.IntValue != 2 {
		t.Fatalf("unexpected counter item: %#v", counter)
	}

	found, expireAt, err := adapter.Expire(ctx, "profiles", "counter", time.Minute)
	if err != nil || !found || expireAt == nil {
		t.Fatalf("expected expire success, found=%t expireAt=%v err=%v", found, expireAt, err)
	}

	if err = adapter.Delete(ctx, "profiles", "current"); err != nil {
		t.Fatalf("delete source plugin cache: %v", err)
	}
	if _, found, err = adapter.Get(ctx, "profiles", "current"); err != nil || found {
		t.Fatalf("expected deleted cache miss, found=%t err=%v", found, err)
	}
}

// TestCacheAdapterScopesByPluginID verifies two plugins using the same
// namespace and key cannot read each other's cached values.
func TestCacheAdapterScopesByPluginID(t *testing.T) {
	ctx := context.Background()
	cacheSvc := kvcache.New(kvcache.WithProvider(kvcache.NewCoordinationKVProvider(coordination.NewMemory(nil))))
	pluginA := newCacheAdapter(cacheSvc, nil, "source-plugin-a")
	pluginB := newCacheAdapter(cacheSvc, nil, "source-plugin-b")

	if _, err := pluginA.Set(ctx, "profiles", "current", "a", 0); err != nil {
		t.Fatalf("set plugin a cache: %v", err)
	}
	if _, err := pluginB.Set(ctx, "profiles", "current", "b", 0); err != nil {
		t.Fatalf("set plugin b cache: %v", err)
	}

	itemA, found, err := pluginA.Get(ctx, "profiles", "current")
	if err != nil || !found || itemA.Value != "a" {
		t.Fatalf("expected plugin a value, item=%#v found=%t err=%v", itemA, found, err)
	}
	itemB, found, err := pluginB.Get(ctx, "profiles", "current")
	if err != nil || !found || itemB.Value != "b" {
		t.Fatalf("expected plugin b value, item=%#v found=%t err=%v", itemB, found, err)
	}
}

// TestCacheAdapterScopesByTenantContext verifies tenant context becomes part
// of the internal cache key when present.
func TestCacheAdapterScopesByTenantContext(t *testing.T) {
	cacheSvc := kvcache.New(kvcache.WithProvider(kvcache.NewCoordinationKVProvider(coordination.NewMemory(nil))))
	adapter := newCacheAdapter(cacheSvc, nil, "source-plugin-a")
	tenantOne := bizctxcap.WithCurrentContext(context.Background(), bizctxcap.CurrentContext{TenantID: 1001})
	tenantTwo := bizctxcap.WithCurrentContext(context.Background(), bizctxcap.CurrentContext{TenantID: 1002})

	if _, err := adapter.Set(tenantOne, "profiles", "current", "tenant-one", 0); err != nil {
		t.Fatalf("set tenant one cache: %v", err)
	}
	if _, err := adapter.Set(tenantTwo, "profiles", "current", "tenant-two", 0); err != nil {
		t.Fatalf("set tenant two cache: %v", err)
	}

	itemOne, found, err := adapter.Get(tenantOne, "profiles", "current")
	if err != nil || !found || itemOne.Value != "tenant-one" {
		t.Fatalf("expected tenant one value, item=%#v found=%t err=%v", itemOne, found, err)
	}
	itemTwo, found, err := adapter.Get(tenantTwo, "profiles", "current")
	if err != nil || !found || itemTwo.Value != "tenant-two" {
		t.Fatalf("expected tenant two value, item=%#v found=%t err=%v", itemTwo, found, err)
	}
}

// TestCacheAdapterRejectsInvalidTTLAndStringIncrement verifies validation and
// type errors from the shared cache service are returned to source plugins.
func TestCacheAdapterRejectsInvalidTTLAndStringIncrement(t *testing.T) {
	ctx := context.Background()
	cacheSvc := kvcache.New(kvcache.WithProvider(kvcache.NewCoordinationKVProvider(coordination.NewMemory(nil))))
	adapter := newCacheAdapter(cacheSvc, nil, "source-plugin-a")

	if _, err := adapter.Set(ctx, "profiles", "invalid-ttl", "value", -time.Second); !bizerr.Is(err, kvcache.CodeKVCacheExpireSecondsNegative) {
		t.Fatalf("expected negative TTL error, got %v", err)
	}
	if _, err := adapter.Set(ctx, "profiles", "typed", "not-int", 0); err != nil {
		t.Fatalf("seed string cache: %v", err)
	}
	if _, err := adapter.Incr(ctx, "profiles", "typed", 1, 0); !bizerr.Is(err, kvcache.CodeKVCacheIncrementValueNotInteger) {
		t.Fatalf("expected string increment error, got %v", err)
	}
}

// TestCacheAdapterReturnsBackendErrors verifies backend failures are not
// converted into cache misses or fake successes.
func TestCacheAdapterReturnsBackendErrors(t *testing.T) {
	expectedErr := errors.New("backend unavailable")
	cacheSvc := kvcache.New(kvcache.WithBackend(failingCacheBackend{err: expectedErr}))
	adapter := newCacheAdapter(cacheSvc, nil, "source-plugin-a")

	if _, _, err := adapter.Get(context.Background(), "profiles", "current"); !errors.Is(err, expectedErr) {
		t.Fatalf("expected get backend error, got %v", err)
	}
	if _, err := adapter.Set(context.Background(), "profiles", "current", "value", 0); !errors.Is(err, expectedErr) {
		t.Fatalf("expected set backend error, got %v", err)
	}
}

// TestCacheAdapterReturnsBizErrorsForMissingScope verifies adapter setup
// failures remain structured for source-plugin callers.
func TestCacheAdapterReturnsBizErrorsForMissingScope(t *testing.T) {
	missingService := newCacheAdapter(nil, nil, "source-plugin-a")
	if _, _, err := missingService.Get(context.Background(), "profiles", "current"); !bizerr.Is(err, CodePluginSourceCacheServiceUnavailable) {
		t.Fatalf("expected missing service bizerr, got %v", err)
	}

	cacheSvc := kvcache.New(kvcache.WithProvider(kvcache.NewCoordinationKVProvider(coordination.NewMemory(nil))))
	missingPluginID := newCacheAdapter(cacheSvc, nil, " ")
	if _, _, err := missingPluginID.Get(context.Background(), "profiles", "current"); !bizerr.Is(err, CodePluginSourceCachePluginIDRequired) {
		t.Fatalf("expected missing plugin id bizerr, got %v", err)
	}
}

// TestServicesForPluginReturnsScopedCache verifies source-plugin host services
// bind cache services to the requested plugin ID.
func TestServicesForPluginReturnsScopedCache(t *testing.T) {
	cacheSvc := kvcache.New(kvcache.WithProvider(kvcache.NewCoordinationKVProvider(coordination.NewMemory(nil))))
	services := &directory{
		bizCtx: nil,
		cache:  cacheSvc,
	}

	scoped := services.ForPlugin("source-plugin-a")
	if scoped == nil || scoped.Cache() == nil {
		t.Fatal("expected scoped host services to expose cache adapter")
	}
	if services.Cache() != nil {
		t.Fatal("expected unscoped base services cache to remain unavailable")
	}
}

// TestServicesExposeCapabilitiesThroughScopedServices verifies the common
// capability helper returns capability-owned consumers.
func TestServicesExposeCapabilitiesThroughScopedServices(t *testing.T) {
	services := &directory{
		org:    orgcap.New(nil),
		tenant: tenantcap.New(nil, nil),
	}

	scoped := capability.ServicesForPlugin(services, "source-plugin-a")
	if scoped == nil {
		t.Fatal("expected scoped services")
	}
	if scoped.Org() == nil {
		t.Fatal("expected organization capability consumer")
	}
	if scoped.Tenant() == nil {
		t.Fatal("expected tenant capability consumer")
	}
}

// failingCacheBackend returns the configured error for every operation.
type failingCacheBackend struct {
	err error
}

// Name returns the stable test backend name.
func (b failingCacheBackend) Name() kvcache.BackendName { return "failing" }

// RequiresExpiredCleanup reports that the failing backend needs no cleanup.
func (b failingCacheBackend) RequiresExpiredCleanup() bool { return false }

// Get returns the configured backend error.
func (b failingCacheBackend) Get(context.Context, kvcache.OwnerType, string) (*kvcache.Item, bool, error) {
	return nil, false, b.err
}

// GetInt returns the configured backend error.
func (b failingCacheBackend) GetInt(context.Context, kvcache.OwnerType, string) (int64, bool, error) {
	return 0, false, b.err
}

// Set returns the configured backend error.
func (b failingCacheBackend) Set(
	context.Context,
	kvcache.OwnerType,
	string,
	string,
	time.Duration,
) (*kvcache.Item, error) {
	return nil, b.err
}

// Delete returns the configured backend error.
func (b failingCacheBackend) Delete(context.Context, kvcache.OwnerType, string) error {
	return b.err
}

// Incr returns the configured backend error.
func (b failingCacheBackend) Incr(
	context.Context,
	kvcache.OwnerType,
	string,
	int64,
	time.Duration,
) (*kvcache.Item, error) {
	return nil, b.err
}

// Expire returns the configured backend error.
func (b failingCacheBackend) Expire(
	context.Context,
	kvcache.OwnerType,
	string,
	time.Duration,
) (bool, *time.Time, error) {
	return false, nil, b.err
}

// CleanupExpired returns the configured backend error.
func (b failingCacheBackend) CleanupExpired(context.Context) error {
	return b.err
}
