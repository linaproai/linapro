// This file tests SQLite result-code classification.

package sqlite

import "testing"

// TestIsRetryablePrimaryCode verifies SQLite write-conflict primary result codes.
func TestIsRetryablePrimaryCode(t *testing.T) {
	t.Parallel()

	if !IsRetryablePrimaryCode(5) {
		t.Fatal("expected SQLITE_BUSY primary code to be retryable")
	}
	if !IsRetryablePrimaryCode(6 | (1 << 8)) {
		t.Fatal("expected extended SQLITE_LOCKED code to be retryable")
	}
	if IsRetryablePrimaryCode(19) {
		t.Fatal("expected SQLITE_CONSTRAINT primary code not to be retryable")
	}
}
