// Package pluginhostservices builds the host-published service directory used
// by source plugins while keeping HTTP startup limited to runtime orchestration.
package pluginhostservices

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/notify"
	"lina-core/internal/service/session"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/pkg/pluginhost"
	pluginserviceconfig "lina-core/pkg/pluginservice/config"
	"lina-core/pkg/pluginservice/contract"
	pluginservicepluginlifecycle "lina-core/pkg/pluginservice/pluginlifecycle"
	pluginservicepluginstate "lina-core/pkg/pluginservice/pluginstate"
	pluginservicetenantfilter "lina-core/pkg/pluginservice/tenantfilter"
)

// PluginStateReader defines the plugin-state read operation required by the
// source-plugin host service directory.
type PluginStateReader interface {
	// IsEnabled reports whether pluginID is installed, enabled, and allowed to
	// expose source-plugin business entries for the current request scope.
	// Implementations must honor tenant/data-scope plugin visibility and return
	// false rather than leaking errors through bridge-facing state checks.
	IsEnabled(ctx context.Context, pluginID string) bool
}

// PluginLifecycleRunner defines the host lifecycle operations published to
// source-plugin services.
type PluginLifecycleRunner interface {
	// Embedded methods must preserve host lifecycle, cache invalidation, i18n,
	// and plugin bridge authorization semantics defined by the stable contract.
	contract.PluginLifecycleRunner
}

// directory implements the pluginhost.HostServices directory.
type directory struct {
	apiDoc       contract.APIDocService // apiDoc exposes localized API-documentation route text.
	auth         contract.AuthService   // auth exposes tenant token operations.
	bizCtx       contract.BizCtxService // bizCtx exposes read-only request business context.
	config       contract.ConfigService // config exposes read-only host configuration.
	i18n         contract.I18nService   // i18n exposes runtime translation lookups.
	notify       contract.NotifyService // notify exposes host notification delivery.
	pluginLife   contract.PluginLifecycleService
	pluginState  contract.PluginStateService  // pluginState exposes plugin enablement lookups.
	route        contract.RouteService        // route exposes dynamic route metadata lookups.
	session      contract.SessionService      // session exposes online-session operations.
	tenantFilter contract.TenantFilterService // tenantFilter exposes plugin table tenant filtering.
}

// Ensure directory satisfies the source-plugin host service contract.
var _ pluginhost.HostServices = (*directory)(nil)

// New creates source-plugin host service adapters from runtime-owned services.
func New(
	apiDocSvc apidoc.Service,
	authSvc auth.Service,
	authTokenIssuer auth.TenantTokenIssuer,
	bizCtxSvc bizctx.Service,
	scopeSvc datascope.Service,
	i18nSvc i18nsvc.Service,
	pluginStateReader PluginStateReader,
	pluginLifecycleRunner PluginLifecycleRunner,
	sessionStore session.Store,
	tenantSvc tenantcapsvc.Service,
	notifySvc notify.Service,
) (pluginhost.HostServices, error) {
	bizCtxAdapter := newBizCtxAdapter(bizCtxSvc)
	tenantFilterSvc, err := pluginservicetenantfilter.New(bizCtxAdapter, tenantSvc)
	if err != nil {
		return nil, gerror.Wrap(err, "create plugin tenant filter service failed")
	}
	return &directory{
		apiDoc:       newAPIDocAdapter(apiDocSvc),
		auth:         newAuthAdapter(authTokenIssuer),
		bizCtx:       bizCtxAdapter,
		config:       pluginserviceconfig.New(),
		i18n:         newI18nAdapter(i18nSvc),
		notify:       newNotifyAdapter(notifySvc),
		pluginLife:   pluginservicepluginlifecycle.New(pluginLifecycleRunner),
		pluginState:  pluginservicepluginstate.New(pluginStateReader),
		route:        newRouteAdapter(),
		session:      newSessionAdapter(authSvc, scopeSvc, sessionStore, tenantSvc),
		tenantFilter: tenantFilterSvc,
	}, nil
}
