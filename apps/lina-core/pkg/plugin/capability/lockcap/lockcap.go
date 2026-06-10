// Package lockcap defines the plugin-visible distributed lock capability.
// The package owns only the stable domain contract and DTOs; host adapters are
// responsible for binding plugin and tenant scope before touching locker
// backends or bridge transport.
package lockcap

import (
	"context"
	"time"
)

// Lock lease boundaries published to source and dynamic plugins.
const (
	// DefaultLease is used when a caller does not request a lease.
	DefaultLease = 30 * time.Second
	// MinLease is the shortest accepted lock lease.
	MinLease = 1 * time.Second
	// MaxLease is the longest accepted lock lease.
	MaxLease = 5 * time.Minute
	// MaxNameBytes is the maximum logical lock name length accepted by the
	// plugin lock capability. Host adapters may reserve additional bytes for
	// internal scope prefixes.
	MaxNameBytes = 64
)

// AcquireInput defines one plugin lock acquisition request.
type AcquireInput struct {
	// Name is the plugin-local logical lock name.
	Name string
	// Lease is the requested lock lease. Zero uses DefaultLease.
	Lease time.Duration
}

// AcquireOutput defines one plugin lock acquisition result.
type AcquireOutput struct {
	// Acquired reports whether this call acquired the lock.
	Acquired bool
	// Ticket is an opaque host-generated token required by Renew and Release.
	Ticket string
	// ExpireAt is the expected expiration time when Acquired is true.
	ExpireAt *time.Time
}

// RenewInput defines one plugin lock renewal request.
type RenewInput struct {
	// Name is the plugin-local logical lock name.
	Name string
	// Ticket is the opaque token returned by Acquire.
	Ticket string
}

// RenewOutput defines one plugin lock renewal result.
type RenewOutput struct {
	// ExpireAt is the expected expiration time after renewal.
	ExpireAt *time.Time
}

// ReleaseInput defines one plugin lock release request.
type ReleaseInput struct {
	// Name is the plugin-local logical lock name.
	Name string
	// Ticket is the opaque token returned by Acquire.
	Ticket string
}

// Service defines the plugin-visible distributed lock operations. The
// implementation must scope every call by current plugin ID and tenant context,
// and must treat Ticket as opaque to callers.
type Service interface {
	// Acquire attempts to acquire one plugin-scoped lock and returns a ticket
	// only when the lock is acquired.
	Acquire(ctx context.Context, in AcquireInput) (*AcquireOutput, error)
	// Renew extends one plugin-scoped lock using the ticket issued by Acquire.
	Renew(ctx context.Context, in RenewInput) (*RenewOutput, error)
	// Release releases one plugin-scoped lock using the ticket issued by Acquire.
	Release(ctx context.Context, in ReleaseInput) error
}
