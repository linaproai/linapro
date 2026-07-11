// This file verifies auth.go public contract surfaces: Service method
// boundaries and user-session ClientType parsing.

package auth

import (
	"reflect"
	"testing"

	"lina-core/pkg/bizerr"
	tokencap "lina-core/pkg/plugin/capability/authcap/token"
)

// TestServiceContractDoesNotExposeTenantWorkflow verifies tenant workflow
// orchestration stays outside the core auth service contract.
func TestServiceContractDoesNotExposeTenantWorkflow(t *testing.T) {
	serviceType := reflect.TypeOf((*Service)(nil)).Elem()
	for _, methodName := range []string{"SelectTenant", "SwitchTenant", "SwitchTenantToken"} {
		if _, ok := serviceType.MethodByName(methodName); ok {
			t.Fatalf("auth.Service must not expose tenant workflow method %s", methodName)
		}
	}
}

// TestClientTypeRejectsNonUserActors verifies service and plugin actors do not
// enter the user-session client type enum.
func TestClientTypeRejectsNonUserActors(t *testing.T) {
	for _, value := range []string{"", "service", "plugin"} {
		if _, err := ParseClientType(value); !bizerr.Is(err, CodeAuthClientTypeInvalid) {
			t.Fatalf("expected invalid client type for %q, got %v", value, err)
		}
	}
	for _, value := range []ClientType{tokencap.ClientTypeWeb, tokencap.ClientTypeMobile, tokencap.ClientTypeDesktop, tokencap.ClientTypeCLI} {
		parsed, err := ParseClientType(value.String())
		if err != nil {
			t.Fatalf("parse allowed client type %q: %v", value, err)
		}
		if parsed != value {
			t.Fatalf("expected client type %q, got %q", value, parsed)
		}
	}
}

// TestLoginErrorMessageKeysMatchRuntimeI18n ensures login-facing auth codes use
// convention-derived message keys (bizerr.MessageKey) that ship in error.json.
func TestLoginErrorMessageKeysMatchRuntimeI18n(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		code *bizerr.Code
		key  string
	}{
		{
			name: "invalid credentials",
			code: CodeAuthInvalidCredentials,
			key:  "error.auth.invalid.credentials",
		},
		{
			name: "user disabled",
			code: CodeAuthUserDisabled,
			key:  "error.auth.user.disabled",
		},
		{
			name: "ip blacklisted",
			code: CodeAuthIPBlacklisted,
			key:  "error.auth.ip.blacklisted",
		},
	}
	for _, tc := range cases {
		if got := tc.code.MessageKey(); got != tc.key {
			t.Fatalf("%s: expected messageKey %q, got %q", tc.name, tc.key, got)
		}
		if got := bizerr.MessageKey(tc.code.RuntimeCode()); got != tc.key {
			t.Fatalf("%s: MessageKey(%q)=%q, want %q", tc.name, tc.code.RuntimeCode(), got, tc.key)
		}
	}
}
