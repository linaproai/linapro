// This file verifies dynamic-demo structured business error metadata.

package dynamicservice

import (
	"testing"

	"lina-core/pkg/bizerr"
)

// TestDynamicDemoBusinessErrorMetadata verifies dynamic-demo errors expose
// stable runtime codes, i18n keys, and named message parameters.
func TestDynamicDemoBusinessErrorMetadata(t *testing.T) {
	err := bizerr.NewCode(CodeDynamicDemoRecordTitleTooLong, bizerr.P("maxChars", 128))
	messageErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured business error, got %T", err)
	}
	if messageErr.RuntimeCode() != "PLUGIN_DEMO_DYNAMIC_RECORD_TITLE_TOO_LONG" {
		t.Fatalf("expected title-too-long code, got %q", messageErr.RuntimeCode())
	}
	if messageErr.MessageKey() != "error.plugin.demo.dynamic.record.title.too.long" {
		t.Fatalf("expected title-too-long key, got %q", messageErr.MessageKey())
	}
	params := messageErr.Params()
	if params["maxChars"] != 128 {
		t.Fatalf("expected maxChars message param 128, got %#v", params["maxChars"])
	}
}
