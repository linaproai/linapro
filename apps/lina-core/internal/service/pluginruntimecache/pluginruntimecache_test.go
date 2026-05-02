// This file verifies plugin runtime cache revision coordination behavior.

package pluginruntimecache

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/service/kvcache"
)

// fakeKVCacheService provides deterministic shared-KV behavior for revision tests.
type fakeKVCacheService struct {
	getIntValue int64
	getIntFound bool
	getIntErr   error
	getIntCalls int32
	getIntKey   string
	incrValue   int64
	incrErr     error
	incrCalls   int32
	incrKey     string
}

// Get returns no string item because revision tests only exercise integer values.
func (f *fakeKVCacheService) Get(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
) (*kvcache.Item, bool, error) {
	return nil, false, nil
}

// GetInt returns the configured shared revision.
func (f *fakeKVCacheService) GetInt(
	_ context.Context,
	_ kvcache.OwnerType,
	cacheKey string,
) (int64, bool, error) {
	atomic.AddInt32(&f.getIntCalls, 1)
	f.getIntKey = cacheKey
	if f.getIntErr != nil {
		return 0, false, f.getIntErr
	}
	return f.getIntValue, f.getIntFound, nil
}

// Set is unused by revision tests.
func (f *fakeKVCacheService) Set(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
	_ string,
	_ int64,
) (*kvcache.Item, error) {
	return nil, nil
}

// Delete is unused by revision tests.
func (f *fakeKVCacheService) Delete(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
) error {
	return nil
}

// Incr returns the configured incremented revision.
func (f *fakeKVCacheService) Incr(
	_ context.Context,
	_ kvcache.OwnerType,
	cacheKey string,
	_ int64,
	_ int64,
) (*kvcache.Item, error) {
	atomic.AddInt32(&f.incrCalls, 1)
	f.incrKey = cacheKey
	if f.incrErr != nil {
		return nil, f.incrErr
	}
	return &kvcache.Item{IntValue: f.incrValue}, nil
}

// Expire is unused by revision tests.
func (f *fakeKVCacheService) Expire(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
	_ int64,
) (bool, *gtime.Time, error) {
	return false, nil, nil
}

// CleanupExpired is unused by revision tests.
func (f *fakeKVCacheService) CleanupExpired(_ context.Context) error {
	return nil
}

// TestControllerSingleNodeNoops verifies single-node deployments avoid shared
// KV reads, writes, and refresh callbacks.
func TestControllerSingleNodeNoops(t *testing.T) {
	fakeKV := &fakeKVCacheService{getIntFound: true, getIntValue: 3, incrValue: 4}
	var refreshCalls int32
	controller := NewController(false, fakeKV, NewObservedRevision(), func(_ context.Context) error {
		atomic.AddInt32(&refreshCalls, 1)
		return nil
	})

	if err := controller.EnsureFresh(context.Background()); err != nil {
		t.Fatalf("single-node ensure fresh failed: %v", err)
	}
	revision, err := controller.MarkChanged(context.Background())
	if err != nil {
		t.Fatalf("single-node mark changed failed: %v", err)
	}
	if revision != 0 {
		t.Fatalf("expected single-node revision 0, got %d", revision)
	}
	if atomic.LoadInt32(&fakeKV.getIntCalls) != 0 || atomic.LoadInt32(&fakeKV.incrCalls) != 0 {
		t.Fatalf("expected no shared KV traffic, got get=%d incr=%d", fakeKV.getIntCalls, fakeKV.incrCalls)
	}
	if atomic.LoadInt32(&refreshCalls) != 0 {
		t.Fatalf("expected no refresh callback, got %d", refreshCalls)
	}
}

// TestControllerEnsureFreshRefreshesOnRevisionChange verifies each cache domain
// refreshes once per newly observed shared revision.
func TestControllerEnsureFreshRefreshesOnRevisionChange(t *testing.T) {
	fakeKV := &fakeKVCacheService{getIntFound: true, getIntValue: 5}
	var refreshCalls int32
	controller := NewController(true, fakeKV, NewObservedRevision(), func(_ context.Context) error {
		atomic.AddInt32(&refreshCalls, 1)
		return nil
	})

	if err := controller.EnsureFresh(context.Background()); err != nil {
		t.Fatalf("first ensure fresh failed: %v", err)
	}
	if err := controller.EnsureFresh(context.Background()); err != nil {
		t.Fatalf("second ensure fresh failed: %v", err)
	}
	if atomic.LoadInt32(&refreshCalls) != 1 {
		t.Fatalf("expected one refresh for revision 5, got %d", refreshCalls)
	}

	fakeKV.getIntValue = 6
	if err := controller.EnsureFresh(context.Background()); err != nil {
		t.Fatalf("third ensure fresh failed: %v", err)
	}
	if atomic.LoadInt32(&refreshCalls) != 2 {
		t.Fatalf("expected second refresh for revision 6, got %d", refreshCalls)
	}
}

// TestControllerMarkChangedStoresReturnedRevision verifies the mutating node
// records the revision it published so its next read path does not refresh again.
func TestControllerMarkChangedStoresReturnedRevision(t *testing.T) {
	fakeKV := &fakeKVCacheService{getIntFound: true, getIntValue: 9, incrValue: 9}
	var refreshCalls int32
	controller := NewController(true, fakeKV, NewObservedRevision(), func(_ context.Context) error {
		atomic.AddInt32(&refreshCalls, 1)
		return nil
	})

	revision, err := controller.MarkChanged(context.Background())
	if err != nil {
		t.Fatalf("mark changed failed: %v", err)
	}
	if revision != 9 {
		t.Fatalf("expected revision 9, got %d", revision)
	}
	if err = controller.EnsureFresh(context.Background()); err != nil {
		t.Fatalf("ensure after mark failed: %v", err)
	}
	if atomic.LoadInt32(&refreshCalls) != 0 {
		t.Fatalf("expected no refresh after local mark, got %d", refreshCalls)
	}
}

// TestControllerPublishChangedLeavesRevisionUnobserved verifies callers can
// publish a revision that the same local process should still consume later.
func TestControllerPublishChangedLeavesRevisionUnobserved(t *testing.T) {
	fakeKV := &fakeKVCacheService{getIntFound: true, getIntValue: 10, incrValue: 10}
	var refreshCalls int32
	controller := NewController(true, fakeKV, NewObservedRevision(), func(_ context.Context) error {
		atomic.AddInt32(&refreshCalls, 1)
		return nil
	})

	revision, err := controller.PublishChanged(context.Background())
	if err != nil {
		t.Fatalf("publish changed failed: %v", err)
	}
	if revision != 10 {
		t.Fatalf("expected revision 10, got %d", revision)
	}
	if err = controller.EnsureFresh(context.Background()); err != nil {
		t.Fatalf("ensure after publish failed: %v", err)
	}
	if atomic.LoadInt32(&refreshCalls) != 1 {
		t.Fatalf("expected refresh after unobserved publish, got %d", refreshCalls)
	}
}

// TestControllerPropagatesSharedKVErrors verifies shared KV failures are
// returned to callers that can fail closed.
func TestControllerPropagatesSharedKVErrors(t *testing.T) {
	readErr := errors.New("read revision failed")
	readController := NewController(
		true,
		&fakeKVCacheService{getIntErr: readErr},
		NewObservedRevision(),
		nil,
	)
	if err := readController.EnsureFresh(context.Background()); !errors.Is(err, readErr) {
		t.Fatalf("expected read error, got %v", err)
	}

	writeErr := errors.New("write revision failed")
	writeController := NewController(
		true,
		&fakeKVCacheService{incrErr: writeErr},
		NewObservedRevision(),
		nil,
	)
	if _, err := writeController.MarkChanged(context.Background()); !errors.Is(err, writeErr) {
		t.Fatalf("expected write error, got %v", err)
	}
}

// TestControllerForKeyUsesExplicitCacheKey verifies non-default coordination
// domains can store revisions under an independent shared KV key.
func TestControllerForKeyUsesExplicitCacheKey(t *testing.T) {
	fakeKV := &fakeKVCacheService{getIntFound: true, getIntValue: 11, incrValue: 12}
	controller := NewControllerForKey(
		ReconcilerRevisionCacheKey,
		true,
		fakeKV,
		NewObservedRevision(),
		nil,
	)

	revision, err := controller.CurrentRevision(context.Background())
	if err != nil {
		t.Fatalf("current revision failed: %v", err)
	}
	if revision != 11 {
		t.Fatalf("expected revision 11, got %d", revision)
	}
	if fakeKV.getIntKey != ReconcilerRevisionCacheKey {
		t.Fatalf("expected get key %q, got %q", ReconcilerRevisionCacheKey, fakeKV.getIntKey)
	}

	revision, err = controller.MarkChanged(context.Background())
	if err != nil {
		t.Fatalf("mark changed failed: %v", err)
	}
	if revision != 12 {
		t.Fatalf("expected incremented revision 12, got %d", revision)
	}
	if fakeKV.incrKey != ReconcilerRevisionCacheKey {
		t.Fatalf("expected incr key %q, got %q", ReconcilerRevisionCacheKey, fakeKV.incrKey)
	}
}
