// This file tests scoped cache-key encoding and parsing behavior.

package kvcache

import (
	"context"
	"testing"
)

// TestBuildCacheKeyEncodesTrimmedParts verifies the public encoder trims and
// encodes each scoped cache-key segment deterministically.
func TestBuildCacheKeyEncodesTrimmedParts(t *testing.T) {
	key := BuildCacheKey(" module.scope ", " runtime/config ", " revision.v1 ")

	const expected = "bW9kdWxlLnNjb3Bl.cnVudGltZS9jb25maWc.cmV2aXNpb24udjE"
	if key != expected {
		t.Fatalf("expected encoded cache key %q, got %q", expected, key)
	}
}

// TestInvalidCacheKeyRejected verifies non-encoded cache keys are rejected
// before any backend read can reach the database.
func TestInvalidCacheKeyRejected(t *testing.T) {
	if _, _, err := New().Get(context.Background(), OwnerTypePlugin, "plain-cache-key"); err == nil {
		t.Fatal("expected invalid cache key format to fail")
	}
}
