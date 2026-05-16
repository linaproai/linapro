// This file covers dynamic-plugin lifecycle precondition response handling.

package runtime

import (
	"net/http"
	"testing"

	bridgecontract "lina-core/pkg/pluginbridge/contract"
	"lina-core/pkg/pluginhost"
)

// TestApplyDynamicLifecycleResponseRecordsVeto verifies explicit guest veto
// responses are preserved as lifecycle decisions instead of bridge errors.
func TestApplyDynamicLifecycleResponseRecordsVeto(t *testing.T) {
	decision := &DynamicLifecycleDecision{
		PluginID:  "plugin-dynamic-veto",
		Operation: pluginhost.LifecycleHookBeforeInstall,
		OK:        true,
	}

	err := applyDynamicLifecycleResponse(decision, &bridgecontract.BridgeResponseEnvelopeV1{
		StatusCode: http.StatusOK,
		Body:       []byte(`{"ok":false,"reason":"plugin.dynamic.veto"}`),
	})
	if err != nil {
		t.Fatalf("expected veto response to decode without bridge error, got %v", err)
	}
	if decision.OK || decision.Reason != "plugin.dynamic.veto" {
		t.Fatalf("expected explicit lifecycle veto decision, got %#v", decision)
	}
}
