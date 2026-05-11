// This file verifies cache coordination revision behavior.

package cachecoord

import (
	"context"
	"sync"
	"testing"

	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
)

const (
	// testRuntimeConfigDomain mirrors the host runtime-config domain in cachecoord tests.
	testRuntimeConfigDomain Domain = "runtime-config"
	// testPluginRuntimeDomain mirrors the host plugin-runtime domain in cachecoord tests.
	testPluginRuntimeDomain Domain = "plugin-runtime"
)

// defaultCoordinatorTestTopology is a non-static topology used to verify that
// real cluster topology wiring is not replaced by later static placeholders.
type defaultCoordinatorTestTopology struct {
	enabled bool
}

// IsEnabled reports the configured cluster mode for the test topology.
func (t defaultCoordinatorTestTopology) IsEnabled() bool {
	return t.enabled
}

// IsPrimary reports this test node as primary.
func (defaultCoordinatorTestTopology) IsPrimary() bool {
	return true
}

// NodeID returns a stable test node identifier.
func (defaultCoordinatorTestTopology) NodeID() string {
	return "default-test-node"
}

// TestSingleNodeMarkChangedUsesProcessLocalRevision verifies local mode avoids
// the shared revision table.
func TestSingleNodeMarkChangedUsesProcessLocalRevision(t *testing.T) {
	ctx := context.Background()
	service := New(NewStaticTopology(false))

	firstRevision, err := service.MarkChanged(
		ctx,
		testRuntimeConfigDomain,
		ScopeGlobal,
		ChangeReason("unit_test_local_first"),
	)
	if err != nil {
		t.Fatalf("first local mark failed: %v", err)
	}
	secondRevision, err := service.MarkChanged(
		ctx,
		testRuntimeConfigDomain,
		ScopeGlobal,
		ChangeReason("unit_test_local_second"),
	)
	if err != nil {
		t.Fatalf("second local mark failed: %v", err)
	}
	if secondRevision != firstRevision+1 {
		t.Fatalf("expected local revision to increment from %d to %d, got %d", firstRevision, firstRevision+1, secondRevision)
	}
}

// TestTenantScopedMarkChangedIsolatesLocalRevisions verifies tenant invalidation scope uses separate revisions.
func TestTenantScopedMarkChangedIsolatesLocalRevisions(t *testing.T) {
	ctx := context.Background()
	service := New(NewStaticTopology(false))
	domain := Domain("unit-tenant-cache")
	scope := Scope("dict")

	tenantOneRevision, err := service.MarkTenantChanged(
		ctx,
		domain,
		scope,
		InvalidationScope{TenantID: 1},
		ChangeReason("tenant_one"),
	)
	if err != nil {
		t.Fatalf("tenant one mark failed: %v", err)
	}
	tenantTwoRevision, err := service.MarkTenantChanged(
		ctx,
		domain,
		scope,
		InvalidationScope{TenantID: 2},
		ChangeReason("tenant_two"),
	)
	if err != nil {
		t.Fatalf("tenant two mark failed: %v", err)
	}
	if tenantOneRevision != 1 || tenantTwoRevision != 1 {
		t.Fatalf("expected isolated first revisions, got tenant1=%d tenant2=%d", tenantOneRevision, tenantTwoRevision)
	}
}

// TestTenantScopedMarkChangedCascadeUsesDistinctScope verifies platform
// cascade invalidation does not overwrite a tenant-only revision bucket.
func TestTenantScopedMarkChangedCascadeUsesDistinctScope(t *testing.T) {
	ctx := context.Background()
	service := New(NewStaticTopology(false))
	domain := Domain("unit-tenant-cache-cascade")
	scope := Scope("permission")

	tenantRevision, err := service.MarkTenantChanged(
		ctx,
		domain,
		scope,
		InvalidationScope{TenantID: 9},
		ChangeReason("tenant_only"),
	)
	if err != nil {
		t.Fatalf("tenant mark failed: %v", err)
	}
	cascadeRevision, err := service.MarkTenantChanged(
		ctx,
		domain,
		scope,
		InvalidationScope{TenantID: 0, CascadeToTenants: true},
		ChangeReason("platform_cascade"),
	)
	if err != nil {
		t.Fatalf("platform cascade mark failed: %v", err)
	}
	if tenantRevision != 1 || cascadeRevision != 1 {
		t.Fatalf("expected isolated tenant/cascade first revisions, got tenant=%d cascade=%d", tenantRevision, cascadeRevision)
	}
	if scoped := ScopedScope(scope, InvalidationScope{TenantID: 9}); scoped == scope {
		t.Fatalf("expected tenant scope to include tenant discriminator, got %q", scoped)
	}
	if cascade := ScopedScope(scope, InvalidationScope{TenantID: 0, CascadeToTenants: true}); cascade == scope {
		t.Fatalf("expected cascade scope to include cascade discriminator, got %q", cascade)
	}
}

// TestDefaultReturnsSharedCoordinatorWithUpdatedTopology verifies production
// constructors reuse one process coordinator while later startup wiring can
// replace the topology view with the real cluster service.
func TestDefaultReturnsSharedCoordinatorWithUpdatedTopology(t *testing.T) {
	first := Default(NewStaticTopology(false))
	impl, ok := first.(*serviceImpl)
	if !ok {
		t.Fatalf("expected default coordinator implementation, got %T", first)
	}
	if impl.clusterEnabled() {
		t.Fatal("expected initial default coordinator topology to be single-node")
	}

	second := Default(NewStaticTopology(true))
	if first != second {
		t.Fatal("expected default coordinator to reuse the same process service")
	}
	if !impl.clusterEnabled() {
		t.Fatal("expected default coordinator topology to be updated to cluster mode")
	}

	third := Default(defaultCoordinatorTestTopology{enabled: true})
	if first != third {
		t.Fatal("expected real topology wiring to reuse the same process service")
	}
	if _, ok := impl.topologySnapshot().(defaultCoordinatorTestTopology); !ok {
		t.Fatalf("expected real topology to be wired, got %T", impl.topologySnapshot())
	}

	fourth := Default(NewStaticTopology(false))
	if first != fourth {
		t.Fatal("expected later static topology calls to reuse the same process service")
	}
	if _, ok := impl.topologySnapshot().(defaultCoordinatorTestTopology); !ok {
		t.Fatalf("expected real topology to survive later static placeholder, got %T", impl.topologySnapshot())
	}

	fifth := Default(defaultCoordinatorTestTopology{enabled: false})
	if first != fifth {
		t.Fatal("expected later disabled topology calls to reuse the same process service")
	}
	if !impl.clusterEnabled() {
		t.Fatal("expected enabled topology to survive later disabled service wiring")
	}
}

// TestClusterMarkChangedPersistsAtomicRevision verifies concurrent clustered
// publishers increment the same persistent row without losing revisions.
func TestClusterMarkChangedPersistsAtomicRevision(t *testing.T) {
	ctx := context.Background()
	service := New(NewStaticTopology(true))
	cleanupCacheRevision(t, ctx, testRuntimeConfigDomain, Scope("unit-test-atomic"))

	const workers = 12
	revisions := make(chan int64, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			revision, err := service.MarkChanged(
				ctx,
				testRuntimeConfigDomain,
				Scope("unit-test-atomic"),
				ChangeReason("unit_test_concurrent_publish"),
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
			t.Fatalf("concurrent publish failed: %v", err)
		}
	}

	seen := make(map[int64]struct{}, workers)
	for revision := range revisions {
		seen[revision] = struct{}{}
	}
	if len(seen) != workers {
		t.Fatalf("expected %d unique revisions, got %d: %#v", workers, len(seen), seen)
	}

	latest, err := service.CurrentRevision(ctx, testRuntimeConfigDomain, Scope("unit-test-atomic"))
	if err != nil {
		t.Fatalf("read shared revision failed: %v", err)
	}
	if latest != workers {
		t.Fatalf("expected latest revision %d, got %d", workers, latest)
	}
}

// TestClusterMarkChangedAcceptsUnconfiguredDomain verifies callers can use a
// new valid domain without changing cachecoord code or configuring metadata.
func TestClusterMarkChangedAcceptsUnconfiguredDomain(t *testing.T) {
	ctx := context.Background()
	service := New(NewStaticTopology(true))
	domain := Domain("plugin:unit-test:custom")
	scope := Scope("unit-test-free-domain")
	cleanupCacheRevision(t, ctx, domain, scope)

	revision, err := service.MarkChanged(ctx, domain, scope, ChangeReason("free_domain"))
	if err != nil {
		t.Fatalf("publish unconfigured domain failed: %v", err)
	}
	if revision != 1 {
		t.Fatalf("expected first unconfigured domain revision 1, got %d", revision)
	}

	items, err := service.Snapshot(ctx)
	if err != nil {
		t.Fatalf("snapshot unconfigured domain failed: %v", err)
	}
	for _, item := range items {
		if item.Domain != domain || item.Scope != scope {
			continue
		}
		if item.AuthoritySource != "caller-owned cache domain" ||
			item.ConsistencyModel != ConsistencySharedRevision ||
			item.MaxStale != DefaultDomainMaxStale ||
			item.FailureStrategy != FailureStrategyReturnVisibleError {
			t.Fatalf("expected default domain spec, got %#v", item)
		}
		return
	}
	t.Fatalf("expected snapshot item for unconfigured domain %q/%q, got %#v", domain, scope, items)
}

// TestClusterCurrentRevisionHandlesMissingSharedRow verifies first reads in
// cluster mode treat a missing revision row as revision zero instead of an
// infrastructure failure.
func TestClusterCurrentRevisionHandlesMissingSharedRow(t *testing.T) {
	ctx := context.Background()
	service := New(NewStaticTopology(true))
	scope := Scope("unit-test-missing-shared-row")
	cleanupCacheRevision(t, ctx, testRuntimeConfigDomain, scope)

	revision, err := service.CurrentRevision(ctx, testRuntimeConfigDomain, scope)
	if err != nil {
		t.Fatalf("expected missing shared revision row to read as zero, got error: %v", err)
	}
	if revision != 0 {
		t.Fatalf("expected missing shared revision row to return 0, got %d", revision)
	}
}

// TestEnsureFreshRefreshesOncePerRevision verifies the refresher only runs when
// the observed revision advances.
func TestEnsureFreshRefreshesOncePerRevision(t *testing.T) {
	ctx := context.Background()
	publisher := New(NewStaticTopology(true))
	consumer := New(NewStaticTopology(true))
	cleanupCacheRevision(t, ctx, testPluginRuntimeDomain, Scope("unit-test-refresh"))

	if _, err := publisher.MarkChanged(ctx, testPluginRuntimeDomain, Scope("unit-test-refresh"), ChangeReason("first")); err != nil {
		t.Fatalf("publish first revision failed: %v", err)
	}

	refreshCalls := 0
	refresher := func(_ context.Context, revision int64) error {
		refreshCalls++
		if revision != 1 && revision != 2 {
			t.Fatalf("unexpected refresh revision %d", revision)
		}
		return nil
	}
	if _, err := consumer.EnsureFresh(ctx, testPluginRuntimeDomain, Scope("unit-test-refresh"), refresher); err != nil {
		t.Fatalf("first ensure fresh failed: %v", err)
	}
	if _, err := consumer.EnsureFresh(ctx, testPluginRuntimeDomain, Scope("unit-test-refresh"), refresher); err != nil {
		t.Fatalf("second ensure fresh failed: %v", err)
	}
	if refreshCalls != 1 {
		t.Fatalf("expected one refresh for first revision, got %d", refreshCalls)
	}

	if _, err := publisher.MarkChanged(ctx, testPluginRuntimeDomain, Scope("unit-test-refresh"), ChangeReason("second")); err != nil {
		t.Fatalf("publish second revision failed: %v", err)
	}
	if _, err := consumer.EnsureFresh(ctx, testPluginRuntimeDomain, Scope("unit-test-refresh"), refresher); err != nil {
		t.Fatalf("third ensure fresh failed: %v", err)
	}
	if refreshCalls != 2 {
		t.Fatalf("expected second refresh after revision change, got %d", refreshCalls)
	}
}

// TestSnapshotIncludesProcessStatusFromOtherInstances verifies diagnostics can
// expose status recorded by cachecoord users that own separate service instances.
func TestSnapshotIncludesProcessStatusFromOtherInstances(t *testing.T) {
	ctx := context.Background()
	scope := Scope("unit-test-process-snapshot")
	publisher := New(NewStaticTopology(true))
	diagnosticReader := New(NewStaticTopology(true))
	cleanupCacheRevision(t, ctx, testRuntimeConfigDomain, scope)

	revision, err := publisher.MarkChanged(ctx, testRuntimeConfigDomain, scope, ChangeReason("diagnostic_snapshot"))
	if err != nil {
		t.Fatalf("publish revision failed: %v", err)
	}

	items, err := diagnosticReader.Snapshot(ctx)
	if err != nil {
		t.Fatalf("snapshot failed: %v", err)
	}
	for _, item := range items {
		if item.Domain != testRuntimeConfigDomain || item.Scope != scope {
			continue
		}
		if item.LocalRevision != revision || item.SharedRevision != revision || item.LastSyncedAt.IsZero() {
			t.Fatalf("expected process status revision %d, got %#v", revision, item)
		}
		return
	}
	t.Fatalf("expected snapshot item for scope %q, got %#v", scope, items)
}

// cleanupCacheRevision removes one test revision row before and after a test.
func cleanupCacheRevision(t *testing.T, ctx context.Context, domain Domain, scope Scope) {
	t.Helper()

	cleanup := func() {
		if _, err := dao.SysCacheRevision.Ctx(ctx).Where(do.SysCacheRevision{
			Domain: domain,
			Scope:  scope,
		}).Delete(); err != nil {
			t.Fatalf("cleanup cache revision failed: %v", err)
		}
	}
	cleanup()
	t.Cleanup(cleanup)
}
