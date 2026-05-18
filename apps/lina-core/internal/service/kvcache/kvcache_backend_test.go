// This file tests kvcache backend/provider wiring.

package kvcache

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gtime"
)

// fakeBackend captures delegated service calls without touching the database.
type fakeBackend struct {
	lastTTL time.Duration
}

// Name returns the fake backend name.
func (f *fakeBackend) Name() BackendName { return "fake" }

// RequiresExpiredCleanup reports that the fake backend does not need cleanup.
func (f *fakeBackend) RequiresExpiredCleanup() bool { return false }

// Get is unused by the backend wiring test.
func (f *fakeBackend) Get(ctx context.Context, ownerType OwnerType, cacheKey string) (*Item, bool, error) {
	return nil, false, nil
}

// GetInt is unused by the backend wiring test.
func (f *fakeBackend) GetInt(ctx context.Context, ownerType OwnerType, cacheKey string) (int64, bool, error) {
	return 0, false, nil
}

// Set records the backend-neutral TTL delegated by the service.
func (f *fakeBackend) Set(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
	value string,
	ttl time.Duration,
) (*Item, error) {
	f.lastTTL = ttl
	return &Item{Key: cacheKey, ValueKind: ValueKindString, Value: value}, nil
}

// Delete is unused by the backend wiring test.
func (f *fakeBackend) Delete(ctx context.Context, ownerType OwnerType, cacheKey string) error {
	return nil
}

// Incr is unused by the backend wiring test.
func (f *fakeBackend) Incr(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
	delta int64,
	ttl time.Duration,
) (*Item, error) {
	return nil, nil
}

// Expire is unused by the backend wiring test.
func (f *fakeBackend) Expire(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
	ttl time.Duration,
) (bool, *gtime.Time, error) {
	return false, nil, nil
}

// CleanupExpired is unused by the backend wiring test.
func (f *fakeBackend) CleanupExpired(ctx context.Context) error {
	return nil
}

// TestServiceDelegatesDurationTTLToBackend verifies service construction can
// use an injected backend and keeps TTL values backend-neutral.
func TestServiceDelegatesDurationTTLToBackend(t *testing.T) {
	backend := &fakeBackend{}
	service := New(WithBackend(backend))

	if service.BackendName() != "fake" {
		t.Fatalf("expected fake backend, got %q", service.BackendName())
	}
	if service.RequiresExpiredCleanup() {
		t.Fatal("expected fake backend cleanup flag to be false")
	}
	if _, err := service.Set(context.Background(), OwnerTypePlugin, "encoded", "value", 2*time.Minute); err != nil {
		t.Fatalf("expected delegated set to succeed, got error: %v", err)
	}
	if backend.lastTTL != 2*time.Minute {
		t.Fatalf("expected TTL 2m, got %s", backend.lastTTL)
	}
}
