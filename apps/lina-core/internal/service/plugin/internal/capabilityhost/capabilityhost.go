// Package capabilityhost builds the host-owned capability service directory
// used by source plugins and dynamic-plugin domain host calls while keeping
// HTTP startup limited to runtime orchestration.
package capabilityhost

import (
	"context"
	"io/fs"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/dict"
	filesvc "lina-core/internal/service/file"
	"lina-core/internal/service/hostlock"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/jobmeta"
	jobcapadapter "lina-core/internal/service/jobmgmt/capabilityadapter"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/notify"
	"lina-core/internal/service/plugin/internal/capabilityowner"
	hostconfigadapter "lina-core/internal/service/plugin/internal/hostconfig"
	"lina-core/internal/service/plugin/internal/manifestresource"
	"lina-core/internal/service/plugin/internal/pluginconfig"
	"lina-core/internal/service/role"
	usersvc "lina-core/internal/service/user"
	usercapadapter "lina-core/internal/service/user/capabilityadapter"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/capability/bizctxcap"
	capabilitydictcap "lina-core/pkg/plugin/capability/dictcap"
	capabilityfilecap "lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/i18ncap"
	capabilityjobcap "lina-core/pkg/plugin/capability/jobcap"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/storagecap"
	capabilitytenantcap "lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/plugin/pluginhost"
)

// directory implements the source-plugin host service directory.
type directory struct {
	apiDoc          apidoccap.Service // apiDoc exposes localized API-documentation route text.
	auth            authcap.Service   // auth exposes authentication and authorization sub capabilities.
	users           capabilityusercap.Service
	bizCtx          bizctxcap.Service // bizCtx exposes read-only request business context.
	cache           kvcache.Service   // cache owns the shared runtime-selected KV backend.
	dict            capabilitydictcap.Service
	files           capabilityfilecap.Service
	hostConfig      hostconfigcap.Service
	i18n            i18ncap.Service // i18n exposes runtime translation lookups.
	jobs            capabilityjobcap.Service
	lock            hostlock.Service // lock owns the shared runtime-selected lock backend.
	manifest        manifestresource.Factory
	notifications   capabilitynotifycap.Service
	org             capabilityorgcap.Service
	plugins         capabilityowner.CapabilityService
	route           routecap.Service // route exposes dynamic route metadata lookups.
	sessions        capabilitysessioncap.Service
	storageRuntime  storagecap.ProviderRuntime // storageRuntime selects the active storage provider.
	storageProvider storagecap.Provider        // storageProvider is the built-in local provider.
	tenant          capabilitytenantcap.Service
}

// pluginStateLookup defines the root plugin enablement methods required by
// capability adapters without making this subcomponent depend on the root
// plugin service package.
type pluginStateLookup interface {
	// IsEnabled reports whether one plugin is enabled in the current scope.
	IsEnabled(ctx context.Context, pluginID string) bool
	// IsProviderEnabled reports whether one plugin may serve provider calls.
	IsProviderEnabled(ctx context.Context, pluginID string) bool
	// IsEnabledAuthoritative reports persisted plugin enablement bypassing local snapshots.
	IsEnabledAuthoritative(ctx context.Context, pluginID string) bool
}

// scopedDirectory wraps a base directory with one plugin-bound cache adapter.
type scopedDirectory struct {
	base     *directory
	pluginID string
}

// Ensure directory satisfies the unified capability services contract.
var _ capability.Services = (*directory)(nil)

// Ensure directory satisfies the plugin-scoped capability factory contract.
var _ capabilityowner.ScopedServicesFactory = (*directory)(nil)

// Ensure scopedDirectory satisfies the unified capability services contract.
var _ capability.Services = (*scopedDirectory)(nil)

// New creates source-plugin host service adapters from runtime-owned services.
func New(
	apiDocSvc apidoc.Service,
	authSvc auth.Service,
	bizCtxSvc bizctxcap.Service,
	roleAccessSvc role.Service,
	hostConfigSvc hostconfigcap.Service,
	scopeSvc datascope.Service,
	cacheCoordSvc cachecoord.Service,
	i18nSvc i18nsvc.Service,
	pluginStateSvc pluginStateLookup,
	pluginLifecycleSvc plugincap.LifecycleService,
	userSvc usersvc.Service,
	fileSvc filesvc.Service,
	jobSvc jobmeta.Owner,
	orgSvc capabilityorgcap.Service,
	tenantSvc tenantspi.Service,
	notifySvc notify.Service,
	kvCacheSvc kvcache.Service,
	lockSvc hostlock.Service,
	pluginConfigFactory pluginconfig.Factory,
	storageRuntime storagecap.ProviderRuntime,
	localStorageProvider storagecap.Provider,
) (capability.Services, error) {
	if hostConfigSvc == nil {
		return nil, gerror.New("create plugin host services failed: host config service is nil")
	}
	if pluginConfigFactory == nil {
		return nil, gerror.New("create plugin host services failed: plugin config factory is nil")
	}
	if kvCacheSvc == nil {
		return nil, gerror.New("create plugin host services failed: cache service is nil")
	}
	if lockSvc == nil {
		return nil, gerror.New("create plugin host services failed: lock service is nil")
	}
	if localStorageProvider == nil {
		return nil, gerror.New("create plugin host services failed: local storage provider is nil")
	}
	if jobSvc == nil {
		return nil, gerror.New("create plugin host services failed: job service is nil")
	}
	if cacheCoordSvc == nil {
		return nil, gerror.New("create plugin host services failed: cachecoord service is nil")
	}
	if tenantSvc == nil {
		return nil, gerror.New("create plugin host services failed: tenant service is nil")
	}
	tenantFilterSvc := tenantSvc.Filter()
	if tenantFilterSvc == nil {
		return nil, gerror.New("create plugin host services failed: tenant filter service is nil")
	}
	bizCtxAdapter := newBizCtxAdapter(bizCtxSvc)
	var (
		i18nAdapter        = newI18nAdapter(i18nSvc)
		userDomain         = usercapadapter.NewCapabilityAdapter(userSvc, tenantSvc, scopeSvc, bizCtxAdapter)
		tokenDomain        = newAuthAdapter(authSvc)
		authzDomain        = role.NewCapabilityAdapter(roleAccessSvc, bizCtxAdapter, cacheCoordSvc)
		dictDomain         = dict.NewCapabilityAdapter(tenantFilterSvc, i18nAdapter, cacheCoordSvc)
		sysConfigDomain    = hostconfigadapter.NewSysConfigCapabilityAdapter(tenantFilterSvc, cacheCoordSvc)
		hostConfigDomain   = hostconfigadapter.NewCapabilityService(hostConfigSvc, sysConfigDomain)
		fileDomain         = filesvc.NewCapabilityAdapter(fileSvc, tenantFilterSvc)
		sessionDomain      = newSessionCapabilityAdapter(authSvc, bizCtxAdapter, userDomain, scopeSvc, tenantSvc)
		notificationDomain = notify.NewCapabilityAdapter(notifySvc, bizCtxAdapter)
		jobDomain          = jobcapadapter.NewCapabilityAdapter(jobSvc, tenantFilterSvc, scopeSvc)
		pluginState        = pluginStateSvc
		pluginLifecycle    = plugincap.NewLifecycle(pluginLifecycleSvc)
		pluginDomain       = capabilityowner.NewCapabilityAdapter(pluginConfigFactory, pluginState, pluginLifecycle, tenantSvc, cacheCoordSvc)
		tenantDomain       = newTenantDomainService(tenantSvc, pluginDomain)
	)
	return &directory{
		apiDoc:          newAPIDocAdapter(apiDocSvc),
		auth:            authcap.New(tokenDomain, authzDomain),
		users:           userDomain,
		bizCtx:          bizCtxAdapter,
		cache:           kvCacheSvc,
		dict:            dictDomain,
		files:           fileDomain,
		hostConfig:      hostConfigDomain,
		i18n:            i18nAdapter,
		jobs:            jobDomain,
		lock:            lockSvc,
		manifest:        manifestresource.NewFactory("", sourcePluginEmbeddedFiles),
		notifications:   notificationDomain,
		org:             orgSvc,
		plugins:         pluginDomain,
		route:           newRouteAdapter(),
		sessions:        sessionDomain,
		storageRuntime:  storageRuntime,
		storageProvider: localStorageProvider,
		tenant:          tenantDomain,
	}, nil
}

// sourcePluginEmbeddedFiles resolves source-plugin embedded assets without
// making manifestresource depend on pluginhost.
func sourcePluginEmbeddedFiles(pluginID string) fs.FS {
	sourcePlugin, ok := pluginhost.GetSourcePlugin(strings.TrimSpace(pluginID))
	if !ok || sourcePlugin == nil {
		return nil
	}
	return sourcePlugin.GetEmbeddedFiles()
}
