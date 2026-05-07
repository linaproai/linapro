// This file translates the project's controlled MySQL SQL asset subset into
// SQLite-compatible SQL.

package sqlite

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/dialect/internal/sqlscan"
)

var (
	reCreateDatabase      = regexp.MustCompile(`(?is)^\s*CREATE\s+DATABASE\b`)
	reUseDatabase         = regexp.MustCompile(`(?is)^\s*USE\s+`)
	reCreateTable         = regexp.MustCompile(`(?is)^\s*CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([^\s(]+)\s*\(`)
	reUnsupportedKeywords = regexp.MustCompile(`(?is)\b(FULLTEXT|SPATIAL|GENERATED\s+ALWAYS\s+AS|PARTITION|ON\s+DUPLICATE\s+KEY\s+UPDATE|FIND_IN_SET|GROUP_CONCAT|IF\s*\()\b`)
	reIntegerAuto         = regexp.MustCompile(`(?i)\b(?:BIGINT|INT)\s+(?:UNSIGNED\s+)?(?:NOT\s+NULL\s+)?(?:PRIMARY\s+KEY\s+)?AUTO_INCREMENT(?:\s+PRIMARY\s+KEY)?\b`)
	reIntegerAutoAlt      = regexp.MustCompile(`(?i)\b(?:BIGINT|INT)\s+(?:UNSIGNED\s+)?PRIMARY\s+KEY\s+AUTO_INCREMENT\b`)
	reIntegerTypes        = regexp.MustCompile(`(?i)\b(?:BIGINT|INT|TINYINT|SMALLINT)(?:\s+UNSIGNED)?\b`)
	reTextTypes           = regexp.MustCompile(`(?i)\b(?:VARCHAR|CHAR)\s*\(\s*\d+\s*\)|\b(?:LONGTEXT|MEDIUMTEXT)\b`)
	reBlobTypes           = regexp.MustCompile(`(?i)\b(?:VARBINARY|BINARY)\s*\(\s*\d+\s*\)|\b(?:BLOB|LONGBLOB|MEDIUMBLOB)\b`)
	reDecimalTypes        = regexp.MustCompile(`(?i)\bDECIMAL\s*\(\s*\d+\s*,\s*\d+\s*\)`)
	reInsertIgnore        = regexp.MustCompile(`(?i)\bINSERT\s+IGNORE\s+INTO\b`)
	reOnUpdate            = regexp.MustCompile(`(?i)\s+ON\s+UPDATE\s+CURRENT_TIMESTAMP`)
	reFromDual            = regexp.MustCompile(`(?i)\s+FROM\s+DUAL\b`)
	reNowFunc             = regexp.MustCompile(`(?i)\bNOW\s*\(\s*\)`)
	reDefaultCharset      = regexp.MustCompile(`(?i)(^|\s+)DEFAULT\s+CHARSET\s*=\s*[A-Za-z0-9_]+`)
	reCharset             = regexp.MustCompile(`(?i)(^|\s+)CHARSET\s*=\s*[A-Za-z0-9_]+`)
	reCollate             = regexp.MustCompile(`(?i)(^|\s+)COLLATE\s*=\s*[A-Za-z0-9_]+`)
	reEngine              = regexp.MustCompile(`(?i)(^|\s+)ENGINE\s*=\s*[A-Za-z0-9_]+`)
	reTableCommentEquals  = regexp.MustCompile(`(?is)(^|\s+)COMMENT\s*=\s*'[^']*(?:''[^']*)*'`)
	reTableCommentSpace   = regexp.MustCompile(`(?is)(^|\s+)COMMENT\s+'[^']*(?:''[^']*)*'`)
)

// translateDDL translates a full SQL asset into SQLite-compatible SQL.
func translateDDL(sourceName string, ddl string) (string, error) {
	var output []string
	for _, statement := range sqlscan.SplitStatements(ddl) {
		if err := rejectUnsupportedSQL(sourceName, ddl, statement); err != nil {
			return "", err
		}
		translated, err := translateSQLiteStatement(statement)
		if err != nil {
			return "", gerror.Wrapf(err, "translate SQL asset %s failed", sourceName)
		}
		if translated == "" {
			continue
		}
		output = append(output, translated)
	}
	if len(output) == 0 {
		return "", nil
	}
	return strings.Join(output, "\n") + "\n", nil
}

// rejectUnsupportedSQL fails fast for MySQL features outside the supported
// project SQL subset.
func rejectUnsupportedSQL(sourceName string, ddl string, statement string) error {
	match := reUnsupportedKeywords.FindString(statement)
	if match == "" {
		return nil
	}
	line := sqlscan.LineNumberForStatement(ddl, statement)
	return gerror.Newf("unsupported MySQL syntax in %s at line %d: %s", sourceName, line, strings.TrimSpace(match))
}

// translateSQLiteStatement translates one SQL statement.
func translateSQLiteStatement(statement string) (string, error) {
	trimmed := strings.TrimSpace(statement)
	switch {
	case trimmed == "":
		return "", nil
	case reCreateDatabase.MatchString(trimmed), reUseDatabase.MatchString(trimmed):
		return "", nil
	case reCreateTable.MatchString(trimmed):
		return translateCreateTable(trimmed)
	default:
		return translateGeneralSQL(trimmed) + ";", nil
	}
}

// translateCreateTable translates one MySQL CREATE TABLE statement and extracts
// inline indexes into standalone CREATE INDEX statements.
func translateCreateTable(statement string) (string, error) {
	openIndex := strings.Index(statement, "(")
	closeIndex := sqlscan.FindMatchingParen(statement, openIndex)
	if openIndex < 0 || closeIndex < 0 {
		return "", gerror.New("CREATE TABLE statement has invalid parentheses")
	}

	header := strings.TrimSpace(statement[:openIndex])
	body := statement[openIndex+1 : closeIndex]
	suffix := strings.TrimSpace(statement[closeIndex+1:])
	tableName := normalizeIdentifier(reCreateTable.FindStringSubmatch(statement)[1])
	items := sqlscan.SplitTopLevelComma(body)

	var (
		columns        []string
		tablePrimaryID string
		indexes        []string
	)
	for _, item := range items {
		translatedItem, extractedIndex, primaryID := translateCreateTableItem(tableName, item)
		if primaryID != "" {
			tablePrimaryID = primaryID
			continue
		}
		if extractedIndex != "" {
			indexes = append(indexes, extractedIndex)
			continue
		}
		if translatedItem != "" {
			columns = append(columns, translatedItem)
		}
	}
	if tablePrimaryID != "" {
		primaryKeyCoveredByColumn := false
		for index, column := range columns {
			if !equalIdentifier(columnName(column), tablePrimaryID) {
				continue
			}
			if isIntegerAutoColumn(column) {
				columns[index] = normalizeAutoIncrementColumn(column)
				primaryKeyCoveredByColumn = true
				break
			}
			if strings.Contains(strings.ToUpper(column), "PRIMARY KEY") {
				primaryKeyCoveredByColumn = true
				break
			}
		}
		if !primaryKeyCoveredByColumn {
			columns = append(columns, fmt.Sprintf("PRIMARY KEY (%s)", tablePrimaryID))
		}
	}

	createSQL := stripBacktickIdentifiers(strings.TrimSpace(header)) + " (\n    " + strings.Join(columns, ",\n    ") + "\n)"
	createSQL += translateTableSuffix(suffix) + ";"
	if len(indexes) == 0 {
		return createSQL, nil
	}
	return createSQL + "\n" + strings.Join(indexes, "\n"), nil
}

// translateCreateTableItem converts one comma-delimited CREATE TABLE item.
func translateCreateTableItem(tableName string, item string) (translated string, indexSQL string, tablePrimaryID string) {
	trimmed := strings.TrimSpace(strings.TrimSuffix(item, ","))
	if trimmed == "" {
		return "", "", ""
	}
	upper := strings.ToUpper(trimmed)
	if strings.HasPrefix(upper, "PRIMARY KEY") {
		cols := extractIndexColumns(trimmed)
		if len(cols) == 1 {
			return "", "", normalizeIdentifier(cols[0])
		}
		return translateGeneralSQL(trimmed), "", ""
	}
	if strings.HasPrefix(upper, "KEY ") ||
		strings.HasPrefix(upper, "INDEX ") ||
		strings.HasPrefix(upper, "UNIQUE KEY ") ||
		strings.HasPrefix(upper, "UNIQUE INDEX ") {
		return "", buildCreateIndexSQL(tableName, trimmed), ""
	}
	return translateColumnDefinition(trimmed), "", ""
}

// translateColumnDefinition converts one column definition.
func translateColumnDefinition(column string) string {
	withoutComment := stripColumnComment(column)
	if isIntegerAutoColumn(withoutComment) {
		return normalizeAutoIncrementColumn(withoutComment)
	}
	return translateGeneralSQL(withoutComment)
}

// translateGeneralSQL performs token-level conversion outside CREATE TABLE structure.
func translateGeneralSQL(sql string) string {
	translated := stripBacktickIdentifiers(sql)
	translated = rewriteSQLiteFunctionCalls(translated)
	translated = reInsertIgnore.ReplaceAllString(translated, "INSERT OR IGNORE INTO")
	translated = reOnUpdate.ReplaceAllString(translated, "")
	translated = reIntegerAuto.ReplaceAllString(translated, "INTEGER PRIMARY KEY AUTOINCREMENT")
	translated = reIntegerAutoAlt.ReplaceAllString(translated, "INTEGER PRIMARY KEY AUTOINCREMENT")
	translated = reIntegerTypes.ReplaceAllString(translated, "INTEGER")
	translated = reBlobTypes.ReplaceAllString(translated, "BLOB")
	translated = reTextTypes.ReplaceAllString(translated, "TEXT")
	translated = reDecimalTypes.ReplaceAllString(translated, "NUMERIC")
	translated = reFromDual.ReplaceAllString(translated, "")
	translated = reNowFunc.ReplaceAllString(translated, "CURRENT_TIMESTAMP")
	translated = stripColumnComment(translated)
	return strings.TrimSpace(translated)
}

// rewriteSQLiteFunctionCalls rewrites MySQL functions that appear in project SQL
// assets but are not available under SQLite with the same name.
func rewriteSQLiteFunctionCalls(sql string) string {
	return rewriteFunctionCall(sql, "CONCAT", rewriteConcatCall)
}

// rewriteFunctionCall rewrites calls to one SQL function outside string and
// identifier quotes while preserving unrelated SQL text byte-for-byte.
func rewriteFunctionCall(
	sql string,
	functionName string,
	rewrite func(args []string) string,
) string {
	var (
		builder       strings.Builder
		inSingleQuote bool
		inDoubleQuote bool
		inBacktick    bool
		upperName     = strings.ToUpper(functionName)
	)
	for i := 0; i < len(sql); i++ {
		current := sql[i]
		switch {
		case inSingleQuote:
			builder.WriteByte(current)
			if sqlscan.ConsumeQuotedEscape(sql, &i, current, '\'') {
				builder.WriteByte(sql[i])
				continue
			}
			if current == '\'' {
				inSingleQuote = false
			}
		case inDoubleQuote:
			builder.WriteByte(current)
			if sqlscan.ConsumeQuotedEscape(sql, &i, current, '"') {
				builder.WriteByte(sql[i])
				continue
			}
			if current == '"' {
				inDoubleQuote = false
			}
		case inBacktick:
			builder.WriteByte(current)
			if current == '`' {
				inBacktick = false
			}
		case current == '\'':
			inSingleQuote = true
			builder.WriteByte(current)
		case current == '"':
			inDoubleQuote = true
			builder.WriteByte(current)
		case current == '`':
			inBacktick = true
			builder.WriteByte(current)
		case sqlscan.IsKeywordAt(sql, i, upperName):
			openIndex := i + sqlscan.KeywordLengthAt(sql, i, upperName)
			for openIndex < len(sql) && sqlscan.IsSQLWhitespace(sql[openIndex]) {
				openIndex++
			}
			if openIndex >= len(sql) || sql[openIndex] != '(' {
				builder.WriteByte(current)
				continue
			}
			closeIndex := sqlscan.FindMatchingParen(sql, openIndex)
			if closeIndex < 0 {
				builder.WriteByte(current)
				continue
			}
			args := sqlscan.SplitTopLevelComma(sql[openIndex+1 : closeIndex])
			builder.WriteString(rewrite(args))
			i = closeIndex
		default:
			builder.WriteByte(current)
		}
	}
	return builder.String()
}

// rewriteConcatCall converts MySQL CONCAT(a, b, ...) to SQLite's string
// concatenation operator.
func rewriteConcatCall(args []string) string {
	rewritten := make([]string, 0, len(args))
	for _, arg := range args {
		rewritten = append(rewritten, rewriteSQLiteFunctionCalls(strings.TrimSpace(arg)))
	}
	return "(" + strings.Join(rewritten, " || ") + ")"
}

// translateTableSuffix removes MySQL table options after the closing paren.
func translateTableSuffix(suffix string) string {
	translated := reEngine.ReplaceAllString(suffix, "")
	translated = reDefaultCharset.ReplaceAllString(translated, "")
	translated = reCharset.ReplaceAllString(translated, "")
	translated = reCollate.ReplaceAllString(translated, "")
	translated = reTableCommentEquals.ReplaceAllString(translated, "")
	translated = reTableCommentSpace.ReplaceAllString(translated, "")
	return strings.TrimSpace(translated)
}

// buildCreateIndexSQL converts one inline KEY/INDEX clause into standalone SQL.
func buildCreateIndexSQL(tableName string, clause string) string {
	normalized := stripBacktickIdentifiers(strings.TrimSpace(clause))
	upper := strings.ToUpper(normalized)
	unique := false
	if strings.HasPrefix(upper, "UNIQUE KEY ") {
		unique = true
		normalized = strings.TrimSpace(normalized[len("UNIQUE KEY "):])
	} else if strings.HasPrefix(upper, "UNIQUE INDEX ") {
		unique = true
		normalized = strings.TrimSpace(normalized[len("UNIQUE INDEX "):])
	} else if strings.HasPrefix(upper, "KEY ") {
		normalized = strings.TrimSpace(normalized[len("KEY "):])
	} else if strings.HasPrefix(upper, "INDEX ") {
		normalized = strings.TrimSpace(normalized[len("INDEX "):])
	}

	openIndex := strings.Index(normalized, "(")
	indexName := strings.TrimSpace(normalized[:openIndex])
	columns := strings.TrimSpace(normalized[openIndex:])
	if unique {
		return fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s %s;", indexName, tableName, columns)
	}
	return fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s %s;", indexName, tableName, columns)
}

// extractIndexColumns returns the top-level comma-separated columns in one
// PRIMARY KEY clause.
func extractIndexColumns(clause string) []string {
	openIndex := strings.Index(clause, "(")
	closeIndex := sqlscan.FindMatchingParen(clause, openIndex)
	if openIndex < 0 || closeIndex < 0 {
		return nil
	}
	return sqlscan.SplitTopLevelComma(clause[openIndex+1 : closeIndex])
}

// isIntegerAutoColumn reports whether a column definition contains AUTO_INCREMENT.
func isIntegerAutoColumn(column string) bool {
	return strings.Contains(strings.ToUpper(column), "AUTO_INCREMENT")
}

// normalizeAutoIncrementColumn converts a column into SQLite integer primary key syntax.
func normalizeAutoIncrementColumn(column string) string {
	name := columnName(column)
	return name + " INTEGER PRIMARY KEY AUTOINCREMENT"
}

// columnName extracts the first identifier from one column definition.
func columnName(column string) string {
	fields := strings.Fields(stripBacktickIdentifiers(column))
	if len(fields) == 0 {
		return ""
	}
	return normalizeIdentifier(fields[0])
}

// normalizeIdentifier removes MySQL identifier quotes from a single identifier.
func normalizeIdentifier(identifier string) string {
	return strings.Trim(strings.TrimSpace(identifier), "`")
}

// equalIdentifier compares SQL identifiers case-insensitively after removing
// MySQL identifier quotes.
func equalIdentifier(left string, right string) bool {
	return strings.EqualFold(normalizeIdentifier(left), normalizeIdentifier(right))
}

// stripColumnComment removes a MySQL column-level COMMENT string.
func stripColumnComment(sql string) string {
	result := stripKeywordStringClause(sql, "COMMENT")
	return strings.TrimSpace(result)
}

// stripKeywordStringClause removes KEYWORD 'string' segments outside quoted strings.
func stripKeywordStringClause(sql string, keyword string) string {
	upperKeyword := strings.ToUpper(keyword)
	var builder strings.Builder
	for i := 0; i < len(sql); i++ {
		if sqlscan.IsKeywordAt(sql, i, upperKeyword) {
			next := i + len(keyword)
			for next < len(sql) && sqlscan.IsSQLWhitespace(sql[next]) {
				next++
			}
			if next < len(sql) && sql[next] == '\'' {
				end := sqlscan.ConsumeSQLString(sql, next)
				if end > next {
					i = end - 1
					continue
				}
			}
		}
		builder.WriteByte(sql[i])
	}
	return builder.String()
}

// stripBacktickIdentifiers removes MySQL backtick identifier quotes outside string literals.
func stripBacktickIdentifiers(sql string) string {
	var (
		builder       strings.Builder
		inSingleQuote bool
		inDoubleQuote bool
	)
	for i := 0; i < len(sql); i++ {
		current := sql[i]
		switch {
		case inSingleQuote:
			builder.WriteByte(current)
			if sqlscan.ConsumeQuotedEscape(sql, &i, current, '\'') {
				builder.WriteByte(sql[i])
				continue
			}
			if current == '\'' {
				inSingleQuote = false
			}
		case inDoubleQuote:
			builder.WriteByte(current)
			if sqlscan.ConsumeQuotedEscape(sql, &i, current, '"') {
				builder.WriteByte(sql[i])
				continue
			}
			if current == '"' {
				inDoubleQuote = false
			}
		case current == '\'':
			inSingleQuote = true
			builder.WriteByte(current)
		case current == '"':
			inDoubleQuote = true
			builder.WriteByte(current)
		case current == '`':
			continue
		default:
			builder.WriteByte(current)
		}
	}
	return builder.String()
}
