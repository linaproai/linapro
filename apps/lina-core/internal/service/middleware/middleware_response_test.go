// This file verifies unified response metadata helpers.
package middleware

import (
	"testing"

	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

// TestApplyRuntimeErrorMetadataCopiesStructuredFields verifies the response
// envelope exposes stable structured-error metadata for frontend localization.
func TestApplyRuntimeErrorMetadataCopiesStructuredFields(t *testing.T) {
	t.Parallel()

	response := &runtimeHandlerResponse{}
	code := bizerr.MustDefine(
		"USER_NOT_FOUND",
		"User {username} does not exist",
		gcode.CodeNotFound,
	)
	err := bizerr.NewCode(code, bizerr.P("username", "alice"))

	applyRuntimeErrorMetadata(response, err)
	if response.ErrorCode != "USER_NOT_FOUND" {
		t.Fatalf("expected error code %q, got %q", "USER_NOT_FOUND", response.ErrorCode)
	}
	if response.MessageKey != "error.user.not.found" {
		t.Fatalf("expected message key %q, got %q", "error.user.not.found", response.MessageKey)
	}
	if response.MessageParams["username"] != "alice" {
		t.Fatalf("expected message parameter username=%q, got %v", "alice", response.MessageParams["username"])
	}
}
