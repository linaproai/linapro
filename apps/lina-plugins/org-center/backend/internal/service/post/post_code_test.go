// This file verifies organization post structured business error metadata.

package post

import (
	"testing"

	"lina-core/pkg/bizerr"
)

// TestPostBusinessErrorMetadata verifies post errors expose stable runtime
// codes, i18n keys, and named message parameters.
func TestPostBusinessErrorMetadata(t *testing.T) {
	err := bizerr.NewCode(CodePostAssignedDeleteDenied, bizerr.P("id", 42))
	messageErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured business error, got %T", err)
	}
	if messageErr.RuntimeCode() != "ORG_POST_ASSIGNED_DELETE_DENIED" {
		t.Fatalf("expected assigned-delete code, got %q", messageErr.RuntimeCode())
	}
	if messageErr.MessageKey() != "error.org.post.assigned.delete.denied" {
		t.Fatalf("expected assigned-delete key, got %q", messageErr.MessageKey())
	}
	params := messageErr.Params()
	if params["id"] != 42 {
		t.Fatalf("expected id message param 42, got %#v", params["id"])
	}

	if !bizerr.Is(err, CodePostAssignedDeleteDenied) {
		t.Fatal("expected error to match post assigned-delete code")
	}
}
