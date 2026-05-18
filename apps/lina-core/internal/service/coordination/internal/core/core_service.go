// This file implements the lightweight service accessors for the backend-neutral
// coordination facade; it intentionally contains no backend-specific behavior.

package core

import "context"

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
