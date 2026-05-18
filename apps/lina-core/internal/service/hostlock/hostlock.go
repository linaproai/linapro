// Package hostlock exposes plugin-facing distributed lock contracts backed by
// the host locker service.
package hostlock

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/service/locker"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// Lock normalization constants shared by acquire, renew, and release paths.
const (
	defaultLease = 30 * time.Second
	minLease     = 1 * time.Second
	maxLease     = 5 * time.Minute
	maxLockBytes = 64
)

// Service defines the hostlock service contract.
type Service interface {
	// Acquire attempts to acquire one plugin-scoped distributed lock. The input
	// identifies the calling plugin, tenant boundary, logical resource, and
	// requested lease. Validation errors are returned as hostlock business
	// errors; locker backend failures are propagated. Successful acquisition
	// returns an opaque ticket required by Renew and Release.
	Acquire(ctx context.Context, in AcquireInput) (*AcquireOutput, error)
	// Renew extends one held lock using the issued lock ticket. Plugin, tenant,
	// and resource parameters must match the ticket claims; mismatches or stale
	// tickets return business or lock ownership errors.
	Renew(ctx context.Context, pluginID string, tenantID int64, resourceRef string, ticket string) (*gtime.Time, error)
	// Release releases one held lock using the issued lock ticket. Plugin,
	// tenant, and resource parameters must match the ticket claims; backend
	// release errors are returned to the caller.
	Release(ctx context.Context, pluginID string, tenantID int64, resourceRef string, ticket string) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	lockerSvc locker.Service // Underlying distributed locker service
}

// AcquireInput defines one distributed lock acquire request.
type AcquireInput struct {
	// PluginID is the current calling plugin identifier.
	PluginID string
	// TenantID is the current tenant boundary for tenant-scoped plugin locks.
	TenantID int64
	// ResourceRef is the logical lock name declared in hostServices.
	ResourceRef string
	// LeaseMillis is the requested lease duration in milliseconds.
	LeaseMillis int64
	// RequestID is the optional host request identifier used in audit reason strings.
	RequestID string
}

// AcquireOutput defines one distributed lock acquire result.
type AcquireOutput struct {
	// Acquired reports whether the lock was acquired successfully.
	Acquired bool
	// Ticket is the opaque lock ticket used for renew and release.
	Ticket string
	// ExpireAt is the next expiration time of the held lock.
	ExpireAt *gtime.Time
}

// New creates and returns a new plugin-facing host lock service instance.
func New(lockerSvc locker.Service) (Service, error) {
	if lockerSvc == nil {
		return nil, gerror.New("hostlock service requires a non-nil locker service")
	}
	return &serviceImpl{
		lockerSvc: lockerSvc,
	}, nil
}

// TenantIDFromIdentity normalizes a plugin identity tenant into the hostlock
// lock-name discriminator. Platform and anonymous executions use tenant 0.
func TenantIDFromIdentity(tenantID int32) int64 {
	if tenantID <= 0 {
		return int64(pkgtenantcap.PLATFORM)
	}
	return int64(tenantID)
}
