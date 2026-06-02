// Package hostservices builds the host-published service directory used
// by source plugins while keeping HTTP startup limited to runtime orchestration.
package hostservices

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/notify"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/capability"
	capabilityai "lina-core/pkg/plugin/capability/ai"
	capabilityaitext "lina-core/pkg/plugin/capability/ai/aitext"
	capabilityconfig "lina-core/pkg/plugin/capability/config"
	"lina-core/pkg/plugin/capability/contract"
	capabilitymanifest "lina-core/pkg/plugin/capability/manifest"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	capabilitypluginlifecycle "lina-core/pkg/plugin/capability/pluginlifecycle"
	capabilitypluginstate "lina-core/pkg/plugin/capability/pluginstate"
	capabilitytenantcap "lina-core/pkg/plugin/capability/tenantcap"
	capabilitytenantfilter "lina-core/pkg/plugin/capability/tenantfilter"
	"lina-core/pkg/plugin/pluginhost"
)

// APIDocResolver defines the route-text slice required by source-plugin apidoc adapters.
type APIDocResolver interface {
	// ResolveRouteText resolves one route's localized module tag and operation summary.
	ResolveRouteText(ctx context.Context, input apidoc.RouteTextInput) apidoc.RouteTextOutput
	// ResolveRouteTexts resolves multiple route texts with a single apidoc catalog load.
	ResolveRouteTexts(ctx context.Context, inputs []apidoc.RouteTextInput) []apidoc.RouteTextOutput
	// FindRouteTitleOperationKeys finds operation key bases whose localized module tag matches keyword.
	FindRouteTitleOperationKeys(ctx context.Context, keyword string) []string
}

// AuthSessionRevoker defines the session revocation slice required by
// source-plugin session adapters.
type AuthSessionRevoker interface {
	// RevokeSession writes a shared revoke marker and removes one online session by token ID.
	RevokeSession(ctx context.Context, tokenID string) error
}

// TenantTokenIssuer defines the tenant-token handoff slice required by
// source-plugin auth adapters.
type TenantTokenIssuer interface {
	// IssueTenantToken consumes a pre-login token and issues a tenant-bound token pair.
	IssueTenantToken(ctx context.Context, in auth.TenantTokenIssueInput) (*auth.TenantTokenOutput, error)
	// ReissueTenantTokenFromBearer parses the current bearer token and issues a new tenant-bound token pair.
	ReissueTenantTokenFromBearer(ctx context.Context, tokenString string, tenantID int) (*auth.TenantTokenOutput, error)
	// IssueImpersonationToken signs and registers one host-owned impersonation token.
	IssueImpersonationToken(ctx context.Context, in auth.ImpersonationTokenIssueInput) (*auth.ImpersonationTokenOutput, error)
	// RevokeImpersonationToken validates and revokes one host impersonation token.
	RevokeImpersonationToken(ctx context.Context, tokenString string, tenantID int) error
}

// BizContextProvider defines the read-only request context projection required
// by source-plugin adapters.
type BizContextProvider interface {
	// Current returns the plugin-visible read-only projection of the current business context.
	Current(ctx context.Context) contract.CurrentContext
}

// RuntimeI18nService defines the runtime translation slice required by
// source-plugin i18n adapters.
type RuntimeI18nService interface {
	// GetLocale returns the effective request locale.
	GetLocale(ctx context.Context) string
	// Translate resolves one runtime message key in the current request locale.
	Translate(ctx context.Context, key string, fallback string) string
	// ExportMessages exports flat runtime messages for one locale.
	ExportMessages(ctx context.Context, locale string) i18nsvc.MessageExportOutput
}

// NotifyPublisher defines the notification slice required by source-plugin adapters.
type NotifyPublisher interface {
	// SendNoticePublication sends one published notice through the built-in inbox channel.
	SendNoticePublication(ctx context.Context, in notify.NoticePublishInput) (*notify.SendOutput, error)
	// DeleteBySource removes notify records for the given business source identifiers.
	DeleteBySource(ctx context.Context, sourceType notify.SourceType, sourceIDs []string) error
}

// PluginLifecycleRunner defines the host lifecycle operations published to
// source-plugin services.
type PluginLifecycleRunner interface {
	// Embedded methods must preserve host lifecycle, cache invalidation, i18n,
	// and plugin bridge authorization semantics defined by the stable contract.
	contract.PluginLifecycleRunner
}

// directory implements the source-plugin host service directory.
type directory struct {
	apiDoc       contract.APIDocService // apiDoc exposes localized API-documentation route text.
	auth         contract.AuthService   // auth exposes tenant token operations.
	ai           capabilityai.Service
	bizCtx       contract.BizCtxService // bizCtx exposes read-only request business context.
	cache        kvcache.Service        // cache owns the shared runtime-selected KV backend.
	config       contract.ConfigServiceFactory
	hostConfig   contract.HostConfigService
	i18n         contract.I18nService // i18n exposes runtime translation lookups.
	manifest     contract.ManifestServiceFactory
	notify       contract.NotifyService // notify exposes host notification delivery.
	org          capabilityorgcap.Service
	pluginLife   contract.PluginLifecycleService
	pluginState  contract.PluginStateService // pluginState exposes plugin enablement lookups.
	route        contract.RouteService       // route exposes dynamic route metadata lookups.
	session      contract.SessionService     // session exposes online-session operations.
	tenant       capabilitytenantcap.Service
	tenantFilter contract.TenantFilterService // tenantFilter exposes plugin table tenant filtering.
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
	apiDocSvc APIDocResolver,
	authSvc AuthSessionRevoker,
	authTokenIssuer TenantTokenIssuer,
	bizCtxSvc BizContextProvider,
	hostConfigSvc contract.HostConfigService,
	scopeSvc datascope.Service,
	i18nSvc RuntimeI18nService,
	pluginStateSvc contract.PluginStateService,
	pluginLifecycleRunner PluginLifecycleRunner,
	sessionStore session.Store,
	aiTextSvc capabilityaitext.Service,
	orgSvc capabilityorgcap.Service,
	tenantSvc capabilitytenantcap.RuntimeService,
	notifySvc NotifyPublisher,
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
		apiDoc:       newAPIDocAdapter(apiDocSvc),
		auth:         newAuthAdapter(authTokenIssuer),
		ai:           capabilityai.New(aiTextSvc),
		bizCtx:       bizCtxAdapter,
		cache:        kvCacheSvc,
		config:       capabilityconfig.NewFactory("", ""),
		hostConfig:   hostConfigSvc,
		i18n:         newI18nAdapter(i18nSvc),
		manifest:     capabilitymanifest.NewFactory(""),
		notify:       newNotifyAdapter(notifySvc),
		org:          orgSvc,
		pluginLife:   capabilitypluginlifecycle.New(pluginLifecycleRunner),
		pluginState:  capabilitypluginstate.New(pluginStateSvc),
		route:        newRouteAdapter(),
		session:      newSessionAdapter(authSvc, scopeSvc, sessionStore, tenantSvc),
		tenant:       tenantSvc,
		tenantFilter: tenantFilterSvc,
	}, nil
}
