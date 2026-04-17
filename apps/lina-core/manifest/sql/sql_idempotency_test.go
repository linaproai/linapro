// This file guards the repository rule that root version SQL scripts must be
// safe to execute multiple times without destructive resets or duplicate-data
// errors.

package sqlscripts

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var (
	reCreateTable = regexp.MustCompile(`(?i)^\s*create\s+table\s+`)
	reInsertInto  = regexp.MustCompile(`(?i)^\s*insert\s+into\s+`)
	reDeleteFrom  = regexp.MustCompile(`(?i)^\s*delete\s+from\s+`)
	reAlterAutoID = regexp.MustCompile(`(?i)\balter\s+table\b.*\bauto_increment\s*=`)
)

func TestRootVersionSQLScriptsAreIdempotent(t *testing.T) {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	paths, err := filepath.Glob(filepath.Join(workingDir, "*.sql"))
	if err != nil {
		t.Fatalf("glob version sql files: %v", err)
	}
	sort.Strings(paths)

	var violations []string
	for _, path := range paths {
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("read %s: %v", path, readErr)
		}
		for index, statement := range splitSQLStatements(string(content)) {
			normalized := normalizeSQLStatement(statement)
			lowered := strings.ToLower(normalized)

			switch {
			case reCreateTable.MatchString(normalized):
				if !strings.Contains(lowered, "create table if not exists") {
					violations = append(
						violations,
						formatSQLViolation(path, index, "CREATE TABLE must use IF NOT EXISTS", normalized),
					)
				}

			case reInsertInto.MatchString(normalized):
				if !strings.Contains(lowered, "insert ignore into") &&
					!strings.Contains(lowered, "on duplicate key update") {
					violations = append(
						violations,
						formatSQLViolation(path, index, "INSERT INTO must use IGNORE or upsert", normalized),
					)
				}

			case reDeleteFrom.MatchString(normalized):
				if !strings.Contains(lowered, " where ") {
					violations = append(
						violations,
						formatSQLViolation(path, index, "unscoped DELETE FROM is not allowed in version SQL", normalized),
					)
				}
			}

			if reAlterAutoID.MatchString(normalized) {
				violations = append(
					violations,
					formatSQLViolation(path, index, "ALTER TABLE ... AUTO_INCREMENT reset is not allowed in version SQL", normalized),
				)
			}
		}
	}

	if len(violations) > 0 {
		t.Fatalf("non-idempotent root SQL statements found:\n%s", strings.Join(violations, "\n"))
	}
}

func splitSQLStatements(content string) []string {
	lines := strings.Split(content, "\n")
	builder := strings.Builder{}
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		builder.WriteString(line)
		builder.WriteByte('\n')
	}
	sanitized := builder.String()

	var (
		statements []string
		current    strings.Builder
		inQuote    bool
	)
	for i := 0; i < len(sanitized); i++ {
		ch := sanitized[i]
		if ch == '\'' {
			if inQuote && i+1 < len(sanitized) && sanitized[i+1] == '\'' {
				current.WriteByte(ch)
				current.WriteByte(sanitized[i+1])
				i++
				continue
			}
			inQuote = !inQuote
		}
		if ch == ';' && !inQuote {
			statement := strings.TrimSpace(current.String())
			if statement != "" {
				statements = append(statements, statement)
			}
			current.Reset()
			continue
		}
		current.WriteByte(ch)
	}
	if statement := strings.TrimSpace(current.String()); statement != "" {
		statements = append(statements, statement)
	}
	return statements
}

func normalizeSQLStatement(statement string) string {
	return strings.Join(strings.Fields(statement), " ")
}

func formatSQLViolation(path string, index int, message string, statement string) string {
	return filepath.Base(path) + ":" + strconv.Itoa(index+1) + ": " + message + " :: " + statement
}
