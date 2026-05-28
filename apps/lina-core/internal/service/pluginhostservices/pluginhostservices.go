// Package pluginhostservices builds the host-published service directory used
// by source plugins while keeping HTTP startup limited to runtime orchestration.
package pluginhostservices

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/notify"
	"lina-core/internal/service/pluginsettings"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/capability"
	capabilityconfig "lina-core/pkg/plugin/capability/config"
	"lina-core/pkg/plugin/capability/contract"
	capabilityhostconfig "lina-core/pkg/plugin/capability/hostconfig"
	capabilitymanifest "lina-core/pkg/plugin/capability/manifest"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	capabilitypluginlifecycle "lina-core/pkg/plugin/capability/pluginlifecycle"
	capabilitypluginstate "lina-core/pkg/plugin/capability/pluginstate"
	capabilitytenantcap "lina-core/pkg/plugin/capability/tenantcap"
	capabilitytenantfilter "lina-core/pkg/plugin/capability/tenantfilter"
	"lina-core/pkg/plugin/pluginhost"
)

// PluginLifecycleRunner defines the host lifecycle operations published to
// source-plugin services.
type PluginLifecycleRunner interface {
	// Embedded methods must preserve host lifecycle, cache invalidation, i18n,
	// and plugin bridge authorization semantics defined by the stable contract.
	contract.PluginLifecycleRunner
}

// directory implements the source-plugin host service directory.
type directory struct {
	apiDoc         contract.APIDocService // apiDoc exposes localized API-documentation route text.
	auth           contract.AuthService   // auth exposes tenant token operations.
	bizCtx         contract.BizCtxService // bizCtx exposes read-only request business context.
	cache          kvcache.Service        // cache owns the shared runtime-selected KV backend.
	config         contract.ConfigServiceFactory
	hostConfig     contract.HostConfigService
	i18n           contract.I18nService // i18n exposes runtime translation lookups.
	manifest       contract.ManifestServiceFactory
	notify         contract.NotifyService // notify exposes host notification delivery.
	org            capabilityorgcap.Service
	pluginLife     contract.PluginLifecycleService
	pluginSettings contract.PluginSettingsService // pluginSettings exposes namespaced key-value settings.
	pluginState    contract.PluginStateService    // pluginState exposes plugin enablement lookups.
	route          contract.RouteService          // route exposes dynamic route metadata lookups.
	session        contract.SessionService        // session exposes online-session operations.
	tenant         capabilitytenantcap.Service
	tenantFilter   contract.TenantFilterService // tenantFilter exposes plugin table tenant filtering.
}

// scopedDirectory wraps a base directory with one plugin-bound cache adapter.
type scopedDirectory struct {
	base     *directory
	pluginID string
}

// Ensure directory satisfies the source-plugin capability contract.
var _ pluginhost.Services = (*directory)(nil)

// Ensure directory satisfies the unified capability services contract.
var _ capability.Services = (*directory)(nil)

// Ensure directory satisfies the plugin-scoped capability factory contract.
var _ capability.ScopedServicesFactory = (*directory)(nil)

// Ensure scopedDirectory satisfies the source-plugin capability contract.
var _ pluginhost.Services = (*scopedDirectory)(nil)

// Ensure scopedDirectory satisfies the unified capability services contract.
var _ capability.Services = (*scopedDirectory)(nil)

// New creates source-plugin host service adapters from runtime-owned services.
func New(
	apiDocSvc apidoc.Service,
	authSvc auth.Service,
	authTokenIssuer auth.TenantTokenIssuer,
	bizCtxSvc bizctx.Service,
	configSvc hostconfig.Service,
	scopeSvc datascope.Service,
	i18nSvc i18nsvc.Service,
	pluginStateSvc contract.PluginStateService,
	pluginLifecycleRunner PluginLifecycleRunner,
	sessionStore session.Store,
	orgSvc capabilityorgcap.Service,
	tenantSvc capabilitytenantcap.RuntimeService,
	notifySvc notify.Service,
	kvCacheSvc kvcache.Service,
) (capability.Services, error) {
	if kvCacheSvc == nil {
		return nil, gerror.New("create plugin host services failed: cache service is nil")
	}
	bizCtxAdapter := newBizCtxAdapter(bizCtxSvc)
	tenantFilterSvc, err := capabilitytenantfilter.New(bizCtxAdapter, tenantSvc)
	if err != nil {
		return nil, gerror.Wrap(err, "create plugin tenant filter service failed")
	}
	return &directory{
		apiDoc:         newAPIDocAdapter(apiDocSvc),
		auth:           newAuthAdapter(authTokenIssuer, authSvc, configSvc),
		bizCtx:         bizCtxAdapter,
		cache:          kvCacheSvc,
		config:         capabilityconfig.NewFactory("", ""),
		hostConfig:     capabilityhostconfig.New(configSvc),
		i18n:           newI18nAdapter(i18nSvc),
		manifest:       capabilitymanifest.NewFactory(""),
		notify:         newNotifyAdapter(notifySvc),
		org:            orgSvc,
		pluginLife:     capabilitypluginlifecycle.New(pluginLifecycleRunner),
		pluginSettings: newPluginSettingsAdapter(pluginsettings.New()),
		pluginState:    capabilitypluginstate.New(pluginStateSvc),
		route:          newRouteAdapter(),
		session:        newSessionAdapter(authSvc, scopeSvc, sessionStore, tenantSvc),
		tenant:         tenantSvc,
		tenantFilter:   tenantFilterSvc,
	}, nil
}
