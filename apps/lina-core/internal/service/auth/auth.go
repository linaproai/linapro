// Package auth implements authentication, JWT issuance, login auditing, and
// online-session persistence for the Lina core host service.
package auth

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mssola/useragent"
	"golang.org/x/crypto/bcrypt"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/config"
	"lina-core/internal/service/orgcap"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// Auth status constants used by login validation.
const (
	// statusDisabled represents a disabled user status.
	// Mirrors user.StatusDisabled; duplicated here to avoid circular import.
	statusDisabled = 0
	// authLoginStatusSuccess marks a successful login lifecycle event.
	authLoginStatusSuccess = 0
	// authLoginStatusFail marks a failed login lifecycle event.
	authLoginStatusFail = 1
)

// Service defines the auth service contract.
type Service interface {
	// SessionStore returns the session store instance.
	SessionStore() session.Store
	// Login verifies credentials and issues JWT token.
	Login(ctx context.Context, in LoginInput) (*LoginOutput, error)
	// ParseToken parses and validates JWT token, returns claims.
	ParseToken(ctx context.Context, tokenString string) (*Claims, error)
	// HashPassword hashes password using bcrypt.
	HashPassword(password string) (string, error)
	// Logout records logout login log and removes session.
	Logout(ctx context.Context, username string, tenantId int, tokenId string)
	// RevokeSession removes one online session by token ID and its cached access context.
	RevokeSession(ctx context.Context, tokenId string) error
}

// TenantTokenIssuer defines the narrow host-owned token handoff used by
// tenant-aware auth adapters.
type TenantTokenIssuer interface {
	// IssueTenantToken consumes a pre-login token and issues a tenant-bound JWT.
	IssueTenantToken(ctx context.Context, in TenantTokenIssueInput) (*TenantTokenOutput, error)
	// ReissueTenantToken validates tenant membership, revokes the current token, and issues a new JWT.
	ReissueTenantToken(ctx context.Context, in TenantTokenReissueInput) (*TenantTokenOutput, error)
	// ReissueTenantTokenFromBearer parses the current token and reissues it for another tenant.
	ReissueTenantTokenFromBearer(ctx context.Context, tokenString string, tenantID int) (*TenantTokenOutput, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// Ensure serviceImpl implements TenantTokenIssuer.
var _ TenantTokenIssuer = (*serviceImpl)(nil)

var (
	defaultPreTokens = newPreTokenStore()
	defaultRevoked   = newRevokeList()
)

// serviceImpl implements Service.
type serviceImpl struct {
	configSvc    authConfigService    // Configuration service
	orgCapSvc    orgcap.Service       // Optional organization capability service
	pluginSvc    pluginsvc.Service    // Plugin service
	roleSvc      authRoleService      // Role service
	tenantSvc    tenantcapsvc.Service // Tenant capability service
	sessionStore session.Store        // Session store
	preTokens    preTokenStore
	revoked      revokeStore
}

// authConfigService is the narrow config surface used by auth.
type authConfigService interface {
	GetJwtSecret(ctx context.Context) string
	GetJwtExpire(ctx context.Context) (time.Duration, error)
	IsLoginIPBlacklisted(ctx context.Context, ip string) (bool, error)
}

// authRoleService is the narrow role access-cache surface used by auth.
type authRoleService interface {
	PrimeTokenAccessContext(ctx context.Context, tokenID string, userID int) (*role.UserAccessContext, error)
	InvalidateTokenAccessContext(ctx context.Context, tokenID string)
}

// New creates and returns a new Service instance.
// Pass a non-nil orgCapSvc to reuse a caller-owned organization capability
// service; pass nil to create the default orgcap service bound to the default
// plugin service instance.
func New(orgCapSvc orgcap.Service) Service {
	return newService(orgCapSvc)
}

// NewTenantTokenIssuer creates the narrowed tenant token issuer.
func NewTenantTokenIssuer(orgCapSvc orgcap.Service) TenantTokenIssuer {
	return newService(orgCapSvc)
}

// newService creates the concrete auth service implementation.
func newService(orgCapSvc orgcap.Service) *serviceImpl {
	pluginSvc := pluginsvc.New(nil)
	if orgCapSvc == nil {
		orgCapSvc = orgcap.New(pluginSvc)
	}
	return &serviceImpl{
		configSvc:    config.New(),
		orgCapSvc:    orgCapSvc,
		pluginSvc:    pluginSvc,
		roleSvc:      role.New(pluginSvc),
		tenantSvc:    tenantcapsvc.New(pluginSvc),
		sessionStore: session.NewDBStore(),
		preTokens:    defaultPreTokens,
		revoked:      defaultRevoked,
	}
}

// SessionStore returns the session store instance.
func (s *serviceImpl) SessionStore() session.Store {
	return s.sessionStore
}

// Claims defines JWT token claims.
type Claims struct {
	TokenId         string `json:"tokenId"`         // Unique token identifier
	UserId          int    `json:"userId"`          // User ID
	Username        string `json:"username"`        // Username
	Status          int    `json:"status"`          // Status
	TenantId        int    `json:"tenantId"`        // Tenant ID, where 0 means platform
	IsImpersonation bool   `json:"isImpersonation"` // Whether the token represents impersonation
	ActingUserId    int    `json:"actingUserId"`    // Real user ID during impersonation
	jwt.RegisteredClaims
}

// LoginInput defines input for Login function.
type LoginInput struct {
	Username string // Username
	Password string // Password
}

// LoginOutput defines output for Login function.
type LoginOutput struct {
	AccessToken string       // JWT access token
	PreToken    string       // Short-lived pre-login token for tenant selection
	Tenants     []TenantInfo // Tenant candidates for two-stage login
}

// TenantInfo defines one tenant candidate returned during two-stage login.
type TenantInfo struct {
	Id     int    // Tenant ID
	Code   string // Tenant code
	Name   string // Tenant display name
	Status string // Tenant status
}

// TenantTokenIssueInput defines input for issuing a tenant token after password login.
type TenantTokenIssueInput struct {
	PreToken string // Short-lived pre-login token
	TenantID int    // Target tenant ID
}

// TenantTokenReissueInput defines input for reissuing the current formal token for a tenant.
type TenantTokenReissueInput struct {
	CurrentClaims *Claims // Current token claims
	TenantID      int     // Target tenant ID
}

// TenantTokenOutput defines a tenant-bound JWT response.
type TenantTokenOutput struct {
	AccessToken string // JWT access token
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

	// Generate JWT token
	token, tokenId, err := s.generateToken(ctx, user, tenantID)
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
	return &LoginOutput{AccessToken: token}, nil
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
	token, tokenID, err := s.generateToken(ctx, user, in.TenantID)
	if err != nil {
		return nil, err
	}
	if err = s.createSession(ctx, user, in.TenantID, tokenID); err != nil {
		logger.Warningf(ctx, "create tenant token session failed tokenId=%s tenantId=%d err=%v", tokenID, in.TenantID, err)
	}
	return &TenantTokenOutput{AccessToken: token}, nil
}

// ReissueTenantToken validates tenant membership, revokes the current token, and issues a new JWT.
func (s *serviceImpl) ReissueTenantToken(ctx context.Context, in TenantTokenReissueInput) (*TenantTokenOutput, error) {
	if in.CurrentClaims == nil {
		return nil, bizerr.NewCode(CodeAuthTokenInvalid)
	}
	if err := s.validateSwitchTenant(ctx, in.CurrentClaims.UserId, in.TenantID); err != nil {
		return nil, err
	}
	if in.CurrentClaims.ExpiresAt != nil {
		if err := s.revoked.Add(ctx, in.CurrentClaims.TokenId, in.CurrentClaims.ExpiresAt.Time); err != nil {
			return nil, bizerr.WrapCode(err, CodeAuthTokenStateUnavailable)
		}
	}
	if err := s.RevokeSession(ctx, in.CurrentClaims.TokenId); err != nil {
		return nil, err
	}
	user := &entity.SysUser{Id: in.CurrentClaims.UserId, Username: in.CurrentClaims.Username, Status: in.CurrentClaims.Status}
	token, tokenID, err := s.generateToken(ctx, user, in.TenantID)
	if err != nil {
		return nil, err
	}
	if err = s.createSession(ctx, user, in.TenantID, tokenID); err != nil {
		logger.Warningf(ctx, "create reissued tenant session failed tokenId=%s tenantId=%d err=%v", tokenID, in.TenantID, err)
	}
	return &TenantTokenOutput{AccessToken: token}, nil
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
	jwtSecret := s.configSvc.GetJwtSecret(ctx)
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, bizerr.NewCode(CodeAuthTokenInvalid)
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
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
func (s *serviceImpl) Logout(ctx context.Context, username string, tenantId int, tokenId string) {
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
}

// RevokeSession removes one online session by token ID and its cached access context.
func (s *serviceImpl) RevokeSession(ctx context.Context, tokenId string) error {
	if tokenId == "" {
		return nil
	}
	if s.roleSvc != nil {
		s.roleSvc.InvalidateTokenAccessContext(ctx, tokenId)
	}
	return s.sessionStore.Delete(ctx, tokenId)
}

// generateToken generates JWT token for given user, returns token string and tokenId.
func (s *serviceImpl) generateToken(ctx context.Context, user *entity.SysUser, tenantID int) (string, string, error) {
	jwtTTL, err := s.configSvc.GetJwtExpire(ctx)
	if err != nil {
		return "", "", err
	}
	var (
		jwtSecret = s.configSvc.GetJwtSecret(ctx)
		tokenId   = guid.S()
	)
	claims := Claims{
		TokenId:  tokenId,
		UserId:   user.Id,
		Username: user.Username,
		Status:   user.Status,
		TenantId: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", err
	}
	return signed, tokenId, nil
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
	var ip, browser, osName string
	if r := g.RequestFromCtx(ctx); r != nil {
		ip = r.GetClientIp()
		ua := useragent.New(r.GetHeader("User-Agent"))
		browserName, browserVersion := ua.Browser()
		browser = browserName + " " + browserVersion
		osName = ua.OS()
	}
	deptName := s.getUserDeptName(ctx, user.Id)
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
	_, err := s.roleSvc.PrimeTokenAccessContext(ctx, tokenID, user.Id)
	return err
}
