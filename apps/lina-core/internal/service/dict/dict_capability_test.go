// This file verifies in-memory dictionary capability behaviors that do not
// require host database fixtures.

package dict

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilitydictcap "lina-core/pkg/plugin/capability/dictcap"
)

// TestEnsureValuesVisibleRejectsMissing verifies dictionary value visibility is
// fail-closed without exposing absent versus invisible values.
func TestEnsureValuesVisibleRejectsMissing(t *testing.T) {
	err := NewCapabilityAdapter(nil, nil, nil).Value().EnsureValuesVisible(context.Background(), capabilitydictcap.ResolveInput{
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
	err := NewCapabilityAdapter(nil, nil, nil).Value().EnsureValuesVisible(context.Background(), capabilitydictcap.ResolveInput{
		Type:   "sys_common_status",
		Values: values,
	})
	if !bizerr.Is(err, capmodel.CodeCapabilityLimitExceeded) {
		t.Fatalf("expected limit error, got %v", err)
	}
}

// TestRefreshRequiresCacheCoord verifies dictionary refresh cannot bypass the
// unified cache coordinator.
func TestRefreshRequiresCacheCoord(t *testing.T) {
	err := NewCapabilityAdapter(nil, nil, nil).Refresh(context.Background(), "sys_common_status")
	if !bizerr.Is(err, capmodel.CodeCapabilityUnavailable) {
		t.Fatalf("expected cachecoord unavailable error, got %v", err)
	}
}
