// This file verifies the host-owned auto-provisioning policy for external
// login: the switch-off fail-closed path, the same-email conflict rejection,
// and the happy path that provisions a least-privilege platform user plus the
// (provider, subject) linkage row inside one transaction. These are
// database-backed tests following the same conventions as the other auth
// integration tests (unique fixtures + t.Cleanup, bizerr code assertions).

package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
	tokencap "lina-core/pkg/plugin/capability/authcap/token"
)

// testExternalProvisioner adapts the real user-owner provisioning service is
// not available in the auth package without an import cycle, so the tests
// bind a minimal provisioner that creates the user row directly with the
// same least-privilege shape and records the received input for assertions.
type testExternalProvisioner struct {
	lastInput ExternalProvisionInput
	failWith  error
}

// ProvisionExternalUser records the input and inserts one minimal platform user.
func (p *testExternalProvisioner) ProvisionExternalUser(ctx context.Context, in ExternalProvisionInput) (int, error) {
	p.lastInput = in
	if p.failWith != nil {
		return 0, p.failWith
	}
	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: fmt.Sprintf("prov-%d", time.Now().UnixNano()),
		Password: "x",
		Nickname: in.DisplayName,
		Email:    in.Email,
		Status:   1,
		TenantId: 0,
	}).InsertAndGetId()
	return int(id), err
}

// cleanupProvisionFixtures removes the user and linkage rows created by one test.
func cleanupProvisionFixtures(t *testing.T, ctx context.Context, provider string, subject string, email string) {
	t.Helper()
	if _, err := dao.SysUserExternalIdentity.Ctx(ctx).
		Unscoped().
		Where(do.SysUserExternalIdentity{Provider: provider, Subject: subject}).
		Delete(); err != nil {
		t.Fatalf("cleanup linkage: %v", err)
	}
	if _, err := dao.SysUser.Ctx(ctx).
		Unscoped().
		Where(do.SysUser{Email: email}).
		Delete(); err != nil {
		t.Fatalf("cleanup user: %v", err)
	}
}

// TestExternalLoginAutoProvisionDisabledFailsClosed verifies that an unlinked
// identity is rejected with not-provisioned when AllowAutoProvision is unset,
// even when a provisioner is bound.
func TestExternalLoginAutoProvisionDisabledFailsClosed(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	svc.BindExternalProvisioner(&testExternalProvisioner{})

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
}

// TestExternalLoginAutoProvisionEmailConflict verifies that an unlinked
// identity whose email already belongs to a local account is rejected with the
// email-conflict code instead of silently linking or provisioning.
func TestExternalLoginAutoProvisionEmailConflict(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	svc.BindExternalProvisioner(&testExternalProvisioner{})

	username := fmt.Sprintf("conflict-user-%d", time.Now().UnixNano())
	email := username + "@example.com"
	userID := insertAuthTestUser(t, ctx, username, "admin123")
	if _, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: userID}).
		Data(do.SysUser{Email: email}).
		Update(); err != nil {
		t.Fatalf("set fixture email: %v", err)
	}

	_, err := svc.LoginByExternalIdentity(ctx, ExternalLoginInput{
		Provider:           "google",
		Subject:            fmt.Sprintf("conflict-%d", time.Now().UnixNano()),
		Email:              email,
		ClientType:         tokencap.ClientTypeWeb,
		AllowAutoProvision: true,
	})
	if !bizerr.Is(err, CodeAuthExternalEmailConflict) {
		t.Fatalf("expected email-conflict, got %v", err)
	}
}

// TestExternalLoginAutoProvisionCreatesUserAndLinkage verifies the happy path:
// switch on, unlinked identity, unused email -> user provisioned, linkage row
// written, token pair issued for the new platform user.
func TestExternalLoginAutoProvisionCreatesUserAndLinkage(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	provisioner := &testExternalProvisioner{}
	svc.BindExternalProvisioner(provisioner)

	provider := "google"
	subject := fmt.Sprintf("prov-sub-%d", time.Now().UnixNano())
	email := fmt.Sprintf("prov-%d@example.com", time.Now().UnixNano())
	t.Cleanup(func() { cleanupProvisionFixtures(t, ctx, provider, subject, email) })

	out, err := svc.LoginByExternalIdentity(ctx, ExternalLoginInput{
		Provider:           provider,
		Subject:            subject,
		Email:              email,
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
	if provisioner.lastInput.Email != email {
		t.Fatalf("provisioner received email %q, want %q", provisioner.lastInput.Email, email)
	}
	// The linkage row must exist and reference the provisioned user.
	var linkage *entity.SysUserExternalIdentity
	if err = dao.SysUserExternalIdentity.Ctx(ctx).
		Where(do.SysUserExternalIdentity{Provider: provider, Subject: subject}).
		Scan(&linkage); err != nil {
		t.Fatalf("scan linkage: %v", err)
	}
	if linkage == nil || linkage.UserId == 0 {
		t.Fatalf("expected linkage row for provisioned user, got %#v", linkage)
	}
	if linkage.EmailSnapshot != email {
		t.Fatalf("linkage email snapshot = %q, want %q", linkage.EmailSnapshot, email)
	}
}
