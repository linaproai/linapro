// auth_external_identity.go implements external-identity login: it resolves a
// source-plugin-verified external identity (provider + immutable subject) to a
// linked local account and issues a host session. The flow deliberately reuses
// the shared login-IP policy, disabled-account check, tenant resolution,
// pre-login-token handoff, token issuance, session persistence, and auth hooks
// owned by the auth service so external login and password login stay
// behavior-compatible. Provisioning is host-owned and closed by default; the
// host never creates a local account from an external identity.

package auth

import (
	"context"
	"strings"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginhost"
	"lina-core/pkg/statusflag"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/mssola/useragent"
)

// tryAutoProvision runs the host-owned auto-provisioning policy for one
// unlinked verified identity. It returns the provisioned user ID, or a
// caller-visible bizerr:
//
//   - AllowAutoProvision unset or no provisioner bound -> not-provisioned
//     (the historical fail-closed behavior).
//   - An enabled local account already uses the same email -> email-conflict.
//     The identity is NOT linked automatically; the user must sign in to the
//     existing account and link the identity through an authenticated
//     confirmation flow. This blocks silent account takeover through an IdP
//     email assertion.
//   - Otherwise the user owner provisions a least-privilege platform user and
//     the (provider, subject) linkage row is recorded inside one transaction.
func (s *serviceImpl) tryAutoProvision(ctx context.Context, in ExternalLoginInput) (int, error) {
	if !in.AllowAutoProvision || s == nil || s.provisioner == nil {
		return 0, bizerr.NewCode(CodeAuthExternalUserNotProvisioned)
	}
	email := strings.TrimSpace(in.Email)
	if email == "" {
		return 0, bizerr.NewCode(CodeAuthExternalUserNotProvisioned)
	}
	emailCount, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Email: email}).
		Count()
	if err != nil {
		return 0, err
	}
	if emailCount > 0 {
		return 0, bizerr.NewCode(CodeAuthExternalEmailConflict)
	}
	var provisionedUserID int
	if err = dao.SysUserExternalIdentity.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		userID, provisionErr := s.provisioner.ProvisionExternalUser(ctx, ExternalProvisionInput{
			Email:       email,
			DisplayName: in.DisplayName,
			Remark:      "auto-provisioned by external login provider " + in.Provider,
		})
		if provisionErr != nil {
			return provisionErr
		}
		if _, linkErr := dao.SysUserExternalIdentity.Ctx(ctx).Data(do.SysUserExternalIdentity{
			UserId:        userID,
			Provider:      in.Provider,
			Subject:       strings.TrimSpace(in.Subject),
			PluginId:      in.PluginID,
			EmailSnapshot: email,
		}).Insert(); linkErr != nil {
			return linkErr
		}
		provisionedUserID = userID
		return nil
	}); err != nil {
		return 0, err
	}
	logger.Infof(ctx, "external login auto-provisioned provider=%s subject=%s userId=%d", in.Provider, in.Subject, provisionedUserID)
	return provisionedUserID, nil
}

// LoginByExternalIdentity resolves a verified external identity to a linked
// local account and issues a host session. See the Service interface for the
// full contract. The caller must have already verified the identity; this
// method performs no OAuth or token exchange.
func (s *serviceImpl) LoginByExternalIdentity(ctx context.Context, in ExternalLoginInput) (*ExternalLoginOutput, error) {
	provider := strings.TrimSpace(in.Provider)
	subject := strings.TrimSpace(in.Subject)
	if provider == "" || subject == "" {
		return nil, bizerr.NewCode(CodeAuthExternalIdentityInvalid)
	}

	clientType, err := ParseClientType(in.ClientType.String())
	if err != nil {
		return nil, err
	}

	// Extract client info for login audit and session metadata, mirroring the
	// password Login path so external logins produce comparable hook payloads.
	var ip, browser, osName string
	if r := g.RequestFromCtx(ctx); r != nil {
		ip = r.GetClientIp()
		ua := useragent.New(r.GetHeader("User-Agent"))
		browserName, browserVersion := ua.Browser()
		browser = browserName + " " + browserVersion
		osName = ua.OS()
	}

	// dispatchExternalLoginFailed publishes a login-failed hook. The username is
	// best-effort: the resolved account username when known, otherwise the
	// captured email used only for audit context.
	dispatchExternalLoginFailed := func(username string, message string, reason string) {
		s.dispatchAuthHookEvent(ctx, pluginhost.ExtensionPointAuthLoginFailed, pluginhost.AuthHookPayloadInput{
			UserName:   username,
			Status:     authLoginStatusFail,
			IP:         ip,
			ClientType: clientType.String(),
			Browser:    browser,
			OS:         osName,
			Message:    message,
			Reason:     reason,
		}, "plugin external login failed hook failed")
	}

	blacklisted, err := s.configSvc.IsLoginIPBlacklisted(ctx, ip)
	if err != nil {
		return nil, err
	}
	if blacklisted {
		dispatchExternalLoginFailed(in.Email, authEventMessageIPBlacklisted, pluginhost.AuthHookReasonIPBlacklisted)
		return nil, bizerr.NewCode(CodeAuthIPBlacklisted)
	}

	// Resolve the linked local user through the authoritative (provider,
	// subject) unique key. Missing linkage returns a uniform not-provisioned
	// error so external login never leaks whether the captured email exists as
	// another account.
	var identity *entity.SysUserExternalIdentity
	if err = dao.SysUserExternalIdentity.Ctx(ctx).
		Where(do.SysUserExternalIdentity{Provider: provider, Subject: subject}).
		Scan(&identity); err != nil {
		return nil, err
	}
	if identity == nil {
		provisionedUserID, provisionErr := s.tryAutoProvision(ctx, in)
		if provisionErr != nil {
			dispatchExternalLoginFailed(in.Email, authEventMessageExternalNotProvisioned, authHookReasonExternalNotProvisioned)
			return nil, provisionErr
		}
		identity = &entity.SysUserExternalIdentity{UserId: provisionedUserID}
	}

	var user *entity.SysUser
	if err = dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: identity.UserId}).
		Scan(&user); err != nil {
		return nil, err
	}
	if user == nil {
		dispatchExternalLoginFailed(in.Email, authEventMessageExternalNotProvisioned, authHookReasonExternalNotProvisioned)
		return nil, bizerr.NewCode(CodeAuthExternalUserNotProvisioned)
	}
	if user.Status == statusflag.Disabled.Int() {
		dispatchExternalLoginFailed(user.Username, authEventMessageUserDisabled, pluginhost.AuthHookReasonUserDisabled)
		return nil, bizerr.NewCode(CodeAuthUserDisabled)
	}

	tenantSvcAvailable := s.tenantSvc != nil && s.tenantSvc.Available(ctx)
	tenants, err := s.loginTenants(ctx, user.Id)
	if err != nil {
		return nil, err
	}
	if user.TenantId != int(tenantcap.PLATFORM) && (!tenantSvcAvailable || len(tenants) == 0) {
		dispatchExternalLoginFailed(user.Username, authEventMessageTenantUnavailable, authHookReasonTenantUnavailable)
		return nil, bizerr.NewCode(CodeAuthTenantUnavailable)
	}
	if len(tenants) > 1 {
		preToken, err := s.preTokens.Create(ctx, preTokenRecord{
			UserID:     user.Id,
			Username:   user.Username,
			Status:     user.Status,
			ClientType: clientType,
		})
		if err != nil {
			return nil, bizerr.WrapCode(err, CodeAuthTokenStateUnavailable)
		}
		return &ExternalLoginOutput{PreToken: preToken, Tenants: tenants}, nil
	}

	tenantID := int(tenantcap.PLATFORM)
	if len(tenants) == 1 {
		tenantID = tenants[0].Id
	}

	accessToken, refreshToken, tokenID, err := s.generateTokenPair(ctx, user, tenantID, clientType)
	if err != nil {
		return nil, err
	}

	loginDate := time.Now()
	if _, err = dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: user.Id}).
		Data(do.SysUser{LoginDate: &loginDate}).
		Update(); err != nil {
		return nil, bizerr.WrapCode(err, CodeAuthLoginStateUpdateFailed)
	}

	if err = s.createSession(ctx, user, tenantID, tokenID, clientType); err != nil {
		logger.Warningf(ctx, "create external login session failed tokenId=%s err=%v", tokenID, err)
	}

	s.dispatchAuthHookEvent(ctx, pluginhost.ExtensionPointAuthLoginSucceeded, pluginhost.AuthHookPayloadInput{
		UserName:   user.Username,
		Status:     authLoginStatusSuccess,
		IP:         ip,
		ClientType: clientType.String(),
		Browser:    browser,
		OS:         osName,
		Message:    authEventMessageLoginSuccessful,
		Reason:     pluginhost.AuthHookReasonLoginSuccessful,
	}, "plugin external login succeeded hook failed")

	return &ExternalLoginOutput{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
