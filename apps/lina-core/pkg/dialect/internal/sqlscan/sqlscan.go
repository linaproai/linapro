// Package sqlscan provides SQL scanning helpers shared by database dialects.
package sqlscan

import "strings"

// SplitStatements parses one SQL file content into ordered executable
// statements while ignoring semicolons inside strings and comments.
func SplitStatements(content string) []string {
	var (
		statements      []string
		builder         strings.Builder
		inSingleQuote   bool
		inDoubleQuote   bool
		inBacktickQuote bool
		inDollarQuote   bool
		inLineComment   bool
		inBlockComment  bool
		dollarQuoteTag  string
	)

	for i := 0; i < len(content); i++ {
		current := content[i]

		switch {
		case inDollarQuote:
			builder.WriteByte(current)
			if current == '$' && strings.HasPrefix(content[i:], dollarQuoteTag) {
				for offset := 1; offset < len(dollarQuoteTag); offset++ {
					builder.WriteByte(content[i+offset])
				}
				i += len(dollarQuoteTag) - 1
				inDollarQuote = false
				dollarQuoteTag = ""
			}
			continue
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
		case current == '$':
			if tag := readDollarQuoteTag(content, i); tag != "" {
				inDollarQuote = true
				dollarQuoteTag = tag
				builder.WriteString(tag)
				i += len(tag) - 1
				continue
			}
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

// isSQLWhitespace reports whether one byte should be treated as SQL whitespace.
func isSQLWhitespace(value byte) bool {
	switch value {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	default:
		return false
	}
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

// readDollarQuoteTag returns a PostgreSQL dollar-quote tag starting at index.
func readDollarQuoteTag(content string, index int) string {
	if index < 0 || index >= len(content) || content[index] != '$' {
		return ""
	}
	for cursor := index + 1; cursor < len(content); cursor++ {
		current := content[cursor]
		if current == '$' {
			return content[index : cursor+1]
		}
		if !isIdentifierByte(current) {
			return ""
		}
	}
	return ""
}

// appendSQLStatement appends one non-empty SQL statement.
func appendSQLStatement(statements *[]string, statement string) {
	trimmed := strings.TrimSpace(statement)
	if trimmed == "" {
		return
	}
	*statements = append(*statements, trimmed)
}

// isIdentifierByte reports whether one byte can be part of a SQL identifier.
func isIdentifierByte(value byte) bool {
	return (value >= 'a' && value <= 'z') ||
		(value >= 'A' && value <= 'Z') ||
		(value >= '0' && value <= '9') ||
		value == '_'
}
