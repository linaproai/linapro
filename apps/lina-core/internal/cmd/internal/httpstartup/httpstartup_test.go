// This file verifies the top-level HTTP startup entrypoint orchestration order.

package httpstartup

import (
	"os"
	"strings"
	"testing"
)

// TestHTTPStartupRegistersSourceRoutesBeforeConsistencyValidation protects the
// startup ordering required by source plugins that register host capability
// providers from HTTP route callbacks.
func TestHTTPStartupRegistersSourceRoutesBeforeConsistencyValidation(t *testing.T) {
	content, err := os.ReadFile("httpstartup.go")
	if err != nil {
		t.Fatalf("read HTTP startup source: %v", err)
	}
	text := string(content)
	beforeRoutesIndex := strings.Index(text, "startHTTPRuntimeBeforeSourceRoutes")
	registerRoutesIndex := strings.Index(text, "registerSourcePluginHTTPRoutes")
	finishRuntimeIndex := strings.Index(text, "finishHTTPRuntimeAfterSourceRoutes")
	completeRoutesIndex := strings.Index(text, "completeSourcePluginHTTPRoutes")
	if beforeRoutesIndex < 0 || registerRoutesIndex < 0 || finishRuntimeIndex < 0 || completeRoutesIndex < 0 {
		t.Fatalf("expected split HTTP startup phases to be present")
	}
	if !(beforeRoutesIndex < registerRoutesIndex &&
		registerRoutesIndex < finishRuntimeIndex &&
		finishRuntimeIndex < completeRoutesIndex) {
		t.Fatalf(
			"expected startup order start-before-routes -> register-source-routes -> finish-runtime -> complete-source-routes, got indexes %d %d %d %d",
			beforeRoutesIndex,
			registerRoutesIndex,
			finishRuntimeIndex,
			completeRoutesIndex,
		)
	}
}
