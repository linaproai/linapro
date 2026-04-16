package locker

import "github.com/gogf/gf/v2/errors/gerror"

// Error definitions for the locker package.
var (
	// ErrLockNotHeld is returned when trying to renew a lock that is not held.
	ErrLockNotHeld = gerror.New("lock not held by current node")

	// ErrRenewalFailed is returned when lease renewal fails.
	ErrRenewalFailed = gerror.New("lease renewal failed")
)
