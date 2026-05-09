// This file classifies SQLite driver errors for shared dialect callers.

package sqlite

import "errors"

// SQLite primary result codes used for retryable write conflicts.
const (
	primaryErrorMask      = 0xff
	errorBusy             = 5
	errorLocked           = 6
	errorConstraint       = 19
	errorConstraintPK     = errorConstraint | (6 << 8)
	errorConstraintUnique = errorConstraint | (8 << 8)
)

// codeError is the common narrow shape exposed by modernc/glebarez SQLite
// errors. Keeping this as an interface avoids binding callers to a specific
// SQLite driver implementation.
type codeError interface {
	error
	Code() int
}

// IsRetryableWriteConflict classifies SQLite lock conflicts by primary result
// code. Extended SQLite codes preserve the primary code in the low byte.
func IsRetryableWriteConflict(err error) bool {
	var sqliteErr codeError
	if !errors.As(err, &sqliteErr) {
		return false
	}
	return IsRetryablePrimaryCode(sqliteErr.Code())
}

// IsRetryablePrimaryCode reports whether an SQLite result code indicates the
// database or table is temporarily locked.
func IsRetryablePrimaryCode(code int) bool {
	switch code & primaryErrorMask {
	case errorBusy, errorLocked:
		return true
	default:
		return false
	}
}

// IsUniqueConstraintViolation reports whether err is a SQLite unique-key or
// primary-key conflict. A plain primary constraint code is intentionally not
// classified because it may represent non-unique constraint failures.
func IsUniqueConstraintViolation(err error) bool {
	var sqliteErr codeError
	if !errors.As(err, &sqliteErr) {
		return false
	}
	switch sqliteErr.Code() {
	case errorConstraintPK, errorConstraintUnique:
		return true
	default:
		return false
	}
}
