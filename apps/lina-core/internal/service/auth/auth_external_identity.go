// auth_external_identity.go implements external-identity login and the
// post-startup provider binding used by that flow. It resolves a
// source-plugin-verified external identity (provider + immutable subject) to a
// linked local account through the bound external-identity provider and issues
// a host session. The flow deliberately reuses the shared login-IP policy,
// disabled-account check, tenant resolution, pre-login-token handoff, token
// issuance, session persistence, and auth hooks owned by the auth service so
// external login and password login stay behavior-compatible. Linkage storage
// and provisioning policy are provider-owned (linapro-extlogin-core); when no
// provider plugin is installed and enabled, external login is fail-closed: no
// linkage resolves, no account is created, and no session is issued.
//
// Auth is constructed before the plugin provider managers finish wiring, so the
// provider cannot be a constructor parameter; runtime assembly calls
// BindExternalIdentityProvider once with the host manager-backed service.

package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/authcap/extlogin/extidspi"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginhost"
	"lina-core/pkg/statusflag"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/mssola/useragent"
)

// BindExternalIdentityProvider attaches the source-plugin external-identity
// provider seam. It is called once from runtime assembly with the host
// manager-backed service, which lazily resolves the enabled provider plugin
// (linapro-extlogin-core) per call. A nil provider keeps external login
// fail-closed.
func (s *serviceImpl) BindExternalIdentityProvider(provider extidspi.Provider) {
	if s == nil {
		return
	}
	s.identityProvider = provider
}

// resolveExternalUserID resolves a verified external identity to a local user
// ID through the bound provider seam. A missing provider, missing linkage, or
// disallowed auto-provisioning uniformly yields the fail-closed
// not-provisioned outcome so external login never leaks account existence.
// Provisioning policy (same-email conflict, email-less anchor derivation,
// idempotent (provider, subject) de-duplication) is provider-owned; policy
// rejections surface as the provider's caller-visible bizerr.
func (s *serviceImpl) resolveExternalUserID(ctx context.Context, in ExternalLoginInput, provider string, subject string) (int, error) {
	if s == nil || s.identityProvider == nil {
		return 0, bizerr.NewCode(CodeAuthExternalUserNotProvisioned)
	}
	userID, found, err := s.identityProvider.Resolve(ctx, extidspi.ResolveInput{
		Provider: provider,
		Subject:  subject,
	})
	if err != nil {
		return 0, err
	}
	if found {
		return userID, nil
	}
	if !in.AllowAutoProvision {
		return 0, bizerr.NewCode(CodeAuthExternalUserNotProvisioned)
	}
	userID, err = s.identityProvider.Provision(ctx, extidspi.ProvisionInput{
		Provider:           provider,
		Subject:            subject,
		Email:              strings.TrimSpace(in.Email),
		DisplayName:        in.DisplayName,
		PluginID:           in.PluginID,
		AllowAutoProvision: true,
	})
	if err != nil {
		// No enabled provider plugin keeps the historical fail-closed outcome.
		if errors.Is(err, extidspi.ErrProviderUnavailable) {
			return 0, bizerr.NewCode(CodeAuthExternalUserNotProvisioned)
		}
		return 0, err
	}
	logger.Infof(ctx, "external login auto-provisioned provider=%s subject=%s userId=%d", provider, subject, userID)
	return userID, nil
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

	// Resolve the linked local user through the provider-owned authoritative
	// (provider, subject) key, provisioning through the provider policy when
	// allowed. Failures keep the uniform not-provisioned outcome so external
	// login never leaks whether the captured email exists as another account.
	userID, err := s.resolveExternalUserID(ctx, in, provider, subject)
	if err != nil {
		dispatchExternalLoginFailed(in.Email, authEventMessageExternalNotProvisioned, authHookReasonExternalNotProvisioned)
		return nil, err
	}

	var user *entity.SysUser
	if err = dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: userID}).
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
