// Package auth implements authentication, JWT issuance, login auditing, and
// online-session persistence for the Lina core host service.
package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"

	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/authcap/extlogin/extidspi"
	tokencap "lina-core/pkg/plugin/capability/authcap/token"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/pluginhost"
)

// ClientType identifies the user-facing client that created an authenticated user session.
type ClientType = tokencap.ClientType

// ParseClientType validates one user-session client type value.
func ParseClientType(value string) (ClientType, error) {
	if clientType, ok := tokencap.ParseClientType(value); ok {
		return clientType, nil
	}
	return "", bizerr.NewCode(CodeAuthClientTypeInvalid)
}

// Auth status constants used by login validation.
const (
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
	// authEventMessageTenantUnavailable is the English fallback for tenant-auth rejection messages.
	authEventMessageTenantUnavailable = "Tenant is not available"
	// authEventMessageLogoutSuccessful is the English fallback for successful logout messages.
	authEventMessageLogoutSuccessful = "Logout successful"
	// authEventMessageExternalNotProvisioned is the English fallback for unlinked external-identity messages.
	authEventMessageExternalNotProvisioned = "No local account is linked to this external identity"
	// authHookReasonTenantUnavailable identifies tenant service or membership failures.
	authHookReasonTenantUnavailable = "tenant_unavailable"
	// authHookReasonExternalNotProvisioned identifies external-identity logins with no linked local account.
	authHookReasonExternalNotProvisioned = "external_not_provisioned"
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
	// BindExternalIdentityProvider attaches the source-plugin external-identity
	// provider seam after startup wiring. The provider owns external-identity
	// storage, resolution, and provisioning/bind policy; auth keeps token,
	// session, and tenant minting. A nil provider keeps external login
	// fail-closed: no linkage is resolved and no account is created.
	BindExternalIdentityProvider(provider extidspi.Provider)
	// LoginByExternalIdentity resolves a plugin-verified external identity
	// (provider + immutable subject) to a linked local account and issues a
	// host session, reusing the same login-IP policy, disabled-account check,
	// tenant resolution, pre-login-token handoff, token issuance, session
	// persistence, and auth hooks as password Login. Linkage storage and
	// provisioning policy are provider-owned and fail-closed by default: an
	// unlinked identity with no provider, or with auto-provisioning disallowed,
	// returns CodeAuthExternalUserNotProvisioned without creating a user. Callers must
	// have already verified the external identity; the host does not perform
	// any OAuth or token exchange. An empty provider or subject returns
	// CodeAuthExternalIdentityInvalid.
	// SPA delivery of the minted session (one-time handoff codes) is owned by
	// linapro-extlogin-core, not this host service.
	LoginByExternalIdentity(ctx context.Context, in ExternalLoginInput) (*ExternalLoginOutput, error)
	// Refresh validates a host refresh token, confirms the online session and
	// tenant membership are still valid, primes role access cache, and returns a
	// fresh access token while preserving the refresh token.
	Refresh(ctx context.Context, in RefreshInput) (*RefreshOutput, error)
	// AuthenticateAccessToken validates an access token and confirms the
	// corresponding sys_online_session row is present, tenant-bound, and not
	// timed out. It returns claims only after the complete login state is valid.
	AuthenticateAccessToken(ctx context.Context, tokenString string) (*Claims, error)
	// IssueTenantToken consumes a pre-login token and issues a tenant-bound token
	// pair while validating tenant membership and creating online-session state.
	IssueTenantToken(ctx context.Context, in TenantTokenIssueInput) (*TenantTokenOutput, error)
	// ReissueTenantTokenFromBearer parses the current access token, validates a
	// tenant switch, revokes the old session, and issues a new tenant token pair.
	ReissueTenantTokenFromBearer(ctx context.Context, tokenString string, tenantID int) (*TenantTokenOutput, error)
	// IssueImpersonationToken signs and registers one host-owned impersonation
	// token using the shared auth/session state.
	IssueImpersonationToken(ctx context.Context, in ImpersonationTokenIssueInput) (*ImpersonationTokenOutput, error)
	// RevokeImpersonationToken validates one host impersonation token and revokes
	// its online-session state when it belongs to the supplied tenant.
	RevokeImpersonationToken(ctx context.Context, tokenString string, tenantID int) error
	// HashPassword hashes a plaintext password with bcrypt for user-account
	// writes; it does not persist the result.
	HashPassword(password string) (string, error)
	// Register creates one public platform account when self-registration is
	// enabled. It enforces username/email uniqueness, assigns the built-in
	// standard user role, and does not issue a login session.
	Register(ctx context.Context, in RegisterInput) (*RegisterOutput, error)
	// RequestPasswordReset starts email password recovery when the feature is
	// enabled and mail delivery is available. The response is always success
	// shaped when the request is accepted so callers cannot enumerate accounts.
	RequestPasswordReset(ctx context.Context, in PasswordResetRequestInput) error
	// ConfirmPasswordReset consumes a one-time reset token and updates the
	// account password. Existing online sessions for the user are removed.
	ConfirmPasswordReset(ctx context.Context, in PasswordResetConfirmInput) error
	// Logout revokes the supplied token ID, clears cached access context,
	// removes the online-session row, and dispatches logout hooks using the
	// current session client type. An empty tokenId only records hook state and
	// leaves sessions unchanged.
	Logout(ctx context.Context, in LogoutInput) error
	// RevokeSession writes a shared revoke marker, removes one online session by
	// token ID, and invalidates token-bound role access cache across callers.
	RevokeSession(ctx context.Context, tokenId string) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	configSvc    configsvc.Service // Configuration service
	orgCapSvc    orgcap.Service
	hookSvc      authHookService // Authentication lifecycle hook dispatcher
	roleSvc      role.Service    // Role service
	tenantSvc    tenantspi.Service
	sessionStore session.Store // Session store
	preTokens    preTokenStore
	resetTokens  passwordResetStore
	rateLimit    rateLimitStore
	kvCache      kvcache.Service
	revoked      revokeStore
	// identityProvider is the source-plugin external-identity provider bound
	// after startup through BindExternalIdentityProvider. Nil keeps external
	// login fail-closed: LoginByExternalIdentity resolves no linkage and
	// provisions no account.
	identityProvider extidspi.Provider
}

// authHookService is the narrow plugin-hook surface required by auth. Auth
// owns authentication state and publishes hook payloads without depending on
// the plugin service root package.
type authHookService interface {
	// DispatchHookEvent dispatches one plugin hook event with already-normalized payload values.
	DispatchHookEvent(ctx context.Context, event pluginhost.ExtensionPoint, values map[string]interface{}) error
}

// New creates the concrete auth service from explicit runtime-owned dependencies.
func New(
	configSvc configsvc.Service,
	hookSvc authHookService,
	orgCapSvc orgcap.Service,
	roleSvc role.Service,
	tenantSvc tenantspi.Service,
	sessionStore session.Store,
	kvCacheSvc kvcache.Service,
) Service {
	return &serviceImpl{
		configSvc:    configSvc,
		orgCapSvc:    orgCapSvc,
		hookSvc:      hookSvc,
		roleSvc:      roleSvc,
		tenantSvc:    tenantSvc,
		sessionStore: sessionStore,
		preTokens:    newKVPreTokenStore(kvCacheSvc),
		resetTokens:  newKVPasswordResetStore(kvCacheSvc),
		rateLimit:    newKVRateLimitStore(kvCacheSvc),
		kvCache:      kvCacheSvc,
		revoked:      newLayeredRevokeStore(newMemoryRevokeStore(), newKVRevokeStore(kvCacheSvc)),
	}
}

// RegisterInput defines input for public self-registration.
type RegisterInput struct {
	Username string // Unique login name
	Password string // Plaintext password
	Email    string // Recovery email
	Nickname string // Optional display name
}

// RegisterOutput defines output for public self-registration.
type RegisterOutput struct {
	UserID int // Created platform user ID
}

// PasswordResetRequestInput defines input for requesting a password-reset email.
type PasswordResetRequestInput struct {
	Email             string // Account email
	PublicOrigin      string // Browser origin used to build the reset link
	WorkspaceBasePath string // Admin workspace base path (for example /admin)
}

// PasswordResetConfirmInput defines input for confirming a password reset.
type PasswordResetConfirmInput struct {
	Token    string // One-time reset token
	Password string // New plaintext password
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

// ExternalLoginInput defines input for LoginByExternalIdentity. The caller is
// a source plugin that has already verified the external identity; PluginID is
// stamped by the host-scoped capability adapter and must not be trusted from
// unauthenticated request payloads.
type ExternalLoginInput struct {
	PluginID    string     // Source-plugin ID that owns Provider; stamped by the host adapter
	Provider    string     // Stable external provider ID owned by PluginID
	Subject     string     // Immutable provider-issued subject identifier
	Email       string     // Verified email; used for provisioning/conflict checks only when AllowAutoProvision is set
	DisplayName string     // Display name captured for audit/hook context only
	ClientType  ClientType // User-session client type
	// AllowAutoProvision declares that the calling plugin permits
	// auto-provisioning for unlinked identities. The provisioning policy is
	// provider-owned (linapro-extlogin-core): a same-email account conflict is
	// rejected instead of silently linking, and account creation is delegated
	// back to the host user owner's least-privilege provisioning path.
	AllowAutoProvision bool
}

// ExternalLoginOutput defines output for LoginByExternalIdentity. It mirrors
// LoginOutput: a token pair is set for 0/1 active tenants, otherwise a
// pre-login token plus tenant candidates are returned for tenant selection.
type ExternalLoginOutput struct {
	AccessToken  string       // JWT access token
	RefreshToken string       // JWT refresh token
	PreToken     string       // Short-lived pre-login token for tenant selection
	Tenants      []TenantInfo // Tenant candidates for two-stage login
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
	CurrentClaims         *Claims // Current token claims
	SkipSessionValidation bool    // Skip session validation when the caller already completed it in the same request.
	TenantID              int     // Target tenant ID
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
