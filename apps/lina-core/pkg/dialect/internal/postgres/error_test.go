// This file tests PostgreSQL SQLSTATE classification helpers.

package postgres

import (
	"errors"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"
)

// fakeSQLStateError mimics the SQLSTATE shape exposed by PostgreSQL drivers.
type fakeSQLStateError struct {
	state string
}

// Error returns a compact fake PostgreSQL error message.
func (e fakeSQLStateError) Error() string {
	return "fake postgres error"
}

// SQLState returns the fake PostgreSQL SQLSTATE.
func (e fakeSQLStateError) SQLState() string {
	return e.state
}

// TestSQLStateExtractsWrappedDriverState verifies SQLSTATE extraction works
// through GoFrame error wrapping and trims driver output.
func TestSQLStateExtractsWrappedDriverState(t *testing.T) {
	t.Parallel()

	err := gerror.Wrap(fakeSQLStateError{state: " 40001 "}, "write failed")
	if got := SQLState(err); got != ErrorSerializationFailure {
		t.Fatalf("expected SQLSTATE %s, got %q", ErrorSerializationFailure, got)
	}
	if got := SQLState(errors.New("plain error")); got != "" {
		t.Fatalf("expected no SQLSTATE for plain error, got %q", got)
	}
}

// TestIsRetryableSQLState verifies transient write-conflict SQLSTATE handling.
func TestIsRetryableSQLState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
		want bool
	}{
		{name: "serialization failure", code: ErrorSerializationFailure, want: true},
		{name: "deadlock", code: ErrorDeadlockDetected, want: true},
		{name: "lock not available", code: ErrorLockNotAvailable, want: true},
		{name: "trimmed", code: " " + ErrorDeadlockDetected + " ", want: true},
		{name: "unique violation", code: ErrorUniqueViolation, want: false},
		{name: "empty", code: "", want: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := IsRetryableSQLState(test.code); got != test.want {
				t.Fatalf("expected retryable=%t, got %t", test.want, got)
			}
		})
	}
}

// TestIsConstraintSQLState verifies stable constraint SQLSTATE handling.
func TestIsConstraintSQLState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
		want bool
	}{
		{name: "unique violation", code: ErrorUniqueViolation, want: true},
		{name: "check violation", code: ErrorCheckViolation, want: true},
		{name: "foreign key violation", code: ErrorForeignKeyViolation, want: true},
		{name: "not null violation", code: ErrorNotNullViolation, want: true},
		{name: "trimmed", code: " " + ErrorUniqueViolation + " ", want: true},
		{name: "serialization failure", code: ErrorSerializationFailure, want: false},
		{name: "empty", code: "", want: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := IsConstraintSQLState(test.code); got != test.want {
				t.Fatalf("expected constraint=%t, got %t", test.want, got)
			}
		})
	}
}

// TestErrorHelpersClassifyWrappedErrors verifies error-level helper functions
// use extracted SQLSTATE values instead of message text.
func TestErrorHelpersClassifyWrappedErrors(t *testing.T) {
	t.Parallel()

	if !IsRetryableWriteConflict(gerror.Wrap(fakeSQLStateError{state: ErrorDeadlockDetected}, "update failed")) {
		t.Fatal("expected wrapped deadlock SQLSTATE to be retryable")
	}
	if !IsConstraintViolation(gerror.Wrap(fakeSQLStateError{state: ErrorForeignKeyViolation}, "insert failed")) {
		t.Fatal("expected wrapped foreign-key SQLSTATE to be a constraint violation")
	}
	if !IsUniqueConstraintViolation(gerror.Wrap(fakeSQLStateError{state: ErrorUniqueViolation}, "insert failed")) {
		t.Fatal("expected wrapped unique SQLSTATE to be a unique constraint violation")
	}
	if IsConstraintViolation(gerror.Wrap(fakeSQLStateError{state: ErrorLockNotAvailable}, "lock failed")) {
		t.Fatal("expected lock-not-available SQLSTATE not to be a constraint violation")
	}
	if IsUniqueConstraintViolation(gerror.Wrap(fakeSQLStateError{state: ErrorNotNullViolation}, "insert failed")) {
		t.Fatal("expected not-null SQLSTATE not to be a unique constraint violation")
	}
}
