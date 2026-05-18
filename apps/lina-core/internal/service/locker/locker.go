// Package locker provides distributed lock acquisition, renewal, and lease
// management services for clustered host components.
package locker

import (
	"context"
	"sync"
	"time"

	"lina-core/internal/service/coordination"
)

// Service defines the locker service contract.
type Service interface {
	// Lock acquires a distributed lock when it is absent or expired. The name
	// and holder identify the resource and owner, reason is audit metadata, and
	// lease controls expiration. It returns the held instance, whether the lock
	// was acquired, and any SQL or coordination backend error.
	Lock(ctx context.Context, name, holder, reason string, lease time.Duration) (*Instance, bool, error)
	// LockFunc acquires a lock and executes the given function.
	// The lock is automatically released after the function completes; function
	// errors are returned after successful acquisition. Failure to acquire
	// returns ok=false with nil error, while backend failures are propagated.
	LockFunc(ctx context.Context, name, holder, reason string, lease time.Duration, f func() error) (bool, error)
	// Unlock releases one SQL-backed distributed lock identified by lock ID and
	// holder. When a coordination backend is active, callers must use name-based
	// release and this method returns ErrLockNotHeld.
	Unlock(ctx context.Context, lockID int64, holder string) error
	// Renew extends one SQL-backed distributed lock identified by lock ID and
	// holder. When a coordination backend is active, callers must use name-based
	// renewal and this method returns ErrLockNotHeld.
	Renew(ctx context.Context, lockID int64, holder string, lease time.Duration) error
	// UnlockByName releases one distributed lock identified by lock name and
	// holder across the active SQL or coordination backend. Unknown or already
	// expired locks are treated according to the backend implementation.
	UnlockByName(ctx context.Context, name string, holder string) error
	// RenewByName extends one distributed lock identified by lock name and
	// holder across the active SQL or coordination backend. It returns
	// ErrLockNotHeld when no live lock is owned by the holder.
	RenewByName(ctx context.Context, name string, holder string, lease time.Duration) error
}

// Interface compliance assertion for the default locker service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct{}

// processCoordinationLockStore stores the deployment-selected coordination
// lock backend used by service instances created before HTTP startup finishes.
var processCoordinationLockStore = struct {
	sync.RWMutex
	store coordination.LockStore
}{}

// New creates and returns a new locker Service instance.
func New() Service {
	return &serviceImpl{}
}
