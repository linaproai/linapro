// This file tests host-side plugindb audit context propagation helpers.

package host

import (
	"context"
	"testing"
)

func TestWithAuditRoundTrip(t *testing.T) {
	metadata := &AuditMetadata{PluginID: "plugin-demo", Table: "sys_plugin_node_state", Method: "list"}
	ctx := WithAudit(context.Background(), metadata)
	decoded := AuditFromContext(ctx)
	if decoded == nil {
		t.Fatal("expected audit metadata in context")
	}
	if decoded.PluginID != metadata.PluginID || decoded.Table != metadata.Table || decoded.Method != metadata.Method {
		t.Fatalf("unexpected decoded metadata: %#v", decoded)
	}
}
