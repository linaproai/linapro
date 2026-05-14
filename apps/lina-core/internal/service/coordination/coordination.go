// Package coordination provides cluster coordination stores for locks,
// short-lived key-value state, cache revisions, cross-node events, and health.
package coordination

import (
	"context"

	"lina-core/internal/service/coordination/internal/core"
	redisprovider "lina-core/internal/service/coordination/internal/redis"
)

// BackendName identifies one coordination backend implementation.
type BackendName = core.BackendName

// Supported coordination backend names.
const (
	// BackendRedis stores cluster coordination state in Redis.
	BackendRedis = core.BackendRedis
	// BackendMemory stores coordination state in process memory for tests.
	BackendMemory = core.BackendMemory
)

// Service exposes deployment-wide coordination stores through narrow contracts.
type Service = core.Service

// Provider creates a coordination service for one backend implementation.
type Provider = core.Provider

// LockStore coordinates distributed lock ownership.
type LockStore = core.LockStore

// KVStore stores short-lived coordination and lossy cache values.
type KVStore = core.KVStore

// RevisionStore coordinates monotonic cache-domain revisions.
type RevisionStore = core.RevisionStore

// EventBus publishes cross-node coordination events.
type EventBus = core.EventBus

// EventHandler handles one coordination event.
type EventHandler = core.EventHandler

// Subscription represents an active event subscription.
type Subscription = core.Subscription

// HealthChecker exposes backend health information.
type HealthChecker = core.HealthChecker

// LockHandle identifies one acquired distributed lock.
type LockHandle = core.LockHandle

// RevisionKey identifies one cache-domain revision.
type RevisionKey = core.RevisionKey

// Event represents one cross-node coordination notification.
type Event = core.Event

// HealthSnapshot describes the current coordination backend state.
type HealthSnapshot = core.HealthSnapshot

// KeyBuilder builds backend keys with stable application and environment
// prefixes so tests and deployments can isolate coordination data.
type KeyBuilder = core.KeyBuilder

// RedisOptions contains normalized Redis connection settings for coordination.
type RedisOptions = redisprovider.Options

// NewKeyBuilder creates a key builder using normalized namespace parts.
func NewKeyBuilder(application string, environment string, instance string) *KeyBuilder {
	return core.NewKeyBuilder(application, environment, instance)
}

// DefaultKeyBuilder creates the default LinaPro key namespace builder.
func DefaultKeyBuilder() *KeyBuilder {
	return core.DefaultKeyBuilder()
}

// NewMemory creates a process-local coordination service for tests.
func NewMemory(keys *KeyBuilder) Service {
	return core.NewMemory(keys)
}

// NewRedis creates a Redis coordination service and verifies connectivity.
func NewRedis(ctx context.Context, options RedisOptions) (Service, error) {
	return redisprovider.New(ctx, options)
}
