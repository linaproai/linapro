// This file defines backend and provider abstractions for the kvcache service.

package kvcache

import (
	"context"
	"time"

	"lina-core/internal/service/coordination"
	"lina-core/internal/service/kvcache/internal/contract"
	"lina-core/internal/service/kvcache/internal/coordkv"
	"lina-core/internal/service/kvcache/internal/memory"
)

// BackendName identifies one kvcache backend implementation.
type BackendName = contract.BackendName

// Supported backend names for the kvcache service.
const (
	// BackendMemory stores cache entries in the host process memory.
	BackendMemory = contract.BackendMemory
	// BackendCoordinationKV stores cache entries in the configured coordination KV backend.
	BackendCoordinationKV = contract.BackendCoordinationKV
)

// Backend defines the backend-specific KV cache operations used by Service.
type Backend = contract.Backend

// Provider creates one kvcache backend implementation.
type Provider = contract.Provider

// Option customizes kvcache service construction.
type Option func(*serviceConfig)

// serviceConfig stores construction-time kvcache options.
type serviceConfig struct {
	provider Provider
	backend  Backend
}

// newServiceConfig returns the default kvcache service configuration.
func newServiceConfig() *serviceConfig {
	return &serviceConfig{provider: NewMemoryProvider()}
}

// NewMemoryProvider returns the single-node in-process kvcache provider.
func NewMemoryProvider() Provider {
	return memory.NewProvider()
}

// NewCoordinationKVProvider returns a kvcache backend provider backed by
// coordination KVStore.
func NewCoordinationKVProvider(coordinationSvc coordination.Service) Provider {
	return coordkv.NewProvider(coordinationSvc)
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
	if err := contract.ValidatePositiveTTL(ttl); err != nil {
		return nil, err
	}
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
	if err := contract.ValidatePositiveTTL(ttl); err != nil {
		return nil, err
	}
	return s.backend.Incr(ctx, ownerType, cacheKey, delta, ttl)
}

// Expire delegates one expiration update to the active backend.
func (s *serviceImpl) Expire(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
	ttl time.Duration,
) (bool, *time.Time, error) {
	if err := contract.ValidatePositiveTTL(ttl); err != nil {
		return false, nil, err
	}
	return s.backend.Expire(ctx, ownerType, cacheKey, ttl)
}

// CleanupExpired delegates expired-entry cleanup to the active backend.
func (s *serviceImpl) CleanupExpired(ctx context.Context) error {
	return s.backend.CleanupExpired(ctx)
}
