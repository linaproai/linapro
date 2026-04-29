// This file verifies source-demo structured business error metadata.

package demo

import (
	"testing"

	"lina-core/pkg/bizerr"
)

// TestDemoBusinessErrorMetadata verifies source-demo errors expose stable
// runtime codes, i18n keys, and named message parameters.
func TestDemoBusinessErrorMetadata(t *testing.T) {
	err := bizerr.NewCode(CodeDemoAttachmentSizeTooLarge, bizerr.P("maxSizeMB", 5))
	messageErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured business error, got %T", err)
	}
	if messageErr.RuntimeCode() != "PLUGIN_DEMO_SOURCE_ATTACHMENT_SIZE_TOO_LARGE" {
		t.Fatalf("expected attachment-size code, got %q", messageErr.RuntimeCode())
	}
	if messageErr.MessageKey() != "error.plugin.demo.source.attachment.size.too.large" {
		t.Fatalf("expected attachment-size key, got %q", messageErr.MessageKey())
	}
	params := messageErr.Params()
	if params["maxSizeMB"] != 5 {
		t.Fatalf("expected maxSizeMB message param 5, got %#v", params["maxSizeMB"])
	}
}
