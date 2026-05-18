// auth_impl.go implements the default authentication flow for credential
// verification, login policy checks, tenant selection, token issuance, session
// persistence, and auth lifecycle hooks. Keep runtime dependencies on the
// serviceImpl fields so callers share cache, session, tenant, and plugin state.

package auth

import (
	"context"
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/session"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
	pkgtenantcap "lina-core/pkg/tenantcap"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mssola/useragent"
	"golang.org/x/crypto/bcrypt"
)

// SessionStore returns the session store instance.
func (s *serviceImpl) SessionStore() session.Store {
	return s.sessionStore
}

// Login verifies credentials and issues JWT token.
func (s *serviceImpl) Login(ctx context.Context, in LoginInput) (*LoginOutput, error) {
	// Extract client info for login log
	var ip, browser, osName string
	if r := g.RequestFromCtx(ctx); r != nil {
		ip = r.GetClientIp()
		ua := useragent.New(r.GetHeader("User-Agent"))
		browserName, browserVersion := ua.Browser()
		browser = browserName + " " + browserVersion
		osName = ua.OS()
	}

	dispatchLoginFailed := func(username string, msg string, reason string) {
		if s.pluginSvc == nil {
			return
		}
		if hookErr := s.pluginSvc.HandleAuthLoginFailed(ctx, pluginsvc.AuthLoginSucceededInput{
			UserName:   username,
			Status:     authLoginStatusFail,
			Ip:         ip,
			ClientType: "web",
			Browser:    browser,
			Os:         osName,
			Message:    msg,
			Reason:     reason,
		}); hookErr != nil {
			logger.Warningf(ctx, "plugin login failed hook failed: %v", hookErr)
		}
	}

	blacklisted, err := s.configSvc.IsLoginIPBlacklisted(ctx, ip)
	if err != nil {
		dispatchLoginFailed(in.Username, pluginsvc.AuthEventMessageInvalidCredentials, pluginhost.AuthHookReasonInvalidCredentials)
		return nil, err
	}
	if blacklisted {
		dispatchLoginFailed(in.Username, pluginsvc.AuthEventMessageIPBlacklisted, pluginhost.AuthHookReasonIPBlacklisted)
		return nil, bizerr.NewCode(CodeAuthIPBlacklisted)
	}

	// Query user by username (GoFrame auto-adds deleted_at IS NULL condition)
	var user *entity.SysUser
	err = dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Username: in.Username}).
		Scan(&user)
	if err != nil {
		return nil, err
	}
	if user == nil {
		dispatchLoginFailed(in.Username, pluginsvc.AuthEventMessageInvalidCredentials, pluginhost.AuthHookReasonInvalidCredentials)
		return nil, bizerr.NewCode(CodeAuthInvalidCredentials)
	}

	// Verify password
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		dispatchLoginFailed(in.Username, pluginsvc.AuthEventMessageInvalidCredentials, pluginhost.AuthHookReasonInvalidCredentials)
		return nil, bizerr.NewCode(CodeAuthInvalidCredentials)
	}

	// Check status
	if user.Status == statusDisabled {
		dispatchLoginFailed(in.Username, pluginsvc.AuthEventMessageUserDisabled, pluginhost.AuthHookReasonUserDisabled)
		return nil, bizerr.NewCode(CodeAuthUserDisabled)
	}

	tenants, err := s.loginTenants(ctx, user.Id)
	if err != nil {
		return nil, err
	}
	if s.tenantSvc != nil && s.tenantSvc.Enabled(ctx) && user.TenantId != int(pkgtenantcap.PLATFORM) && len(tenants) == 0 {
		dispatchLoginFailed(in.Username, "Tenant is not available", "tenant_unavailable")
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

	tenantID := int(pkgtenantcap.PLATFORM)
	if len(tenants) == 1 {
		tenantID = tenants[0].Id
	}

	// Generate JWT token pair
	accessToken, refreshToken, tokenId, err := s.generateTokenPair(ctx, user, tenantID)
	if err != nil {
		return nil, err
	}

	// Record login time
	if _, err = dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: user.Id}).
		Data(do.SysUser{LoginDate: gtime.Now()}).
		Update(); err != nil {
		return nil, bizerr.WrapCode(err, CodeAuthLoginStateUpdateFailed)
	}

	// Create online session
	if err = s.createSession(ctx, user, tenantID, tokenId); err != nil {
		logger.Warningf(ctx, "create online session failed tokenId=%s err=%v", tokenId, err)
	}

	if s.pluginSvc != nil {
		if err := s.pluginSvc.HandleAuthLoginSucceeded(ctx, pluginsvc.AuthLoginSucceededInput{
			UserName:   in.Username,
			Status:     authLoginStatusSuccess,
			Ip:         ip,
			ClientType: "web",
			Browser:    browser,
			Os:         osName,
			Message:    pluginsvc.AuthEventMessageLoginSuccessful,
			Reason:     pluginhost.AuthHookReasonLoginSuccessful,
		}); err != nil {
			logger.Warningf(ctx, "plugin login succeeded hook failed: %v", err)
		}
	}
	return &LoginOutput{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// IssueTenantToken consumes a pre-login token and issues a tenant-bound JWT.
func (s *serviceImpl) IssueTenantToken(ctx context.Context, in TenantTokenIssueInput) (*TenantTokenOutput, error) {
	record, ok, err := s.preTokens.Consume(ctx, in.PreToken)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeAuthTokenStateUnavailable)
	}
	if !ok {
		return nil, bizerr.NewCode(CodeAuthPreTokenInvalid)
	}
	if err := s.validateUserTenant(ctx, record.UserID, in.TenantID); err != nil {
		return nil, err
	}
	user := &entity.SysUser{Id: record.UserID, Username: record.Username, Status: record.Status}
	accessToken, refreshToken, tokenID, err := s.generateTokenPair(ctx, user, in.TenantID)
	if err != nil {
		return nil, err
	}
	if err = s.createSession(ctx, user, in.TenantID, tokenID); err != nil {
		logger.Warningf(ctx, "create tenant token session failed tokenId=%s tenantId=%d err=%v", tokenID, in.TenantID, err)
	}
	return &TenantTokenOutput{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// ReissueTenantToken validates tenant membership, revokes the current token, and issues a new JWT.
func (s *serviceImpl) ReissueTenantToken(ctx context.Context, in TenantTokenReissueInput) (*TenantTokenOutput, error) {
	if in.CurrentClaims == nil {
		return nil, bizerr.NewCode(CodeAuthTokenInvalid)
	}
	if err := s.validateSwitchTenant(ctx, in.CurrentClaims.UserId, in.TenantID); err != nil {
		return nil, err
	}
	expiresAt := time.Time{}
	if in.CurrentClaims.ExpiresAt != nil {
		expiresAt = in.CurrentClaims.ExpiresAt.Time
	}
	if err := s.revokeSession(ctx, in.CurrentClaims.TokenId, expiresAt); err != nil {
		return nil, err
	}
	user := &entity.SysUser{Id: in.CurrentClaims.UserId, Username: in.CurrentClaims.Username, Status: in.CurrentClaims.Status}
	accessToken, refreshToken, tokenID, err := s.generateTokenPair(ctx, user, in.TenantID)
	if err != nil {
		return nil, err
	}
	if err = s.createSession(ctx, user, in.TenantID, tokenID); err != nil {
		logger.Warningf(ctx, "create reissued tenant session failed tokenId=%s tenantId=%d err=%v", tokenID, in.TenantID, err)
	}
	return &TenantTokenOutput{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// ReissueTenantTokenFromBearer parses the current token and reissues it for another tenant.
func (s *serviceImpl) ReissueTenantTokenFromBearer(ctx context.Context, tokenString string, tenantID int) (*TenantTokenOutput, error) {
	claims, err := s.ParseToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	return s.ReissueTenantToken(ctx, TenantTokenReissueInput{
		CurrentClaims: claims,
		TenantID:      tenantID,
	})
}

// ParseToken parses and validates JWT token, returns claims.
func (s *serviceImpl) ParseToken(ctx context.Context, tokenString string) (*Claims, error) {
	return s.parseToken(ctx, tokenString, tokenKindAccess)
}

// Refresh validates a refresh token and issues a fresh access token for the
// existing online session.
func (s *serviceImpl) Refresh(ctx context.Context, in RefreshInput) (*RefreshOutput, error) {
	claims, err := s.parseToken(ctx, in.RefreshToken, tokenKindRefresh)
	if err != nil {
		return nil, err
	}
	sessionTimeout, err := s.configSvc.GetSessionTimeout(ctx)
	if err != nil {
		return nil, err
	}
	active, err := s.sessionStore.TouchOrValidate(ctx, claims.TenantId, claims.TokenId, sessionTimeout)
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, bizerr.NewCode(CodeAuthTokenInvalid)
	}

	var user *entity.SysUser
	err = dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: claims.UserId}).
		Scan(&user)
	if err != nil {
		return nil, err
	}
	if user == nil {
		if revokeErr := s.RevokeSession(ctx, claims.TokenId); revokeErr != nil {
			logger.Warningf(ctx, "revoke missing-user refresh session failed tokenId=%s err=%v", claims.TokenId, revokeErr)
		}
		return nil, bizerr.NewCode(CodeAuthTokenInvalid)
	}
	if user.Status == statusDisabled {
		if revokeErr := s.RevokeSession(ctx, claims.TokenId); revokeErr != nil {
			logger.Warningf(ctx, "revoke disabled-user refresh session failed tokenId=%s userId=%d err=%v", claims.TokenId, user.Id, revokeErr)
		}
		return nil, bizerr.NewCode(CodeAuthUserDisabled)
	}

	// The host signer only ever issues access/refresh tokens with
	// TenantId == PLATFORM (single-tenant / platform login) or a real
	// positive tenant ID. A refresh token claiming a negative/sentinel tenant
	// ID never originates from the host, so treat it as forged or corrupt:
	// tear down the session and reject the refresh.
	if claims.TenantId < int(pkgtenantcap.PLATFORM) {
		if revokeErr := s.RevokeSession(ctx, claims.TokenId); revokeErr != nil {
			logger.Warningf(ctx, "revoke invalid-tenant refresh session failed tokenId=%s userId=%d tenantId=%d err=%v", claims.TokenId, user.Id, claims.TenantId, revokeErr)
		}
		return nil, bizerr.NewCode(CodeAuthTokenInvalid)
	}
	// Re-validate tenant membership so a user removed from the token's tenant
	// cannot keep minting access tokens just because their refresh token JWT
	// and online session row still exist. Platform-scoped tokens skip this
	// check because they do not represent a tenant membership.
	//
	// We split the failure modes by error shape:
	//   - bizerr.As(err) == true: the tenant provider made a definitive
	//     authorization decision (CodeMembershipNotFound, CodeTenantUnavailable,
	//     ...). Tear down the session so the user is forced through login.
	//   - bizerr.As(err) == false: the provider hit an infrastructure error
	//     (DB outage, timeout, plugin transport failure, ...). The membership
	//     state is unknowable, so we surface the error to the client without
	//     destroying the session — a transient blip should not kick every
	//     active tenant user offline. Access tokens are short-lived; if the
	//     eviction is real, the next refresh after infra recovery will see a
	//     definitive bizerr and revoke at that point.
	if claims.TenantId > int(pkgtenantcap.PLATFORM) {
		if err = s.validateUserTenant(ctx, user.Id, claims.TenantId); err != nil {
			if _, definitive := bizerr.As(err); definitive {
				if revokeErr := s.RevokeSession(ctx, claims.TokenId); revokeErr != nil {
					logger.Warningf(ctx, "revoke evicted-tenant refresh session failed tokenId=%s userId=%d tenantId=%d err=%v", claims.TokenId, user.Id, claims.TenantId, revokeErr)
				}
			} else {
				logger.Warningf(ctx, "tenant membership lookup failed during refresh tokenId=%s userId=%d tenantId=%d err=%v", claims.TokenId, user.Id, claims.TenantId, err)
			}
			return nil, err
		}
	}

	accessToken, err := s.signToken(ctx, user, claims.TenantId, claims.TokenId, tokenKindAccess, claims.IsImpersonation, claims.ActingUserId)
	if err != nil {
		return nil, err
	}
	if s.roleSvc != nil {
		if _, err = s.roleSvc.PrimeTokenAccessContext(datascope.WithTenantScope(ctx, claims.TenantId), claims.TokenId, user.Id); err != nil {
			return nil, err
		}
	}
	return &RefreshOutput{AccessToken: accessToken, RefreshToken: in.RefreshToken}, nil
}

// parseToken parses and validates a JWT for the expected token kind.
func (s *serviceImpl) parseToken(ctx context.Context, tokenString string, expected tokenKind) (*Claims, error) {
	jwtSecret := s.configSvc.GetJwtSecret(ctx)
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, bizerr.NewCode(CodeAuthTokenInvalid)
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.TokenType != expected {
			return nil, bizerr.NewCode(CodeAuthTokenInvalid)
		}
		revoked, err := s.revoked.Revoked(ctx, claims.TokenId)
		if err != nil {
			return nil, bizerr.WrapCode(err, CodeAuthTokenStateUnavailable)
		}
		if revoked {
			return nil, bizerr.NewCode(CodeAuthTokenInvalid)
		}
		return claims, nil
	}
	return nil, bizerr.NewCode(CodeAuthTokenInvalid)
}

// HashPassword hashes password using bcrypt.
func (s *serviceImpl) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", gerror.Wrap(err, "error.auth.password.hashFailed")
	}
	return string(hash), nil
}

// Logout records logout login log and removes session.
func (s *serviceImpl) Logout(ctx context.Context, username string, tenantId int, tokenId string) error {
	var ip, browser, osName string
	if r := g.RequestFromCtx(ctx); r != nil {
		ip = r.GetClientIp()
		ua := useragent.New(r.GetHeader("User-Agent"))
		browserName, browserVersion := ua.Browser()
		browser = browserName + " " + browserVersion
		osName = ua.OS()
	}
	// Delete session
	if tokenId != "" {
		if err := s.RevokeSession(ctx, tokenId); err != nil {
			logger.Warningf(ctx, "revoke session during logout failed tokenId=%s tenantId=%d err=%v", tokenId, tenantId, err)
			return err
		}
	}
	if s.pluginSvc != nil {
		if err := s.pluginSvc.HandleAuthLogoutSucceeded(ctx, pluginsvc.AuthLoginSucceededInput{
			UserName:   username,
			Status:     authLoginStatusSuccess,
			Ip:         ip,
			ClientType: "web",
			Browser:    browser,
			Os:         osName,
			Message:    pluginsvc.AuthEventMessageLogoutSuccessful,
			Reason:     pluginhost.AuthHookReasonLogoutSuccessful,
		}); err != nil {
			logger.Warningf(ctx, "plugin logout succeeded hook failed: %v", err)
		}
	}
	return nil
}

// RevokeSession removes one online session by token ID and its cached access context.
func (s *serviceImpl) RevokeSession(ctx context.Context, tokenId string) error {
	return s.revokeSession(ctx, tokenId, time.Time{})
}

// revokeSession marks a token ID as revoked and removes the online-session
// projection. A zero expiration falls back to the longest host-issued token TTL
// because force-logout callers only know the token ID, not the signed JWT.
func (s *serviceImpl) revokeSession(ctx context.Context, tokenId string, expiresAt time.Time) error {
	if tokenId == "" {
		return nil
	}
	if s.roleSvc != nil {
		s.roleSvc.InvalidateTokenAccessContext(ctx, tokenId)
	}
	if err := s.revokeTokenID(ctx, tokenId, expiresAt); err != nil {
		return err
	}
	if s.sessionStore == nil {
		return nil
	}
	return s.sessionStore.Delete(ctx, tokenId)
}

// revokeTokenID writes the shared JWT revoke marker used by all cluster nodes
// before local session state is considered invalidated.
func (s *serviceImpl) revokeTokenID(ctx context.Context, tokenID string, expiresAt time.Time) error {
	if tokenID == "" || s == nil || s.revoked == nil {
		return nil
	}
	if expiresAt.IsZero() {
		var err error
		expiresAt, err = s.fallbackRevocationExpiresAt(ctx)
		if err != nil {
			return err
		}
	}
	if err := s.revoked.Add(ctx, tokenID, expiresAt); err != nil {
		return bizerr.WrapCode(err, CodeAuthTokenStateUnavailable)
	}
	return nil
}

// fallbackRevocationExpiresAt returns a conservative revoke expiration for
// token-ID-only invalidation paths such as logout and monitor force-logout.
func (s *serviceImpl) fallbackRevocationExpiresAt(ctx context.Context) (time.Time, error) {
	ttl, err := s.tokenTTL(ctx, tokenKindRefresh)
	if err != nil {
		return time.Time{}, bizerr.WrapCode(err, CodeAuthTokenStateUnavailable)
	}
	return time.Now().Add(ttl), nil
}

// generateToken generates JWT token for given user, returns token string and tokenId.
func (s *serviceImpl) generateToken(ctx context.Context, user *entity.SysUser, tenantID int) (string, string, error) {
	tokenID := guid.S()
	token, err := s.signToken(ctx, user, tenantID, tokenID, tokenKindAccess, false, 0)
	if err != nil {
		return "", "", err
	}
	return token, tokenID, nil
}

// generateTokenPair signs access and refresh JWTs for one online session.
func (s *serviceImpl) generateTokenPair(ctx context.Context, user *entity.SysUser, tenantID int) (string, string, string, error) {
	tokenID := guid.S()
	accessToken, err := s.signToken(ctx, user, tenantID, tokenID, tokenKindAccess, false, 0)
	if err != nil {
		return "", "", "", err
	}
	refreshToken, err := s.signToken(ctx, user, tenantID, tokenID, tokenKindRefresh, false, 0)
	if err != nil {
		return "", "", "", err
	}
	return accessToken, refreshToken, tokenID, nil
}

// signToken signs one JWT for the supplied token kind.
func (s *serviceImpl) signToken(
	ctx context.Context,
	user *entity.SysUser,
	tenantID int,
	tokenID string,
	kind tokenKind,
	isImpersonation bool,
	actingUserID int,
) (string, error) {
	jwtTTL, err := s.tokenTTL(ctx, kind)
	if err != nil {
		return "", err
	}
	jwtSecret := s.configSvc.GetJwtSecret(ctx)
	claims := Claims{
		TokenId:         tokenID,
		TokenType:       kind,
		UserId:          user.Id,
		Username:        user.Username,
		Status:          user.Status,
		TenantId:        tenantID,
		IsImpersonation: isImpersonation,
		ActingUserId:    actingUserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}
	return signed, nil
}

// tokenTTL returns the effective lifetime for a token kind.
func (s *serviceImpl) tokenTTL(ctx context.Context, kind tokenKind) (time.Duration, error) {
	accessTTL, err := s.configSvc.GetJwtExpire(ctx)
	if err != nil {
		return 0, err
	}
	if kind == tokenKindAccess {
		return accessTTL, nil
	}
	sessionTTL, err := s.configSvc.GetSessionTimeout(ctx)
	if err != nil {
		return 0, err
	}
	refreshTTL := defaultRefreshTokenTTL
	if sessionTTL > refreshTTL {
		refreshTTL = sessionTTL
	}
	if accessTTL > refreshTTL {
		refreshTTL = accessTTL
	}
	return refreshTTL, nil
}

// getUserDeptName queries the department name for a user by userId.
func (s *serviceImpl) getUserDeptName(ctx context.Context, userId int) string {
	if s == nil || s.orgCapSvc == nil {
		return ""
	}
	deptName, err := s.orgCapSvc.GetUserDeptName(ctx, userId)
	if err != nil {
		return ""
	}
	return deptName
}

// loginTenants returns active tenant candidates for a login user.
func (s *serviceImpl) loginTenants(ctx context.Context, userID int) ([]TenantInfo, error) {
	if s == nil || s.tenantSvc == nil || !s.tenantSvc.Enabled(ctx) {
		return nil, nil
	}
	provider := pkgtenantcap.CurrentProvider()
	if provider == nil {
		return nil, nil
	}
	providerTenants, err := provider.ListUserTenants(ctx, userID)
	if err != nil {
		return nil, err
	}
	tenants := make([]TenantInfo, 0, len(providerTenants))
	for _, item := range providerTenants {
		tenants = append(tenants, TenantInfo{
			Id:     int(item.ID),
			Code:   item.Code,
			Name:   item.Name,
			Status: item.Status,
		})
	}
	return tenants, nil
}

// validateUserTenant verifies that a user can sign into tenantID.
func (s *serviceImpl) validateUserTenant(ctx context.Context, userID int, tenantID int) error {
	if s == nil || s.tenantSvc == nil || !s.tenantSvc.Enabled(ctx) {
		return nil
	}
	provider := pkgtenantcap.CurrentProvider()
	if provider == nil {
		return nil
	}
	return provider.ValidateUserInTenant(ctx, userID, pkgtenantcap.TenantID(tenantID))
}

// validateSwitchTenant verifies that a user can switch into tenantID.
func (s *serviceImpl) validateSwitchTenant(ctx context.Context, userID int, tenantID int) error {
	if s == nil || s.tenantSvc == nil || !s.tenantSvc.Enabled(ctx) {
		return nil
	}
	provider := pkgtenantcap.CurrentProvider()
	if provider == nil {
		return nil
	}
	return provider.SwitchTenant(ctx, userID, pkgtenantcap.TenantID(tenantID))
}

// createSession persists a tenant-bound online-session row.
func (s *serviceImpl) createSession(ctx context.Context, user *entity.SysUser, tenantID int, tokenID string) error {
	tenantScopedCtx := datascope.WithTenantScope(ctx, tenantID)
	var ip, browser, osName string
	if r := g.RequestFromCtx(ctx); r != nil {
		ip = r.GetClientIp()
		ua := useragent.New(r.GetHeader("User-Agent"))
		browserName, browserVersion := ua.Browser()
		browser = browserName + " " + browserVersion
		osName = ua.OS()
	}
	deptName := s.getUserDeptName(tenantScopedCtx, user.Id)
	if ttlSetter, ok := s.sessionStore.(interface{ SetDefaultTTL(time.Duration) }); ok {
		timeout, err := s.configSvc.GetSessionTimeout(ctx)
		if err != nil {
			return err
		}
		ttlSetter.SetDefaultTTL(timeout)
	}
	if err := s.sessionStore.Set(ctx, &session.Session{
		TokenId:   tokenID,
		TenantId:  tenantID,
		UserId:    user.Id,
		Username:  user.Username,
		DeptName:  deptName,
		Ip:        ip,
		Browser:   browser,
		Os:        osName,
		LoginTime: gtime.Now(),
	}); err != nil {
		return err
	}
	if s.roleSvc == nil {
		return nil
	}
	_, err := s.roleSvc.PrimeTokenAccessContext(tenantScopedCtx, tokenID, user.Id)
	return err
}
