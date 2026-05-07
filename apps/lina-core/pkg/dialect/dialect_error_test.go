// This file tests dialect-level database driver error classification.

package dialect

import (
	"errors"
	"testing"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/gogf/gf/v2/errors/gerror"
)

// Retryable test error codes reported by the MySQL driver.
const (
	testMySQLDeadlock        uint16 = 1213
	testMySQLLockWaitTimeout uint16 = 1205
)

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

// TestIsRetryableWriteConflictClassifiesMySQLErrors verifies MySQL retryable
// lock conflicts are recognized by driver error number rather than text.
func TestIsRetryableWriteConflictClassifiesMySQLErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "deadlock",
			err:  &mysql.MySQLError{Number: testMySQLDeadlock, Message: "deadlock found"},
			want: true,
		},
		{
			name: "lock wait timeout wrapped by goframe",
			err: gerror.Wrap(
				&mysql.MySQLError{Number: testMySQLLockWaitTimeout, Message: "lock wait timeout exceeded"},
				"update failed",
			),
			want: true,
		},
		{
			name: "duplicate key is not retryable",
			err:  &mysql.MySQLError{Number: 1062, Message: "duplicate entry"},
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
