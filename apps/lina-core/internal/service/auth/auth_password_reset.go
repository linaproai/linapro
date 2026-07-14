// This file implements email-based password recovery for public accounts.

package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/notifycap"
	"lina-core/pkg/statusflag"
)

// RequestPasswordReset starts email password recovery when the feature is enabled.
// On acceptance it always returns nil so callers cannot enumerate accounts by email.
func (s *serviceImpl) RequestPasswordReset(ctx context.Context, in PasswordResetRequestInput) error {
	publicCfg, err := s.configSvc.GetPublicFrontend(ctx)
	if err != nil {
		return err
	}
	if publicCfg == nil || !publicCfg.Auth.ForgetPasswordEnabled {
		return bizerr.NewCode(CodeAuthForgetPasswordDisabled)
	}

	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" {
		return bizerr.NewCode(CodeAuthClientTypeInvalid)
	}

	if s.rateLimit != nil {
		ok, limitErr := s.rateLimit.Allow(ctx, "forget:"+email, forgetPasswordRateLimitMax)
		if limitErr != nil {
			return bizerr.WrapCode(limitErr, CodeAuthTokenStateUnavailable)
		}
		if !ok {
			return bizerr.NewCode(CodeAuthForgetPasswordRateLimited)
		}
	}

	delivery := notifycap.EmailDeliveryOrNil()
	if delivery == nil {
		return bizerr.NewCode(CodeAuthPasswordResetUnavailable)
	}

	// Constant-ish delay reduces simple timing enumeration.
	start := time.Now()
	defer func() {
		const minLatency = 200 * time.Millisecond
		if elapsed := time.Since(start); elapsed < minLatency {
			time.Sleep(minLatency - elapsed)
		}
	}()

	user, err := s.findEnabledUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		// Accepted without sending so account existence is not leaked.
		return nil
	}

	token, err := s.resetTokens.Create(ctx, passwordResetRecord{
		UserID: user.Id,
		Email:  email,
	})
	if err != nil {
		return bizerr.WrapCode(err, CodeAuthTokenStateUnavailable)
	}

	var (
		resetURL = buildPasswordResetURL(in.PublicOrigin, in.WorkspaceBasePath, token)
		subject  = "Reset your LinaPro password"
		content  = fmt.Sprintf(
			"You requested a password reset for account %s.\n\nOpen this link within 30 minutes to set a new password:\n%s\n\nIf you did not request this, you can ignore this email.",
			user.Username,
			resetURL,
		)
	)
	if _, err = delivery.Deliver(ctx, notifycap.EmailDeliveryInput{
		AccountID: 0,
		To:        []string{email},
		Subject:   subject,
		Content:   content,
	}); err != nil {
		logger.Warningf(ctx, "password reset email delivery failed userId=%d: %v", user.Id, err)
		return bizerr.NewCode(CodeAuthPasswordResetUnavailable)
	}

	logger.Infof(ctx, "password reset email accepted userId=%d", user.Id)
	return nil
}

// ConfirmPasswordReset consumes a one-time reset token and updates the password.
func (s *serviceImpl) ConfirmPasswordReset(ctx context.Context, in PasswordResetConfirmInput) error {
	publicCfg, err := s.configSvc.GetPublicFrontend(ctx)
	if err != nil {
		return err
	}
	if publicCfg == nil || !publicCfg.Auth.ForgetPasswordEnabled {
		return bizerr.NewCode(CodeAuthForgetPasswordDisabled)
	}

	token := strings.TrimSpace(in.Token)
	password := strings.TrimSpace(in.Password)
	if token == "" || password == "" {
		return bizerr.NewCode(CodeAuthPasswordResetTokenInvalid)
	}

	record, ok, err := s.resetTokens.Consume(ctx, token)
	if err != nil {
		return bizerr.WrapCode(err, CodeAuthTokenStateUnavailable)
	}
	if !ok || record.UserID <= 0 {
		return bizerr.NewCode(CodeAuthPasswordResetTokenInvalid)
	}

	var user *entity.SysUser
	if err = dao.SysUser.Ctx(ctx).Where(do.SysUser{Id: record.UserID}).Scan(&user); err != nil {
		return err
	}
	if user == nil || user.Status != statusflag.EnabledValue.Int() {
		return bizerr.NewCode(CodeAuthPasswordResetTokenInvalid)
	}

	hash, err := s.HashPassword(password)
	if err != nil {
		return err
	}
	if _, err = dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: user.Id}).
		Data(do.SysUser{Password: hash}).
		Update(); err != nil {
		return err
	}

	// Best-effort session cleanup for platform sessions of this user.
	if s.sessionStore != nil {
		if delErr := s.sessionStore.DeleteByUserId(ctx, 0, user.Id); delErr != nil {
			logger.Warningf(ctx, "password reset session cleanup failed userId=%d: %v", user.Id, delErr)
		}
	}

	logger.Infof(ctx, "password reset completed userId=%d", user.Id)
	return nil
}

// findEnabledUserByEmail returns one enabled account matching the email.
// When multiple rows exist, the lowest ID is chosen for stable recovery.
func (s *serviceImpl) findEnabledUserByEmail(ctx context.Context, email string) (*entity.SysUser, error) {
	var user *entity.SysUser
	err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Email: email, Status: statusflag.EnabledValue.Int()}).
		OrderAsc(dao.SysUser.Columns().Id).
		Limit(1).
		Scan(&user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// buildPasswordResetURL composes the SPA recovery URL.
func buildPasswordResetURL(publicOrigin, workspaceBasePath, token string) string {
	origin := strings.TrimRight(strings.TrimSpace(publicOrigin), "/")
	basePath := strings.TrimSpace(workspaceBasePath)
	if basePath == "" {
		basePath = "/admin"
	}
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	basePath = strings.TrimRight(basePath, "/")
	if origin == "" {
		return fmt.Sprintf("%s/auth/reset-password?token=%s", basePath, token)
	}
	return fmt.Sprintf("%s%s/auth/reset-password?token=%s", origin, basePath, token)
}
