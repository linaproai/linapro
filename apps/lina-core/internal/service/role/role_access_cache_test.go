// This file covers local token access-context cache behavior, invalidation,
// and cloning safety for request-scoped mutations.

package role

import (
	"context"
	"reflect"
	"testing"
)

func TestTokenAccessContextCacheLifecycle(t *testing.T) {
	ctx := context.Background()
	svc := New().(*serviceImpl)
	svc.clearLocalAccessCache(ctx)
	t.Cleanup(func() {
		svc.clearLocalAccessCache(ctx)
	})

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
	svc.clearLocalAccessCache(ctx)
	t.Cleanup(func() {
		svc.clearLocalAccessCache(ctx)
	})

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
