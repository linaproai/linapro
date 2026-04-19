// This file tests scoped cache-key encoding and parsing behavior.

package kvcache

import "testing"

// TestBuildCacheKeyRoundTrip verifies cache keys round-trip through the public
// encoder and internal decoder.
func TestBuildCacheKeyRoundTrip(t *testing.T) {
	key := BuildCacheKey(" module.scope ", " runtime/config ", " revision.v1 ")

	identity, err := parseCacheKey(key)
	if err != nil {
		t.Fatalf("expected cache key to parse, got %v", err)
	}
	if identity.ownerKey != "module.scope" {
		t.Fatalf("expected owner key to round-trip, got %q", identity.ownerKey)
	}
	if identity.namespace != "runtime/config" {
		t.Fatalf("expected namespace to round-trip, got %q", identity.namespace)
	}
	if identity.cacheKey != "revision.v1" {
		t.Fatalf("expected logical cache key to round-trip, got %q", identity.cacheKey)
	}
}

// TestParseCacheKeyRejectsInvalidFormat verifies non-encoded cache keys are
// rejected with a validation error.
func TestParseCacheKeyRejectsInvalidFormat(t *testing.T) {
	if _, err := parseCacheKey("plain-cache-key"); err == nil {
		t.Fatal("expected invalid cache key format to fail")
	}
}
