// This file implements the guest-side distributed lock host-service client.

package domainhostcall

import (
	"context"
	"time"

	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// lockService adapts the lock host service to lockcap.Service.
type lockService struct{ baseService }

// Lock creates the distributed lock domain guest client.
func Lock(invoker HostServiceInvoker) lockcap.Service {
	return &lockService{baseService: newBaseServiceWithHostService(nil, invoker)}
}

// Acquire attempts to acquire one governed distributed lock.
func (s *lockService) Acquire(_ context.Context, in lockcap.AcquireInput) (*lockcap.AcquireOutput, error) {
	payload, err := s.callHostService(
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
func (s *lockService) Renew(_ context.Context, in lockcap.RenewInput) (*lockcap.RenewOutput, error) {
	payload, err := s.callHostService(
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
func (s *lockService) Release(_ context.Context, in lockcap.ReleaseInput) error {
	_, err := s.callHostService(
		protocol.HostServiceLock,
		protocol.HostServiceMethodLockRelease,
		in.Name,
		"",
		protocol.MarshalHostServiceLockReleaseRequest(&protocol.HostServiceLockReleaseRequest{Ticket: in.Ticket}),
	)
	return err
}

func leaseMillis(lease time.Duration) int64 {
	if lease <= 0 {
		return 0
	}
	return int64(lease / time.Millisecond)
}
