// Package shared defines plugin-local constants and table helpers shared by
// multi-tenant service packages.
package shared

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// Plugin-local table names.
const (
	TableTenant     = "plugin_multi_tenant_tenant"
	TableMembership = "plugin_multi_tenant_user_membership"
	TableSysUser    = "sys_user"
)

// TenantStatus is the tenant lifecycle status stored by the plugin.
type TenantStatus string

const (
	// TenantStatusActive allows normal tenant access.
	TenantStatusActive TenantStatus = "active"
	// TenantStatusSuspended blocks tenant user access until resumed.
	TenantStatusSuspended TenantStatus = "suspended"
)

// Resolver names supported by the plugin-owned resolution chain.
const (
	ResolverOverride  = "override"
	ResolverJWT       = "jwt"
	ResolverSession   = "session"
	ResolverHeader    = "header"
	ResolverSubdomain = "subdomain"
	ResolverDefault   = "default"
)

// defaultResolverChain is the code-owned request tenant resolution order.
//
// Resolution is intentionally fixed in code rather than stored in a plugin
// table. Override is evaluated before JWT so platform administrators can enter
// audited tenant impersonation. JWT remains authoritative for normal
// authenticated requests, so header and subdomain values are only login-stage
// hints and can never override a signed tenant claim. Session keeps browser
// continuity after tenant selection, while default is the final membership
// fallback used when the request still has no tenant decision.
var defaultResolverChain = []string{
	ResolverOverride,
	ResolverJWT,
	ResolverSession,
	ResolverHeader,
	ResolverSubdomain,
	ResolverDefault,
}

// Resolver behavior values.
const (
	// OnAmbiguousPrompt asks the caller to select a tenant when no resolver can
	// determine one unambiguously.
	OnAmbiguousPrompt = "prompt"
	// OnAmbiguousReject rejects requests without an unambiguous tenant instead
	// of presenting a selector.
	OnAmbiguousReject = "reject"
	// OnAmbiguousFirstOwned chooses the first active membership as a convenience
	// fallback. It is retained as a supported runtime policy but is not the code
	// default.
	OnAmbiguousFirstOwned = "first_owned"
)

// Built-in tenant configuration defaults. They are intentionally code-owned
// instead of host config-file values so new installations cannot drift before
// the tenant management UI exposes supported settings.
const (
	// DefaultIsolationMode selects the Pool model: one database/schema with
	// tenant_id columns on tenant-sensitive tables. This is the only supported
	// isolation mode in the first multi-tenant iteration.
	DefaultIsolationMode = "pool"
	// DefaultCardinality allows one user identity to hold memberships in multiple
	// tenants. This matches internal BU usage and is a superset of single-tenant
	// membership.
	DefaultCardinality = "multi"
	// SingleCardinality allows at most one active tenant membership per user.
	SingleCardinality = "single"
	// DefaultRootDomain is empty for now, which disables subdomain tenant
	// resolution. Root-domain configuration will be exposed in a later iteration.
	DefaultRootDomain = ""
	// DefaultTenantCodeHeader is the login-stage tenant hint header. Authenticated
	// business requests must not use it to override JWT TenantId.
	DefaultTenantCodeHeader = "X-Tenant-Code"
	// DefaultTenantOverrideHeader is the platform-admin impersonation header.
	// It is accepted only after platform authorization checks.
	DefaultTenantOverrideHeader = "X-Tenant-Override"
)

// defaultReservedSubdomains are labels that must never be interpreted as tenant
// codes when subdomain resolution is enabled later.
var defaultReservedSubdomains = []string{"www", "api", "admin", "static", "docs"}

// DefaultResolverChain returns a detached copy of the built-in request tenant
// resolution order.
func DefaultResolverChain() []string {
	return cloneStrings(defaultResolverChain)
}

// DefaultReservedSubdomains returns a detached copy of the built-in reserved
// subdomain labels.
func DefaultReservedSubdomains() []string {
	return cloneStrings(defaultReservedSubdomains)
}

// cloneStrings returns a detached copy of a string slice.
func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

// Membership status values.
const (
	MembershipStatusDisabled = 0
	MembershipStatusEnabled  = 1
)

// PlatformTenantID is the tenant ID used for platform context.
const PlatformTenantID int64 = 0

// Model returns a context-bound GoFrame model for plugin-local table access.
func Model(ctx context.Context, table string) *gdb.Model {
	return g.DB().Model(table).Safe().Ctx(ctx)
}
