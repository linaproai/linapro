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
	"lina-core/internal/service/hostlock"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/notify"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	capabilityhostconfig "lina-core/pkg/plugin/capability/hostconfigcap"
	capabilitymanifest "lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	capabilityconfig "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// newUserTestService constructs user service tests through explicit dependencies.
func newUserTestService(tenantManagersAndRuntimes ...any) Service {
	bizCtxSvc := bizctx.New()
	configSvc := hostconfig.New()
	clusterSvc := cluster.New(configSvc.GetCluster(context.Background()))
	cacheCoordSvc := cachecoord.Default(clusterSvc)
	i18nSvc := i18nsvc.New(bizCtxSvc, configSvc, cacheCoordSvc)
	sessionStore := session.NewDBStore()
	lockerSvc := locker.New()
	pluginRuntime := pluginsvc.NewRuntimeDelegate()
	orgCapSvc := orgspi.New(nil, pluginRuntime)
	tenantSvc := tenantspi.New(nil, nil, nil)
	if len(tenantManagersAndRuntimes) > 0 {
		var (
			manager *tenantspi.Manager
			runtime tenantspi.ProviderRuntime
		)
		if value, ok := tenantManagersAndRuntimes[0].(*tenantspi.Manager); ok {
			manager = value
			if len(tenantManagersAndRuntimes) > 1 {
				runtime, _ = tenantManagersAndRuntimes[1].(tenantspi.ProviderRuntime)
			}
		} else {
			runtime, _ = tenantManagersAndRuntimes[0].(tenantspi.ProviderRuntime)
		}
		tenantSvc = tenantspi.New(manager, runtime, nil)
	}
	roleSvc := role.New(pluginRuntime, bizCtxSvc, configSvc, i18nSvc, orgCapSvc, tenantSvc)
	scopeSvc := datascope.New(bizCtxSvc, roleSvc, orgCapSvc)
	roleSvc.SetDataScopeService(scopeSvc)
	kvCacheSvc := kvcache.New()
	authSvc := auth.New(configSvc, pluginRuntime, orgCapSvc, roleSvc, tenantSvc, sessionStore, kvCacheSvc)
	notifySvc := notify.New(tenantSvc)
	apiDocSvc := apidoc.New(configSvc, bizCtxSvc, i18nSvc, pluginRuntime)
	hostConfigReader, ok := configSvc.(capabilityhostconfig.RawConfigReader)
	if !ok {
		panic("test config service does not support raw host config reads")
	}
	hostConfigSvc := capabilityhostconfig.New(hostConfigReader)
	lockSvc, err := hostlock.New(lockerSvc)
	if err != nil {
		panic(err)
	}
	capabilities, err := pluginsvc.NewHostServices(
		apiDocSvc,
		authSvc,
		nil,
		bizCtxSvc,
		hostConfigSvc,
		scopeSvc,
		i18nSvc,
		pluginRuntime,
		pluginRuntime,
		sessionStore,
		aitext.New(nil, pluginRuntime),
		orgCapSvc,
		tenantSvc,
		notifySvc,
		kvCacheSvc,
		lockSvc,
		pluginsvc.NewStorageProviderRuntime(pluginRuntime),
		pluginsvc.NewLocalStorageProvider(configSvc.GetPluginDynamicStoragePath(context.Background())),
	)
	if err != nil {
		panic(err)
	}
	pluginSvc, err := pluginsvc.New(
		clusterSvc,
		configSvc,
		bizCtxSvc,
		cacheCoordSvc,
		i18nSvc,
		sessionStore,
		roleSvc,
		lockerSvc,
		nil,
		capabilities,
		orgCapSvc,
		tenantSvc,
		tenantSvc,
		tenantSvc,
		capabilityconfig.NewConfigFactory("", ""),
		hostConfigSvc,
		capabilitymanifest.NewFactory(""),
	)
	if err != nil {
		panic(err)
	}
	if err = pluginRuntime.BindService(pluginSvc); err != nil {
		panic(err)
	}
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
	if orgScope, ok := orgCapSvc.(orgspi.ScopeService); ok {
		svc.orgScope = orgScope
	} else {
		svc.orgScope = nil
	}
	if orgAssignment, ok := orgCapSvc.(orgspi.AssignmentService); ok {
		svc.orgAssignment = orgAssignment
	} else {
		svc.orgAssignment = nil
	}
	refreshUserTestScope(svc)
}

// refreshUserTestScope rebuilds the stateless data-scope helper from the
// current explicit fake dependencies.
func refreshUserTestScope(svc *serviceImpl) {
	var orgScope orgspi.ScopeService
	if svc.orgScope != nil {
		orgScope = svc.orgScope
	} else if scope, ok := svc.orgCapSvc.(orgspi.ScopeService); ok {
		orgScope = scope
	}
	svc.scopeSvc = datascope.New(svc.bizCtxSvc, svc.roleSvc, orgScope)
}
