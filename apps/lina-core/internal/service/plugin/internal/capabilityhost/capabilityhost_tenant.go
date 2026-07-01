// This file composes the tenant-domain capability exposed through plugin host
// services with governance sub-capabilities owned by adjacent host domains.

package capabilityhost

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	capabilitytenantcap "lina-core/pkg/plugin/capability/tenantcap"
)

// tenantDomainService delegates ordinary tenant operations to the tenant owner
// and attaches tenant-governance sub-capabilities for plugin callers.
type tenantDomainService struct {
	base    capabilitytenantcap.Service
	plugins capabilitytenantcap.PluginService
}

// newTenantDomainService creates the tenant capability view published to plugins.
func newTenantDomainService(
	base capabilitytenantcap.Service,
	plugins capabilitytenantcap.PluginService,
) capabilitytenantcap.Service {
	return tenantDomainService{
		base:    base,
		plugins: plugins,
	}
}

// Available reports whether tenant capability is available.
func (s tenantDomainService) Available(ctx context.Context) bool {
	if s.base == nil {
		return false
	}
	return s.base.Available(ctx)
}

// Status returns the current tenant capability status.
func (s tenantDomainService) Status(ctx context.Context) capmodel.CapabilityStatus {
	if s.base == nil {
		return capmodel.CapabilityStatus{}
	}
	return s.base.Status(ctx)
}

// Context returns current-tenant context operations.
func (s tenantDomainService) Context() capabilitytenantcap.ContextService {
	if s.base == nil {
		return nil
	}
	return s.base.Context()
}

// Directory returns tenant directory operations.
func (s tenantDomainService) Directory() capabilitytenantcap.DirectoryService {
	if s.base == nil {
		return nil
	}
	return s.base.Directory()
}

// Membership returns user-to-tenant membership operations.
func (s tenantDomainService) Membership() capabilitytenantcap.MembershipService {
	if s.base == nil {
		return nil
	}
	return s.base.Membership()
}

// Plugins returns tenant-plugin governance operations.
func (s tenantDomainService) Plugins() capabilitytenantcap.PluginService {
	if s.plugins != nil {
		return s.plugins
	}
	if s.base == nil {
		return nil
	}
	return s.base.Plugins()
}

// Filter returns tenant filter context operations.
func (s tenantDomainService) Filter() capabilitytenantcap.FilterService {
	if s.base == nil {
		return nil
	}
	return s.base.Filter()
}

var _ capabilitytenantcap.Service = (*tenantDomainService)(nil)
