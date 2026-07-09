// This file verifies external-identity login resolves verified identities to
// linked local accounts through the bound provider seam, stays fail-closed
// without a provider, and reuses the shared login policy. Linkage storage is
// provider-owned (linapro-oidc-core), so these tests bind a mock
// externalidentityspi.Provider; only sys_user fixtures touch the database.
// Assertions use stable bizerr codes rather than localized text so they do not
// depend on i18n resource loading.

package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/authcap/externallogin/externalidentityspi"
	tokencap "lina-core/pkg/plugin/capability/authcap/token"
	"lina-core/pkg/statusflag"
)

// mockIdentityProvider is a test double for the provider seam. It resolves
// from an in-memory linkage map and records provisioning calls.
type mockIdentityProvider struct {
	links          map[string]int
	resolveErr     error
	provisionErr   error
	provisionID    int
	provisionCalls []externalidentityspi.ProvisionInput
}

// linkKey builds the in-memory (provider, subject) resolution key.
func linkKey(provider string, subject string) string {
	return provider + "\x00" + subject
}

// Resolve maps (provider, subject) through the in-memory linkage map.
func (m *mockIdentityProvider) Resolve(_ context.Context, in externalidentityspi.ResolveInput) (int, bool, error) {
	if m.resolveErr != nil {
		return 0, false, m.resolveErr
	}
	userID, found := m.links[linkKey(in.Provider, in.Subject)]
	return userID, found, nil
}

// Provision records the call and returns the configured outcome.
func (m *mockIdentityProvider) Provision(_ context.Context, in externalidentityspi.ProvisionInput) (int, error) {
	m.provisionCalls = append(m.provisionCalls, in)
	if m.provisionErr != nil {
		return 0, m.provisionErr
	}
	return m.provisionID, nil
}

// Bind is unused by external-login tests.
func (m *mockIdentityProvider) Bind(context.Context, externalidentityspi.BindInput) error {
	return nil
}

// Unbind is unused by external-login tests.
func (m *mockIdentityProvider) Unbind(context.Context, externalidentityspi.UnbindInput) error {
	return nil
}

// List is unused by external-login tests.
func (m *mockIdentityProvider) List(context.Context, int) ([]externalidentityspi.BoundIdentity, error) {
	return nil, nil
}

// disableAuthTestUser flips one seeded user to the disabled status.
func disableAuthTestUser(t *testing.T, ctx context.Context, userID int) {
	t.Helper()
	if _, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: userID}).
		Data(do.SysUser{Status: statusflag.Disabled.Int()}).
		Update(); err != nil {
		t.Fatalf("disable test user: %v", err)
	}
}

// TestLoginByExternalIdentityRejectsEmptyProviderOrSubject verifies an external
// login with an empty provider or subject fails before any provider call.
func TestLoginByExternalIdentityRejectsEmptyProviderOrSubject(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()

	if _, err := svc.LoginByExternalIdentity(ctx, ExternalLoginInput{
		Provider:   "",
		Subject:    "sub-1",
		ClientType: tokencap.ClientTypeWeb,
	}); !bizerr.Is(err, CodeAuthExternalIdentityInvalid) {
		t.Fatalf("empty provider: expected identity-invalid, got %v", err)
	}
	if _, err := svc.LoginByExternalIdentity(ctx, ExternalLoginInput{
		Provider:   "google",
		Subject:    "  ",
		ClientType: tokencap.ClientTypeWeb,
	}); !bizerr.Is(err, CodeAuthExternalIdentityInvalid) {
		t.Fatalf("blank subject: expected identity-invalid, got %v", err)
	}
}

// TestLoginByExternalIdentityFailsClosedWithoutProvider verifies external login
// rejects every identity when no provider is bound: the provider plugin being
// absent or disabled must not resolve linkages or mint accounts.
func TestLoginByExternalIdentityFailsClosedWithoutProvider(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()

	_, err := svc.LoginByExternalIdentity(ctx, ExternalLoginInput{
		Provider:           "google",
		Subject:            fmt.Sprintf("noprovider-%d", time.Now().UnixNano()),
		Email:              "noprovider@example.com",
		ClientType:         tokencap.ClientTypeWeb,
		AllowAutoProvision: true,
	})
	if !bizerr.Is(err, CodeAuthExternalUserNotProvisioned) {
		t.Fatalf("expected not-provisioned, got %v", err)
	}
}

// TestLoginByExternalIdentityRejectsUnprovisionedIdentity verifies a verified
// identity with no linkage is rejected without provisioning when
// AllowAutoProvision is unset.
func TestLoginByExternalIdentityRejectsUnprovisionedIdentity(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	provider := &mockIdentityProvider{links: map[string]int{}}
	svc.BindExternalIdentityProvider(provider)

	_, err := svc.LoginByExternalIdentity(ctx, ExternalLoginInput{
		Provider:   "google",
		Subject:    fmt.Sprintf("unlinked-%d", time.Now().UnixNano()),
		Email:      "unlinked@example.com",
		ClientType: tokencap.ClientTypeWeb,
	})
	if !bizerr.Is(err, CodeAuthExternalUserNotProvisioned) {
		t.Fatalf("expected not-provisioned, got %v", err)
	}
	if len(provider.provisionCalls) != 0 {
		t.Fatalf("expected no provisioning call, got %d", len(provider.provisionCalls))
	}
}

// TestLoginByExternalIdentityIssuesTokenPairForLinkedUser verifies a resolved
// linkage enters the unchanged token-minting flow and receives a token pair
// whose claims carry the resolved user.
func TestLoginByExternalIdentityIssuesTokenPairForLinkedUser(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()

	username := fmt.Sprintf("external-login-user-%d", time.Now().UnixNano())
	userID := insertAuthTestUser(t, ctx, username, "admin123")
	provider := "google"
	subject := fmt.Sprintf("sub-%d", time.Now().UnixNano())
	svc.BindExternalIdentityProvider(&mockIdentityProvider{
		links: map[string]int{linkKey(provider, subject): userID},
	})

	out, err := svc.LoginByExternalIdentity(ctx, ExternalLoginInput{
		Provider:    provider,
		Subject:     subject,
		Email:       "linked@example.com",
		DisplayName: "Linked User",
		ClientType:  tokencap.ClientTypeWeb,
	})
	if err != nil {
		t.Fatalf("external login: %v", err)
	}
	if out.AccessToken == "" || out.RefreshToken == "" {
		t.Fatalf("expected token pair, got %#v", out)
	}
	if out.PreToken != "" {
		t.Fatalf("platform user should not receive a pre-token, got %q", out.PreToken)
	}
	claims, err := svc.parseAccessTokenForTest(ctx, out.AccessToken)
	if err != nil {
		t.Fatalf("parse external login token: %v", err)
	}
	if claims.UserId != userID {
		t.Fatalf("expected token for user %d, got %d", userID, claims.UserId)
	}
	if claims.ClientType != tokencap.ClientTypeWeb {
		t.Fatalf("expected web client type, got %q", claims.ClientType)
	}
}

// TestLoginByExternalIdentityRejectsDisabledUser verifies a linked but disabled
// account cannot sign in through external login.
func TestLoginByExternalIdentityRejectsDisabledUser(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()

	username := fmt.Sprintf("external-disabled-user-%d", time.Now().UnixNano())
	userID := insertAuthTestUser(t, ctx, username, "admin123")
	provider := "discord"
	subject := fmt.Sprintf("sub-%d", time.Now().UnixNano())
	svc.BindExternalIdentityProvider(&mockIdentityProvider{
		links: map[string]int{linkKey(provider, subject): userID},
	})
	disableAuthTestUser(t, ctx, userID)

	_, err := svc.LoginByExternalIdentity(ctx, ExternalLoginInput{
		Provider:   provider,
		Subject:    subject,
		ClientType: tokencap.ClientTypeWeb,
	})
	if !bizerr.Is(err, CodeAuthUserDisabled) {
		t.Fatalf("expected user-disabled, got %v", err)
	}
}
