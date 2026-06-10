//go:build wasip1

// This file adapts the governed lock host service transport to lockcap.Service.

package pluginbridge

import (
	"context"
	"time"

	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// lockHostService is the default guest-side distributed lock host-service
// client.
type lockHostService struct{}

// defaultLockHostService stores the singleton lock host-service client used by
// package-level helpers.
var defaultLockHostService lockcap.Service = &lockHostService{}

// Lock returns the distributed lock domain guest client.
func Lock() lockcap.Service {
	return defaultLockHostService
}

// Acquire attempts to acquire one governed distributed lock.
func (s *lockHostService) Acquire(_ context.Context, in lockcap.AcquireInput) (*lockcap.AcquireOutput, error) {
	payload, err := invokeGuestHostService(
		protocol.HostServiceLock,
		protocol.HostServiceMethodLockAcquire,
		in.Name,
		"",
		protocol.MarshalHostServiceLockAcquireRequest(&protocol.HostServiceLockAcquireRequest{LeaseMillis: leaseMillis(in.Lease)}),
	)
	if err != nil {
		return nil, err
	}
	response, err := protocol.UnmarshalHostServiceLockAcquireResponse(payload)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil
	}
	return &lockcap.AcquireOutput{
		Acquired: response.Acquired,
		Ticket:   response.Ticket,
		ExpireAt: parseWireTime(response.ExpireAt),
	}, nil
}

// Renew extends one governed distributed lock using the issued ticket.
func (s *lockHostService) Renew(_ context.Context, in lockcap.RenewInput) (*lockcap.RenewOutput, error) {
	payload, err := invokeGuestHostService(
		protocol.HostServiceLock,
		protocol.HostServiceMethodLockRenew,
		in.Name,
		"",
		protocol.MarshalHostServiceLockRenewRequest(&protocol.HostServiceLockRenewRequest{Ticket: in.Ticket}),
	)
	if err != nil {
		return nil, err
	}
	response, err := protocol.UnmarshalHostServiceLockRenewResponse(payload)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil
	}
	return &lockcap.RenewOutput{ExpireAt: parseWireTime(response.ExpireAt)}, nil
}

// Release releases one governed distributed lock using the issued ticket.
func (s *lockHostService) Release(_ context.Context, in lockcap.ReleaseInput) error {
	_, err := invokeGuestHostService(
		protocol.HostServiceLock,
		protocol.HostServiceMethodLockRelease,
		in.Name,
		"",
		protocol.MarshalHostServiceLockReleaseRequest(&protocol.HostServiceLockReleaseRequest{Ticket: in.Ticket}),
	)
	return err
}

// leaseMillis converts a domain lock lease to wire milliseconds.
func leaseMillis(lease time.Duration) int64 {
	if lease <= 0 {
		return 0
	}
	return int64(lease / time.Millisecond)
}
