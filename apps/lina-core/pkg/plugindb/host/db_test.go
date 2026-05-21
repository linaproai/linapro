// This file tests host-side plugindb DB wrapper and DoCommit governance
// interception.

package host

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"

	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/dialect"
)

// TestPluginDataDriverTypeUsesSharedSupportedDrivers verifies governed driver
// wrappers are derived from LinaPro's shared database driver registry.
func TestPluginDataDriverTypeUsesSharedSupportedDrivers(t *testing.T) {
	tests := []struct {
		name     string
		baseType string
		want     string
	}{
		{name: "postgresql", baseType: " pgsql ", want: "plugin-data-pgsql"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := pluginDataDriverType(test.baseType)
			if err != nil {
				t.Fatalf("pluginDataDriverType failed: %v", err)
			}
			if got != test.want {
				t.Fatalf("expected driver type %q, got %q", test.want, got)
			}
		})
	}

	if _, err := pluginDataDriverType("mysql"); err == nil {
		t.Fatal("expected mysql to be rejected by plugin data driver registry")
	}
	if _, err := pluginDataDriverType("sqlite"); err == nil {
		t.Fatal("expected sqlite to be rejected by plugin data driver registry")
	}
}

// TestDBDoCommitRejectsUnauthorizedTable verifies the governed DB wrapper
// rejects SQL targeting a table outside the authorized resource scope.
func TestDBDoCommitRejectsUnauthorizedTable(t *testing.T) {
	db, err := DB()
	if err != nil {
		t.Fatalf("DB failed: %v", err)
	}
	ctx := WithAudit(context.Background(), &AuditMetadata{
		PluginID:      "test-plugin-data",
		Table:         "sys_plugin_node_state",
		Method:        "delete",
		ResourceTable: "sys_plugin_node_state",
	})
	_, err = db.Ctx(ctx).Exec(ctx, "DELETE FROM sys_plugin WHERE plugin_id = ?", "forbidden")
	if err == nil {
		t.Fatal("expected DoCommit to reject unauthorized table")
	}
	if !strings.Contains(err.Error(), "authorized table") {
		t.Fatalf("expected unauthorized table error, got %v", err)
	}
}

// TestValidatePluginDataCommitTableUsesExactIdentifier verifies table checks
// do not accept unauthorized identifiers that merely contain the table name.
func TestValidatePluginDataCommitTableUsesExactIdentifier(t *testing.T) {
	metadata := &AuditMetadata{
		PluginID:      "test-plugin-data",
		Table:         "sys_plugin_node_state",
		Method:        "list",
		ResourceTable: "sys_plugin_node_state",
	}
	err := validatePluginDataCommitTable(metadata, gdbDoCommitInputForTest(
		`SELECT * FROM sys_plugin_node_state_archive WHERE plugin_id = ?`,
	), mustPostgresDialectForTest(t))
	if err == nil {
		t.Fatal("expected lookalike table name to be rejected")
	}
	if !strings.Contains(err.Error(), "authorized table") {
		t.Fatalf("expected authorized-table error, got %v", err)
	}

	if err = validatePluginDataCommitTable(metadata, gdbDoCommitInputForTest(
		`SELECT * FROM "sys_plugin_node_state" WHERE plugin_id = ?`,
	), mustPostgresDialectForTest(t)); err != nil {
		t.Fatalf("expected exact quoted table name to pass, got %v", err)
	}
}

// TestValidatePluginDataCommitTableAllowsMetadataLookup verifies GoFrame
// metadata queries can reference the authorized table through bind arguments.
func TestValidatePluginDataCommitTableAllowsMetadataLookup(t *testing.T) {
	metadata := &AuditMetadata{
		PluginID:      "test-plugin-data",
		Table:         "plugin_linapro_demo_dynamic_record",
		Method:        "list",
		ResourceTable: "plugin_linapro_demo_dynamic_record",
	}
	err := validatePluginDataCommitTable(metadata, gdbDoCommitInputForTest(
		`SELECT column_name FROM information_schema.columns WHERE table_name = ?`,
		"plugin_linapro_demo_dynamic_record",
	), mustPostgresDialectForTest(t))
	if err != nil {
		t.Fatalf("expected metadata lookup for authorized table to pass, got %v", err)
	}
}

// TestValidatePluginDataCommitTableAllowsSchemaProbe verifies read-only driver
// schema probes do not require a table literal.
func TestValidatePluginDataCommitTableAllowsSchemaProbe(t *testing.T) {
	metadata := &AuditMetadata{
		PluginID:      "test-plugin-data",
		Table:         "plugin_linapro_demo_dynamic_record",
		Method:        "list",
		ResourceTable: "plugin_linapro_demo_dynamic_record",
	}
	err := validatePluginDataCommitTable(
		metadata,
		gdbDoCommitInputForTest(`SELECT current_schema()`),
		mustPostgresDialectForTest(t),
	)
	if err != nil {
		t.Fatalf("expected schema probe to pass, got %v", err)
	}

	err = validatePluginDataCommitTable(metadata, gdbDoCommitInputForTest(
		`SELECT c.relname FROM pg_class c INNER JOIN pg_namespace n ON c.relnamespace = n.oid WHERE n.nspname = 'public' AND c.relkind IN ('r', 'p') AND c.relpartbound IS NULL ORDER BY c.relname`,
	), mustPostgresDialectForTest(t))
	if err != nil {
		t.Fatalf("expected schema table listing to pass, got %v", err)
	}
}

// TestValidatePluginDataCommitTableRejectsMutationSchemaProbe verifies
// metadata probe allowlists cannot bypass table checks for mutations.
func TestValidatePluginDataCommitTableRejectsMutationSchemaProbe(t *testing.T) {
	metadata := &AuditMetadata{
		PluginID:      "test-plugin-data",
		Table:         "plugin_linapro_demo_dynamic_record",
		Method:        "transaction",
		ResourceTable: "plugin_linapro_demo_dynamic_record",
	}
	err := validatePluginDataCommitTable(metadata, gdb.DoCommitInput{
		Type: gdb.SqlTypeExecContext,
		Sql:  `DELETE FROM sys_plugin WHERE id IN (SELECT id FROM pg_class WHERE relname = ?)`,
		Args: []any{"plugin_linapro_demo_dynamic_record"},
	}, mustPostgresDialectForTest(t))
	if err == nil {
		t.Fatal("expected mutation schema probe to be rejected")
	}
	if !strings.Contains(err.Error(), "authorized table") {
		t.Fatalf("expected authorized-table error, got %v", err)
	}
}

// TestValidatePluginDataCommitTableRejectsCatalogSubqueryAroundUnauthorizedTable
// verifies dialect metadata classification cannot bypass unauthorized
// application table checks when catalog SQL appears in a subquery.
func TestValidatePluginDataCommitTableRejectsCatalogSubqueryAroundUnauthorizedTable(t *testing.T) {
	metadata := &AuditMetadata{
		PluginID:      "test-plugin-data",
		Table:         "plugin_linapro_demo_dynamic_record",
		Method:        "list",
		ResourceTable: "plugin_linapro_demo_dynamic_record",
	}
	err := validatePluginDataCommitTable(metadata, gdbDoCommitInputForTest(
		`SELECT * FROM sys_plugin WHERE id IN (SELECT id FROM pg_class WHERE relname = ?)`,
		"plugin_linapro_demo_dynamic_record",
	), mustPostgresDialectForTest(t))
	if err == nil {
		t.Fatal("expected catalog subquery around unauthorized table to be rejected")
	}
	if !strings.Contains(err.Error(), "authorized table") {
		t.Fatalf("expected authorized-table error, got %v", err)
	}
}

// TestPluginDataHostDoesNotEmbedPostgreSQLCatalogSQL verifies database-specific
// catalog strings stay behind the dialect boundary.
func TestPluginDataHostDoesNotEmbedPostgreSQLCatalogSQL(t *testing.T) {
	t.Parallel()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file failed")
	}
	content, err := os.ReadFile(filepath.Join(filepath.Dir(file), "db.go"))
	if err != nil {
		t.Fatalf("read db.go failed: %v", err)
	}
	source := string(content)
	for _, forbidden := range []string{
		"information_schema",
		"pg_catalog",
		"current_schema",
		"pg_class",
		"pg_namespace",
		"version()",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("plugindb host must not embed database-specific catalog SQL %q", forbidden)
		}
	}
}

// gdbDoCommitInputForTest creates one minimal commit input for table-guard
// tests without requiring a live database.
func gdbDoCommitInputForTest(sql string, args ...any) gdb.DoCommitInput {
	return gdb.DoCommitInput{
		Type: gdb.SqlTypeQueryContext,
		Sql:  sql,
		Args: args,
	}
}

// mustPostgresDialectForTest resolves the PostgreSQL dialect used by table
// guard tests without requiring a live database.
func mustPostgresDialectForTest(t *testing.T) dialect.Dialect {
	t.Helper()

	dbDialect, err := dialect.FromDriverType("pgsql")
	if err != nil {
		t.Fatalf("resolve PostgreSQL dialect failed: %v", err)
	}
	return dbDialect
}
