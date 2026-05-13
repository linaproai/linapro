// Package pluginhostservices builds the host-published service directory used
// by source plugins while keeping HTTP startup limited to runtime orchestration.
package pluginhostservices

import (
	"context"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/notify"
	"lina-core/internal/service/orgcap"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/pkg/pluginhost"
	pluginserviceconfig "lina-core/pkg/pluginservice/config"
	"lina-core/pkg/pluginservice/contract"
	pluginservicepluginstate "lina-core/pkg/pluginservice/pluginstate"
	pluginservicetenantfilter "lina-core/pkg/pluginservice/tenantfilter"
)

// PluginStateReader defines the plugin-state read operation required by the
// source-plugin host service directory.
type PluginStateReader interface {
	// IsEnabled reports whether the plugin is currently installed and enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
}

// directory implements the pluginhost.HostServices directory.
type directory struct {
	apiDoc       contract.APIDocService       // apiDoc exposes localized API-documentation route text.
	auth         contract.AuthService         // auth exposes tenant token operations.
	bizCtx       contract.BizCtxService       // bizCtx exposes read-only request business context.
	config       contract.ConfigService       // config exposes read-only host configuration.
	i18n         contract.I18nService         // i18n exposes runtime translation lookups.
	notify       contract.NotifyService       // notify exposes host notification delivery.
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
	i18nSvc i18nsvc.Service,
	orgCapSvc orgcap.Service,
	pluginStateReader PluginStateReader,
	roleSvc role.Service,
	sessionStore session.Store,
	tenantSvc tenantcapsvc.Service,
	notifySvc notify.Service,
) pluginhost.HostServices {
	bizCtxAdapter := newBizCtxAdapter(bizCtxSvc)
	scopeSvc := datascope.New(bizCtxSvc, roleSvc, orgCapSvc)
	return &directory{
		apiDoc:       newAPIDocAdapter(apiDocSvc),
		auth:         newAuthAdapter(authTokenIssuer),
		bizCtx:       bizCtxAdapter,
		config:       pluginserviceconfig.New(),
		i18n:         newI18nAdapter(i18nSvc),
		notify:       newNotifyAdapter(notifySvc),
		pluginState:  pluginservicepluginstate.New(pluginStateReader),
		route:        newRouteAdapter(),
		session:      newSessionAdapter(authSvc, scopeSvc, sessionStore, tenantSvc),
		tenantFilter: pluginservicetenantfilter.New(bizCtxAdapter, tenantSvc),
	}
}

// APIDoc returns the host API-documentation localization adapter.
func (s *directory) APIDoc() contract.APIDocService {
	if s == nil {
		return nil
	}
	return s.apiDoc
}

// Auth returns the host tenant-auth adapter.
func (s *directory) Auth() contract.AuthService {
	if s == nil {
		return nil
	}
	return s.auth
}

// BizCtx returns the host business-context adapter.
func (s *directory) BizCtx() contract.BizCtxService {
	if s == nil {
		return nil
	}
	return s.bizCtx
}

// Config returns the host static configuration adapter.
func (s *directory) Config() contract.ConfigService {
	if s == nil {
		return nil
	}
	return s.config
}

// I18n returns the host runtime translation adapter.
func (s *directory) I18n() contract.I18nService {
	if s == nil {
		return nil
	}
	return s.i18n
}

// Notify returns the host notification adapter.
func (s *directory) Notify() contract.NotifyService {
	if s == nil {
		return nil
	}
	return s.notify
}

// PluginState returns the host plugin enablement adapter.
func (s *directory) PluginState() contract.PluginStateService {
	if s == nil {
		return nil
	}
	return s.pluginState
}

// Route returns the host dynamic-route metadata adapter.
func (s *directory) Route() contract.RouteService {
	if s == nil {
		return nil
	}
	return s.route
}

// Session returns the host online-session adapter.
func (s *directory) Session() contract.SessionService {
	if s == nil {
		return nil
	}
	return s.session
}

// TenantFilter returns the host tenant-filter adapter.
func (s *directory) TenantFilter() contract.TenantFilterService {
	if s == nil {
		return nil
	}
	return s.tenantFilter
}
