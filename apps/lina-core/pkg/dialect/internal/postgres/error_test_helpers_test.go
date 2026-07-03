// This file keeps PostgreSQL constraint classification helpers scoped to tests.

package postgres

import "strings"

// isConstraintViolation classifies PostgreSQL constraint failures by stable SQLSTATE code.
func isConstraintViolation(err error) bool {
	return isConstraintSQLState(sqlState(err))
}

// isConstraintSQLState reports whether a PostgreSQL SQLSTATE is a constraint violation.
func isConstraintSQLState(code string) bool {
	switch strings.TrimSpace(code) {
	case errorUniqueViolation, errorCheckViolation, errorForeignKeyViolation, errorNotNullViolation:
		return true
	default:
		return false
	}
}
