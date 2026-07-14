// auth_impl.go implements the default authentication flow for credential
// verification, login policy checks, tenant selection, token issuance, session
// persistence, and auth lifecycle hooks. Keep runtime dependencies on the
// serviceImpl fields so callers share cache, session, tenant, and plugin state.

package auth

import (
	"context"
	"strings"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/session"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginhost"
	"lina-core/pkg/statusflag"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
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

	clientType, err := ParseClientType(in.ClientType.String())
	if err != nil {
		return nil, err
	}

	dispatchLoginFailed := func(username string, msg string, reason string) {
		s.dispatchAuthHookEvent(ctx, pluginhost.ExtensionPointAuthLoginFailed, pluginhost.AuthHookPayloadInput{
			UserName:   username,
			Status:     authLoginStatusFail,
			IP:         ip,
			ClientType: clientType.String(),
			Browser:    browser,
			OS:         osName,
			Message:    msg,
			Reason:     reason,
		}, "plugin login failed hook failed")
	}

	blacklisted, err := s.configSvc.IsLoginIPBlacklisted(ctx, ip)
	if err != nil {
		dispatchLoginFailed(in.Username, authEventMessageInvalidCredentials, pluginhost.AuthHookReasonInvalidCredentials)
		return nil, err
	}
	if blacklisted {
		dispatchLoginFailed(in.Username, authEventMessageIPBlacklisted, pluginhost.AuthHookReasonIPBlacklisted)
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
		dispatchLoginFailed(in.Username, authEventMessageInvalidCredentials, pluginhost.AuthHookReasonInvalidCredentials)
		return nil, bizerr.NewCode(CodeAuthInvalidCredentials)
	}

	// Verify password
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		dispatchLoginFailed(in.Username, authEventMessageInvalidCredentials, pluginhost.AuthHookReasonInvalidCredentials)
		return nil, bizerr.NewCode(CodeAuthInvalidCredentials)
	}

	// Check status
	if user.Status == statusflag.Disabled.Int() {
		dispatchLoginFailed(in.Username, authEventMessageUserDisabled, pluginhost.AuthHookReasonUserDisabled)
		return nil, bizerr.NewCode(CodeAuthUserDisabled)
	}

	tenantSvcAvailable := s.tenantSvc != nil && s.tenantSvc.Available(ctx)
	tenants, err := s.loginTenants(ctx, user.Id)
	if err != nil {
		return nil, err
	}
	if user.TenantId != int(tenantcap.PLATFORM) && (!tenantSvcAvailable || len(tenants) == 0) {
		dispatchLoginFailed(in.Username, authEventMessageTenantUnavailable, authHookReasonTenantUnavailable)
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
		return &LoginOutput{PreToken: preToken, Tenants: tenants}, nil
	}

	tenantID := int(tenantcap.PLATFORM)
	if len(tenants) == 1 {
		tenantID = tenants[0].Id
	}

	// Generate JWT token pair
	accessToken, refreshToken, tokenId, err := s.generateTokenPair(ctx, user, tenantID, clientType)
	if err != nil {
		return nil, err
	}

	// Record login time
	loginDate := time.Now()
	if _, err = dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: user.Id}).
		Data(do.SysUser{LoginDate: &loginDate}).
		Update(); err != nil {
		return nil, bizerr.WrapCode(err, CodeAuthLoginStateUpdateFailed)
	}

	// Create online session
	if err = s.createSession(ctx, user, tenantID, tokenId, clientType); err != nil {
		logger.Warningf(ctx, "create online session failed tokenId=%s err=%v", tokenId, err)
	}

	s.dispatchAuthHookEvent(ctx, pluginhost.ExtensionPointAuthLoginSucceeded, pluginhost.AuthHookPayloadInput{
		UserName:   in.Username,
		Status:     authLoginStatusSuccess,
		IP:         ip,
		ClientType: clientType.String(),
		Browser:    browser,
		OS:         osName,
		Message:    authEventMessageLoginSuccessful,
		Reason:     pluginhost.AuthHookReasonLoginSuccessful,
	}, "plugin login succeeded hook failed")
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
	clientType, err := ParseClientType(record.ClientType.String())
	if err != nil {
		return nil, err
	}
	user := &entity.SysUser{Id: record.UserID, Username: record.Username, Status: record.Status}
	accessToken, refreshToken, tokenID, err := s.generateTokenPair(ctx, user, in.TenantID, clientType)
	if err != nil {
		return nil, err
	}
	if err = s.createSession(ctx, user, in.TenantID, tokenID, clientType); err != nil {
		logger.Warningf(ctx, "create tenant token session failed tokenId=%s tenantId=%d err=%v", tokenID, in.TenantID, err)
	}
	return &TenantTokenOutput{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// ReissueTenantToken validates tenant membership, revokes the current token, and issues a new JWT.
func (s *serviceImpl) ReissueTenantToken(ctx context.Context, in TenantTokenReissueInput) (*TenantTokenOutput, error) {
	if in.CurrentClaims == nil {
		return nil, bizerr.NewCode(CodeAuthTokenInvalid)
	}
	if !in.SkipSessionValidation {
		if err := s.validateAccessSession(ctx, in.CurrentClaims); err != nil {
			return nil, err
		}
	}
	if err := s.validateSwitchTenant(ctx, in.CurrentClaims.UserId, in.TenantID); err != nil {
		return nil, err
	}
	clientType, err := claimsClientType(in.CurrentClaims)
	if err != nil {
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
	accessToken, refreshToken, tokenID, err := s.generateTokenPair(ctx, user, in.TenantID, clientType)
	if err != nil {
		return nil, err
	}
	if err = s.createSession(ctx, user, in.TenantID, tokenID, clientType); err != nil {
		logger.Warningf(ctx, "create reissued tenant session failed tokenId=%s tenantId=%d err=%v", tokenID, in.TenantID, err)
	}
	return &TenantTokenOutput{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// ReissueTenantTokenFromBearer parses the current token and reissues it for another tenant.
func (s *serviceImpl) ReissueTenantTokenFromBearer(ctx context.Context, tokenString string, tenantID int) (*TenantTokenOutput, error) {
	claims, err := s.AuthenticateAccessToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	return s.ReissueTenantToken(ctx, TenantTokenReissueInput{
		CurrentClaims:         claims,
		SkipSessionValidation: true,
		TenantID:              tenantID,
	})
}

// IssueImpersonationToken signs and registers a host-owned impersonation token.
func (s *serviceImpl) IssueImpersonationToken(ctx context.Context, in ImpersonationTokenIssueInput) (*ImpersonationTokenOutput, error) {
	if in.ActingUserID <= 0 || in.TenantID <= 0 {
		return nil, bizerr.NewCode(CodeAuthTokenInvalid)
	}

	var user *entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: in.ActingUserID}).
		Scan(&user); err != nil {
		return nil, err
	}
	if user == nil {
		return nil, bizerr.NewCode(CodeAuthTokenInvalid)
	}
	if user.Status == statusflag.Disabled.Int() {
		return nil, bizerr.NewCode(CodeAuthUserDisabled)
	}
	businessCtx, _ := ctx.Value(bizctx.ContextKey).(*model.Context)
	if businessCtx == nil {
		return nil, bizerr.NewCode(CodeAuthTokenInvalid)
	}
	clientType, err := ParseClientType(businessCtx.ClientType)
	if err != nil {
		return nil, bizerr.NewCode(CodeAuthTokenInvalid)
	}

	tokenID := guid.S()
	accessToken, err := s.signToken(ctx, user, in.TenantID, tokenID, tokenKindAccess, clientType, true, in.ActingUserID)
	if err != nil {
		return nil, err
	}
	if err = s.createImpersonationSession(ctx, user, in.TenantID, tokenID, clientType, in.ActingUserID); err != nil {
		return nil, err
	}
	return &ImpersonationTokenOutput{
		AccessToken:  accessToken,
		TokenID:      tokenID,
		TenantID:     in.TenantID,
		ActingUserID: in.ActingUserID,
	}, nil
}

// RevokeImpersonationToken validates and revokes one host impersonation token.
func (s *serviceImpl) RevokeImpersonationToken(ctx context.Context, tokenString string, tenantID int) error {
	claims, err := s.parseToken(ctx, strings.TrimSpace(strings.TrimPrefix(tokenString, "Bearer ")), tokenKindAccess)
	if err != nil {
		return err
	}
	if claims == nil || !claims.IsImpersonation || claims.TokenId == "" || claims.ActingUserId <= 0 {
		return bizerr.NewCode(CodeAuthTokenInvalid)
	}
	if tenantID > 0 && claims.TenantId != tenantID {
		return bizerr.NewCode(CodeAuthTokenInvalid)
	}
	if err = s.validateAccessSession(ctx, claims); err != nil {
		return err
	}
	expiresAt := time.Time{}
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	}
	return s.revokeSession(ctx, claims.TokenId, expiresAt)
}

// AuthenticateAccessToken parses an access token and validates its online session.
func (s *serviceImpl) AuthenticateAccessToken(ctx context.Context, tokenString string) (*Claims, error) {
	claims, err := s.parseToken(ctx, normalizeBearerToken(tokenString), tokenKindAccess)
	if err != nil {
		return nil, err
	}
	if err = s.validateAccessSession(ctx, claims); err != nil {
		return nil, err
	}
	return claims, nil
}

// normalizeBearerToken accepts either an Authorization header value or a raw
// access-token string and returns the token segment consumed by JWT parsing.
func normalizeBearerToken(tokenString string) string {
	token := strings.TrimSpace(tokenString)
	const bearerPrefix = "Bearer "
	if strings.HasPrefix(token, bearerPrefix) {
		return strings.TrimSpace(strings.TrimPrefix(token, bearerPrefix))
	}
	return token
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
	if user.Status == statusflag.Disabled.Int() {
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
	if claims.TenantId < int(tenantcap.PLATFORM) {
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
	if claims.TenantId > int(tenantcap.PLATFORM) {
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

	accessToken, err := s.signToken(ctx, user, claims.TenantId, claims.TokenId, tokenKindAccess, claims.ClientType, claims.IsImpersonation, claims.ActingUserId)
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
		if _, err := claimsClientType(claims); err != nil {
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

// validateAccessSession confirms the parsed access token still has a valid
// sys_online_session row. This is the complete login-state authority used by
// middleware, tenant switching, and impersonation revocation.
func (s *serviceImpl) validateAccessSession(ctx context.Context, claims *Claims) error {
	if claims == nil || claims.TokenId == "" {
		return bizerr.NewCode(CodeAuthTokenInvalid)
	}
	if s == nil || s.sessionStore == nil || s.configSvc == nil {
		return bizerr.NewCode(CodeAuthTokenInvalid)
	}
	sessionTimeout, err := s.configSvc.GetSessionTimeout(ctx)
	if err != nil {
		return err
	}
	active, err := s.sessionStore.TouchOrValidate(ctx, claims.TenantId, claims.TokenId, sessionTimeout)
	if err != nil {
		return err
	}
	if !active {
		if s.roleSvc != nil {
			s.roleSvc.InvalidateTokenAccessContext(ctx, claims.TokenId)
		}
		return bizerr.NewCode(CodeAuthTokenInvalid)
	}
	return nil
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
func (s *serviceImpl) Logout(ctx context.Context, in LogoutInput) error {
	clientType, err := ParseClientType(in.ClientType.String())
	if err != nil {
		return err
	}
	var ip, browser, osName string
	if r := g.RequestFromCtx(ctx); r != nil {
		ip = r.GetClientIp()
		ua := useragent.New(r.GetHeader("User-Agent"))
		browserName, browserVersion := ua.Browser()
		browser = browserName + " " + browserVersion
		osName = ua.OS()
	}
	// Delete session
	if in.TokenID != "" {
		if err := s.RevokeSession(ctx, in.TokenID); err != nil {
			logger.Warningf(ctx, "revoke session during logout failed tokenId=%s tenantId=%d err=%v", in.TokenID, in.TenantID, err)
			return err
		}
	}
	s.dispatchAuthHookEvent(ctx, pluginhost.ExtensionPointAuthLogoutSucceeded, pluginhost.AuthHookPayloadInput{
		UserName:   in.Username,
		Status:     authLoginStatusSuccess,
		IP:         ip,
		ClientType: clientType.String(),
		Browser:    browser,
		OS:         osName,
		Message:    authEventMessageLogoutSuccessful,
		Reason:     pluginhost.AuthHookReasonLogoutSuccessful,
	}, "plugin logout succeeded hook failed")
	return nil
}

// dispatchAuthHookEvent publishes one auth lifecycle hook through the generic
// plugin hook dispatcher while auth keeps ownership of the auth payload shape.
func (s *serviceImpl) dispatchAuthHookEvent(
	ctx context.Context,
	event pluginhost.ExtensionPoint,
	input pluginhost.AuthHookPayloadInput,
	warning string,
) {
	if s == nil || s.hookSvc == nil {
		return
	}
	if err := s.hookSvc.DispatchHookEvent(ctx, event, pluginhost.BuildAuthHookPayloadValues(input)); err != nil {
		logger.Warningf(ctx, "%s: %v", warning, err)
	}
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

// generateTokenPair signs access and refresh JWTs for one online session.
func (s *serviceImpl) generateTokenPair(ctx context.Context, user *entity.SysUser, tenantID int, clientType ClientType) (string, string, string, error) {
	tokenID := guid.S()
	accessToken, err := s.signToken(ctx, user, tenantID, tokenID, tokenKindAccess, clientType, false, 0)
	if err != nil {
		return "", "", "", err
	}
	refreshToken, err := s.signToken(ctx, user, tenantID, tokenID, tokenKindRefresh, clientType, false, 0)
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
	clientType ClientType,
	isImpersonation bool,
	actingUserID int,
) (string, error) {
	clientType, err := ParseClientType(clientType.String())
	if err != nil {
		return "", err
	}
	jwtTTL, err := s.tokenTTL(ctx, kind)
	if err != nil {
		return "", err
	}
	var (
		jwtSecret = s.configSvc.GetJwtSecret(ctx)
		claims    = Claims{
			TokenId:         tokenID,
			TokenType:       kind,
			ClientType:      clientType,
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
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	)
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
	_, deptName, err := s.orgCapSvc.Assignment().GetUserDeptInfo(ctx, userId)
	if err != nil {
		return ""
	}
	return deptName
}

// claimsClientType validates and returns the user-session client type from
// parsed JWT claims.
func claimsClientType(claims *Claims) (ClientType, error) {
	if claims == nil {
		return "", bizerr.NewCode(CodeAuthTokenInvalid)
	}
	clientType, err := ParseClientType(claims.ClientType.String())
	if err != nil {
		return "", bizerr.NewCode(CodeAuthTokenInvalid)
	}
	return clientType, nil
}

// loginTenants returns active tenant candidates for a login user.
func (s *serviceImpl) loginTenants(ctx context.Context, userID int) ([]TenantInfo, error) {
	if s == nil || s.tenantSvc == nil || !s.tenantSvc.Available(ctx) {
		return nil, nil
	}
	providerTenants, err := s.tenantSvc.ListUserTenants(ctx, userID)
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
	if s == nil || s.tenantSvc == nil || !s.tenantSvc.Available(ctx) {
		return nil
	}
	return s.tenantSvc.ValidateUserInTenant(ctx, userID, tenantcap.TenantID(tenantID))
}

// validateSwitchTenant verifies that a user can switch into tenantID.
func (s *serviceImpl) validateSwitchTenant(ctx context.Context, userID int, tenantID int) error {
	if s == nil || s.tenantSvc == nil || !s.tenantSvc.Available(ctx) {
		return nil
	}
	return s.tenantSvc.SwitchTenant(ctx, userID, tenantcap.TenantID(tenantID))
}

// createSession persists a tenant-bound online-session row.
func (s *serviceImpl) createSession(ctx context.Context, user *entity.SysUser, tenantID int, tokenID string, clientType ClientType) error {
	tenantScopedCtx := datascope.WithTenantScope(ctx, tenantID)
	return s.createSessionWithPrimeContext(ctx, tenantScopedCtx, user, tenantID, tokenID, clientType)
}

// createImpersonationSession persists an impersonation session and primes role
// access with platform-admin grants while keeping target tenant cache scope.
func (s *serviceImpl) createImpersonationSession(
	ctx context.Context,
	user *entity.SysUser,
	tenantID int,
	tokenID string,
	clientType ClientType,
	actingUserID int,
) error {
	clientType, err := ParseClientType(clientType.String())
	if err != nil {
		return err
	}
	impersonationCtx := context.WithValue(datascope.WithTenantScope(ctx, tenantID), bizctx.ContextKey, &model.Context{
		TokenId:         tokenID,
		UserId:          user.Id,
		Username:        user.Username,
		Status:          user.Status,
		ClientType:      clientType.String(),
		TenantId:        tenantID,
		ActingAsTenant:  true,
		ActingUserId:    actingUserID,
		IsImpersonation: true,
	})
	return s.createSessionWithPrimeContext(impersonationCtx, impersonationCtx, user, tenantID, tokenID, clientType)
}

// createSessionWithPrimeContext persists a tenant-bound online-session row and
// primes role access through the provided permission context.
func (s *serviceImpl) createSessionWithPrimeContext(
	ctx context.Context,
	primeCtx context.Context,
	user *entity.SysUser,
	tenantID int,
	tokenID string,
	clientType ClientType,
) error {
	clientType, err := ParseClientType(clientType.String())
	if err != nil {
		return err
	}
	var ip, browser, osName string
	if r := g.RequestFromCtx(ctx); r != nil {
		ip = r.GetClientIp()
		ua := useragent.New(r.GetHeader("User-Agent"))
		browserName, browserVersion := ua.Browser()
		browser = browserName + " " + browserVersion
		osName = ua.OS()
	}
	deptName := s.getUserDeptName(ctx, user.Id)
	if ttlSetter, ok := s.sessionStore.(interface{ SetDefaultTTL(time.Duration) }); ok {
		timeout, err := s.configSvc.GetSessionTimeout(ctx)
		if err != nil {
			return err
		}
		ttlSetter.SetDefaultTTL(timeout)
	}
	loginTime := time.Now()
	if err := s.sessionStore.Set(ctx, &session.Session{
		TokenId:    tokenID,
		TenantId:   tenantID,
		UserId:     user.Id,
		Username:   user.Username,
		ClientType: clientType.String(),
		DeptName:   deptName,
		Ip:         ip,
		Browser:    browser,
		Os:         osName,
		LoginTime:  &loginTime,
	}); err != nil {
		return err
	}
	if s.roleSvc == nil {
		return nil
	}
	_, err = s.roleSvc.PrimeTokenAccessContext(primeCtx, tokenID, user.Id)
	return err
}
