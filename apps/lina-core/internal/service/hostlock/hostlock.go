// This file defines the plugin-facing distributed lock adapter built on top of the host locker service.

package hostlock

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/service/locker"
	"lina-core/pkg/bizerr"
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
	// Acquire attempts to acquire one plugin-scoped distributed lock.
	Acquire(ctx context.Context, in AcquireInput) (*AcquireOutput, error)
	// Renew extends one held lock using the issued lock ticket.
	Renew(ctx context.Context, pluginID string, tenantID int64, resourceRef string, ticket string) (*gtime.Time, error)
	// Release releases one held lock using the issued lock ticket.
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
func New(lockerSvc locker.Service) Service {
	if lockerSvc == nil {
		panic("hostlock service requires a non-nil locker service")
	}
	return &serviceImpl{
		lockerSvc: lockerSvc,
	}
}

// Acquire attempts to acquire one plugin-scoped distributed lock.
func (s *serviceImpl) Acquire(ctx context.Context, in AcquireInput) (*AcquireOutput, error) {
	actualLockName, err := buildActualLockName(in.PluginID, in.TenantID, in.ResourceRef)
	if err != nil {
		return nil, err
	}

	lease, err := normalizeLease(in.LeaseMillis)
	if err != nil {
		return nil, err
	}

	holder := buildLockHolder()
	instance, ok, err := s.lockerSvc.Lock(ctx, actualLockName, holder, buildLockReason(in.ResourceRef, in.RequestID), lease)
	if err != nil {
		return nil, err
	}
	if !ok || instance == nil {
		return &AcquireOutput{Acquired: false}, nil
	}

	ticket, err := encodeLockTicket(lockTicketClaims{
		LockID:      instance.ID(),
		LockName:    actualLockName,
		TenantID:    in.TenantID,
		PluginID:    strings.TrimSpace(in.PluginID),
		ResourceRef: strings.TrimSpace(in.ResourceRef),
		Holder:      instance.Holder(),
		LeaseMillis: lease.Milliseconds(),
	})
	if err != nil {
		return nil, err
	}

	return &AcquireOutput{
		Acquired: true,
		Ticket:   ticket,
		ExpireAt: gtime.Now().Add(lease),
	}, nil
}

// Renew extends one held lock using the issued lock ticket.
func (s *serviceImpl) Renew(ctx context.Context, pluginID string, tenantID int64, resourceRef string, ticket string) (*gtime.Time, error) {
	claims, err := decodeAndValidateTicket(ticket, pluginID, tenantID, resourceRef)
	if err != nil {
		return nil, err
	}

	lease, err := normalizeLease(claims.LeaseMillis)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(claims.LockName) != "" {
		err = s.lockerSvc.RenewByName(ctx, claims.LockName, claims.Holder, lease)
	} else {
		err = s.lockerSvc.Renew(ctx, claims.LockID, claims.Holder, lease)
	}
	if err != nil {
		return nil, err
	}
	return gtime.Now().Add(lease), nil
}

// Release releases one held lock using the issued lock ticket.
func (s *serviceImpl) Release(ctx context.Context, pluginID string, tenantID int64, resourceRef string, ticket string) error {
	claims, err := decodeAndValidateTicket(ticket, pluginID, tenantID, resourceRef)
	if err != nil {
		return err
	}
	if strings.TrimSpace(claims.LockName) != "" {
		return s.lockerSvc.UnlockByName(ctx, claims.LockName, claims.Holder)
	}
	return s.lockerSvc.Unlock(ctx, claims.LockID, claims.Holder)
}

// buildActualLockName combines the plugin, tenant, and logical resource name
// into one bounded lock key accepted by the underlying locker service.
func buildActualLockName(pluginID string, tenantID int64, resourceRef string) (string, error) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	normalizedResourceRef := strings.TrimSpace(resourceRef)
	if normalizedPluginID == "" {
		return "", bizerr.NewCode(CodeHostLockPluginIDRequired)
	}
	if normalizedResourceRef == "" {
		return "", bizerr.NewCode(CodeHostLockResourceRequired)
	}

	actualLockName := "plugin:" + normalizedPluginID + ":tenant=" + strconv.FormatInt(tenantID, 10) + ":" + normalizedResourceRef
	if len([]byte(actualLockName)) > maxLockBytes {
		return "", bizerr.NewCode(CodeHostLockNameTooLong, bizerr.P("maxBytes", maxLockBytes))
	}
	return actualLockName, nil
}

// TenantIDFromIdentity normalizes a plugin identity tenant into the hostlock
// lock-name discriminator. Platform and anonymous executions use tenant 0.
func TenantIDFromIdentity(tenantID int32) int64 {
	if tenantID <= 0 {
		return int64(pkgtenantcap.PLATFORM)
	}
	return int64(tenantID)
}

// normalizeLease converts the request lease in milliseconds into a validated
// duration while enforcing host-level defaults and boundaries.
func normalizeLease(leaseMillis int64) (time.Duration, error) {
	if leaseMillis <= 0 {
		return defaultLease, nil
	}

	lease := time.Duration(leaseMillis) * time.Millisecond
	if lease < minLease {
		return 0, bizerr.NewCode(CodeHostLockLeaseTooShort, bizerr.P("minLease", minLease))
	}
	if lease > maxLease {
		return 0, bizerr.NewCode(CodeHostLockLeaseTooLong, bizerr.P("maxLease", maxLease))
	}
	return lease, nil
}

// buildLockHolder creates one unique holder token used by the distributed
// locker implementation to identify the current caller.
func buildLockHolder() string {
	return "pl:" + guid.S()
}

// buildLockReason generates one audit-friendly lock reason string that records
// the logical resource and optional request identifier.
func buildLockReason(resourceRef string, requestID string) string {
	normalizedResourceRef := strings.TrimSpace(resourceRef)
	normalizedRequestID := strings.TrimSpace(requestID)
	if normalizedRequestID == "" {
		return "plugin host lock: " + normalizedResourceRef
	}
	return "plugin host lock: " + normalizedResourceRef + " request=" + normalizedRequestID
}
