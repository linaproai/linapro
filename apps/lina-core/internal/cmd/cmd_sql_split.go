// This file splits development SQL assets into executable statements so local
// init/mock flows no longer depend on driver-level multi-statement execution.

package cmd

import "lina-core/pkg/dialect"

// splitSQLStatements parses one SQL file content into ordered executable
// statements while ignoring semicolons inside strings and comments.
func splitSQLStatements(content string) []string {
	return dialect.SplitSQLStatements(content)
}
