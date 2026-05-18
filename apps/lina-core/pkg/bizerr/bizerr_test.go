// This file verifies structured business error construction and fallback
// formatting.

package bizerr

import (
	"errors"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// TestMessageKeyDerivesFromErrorCode verifies stable error codes map to
// predictable runtime i18n keys.
func TestMessageKeyDerivesFromErrorCode(t *testing.T) {
	t.Parallel()

	actual := MessageKey("DICT_TYPE_EXISTS")
	expected := "error.dict.type.exists"
	if actual != expected {
		t.Fatalf("expected message key %q, got %q", expected, actual)
	}
}

// TestFormatReplacesNamedParameters verifies named placeholders are rendered
// without requiring fmt-style positional arguments.
func TestFormatReplacesNamedParameters(t *testing.T) {
	t.Parallel()

	actual := Format("User {username} has {count} pending tasks", map[string]any{
		"username": "alice",
		"count":    3,
	})
	expected := "User alice has 3 pending tasks"
	if actual != expected {
		t.Fatalf("expected formatted message %q, got %q", expected, actual)
	}
}

// TestCodeAccessorsAndMatching verifies reusable definitions expose their
// structured metadata without forcing callers to duplicate raw error strings.
func TestCodeAccessorsAndMatching(t *testing.T) {
	t.Parallel()

	runtimeCode := "USER_LOCKED"
	messageKey := "error.user.locked.custom"
	fallback := "User is locked"
	code := MustDefineWithKey(
		runtimeCode,
		messageKey,
		fallback,
		gcode.CodeNotAuthorized,
	)
	var nilCode *Code

	if code.RuntimeCode() != runtimeCode {
		t.Fatalf("expected runtime code %q, got %q", runtimeCode, code.RuntimeCode())
	}
	if code.MessageKey() != messageKey {
		t.Fatalf("expected message key %q, got %q", messageKey, code.MessageKey())
	}
	if code.Fallback() != fallback {
		t.Fatalf("expected fallback %q, got %q", fallback, code.Fallback())
	}
	if code.TypeCode() != gcode.CodeNotAuthorized {
		t.Fatalf("expected GoFrame type code %v, got %v", gcode.CodeNotAuthorized, code.TypeCode())
	}
	if nilCode.RuntimeCode() != "" || nilCode.MessageKey() != "" || nilCode.Fallback() != "" {
		t.Fatal("expected nil code string metadata accessors to return empty values")
	}
	if nilCode.TypeCode() != gcode.CodeUnknown {
		t.Fatalf("expected nil code type to be unknown, got %v", nilCode.TypeCode())
	}

	err := NewCode(code)
	messageErr, ok := As(err)
	if !ok {
		t.Fatal("expected structured error to be discoverable from error chain")
	}
	if !messageErr.Matches(code) {
		t.Fatal("expected structured error to match its reusable definition")
	}
	if !Is(err, code) {
		t.Fatal("expected package-level matcher to match its reusable definition")
	}
	if Is(errors.New("plain error"), code) {
		t.Fatal("expected package-level matcher to reject unstructured errors")
	}
	if Is(err, nil) {
		t.Fatal("expected package-level matcher to reject nil definitions")
	}
}

// TestStructuredErrorCarriesRuntimeMetadata verifies wrapped errors preserve
// GoFrame response code semantics and runtime-message metadata.
func TestStructuredErrorCarriesRuntimeMetadata(t *testing.T) {
	t.Parallel()

	code := MustDefine(
		"USER_NOT_FOUND",
		"User {username} does not exist",
		gcode.CodeNotFound,
	)
	cause := errors.New("storage unavailable")
	err := WrapCode(cause, code, P("username", "alice"))

	if !errors.Is(err, cause) {
		t.Fatal("expected structured error to wrap the original cause")
	}
	if actual := gerror.Code(err); actual != gcode.CodeNotFound {
		t.Fatalf("expected GoFrame type code %v, got %v", gcode.CodeNotFound, actual)
	}
	stack := gerror.Stack(err)
	if !strings.Contains(stack, "TestStructuredErrorCarriesRuntimeMetadata") {
		t.Fatalf("expected stack to include test creation site, got %q", stack)
	}
	messageErr, ok := As(err)
	if !ok {
		t.Fatal("expected structured error to be discoverable from error chain")
	}
	if messageErr.TypeCode() != gcode.CodeNotFound {
		t.Fatalf("expected GoFrame type code %v, got %v", gcode.CodeNotFound, messageErr.TypeCode())
	}
	if messageErr.RuntimeCode() != code.RuntimeCode() {
		t.Fatalf("expected runtime code %q, got %q", code.RuntimeCode(), messageErr.RuntimeCode())
	}
	if messageErr.MessageKey() != code.MessageKey() {
		t.Fatalf("expected message key %q, got %q", code.MessageKey(), messageErr.MessageKey())
	}
	expectedErrorText := "User alice does not exist: storage unavailable"
	if actual := messageErr.Error(); actual != expectedErrorText {
		t.Fatalf("expected fallback error %q, got %q", expectedErrorText, actual)
	}
	params := messageErr.Params()
	params["username"] = "bob"
	if actual := messageErr.Error(); actual != expectedErrorText {
		t.Fatalf("expected Params to return a defensive copy, got %q", actual)
	}
}
