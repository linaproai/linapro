// This file verifies organization department structured business error metadata.

package dept

import (
	"testing"

	"lina-core/pkg/bizerr"
)

// TestDeptBusinessErrorMetadata verifies department errors expose stable
// runtime codes and i18n keys instead of fixed-language text.
func TestDeptBusinessErrorMetadata(t *testing.T) {
	testCases := []struct {
		name        string
		code        *bizerr.Code
		runtimeCode string
		messageKey  string
	}{
		{
			name:        "not found",
			code:        CodeDeptNotFound,
			runtimeCode: "ORG_DEPT_NOT_FOUND",
			messageKey:  "error.org.dept.not.found",
		},
		{
			name:        "has children",
			code:        CodeDeptHasChildrenDeleteDenied,
			runtimeCode: "ORG_DEPT_HAS_CHILDREN_DELETE_DENIED",
			messageKey:  "error.org.dept.has.children.delete.denied",
		},
		{
			name:        "has users",
			code:        CodeDeptHasUsersDeleteDenied,
			runtimeCode: "ORG_DEPT_HAS_USERS_DELETE_DENIED",
			messageKey:  "error.org.dept.has.users.delete.denied",
		},
		{
			name:        "code exists",
			code:        CodeDeptCodeExists,
			runtimeCode: "ORG_DEPT_CODE_EXISTS",
			messageKey:  "error.org.dept.code.exists",
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
