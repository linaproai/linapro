// This file keeps role service tests aligned with explicit dependency
// injection when individual fakes are replaced after construction.

package role

import (
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// newRoleTestService constructs a role service with explicit test dependencies,
// including the shared data-scope service required by role user operations.
func newRoleTestService(permissionFilter PermissionMenuFilter, organizationState OrganizationCapabilityState) *serviceImpl {
	var (
		bizCtxSvc = bizctx.New()
		configSvc = hostconfig.New()
		i18nSvc   = i18nsvc.New(bizCtxSvc, configSvc, cachecoord.Default(nil))
		orgCapSvc = orgspi.New(nil, nil)
		tenantSvc = tenantspi.New(nil, nil, nil)
	)
	if organizationState == nil {
		organizationState = orgCapSvc
	}
	svc := New(permissionFilter, bizCtxSvc, configSvc, i18nSvc, organizationState, tenantSvc).(*serviceImpl)
	refreshRoleTestScope(svc, orgCapSvc)
	return svc
}

// newDefaultRoleTestService constructs the default role service used by most tests.
func newDefaultRoleTestService() *serviceImpl {
	return newRoleTestService(nil, nil)
}

// setRoleTestBizCtx replaces the business context dependency and refreshes
// the derived data-scope service used by role-management tests.
func setRoleTestBizCtx(svc *serviceImpl, bizCtxSvc bizctx.Service) {
	svc.bizCtxSvc = bizCtxSvc
	refreshRoleTestScope(svc, nil)
}

// refreshRoleTestScope rebuilds the stateless data-scope helper from the
// current explicit fake dependencies.
func refreshRoleTestScope(svc *serviceImpl, orgCapSvc orgcap.Service) {
	var orgScope orgspi.ScopeService
	if scope, ok := orgCapSvc.(orgspi.ScopeService); ok {
		orgScope = scope
	}
	svc.SetDataScopeService(datascope.New(svc.bizCtxSvc, svc, orgScope))
}
