// This file tests host-side plugindb DB wrapper and DoCommit governance
// interception.

package host

import (
	"context"
	"strings"
	"testing"

	_ "lina-core/pkg/dbdriver"
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
		{name: "sqlite", baseType: "SQLITE", want: "plugin-data-sqlite"},
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
