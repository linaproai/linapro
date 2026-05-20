// This file verifies plugin test database cleanup helpers.

package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// TestCleanupPluginGovernanceRowsHardUsesExplicitPluginConditions verifies the
// shared cleanup helper keeps DELETE statements scoped to the target plugin ID.
func TestCleanupPluginGovernanceRowsHardUsesExplicitPluginConditions(t *testing.T) {
	ctx := context.Background()
	targetPluginID := fmt.Sprintf("plugin-cleanup-target-%d", time.Now().UnixNano())
	otherPluginID := fmt.Sprintf("plugin-cleanup-other-%d", time.Now().UnixNano())

	createMinimalPluginGovernanceTables(t, ctx)
	t.Cleanup(func() {
		CleanupPluginGovernanceRowsHard(t, ctx, targetPluginID)
		CleanupPluginGovernanceRowsHard(t, ctx, otherPluginID)
	})
	insertPluginGovernanceRows(t, ctx, targetPluginID)
	insertPluginGovernanceRows(t, ctx, otherPluginID)

	CleanupPluginGovernanceRowsHard(t, ctx, targetPluginID)

	assertPluginGovernanceRowCounts(t, ctx, targetPluginID, 0)
	assertPluginGovernanceRowCounts(t, ctx, otherPluginID, 6)
}

// createMinimalPluginGovernanceTables creates only the columns needed by the
// cleanup helper so the test stays focused on DELETE scoping behavior.
func createMinimalPluginGovernanceTables(t *testing.T, ctx context.Context) {
	t.Helper()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS sys_plugin_node_state ("id" INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, "plugin_id" VARCHAR(64) NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS sys_plugin_resource_ref ("id" INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, "plugin_id" VARCHAR(64) NOT NULL, "deleted_at" TIMESTAMP NULL);`,
		`CREATE TABLE IF NOT EXISTS sys_plugin_state ("id" INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, "plugin_id" VARCHAR(64) NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS sys_plugin_migration ("id" INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, "plugin_id" VARCHAR(64) NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS sys_plugin_release ("id" INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, "plugin_id" VARCHAR(64) NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS sys_plugin ("id" INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, "plugin_id" VARCHAR(64) NOT NULL, "deleted_at" TIMESTAMP NULL);`,
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
