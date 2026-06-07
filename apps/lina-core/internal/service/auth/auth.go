// Package auth implements authentication, JWT issuance, login auditing, and
// online-session persistence for the Lina core host service.
package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	tokencap "lina-core/pkg/plugin/capability/authcap/token"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginhost"
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

// English fallback messages published with host authentication lifecycle hooks.
const (
	// authEventMessageLoginSuccessful is the English fallback for successful login messages.
	authEventMessageLoginSuccessful = "Login successful"
	// authEventMessageInvalidCredentials is the English fallback for invalid credential messages.
	authEventMessageInvalidCredentials = "Invalid username or password"
	// authEventMessageUserDisabled is the English fallback for disabled account messages.
	authEventMessageUserDisabled = "User account is disabled"
	// authEventMessageIPBlacklisted is the English fallback for blocked login IP messages.
	authEventMessageIPBlacklisted = "Login IP is blacklisted"
	// authEventMessageLogoutSuccessful is the English fallback for successful logout messages.
	authEventMessageLogoutSuccessful = "Logout successful"
)

// tokenKind identifies the intended use of one signed JWT. The underlying
// string values are owned by `authcap/token` so host signers, host parsers,
// dynamic plugin routes, and source plugins stay in lock-step.
type tokenKind string

const (
	// tokenKindAccess marks JWTs accepted by protected API middleware.
	tokenKindAccess tokenKind = tokencap.KindAccess
	// tokenKindRefresh marks JWTs accepted only by the refresh-token endpoint.
	tokenKindRefresh tokenKind = tokencap.KindRefresh
	// defaultRefreshTokenTTL is the minimum lifetime for refresh tokens.
	defaultRefreshTokenTTL time.Duration = 7 * 24 * time.Hour
)

// Service defines authentication, token lifecycle, and online-session
// operations used by host HTTP handlers and tenant-aware adapters.
type Service interface {
	// SessionStore returns the shared online-session store used by middleware,
	// cleanup jobs, and forced logout paths. Callers must treat the returned
	// store as runtime-owned state because it may include cluster hot-state
	// coordination.
	SessionStore() session.Store
	// Login verifies credentials, applies login IP policy, resolves tenant
	// candidates, and either issues a token pair or returns a pre-login token
	// for tenant selection. It persists session state and dispatches auth hooks;
	// user-visible failures are returned as bizerr codes.
	Login(ctx context.Context, in LoginInput) (*LoginOutput, error)
	// Refresh validates a host refresh token, confirms the online session and
	// tenant membership are still valid, primes role access cache, and returns a
	// fresh access token while preserving the refresh token.
	Refresh(ctx context.Context, in RefreshInput) (*RefreshOutput, error)
	// ParseToken parses an access token, validates its token kind and revoke
	// marker, and returns claims for middleware context injection.
	ParseToken(ctx context.Context, tokenString string) (*Claims, error)
	// HashPassword hashes a plaintext password with bcrypt for user-account
	// writes; it does not persist the result.
	HashPassword(password string) (string, error)
	// Logout revokes the supplied token ID, clears cached access context,
	// removes the online-session row, and dispatches logout hooks using the
	// current session client type. An empty tokenId only records hook state and
	// leaves sessions unchanged.
	Logout(ctx context.Context, in LogoutInput) error
	// RevokeSession writes a shared revoke marker, removes one online session by
	// token ID, and invalidates token-bound role access cache across callers.
	RevokeSession(ctx context.Context, tokenId string) error
}

// TenantTokenIssuer defines the narrow host-owned token handoff used by
// tenant-aware auth adapters.
type TenantTokenIssuer interface {
	// IssueTenantToken consumes a short-lived pre-login token, validates the
	// selected tenant membership, and issues a tenant-bound token pair.
	IssueTenantToken(ctx context.Context, in TenantTokenIssueInput) (*TenantTokenOutput, error)
	// ReissueTenantToken validates tenant membership, revokes the current token
	// and cached access context, and issues a new tenant-bound token pair.
	ReissueTenantToken(ctx context.Context, in TenantTokenReissueInput) (*TenantTokenOutput, error)
	// ReissueTenantTokenFromBearer parses the current bearer token and delegates
	// to ReissueTenantToken for tenant-switch validation and revocation.
	ReissueTenantTokenFromBearer(ctx context.Context, tokenString string, tenantID int) (*TenantTokenOutput, error)
	// IssueImpersonationToken signs a host-owned impersonation access token for
	// a platform administrator entering target tenant scope. It creates the
	// online-session row and primes role access using platform administrator
	// grants while retaining the target tenant as the request data boundary.
	IssueImpersonationToken(ctx context.Context, in ImpersonationTokenIssueInput) (*ImpersonationTokenOutput, error)
	// RevokeImpersonationToken validates that tokenString is an impersonation
	// access token for tenantID when tenantID is non-zero, then revokes the
	// session and token access cache through the host auth state store.
	RevokeImpersonationToken(ctx context.Context, tokenString string, tenantID int) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// Ensure serviceImpl implements TenantTokenIssuer.
var _ TenantTokenIssuer = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	configSvc    authConfigService // Configuration service
	orgCapSvc    authOrgProjectionService
	hookSvc      authHookService // Authentication lifecycle hook dispatcher
	roleSvc      authRoleService // Role service
	tenantSvc    authTenantAccessService
	sessionStore session.Store // Session store
	preTokens    preTokenStore
	revoked      revokeStore
}

// authHookService is the narrow plugin-hook surface required by auth. Auth
// owns authentication state and publishes hook payloads without depending on
// the plugin service root package.
type authHookService interface {
	// HandleAuthLoginSucceeded dispatches a login-succeeded hook to enabled plugins.
	HandleAuthLoginSucceeded(ctx context.Context, input pluginhost.AuthHookPayloadInput) error
	// HandleAuthLoginFailed dispatches a login-failed hook to enabled plugins.
	HandleAuthLoginFailed(ctx context.Context, input pluginhost.AuthHookPayloadInput) error
	// HandleAuthLogoutSucceeded dispatches a logout-succeeded hook to enabled plugins.
	HandleAuthLogoutSucceeded(ctx context.Context, input pluginhost.AuthHookPayloadInput) error
}

// authConfigService is the narrow config surface used by auth.
type authConfigService interface {
	// GetJwtSecret returns the signing secret used for host access and refresh
	// tokens. Callers treat an empty value as a configuration error at token
	// issue or parse time rather than caching it locally.
	GetJwtSecret(ctx context.Context) string
	// GetJwtExpire returns the configured token lifetime. It may return a
	// configuration parsing error when the duration value is invalid.
	GetJwtExpire(ctx context.Context) (time.Duration, error)
	// GetSessionTimeout returns the online-session timeout used for session and
	// token revocation bookkeeping. It returns configuration errors from the
	// underlying system-config service unchanged.
	GetSessionTimeout(ctx context.Context) (time.Duration, error)
	// IsLoginIPBlacklisted reports whether the given client IP is denied by
	// login policy. It returns lookup or configuration errors so authentication
	// can fail closed when policy state cannot be trusted.
	IsLoginIPBlacklisted(ctx context.Context, ip string) (bool, error)
}

// authRoleService is the narrow role access-cache surface used by auth.
type authRoleService interface {
	// PrimeTokenAccessContext builds and caches the role/menu/data-scope access
	// snapshot for one issued token. It returns permission-data lookup errors so
	// token issuance can be rolled back when access context cannot be prepared.
	PrimeTokenAccessContext(ctx context.Context, tokenID string, userID int) (*role.UserAccessContext, error)
	// InvalidateTokenAccessContext removes the cached access snapshot for one
	// token. Missing snapshots are treated as success because revocation paths
	// must remain idempotent.
	InvalidateTokenAccessContext(ctx context.Context, tokenID string)
}

// authOrgProjectionService is the organization read slice used by auth session
// projection. It deliberately excludes organization assignment and data-scope
// query builder methods.
type authOrgProjectionService interface {
	// GetUserDeptName returns one user's department display name for login and
	// online-session projection. Missing organization capability should degrade
	// to an empty name.
	GetUserDeptName(ctx context.Context, userID int) (string, error)
}

// authTenantAccessService is the tenant membership slice required by auth. It
// excludes tenant query-scope, provisioning, and startup consistency methods.
type authTenantAccessService interface {
	// Available reports whether tenant membership checks should run.
	Available(ctx context.Context) bool
	// ListUserTenants returns active tenant candidates visible to one login user.
	ListUserTenants(ctx context.Context, userID int) ([]tenantcapsvc.TenantInfo, error)
	// ValidateUserInTenant verifies that userID may issue a token for tenantID.
	ValidateUserInTenant(ctx context.Context, userID int, tenantID tenantcapsvc.TenantID) error
	// SwitchTenant validates a tenant switch before token reissue.
	SwitchTenant(ctx context.Context, userID int, target tenantcapsvc.TenantID) error
}

// New creates the concrete auth service from explicit runtime-owned dependencies.
func New(configSvc authConfigService, hookSvc authHookService, orgCapSvc authOrgProjectionService, roleSvc authRoleService, tenantSvc authTenantAccessService, sessionStore session.Store, kvCacheSvc kvcache.Service) Service {
	return &serviceImpl{
		configSvc:    configSvc,
		orgCapSvc:    orgCapSvc,
		hookSvc:      hookSvc,
		roleSvc:      roleSvc,
		tenantSvc:    tenantSvc,
		sessionStore: sessionStore,
		preTokens:    newKVPreTokenStore(kvCacheSvc),
		revoked:      newLayeredRevokeStore(newMemoryRevokeStore(), newKVRevokeStore(kvCacheSvc)),
	}
}

// Claims defines JWT token claims.
type Claims struct {
	TokenId         string     `json:"tokenId"`         // Unique token identifier
	TokenType       tokenKind  `json:"tokenType"`       // TokenType identifies access or refresh token usage.
	ClientType      ClientType `json:"clientType"`      // ClientType identifies the user-session client.
	UserId          int        `json:"userId"`          // User ID
	Username        string     `json:"username"`        // Username
	Status          int        `json:"status"`          // Status
	TenantId        int        `json:"tenantId"`        // Tenant ID, where 0 means platform
	IsImpersonation bool       `json:"isImpersonation"` // Whether the token represents impersonation
	ActingUserId    int        `json:"actingUserId"`    // Real user ID during impersonation
	jwt.RegisteredClaims
}

// LoginInput defines input for Login function.
type LoginInput struct {
	Username   string     // Username
	Password   string     // Password
	ClientType ClientType // User-session client type
}

// LogoutInput defines input for Logout function.
type LogoutInput struct {
	Username   string     // Username
	TenantID   int        // Tenant ID, where 0 means platform
	TokenID    string     // Token/session identifier
	ClientType ClientType // User-session client type
}

// LoginOutput defines output for Login function.
type LoginOutput struct {
	AccessToken  string       // JWT access token
	RefreshToken string       // JWT refresh token
	PreToken     string       // Short-lived pre-login token for tenant selection
	Tenants      []TenantInfo // Tenant candidates for two-stage login
}

// RefreshInput defines input for Refresh function.
type RefreshInput struct {
	RefreshToken string // JWT refresh token
}

// RefreshOutput defines output for Refresh function.
type RefreshOutput struct {
	AccessToken  string // Newly issued JWT access token
	RefreshToken string // Refresh token that remains valid for the session
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
	AccessToken  string // JWT access token
	RefreshToken string // JWT refresh token
}

// ImpersonationTokenIssueInput defines host-owned impersonation token issue fields.
type ImpersonationTokenIssueInput struct {
	ActingUserID int // Platform administrator user ID
	TenantID     int // Target tenant ID
}

// ImpersonationTokenOutput defines a host-owned impersonation token response.
type ImpersonationTokenOutput struct {
	AccessToken  string // JWT access token
	TokenID      string // Host token/session identifier
	TenantID     int    // Target tenant ID
	ActingUserID int    // Platform administrator user ID
}
