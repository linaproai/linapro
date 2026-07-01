// This file verifies in-memory authorization capability behaviors that do not
// require host database fixtures.

package role

import (
	"context"
	"reflect"
	"testing"

	"lina-core/pkg/bizerr"
	capabilityauthz "lina-core/pkg/plugin/capability/authcap/authz"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
)

// TestBatchHasPermissionsUsesBusinessContext verifies batch permission checks
// use the standard business context as a set rather than per-key storage calls.
func TestBatchHasPermissionsUsesBusinessContext(t *testing.T) {
	ctx := bizctxcap.WithCurrentContext(context.Background(), bizctxcap.CurrentContext{
		Permissions: []string{"system:user:list", "system:user:create"},
	})
	result, err := NewCapabilityAdapter(nil, nil, nil).BatchHasPermissions(ctx, []capabilityauthz.PermissionKey{"system:user:list", "system:user:delete", ""})
	if err != nil {
		t.Fatalf("batch has permissions failed: %v", err)
	}
	expected := map[capabilityauthz.PermissionKey]bool{
		"system:user:list":   true,
		"system:user:delete": false,
		"":                   false,
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("unexpected permission result: %#v", result)
	}
}

// TestBatchHasPermissionsRejectsLimit verifies the published batch limit is enforced.
func TestBatchHasPermissionsRejectsLimit(t *testing.T) {
	keys := make([]capabilityauthz.PermissionKey, capabilityauthz.MaxBatchHasPermissions+1)
	_, err := NewCapabilityAdapter(nil, nil, nil).BatchHasPermissions(context.Background(), keys)
	if !bizerr.Is(err, capmodel.CodeCapabilityLimitExceeded) {
		t.Fatalf("expected limit error, got %v", err)
	}
}

// TestReplaceRolePermissionsRequiresCacheCoord verifies write paths fail before
// touching role bindings when no shared cache coordinator is injected.
func TestReplaceRolePermissionsRequiresCacheCoord(t *testing.T) {
	ctx := bizctxcap.WithCurrentContext(context.Background(), bizctxcap.CurrentContext{
		UserID:   1,
		TenantID: 1,
	})
	err := NewCapabilityAdapter(nil, nil, nil).ReplaceRolePermissions(ctx, "1", nil)
	if !bizerr.Is(err, capmodel.CodeCapabilityUnavailable) {
		t.Fatalf("expected cachecoord unavailable error, got %v", err)
	}
}
