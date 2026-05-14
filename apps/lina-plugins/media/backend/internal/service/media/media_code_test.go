// This file verifies media structured business error metadata.

package media

import (
	"testing"

	"lina-core/pkg/bizerr"
)

// TestMediaBusinessErrorMetadata verifies media errors expose stable runtime codes and message keys.
func TestMediaBusinessErrorMetadata(t *testing.T) {
	testCases := []struct {
		name        string
		code        *bizerr.Code
		runtimeCode string
		messageKey  string
	}{
		{
			name:        "strategy not found",
			code:        CodeMediaStrategyNotFound,
			runtimeCode: "MEDIA_STRATEGY_NOT_FOUND",
			messageKey:  "error.media.strategy.not.found",
		},
		{
			name:        "binding device required",
			code:        CodeMediaBindingDeviceRequired,
			runtimeCode: "MEDIA_BINDING_DEVICE_REQUIRED",
			messageKey:  "error.media.binding.device.required",
		},
		{
			name:        "alias not found",
			code:        CodeMediaAliasNotFound,
			runtimeCode: "MEDIA_ALIAS_NOT_FOUND",
			messageKey:  "error.media.alias.not.found",
		},
		{
			name:        "tenant whitelist duplicate",
			code:        CodeMediaTenantWhiteDuplicate,
			runtimeCode: "MEDIA_TENANT_WHITE_DUPLICATE",
			messageKey:  "error.media.tenant.white.duplicate",
		},
		{
			name:        "tenant whitelist ip invalid",
			code:        CodeMediaTenantWhiteIPInvalid,
			runtimeCode: "MEDIA_TENANT_WHITE_IP_INVALID",
			messageKey:  "error.media.tenant.white.ip.invalid",
		},
		{
			name:        "media node referenced",
			code:        CodeMediaNodeReferenced,
			runtimeCode: "MEDIA_NODE_REFERENCED",
			messageKey:  "error.media.node.referenced",
		},
		{
			name:        "device node duplicate",
			code:        CodeMediaDeviceNodeDuplicate,
			runtimeCode: "MEDIA_DEVICE_NODE_DUPLICATE",
			messageKey:  "error.media.device.node.duplicate",
		},
		{
			name:        "tenant stream duplicate",
			code:        CodeMediaTenantStreamDuplicate,
			runtimeCode: "MEDIA_TENANT_STREAM_DUPLICATE",
			messageKey:  "error.media.tenant.stream.duplicate",
		},
		{
			name:        "tieta tenant mismatch",
			code:        CodeMediaTietaTenantMismatch,
			runtimeCode: "MEDIA_TIETA_TENANT_MISMATCH",
			messageKey:  "error.media.tieta.tenant.mismatch",
		},
		{
			name:        "tieta device permission denied",
			code:        CodeMediaTietaDevicePermissionDenied,
			runtimeCode: "MEDIA_TIETA_DEVICE_PERMISSION_DENIED",
			messageKey:  "error.media.tieta.device.permission.denied",
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
