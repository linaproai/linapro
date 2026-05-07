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
			if ConsumeQuotedEscape(content, &i, current, '\'') {
				builder.WriteByte(content[i])
				continue
			}
			if current == '\'' {
				inSingleQuote = false
			}
			continue
		case inDoubleQuote:
			builder.WriteByte(current)
			if ConsumeQuotedEscape(content, &i, current, '"') {
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

// SplitTopLevelComma splits a SQL fragment on commas not nested in parentheses
// and not inside string or identifier quotes.
func SplitTopLevelComma(content string) []string {
	var (
		parts          []string
		builder        strings.Builder
		depth          int
		inSingleQuote  bool
		inDoubleQuote  bool
		inBacktickName bool
	)
	for i := 0; i < len(content); i++ {
		current := content[i]
		switch {
		case inSingleQuote:
			builder.WriteByte(current)
			if ConsumeQuotedEscape(content, &i, current, '\'') {
				builder.WriteByte(content[i])
				continue
			}
			if current == '\'' {
				inSingleQuote = false
			}
			continue
		case inDoubleQuote:
			builder.WriteByte(current)
			if ConsumeQuotedEscape(content, &i, current, '"') {
				builder.WriteByte(content[i])
				continue
			}
			if current == '"' {
				inDoubleQuote = false
			}
			continue
		case inBacktickName:
			builder.WriteByte(current)
			if current == '`' {
				inBacktickName = false
			}
			continue
		case current == '\'':
			inSingleQuote = true
		case current == '"':
			inDoubleQuote = true
		case current == '`':
			inBacktickName = true
		case current == '(':
			depth++
		case current == ')':
			if depth > 0 {
				depth--
			}
		case current == ',' && depth == 0:
			appendSQLPart(&parts, builder.String())
			builder.Reset()
			continue
		}
		builder.WriteByte(current)
	}
	appendSQLPart(&parts, builder.String())
	return parts
}

// FindMatchingParen finds the closing paren for one opening paren index.
func FindMatchingParen(content string, openIndex int) int {
	if openIndex < 0 || openIndex >= len(content) || content[openIndex] != '(' {
		return -1
	}
	var (
		depth         int
		inSingleQuote bool
		inDoubleQuote bool
		inBacktick    bool
	)
	for i := openIndex; i < len(content); i++ {
		current := content[i]
		switch {
		case inSingleQuote:
			if ConsumeQuotedEscape(content, &i, current, '\'') {
				continue
			}
			if current == '\'' {
				inSingleQuote = false
			}
			continue
		case inDoubleQuote:
			if ConsumeQuotedEscape(content, &i, current, '"') {
				continue
			}
			if current == '"' {
				inDoubleQuote = false
			}
			continue
		case inBacktick:
			if current == '`' {
				inBacktick = false
			}
			continue
		case current == '\'':
			inSingleQuote = true
		case current == '"':
			inDoubleQuote = true
		case current == '`':
			inBacktick = true
		case current == '(':
			depth++
		case current == ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// ConsumeQuotedEscape advances the parser across one escaped quote or escaped
// character inside a quoted SQL literal.
func ConsumeQuotedEscape(content string, index *int, current byte, quote byte) bool {
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

// ConsumeSQLString returns the index after one single-quoted SQL string.
func ConsumeSQLString(content string, start int) int {
	for i := start + 1; i < len(content); i++ {
		current := content[i]
		if ConsumeQuotedEscape(content, &i, current, '\'') {
			continue
		}
		if current == '\'' {
			return i + 1
		}
	}
	return -1
}

// IsSQLWhitespace reports whether one byte should be treated as SQL whitespace.
func IsSQLWhitespace(value byte) bool {
	switch value {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	default:
		return false
	}
}

// IsKeywordAt reports whether keyword starts at index on identifier boundaries.
func IsKeywordAt(content string, index int, keyword string) bool {
	return KeywordLengthAt(content, index, keyword) > 0
}

// KeywordLengthAt returns the matched keyword length when keyword starts at
// index on identifier boundaries, otherwise 0.
func KeywordLengthAt(content string, index int, keyword string) int {
	if index < 0 || index+len(keyword) > len(content) {
		return 0
	}
	if strings.ToUpper(content[index:index+len(keyword)]) != keyword {
		return 0
	}
	if index > 0 && isIdentifierByte(content[index-1]) {
		return 0
	}
	nextIndex := index + len(keyword)
	if nextIndex < len(content) && isIdentifierByte(content[nextIndex]) {
		return 0
	}
	return len(keyword)
}

// LineNumberForStatement estimates a statement's 1-based line number.
func LineNumberForStatement(content string, statement string) int {
	index := strings.Index(content, statement)
	if index < 0 {
		return 1
	}
	return strings.Count(content[:index], "\n") + 1
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
	return IsSQLWhitespace(content[index+2])
}

// appendSQLStatement appends one non-empty SQL statement.
func appendSQLStatement(statements *[]string, statement string) {
	trimmed := strings.TrimSpace(statement)
	if trimmed == "" {
		return
	}
	*statements = append(*statements, trimmed)
}

// appendSQLPart appends one non-empty comma-delimited SQL fragment.
func appendSQLPart(parts *[]string, part string) {
	trimmed := strings.TrimSpace(part)
	if trimmed == "" {
		return
	}
	*parts = append(*parts, trimmed)
}

// isIdentifierByte reports whether one byte can be part of a SQL identifier.
func isIdentifierByte(value byte) bool {
	return (value >= 'a' && value <= 'z') ||
		(value >= 'A' && value <= 'Z') ||
		(value >= '0' && value <= '9') ||
		value == '_'
}
