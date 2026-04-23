// This file splits development SQL assets into executable statements so local
// init/mock flows no longer depend on driver-level multi-statement execution.

package cmd

import "strings"

// splitSQLStatements parses one SQL file content into ordered executable
// statements while ignoring semicolons inside strings and comments.
func splitSQLStatements(content string) []string {
	var (
		statements      []string
		builder         strings.Builder
		inSingleQuote   bool
		inDoubleQuote   bool
		inBacktickQuote bool
		inLineComment   bool
		inBlockComment  bool
	)

	for i := 0; i < len(content); i++ {
		current := content[i]

		switch {
		case inLineComment:
			if current == '\n' {
				inLineComment = false
				builder.WriteByte('\n')
			}
			continue

		case inBlockComment:
			if current == '*' && i+1 < len(content) && content[i+1] == '/' {
				inBlockComment = false
				i++
				builder.WriteByte(' ')
			}
			continue

		case inSingleQuote:
			builder.WriteByte(current)
			if consumeQuotedEscape(content, &i, current, '\'') {
				builder.WriteByte(content[i])
				continue
			}
			if current == '\'' {
				inSingleQuote = false
			}
			continue

		case inDoubleQuote:
			builder.WriteByte(current)
			if consumeQuotedEscape(content, &i, current, '"') {
				builder.WriteByte(content[i])
				continue
			}
			if current == '"' {
				inDoubleQuote = false
			}
			continue

		case inBacktickQuote:
			builder.WriteByte(current)
			if current == '`' && i+1 < len(content) && content[i+1] == '`' {
				i++
				builder.WriteByte(content[i])
				continue
			}
			if current == '`' {
				inBacktickQuote = false
			}
			continue

		case isLineCommentStart(content, i):
			inLineComment = true
			i++
			continue

		case current == '#':
			inLineComment = true
			continue

		case current == '/' && i+1 < len(content) && content[i+1] == '*':
			inBlockComment = true
			i++
			continue

		case current == ';':
			appendSQLStatement(&statements, builder.String())
			builder.Reset()
			continue

		case current == '\'':
			inSingleQuote = true

		case current == '"':
			inDoubleQuote = true

		case current == '`':
			inBacktickQuote = true
		}

		builder.WriteByte(current)
	}

	appendSQLStatement(&statements, builder.String())
	return statements
}

// consumeQuotedEscape advances the parser across one escaped quote or escaped
// character inside a quoted SQL literal.
func consumeQuotedEscape(content string, index *int, current byte, quote byte) bool {
	if current == '\\' && *index+1 < len(content) {
		*index = *index + 1
		return true
	}
	if current == quote && *index+1 < len(content) && content[*index+1] == quote {
		*index = *index + 1
		return true
	}
	return false
}

// isLineCommentStart reports whether the current index begins a MySQL-style
// double-dash line comment that requires trailing whitespace or EOF.
func isLineCommentStart(content string, index int) bool {
	if index+1 >= len(content) || content[index] != '-' || content[index+1] != '-' {
		return false
	}
	if index+2 >= len(content) {
		return true
	}
	return isSQLWhitespace(content[index+2])
}

// isSQLWhitespace reports whether one byte should be treated as SQL whitespace
// when checking comment boundaries.
func isSQLWhitespace(value byte) bool {
	switch value {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	default:
		return false
	}
}

// appendSQLStatement appends one non-empty SQL statement after trimming leading
// and trailing whitespace introduced by parsing.
func appendSQLStatement(statements *[]string, statement string) {
	trimmed := strings.TrimSpace(statement)
	if trimmed == "" {
		return
	}
	*statements = append(*statements, trimmed)
}
