// This file tests exported locker error values and their basic semantics.

package locker

import (
	"errors"
	"testing"

	"github.com/gogf/gf/v2/test/gtest"
)

// TestErrors verifies locker error sentinels remain non-nil and stable.
func TestErrors(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// Verify error definitions
		t.Assert(ErrLockNotHeld != nil, true)
		t.Assert(ErrRenewalFailed != nil, true)

		// Verify error messages
		t.Assert(errors.Is(ErrLockNotHeld, ErrLockNotHeld), true)
		t.Assert(errors.Is(ErrRenewalFailed, ErrRenewalFailed), true)

		// Verify error strings
		t.Assert(ErrLockNotHeld.Error(), "lock not held by current node")
		t.Assert(ErrRenewalFailed.Error(), "lease renewal failed")
	})
}
