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
	errorUniqueViolation      = "23505"
	errorSerializationFailure = "40001"
	errorDeadlockDetected     = "40P01"
	errorLockNotAvailable     = "55P03"
	errorCheckViolation       = "23514"
	errorForeignKeyViolation  = "23503"
	errorNotNullViolation     = "23502"
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
	return isRetryableSQLState(sqlState(err))
}

// isRetryableSQLState reports whether a PostgreSQL SQLSTATE represents a
// transient write conflict that a caller may retry.
func isRetryableSQLState(code string) bool {
	switch strings.TrimSpace(code) {
	case errorSerializationFailure, errorDeadlockDetected, errorLockNotAvailable:
		return true
	default:
		return false
	}
}

// isConstraintViolation classifies PostgreSQL constraint failures by stable
// SQLSTATE code.
func isConstraintViolation(err error) bool {
	return isConstraintSQLState(sqlState(err))
}

// IsUniqueConstraintViolation reports whether err is a PostgreSQL unique-key
// conflict.
func IsUniqueConstraintViolation(err error) bool {
	return strings.TrimSpace(sqlState(err)) == errorUniqueViolation
}

// isConstraintSQLState reports whether a PostgreSQL SQLSTATE represents one of
// the constraint violations the dialect layer needs to classify.
func isConstraintSQLState(code string) bool {
	switch strings.TrimSpace(code) {
	case errorUniqueViolation, errorCheckViolation, errorForeignKeyViolation, errorNotNullViolation:
		return true
	default:
		return false
	}
}

// sqlState extracts one PostgreSQL SQLSTATE code from err when available.
func sqlState(err error) string {
	if err == nil {
		return ""
	}
	var stateErr sqlStateError
	if errors.As(err, &stateErr) {
		return strings.TrimSpace(stateErr.SQLState())
	}
	return ""
}
