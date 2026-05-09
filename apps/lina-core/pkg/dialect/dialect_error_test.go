// This file tests dialect-level database driver error classification.

package dialect

import (
	"errors"
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

// fakeSQLiteCodeError mimics the narrow Code() shape exposed by supported
// SQLite drivers without depending on unexported driver error fields.
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

// TestIsRetryableWriteConflictClassifiesSQLiteErrors verifies SQLite retryable
// lock conflicts are recognized by result code rather than text.
func TestIsRetryableWriteConflictClassifiesSQLiteErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "busy",
			err:  fakeSQLiteCodeError{code: 5},
			want: true,
		},
		{
			name: "locked wrapped by goframe",
			err:  gerror.Wrap(fakeSQLiteCodeError{code: 6}, "update failed"),
			want: true,
		},
		{
			name: "busy extended",
			err:  fakeSQLiteCodeError{code: 5 | (1 << 8)},
			want: true,
		},
		{
			name: "constraint is not retryable",
			err:  fakeSQLiteCodeError{code: 19},
			want: false,
		},
		{
			name: "plain error is not retryable",
			err:  errors.New("database is locked"),
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
		{
			name: "sqlite unique",
			err:  fakeSQLiteCodeError{code: 19 | (8 << 8)},
			want: true,
		},
		{
			name: "sqlite generic constraint",
			err:  fakeSQLiteCodeError{code: 19},
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
