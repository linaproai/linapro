// This file verifies monitor-operlog structured business error metadata.

package operlog

import (
	"testing"

	"lina-core/pkg/bizerr"
)

// TestOperLogBusinessErrorMetadata verifies operation-log errors expose stable
// runtime codes and i18n keys instead of fixed-language text.
func TestOperLogBusinessErrorMetadata(t *testing.T) {
	err := bizerr.NewCode(CodeOperLogNotFound)
	messageErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured business error, got %T", err)
	}
	if messageErr.RuntimeCode() != "MONITOR_OPERLOG_NOT_FOUND" {
		t.Fatalf("expected operation-log not-found code, got %q", messageErr.RuntimeCode())
	}
	if messageErr.MessageKey() != "error.monitor.operlog.not.found" {
		t.Fatalf("expected operation-log not-found key, got %q", messageErr.MessageKey())
	}
}
