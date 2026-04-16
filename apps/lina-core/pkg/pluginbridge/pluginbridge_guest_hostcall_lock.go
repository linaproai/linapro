//go:build wasip1

// This file provides guest-side helpers for the governed distributed lock host service.

package pluginbridge

// LockHostService exposes guest-side helpers for the governed distributed lock host service.
type LockHostService interface {
	// Acquire attempts to acquire one governed distributed lock.
	Acquire(lockName string, leaseMillis int64) (*HostServiceLockAcquireResponse, error)
	// Renew extends one governed distributed lock using the issued ticket.
	Renew(lockName string, ticket string) (*HostServiceLockRenewResponse, error)
	// Release releases one governed distributed lock using the issued ticket.
	Release(lockName string, ticket string) error
}

type lockHostService struct{}

var defaultLockHostService LockHostService = &lockHostService{}

// Lock returns the distributed lock host service guest client.
func Lock() LockHostService {
	return defaultLockHostService
}

// Acquire attempts to acquire one governed distributed lock.
func (s *lockHostService) Acquire(lockName string, leaseMillis int64) (*HostServiceLockAcquireResponse, error) {
	payload, err := invokeHostService(
		HostServiceLock,
		HostServiceMethodLockAcquire,
		lockName,
		"",
		MarshalHostServiceLockAcquireRequest(&HostServiceLockAcquireRequest{LeaseMillis: leaseMillis}),
	)
	if err != nil {
		return nil, err
	}
	return UnmarshalHostServiceLockAcquireResponse(payload)
}

// Renew extends one governed distributed lock using the issued ticket.
func (s *lockHostService) Renew(lockName string, ticket string) (*HostServiceLockRenewResponse, error) {
	payload, err := invokeHostService(
		HostServiceLock,
		HostServiceMethodLockRenew,
		lockName,
		"",
		MarshalHostServiceLockRenewRequest(&HostServiceLockRenewRequest{Ticket: ticket}),
	)
	if err != nil {
		return nil, err
	}
	return UnmarshalHostServiceLockRenewResponse(payload)
}

// Release releases one governed distributed lock using the issued ticket.
func (s *lockHostService) Release(lockName string, ticket string) error {
	_, err := invokeHostService(
		HostServiceLock,
		HostServiceMethodLockRelease,
		lockName,
		"",
		MarshalHostServiceLockReleaseRequest(&HostServiceLockReleaseRequest{Ticket: ticket}),
	)
	return err
}
