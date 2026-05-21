// This file tests dialect-level database driver error classification.

package dialect

import (
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"
)

// Retryable test SQLSTATE codes reported by PostgreSQL drivers.
const (
	testPGSerializationFailure = "40001"
	testPGDeadlockDetected     = "40P01"
	testPGUniqueViolation      = "23505"
)

// fakePostgreSQLError mimics the narrow SQLState shape exposed by supported
// PostgreSQL drivers.
type fakePostgreSQLError struct {
	state string
}

// Error returns a compact fake PostgreSQL error message.
func (e fakePostgreSQLError) Error() string {
	return "fake postgres error"
}

// SQLState returns the fake PostgreSQL SQLSTATE.
func (e fakePostgreSQLError) SQLState() string {
	return e.state
}

// TestIsRetryableWriteConflictClassifiesPostgreSQLErrors verifies PostgreSQL
// retryable write conflicts are recognized by SQLSTATE rather than text.
func TestIsRetryableWriteConflictClassifiesPostgreSQLErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "serialization failure",
			err:  fakePostgreSQLError{state: testPGSerializationFailure},
			want: true,
		},
		{
			name: "deadlock wrapped by goframe",
			err: gerror.Wrap(
				fakePostgreSQLError{state: testPGDeadlockDetected},
				"update failed",
			),
			want: true,
		},
		{
			name: "unique violation is not retryable",
			err:  fakePostgreSQLError{state: testPGUniqueViolation},
			want: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := IsRetryableWriteConflict(test.err); got != test.want {
				t.Fatalf("expected retryable=%t, got %t", test.want, got)
			}
		})
	}
}

// TestIsUniqueConstraintViolationClassifiesDatabaseErrors verifies duplicate
// writes are recognized without matching driver error text.
func TestIsUniqueConstraintViolationClassifiesDatabaseErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "postgres unique",
			err:  fakePostgreSQLError{state: testPGUniqueViolation},
			want: true,
		},
		{
			name: "postgres deadlock",
			err:  fakePostgreSQLError{state: testPGDeadlockDetected},
			want: false,
		},
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
