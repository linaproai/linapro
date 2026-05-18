// This file verifies backend-neutral coordination behavior using the in-memory
// provider.

package coordination

import (
	"context"
	"net"
	"testing"
	"time"

	"lina-core/pkg/bizerr"
)

// TestKeyBuilderEncodesSegments verifies generated keys are namespaced and
// reject missing required segments.
func TestKeyBuilderEncodesSegments(t *testing.T) {
	keys := NewKeyBuilder("app one", "dev", "node-a")

	lockKey, err := keys.LockKey("leader election")
	if err != nil {
		t.Fatalf("build lock key: %v", err)
	}
	if lockKey == "app one:dev:node-a:lock:leader election" {
		t.Fatalf("expected encoded key segments, got %q", lockKey)
	}
	if _, err = keys.LockKey(" "); !bizerr.Is(err, CodeCoordinationKeyInvalid) {
		t.Fatalf("expected invalid key error, got %v", err)
	}
}

// TestMemoryLockStoreVerifiesOwnerToken covers acquire, renew, ownership, and
// release semantics for the LockStore contract.
func TestMemoryLockStoreVerifiesOwnerToken(t *testing.T) {
	ctx := context.Background()
	lockStore := NewMemory(nil).Lock()

	handle, ok, err := lockStore.Acquire(ctx, "unit-lock", "node-a", "unit", time.Second)
	if err != nil {
		t.Fatalf("acquire lock: %v", err)
	}
	if !ok || handle == nil {
		t.Fatal("expected first acquire to succeed")
	}

	competing, ok, err := lockStore.Acquire(ctx, "unit-lock", "node-b", "unit", time.Second)
	if err != nil {
		t.Fatalf("acquire competing lock: %v", err)
	}
	if ok || competing != nil {
		t.Fatal("expected competing lock acquire to miss")
	}

	if err = lockStore.Renew(ctx, handle, time.Second); err != nil {
		t.Fatalf("renew lock: %v", err)
	}
	held, err := lockStore.IsHeld(ctx, handle)
	if err != nil {
		t.Fatalf("check held: %v", err)
	}
	if !held {
		t.Fatal("expected renewed lock to be held")
	}

	stale := *handle
	stale.Token = "other-token"
	if err = lockStore.Release(ctx, &stale); !bizerr.Is(err, CodeCoordinationLockNotHeld) {
		t.Fatalf("expected stale release to fail with lock-not-held, got %v", err)
	}
	if err = lockStore.Release(ctx, handle); err != nil {
		t.Fatalf("release lock: %v", err)
	}
}

// TestMemoryKVStoreHonorsTTLAndCompareDelete verifies TTL visibility and
// compare-delete semantics.
func TestMemoryKVStoreHonorsTTLAndCompareDelete(t *testing.T) {
	ctx := context.Background()
	kv := NewMemory(nil).KV()

	if err := kv.Set(ctx, "unit-key", "value", 20*time.Millisecond); err != nil {
		t.Fatalf("set kv: %v", err)
	}
	if value, ok, err := kv.Get(ctx, "unit-key"); err != nil || !ok || value != "value" {
		t.Fatalf("expected visible value, value=%q ok=%t err=%v", value, ok, err)
	}
	time.Sleep(40 * time.Millisecond)
	if _, ok, err := kv.Get(ctx, "unit-key"); err != nil || ok {
		t.Fatalf("expected expired value to be hidden, ok=%t err=%v", ok, err)
	}

	if ok, err := kv.SetNX(ctx, "unit-key", "fresh", time.Second); err != nil || !ok {
		t.Fatalf("expected setnx after expiry, ok=%t err=%v", ok, err)
	}
	if ok, err := kv.CompareAndDelete(ctx, "unit-key", "wrong"); err != nil || ok {
		t.Fatalf("expected compare delete mismatch, ok=%t err=%v", ok, err)
	}
	if ok, err := kv.CompareAndDelete(ctx, "unit-key", "fresh"); err != nil || !ok {
		t.Fatalf("expected compare delete match, ok=%t err=%v", ok, err)
	}
}

// TestMemoryRevisionStoreIsSharedByKey verifies revision increments are scoped
// by tenant, domain, and scope.
func TestMemoryRevisionStoreIsSharedByKey(t *testing.T) {
	ctx := context.Background()
	revisions := NewMemory(nil).Revision()
	key := RevisionKey{TenantID: 1, Domain: "role", Scope: "global"}

	first, err := revisions.Bump(ctx, key, "first")
	if err != nil {
		t.Fatalf("bump first revision: %v", err)
	}
	second, err := revisions.Bump(ctx, key, "second")
	if err != nil {
		t.Fatalf("bump second revision: %v", err)
	}
	current, err := revisions.Current(ctx, key)
	if err != nil {
		t.Fatalf("read current revision: %v", err)
	}
	if first != 1 || second != 2 || current != 2 {
		t.Fatalf("unexpected revisions first=%d second=%d current=%d", first, second, current)
	}
}

// TestMemoryEventBusPublishesToSubscribers verifies published events are
// delivered and subscription close stops delivery.
func TestMemoryEventBusPublishesToSubscribers(t *testing.T) {
	ctx := context.Background()
	events := NewMemory(nil).Events()
	received := make(chan Event, 1)

	subscription, err := events.Subscribe(ctx, func(_ context.Context, event Event) error {
		received <- event
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err = events.Publish(ctx, Event{ID: "event-1", Kind: "cache.invalidate"}); err != nil {
		t.Fatalf("publish: %v", err)
	}
	select {
	case event := <-received:
		if event.ID != "event-1" {
			t.Fatalf("expected event-1, got %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("expected event delivery")
	}
	if err = subscription.Close(ctx); err != nil {
		t.Fatalf("close subscription: %v", err)
	}
}

// TestRedisProviderReportsUnavailableWhenEndpointCloses verifies startup
// probing returns a structured Redis-unavailable error for broken endpoints.
func TestRedisProviderReportsUnavailableWhenEndpointCloses(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on local tcp endpoint: %v", err)
	}
	t.Cleanup(func() {
		if err := listener.Close(); err != nil {
			t.Fatalf("close local tcp listener: %v", err)
		}
	})

	accepted := make(chan struct{})
	go func() {
		defer close(accepted)
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("close accepted tcp connection: %v", closeErr)
		}
	}()

	_, err = NewRedis(context.Background(), RedisOptions{
		Address:        listener.Addr().String(),
		ConnectTimeout: 100 * time.Millisecond,
		ReadTimeout:    100 * time.Millisecond,
		WriteTimeout:   100 * time.Millisecond,
		KeyBuilder:     NewKeyBuilder("linapro-test", "redis-failure", "unit"),
	})
	if !bizerr.Is(err, CodeCoordinationRedisUnavailable) {
		t.Fatalf("expected Redis unavailable error, got %v", err)
	}

	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("expected local endpoint to accept Redis probe")
	}
}

// TestRedisProviderReportsUnavailableWhenEndpointTimesOut verifies startup
// probing respects configured timeouts for endpoints that accept but do not
// speak the Redis protocol.
func TestRedisProviderReportsUnavailableWhenEndpointTimesOut(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on local tcp endpoint: %v", err)
	}
	t.Cleanup(func() {
		if err := listener.Close(); err != nil {
			t.Fatalf("close local tcp listener: %v", err)
		}
	})

	accepted := make(chan struct{})
	release := make(chan struct{})
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			close(accepted)
			return
		}
		close(accepted)
		<-release
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("close accepted tcp connection: %v", closeErr)
		}
	}()
	t.Cleanup(func() {
		close(release)
	})

	_, err = NewRedis(context.Background(), RedisOptions{
		Address:        listener.Addr().String(),
		ConnectTimeout: 100 * time.Millisecond,
		ReadTimeout:    20 * time.Millisecond,
		WriteTimeout:   20 * time.Millisecond,
		KeyBuilder:     NewKeyBuilder("linapro-test", "redis-timeout", "unit"),
	})
	if !bizerr.Is(err, CodeCoordinationRedisUnavailable) {
		t.Fatalf("expected Redis unavailable timeout error, got %v", err)
	}

	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("expected local endpoint to accept Redis probe")
	}
}
