// This file verifies the auto-provisioning delegation for external login: the
// switch-off fail-closed path, provider policy error pass-through, the
// provider-unavailable sentinel mapping, and the happy path where a
// provider-provisioned account enters the unchanged token-minting flow.
// Provisioning policy (same-email conflict, anchor derivation, linkage
// de-duplication) is provider-owned and covered by linapro-oidc-core's own
// tests; these tests bind a mock provider and only seed sys_user fixtures.

package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/authcap/externallogin/externalidentityspi"
	tokencap "lina-core/pkg/plugin/capability/authcap/token"

	"github.com/gogf/gf/v2/errors/gcode"
)

// codeAuthTestProviderPolicyRejected simulates a provider-owned policy
// rejection (for example the plugin's same-email conflict) crossing the seam.
var codeAuthTestProviderPolicyRejected = bizerr.MustDefine(
	"AUTH_TEST_PROVIDER_POLICY_REJECTED",
	"Test provider policy rejection",
	gcode.CodeInvalidOperation,
)

// TestExternalLoginAutoProvisionDisabledFailsClosed verifies that an unlinked
// identity is rejected with not-provisioned when AllowAutoProvision is unset,
// even when a provider is bound, and that the provider is never asked to
// provision.
func TestExternalLoginAutoProvisionDisabledFailsClosed(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	provider := &mockIdentityProvider{links: map[string]int{}, provisionID: 1}
	svc.BindExternalIdentityProvider(provider)

	_, err := svc.LoginByExternalIdentity(ctx, ExternalLoginInput{
		Provider:   "google",
		Subject:    fmt.Sprintf("noprov-%d", time.Now().UnixNano()),
		Email:      "noprov@example.com",
		ClientType: tokencap.ClientTypeWeb,
		// AllowAutoProvision deliberately unset.
	})
	if !bizerr.Is(err, CodeAuthExternalUserNotProvisioned) {
		t.Fatalf("expected not-provisioned, got %v", err)
	}
	if len(provider.provisionCalls) != 0 {
		t.Fatalf("expected no provisioning call, got %d", len(provider.provisionCalls))
	}
}

// TestExternalLoginProviderPolicyErrorPassesThrough verifies that a
// provider-owned policy rejection (for example the plugin's same-email
// conflict) surfaces to the caller unchanged, and that the provider received
// the verified identity fields.
func TestExternalLoginProviderPolicyErrorPassesThrough(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	provider := &mockIdentityProvider{
		links:        map[string]int{},
		provisionErr: bizerr.NewCode(codeAuthTestProviderPolicyRejected),
	}
	svc.BindExternalIdentityProvider(provider)

	subject := fmt.Sprintf("conflict-%d", time.Now().UnixNano())
	email := "conflict@example.com"
	_, err := svc.LoginByExternalIdentity(ctx, ExternalLoginInput{
		PluginID:           "linapro-oidc-google",
		Provider:           "google",
		Subject:            subject,
		Email:              email,
		ClientType:         tokencap.ClientTypeWeb,
		AllowAutoProvision: true,
	})
	if !bizerr.Is(err, codeAuthTestProviderPolicyRejected) {
		t.Fatalf("expected provider policy rejection to pass through, got %v", err)
	}
	if len(provider.provisionCalls) != 1 {
		t.Fatalf("expected one provisioning call, got %d", len(provider.provisionCalls))
	}
	call := provider.provisionCalls[0]
	if call.Provider != "google" || call.Subject != subject || call.Email != email {
		t.Fatalf("provider received unexpected provisioning input: %#v", call)
	}
	if call.PluginID != "linapro-oidc-google" {
		t.Fatalf("expected host-stamped plugin ID, got %q", call.PluginID)
	}
	if !call.AllowAutoProvision {
		t.Fatal("expected AllowAutoProvision to be forwarded as true")
	}
}

// TestExternalLoginProviderUnavailableMapsToNotProvisioned verifies that the
// provider-unavailable sentinel (no enabled provider plugin) keeps the
// historical uniform fail-closed outcome.
func TestExternalLoginProviderUnavailableMapsToNotProvisioned(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	svc.BindExternalIdentityProvider(&mockIdentityProvider{
		links:        map[string]int{},
		provisionErr: externalidentityspi.ErrProviderUnavailable,
	})

	_, err := svc.LoginByExternalIdentity(ctx, ExternalLoginInput{
		Provider:           "google",
		Subject:            fmt.Sprintf("unavailable-%d", time.Now().UnixNano()),
		Email:              "unavailable@example.com",
		ClientType:         tokencap.ClientTypeWeb,
		AllowAutoProvision: true,
	})
	if !bizerr.Is(err, CodeAuthExternalUserNotProvisioned) {
		t.Fatalf("expected not-provisioned, got %v", err)
	}
}

// TestExternalLoginAutoProvisionIssuesSession verifies the happy path: switch
// on, unlinked identity, provider provisions an account, and the resolved user
// enters the unchanged token-minting flow.
func TestExternalLoginAutoProvisionIssuesSession(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()

	username := fmt.Sprintf("prov-user-%d", time.Now().UnixNano())
	userID := insertAuthTestUser(t, ctx, username, "admin123")
	provider := &mockIdentityProvider{links: map[string]int{}, provisionID: userID}
	svc.BindExternalIdentityProvider(provider)

	out, err := svc.LoginByExternalIdentity(ctx, ExternalLoginInput{
		Provider:           "google",
		Subject:            fmt.Sprintf("prov-sub-%d", time.Now().UnixNano()),
		Email:              fmt.Sprintf("prov-%d@example.com", time.Now().UnixNano()),
		DisplayName:        "Provisioned User",
		ClientType:         tokencap.ClientTypeWeb,
		AllowAutoProvision: true,
	})
	if err != nil {
		t.Fatalf("auto-provision login: %v", err)
	}
	if out.AccessToken == "" || out.RefreshToken == "" {
		t.Fatalf("expected token pair for provisioned user, got %#v", out)
	}
	claims, err := svc.parseAccessTokenForTest(ctx, out.AccessToken)
	if err != nil {
		t.Fatalf("parse provisioned login token: %v", err)
	}
	if claims.UserId != userID {
		t.Fatalf("expected token for provisioned user %d, got %d", userID, claims.UserId)
	}
	if len(provider.provisionCalls) != 1 {
		t.Fatalf("expected one provisioning call, got %d", len(provider.provisionCalls))
	}
}
