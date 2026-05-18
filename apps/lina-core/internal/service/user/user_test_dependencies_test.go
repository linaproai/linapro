// This file keeps user service tests aligned with explicit dependency
// injection when individual fakes are replaced after construction.

package user

import (
	"context"

	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cluster"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/orgcap"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	tenantcapsvc "lina-core/internal/service/tenantcap"
)

// newUserTestService constructs user service tests through explicit dependencies.
func newUserTestService(tenantEnablementReaders ...tenantcapsvc.PluginEnablementReader) Service {
	bizCtxSvc := bizctx.New()
	configSvc := hostconfig.New()
	clusterSvc := cluster.New(configSvc.GetCluster(context.Background()))
	cacheCoordSvc := cachecoord.Default(clusterSvc)
	i18nSvc := i18nsvc.New(bizCtxSvc, configSvc, cacheCoordSvc)
	sessionStore := session.NewDBStore()
	pluginSvc, err := pluginsvc.New(clusterSvc, configSvc, bizCtxSvc, cacheCoordSvc, i18nSvc, sessionStore, nil)
	if err != nil {
		panic(err)
	}
	orgCapSvc := orgcap.New(pluginSvc)
	tenantSvc := tenantcapsvc.New(nil, nil)
	if len(tenantEnablementReaders) > 0 {
		tenantSvc = tenantcapsvc.New(tenantEnablementReaders[0], nil)
	}
	roleSvc := role.New(pluginSvc, bizCtxSvc, configSvc, i18nSvc, nil, orgCapSvc, tenantSvc)
	scopeSvc := datascope.New(bizCtxSvc, roleSvc, orgCapSvc)
	roleSvc.SetDataScopeService(scopeSvc)
	authSvc := auth.New(configSvc, pluginSvc, orgCapSvc, roleSvc, tenantSvc, sessionStore, kvcache.New())
	return New(authSvc, bizCtxSvc, i18nSvc, orgCapSvc, roleSvc, scopeSvc, tenantSvc).(*serviceImpl)
}

// setUserTestBizCtx replaces the business context dependency and refreshes
// the derived data-scope service used by user-management tests.
func setUserTestBizCtx(svc *serviceImpl, bizCtxSvc bizctx.Service) {
	svc.bizCtxSvc = bizCtxSvc
	refreshUserTestScope(svc)
}

// setUserTestOrgCap replaces the organization capability dependency and
// refreshes the derived data-scope service used by user-management tests.
func setUserTestOrgCap(svc *serviceImpl, orgCapSvc orgcap.Service) {
	svc.orgCapSvc = orgCapSvc
	refreshUserTestScope(svc)
}

// refreshUserTestScope rebuilds the stateless data-scope helper from the
// current explicit fake dependencies.
func refreshUserTestScope(svc *serviceImpl) {
	svc.scopeSvc = datascope.New(svc.bizCtxSvc, svc.roleSvc, svc.orgCapSvc)
}
