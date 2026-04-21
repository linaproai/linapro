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
	"lina-core/pkg/logger"
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
	Logout(ctx context.Context, username string, tokenId string)
	// RevokeSession removes one online session and its cached access context.
	RevokeSession(ctx context.Context, tokenId string) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	configSvc    config.Service    // Configuration service
	orgCapSvc    orgcap.Service    // Optional organization capability service
	pluginSvc    pluginsvc.Service // Plugin service
	roleSvc      role.Service      // Role service
	sessionStore session.Store     // Session store
}

// New creates and returns a new Service instance.
// Pass a non-nil orgCapSvc to reuse a caller-owned organization capability
// service; pass nil to create the default orgcap service bound to the default
// plugin service instance.
func New(orgCapSvc orgcap.Service) Service {
	pluginSvc := pluginsvc.New(nil)
	if orgCapSvc == nil {
		orgCapSvc = orgcap.New(pluginSvc)
	}
	return &serviceImpl{
		configSvc:    config.New(),
		orgCapSvc:    orgCapSvc,
		pluginSvc:    pluginSvc,
		roleSvc:      role.New(pluginSvc),
		sessionStore: session.NewDBStore(),
	}
}

// SessionStore returns the session store instance.
func (s *serviceImpl) SessionStore() session.Store {
	return s.sessionStore
}

// Claims defines JWT token claims.
type Claims struct {
	TokenId  string `json:"tokenId"`  // Unique token identifier
	UserId   int    `json:"userId"`   // User ID
	Username string `json:"username"` // Username
	Status   int    `json:"status"`   // Status
	jwt.RegisteredClaims
}

// LoginInput defines input for Login function.
type LoginInput struct {
	Username string // Username
	Password string // Password
}

// LoginOutput defines output for Login function.
type LoginOutput struct {
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

	dispatchLoginFailed := func(username string, msg string) {
		if hookErr := s.pluginSvc.HandleAuthLoginFailed(ctx, pluginsvc.AuthLoginSucceededInput{
			UserName:   username,
			Status:     authLoginStatusFail,
			Ip:         ip,
			ClientType: "web",
			Browser:    browser,
			Os:         osName,
			Message:    msg,
		}); hookErr != nil {
			logger.Warningf(ctx, "plugin login failed hook failed: %v", hookErr)
		}
	}

	if s.configSvc.IsLoginIPBlacklisted(ctx, ip) {
		dispatchLoginFailed(in.Username, "登录IP已被禁止")
		return nil, gerror.New("登录IP已被禁止")
	}

	// Query user by username (GoFrame auto-adds deleted_at IS NULL condition)
	var user *entity.SysUser
	err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Username: in.Username}).
		Scan(&user)
	if err != nil {
		return nil, err
	}
	if user == nil {
		dispatchLoginFailed(in.Username, "用户名或密码错误")
		return nil, gerror.New("用户名或密码错误")
	}

	// Verify password
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		dispatchLoginFailed(in.Username, "用户名或密码错误")
		return nil, gerror.New("用户名或密码错误")
	}

	// Check status
	if user.Status == statusDisabled {
		dispatchLoginFailed(in.Username, "用户已停用")
		return nil, gerror.New("用户已停用")
	}

	// Generate JWT token
	token, tokenId, err := s.generateToken(ctx, user)
	if err != nil {
		return nil, err
	}

	// Record login time
	if _, err = dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: user.Id}).
		Data(do.SysUser{LoginDate: gtime.Now()}).
		Update(); err != nil {
		return nil, gerror.Wrap(err, "更新最后登录时间失败")
	}

	// Create online session
	deptName := s.getUserDeptName(ctx, user.Id)
	if err = s.sessionStore.Set(ctx, &session.Session{
		TokenId:   tokenId,
		UserId:    user.Id,
		Username:  user.Username,
		DeptName:  deptName,
		Ip:        ip,
		Browser:   browser,
		Os:        osName,
		LoginTime: gtime.Now(),
	}); err != nil {
		logger.Warningf(ctx, "create online session failed tokenId=%s err=%v", tokenId, err)
	} else if _, err = s.roleSvc.PrimeTokenAccessContext(ctx, tokenId, user.Id); err != nil {
		logger.Warningf(ctx, "prime access context cache failed tokenId=%s err=%v", tokenId, err)
	}

	if err := s.pluginSvc.HandleAuthLoginSucceeded(ctx, pluginsvc.AuthLoginSucceededInput{
		UserName:   in.Username,
		Status:     authLoginStatusSuccess,
		Ip:         ip,
		ClientType: "web",
		Browser:    browser,
		Os:         osName,
		Message:    "登录成功",
	}); err != nil {
		logger.Warningf(ctx, "plugin login succeeded hook failed: %v", err)
	}
	return &LoginOutput{AccessToken: token}, nil
}

// ParseToken parses and validates JWT token, returns claims.
func (s *serviceImpl) ParseToken(ctx context.Context, tokenString string) (*Claims, error) {
	jwtSecret := s.configSvc.GetJwtSecret(ctx)
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, gerror.New("无效的Token")
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, gerror.New("无效的Token")
}

// HashPassword hashes password using bcrypt.
func (s *serviceImpl) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", gerror.Wrap(err, "密码加密失败")
	}
	return string(hash), nil
}

// Logout records logout login log and removes session.
func (s *serviceImpl) Logout(ctx context.Context, username string, tokenId string) {
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
			logger.Warningf(ctx, "revoke session during logout failed tokenId=%s err=%v", tokenId, err)
		}
	}
	if err := s.pluginSvc.HandleAuthLogoutSucceeded(ctx, pluginsvc.AuthLoginSucceededInput{
		UserName:   username,
		Status:     authLoginStatusSuccess,
		Ip:         ip,
		ClientType: "web",
		Browser:    browser,
		Os:         osName,
		Message:    "登出成功",
	}); err != nil {
		logger.Warningf(ctx, "plugin logout succeeded hook failed: %v", err)
	}
}

// RevokeSession removes one online session and its cached access context.
func (s *serviceImpl) RevokeSession(ctx context.Context, tokenId string) error {
	if tokenId == "" {
		return nil
	}
	s.roleSvc.InvalidateTokenAccessContext(ctx, tokenId)
	return s.sessionStore.Delete(ctx, tokenId)
}

// generateToken generates JWT token for given user, returns token string and tokenId.
func (s *serviceImpl) generateToken(ctx context.Context, user *entity.SysUser) (string, string, error) {
	var (
		jwtTTL    = s.configSvc.GetJwtExpire(ctx)
		jwtSecret = s.configSvc.GetJwtSecret(ctx)
		tokenId   = guid.S()
	)
	claims := Claims{
		TokenId:  tokenId,
		UserId:   user.Id,
		Username: user.Username,
		Status:   user.Status,
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
