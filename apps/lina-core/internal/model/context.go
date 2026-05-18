// This file defines the request-scoped business context shared across host
// middleware, controllers, and services.

package model

// Context is the business context for each request.
type Context struct {
	// TokenId is the authenticated session or access-token identifier for the
	// current request.
	// Example: "session_20260507153000_admin".
	TokenId string `json:"tokenId"`

	// UserId is the current authenticated user's numeric ID.
	// Example: 1.
	UserId int `json:"userId"`

	// Username is the current authenticated user's login name.
	// Example: "admin".
	Username string `json:"username"`

	// Status is the current authenticated user's account status, where 1 means
	// enabled and 0 means disabled.
	// Example: 1.
	Status int `json:"status"`

	// TenantId is the current request tenant, where 0 means platform context.
	// Example: 1001.
	TenantId int `json:"tenantId"`

	// ActingAsTenant reports that a platform user is operating through a tenant view.
	// Example: true.
	ActingAsTenant bool `json:"actingAsTenant"`

	// ActingUserId is the real user ID when impersonation is active.
	// Example: 1.
	ActingUserId int `json:"actingUserId"`

	// IsImpersonation reports whether the current token or override is impersonating a tenant.
	// Example: false.
	IsImpersonation bool `json:"isImpersonation"`

	// DataScope is the effective role data scope cached for the current request:
	// 0 means no governed data access, 1 means all data, 2 means current
	// tenant data, 3 means current department data, and 4 means self-owned data.
	// Example: 3.
	DataScope int `json:"dataScope"`

	// DataScopeUnsupported reports whether an enabled role carries a data scope
	// value that the host cannot interpret.
	// Example: false.
	DataScopeUnsupported bool `json:"dataScopeUnsupported"`

	// UnsupportedDataScope stores the first unsupported role data-scope value
	// when DataScopeUnsupported is true.
	// Example: 99.
	UnsupportedDataScope int `json:"unsupportedDataScope"`

	// Locale is the resolved request locale used by runtime i18n projection.
	// Example: "zh-CN".
	Locale string `json:"locale"`
}
