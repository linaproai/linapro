// This file keeps user service tests aligned with explicit dependency
// injection when individual fakes are replaced after construction.

package user

import (
	"context"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cluster"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/notify"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	capabilityhostconfig "lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/orgcap"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
)

// newUserTestService constructs user service tests through explicit dependencies.
func newUserTestService(tenantRuntimes ...tenantcapsvc.ProviderRuntime) Service {
	bizCtxSvc := bizctx.New()
	configSvc := hostconfig.New()
	clusterSvc := cluster.New(configSvc.GetCluster(context.Background()))
	cacheCoordSvc := cachecoord.Default(clusterSvc)
	i18nSvc := i18nsvc.New(bizCtxSvc, configSvc, cacheCoordSvc)
	sessionStore := session.NewDBStore()
	pluginSvc, err := pluginsvc.New(clusterSvc, configSvc, bizCtxSvc, cacheCoordSvc, i18nSvc, sessionStore, locker.New(), nil)
	if err != nil {
		panic(err)
	}
	orgCapSvc := orgcap.New(pluginSvc)
	tenantSvc := tenantcapsvc.New(nil, nil)
	if len(tenantRuntimes) > 0 {
		tenantSvc = tenantcapsvc.New(tenantRuntimes[0], nil)
	}
	roleSvc := role.New(pluginSvc, bizCtxSvc, configSvc, i18nSvc, nil, tenantSvc)
	scopeSvc := datascope.New(bizCtxSvc, roleSvc, orgCapSvc)
	roleSvc.SetDataScopeService(scopeSvc)
	kvCacheSvc := kvcache.New()
	authSvc := auth.New(configSvc, pluginSvc, orgCapSvc, roleSvc, tenantSvc, sessionStore, kvCacheSvc)
	notifySvc := notify.New(tenantSvc)
	apiDocSvc := apidoc.New(configSvc, bizCtxSvc, i18nSvc, pluginSvc)
	hostConfigReader, ok := configSvc.(capabilityhostconfig.RawConfigReader)
	if !ok {
		panic("test config service does not support raw host config reads")
	}
	hostConfigSvc := capabilityhostconfig.New(hostConfigReader)
	capabilities, err := pluginsvc.NewHostServices(
		apiDocSvc,
		authSvc,
		nil,
		bizCtxSvc,
		hostConfigSvc,
		scopeSvc,
		i18nSvc,
		pluginSvc,
		pluginSvc,
		sessionStore,
		aitext.New(pluginSvc),
		orgCapSvc,
		tenantSvc,
		notifySvc,
		kvCacheSvc,
	)
	if err != nil {
		panic(err)
	}
	pluginSvc.SetCapabilities(capabilities)
	return New(authSvc, bizCtxSvc, i18nSvc, orgCapSvc, orgCapSvc, orgCapSvc, roleSvc, scopeSvc, tenantSvc, tenantSvc, tenantSvc).(*serviceImpl)
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
	if orgScope, ok := orgCapSvc.(orgcap.ScopeService); ok {
		svc.orgScope = orgScope
	} else {
		svc.orgScope = nil
	}
	if orgAssignment, ok := orgCapSvc.(orgcap.AssignmentService); ok {
		svc.orgAssignment = orgAssignment
	} else {
		svc.orgAssignment = nil
	}
	refreshUserTestScope(svc)
}

// refreshUserTestScope rebuilds the stateless data-scope helper from the
// current explicit fake dependencies.
func refreshUserTestScope(svc *serviceImpl) {
	var orgScope orgcap.ScopeService
	if svc.orgScope != nil {
		orgScope = svc.orgScope
	} else if scope, ok := svc.orgCapSvc.(orgcap.ScopeService); ok {
		orgScope = scope
	}
	svc.scopeSvc = datascope.New(svc.bizCtxSvc, svc.roleSvc, orgScope)
}
