// This file implements external-provider login handoff used by source-plugin
// OAuth callbacks. It mirrors the password Login() flow but trusts the caller
// to have verified the provider exchange, uses email as the local user join
// key, and emits the same auth lifecycle hooks so audit and notification
// pipelines treat external logins identically to password logins.

package auth

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/mssola/useragent"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginhost"
)

// LoginByExternal resolves a verified external provider identity into a host
// login outcome. The flow intentionally mirrors Login() so single-tenant users
// receive an access/refresh pair while multi-tenant users receive a pre-login
// token plus tenant candidates. The caller is expected to be a source-plugin
// OAuth callback that has already validated the provider exchange.
func (s *serviceImpl) LoginByExternal(ctx context.Context, in ExternalLoginInput) (*LoginOutput, error) {
	email := strings.TrimSpace(in.Email)
	if email == "" {
		return nil, bizerr.NewCode(CodeAuthExternalIdentityInvalid)
	}

	clientIP := strings.TrimSpace(in.ClientIP)
	var browser, osName string
	if r := g.RequestFromCtx(ctx); r != nil {
		if clientIP == "" {
			clientIP = r.GetClientIp()
		}
		ua := useragent.New(r.GetHeader("User-Agent"))
		browserName, browserVersion := ua.Browser()
		browser = strings.TrimSpace(browserName + " " + browserVersion)
		osName = ua.OS()
	}

	username := strings.TrimSpace(in.DisplayName)
	if username == "" {
		username = email
	}

	dispatchLoginFailed := func(msg string, reason string) {
		if s == nil || s.pluginSvc == nil {
			return
		}
		if hookErr := s.pluginSvc.HandleAuthLoginFailed(ctx, pluginsvc.AuthLoginSucceededInput{
			UserName:   username,
			Status:     authLoginStatusFail,
			Ip:         clientIP,
			ClientType: "web",
			Browser:    browser,
			Os:         osName,
			Message:    msg,
			Reason:     reason,
		}); hookErr != nil {
			logger.Warningf(ctx, "plugin external login failed hook failed provider=%s err=%v", in.ProviderID, hookErr)
		}
	}

	blacklisted, err := s.configSvc.IsLoginIPBlacklisted(ctx, clientIP)
	if err != nil {
		dispatchLoginFailed(pluginsvc.AuthEventMessageInvalidCredentials, pluginhost.AuthHookReasonInvalidCredentials)
		return nil, err
	}
	if blacklisted {
		dispatchLoginFailed(pluginsvc.AuthEventMessageIPBlacklisted, pluginhost.AuthHookReasonIPBlacklisted)
		return nil, bizerr.NewCode(CodeAuthIPBlacklisted)
	}

	var user *entity.SysUser
	if err = dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Email: email}).
		Scan(&user); err != nil {
		return nil, err
	}
	if user == nil {
		dispatchLoginFailed(pluginsvc.AuthEventMessageInvalidCredentials, pluginhost.AuthHookReasonInvalidCredentials)
		return nil, bizerr.NewCode(CodeAuthExternalUserNotProvisioned)
	}
	if user.Status == statusDisabled {
		dispatchLoginFailed(pluginsvc.AuthEventMessageUserDisabled, pluginhost.AuthHookReasonUserDisabled)
		return nil, bizerr.NewCode(CodeAuthUserDisabled)
	}

	tenants, err := s.loginTenants(ctx, user.Id)
	if err != nil {
		return nil, err
	}
	if s.tenantSvc != nil && s.tenantSvc.Available(ctx) && user.TenantId != int(tenantcap.PLATFORM) && len(tenants) == 0 {
		dispatchLoginFailed("Tenant is not available", "tenant_unavailable")
		return nil, bizerr.NewCode(CodeAuthTenantUnavailable)
	}

	if len(tenants) > 1 {
		preToken, err := s.preTokens.Create(ctx, preTokenRecord{
			UserID:   user.Id,
			Username: user.Username,
			Status:   user.Status,
		})
		if err != nil {
			return nil, bizerr.WrapCode(err, CodeAuthTokenStateUnavailable)
		}
		return &LoginOutput{PreToken: preToken, Tenants: tenants}, nil
	}

	tenantID := int(tenantcap.PLATFORM)
	if len(tenants) == 1 {
		tenantID = tenants[0].Id
	}

	accessToken, refreshToken, tokenId, err := s.generateTokenPair(ctx, user, tenantID)
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

	if err = s.createSession(ctx, user, tenantID, tokenId); err != nil {
		logger.Warningf(ctx, "create external login session failed provider=%s tokenId=%s err=%v", in.ProviderID, tokenId, err)
	}

	if s.pluginSvc != nil {
		if hookErr := s.pluginSvc.HandleAuthLoginSucceeded(ctx, pluginsvc.AuthLoginSucceededInput{
			UserName:   user.Username,
			Status:     authLoginStatusSuccess,
			Ip:         clientIP,
			ClientType: "web",
			Browser:    browser,
			Os:         osName,
			Message:    pluginsvc.AuthEventMessageLoginSuccessful,
			Reason:     pluginhost.AuthHookReasonLoginSuccessful,
		}); hookErr != nil {
			logger.Warningf(ctx, "plugin external login succeeded hook failed provider=%s err=%v", in.ProviderID, hookErr)
		}
	}

	return &LoginOutput{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
