// This file defines backend and provider abstractions for the kvcache service.

package kvcache

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/os/gtime"

	mysqlmemory "lina-core/internal/service/kvcache/internal/mysql-memory"
)

// BackendName identifies one kvcache backend implementation.
type BackendName string

// Supported backend names for the kvcache service.
const (
	// BackendMySQLMemory stores cache entries in the host MySQL MEMORY table.
	BackendMySQLMemory BackendName = "mysql-memory"
)

// Backend defines the backend-specific KV cache operations used by Service.
type Backend interface {
	// Name returns the stable backend implementation name.
	Name() BackendName
	// RequiresExpiredCleanup reports whether the backend needs external cleanup
	// for expired entries.
	RequiresExpiredCleanup() bool
	// Get returns the current cache entry snapshot identified by ownerType and cacheKey.
	Get(ctx context.Context, ownerType OwnerType, cacheKey string) (*Item, bool, error)
	// GetInt returns the current integer cache value identified by ownerType and cacheKey.
	GetInt(ctx context.Context, ownerType OwnerType, cacheKey string) (int64, bool, error)
	// Set stores or replaces a string cache value with a backend-neutral TTL.
	Set(ctx context.Context, ownerType OwnerType, cacheKey string, value string, ttl time.Duration) (*Item, error)
	// Delete removes the cache entry identified by ownerType and cacheKey.
	Delete(ctx context.Context, ownerType OwnerType, cacheKey string) error
	// Incr increments an integer cache value by delta using a backend-neutral TTL.
	Incr(ctx context.Context, ownerType OwnerType, cacheKey string, delta int64, ttl time.Duration) (*Item, error)
	// Expire updates the expiration policy of a cache entry.
	Expire(ctx context.Context, ownerType OwnerType, cacheKey string, ttl time.Duration) (bool, *gtime.Time, error)
	// CleanupExpired removes expired entries when the backend needs external cleanup.
	CleanupExpired(ctx context.Context) error
}

// Provider creates one kvcache backend implementation.
type Provider interface {
	// NewBackend returns one backend instance for the provider implementation.
	NewBackend() Backend
}

// Option customizes kvcache service construction.
type Option func(*serviceConfig)

// serviceConfig stores construction-time kvcache options.
type serviceConfig struct {
	provider Provider
	backend  Backend
}

// mysqlMemoryProvider adapts the internal MySQL MEMORY implementation to the
// public kvcache provider contract.
type mysqlMemoryProvider struct{}

// mysqlMemoryBackend adapts the internal MySQL MEMORY implementation to the
// public kvcache backend contract.
type mysqlMemoryBackend struct {
	backend *mysqlmemory.MySQLMemoryBackend
}

// newServiceConfig returns the default kvcache service configuration.
func newServiceConfig() *serviceConfig {
	return &serviceConfig{provider: NewMySQLMemoryProvider()}
}

// WithProvider configures the backend provider used by the kvcache service.
func WithProvider(provider Provider) Option {
	return func(config *serviceConfig) {
		if config != nil && provider != nil {
			config.provider = provider
			config.backend = nil
		}
	}
}

// WithBackend configures one concrete backend instance, primarily for tests and
// explicit in-process composition.
func WithBackend(backend Backend) Option {
	return func(config *serviceConfig) {
		if config != nil && backend != nil {
			config.backend = backend
		}
	}
}

// NewMySQLMemoryProvider returns the default MySQL MEMORY backend provider.
func NewMySQLMemoryProvider() Provider {
	return mysqlMemoryProvider{}
}

// NewBackend creates one MySQL MEMORY backend instance.
func (mysqlMemoryProvider) NewBackend() Backend {
	return &mysqlMemoryBackend{backend: mysqlmemory.NewMySQLMemoryBackend()}
}

// Name returns the stable MySQL MEMORY backend name.
func (b *mysqlMemoryBackend) Name() BackendName {
	return BackendMySQLMemory
}

// RequiresExpiredCleanup reports that MySQL MEMORY entries need scheduled
// cleanup because expiration is enforced by kvcache rather than the database.
func (b *mysqlMemoryBackend) RequiresExpiredCleanup() bool {
	return true
}

// Get reads one cache entry through the internal MySQL MEMORY implementation.
func (b *mysqlMemoryBackend) Get(ctx context.Context, ownerType OwnerType, cacheKey string) (*Item, bool, error) {
	item, ok, err := b.backend.Get(ctx, mysqlmemory.OwnerType(ownerType.String()), cacheKey)
	return convertInternalItem(item), ok, err
}

// GetInt reads one integer cache value through the internal MySQL MEMORY implementation.
func (b *mysqlMemoryBackend) GetInt(ctx context.Context, ownerType OwnerType, cacheKey string) (int64, bool, error) {
	return b.backend.GetInt(ctx, mysqlmemory.OwnerType(ownerType.String()), cacheKey)
}

// Set writes one string cache value through the internal MySQL MEMORY implementation.
func (b *mysqlMemoryBackend) Set(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
	value string,
	ttl time.Duration,
) (*Item, error) {
	item, err := b.backend.Set(ctx, mysqlmemory.OwnerType(ownerType.String()), cacheKey, value, ttl)
	return convertInternalItem(item), err
}

// Delete removes one cache entry through the internal MySQL MEMORY implementation.
func (b *mysqlMemoryBackend) Delete(ctx context.Context, ownerType OwnerType, cacheKey string) error {
	return b.backend.Delete(ctx, mysqlmemory.OwnerType(ownerType.String()), cacheKey)
}

// Incr increments one integer cache value through the internal MySQL MEMORY implementation.
func (b *mysqlMemoryBackend) Incr(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
	delta int64,
	ttl time.Duration,
) (*Item, error) {
	item, err := b.backend.Incr(ctx, mysqlmemory.OwnerType(ownerType.String()), cacheKey, delta, ttl)
	return convertInternalItem(item), err
}

// Expire updates one cache entry expiration through the internal MySQL MEMORY implementation.
func (b *mysqlMemoryBackend) Expire(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
	ttl time.Duration,
) (bool, *gtime.Time, error) {
	return b.backend.Expire(ctx, mysqlmemory.OwnerType(ownerType.String()), cacheKey, ttl)
}

// CleanupExpired removes expired entries through the internal MySQL MEMORY implementation.
func (b *mysqlMemoryBackend) CleanupExpired(ctx context.Context) error {
	return b.backend.CleanupExpired(ctx)
}

// convertInternalItem maps one internal backend item into the public kvcache
// item shape.
func convertInternalItem(item *mysqlmemory.Item) *Item {
	if item == nil {
		return nil
	}
	return &Item{
		Key:       item.Key,
		ValueKind: item.ValueKind,
		Value:     item.Value,
		IntValue:  item.IntValue,
		ExpireAt:  item.ExpireAt,
	}
}

// BackendName returns the active backend name.
func (s *serviceImpl) BackendName() BackendName {
	if s == nil || s.backend == nil {
		return ""
	}
	return s.backend.Name()
}

// RequiresExpiredCleanup reports whether the active backend needs external
// expired-entry cleanup.
func (s *serviceImpl) RequiresExpiredCleanup() bool {
	return s != nil && s.backend != nil && s.backend.RequiresExpiredCleanup()
}

// Get delegates one cache read to the active backend.
func (s *serviceImpl) Get(ctx context.Context, ownerType OwnerType, cacheKey string) (*Item, bool, error) {
	return s.backend.Get(ctx, ownerType, cacheKey)
}

// GetInt delegates one integer cache read to the active backend.
func (s *serviceImpl) GetInt(ctx context.Context, ownerType OwnerType, cacheKey string) (int64, bool, error) {
	return s.backend.GetInt(ctx, ownerType, cacheKey)
}

// Set delegates one cache write to the active backend.
func (s *serviceImpl) Set(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
	value string,
	ttl time.Duration,
) (*Item, error) {
	return s.backend.Set(ctx, ownerType, cacheKey, value, ttl)
}

// Delete delegates one cache delete to the active backend.
func (s *serviceImpl) Delete(ctx context.Context, ownerType OwnerType, cacheKey string) error {
	return s.backend.Delete(ctx, ownerType, cacheKey)
}

// Incr delegates one integer cache increment to the active backend.
func (s *serviceImpl) Incr(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
	delta int64,
	ttl time.Duration,
) (*Item, error) {
	return s.backend.Incr(ctx, ownerType, cacheKey, delta, ttl)
}

// Expire delegates one expiration update to the active backend.
func (s *serviceImpl) Expire(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
	ttl time.Duration,
) (bool, *gtime.Time, error) {
	return s.backend.Expire(ctx, ownerType, cacheKey, ttl)
}

// CleanupExpired delegates expired-entry cleanup to the active backend.
func (s *serviceImpl) CleanupExpired(ctx context.Context) error {
	return s.backend.CleanupExpired(ctx)
}
