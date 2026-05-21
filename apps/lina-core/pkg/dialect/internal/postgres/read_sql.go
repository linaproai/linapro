// This file classifies PostgreSQL driver and ORM read-only SQL that belongs to
// catalog inspection or connection probing rather than application data access.

package postgres

import "strings"

// ReadSQLClassification describes PostgreSQL read-only SQL emitted by database
// drivers or ORM internals.
type ReadSQLClassification struct {
	MetadataLookup bool
	SchemaProbe    bool
}

// ClassifyReadSQL classifies PostgreSQL catalog lookup and schema probe SQL.
func ClassifyReadSQL(sql string) ReadSQLClassification {
	normalizedSQL := strings.ToLower(strings.TrimSpace(sql))
	return ReadSQLClassification{
		MetadataLookup: postgresSQLReadsMetadata(normalizedSQL) && !postgresSQLReadsApplicationTable(normalizedSQL),
		SchemaProbe:    postgresSQLReadsSchemaProbe(normalizedSQL),
	}
}

// postgresSQLReadsMetadata identifies metadata queries that refer to the target
// table through bind arguments rather than literal SQL text.
func postgresSQLReadsMetadata(sql string) bool {
	return strings.Contains(sql, "from information_schema.") ||
		strings.Contains(sql, "join information_schema.") ||
		strings.Contains(sql, "from pg_catalog.") ||
		strings.Contains(sql, "join pg_catalog.") ||
		strings.Contains(sql, "from pg_") ||
		strings.Contains(sql, "join pg_")
}

// postgresSQLReadsSchemaProbe identifies PostgreSQL schema and version probes
// emitted before table-specific metadata queries run.
func postgresSQLReadsSchemaProbe(sql string) bool {
	return postgresSQLMatchesProbeExpression(sql, "current_schema()") ||
		postgresSQLMatchesProbeExpression(sql, "version()") ||
		(postgresSQLReadsCatalogOnly(sql) &&
			strings.Contains(sql, "from pg_class") &&
			strings.Contains(sql, "pg_namespace") &&
			strings.Contains(sql, "relname"))
}

// postgresSQLMatchesProbeExpression reports whether SQL is a simple SELECT of
// one connection/session probe expression, with no FROM clause.
func postgresSQLMatchesProbeExpression(sql string, expression string) bool {
	return strings.HasPrefix(sql, "select ") &&
		strings.Contains(sql, expression) &&
		!strings.Contains(sql, " from ")
}

// postgresSQLReadsCatalogOnly reports whether every FROM/JOIN target in SQL is
// a PostgreSQL catalog or information_schema relation.
func postgresSQLReadsCatalogOnly(sql string) bool {
	tables := postgresSQLTableReferences(sql)
	if len(tables) == 0 {
		return false
	}
	for _, table := range tables {
		if !postgresSQLIsCatalogTable(table) {
			return false
		}
	}
	return true
}

// postgresSQLReadsApplicationTable reports whether SQL references a table
// outside PostgreSQL catalog and information_schema relations.
func postgresSQLReadsApplicationTable(sql string) bool {
	for _, table := range postgresSQLTableReferences(sql) {
		if !postgresSQLIsCatalogTable(table) {
			return true
		}
	}
	return false
}

// postgresSQLTableReferences extracts lightweight FROM/JOIN table tokens used
// only for classifying driver-emitted read SQL.
func postgresSQLTableReferences(sql string) []string {
	tokens := strings.FieldsFunc(sql, func(r rune) bool {
		return r == ' ' ||
			r == '\t' ||
			r == '\n' ||
			r == '\r' ||
			r == ',' ||
			r == '(' ||
			r == ')' ||
			r == ';'
	})
	references := make([]string, 0)
	for index, token := range tokens {
		if token != "from" && token != "join" {
			continue
		}
		if index+1 >= len(tokens) {
			continue
		}
		references = append(references, normalizePostgresSQLTableReference(tokens[index+1]))
	}
	return references
}

// normalizePostgresSQLTableReference trims SQL identifier quoting from a table
// token while preserving schema qualification.
func normalizePostgresSQLTableReference(table string) string {
	return strings.Trim(strings.TrimSpace(table), "`\"[]")
}

// postgresSQLIsCatalogTable reports whether a table token belongs to
// PostgreSQL catalog or information_schema metadata.
func postgresSQLIsCatalogTable(table string) bool {
	normalized := normalizePostgresSQLTableReference(table)
	return strings.HasPrefix(normalized, "information_schema.") ||
		strings.HasPrefix(normalized, "pg_catalog.") ||
		strings.HasPrefix(normalized, "pg_")
}
