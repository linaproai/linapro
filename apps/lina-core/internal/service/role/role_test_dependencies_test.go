// This file keeps role service tests aligned with explicit dependency
// injection when individual fakes are replaced after construction.

package role

import (
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/orgcap"
	tenantcapsvc "lina-core/internal/service/tenantcap"
)

// newRoleTestService constructs a role service with explicit test dependencies,
// including the shared data-scope service required by role user operations.
func newRoleTestService(permissionFilter PermissionMenuFilter, organizationState OrganizationCapabilityState) *serviceImpl {
	var (
		bizCtxSvc = bizctx.New()
		configSvc = hostconfig.New()
		i18nSvc   = i18nsvc.New(bizCtxSvc, configSvc, cachecoord.Default(nil))
		orgCapSvc = orgcap.New(nil)
		tenantSvc = tenantcapsvc.New(nil, nil)
	)
	svc := New(permissionFilter, bizCtxSvc, configSvc, i18nSvc, organizationState, orgCapSvc, tenantSvc).(*serviceImpl)
	refreshRoleTestScope(svc)
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
	refreshRoleTestScope(svc)
}

// refreshRoleTestScope rebuilds the stateless data-scope helper from the
// current explicit fake dependencies.
func refreshRoleTestScope(svc *serviceImpl) {
	svc.SetDataScopeService(datascope.New(svc.bizCtxSvc, svc, svc.orgCapSvc))
}
