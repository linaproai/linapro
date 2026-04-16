// This file tests host-side plugindb DB wrapper and DoCommit governance
// interception.

package host

import (
	"context"
	"strings"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

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
	if !strings.Contains(err.Error(), "授权表") {
		t.Fatalf("expected unauthorized table error, got %v", err)
	}
}
