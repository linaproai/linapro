// This file verifies the guest-side user capability boundary: external-identity
// provisioning must fail closed for dynamic plugins and must never reach the
// host-service transport, mirroring the Auth.ExternalLogin() stub.

package domainhostcall

import (
	"context"
	"testing"

	"lina-core/pkg/plugin/capability/usercap"
)

// TestUsersProvisionExternalFailsClosed verifies the dynamic users guest client
// rejects external-identity provisioning without invoking the host transport.
// The primitive is operator-less account minting and is deliberately not
// published as a users host-service method, so the guest must fail closed.
func TestUsersProvisionExternalFailsClosed(t *testing.T) {
	invoked := false
	invoker := func(service string, method string, _ []byte, _ any) error {
		invoked = true
		t.Fatalf("ProvisionExternal must not reach host transport, but called service=%q method=%q", service, method)
		return nil
	}

	client := Users(invoker)

	id, err := client.ProvisionExternal(context.Background(), usercap.ProvisionExternalInput{
		Email:       "someone@example.com",
		DisplayName: "Someone",
		Remark:      "test",
	})
	if err == nil {
		t.Fatal("ProvisionExternal must fail closed for dynamic plugins, got nil error")
	}
	if id != "" {
		t.Fatalf("ProvisionExternal must not return a user ID on fail-closed path, got %q", id)
	}
	if invoked {
		t.Fatal("ProvisionExternal must not invoke the host-service transport")
	}
}
