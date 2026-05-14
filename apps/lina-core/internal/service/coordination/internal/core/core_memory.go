// This file implements an in-memory coordination backend for deterministic
// unit tests and local component composition that must not connect to Redis.

package core

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
)

// memoryBackend implements all coordination stores against process memory.
type memoryBackend struct {
	keys        *KeyBuilder
	mu          sync.Mutex
	locks       map[string]*memoryLockRecord
	values      map[string]*memoryKVRecord
	revisions   map[string]int64
	subscribers map[int]EventHandler
	nextSubID   int
	closed      bool
}

// memoryLockRecord stores one in-memory lock owner and expiration.
type memoryLockRecord struct {
	handle   LockHandle
	expireAt time.Time
}

// memoryKVRecord stores one in-memory KV value and optional expiration.
type memoryKVRecord struct {
	value    string
	expireAt time.Time
}

// memorySubscription represents one active in-memory event subscriber.
type memorySubscription struct {
	backend *memoryBackend
	id      int
	once    sync.Once
}

// NewMemory creates a process-local coordination service for tests.
func NewMemory(keys *KeyBuilder) Service {
	if keys == nil {
		keys = DefaultKeyBuilder()
	}
	backend := &memoryBackend{
		keys:        keys,
		locks:       make(map[string]*memoryLockRecord),
		values:      make(map[string]*memoryKVRecord),
		revisions:   make(map[string]int64),
		subscribers: make(map[int]EventHandler),
	}
	return NewService(BackendMemory, keys, backend, backend, backend, backend, backend, backend.Close)
}

// Acquire obtains a lock when it is absent or expired.
func (m *memoryBackend) Acquire(
	ctx context.Context,
	name string,
	owner string,
	reason string,
	ttl time.Duration,
) (*LockHandle, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	if ttl <= 0 {
		return nil, false, BizerrTTLInvalid()
	}
	lockKey, err := m.keys.LockKey(name)
	if err != nil {
		return nil, false, err
	}
	fenceKey, err := m.keys.LockFenceKey(name)
	if err != nil {
		return nil, false, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, false, gerror.New("coordination backend is closed")
	}

	now := time.Now()
	if existing := m.locks[lockKey]; existing != nil && existing.expireAt.After(now) {
		if existing.handle.Owner != owner {
			return nil, false, nil
		}
		existing.expireAt = now.Add(ttl)
		existing.handle.Lease = ttl
		existing.handle.Reason = reason
		handle := existing.handle
		return &handle, true, nil
	}

	m.revisions[fenceKey]++
	token, err := NewOwnerToken(owner)
	if err != nil {
		return nil, false, err
	}
	handle := LockHandle{
		Name:         name,
		Owner:        owner,
		Token:        token,
		Reason:       reason,
		Lease:        ttl,
		FencingToken: m.revisions[fenceKey],
		AcquiredAt:   now,
	}
	m.locks[lockKey] = &memoryLockRecord{
		handle:   handle,
		expireAt: now.Add(ttl),
	}
	return &handle, true, nil
}

// Renew extends a lock only when the caller still owns it.
func (m *memoryBackend) Renew(ctx context.Context, handle *LockHandle, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if ttl <= 0 {
		return BizerrTTLInvalid()
	}
	if handle == nil {
		return bizerr.NewCode(CodeCoordinationLockNotHeld)
	}
	lockKey, err := m.keys.LockKey(HandleName(handle))
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	record := m.locks[lockKey]
	if record == nil || record.handle.Token != handle.Token || !record.expireAt.After(time.Now()) {
		return bizerr.NewCode(CodeCoordinationLockNotHeld)
	}
	record.expireAt = time.Now().Add(ttl)
	record.handle.Lease = ttl
	return nil
}

// Release releases a lock only when the caller still owns it.
func (m *memoryBackend) Release(ctx context.Context, handle *LockHandle) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if handle == nil {
		return bizerr.NewCode(CodeCoordinationLockNotHeld)
	}
	lockKey, err := m.keys.LockKey(HandleName(handle))
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	record := m.locks[lockKey]
	if record == nil {
		return nil
	}
	if record.handle.Token != handle.Token {
		return bizerr.NewCode(CodeCoordinationLockNotHeld)
	}
	delete(m.locks, lockKey)
	return nil
}

// IsHeld reports whether the handle still owns the lock.
func (m *memoryBackend) IsHeld(ctx context.Context, handle *LockHandle) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if handle == nil {
		return false, nil
	}
	lockKey, err := m.keys.LockKey(HandleName(handle))
	if err != nil {
		return false, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	record := m.locks[lockKey]
	return record != nil && record.handle.Token == handle.Token && record.expireAt.After(time.Now()), nil
}

// Get returns a string value for key when it exists.
func (m *memoryBackend) Get(ctx context.Context, key string) (string, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	normalizedKey, err := RequireLogicalKey(key)
	if err != nil {
		return "", false, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	record := m.values[normalizedKey]
	if record == nil || RecordExpired(record.expireAt) {
		delete(m.values, normalizedKey)
		return "", false, nil
	}
	return record.value, true, nil
}

// Set stores a string value with optional TTL. ttl=0 means no expiration.
func (m *memoryBackend) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	normalizedKey, expireAt, err := NormalizeKVWrite(key, ttl)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.values[normalizedKey] = &memoryKVRecord{value: value, expireAt: expireAt}
	m.mu.Unlock()
	return nil
}

// SetNX stores a string value only when the key is absent.
func (m *memoryBackend) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	normalizedKey, expireAt, err := NormalizeKVWrite(key, ttl)
	if err != nil {
		return false, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	record := m.values[normalizedKey]
	if record != nil && !RecordExpired(record.expireAt) {
		return false, nil
	}
	m.values[normalizedKey] = &memoryKVRecord{value: value, expireAt: expireAt}
	return true, nil
}

// Delete removes a key and treats missing keys as success.
func (m *memoryBackend) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	normalizedKey, err := RequireLogicalKey(key)
	if err != nil {
		return err
	}
	m.mu.Lock()
	delete(m.values, normalizedKey)
	m.mu.Unlock()
	return nil
}

// CompareAndDelete removes a key only when its current value matches expected.
func (m *memoryBackend) CompareAndDelete(ctx context.Context, key string, expected string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	normalizedKey, err := RequireLogicalKey(key)
	if err != nil {
		return false, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	record := m.values[normalizedKey]
	if record == nil || RecordExpired(record.expireAt) {
		delete(m.values, normalizedKey)
		return false, nil
	}
	if record.value != expected {
		return false, nil
	}
	delete(m.values, normalizedKey)
	return true, nil
}

// IncrBy increments an integer key by delta and returns the new value.
func (m *memoryBackend) IncrBy(ctx context.Context, key string, delta int64, ttl time.Duration) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	normalizedKey, expireAt, err := NormalizeKVWrite(key, ttl)
	if err != nil {
		return 0, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	current := int64(0)
	record := m.values[normalizedKey]
	if record != nil && !RecordExpired(record.expireAt) {
		parsed, parseErr := ParseStoredInt(record.value)
		if parseErr != nil {
			return 0, parseErr
		}
		current = parsed
	}
	current += delta
	if ttl == 0 && record != nil && !RecordExpired(record.expireAt) {
		expireAt = record.expireAt
	}
	m.values[normalizedKey] = &memoryKVRecord{
		value:    FormatStoredInt(current),
		expireAt: expireAt,
	}
	return current, nil
}

// Expire updates a key expiration. ttl=0 clears expiration.
func (m *memoryBackend) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	normalizedKey, expireAt, err := NormalizeKVWrite(key, ttl)
	if err != nil {
		return false, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	record := m.values[normalizedKey]
	if record == nil || RecordExpired(record.expireAt) {
		delete(m.values, normalizedKey)
		return false, nil
	}
	record.expireAt = expireAt
	return true, nil
}

// TTL returns the remaining key lifetime. A negative duration means no TTL or missing key.
func (m *memoryBackend) TTL(ctx context.Context, key string) (time.Duration, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	normalizedKey, err := RequireLogicalKey(key)
	if err != nil {
		return 0, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	record := m.values[normalizedKey]
	if record == nil || RecordExpired(record.expireAt) {
		delete(m.values, normalizedKey)
		return -1, nil
	}
	if record.expireAt.IsZero() {
		return -1, nil
	}
	return time.Until(record.expireAt), nil
}

// Bump increments and returns one revision.
func (m *memoryBackend) Bump(ctx context.Context, key RevisionKey, reason string) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	backendKey, err := m.keys.RevisionKey(key)
	if err != nil {
		return 0, err
	}

	m.mu.Lock()
	m.revisions[backendKey]++
	revision := m.revisions[backendKey]
	m.mu.Unlock()
	return revision, nil
}

// Current returns the latest revision, initializing missing revisions as 0.
func (m *memoryBackend) Current(ctx context.Context, key RevisionKey) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	backendKey, err := m.keys.RevisionKey(key)
	if err != nil {
		return 0, err
	}

	m.mu.Lock()
	revision := m.revisions[backendKey]
	m.mu.Unlock()
	return revision, nil
}

// Publish publishes one event to peer nodes.
func (m *memoryBackend) Publish(ctx context.Context, event Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	handlers := make([]EventHandler, 0, len(m.subscribers))
	for _, handler := range m.subscribers {
		handlers = append(handlers, handler)
	}
	m.mu.Unlock()
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe starts consuming events until the subscription is closed.
func (m *memoryBackend) Subscribe(ctx context.Context, handler EventHandler) (Subscription, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if handler == nil {
		return nil, gerror.New("coordination event handler is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil, gerror.New("coordination backend is closed")
	}
	m.nextSubID++
	id := m.nextSubID
	m.subscribers[id] = handler
	return &memorySubscription{backend: m, id: id}, nil
}

// Ping verifies backend connectivity.
func (m *memoryBackend) Ping(ctx context.Context) error {
	return ctx.Err()
}

// Snapshot returns the latest health state.
func (m *memoryBackend) Snapshot(ctx context.Context) HealthSnapshot {
	healthy := ctx.Err() == nil
	return HealthSnapshot{
		Backend:           BackendMemory,
		Healthy:           healthy,
		LastSuccessAt:     time.Now(),
		SubscriberRunning: true,
	}
}

// Close releases in-memory subscribers.
func (m *memoryBackend) Close(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	m.closed = true
	clear(m.subscribers)
	m.mu.Unlock()
	return nil
}

// Close stops the subscription and releases resources.
func (s *memorySubscription) Close(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.once.Do(func() {
		if s.backend == nil {
			return
		}
		s.backend.mu.Lock()
		delete(s.backend.subscribers, s.id)
		s.backend.mu.Unlock()
	})
	return nil
}
