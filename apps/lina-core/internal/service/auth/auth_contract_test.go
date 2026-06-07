// This file verifies authentication service interface boundaries.

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

// TestTenantTokenIssuerOwnsTenantTokenWorkflow verifies tenant token handoff is
// available through the explicit narrow interface.
func TestTenantTokenIssuerOwnsTenantTokenWorkflow(t *testing.T) {
	issuerType := reflect.TypeOf((*TenantTokenIssuer)(nil)).Elem()
	for _, methodName := range []string{
		"IssueTenantToken",
		"ReissueTenantToken",
		"ReissueTenantTokenFromBearer",
		"IssueImpersonationToken",
		"RevokeImpersonationToken",
	} {
		if _, ok := issuerType.MethodByName(methodName); !ok {
			t.Fatalf("TenantTokenIssuer must expose tenant token method %s", methodName)
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
