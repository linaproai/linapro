// Package core defines backend-neutral coordination contracts and in-process
// implementations shared by concrete coordination providers.
package core

import (
	"context"
	"time"
)

// BackendName identifies one coordination backend implementation.
type BackendName string

// Supported coordination backend names.
const (
	// BackendRedis stores cluster coordination state in Redis.
	BackendRedis BackendName = "redis"
	// BackendMemory stores coordination state in process memory for tests.
	BackendMemory BackendName = "memory"
)

// Service exposes deployment-wide coordination stores through narrow contracts.
type Service interface {
	// BackendName returns the active coordination backend name.
	BackendName() BackendName
	// KeyBuilder returns the backend key builder used by this service.
	KeyBuilder() *KeyBuilder
	// Lock returns the distributed lock store.
	Lock() LockStore
	// KV returns the short-lived key-value store.
	KV() KVStore
	// Revision returns the cache revision store.
	Revision() RevisionStore
	// Events returns the cross-node event bus.
	Events() EventBus
	// Health returns the backend health checker.
	Health() HealthChecker
	// Close releases backend resources such as Redis subscriptions and clients.
	Close(ctx context.Context) error
}

// Provider creates a coordination service for one backend implementation.
type Provider interface {
	// BackendName returns the backend name produced by this provider.
	BackendName() BackendName
	// New creates and verifies one coordination service instance.
	New(ctx context.Context) (Service, error)
}

// LockStore coordinates distributed lock ownership.
type LockStore interface {
	// Acquire obtains a lock when it is absent or expired.
	Acquire(ctx context.Context, name string, owner string, reason string, ttl time.Duration) (*LockHandle, bool, error)
	// Renew extends a lock only when the caller still owns it.
	Renew(ctx context.Context, handle *LockHandle, ttl time.Duration) error
	// Release releases a lock only when the caller still owns it.
	Release(ctx context.Context, handle *LockHandle) error
	// IsHeld reports whether the handle still owns the lock.
	IsHeld(ctx context.Context, handle *LockHandle) (bool, error)
}

// KVStore stores short-lived coordination and lossy cache values.
type KVStore interface {
	// Get returns a string value for key when it exists.
	Get(ctx context.Context, key string) (string, bool, error)
	// Set stores a string value with optional TTL. ttl=0 means no expiration.
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	// SetNX stores a string value only when the key is absent.
	SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error)
	// Delete removes a key and treats missing keys as success.
	Delete(ctx context.Context, key string) error
	// CompareAndDelete removes a key only when its current value matches expected.
	CompareAndDelete(ctx context.Context, key string, expected string) (bool, error)
	// IncrBy increments an integer key by delta and returns the new value.
	IncrBy(ctx context.Context, key string, delta int64, ttl time.Duration) (int64, error)
	// Expire updates a key expiration. ttl=0 clears expiration.
	Expire(ctx context.Context, key string, ttl time.Duration) (bool, error)
	// TTL returns the remaining key lifetime. A negative duration means no TTL or missing key.
	TTL(ctx context.Context, key string) (time.Duration, error)
}

// RevisionStore coordinates monotonic cache-domain revisions.
type RevisionStore interface {
	// Bump increments and returns one revision.
	Bump(ctx context.Context, key RevisionKey, reason string) (int64, error)
	// Current returns the latest revision, initializing missing revisions as 0.
	Current(ctx context.Context, key RevisionKey) (int64, error)
}

// EventBus publishes cross-node coordination events.
type EventBus interface {
	// Publish publishes one event to peer nodes.
	Publish(ctx context.Context, event Event) error
	// Subscribe starts consuming events until the subscription is closed.
	Subscribe(ctx context.Context, handler EventHandler) (Subscription, error)
}

// EventHandler handles one coordination event.
type EventHandler func(ctx context.Context, event Event) error

// Subscription represents an active event subscription.
type Subscription interface {
	// Close stops the subscription and releases resources.
	Close(ctx context.Context) error
}

// HealthChecker exposes backend health information.
type HealthChecker interface {
	// Ping verifies backend connectivity.
	Ping(ctx context.Context) error
	// Snapshot returns the latest health state.
	Snapshot(ctx context.Context) HealthSnapshot
}

// LockHandle identifies one acquired distributed lock.
type LockHandle struct {
	Name         string        // Name is the logical lock name.
	Owner        string        // Owner identifies the requesting node or component.
	Token        string        // Token is the unique owner token stored in the backend.
	Reason       string        // Reason describes why the lock was acquired.
	Lease        time.Duration // Lease is the requested lock lifetime.
	FencingToken int64         // FencingToken monotonically increases per lock when supported.
	AcquiredAt   time.Time     // AcquiredAt records local acquisition time.
}

// RevisionKey identifies one cache-domain revision.
type RevisionKey struct {
	TenantID         int64  // TenantID is the target tenant, 0 platform, or -1 all tenants.
	Domain           string // Domain is the cache domain identifier.
	Scope            string // Scope is the explicit invalidation scope.
	CascadeToTenants bool   // CascadeToTenants expands platform changes to tenant views.
}

// Event represents one cross-node coordination notification.
type Event struct {
	ID               string    `json:"id"`
	Kind             string    `json:"kind"`
	Domain           string    `json:"domain"`
	Scope            string    `json:"scope"`
	TenantID         int64     `json:"tenantId"`
	CascadeToTenants bool      `json:"cascadeToTenants"`
	Revision         int64     `json:"revision"`
	Reason           string    `json:"reason"`
	SourceNode       string    `json:"sourceNode"`
	CreatedAt        time.Time `json:"createdAt"`
}

// HealthSnapshot describes the current coordination backend state.
type HealthSnapshot struct {
	Backend             BackendName // Backend is the active coordination implementation.
	Healthy             bool        // Healthy reports whether the latest ping succeeded.
	LastSuccessAt       time.Time   // LastSuccessAt records the latest successful health check.
	LastError           string      // LastError stores the latest health or subscription error.
	SubscriberRunning   bool        // SubscriberRunning reports whether event subscription is active.
	LastEventReceivedAt time.Time   // LastEventReceivedAt records the latest consumed event time.
}

// serviceImpl composes one backend implementation into the Service interface.
type serviceImpl struct {
	backend  BackendName
	keys     *KeyBuilder
	lock     LockStore
	kv       KVStore
	revision RevisionStore
	events   EventBus
	health   HealthChecker
	close    func(context.Context) error
}

// NewService composes backend-specific store implementations into a
// backend-neutral coordination service.
func NewService(
	backend BackendName,
	keys *KeyBuilder,
	lock LockStore,
	kv KVStore,
	revision RevisionStore,
	events EventBus,
	health HealthChecker,
	close func(context.Context) error,
) Service {
	if keys == nil {
		keys = DefaultKeyBuilder()
	}
	return &serviceImpl{
		backend:  backend,
		keys:     keys,
		lock:     lock,
		kv:       kv,
		revision: revision,
		events:   events,
		health:   health,
		close:    close,
	}
}

// BackendName returns the active coordination backend name.
func (s *serviceImpl) BackendName() BackendName {
	if s == nil {
		return ""
	}
	return s.backend
}

// KeyBuilder returns the backend key builder used by this service.
func (s *serviceImpl) KeyBuilder() *KeyBuilder {
	if s == nil || s.keys == nil {
		return DefaultKeyBuilder()
	}
	return s.keys
}

// Lock returns the distributed lock store.
func (s *serviceImpl) Lock() LockStore {
	if s == nil {
		return nil
	}
	return s.lock
}

// KV returns the short-lived key-value store.
func (s *serviceImpl) KV() KVStore {
	if s == nil {
		return nil
	}
	return s.kv
}

// Revision returns the cache revision store.
func (s *serviceImpl) Revision() RevisionStore {
	if s == nil {
		return nil
	}
	return s.revision
}

// Events returns the cross-node event bus.
func (s *serviceImpl) Events() EventBus {
	if s == nil {
		return nil
	}
	return s.events
}

// Health returns the backend health checker.
func (s *serviceImpl) Health() HealthChecker {
	if s == nil {
		return nil
	}
	return s.health
}

// Close releases backend resources.
func (s *serviceImpl) Close(ctx context.Context) error {
	if s == nil || s.close == nil {
		return nil
	}
	return s.close(ctx)
}
