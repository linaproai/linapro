// This file verifies sys_config capability write guards that do not require
// database fixtures.

package hostconfig

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
)

// TestSetValueRequiresCacheCoord verifies sys_config writes fail before data
// mutation when no unified cache coordinator is injected.
func TestSetValueRequiresCacheCoord(t *testing.T) {
	err := NewSysConfigCapabilityAdapter(nil, nil).SetValue(context.Background(), "site.title", "Lina")
	if !bizerr.Is(err, capmodel.CodeCapabilityUnavailable) {
		t.Fatalf("expected cachecoord unavailable error, got %v", err)
	}
}
