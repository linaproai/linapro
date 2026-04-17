// This file covers local token access-context cache behavior, invalidation,
// and cloning safety for request-scoped mutations.

package role

import (
	"context"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/service/kvcache"
)

type fakeKVCacheService struct {
	getIntValue int64
	getIntErr   error
	getIntCalls int32
	incrCalls   int32
}

func (f *fakeKVCacheService) Get(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
	_ string,
	_ string,
) (*kvcache.Item, bool, error) {
	return nil, false, nil
}

func (f *fakeKVCacheService) GetInt(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
	_ string,
	_ string,
) (int64, bool, error) {
	atomic.AddInt32(&f.getIntCalls, 1)
	return f.getIntValue, true, f.getIntErr
}

func (f *fakeKVCacheService) Set(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
	_ string,
	_ string,
	_ string,
	_ int64,
) (*kvcache.Item, error) {
	return nil, nil
}

func (f *fakeKVCacheService) Delete(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
	_ string,
	_ string,
) error {
	return nil
}

func (f *fakeKVCacheService) Incr(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
	_ string,
	_ string,
	_ int64,
	_ int64,
) (*kvcache.Item, error) {
	atomic.AddInt32(&f.incrCalls, 1)
	return &kvcache.Item{IntValue: f.getIntValue}, nil
}

func (f *fakeKVCacheService) Expire(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
	_ string,
	_ string,
	_ int64,
) (bool, *gtime.Time, error) {
	return false, nil, nil
}

func (f *fakeKVCacheService) CleanupExpired(_ context.Context) error {
	return nil
}

func resetRoleAccessCacheTestState(t *testing.T, svc *serviceImpl) {
	t.Helper()

	ctx := context.Background()
	accessContextCache = gcache.New()
	svc.clearLocalAccessCache(ctx)
	svc.clearLocalAccessRevision()
	t.Cleanup(func() {
		accessContextCache = gcache.New()
		svc.clearLocalAccessCache(ctx)
		svc.clearLocalAccessRevision()
	})
}

func TestTokenAccessContextCacheLifecycle(t *testing.T) {
	ctx := context.Background()
	svc := New().(*serviceImpl)
	resetRoleAccessCacheTestState(t, svc)

	tokenID := "token-cache-lifecycle"
	userID := 101
	access := &UserAccessContext{
		RoleIds:      []int{1, 2},
		RoleNames:    []string{"admin", "editor"},
		MenuIds:      []int{10, 11},
		Permissions:  []string{"system:user:query", "system:user:edit"},
		IsSuperAdmin: false,
	}

	svc.cacheTokenAccessContext(ctx, tokenID, userID, 7, access)

	cached := svc.getCachedTokenAccessContext(ctx, tokenID, userID, 7)
	if !reflect.DeepEqual(cached, access) {
		t.Fatalf("expected cached access %#v, got %#v", access, cached)
	}

	// Returned snapshots must be detached from the cached entry so request-level
	// mutations do not leak into the shared token cache.
	cached.Permissions[0] = "mutated"
	reloaded := svc.getCachedTokenAccessContext(ctx, tokenID, userID, 7)
	if reloaded == nil || reloaded.Permissions[0] != "system:user:query" {
		t.Fatalf("expected cached permissions to stay immutable, got %#v", reloaded)
	}

	if stale := svc.getCachedTokenAccessContext(ctx, tokenID, userID, 8); stale != nil {
		t.Fatalf("expected revision mismatch to force cache miss, got %#v", stale)
	}

	svc.InvalidateTokenAccessContext(ctx, tokenID)
	if stale := svc.getCachedTokenAccessContext(ctx, tokenID, userID, 7); stale != nil {
		t.Fatalf("expected invalidated token cache to be empty, got %#v", stale)
	}
}

func TestInvalidateUserAccessContextsRemovesBoundTokensOnly(t *testing.T) {
	ctx := context.Background()
	svc := New().(*serviceImpl)
	resetRoleAccessCacheTestState(t, svc)

	sharedAccess := &UserAccessContext{
		Permissions: []string{"system:role:auth"},
	}

	svc.cacheTokenAccessContext(ctx, "user-1-token-a", 1, 3, sharedAccess)
	svc.cacheTokenAccessContext(ctx, "user-1-token-b", 1, 3, sharedAccess)
	svc.cacheTokenAccessContext(ctx, "user-2-token-a", 2, 3, sharedAccess)

	svc.InvalidateUserAccessContexts(ctx, 1)

	if access := svc.getCachedTokenAccessContext(ctx, "user-1-token-a", 1, 3); access != nil {
		t.Fatalf("expected first token for invalidated user to be removed, got %#v", access)
	}
	if access := svc.getCachedTokenAccessContext(ctx, "user-1-token-b", 1, 3); access != nil {
		t.Fatalf("expected second token for invalidated user to be removed, got %#v", access)
	}
	if access := svc.getCachedTokenAccessContext(ctx, "user-2-token-a", 2, 3); access == nil {
		t.Fatal("expected other users' cached tokens to remain available")
	}
}

func TestCloneUserAccessContextCopiesSlices(t *testing.T) {
	original := &UserAccessContext{
		RoleIds:      []int{1, 2},
		RoleNames:    []string{"admin", "ops"},
		MenuIds:      []int{10, 20},
		Permissions:  []string{"user:list", "user:update"},
		IsSuperAdmin: true,
	}

	cloned := cloneUserAccessContext(original)
	if cloned == nil {
		t.Fatal("expected cloned access context")
	}

	cloned.RoleIds[0] = 99
	cloned.RoleNames[0] = "guest"
	cloned.MenuIds[0] = 88
	cloned.Permissions[0] = "guest:list"
	cloned.IsSuperAdmin = false

	if original.RoleIds[0] != 1 {
		t.Fatalf("expected original RoleIds to stay unchanged, got %v", original.RoleIds)
	}
	if original.RoleNames[0] != "admin" {
		t.Fatalf("expected original RoleNames to stay unchanged, got %v", original.RoleNames)
	}
	if original.MenuIds[0] != 10 {
		t.Fatalf("expected original MenuIds to stay unchanged, got %v", original.MenuIds)
	}
	if original.Permissions[0] != "user:list" {
		t.Fatalf("expected original Permissions to stay unchanged, got %v", original.Permissions)
	}
	if !original.IsSuperAdmin {
		t.Fatal("expected original IsSuperAdmin to stay unchanged")
	}
}

func TestCloneSliceWithCopyPreservesNilAndValues(t *testing.T) {
	if cloned := cloneSliceWithCopy[int](nil); cloned != nil {
		t.Fatalf("expected nil clone for nil slice, got %#v", cloned)
	}

	values := []string{"a", "b"}
	cloned := cloneSliceWithCopy(values)
	if len(cloned) != len(values) {
		t.Fatalf("expected cloned length %d, got %d", len(values), len(cloned))
	}
	if &cloned[0] == &values[0] {
		t.Fatal("expected cloned slice to have independent backing array")
	}
}

func TestGetAccessRevisionUsesPureReadPath(t *testing.T) {
	ctx := context.Background()
	svc := New().(*serviceImpl)
	resetRoleAccessCacheTestState(t, svc)

	fakeKV := &fakeKVCacheService{getIntValue: 9}
	svc.kvCacheSvc = fakeKV

	revision, err := svc.getAccessRevision(ctx)
	if err != nil {
		t.Fatalf("get access revision failed: %v", err)
	}
	if revision != 9 {
		t.Fatalf("expected revision 9, got %d", revision)
	}
	if atomic.LoadInt32(&fakeKV.getIntCalls) != 1 {
		t.Fatalf("expected exactly one GetInt call, got %d", atomic.LoadInt32(&fakeKV.getIntCalls))
	}
	if atomic.LoadInt32(&fakeKV.incrCalls) != 0 {
		t.Fatalf("expected no Incr calls for read path, got %d", atomic.LoadInt32(&fakeKV.incrCalls))
	}

	revision, err = svc.getAccessRevision(ctx)
	if err != nil {
		t.Fatalf("second get access revision failed: %v", err)
	}
	if revision != 9 {
		t.Fatalf("expected cached revision 9, got %d", revision)
	}
	if atomic.LoadInt32(&fakeKV.getIntCalls) != 1 {
		t.Fatalf("expected cached local revision to avoid extra GetInt calls, got %d", atomic.LoadInt32(&fakeKV.getIntCalls))
	}
}

func TestSyncAccessTopologyRevisionKeepsCacheWhenRevisionUnchanged(t *testing.T) {
	ctx := context.Background()
	svc := New().(*serviceImpl)
	resetRoleAccessCacheTestState(t, svc)

	fakeKV := &fakeKVCacheService{getIntValue: 7}
	svc.kvCacheSvc = fakeKV
	svc.storeLocalAccessRevision(7)
	svc.cacheTokenAccessContext(ctx, "sync-same-revision", 1, 7, &UserAccessContext{
		Permissions: []string{"system:user:list"},
	})

	if err := svc.SyncAccessTopologyRevision(ctx); err != nil {
		t.Fatalf("sync access topology revision failed: %v", err)
	}

	cachedVar, err := accessContextCache.Get(ctx, accessCacheKey("sync-same-revision"))
	if err != nil {
		t.Fatalf("get cached access context after unchanged sync: %v", err)
	}
	if cachedVar == nil {
		t.Fatal("expected cached token access context to remain after unchanged revision sync")
	}
}

func TestSyncAccessTopologyRevisionClearsCacheWhenRevisionChanges(t *testing.T) {
	ctx := context.Background()
	svc := New().(*serviceImpl)
	resetRoleAccessCacheTestState(t, svc)

	fakeKV := &fakeKVCacheService{getIntValue: 8}
	svc.kvCacheSvc = fakeKV
	svc.storeLocalAccessRevision(7)
	svc.cacheTokenAccessContext(ctx, "sync-new-revision", 1, 7, &UserAccessContext{
		Permissions: []string{"system:user:list"},
	})

	if err := svc.SyncAccessTopologyRevision(ctx); err != nil {
		t.Fatalf("sync access topology revision failed: %v", err)
	}

	cachedVar, err := accessContextCache.Get(ctx, accessCacheKey("sync-new-revision"))
	if err != nil {
		t.Fatalf("get cached access context after changed sync: %v", err)
	}
	if cachedVar != nil {
		t.Fatal("expected stale token access context to be evicted after revision change")
	}

	revision, ok := svc.getLocalAccessRevision()
	if !ok {
		t.Fatal("expected synced revision to remain locally cached")
	}
	if revision != 8 {
		t.Fatalf("expected local revision 8 after sync, got %d", revision)
	}
}

func TestLoadTokenAccessContextWithCacheLockSuppressesDuplicateLoads(t *testing.T) {
	ctx := context.Background()
	svc := New().(*serviceImpl)
	resetRoleAccessCacheTestState(t, svc)

	var loadCalls atomic.Int32
	loader := func(context.Context) (*UserAccessContext, error) {
		loadCalls.Add(1)
		time.Sleep(30 * time.Millisecond)
		return &UserAccessContext{
			RoleIds:      []int{1},
			RoleNames:    []string{"admin"},
			MenuIds:      []int{101},
			Permissions:  []string{"system:user:list"},
			IsSuperAdmin: true,
		}, nil
	}

	const workers = 8
	results := make(chan *UserAccessContext, workers)
	errs := make(chan error, workers)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			access, err := svc.loadTokenAccessContextWithCacheLock(
				ctx,
				"concurrent-token",
				1,
				3,
				loader,
			)
			if err != nil {
				errs <- err
				return
			}
			results <- access
		}()
	}

	close(start)
	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("load token access context with cache lock failed: %v", err)
		}
	}
	if loadCalls.Load() != 1 {
		t.Fatalf("expected exactly one cold-load execution, got %d", loadCalls.Load())
	}

	count := 0
	for access := range results {
		count++
		if access == nil {
			t.Fatal("expected non-nil access context from cache lock loader")
		}
		if len(access.Permissions) != 1 || access.Permissions[0] != "system:user:list" {
			t.Fatalf("unexpected permissions from cached access context: %#v", access)
		}
	}
	if count != workers {
		t.Fatalf("expected %d access results, got %d", workers, count)
	}
}
