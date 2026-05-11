// This file verifies tenant service business error metadata.

package tenant

import (
	"testing"

	"lina-core/pkg/bizerr"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
)

// TestTenantBusinessErrorMetadata verifies tenant errors expose stable runtime
// codes and i18n keys.
func TestTenantBusinessErrorMetadata(t *testing.T) {
	testCases := []struct {
		name        string
		code        *bizerr.Code
		runtimeCode string
		messageKey  string
	}{
		{name: "not found", code: CodeTenantNotFound, runtimeCode: "MULTI_TENANT_TENANT_NOT_FOUND", messageKey: "error.multi.tenant.tenant.not.found"},
		{name: "code exists", code: CodeTenantCodeExists, runtimeCode: "MULTI_TENANT_TENANT_CODE_EXISTS", messageKey: "error.multi.tenant.tenant.code.exists"},
		{name: "code invalid", code: CodeTenantCodeInvalid, runtimeCode: "MULTI_TENANT_TENANT_CODE_INVALID", messageKey: "error.multi.tenant.tenant.code.invalid"},
		{name: "code reserved", code: CodeTenantCodeReserved, runtimeCode: "MULTI_TENANT_TENANT_CODE_RESERVED", messageKey: "error.multi.tenant.tenant.code.reserved"},
		{name: "invalid status", code: CodeTenantInvalidStatus, runtimeCode: "MULTI_TENANT_INVALID_TENANT_STATUS", messageKey: "error.multi.tenant.invalid.tenant.status"},
		{name: "transition invalid", code: CodeTenantStatusTransitionInvalid, runtimeCode: "MULTI_TENANT_TENANT_STATUS_TRANSITION_INVALID", messageKey: "error.multi.tenant.tenant.status.transition.invalid"},
		{name: "delete guard vetoed", code: CodeTenantDeleteGuardVetoed, runtimeCode: "MULTI_TENANT_DELETE_GUARD_VETOED", messageKey: "error.multi.tenant.delete.guard.vetoed"},
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

// TestTenantCodeValidation verifies the public tenant-code contract.
func TestTenantCodeValidation(t *testing.T) {
	validCodes := []string{"aa", "a1", "acme-01", "tenant123"}
	for _, code := range validCodes {
		if err := validateTenantCode(code); err != nil {
			t.Fatalf("expected tenant code %q to be valid: %v", code, err)
		}
	}

	invalidCodes := []string{"a", "A1", "-aa", "aa-", "aa_beta", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
	for _, code := range invalidCodes {
		if err := validateTenantCode(code); !bizerr.Is(err, CodeTenantCodeInvalid) {
			t.Fatalf("expected tenant code %q to be invalid, got %v", code, err)
		}
	}
}

// TestTenantStatusTransitionRules verifies the lifecycle state machine.
func TestTenantStatusTransitionRules(t *testing.T) {
	allowed := [][2]string{
		{"active", "suspended"},
		{"suspended", "active"},
	}
	for _, item := range allowed {
		if !isStatusTransitionAllowed(sharedTenantStatus(item[0]), sharedTenantStatus(item[1])) {
			t.Fatalf("expected transition %s -> %s to be allowed", item[0], item[1])
		}
	}

	rejected := [][2]string{
		{"active", "active"},
		{"active", "deleted"},
		{"suspended", "deleted"},
		{"suspended", "suspended"},
	}
	for _, item := range rejected {
		if isStatusTransitionAllowed(sharedTenantStatus(item[0]), sharedTenantStatus(item[1])) {
			t.Fatalf("expected transition %s -> %s to be rejected", item[0], item[1])
		}
	}
}

// sharedTenantStatus keeps table-driven tests readable.
func sharedTenantStatus(value string) shared.TenantStatus {
	return shared.TenantStatus(value)
}
