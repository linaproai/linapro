// error.go classifies PostgreSQL driver errors for shared dialect callers. It
// centralizes SQLSTATE handling so retry and constraint decisions remain stable
// across host services and plugin database helpers.

package postgres

import (
	"errors"
	"strings"
)

// PostgreSQL SQLSTATE codes relevant to LinaPro's write paths.
const (
	ErrorUniqueViolation      = "23505"
	ErrorSerializationFailure = "40001"
	ErrorDeadlockDetected     = "40P01"
	ErrorLockNotAvailable     = "55P03"
	ErrorCheckViolation       = "23514"
	ErrorForeignKeyViolation  = "23503"
	ErrorNotNullViolation     = "23502"
)

// sqlStateError is the narrow SQLSTATE shape exposed by PostgreSQL drivers.
type sqlStateError interface {
	error
	// SQLState returns the PostgreSQL SQLSTATE code used to classify retryable
	// write conflicts and constraint failures.
	SQLState() string
}

// IsRetryableWriteConflict classifies PostgreSQL transient write conflicts by
// stable SQLSTATE code.
func IsRetryableWriteConflict(err error) bool {
	return IsRetryableSQLState(SQLState(err))
}

// IsRetryableSQLState reports whether a PostgreSQL SQLSTATE represents a
// transient write conflict that a caller may retry.
func IsRetryableSQLState(code string) bool {
	switch strings.TrimSpace(code) {
	case ErrorSerializationFailure, ErrorDeadlockDetected, ErrorLockNotAvailable:
		return true
	default:
		return false
	}
}

// IsConstraintViolation classifies PostgreSQL constraint failures by stable
// SQLSTATE code.
func IsConstraintViolation(err error) bool {
	return IsConstraintSQLState(SQLState(err))
}

// IsUniqueConstraintViolation reports whether err is a PostgreSQL unique-key
// conflict.
func IsUniqueConstraintViolation(err error) bool {
	return strings.TrimSpace(SQLState(err)) == ErrorUniqueViolation
}

// IsConstraintSQLState reports whether a PostgreSQL SQLSTATE represents one of
// the constraint violations the dialect layer needs to classify.
func IsConstraintSQLState(code string) bool {
	switch strings.TrimSpace(code) {
	case ErrorUniqueViolation, ErrorCheckViolation, ErrorForeignKeyViolation, ErrorNotNullViolation:
		return true
	default:
		return false
	}
}

// SQLState extracts one PostgreSQL SQLSTATE code from err when available.
func SQLState(err error) string {
	if err == nil {
		return ""
	}
	var stateErr sqlStateError
	if errors.As(err, &stateErr) {
		return strings.TrimSpace(stateErr.SQLState())
	}
	return ""
}
