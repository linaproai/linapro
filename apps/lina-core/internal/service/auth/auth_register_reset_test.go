// This file verifies public self-registration and email password-reset flows.

package auth

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/notifycap"
	"lina-core/pkg/statusflag"
)

// capturingEmailDelivery records outbound recovery emails for tests.
type capturingEmailDelivery struct {
	mu       sync.Mutex
	messages []notifycap.EmailDeliveryInput
}

// Deliver records the message and returns a synthetic provider ID.
func (d *capturingEmailDelivery) Deliver(
	_ context.Context,
	in notifycap.EmailDeliveryInput,
) (*notifycap.EmailDeliveryResult, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.messages = append(d.messages, in)
	return &notifycap.EmailDeliveryResult{ProviderMessageID: "test-msg"}, nil
}

// lastContent returns the body of the latest delivered message.
func (d *capturingEmailDelivery) lastContent() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.messages) == 0 {
		return ""
	}
	return d.messages[len(d.messages)-1].Content
}

// TestRegisterCreatesPlatformUserWithDefaultRole verifies public registration.
func TestRegisterCreatesPlatformUserWithDefaultRole(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	withRuntimeParamValue(t, "sys.auth.registerEnabled", "true")

	username := fmt.Sprintf("selfreg-%d", time.Now().UnixNano())
	email := username + "@example.com"
	out, err := svc.Register(ctx, RegisterInput{
		Username: username,
		Password: "Passw0rd!",
		Email:    email,
		Nickname: "Self Reg",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if out == nil || out.UserID <= 0 {
		t.Fatal("expected created user id")
	}
	t.Cleanup(func() {
		_, _ = dao.SysUserRole.Ctx(ctx).Where(do.SysUserRole{UserId: out.UserID}).Delete()
		_, _ = dao.SysUser.Ctx(ctx).Unscoped().Where(do.SysUser{Id: out.UserID}).Delete()
	})

	loginOut, err := svc.Login(ctx, LoginInput{
		Username:   username,
		Password:   "Passw0rd!",
		ClientType: "web",
	})
	if err != nil {
		t.Fatalf("login after register: %v", err)
	}
	if loginOut.AccessToken == "" && loginOut.PreToken == "" {
		t.Fatal("expected login session after register")
	}

	count, err := dao.SysUserRole.Ctx(ctx).Where(do.SysUserRole{UserId: out.UserID}).Count()
	if err != nil {
		t.Fatalf("count roles: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one default role assignment, got %d", count)
	}
}

// TestRegisterRejectsDuplicateUsername verifies username uniqueness.
func TestRegisterRejectsDuplicateUsername(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	withRuntimeParamValue(t, "sys.auth.registerEnabled", "true")

	username := fmt.Sprintf("selfreg-dup-%d", time.Now().UnixNano())
	email1 := username + "-1@example.com"
	out, err := svc.Register(ctx, RegisterInput{
		Username: username,
		Password: "Passw0rd!",
		Email:    email1,
	})
	if err != nil {
		t.Fatalf("first register: %v", err)
	}
	t.Cleanup(func() {
		_, _ = dao.SysUserRole.Ctx(ctx).Where(do.SysUserRole{UserId: out.UserID}).Delete()
		_, _ = dao.SysUser.Ctx(ctx).Unscoped().Where(do.SysUser{Id: out.UserID}).Delete()
	})

	_, err = svc.Register(ctx, RegisterInput{
		Username: username,
		Password: "Passw0rd!",
		Email:    username + "-2@example.com",
	})
	if !bizerr.Is(err, CodeAuthRegisterUsernameExists) {
		t.Fatalf("expected username exists, got %v", err)
	}
}

// TestPasswordResetFlowWithEmailDelivery verifies request+confirm recovery.
func TestPasswordResetFlowWithEmailDelivery(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	withRuntimeParamValue(t, "sys.auth.forgetPasswordEnabled", "true")

	delivery := &capturingEmailDelivery{}
	notifycap.ResetEmailDeliveryForTest()
	if err := notifycap.ProvideEmailDelivery(delivery); err != nil {
		t.Fatalf("provide email delivery: %v", err)
	}
	t.Cleanup(notifycap.ResetEmailDeliveryForTest)

	username := fmt.Sprintf("reset-%d", time.Now().UnixNano())
	email := username + "@example.com"
	password := "Passw0rd!"
	userID := insertAuthTestUser(t, ctx, username, password)
	if _, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: userID}).
		Data(do.SysUser{Email: email, Status: statusflag.EnabledValue.Int()}).
		Update(); err != nil {
		t.Fatalf("set email: %v", err)
	}

	if err := svc.RequestPasswordReset(ctx, PasswordResetRequestInput{
		Email:             email,
		PublicOrigin:      "http://127.0.0.1:9120",
		WorkspaceBasePath: "/admin",
	}); err != nil {
		t.Fatalf("request password reset: %v", err)
	}

	content := delivery.lastContent()
	if content == "" {
		t.Fatal("expected recovery email content")
	}
	token := extractResetTokenFromContent(content)
	if token == "" {
		t.Fatalf("expected reset token in email content: %q", content)
	}

	newPassword := "N3wPassw0rd!"
	if err := svc.ConfirmPasswordReset(ctx, PasswordResetConfirmInput{
		Token:    token,
		Password: newPassword,
	}); err != nil {
		t.Fatalf("confirm password reset: %v", err)
	}

	if _, err := svc.Login(ctx, LoginInput{
		Username:   username,
		Password:   password,
		ClientType: "web",
	}); err == nil {
		t.Fatal("expected old password to fail after reset")
	}
	if _, err := svc.Login(ctx, LoginInput{
		Username:   username,
		Password:   newPassword,
		ClientType: "web",
	}); err != nil {
		t.Fatalf("login with new password: %v", err)
	}

	// Token is single-use.
	if err := svc.ConfirmPasswordReset(ctx, PasswordResetConfirmInput{
		Token:    token,
		Password: "Another1!",
	}); !bizerr.Is(err, CodeAuthPasswordResetTokenInvalid) {
		t.Fatalf("expected invalid token on reuse, got %v", err)
	}
}

// TestRequestPasswordResetDoesNotEnumerateMissingEmail verifies silent accept.
func TestRequestPasswordResetDoesNotEnumerateMissingEmail(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	withRuntimeParamValue(t, "sys.auth.forgetPasswordEnabled", "true")

	delivery := &capturingEmailDelivery{}
	notifycap.ResetEmailDeliveryForTest()
	if err := notifycap.ProvideEmailDelivery(delivery); err != nil {
		t.Fatalf("provide email delivery: %v", err)
	}
	t.Cleanup(notifycap.ResetEmailDeliveryForTest)

	if err := svc.RequestPasswordReset(ctx, PasswordResetRequestInput{
		Email:             fmt.Sprintf("missing-%d@example.com", time.Now().UnixNano()),
		PublicOrigin:      "http://127.0.0.1:9120",
		WorkspaceBasePath: "/admin",
	}); err != nil {
		t.Fatalf("request password reset for missing email: %v", err)
	}
	if delivery.lastContent() != "" {
		t.Fatal("expected no email for missing account")
	}
}

// extractResetTokenFromContent parses rst_ tokens from recovery email bodies.
func extractResetTokenFromContent(content string) string {
	for _, part := range strings.Fields(content) {
		if idx := strings.Index(part, "token="); idx >= 0 {
			return strings.TrimSpace(part[idx+len("token="):])
		}
	}
	return ""
}
