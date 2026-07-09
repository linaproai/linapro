// This file adapts the host shared lock backend into the plugin-visible
// lockcap contract while binding every call to one plugin and tenant scope.

package capabilityhost

import (
	"context"
	"strings"
	"time"

	"lina-core/internal/service/hostlock"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// lockAdapter binds the shared host lock service to one plugin.
type lockAdapter struct {
	service  hostlock.Service
	bizCtx   bizctxcap.Service
	pluginID string
}

// newLockAdapter creates one plugin-scoped lock adapter.
func newLockAdapter(
	service hostlock.Service,
	bizCtx bizctxcap.Service,
	pluginID string,
) lockcap.Service {
	return &lockAdapter{
		service:  service,
		bizCtx:   bizCtx,
		pluginID: strings.TrimSpace(pluginID),
	}
}

// Acquire attempts to acquire one plugin-scoped lock.
func (s *lockAdapter) Acquire(ctx context.Context, in lockcap.AcquireInput) (*lockcap.AcquireOutput, error) {
	lockName, lease, err := s.validateAcquireInput(in)
	if err != nil {
		return nil, err
	}
	output, err := s.service.Acquire(ctx, hostlock.AcquireInput{
		PluginID:    s.pluginID,
		TenantID:    s.currentTenantID(ctx),
		ResourceRef: lockName,
		LeaseMillis: lease.Milliseconds(),
	})
	if err != nil {
		return nil, err
	}
	if output == nil {
		return &lockcap.AcquireOutput{Acquired: false}, nil
	}
	return &lockcap.AcquireOutput{
		Acquired: output.Acquired,
		Ticket:   output.Ticket,
		ExpireAt: output.ExpireAt,
	}, nil
}

// Renew extends one plugin-scoped lock using a previously issued ticket.
func (s *lockAdapter) Renew(ctx context.Context, in lockcap.RenewInput) (*lockcap.RenewOutput, error) {
	lockName, err := s.validateTicketInput(in.Name, in.Ticket)
	if err != nil {
		return nil, err
	}
	expireAt, err := s.service.Renew(ctx, s.pluginID, s.currentTenantID(ctx), lockName, in.Ticket)
	if err != nil {
		return nil, err
	}
	return &lockcap.RenewOutput{ExpireAt: expireAt}, nil
}

// Release releases one plugin-scoped lock using a previously issued ticket.
func (s *lockAdapter) Release(ctx context.Context, in lockcap.ReleaseInput) error {
	lockName, err := s.validateTicketInput(in.Name, in.Ticket)
	if err != nil {
		return err
	}
	return s.service.Release(ctx, s.pluginID, s.currentTenantID(ctx), lockName, in.Ticket)
}

// validateAcquireInput validates plugin scope, lock name, and lease duration.
func (s *lockAdapter) validateAcquireInput(in lockcap.AcquireInput) (string, time.Duration, error) {
	if err := s.validateServiceScope(); err != nil {
		return "", 0, err
	}
	lockName, err := normalizeLockName(in.Name)
	if err != nil {
		return "", 0, err
	}
	lease := in.Lease
	if lease == 0 {
		lease = lockcap.DefaultLease
	}
	if lease < lockcap.MinLease {
		return "", 0, bizerr.NewCode(lockcap.CodeLockLeaseTooShort, bizerr.P("minLease", lockcap.MinLease))
	}
	if lease > lockcap.MaxLease {
		return "", 0, bizerr.NewCode(lockcap.CodeLockLeaseTooLong, bizerr.P("maxLease", lockcap.MaxLease))
	}
	return lockName, lease, nil
}

// validateTicketInput validates plugin scope, logical lock name, and ticket.
func (s *lockAdapter) validateTicketInput(name string, ticket string) (string, error) {
	if err := s.validateServiceScope(); err != nil {
		return "", err
	}
	lockName, err := normalizeLockName(name)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(ticket) == "" {
		return "", bizerr.NewCode(lockcap.CodeLockTicketRequired)
	}
	return lockName, nil
}

// validateServiceScope ensures the adapter is bound to a runtime lock backend and plugin ID.
func (s *lockAdapter) validateServiceScope() error {
	if s == nil || s.service == nil {
		return bizerr.NewCode(lockcap.CodeLockServiceUnavailable)
	}
	if strings.TrimSpace(s.pluginID) == "" {
		return bizerr.NewCode(lockcap.CodeLockPluginIDRequired)
	}
	return nil
}

// normalizeLockName trims and validates one logical lock name.
func normalizeLockName(name string) (string, error) {
	normalized := strings.TrimSpace(name)
	if normalized == "" {
		return "", bizerr.NewCode(lockcap.CodeLockNameRequired)
	}
	if len([]byte(normalized)) > lockcap.MaxNameBytes {
		return "", bizerr.NewCode(lockcap.CodeLockNameTooLong, bizerr.P("maxBytes", lockcap.MaxNameBytes))
	}
	return normalized, nil
}

// currentTenantID returns the current plugin lock tenant scope.
func (s *lockAdapter) currentTenantID(ctx context.Context) int64 {
	if s != nil && s.bizCtx != nil {
		if tenantID := s.bizCtx.Current(ctx).TenantID; tenantID > 0 {
			return int64(tenantID)
		}
	}
	if tenantID := bizctxcap.CurrentFromContext(ctx).TenantID; tenantID > 0 {
		return int64(tenantID)
	}
	return int64(tenantcap.PLATFORM)
}
