// This file contains the built-in tenant resolver implementations.

package resolver

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	"lina-plugin-multi-tenant/backend/internal/service/membership"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
)

// overrideResolver resolves platform impersonation from X-Tenant-Override.
type overrideResolver struct{}

// Name returns the configured resolver name.
func (r overrideResolver) Name() string {
	return shared.ResolverOverride
}

// Resolve returns a tenant result from X-Tenant-Override.
func (r overrideResolver) Resolve(ctx context.Context, request *ghttp.Request, identity Identity, config Config) (*Result, bool, error) {
	value := strings.TrimSpace(request.Header.Get(shared.DefaultTenantOverrideHeader))
	if value == "" {
		return nil, false, nil
	}
	if !identity.IsPlatform {
		return nil, true, bizerr.NewCode(CodePlatformPermissionRequired)
	}
	tenantID, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, true, bizerr.WrapCode(err, CodeTenantOverrideInvalid)
	}
	return &Result{TenantID: tenantID, Source: shared.ResolverOverride, ActingAsTenant: true}, true, nil
}

// jwtResolver resolves TenantId from host business context set by the auth layer.
type jwtResolver struct{}

// Name returns the configured resolver name.
func (r jwtResolver) Name() string {
	return shared.ResolverJWT
}

// Resolve returns a tenant result from host-attached JWT metadata.
func (r jwtResolver) Resolve(ctx context.Context, request *ghttp.Request, identity Identity, config Config) (*Result, bool, error) {
	if identity.TenantID == shared.PlatformTenantID {
		return nil, false, nil
	}
	return &Result{
		TenantID:       identity.TenantID,
		Source:         shared.ResolverJWT,
		ActingAsTenant: identity.ActingAsTenant || identity.IsImpersonation,
	}, true, nil
}

// sessionResolver resolves TenantId from a session value.
type sessionResolver struct{}

// Name returns the configured resolver name.
func (r sessionResolver) Name() string {
	return shared.ResolverSession
}

// Resolve returns a tenant result from session metadata.
func (r sessionResolver) Resolve(ctx context.Context, request *ghttp.Request, identity Identity, config Config) (*Result, bool, error) {
	value := request.Session.MustGet("tenant_id").Int64()
	if value == 0 {
		return nil, false, nil
	}
	return &Result{TenantID: value, Source: shared.ResolverSession}, true, nil
}

// headerResolver resolves a login-stage tenant hint from X-Tenant-Code.
type headerResolver struct{}

// Name returns the configured resolver name.
func (r headerResolver) Name() string {
	return shared.ResolverHeader
}

// Resolve returns a tenant result from X-Tenant-Code.
func (r headerResolver) Resolve(ctx context.Context, request *ghttp.Request, identity Identity, config Config) (*Result, bool, error) {
	code := strings.TrimSpace(request.Header.Get(shared.DefaultTenantCodeHeader))
	if code == "" {
		return nil, false, nil
	}
	return findTenantByCode(ctx, code, shared.ResolverHeader)
}

// subdomainResolver resolves a login-stage tenant hint from host subdomain.
type subdomainResolver struct{}

// Name returns the configured resolver name.
func (r subdomainResolver) Name() string {
	return shared.ResolverSubdomain
}

// Resolve returns a tenant result from the first host label.
func (r subdomainResolver) Resolve(ctx context.Context, request *ghttp.Request, identity Identity, config Config) (*Result, bool, error) {
	label := subdomainLabel(request.Host, config.RootDomain)
	if label == "" {
		return nil, false, nil
	}
	if isReservedSubdomain(label, config.ReservedSubdomains) {
		return nil, false, nil
	}
	return findTenantByCode(ctx, label, shared.ResolverSubdomain)
}

// userTenantLister lists tenant memberships needed by the default resolver.
type userTenantLister interface {
	// ListUserTenants returns enabled tenant memberships for one user.
	ListUserTenants(ctx context.Context, userID int64) ([]*membership.TenantInfo, error)
}

// defaultResolver chooses the first enabled membership for tenant users.
type defaultResolver struct {
	membershipSvc userTenantLister
}

// Name returns the configured resolver name.
func (r defaultResolver) Name() string {
	return shared.ResolverDefault
}

// Resolve returns a tenant result from the user's first enabled membership.
func (r defaultResolver) Resolve(ctx context.Context, request *ghttp.Request, identity Identity, config Config) (*Result, bool, error) {
	if identity.UserID == 0 {
		return &Result{TenantID: shared.PlatformTenantID, Source: shared.ResolverDefault}, true, nil
	}
	if identity.IsPlatform && !identity.ActingAsTenant && !identity.IsImpersonation {
		return &Result{TenantID: shared.PlatformTenantID, Source: shared.ResolverDefault}, true, nil
	}
	tenants, err := r.membershipSvc.ListUserTenants(ctx, identity.UserID)
	if err != nil {
		return nil, true, err
	}
	if len(tenants) == 0 {
		return &Result{TenantID: shared.PlatformTenantID, Source: shared.ResolverDefault}, true, nil
	}
	return &Result{TenantID: tenants[0].Id, Source: shared.ResolverDefault}, true, nil
}

// subdomainLabel extracts the tenant label from a request host and optional root domain.
func subdomainLabel(host string, rootDomain string) string {
	hostname := strings.ToLower(strings.TrimSpace(host))
	if colon := strings.LastIndex(hostname, ":"); colon >= 0 {
		hostname = hostname[:colon]
	}
	root := strings.Trim(strings.ToLower(strings.TrimSpace(rootDomain)), ".")
	if root == shared.DefaultRootDomain {
		return ""
	}
	suffix := "." + root
	if !strings.HasSuffix(hostname, suffix) {
		return ""
	}
	label := strings.TrimSuffix(hostname, suffix)
	if label == "" || strings.Contains(label, ".") {
		return ""
	}
	return label
}

// isReservedSubdomain reports whether label is blocked by resolver policy.
func isReservedSubdomain(label string, configured []string) bool {
	reserved := configured
	if len(reserved) == 0 {
		reserved = shared.DefaultReservedSubdomains()
	}
	for _, item := range reserved {
		if label == strings.ToLower(strings.TrimSpace(item)) {
			return true
		}
	}
	return false
}

// findTenantByCode resolves a tenant ID by its stable code.
func findTenantByCode(ctx context.Context, code string, source string) (*Result, bool, error) {
	row := struct {
		Id int64 `json:"id" orm:"id"`
	}{}
	err := shared.Model(ctx, shared.TableTenant).Fields("id").Where("code", code).Scan(&row)
	if err != nil {
		if gerror.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, true, err
	}
	if row.Id == 0 {
		return nil, false, nil
	}
	return &Result{TenantID: row.Id, Source: source}, true, nil
}
