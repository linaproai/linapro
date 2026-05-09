// This file classifies database driver errors that require dialect-specific
// interpretation but are useful to shared host services.

package dialect

import (
	internalpostgres "lina-core/pkg/dialect/internal/postgres"
	internalsqlite "lina-core/pkg/dialect/internal/sqlite"
)

// IsRetryableWriteConflict reports whether err represents a transient database
// write conflict that a compare-and-swap write path may safely retry.
func IsRetryableWriteConflict(err error) bool {
	if err == nil {
		return false
	}
	return internalpostgres.IsRetryableWriteConflict(err) || internalsqlite.IsRetryableWriteConflict(err)
}

// IsUniqueConstraintViolation reports whether err represents a database
// unique-key conflict.
func IsUniqueConstraintViolation(err error) bool {
	if err == nil {
		return false
	}
	return internalpostgres.IsUniqueConstraintViolation(err) || internalsqlite.IsUniqueConstraintViolation(err)
}
