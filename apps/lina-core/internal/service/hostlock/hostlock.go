// This file defines the plugin-facing distributed lock adapter built on top of the host locker service.

package hostlock

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/service/locker"
)

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
	Renew(ctx context.Context, pluginID string, resourceRef string, ticket string) (*gtime.Time, error)
	// Release releases one held lock using the issued lock ticket.
	Release(ctx context.Context, pluginID string, resourceRef string, ticket string) error
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	lockerSvc locker.Service // Underlying distributed locker service
}

// AcquireInput defines one distributed lock acquire request.
type AcquireInput struct {
	// PluginID is the current calling plugin identifier.
	PluginID string
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
func New() Service {
	return &serviceImpl{
		lockerSvc: locker.New(),
	}
}

// Acquire attempts to acquire one plugin-scoped distributed lock.
func (s *serviceImpl) Acquire(ctx context.Context, in AcquireInput) (*AcquireOutput, error) {
	actualLockName, err := buildActualLockName(in.PluginID, in.ResourceRef)
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
func (s *serviceImpl) Renew(ctx context.Context, pluginID string, resourceRef string, ticket string) (*gtime.Time, error) {
	claims, err := decodeAndValidateTicket(ticket, pluginID, resourceRef)
	if err != nil {
		return nil, err
	}

	lease, err := normalizeLease(claims.LeaseMillis)
	if err != nil {
		return nil, err
	}
	if err = s.lockerSvc.Renew(ctx, claims.LockID, claims.Holder, lease); err != nil {
		return nil, err
	}
	return gtime.Now().Add(lease), nil
}

// Release releases one held lock using the issued lock ticket.
func (s *serviceImpl) Release(ctx context.Context, pluginID string, resourceRef string, ticket string) error {
	claims, err := decodeAndValidateTicket(ticket, pluginID, resourceRef)
	if err != nil {
		return err
	}
	return s.lockerSvc.Unlock(ctx, claims.LockID, claims.Holder)
}

func buildActualLockName(pluginID string, resourceRef string) (string, error) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	normalizedResourceRef := strings.TrimSpace(resourceRef)
	if normalizedPluginID == "" {
		return "", gerror.New("插件ID不能为空")
	}
	if normalizedResourceRef == "" {
		return "", gerror.New("逻辑锁名不能为空")
	}

	actualLockName := "plugin:" + normalizedPluginID + ":" + normalizedResourceRef
	if len([]byte(actualLockName)) > maxLockBytes {
		return "", gerror.Newf("实际锁名长度超出限制，最大允许 %d 字节", maxLockBytes)
	}
	return actualLockName, nil
}

func normalizeLease(leaseMillis int64) (time.Duration, error) {
	if leaseMillis <= 0 {
		return defaultLease, nil
	}

	lease := time.Duration(leaseMillis) * time.Millisecond
	if lease < minLease {
		return 0, gerror.Newf("锁租期不能小于 %s", minLease)
	}
	if lease > maxLease {
		return 0, gerror.Newf("锁租期不能大于 %s", maxLease)
	}
	return lease, nil
}

func buildLockHolder() string {
	return "pl:" + guid.S()
}

func buildLockReason(resourceRef string, requestID string) string {
	normalizedResourceRef := strings.TrimSpace(resourceRef)
	normalizedRequestID := strings.TrimSpace(requestID)
	if normalizedRequestID == "" {
		return "plugin host lock: " + normalizedResourceRef
	}
	return "plugin host lock: " + normalizedResourceRef + " request=" + normalizedRequestID
}
