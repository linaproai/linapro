// This file verifies authentication service interface boundaries.

package auth

import (
	"reflect"
	"testing"
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
	for _, methodName := range []string{"IssueTenantToken", "ReissueTenantToken", "ReissueTenantTokenFromBearer"} {
		if _, ok := issuerType.MethodByName(methodName); !ok {
			t.Fatalf("TenantTokenIssuer must expose tenant token method %s", methodName)
		}
	}
}
