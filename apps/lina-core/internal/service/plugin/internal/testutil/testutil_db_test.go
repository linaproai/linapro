// This file verifies plugin test database cleanup helpers across supported
// database dialects.

package testutil

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TestCleanupPluginGovernanceRowsHardUsesExplicitPluginConditions verifies the
// shared cleanup helper keeps DELETE statements scoped on SQLite, where GoFrame
// rejects unqualified DELETE operations.
func TestCleanupPluginGovernanceRowsHardUsesExplicitPluginConditions(t *testing.T) {
	ctx := context.Background()
	link := "sqlite::@file(" + t.TempDir() + "/plugin-cleanup.db)"
	originalConfig := gdb.GetAllConfig()
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: link}},
	}); err != nil {
		t.Fatalf("configure SQLite cleanup database: %v", err)
	}
	db := g.DB()
	t.Cleanup(func() {
		if closeErr := db.Close(ctx); closeErr != nil {
			t.Errorf("close SQLite cleanup database: %v", closeErr)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore GoFrame database config: %v", err)
		}
	})

	createMinimalPluginGovernanceTables(t, ctx)
	insertPluginGovernanceRows(t, ctx, "plugin-cleanup-target")
	insertPluginGovernanceRows(t, ctx, "plugin-cleanup-other")

	CleanupPluginGovernanceRowsHard(t, ctx, "plugin-cleanup-target")

	assertPluginGovernanceRowCounts(t, ctx, "plugin-cleanup-target", 0)
	assertPluginGovernanceRowCounts(t, ctx, "plugin-cleanup-other", 6)
}

// createMinimalPluginGovernanceTables creates only the columns needed by the
// cleanup helper so the test stays focused on DELETE scoping behavior.
func createMinimalPluginGovernanceTables(t *testing.T, ctx context.Context) {
	t.Helper()

	statements := []string{
		"CREATE TABLE sys_plugin_node_state (id INTEGER PRIMARY KEY AUTOINCREMENT, plugin_id TEXT NOT NULL);",
		"CREATE TABLE sys_plugin_resource_ref (id INTEGER PRIMARY KEY AUTOINCREMENT, plugin_id TEXT NOT NULL, deleted_at TIMESTAMP NULL);",
		"CREATE TABLE sys_plugin_state (id INTEGER PRIMARY KEY AUTOINCREMENT, plugin_id TEXT NOT NULL);",
		"CREATE TABLE sys_plugin_migration (id INTEGER PRIMARY KEY AUTOINCREMENT, plugin_id TEXT NOT NULL);",
		"CREATE TABLE sys_plugin_release (id INTEGER PRIMARY KEY AUTOINCREMENT, plugin_id TEXT NOT NULL);",
		"CREATE TABLE sys_plugin (id INTEGER PRIMARY KEY AUTOINCREMENT, plugin_id TEXT NOT NULL, deleted_at TIMESTAMP NULL);",
	}
	for index, statement := range statements {
		if _, err := g.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("create cleanup table statement %d: %v", index+1, err)
		}
	}
}

// insertPluginGovernanceRows inserts one row per table for the given plugin ID.
func insertPluginGovernanceRows(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()

	tables := []string{
		"sys_plugin_node_state",
		"sys_plugin_resource_ref",
		"sys_plugin_state",
		"sys_plugin_migration",
		"sys_plugin_release",
		"sys_plugin",
	}
	for _, table := range tables {
		if _, err := g.DB().Exec(ctx, "INSERT INTO "+table+" (plugin_id) VALUES (?)", pluginID); err != nil {
			t.Fatalf("insert cleanup row into %s: %v", table, err)
		}
	}
}

// assertPluginGovernanceRowCounts verifies the remaining rows for one plugin ID
// across every table touched by the cleanup helper.
func assertPluginGovernanceRowCounts(t *testing.T, ctx context.Context, pluginID string, expectedTotal int) {
	t.Helper()

	tables := []string{
		"sys_plugin_node_state",
		"sys_plugin_resource_ref",
		"sys_plugin_state",
		"sys_plugin_migration",
		"sys_plugin_release",
		"sys_plugin",
	}
	total := 0
	for _, table := range tables {
		count, err := g.DB().GetValue(ctx, "SELECT COUNT(1) FROM "+table+" WHERE plugin_id = ?", pluginID)
		if err != nil {
			t.Fatalf("count cleanup rows in %s: %v", table, err)
		}
		total += count.Int()
	}
	if total != expectedTotal {
		t.Fatalf("expected %d cleanup rows for %s, got %d", expectedTotal, pluginID, total)
	}
}
