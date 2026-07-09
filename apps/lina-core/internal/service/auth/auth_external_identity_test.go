// This file verifies external-identity login resolves verified identities to
// linked local accounts, enforces host-owned provisioning, and reuses the
// shared login policy. These are database-backed tests: they seed sys_user and
// sys_user_external_identity rows and assert on stable bizerr codes rather than
// localized text so they do not depend on i18n resource loading.

package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/pkg/bizerr"
	tokencap "lina-core/pkg/plugin/capability/authcap/token"
	"lina-core/pkg/statusflag"
)

// insertAuthTestExternalIdentity seeds one (provider, subject) -> userID linkage
// row and removes it during test cleanup.
func insertAuthTestExternalIdentity(t *testing.T, ctx context.Context, userID int, provider string, subject string) {
	t.Helper()
	id, err := dao.SysUserExternalIdentity.Ctx(ctx).Data(do.SysUserExternalIdentity{
		UserId:   userID,
		Provider: provider,
		Subject:  subject,
		PluginId: "linapro-oidc-test",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert external identity: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := dao.SysUserExternalIdentity.Ctx(ctx).
			Unscoped().
			Where(do.SysUserExternalIdentity{Id: int(id)}).
			Delete(); cleanupErr != nil {
			t.Fatalf("cleanup external identity: %v", cleanupErr)
		}
	})
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
// login with an empty provider or subject fails before any database lookup.
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

// TestLoginByExternalIdentityRejectsUnprovisionedIdentity verifies a verified
// identity with no linkage row is rejected without provisioning.
func TestLoginByExternalIdentityRejectsUnprovisionedIdentity(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()

	_, err := svc.LoginByExternalIdentity(ctx, ExternalLoginInput{
		Provider:   "google",
		Subject:    fmt.Sprintf("unlinked-%d", time.Now().UnixNano()),
		Email:      "unlinked@example.com",
		ClientType: tokencap.ClientTypeWeb,
	})
	if !bizerr.Is(err, CodeAuthExternalUserNotProvisioned) {
		t.Fatalf("expected not-provisioned, got %v", err)
	}
}

// TestLoginByExternalIdentityIssuesTokenPairForLinkedUser verifies a linked
// platform user receives a token pair whose claims carry the resolved user.
func TestLoginByExternalIdentityIssuesTokenPairForLinkedUser(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()

	username := fmt.Sprintf("external-login-user-%d", time.Now().UnixNano())
	userID := insertAuthTestUser(t, ctx, username, "admin123")
	provider := "google"
	subject := fmt.Sprintf("sub-%d", time.Now().UnixNano())
	insertAuthTestExternalIdentity(t, ctx, userID, provider, subject)

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
	insertAuthTestExternalIdentity(t, ctx, userID, provider, subject)
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
