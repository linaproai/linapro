// This file verifies the role-owned dynamic route access projection contract
// consumed by dynamic plugin runtime authentication.

package role

import (
	"context"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/datascope"
)

// TestBuildDynamicRouteAccessProjectionReturnsDetachedSnapshot verifies callers
// cannot mutate the token access snapshot through the dynamic route projection.
func TestBuildDynamicRouteAccessProjectionReturnsDetachedSnapshot(t *testing.T) {
	ctx := context.Background()
	svc := newDefaultRoleTestService()
	resetRoleAccessCacheTestState(t, svc)
	svc.accessRevisionCtrl = &fakeAccessRevisionController{revision: 7}

	access := &UserAccessContext{
		Permissions:          []string{"plugin:route:read"},
		RoleNames:            []string{"tenant-operator"},
		DataScope:            datascope.ScopeTenant,
		DataScopeUnsupported: true,
		UnsupportedDataScope: 99,
		IsSuperAdmin:         true,
	}
	if err := svc.cacheTokenAccessContext(
		datascope.WithTenantForTest(ctx, 51),
		"detached-token",
		11,
		7,
		access,
	); err != nil {
		t.Fatalf("seed token access context: %v", err)
	}

	projection, err := svc.BuildDynamicRouteAccessProjection(ctx, "detached-token", 11, 51)
	if err != nil {
		t.Fatalf("build dynamic route projection: %v", err)
	}
	projection.Permissions[0] = "mutated:permission"
	projection.RoleNames[0] = "mutated-role"

	reloaded, err := svc.BuildDynamicRouteAccessProjection(ctx, "detached-token", 11, 51)
	if err != nil {
		t.Fatalf("reload dynamic route projection: %v", err)
	}
	if reloaded.Permissions[0] != "plugin:route:read" {
		t.Fatalf("expected cached permission to stay detached, got %#v", reloaded.Permissions)
	}
	if reloaded.RoleNames[0] != "tenant-operator" {
		t.Fatalf("expected cached role name to stay detached, got %#v", reloaded.RoleNames)
	}
	if reloaded.DataScope != datascope.ScopeTenant ||
		!reloaded.DataScopeUnsupported ||
		reloaded.UnsupportedDataScope != 99 ||
		!reloaded.IsSuperAdmin {
		t.Fatalf("expected data-scope fields to be preserved, got %#v", reloaded)
	}
}

// TestBuildDynamicRouteAccessProjectionUsesTenantScopedTokenBucket verifies
// identical token IDs in different tenants resolve to isolated access snapshots.
func TestBuildDynamicRouteAccessProjectionUsesTenantScopedTokenBucket(t *testing.T) {
	ctx := context.Background()
	svc := newDefaultRoleTestService()
	resetRoleAccessCacheTestState(t, svc)
	svc.accessRevisionCtrl = &fakeAccessRevisionController{revision: 8}

	tokenID := "shared-dynamic-token"
	if err := svc.cacheTokenAccessContext(
		datascope.WithTenantForTest(ctx, 61),
		tokenID,
		21,
		8,
		&UserAccessContext{
			Permissions: []string{"plugin:tenant-a:read"},
			RoleNames:   []string{"tenant-a"},
			DataScope:   datascope.ScopeSelf,
		},
	); err != nil {
		t.Fatalf("seed tenant A access context: %v", err)
	}
	if err := svc.cacheTokenAccessContext(
		datascope.WithTenantForTest(ctx, 62),
		tokenID,
		22,
		8,
		&UserAccessContext{
			Permissions: []string{"plugin:tenant-b:read"},
			RoleNames:   []string{"tenant-b"},
			DataScope:   datascope.ScopeDept,
		},
	); err != nil {
		t.Fatalf("seed tenant B access context: %v", err)
	}

	tenantA, err := svc.BuildDynamicRouteAccessProjection(ctx, tokenID, 21, 61)
	if err != nil {
		t.Fatalf("build tenant A projection: %v", err)
	}
	tenantB, err := svc.BuildDynamicRouteAccessProjection(ctx, tokenID, 22, 62)
	if err != nil {
		t.Fatalf("build tenant B projection: %v", err)
	}

	if len(tenantA.Permissions) != 1 || tenantA.Permissions[0] != "plugin:tenant-a:read" {
		t.Fatalf("expected tenant A permission, got %#v", tenantA.Permissions)
	}
	if len(tenantB.Permissions) != 1 || tenantB.Permissions[0] != "plugin:tenant-b:read" {
		t.Fatalf("expected tenant B permission, got %#v", tenantB.Permissions)
	}
	if tenantA.DataScope != datascope.ScopeSelf || tenantB.DataScope != datascope.ScopeDept {
		t.Fatalf("expected tenant-scoped data scopes, got A=%d B=%d", tenantA.DataScope, tenantB.DataScope)
	}
}

// TestBuildDynamicRouteAccessProjectionPropagatesFreshnessFailure verifies
// permission revision failures keep dynamic route auth fail-closed.
func TestBuildDynamicRouteAccessProjectionPropagatesFreshnessFailure(t *testing.T) {
	ctx := context.Background()
	svc := newDefaultRoleTestService()
	resetRoleAccessCacheTestState(t, svc)
	svc.accessRevisionCtrl = &fakeAccessRevisionController{currentErr: gerror.New("revision unavailable")}

	_, err := svc.BuildDynamicRouteAccessProjection(ctx, "freshness-token", 31, 71)
	if err == nil {
		t.Fatal("expected dynamic route projection to fail when freshness is unavailable")
	}
	if !strings.Contains(err.Error(), "revision unavailable") {
		t.Fatalf("expected freshness error to propagate, got %v", err)
	}
}

// TestBuildDynamicRouteAccessProjectionRejectsInvalidInputs verifies the
// projection contract refuses ambiguous token, user, and tenant identities.
func TestBuildDynamicRouteAccessProjectionRejectsInvalidInputs(t *testing.T) {
	ctx := context.Background()
	svc := newDefaultRoleTestService()
	resetRoleAccessCacheTestState(t, svc)
	svc.accessRevisionCtrl = &fakeAccessRevisionController{revision: 9}

	testCases := []struct {
		name     string
		tokenID  string
		userID   int
		tenantID int
	}{
		{name: "blank token", tokenID: " ", userID: 1, tenantID: 1},
		{name: "missing user", tokenID: "token", userID: 0, tenantID: 1},
		{name: "negative tenant", tokenID: "token", userID: 1, tenantID: -1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := svc.BuildDynamicRouteAccessProjection(
				ctx,
				testCase.tokenID,
				testCase.userID,
				testCase.tenantID,
			)
			if err == nil {
				t.Fatal("expected invalid projection input to fail")
			}
		})
	}
}
