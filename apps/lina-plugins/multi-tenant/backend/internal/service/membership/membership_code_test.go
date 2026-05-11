// This file verifies membership service business error metadata.

package membership

import (
	"testing"

	"lina-core/pkg/bizerr"
)

// TestMembershipBusinessErrorMetadata verifies membership errors expose stable
// runtime codes and i18n keys.
func TestMembershipBusinessErrorMetadata(t *testing.T) {
	testCases := []struct {
		name        string
		code        *bizerr.Code
		runtimeCode string
		messageKey  string
	}{
		{name: "not found", code: CodeMembershipNotFound, runtimeCode: "MULTI_TENANT_MEMBERSHIP_NOT_FOUND", messageKey: "error.multi.tenant.membership.not.found"},
		{name: "exists", code: CodeMembershipExists, runtimeCode: "MULTI_TENANT_MEMBERSHIP_EXISTS", messageKey: "error.multi.tenant.membership.exists"},
		{name: "platform forbidden", code: CodePlatformMembershipForbidden, runtimeCode: "MULTI_TENANT_PLATFORM_MEMBERSHIP_FORBIDDEN", messageKey: "error.multi.tenant.platform.membership.forbidden"},
		{name: "single cardinality", code: CodeSingleCardinalityViolation, runtimeCode: "MULTI_TENANT_SINGLE_CARDINALITY_VIOLATION", messageKey: "error.multi.tenant.single.cardinality.violation"},
		{name: "tenant unavailable", code: CodeTenantUnavailable, runtimeCode: "MULTI_TENANT_TENANT_UNAVAILABLE", messageKey: "error.multi.tenant.tenant.unavailable"},
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
