// This file verifies impersonation token metadata and error contracts.

package impersonate

import (
	"testing"
	"time"

	"lina-core/pkg/authtoken"
	"lina-core/pkg/bizerr"
)

// TestImpersonationBusinessErrorMetadata verifies impersonation errors expose stable metadata.
func TestImpersonationBusinessErrorMetadata(t *testing.T) {
	testCases := []struct {
		name        string
		code        *bizerr.Code
		runtimeCode string
		messageKey  string
	}{
		{name: "permission denied", code: CodeImpersonationPermissionDenied, runtimeCode: "MULTI_TENANT_IMPERSONATION_PERMISSION_DENIED", messageKey: "error.multi.tenant.impersonation.permission.denied"},
		{name: "tenant unavailable", code: CodeImpersonationTenantUnavailable, runtimeCode: "MULTI_TENANT_IMPERSONATION_TENANT_UNAVAILABLE", messageKey: "error.multi.tenant.impersonation.tenant.unavailable"},
		{name: "token invalid", code: CodeImpersonationTokenInvalid, runtimeCode: "MULTI_TENANT_IMPERSONATION_TOKEN_INVALID", messageKey: "error.multi.tenant.impersonation.token.invalid"},
		{name: "token unavailable", code: CodeImpersonationTokenUnavailable, runtimeCode: "MULTI_TENANT_IMPERSONATION_TOKEN_UNAVAILABLE", messageKey: "error.multi.tenant.impersonation.token.unavailable"},
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

// TestJWTTokenSignerRoundTrip verifies impersonation tokens carry host-compatible claims.
func TestJWTTokenSignerRoundTrip(t *testing.T) {
	signer := jwtTokenSigner{}
	user := &userRow{Id: 7, Username: "admin", Status: 1}
	token, err := signer.Sign("unit-secret", time.Hour, user, 42, "token-1")
	if err != nil {
		t.Fatalf("sign token failed: %v", err)
	}
	claims, err := signer.Parse("unit-secret", token)
	if err != nil {
		t.Fatalf("parse token failed: %v", err)
	}
	if claims.UserId != 7 || claims.TenantId != 42 || claims.ActingUserId != 7 {
		t.Fatalf("unexpected claims: %#v", claims)
	}
	if claims.TokenType != authtoken.KindAccess {
		t.Fatalf("expected host access token type, got %q", claims.TokenType)
	}
	if !claims.IsImpersonation {
		t.Fatalf("expected impersonation claims, got %#v", claims)
	}
}
