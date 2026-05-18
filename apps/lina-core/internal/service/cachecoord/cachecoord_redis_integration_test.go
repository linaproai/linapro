// This file verifies cache coordination against a real Redis coordination
// backend when explicitly enabled by the test environment.

package cachecoord

import (
	"context"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"lina-core/internal/service/coordination"
)

// redisCacheCoordTestTopology gives Redis integration tests distinct node IDs.
type redisCacheCoordTestTopology struct {
	nodeID string
}

// IsEnabled reports clustered mode for Redis cachecoord integration tests.
func (t redisCacheCoordTestTopology) IsEnabled() bool {
	return true
}

// IsPrimary reports the test node as primary.
func (t redisCacheCoordTestTopology) IsPrimary() bool {
	return true
}

// NodeID returns the test node identifier.
func (t redisCacheCoordTestTopology) NodeID() string {
	if t.nodeID == "" {
		return "redis-cachecoord-test-node"
	}
	return t.nodeID
}

// TestRedisCacheCoordIntegrationConcurrentRevisionAndEvent verifies Redis
// backed cachecoord revisions are atomic across concurrent publishers and that
// another node receives the Redis pub/sub notification needed to refresh local
// state.
func TestRedisCacheCoordIntegrationConcurrentRevisionAndEvent(t *testing.T) {
	ctx := context.Background()
	keys := newRedisCacheCoordIntegrationKeyBuilder(t)
	writerCoord := newRedisCacheCoordIntegrationService(t, keys)
	readerCoord := newRedisCacheCoordIntegrationService(t, keys)
	publisher := NewWithCoordination(redisCacheCoordTestTopology{nodeID: "redis-cachecoord-writer"}, writerCoord)
	consumer := NewWithCoordination(redisCacheCoordTestTopology{nodeID: "redis-cachecoord-reader"}, readerCoord)
	domain := Domain("redis-cachecoord")
	scope := Scope("concurrent-event")

	revisionRedisKey, err := keys.RevisionKey(coordination.RevisionKey{
		TenantID: 0,
		Domain:   string(domain),
		Scope:    string(scope),
	})
	if err != nil {
		t.Fatalf("build cachecoord redis revision key: %v", err)
	}
	t.Cleanup(func() {
		cleanupRedisCacheCoordIntegrationKeys(t, revisionRedisKey)
	})

	const workers = 8
	events := make(chan coordination.Event, workers)
	refreshed := make(chan int64, workers)
	handlerErrs := make(chan error, workers)
	subscription, err := readerCoord.Events().Subscribe(ctx, func(handlerCtx context.Context, event coordination.Event) error {
		if event.Kind != cacheInvalidateEventKind ||
			event.Domain != string(domain) ||
			event.Scope != string(scope) {
			return nil
		}
		events <- event
		_, ensureErr := consumer.EnsureFresh(handlerCtx, domain, scope, func(_ context.Context, revision int64) error {
			refreshed <- revision
			return nil
		})
		if ensureErr != nil {
			handlerErrs <- ensureErr
			return ensureErr
		}
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe cachecoord redis events: %v", err)
	}
	t.Cleanup(func() {
		if err := subscription.Close(ctx); err != nil {
			t.Fatalf("close cachecoord redis subscription: %v", err)
		}
	})

	revisions := publishRedisCacheCoordChanges(t, ctx, publisher, domain, scope, workers)
	assertRedisCacheCoordRevisions(t, revisions, workers)

	current, err := consumer.CurrentRevision(ctx, domain, scope)
	if err != nil {
		t.Fatalf("read cachecoord redis revision: %v", err)
	}
	if current != workers {
		t.Fatalf("expected redis cachecoord revision %d, got %d", workers, current)
	}

	event := waitForRedisCacheCoordEvent(t, events, handlerErrs, 3*time.Second)
	if event.SourceNode != "redis-cachecoord-writer" {
		t.Fatalf("expected event source node redis-cachecoord-writer, got %#v", event)
	}
	if revision := waitForRedisCacheCoordRefresh(t, refreshed, handlerErrs, workers, 3*time.Second); revision != workers {
		t.Fatalf("expected redis event refresh revision %d, got %d", workers, revision)
	}

	items, err := consumer.Snapshot(ctx)
	if err != nil {
		t.Fatalf("read cachecoord redis snapshot: %v", err)
	}
	assertRedisCacheCoordSnapshot(t, items, domain, scope, workers)
}

// newRedisCacheCoordIntegrationKeyBuilder creates a unique Redis key namespace
// and skips unless Redis integration tests are explicitly enabled.
func newRedisCacheCoordIntegrationKeyBuilder(t *testing.T) *coordination.KeyBuilder {
	t.Helper()

	if os.Getenv("LINA_TEST_REDIS_ADDR") == "" {
		t.Skip("set LINA_TEST_REDIS_ADDR to enable Redis cachecoord integration tests")
	}
	return coordination.NewKeyBuilder(
		"linapro-test",
		"cachecoord-redis",
		strconv.FormatInt(time.Now().UnixNano(), 10),
	)
}

// newRedisCacheCoordIntegrationService creates one Redis-backed coordination
// service sharing the provided test key namespace.
func newRedisCacheCoordIntegrationService(t *testing.T, keys *coordination.KeyBuilder) coordination.Service {
	t.Helper()

	db := 0
	if rawDB := os.Getenv("LINA_TEST_REDIS_DB"); rawDB != "" {
		parsedDB, err := strconv.Atoi(rawDB)
		if err != nil {
			t.Fatalf("parse LINA_TEST_REDIS_DB: %v", err)
		}
		db = parsedDB
	}

	ctx := context.Background()
	service, err := coordination.NewRedis(ctx, coordination.RedisOptions{
		Address:        os.Getenv("LINA_TEST_REDIS_ADDR"),
		DB:             db,
		Password:       os.Getenv("LINA_TEST_REDIS_PASSWORD"),
		ConnectTimeout: time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
		KeyBuilder:     keys,
	})
	if err != nil {
		t.Fatalf("create redis cachecoord coordination service: %v", err)
	}
	t.Cleanup(func() {
		if err := service.Close(ctx); err != nil {
			t.Fatalf("close redis cachecoord coordination service: %v", err)
		}
	})
	return service
}

// publishRedisCacheCoordChanges publishes concurrent cachecoord revisions and
// returns every revision observed by the publishers.
func publishRedisCacheCoordChanges(
	t *testing.T,
	ctx context.Context,
	publisher Service,
	domain Domain,
	scope Scope,
	workers int,
) []int64 {
	t.Helper()

	revisions := make(chan int64, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			revision, err := publisher.MarkChanged(
				ctx,
				domain,
				scope,
				ChangeReason("redis_cachecoord_concurrent_publish"),
			)
			if err != nil {
				errs <- err
				return
			}
			revisions <- revision
		}()
	}
	wg.Wait()
	close(revisions)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("publish redis cachecoord revision: %v", err)
		}
	}
	values := make([]int64, 0, workers)
	for revision := range revisions {
		values = append(values, revision)
	}
	return values
}

// assertRedisCacheCoordRevisions verifies concurrent Redis revision bumps are
// atomic and no increment is lost.
func assertRedisCacheCoordRevisions(t *testing.T, revisions []int64, workers int) {
	t.Helper()

	if len(revisions) != workers {
		t.Fatalf("expected %d redis cachecoord revisions, got %d: %#v", workers, len(revisions), revisions)
	}
	seen := make(map[int64]struct{}, workers)
	for _, revision := range revisions {
		seen[revision] = struct{}{}
	}
	if len(seen) != workers {
		t.Fatalf("expected %d unique redis cachecoord revisions, got %d: %#v", workers, len(seen), seen)
	}
	for revision := int64(1); revision <= int64(workers); revision++ {
		if _, ok := seen[revision]; !ok {
			t.Fatalf("expected redis cachecoord revision %d in %#v", revision, seen)
		}
	}
}

// waitForRedisCacheCoordEvent waits for the peer-node Redis invalidation event.
func waitForRedisCacheCoordEvent(
	t *testing.T,
	events <-chan coordination.Event,
	errs <-chan error,
	timeout time.Duration,
) coordination.Event {
	t.Helper()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case event := <-events:
		return event
	case err := <-errs:
		t.Fatalf("handle cachecoord redis event: %v", err)
	case <-timer.C:
		t.Fatal("expected redis cachecoord invalidation event")
	}
	return coordination.Event{}
}

// waitForRedisCacheCoordRefresh waits until Redis pub/sub drives the consumer
// instance to the requested revision.
func waitForRedisCacheCoordRefresh(
	t *testing.T,
	revisions <-chan int64,
	errs <-chan error,
	target int64,
	timeout time.Duration,
) int64 {
	t.Helper()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	latest := int64(0)
	for latest < target {
		select {
		case revision := <-revisions:
			if revision > latest {
				latest = revision
			}
		case err := <-errs:
			t.Fatalf("handle cachecoord redis refresh: %v", err)
		case <-timer.C:
			t.Fatalf("expected redis cachecoord refresh revision %d, latest=%d", target, latest)
		}
	}
	return latest
}

// assertRedisCacheCoordSnapshot verifies cachecoord diagnostics reflect Redis
// backend state after the peer-node event refresh.
func assertRedisCacheCoordSnapshot(
	t *testing.T,
	items []SnapshotItem,
	domain Domain,
	scope Scope,
	revision int64,
) {
	t.Helper()

	for _, item := range items {
		if item.Domain != domain || item.Scope != scope {
			continue
		}
		if item.Backend != coordination.BackendRedis ||
			!item.CoordinationHealthy ||
			!item.EventSubscriberRunning ||
			item.LastEventReceivedAt.IsZero() {
			t.Fatalf("expected healthy redis cachecoord diagnostics, got %#v", item)
		}
		if item.LocalRevision != revision || item.SharedRevision != revision {
			t.Fatalf("expected redis cachecoord local/shared revision %d, got %#v", revision, item)
		}
		return
	}
	t.Fatalf("expected redis cachecoord snapshot for %q/%q, got %#v", domain, scope, items)
}

// cleanupRedisCacheCoordIntegrationKeys deletes exact Redis keys created by
// cachecoord integration tests without scanning or flushing the database.
func cleanupRedisCacheCoordIntegrationKeys(t *testing.T, keys ...string) {
	t.Helper()

	address := os.Getenv("LINA_TEST_REDIS_ADDR")
	if address == "" || len(keys) == 0 {
		return
	}
	db := 0
	if rawDB := os.Getenv("LINA_TEST_REDIS_DB"); rawDB != "" {
		parsedDB, err := strconv.Atoi(rawDB)
		if err != nil {
			t.Fatalf("parse LINA_TEST_REDIS_DB for cleanup: %v", err)
		}
		db = parsedDB
	}
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		DB:       db,
		Password: os.Getenv("LINA_TEST_REDIS_PASSWORD"),
	})
	defer func() {
		if err := client.Close(); err != nil {
			t.Fatalf("close redis cachecoord cleanup client: %v", err)
		}
	}()
	if err := client.Del(context.Background(), keys...).Err(); err != nil {
		t.Fatalf("cleanup redis cachecoord integration keys: %v", err)
	}
}
