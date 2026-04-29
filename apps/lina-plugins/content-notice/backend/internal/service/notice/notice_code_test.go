// This file verifies content-notice structured business error metadata.

package notice

import (
	"testing"

	"lina-core/pkg/bizerr"
)

// TestNoticeBusinessErrorMetadata verifies notice errors expose stable runtime
// codes and i18n keys instead of fixed-language text.
func TestNoticeBusinessErrorMetadata(t *testing.T) {
	testCases := []struct {
		name        string
		code        *bizerr.Code
		runtimeCode string
		messageKey  string
	}{
		{
			name:        "not found",
			code:        CodeNoticeNotFound,
			runtimeCode: "CONTENT_NOTICE_NOT_FOUND",
			messageKey:  "error.content.notice.not.found",
		},
		{
			name:        "delete required",
			code:        CodeNoticeDeleteRequired,
			runtimeCode: "CONTENT_NOTICE_DELETE_REQUIRED",
			messageKey:  "error.content.notice.delete.required",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			err := bizerr.NewCode(testCase.code)
			messageErr, ok := bizerr.As(err)
			if !ok {
				t.Fatalf("expected structured business error, got %T", err)
			}
			if messageErr.RuntimeCode() != testCase.runtimeCode {
				t.Fatalf("expected runtime code %q, got %q", testCase.runtimeCode, messageErr.RuntimeCode())
			}
			if messageErr.MessageKey() != testCase.messageKey {
				t.Fatalf("expected message key %q, got %q", testCase.messageKey, messageErr.MessageKey())
			}
		})
	}
}
