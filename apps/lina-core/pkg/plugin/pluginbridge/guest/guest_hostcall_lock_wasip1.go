//go:build wasip1

// This file provides guest-side helpers for the governed distributed lock host service.

package guest

import "lina-core/pkg/plugin/pluginbridge/protocol"

// lockHostService is the default guest-side distributed lock host-service
// client.
type lockHostService struct{}

// defaultLockHostService stores the singleton lock host-service client used by
// package-level helpers.
var defaultLockHostService LockHostService = &lockHostService{}

// Lock returns the distributed lock host service guest client.
func Lock() LockHostService {
	return defaultLockHostService
}

// Acquire attempts to acquire one governed distributed lock.
func (s *lockHostService) Acquire(lockName string, leaseMillis int64) (*protocol.HostServiceLockAcquireResponse, error) {
	payload, err := invokeGuestHostService(
		protocol.HostServiceLock,
		protocol.HostServiceMethodLockAcquire,
		lockName,
		"",
		protocol.MarshalHostServiceLockAcquireRequest(&protocol.HostServiceLockAcquireRequest{LeaseMillis: leaseMillis}),
	)
	if err != nil {
		return nil, err
	}
	return protocol.UnmarshalHostServiceLockAcquireResponse(payload)
}

// Renew extends one governed distributed lock using the issued ticket.
func (s *lockHostService) Renew(lockName string, ticket string) (*protocol.HostServiceLockRenewResponse, error) {
	payload, err := invokeGuestHostService(
		protocol.HostServiceLock,
		protocol.HostServiceMethodLockRenew,
		lockName,
		"",
		protocol.MarshalHostServiceLockRenewRequest(&protocol.HostServiceLockRenewRequest{Ticket: ticket}),
	)
	if err != nil {
		return nil, err
	}
	return protocol.UnmarshalHostServiceLockRenewResponse(payload)
}

// Release releases one governed distributed lock using the issued ticket.
func (s *lockHostService) Release(lockName string, ticket string) error {
	_, err := invokeGuestHostService(
		protocol.HostServiceLock,
		protocol.HostServiceMethodLockRelease,
		lockName,
		"",
		protocol.MarshalHostServiceLockReleaseRequest(&protocol.HostServiceLockReleaseRequest{Ticket: ticket}),
	)
	return err
}
