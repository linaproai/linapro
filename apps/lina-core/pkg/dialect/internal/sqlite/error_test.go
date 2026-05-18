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

// fakeSQLiteCodeError mimics the narrow Code() shape exposed by supported
// SQLite drivers.
type fakeSQLiteCodeError struct {
	code int
}

// Error returns a compact fake SQLite error message.
func (e fakeSQLiteCodeError) Error() string {
	return "fake sqlite error"
}

// Code returns the fake SQLite result code.
func (e fakeSQLiteCodeError) Code() int {
	return e.code
}

// TestIsUniqueConstraintViolation verifies only unique and primary-key
// constraint extended codes are treated as duplicate-write conflicts.
func TestIsUniqueConstraintViolation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "primary key", err: fakeSQLiteCodeError{code: errorConstraintPK}, want: true},
		{name: "unique", err: fakeSQLiteCodeError{code: errorConstraintUnique}, want: true},
		{name: "primary constraint only", err: fakeSQLiteCodeError{code: errorConstraint}, want: false},
		{name: "busy", err: fakeSQLiteCodeError{code: errorBusy}, want: false},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := IsUniqueConstraintViolation(test.err); got != test.want {
				t.Fatalf("expected unique constraint=%t, got %t", test.want, got)
			}
		})
	}
}
