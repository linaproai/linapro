// This file verifies monitor-loginlog structured business error metadata.

package loginlog

import (
	"testing"

	"lina-core/pkg/bizerr"
)

// TestLoginLogBusinessErrorMetadata verifies login-log errors expose stable
// runtime codes and i18n keys instead of fixed-language text.
func TestLoginLogBusinessErrorMetadata(t *testing.T) {
	err := bizerr.NewCode(CodeLoginLogNotFound)
	messageErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured business error, got %T", err)
	}
	if messageErr.RuntimeCode() != "MONITOR_LOGINLOG_NOT_FOUND" {
		t.Fatalf("expected login-log not-found code, got %q", messageErr.RuntimeCode())
	}
	if messageErr.MessageKey() != "error.monitor.loginlog.not.found" {
		t.Fatalf("expected login-log not-found key, got %q", messageErr.MessageKey())
	}
}
