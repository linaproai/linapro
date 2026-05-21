// This file tests PostgreSQL read-only SQL classification used by governance
// layers that must avoid depending on PostgreSQL catalog syntax directly.

package postgres

import "testing"

// TestClassifyReadSQL verifies PostgreSQL catalog lookup and schema probe SQL
// are classified without requiring a live database.
func TestClassifyReadSQL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		sql                string
		wantMetadataLookup bool
		wantSchemaProbe    bool
	}{
		{
			name:               "information schema metadata lookup",
			sql:                `SELECT column_name FROM information_schema.columns WHERE table_name = ?`,
			wantMetadataLookup: true,
		},
		{
			name:               "pg catalog metadata lookup",
			sql:                `SELECT a.attname FROM pg_catalog.pg_attribute a WHERE a.attrelid = ?`,
			wantMetadataLookup: true,
		},
		{
			name:            "current schema probe",
			sql:             `SELECT current_schema()`,
			wantSchemaProbe: true,
		},
		{
			name:               "table listing schema probe",
			sql:                `SELECT c.relname FROM pg_class c INNER JOIN pg_namespace n ON c.relnamespace = n.oid WHERE c.relkind IN ('r', 'p') ORDER BY c.relname`,
			wantMetadataLookup: true,
			wantSchemaProbe:    true,
		},
		{
			name: "application data read",
			sql:  `SELECT * FROM plugin_demo_record WHERE id = ?`,
		},
		{
			name: "application read containing catalog subquery is not metadata lookup",
			sql:  `SELECT * FROM sys_plugin WHERE id IN (SELECT id FROM pg_class WHERE relname = ?)`,
		},
		{
			name: "application read containing current schema is not schema probe",
			sql:  `SELECT id, current_schema() FROM sys_plugin`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			classification := ClassifyReadSQL(test.sql)
			if classification.MetadataLookup != test.wantMetadataLookup {
				t.Fatalf("expected MetadataLookup=%t, got %t", test.wantMetadataLookup, classification.MetadataLookup)
			}
			if classification.SchemaProbe != test.wantSchemaProbe {
				t.Fatalf("expected SchemaProbe=%t, got %t", test.wantSchemaProbe, classification.SchemaProbe)
			}
		})
	}
}
