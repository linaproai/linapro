// This file verifies the guest-side user capability routes CreateFromExternal
// through the host-service transport under the same-trust model for installed
// dynamic plugins.

package domainhostcall

import (
	"context"
	"testing"

	"lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestUsersCreateFromExternalInvokesHostTransport verifies the dynamic users
// guest client publishes CreateFromExternal over the host-service transport.
func TestUsersCreateFromExternalInvokesHostTransport(t *testing.T) {
	var gotService, gotMethod string
	invoker := func(service string, method string, _ []byte, out any) error {
		gotService = service
		gotMethod = method
		if id, ok := out.(*usercap.UserID); ok {
			*id = usercap.UserID("42")
		}
		return nil
	}

	client := Users(invoker)
	id, err := client.CreateFromExternal(context.Background(), usercap.CreateFromExternalInput{
		Email:       "someone@example.com",
		DisplayName: "Someone",
		Remark:      "test",
	})
	if err != nil {
		t.Fatalf("CreateFromExternal: %v", err)
	}
	if id != "42" {
		t.Fatalf("expected user id 42, got %q", id)
	}
	if gotService != protocol.HostServiceUsers {
		t.Fatalf("expected service %q, got %q", protocol.HostServiceUsers, gotService)
	}
	if gotMethod != protocol.HostServiceMethodUsersCreateFromExternal {
		t.Fatalf("expected method %q, got %q", protocol.HostServiceMethodUsersCreateFromExternal, gotMethod)
	}
}
