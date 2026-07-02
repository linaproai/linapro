// This file keeps lock-expiration helpers scoped to tests.

package locker

import "time"

// isExpiredLock reports whether one lock row is available for takeover.
func isExpiredLock(expireTime *time.Time, now time.Time) bool {
	if expireTime == nil {
		return true
	}
	return now.After(*expireTime)
}
