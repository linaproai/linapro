// This file contains test-only helpers for resetting role access revision state.

package role

import "time"

// clearLocalAccessRevision drops the process-local revision so tests can force
// the next read to resynchronize from the configured revision controller.
func clearLocalAccessRevision() {
	accessRevisionState.Lock()
	accessRevisionState.value = 0
	accessRevisionState.expireAt = time.Time{}
	accessRevisionState.Unlock()
}
