// This file verifies small in-memory authorization and dictionary capability
// behaviors that do not require host database fixtures.

package capabilityhost

import (
	"context"
	"reflect"
	"testing"

	"lina-core/pkg/bizerr"
	capabilityauthz "lina-core/pkg/plugin/capability/authcap/authz"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilitydictcap "lina-core/pkg/plugin/capability/dictcap"
)

// TestBatchHasPermissionsUsesSnapshot verifies batch permission checks use the
// capability authorization snapshot as a set rather than per-key storage calls.
func TestBatchHasPermissionsUsesSnapshot(t *testing.T) {
	ctx := context.Background()
	result, err := newAuthzCapabilityAdapter().BatchHasPermissions(ctx, capmodel.CapabilityContext{
		Authorization: capmodel.CapabilityAuthorizationSnapshot{
			Permissions: []string{"system:user:list", "system:user:create"},
		},
	}, []capabilityauthz.PermissionKey{"system:user:list", "system:user:delete", ""})
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
	_, err := newAuthzCapabilityAdapter().BatchHasPermissions(context.Background(), capmodel.CapabilityContext{}, keys)
	if !bizerr.Is(err, capmodel.CodeCapabilityLimitExceeded) {
		t.Fatalf("expected limit error, got %v", err)
	}
}

// TestEnsureValuesVisibleRejectsMissing verifies dictionary value visibility is
// fail-closed without exposing absent versus invisible values.
func TestEnsureValuesVisibleRejectsMissing(t *testing.T) {
	err := newDictCapabilityAdapter(nil, nil).EnsureValuesVisible(context.Background(), capmodel.CapabilityContext{}, capabilitydictcap.ResolveInput{
		Type:   "sys_common_status",
		Values: []capabilitydictcap.Value{"missing"},
	})
	if !bizerr.Is(err, capmodel.CodeCapabilityDenied) {
		t.Fatalf("expected denied error for missing dictionary value, got %v", err)
	}
}

// TestEnsureValuesVisibleRejectsLimit verifies dictionary value checks are bounded.
func TestEnsureValuesVisibleRejectsLimit(t *testing.T) {
	values := make([]capabilitydictcap.Value, capabilitydictcap.MaxEnsureValuesVisible+1)
	err := newDictCapabilityAdapter(nil, nil).EnsureValuesVisible(context.Background(), capmodel.CapabilityContext{}, capabilitydictcap.ResolveInput{
		Type:   "sys_common_status",
		Values: values,
	})
	if !bizerr.Is(err, capmodel.CodeCapabilityLimitExceeded) {
		t.Fatalf("expected limit error, got %v", err)
	}
}
