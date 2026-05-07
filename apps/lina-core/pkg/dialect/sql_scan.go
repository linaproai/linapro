// This file provides SQL scanning helpers shared by dialect translators.

package dialect

import "lina-core/pkg/dialect/internal/sqlscan"

// SplitSQLStatements parses one SQL file content into ordered executable
// statements while ignoring semicolons inside strings and comments.
func SplitSQLStatements(content string) []string {
	return sqlscan.SplitStatements(content)
}
