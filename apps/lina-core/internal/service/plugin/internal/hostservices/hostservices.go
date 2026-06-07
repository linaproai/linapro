// Package hostservices builds the host-published service directory used
// by source plugins while keeping HTTP startup limited to runtime orchestration.
package hostservices

import (
	"context"
	"io/fs"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/notify"
	hostauthzcap "lina-core/internal/service/plugin/internal/hostservices/internal/authzcap"
	hostruntimeconfigcap "lina-core/internal/service/plugin/internal/hostservices/internal/configcap"
	hostdictcap "lina-core/internal/service/plugin/internal/hostservices/internal/dictcap"
	hostfilecap "lina-core/internal/service/plugin/internal/hostservices/internal/filecap"
	hostinfracap "lina-core/internal/service/plugin/internal/hostservices/internal/infracap"
	hostjobcap "lina-core/internal/service/plugin/internal/hostservices/internal/jobcap"
	hostnotifycap "lina-core/internal/service/plugin/internal/hostservices/internal/notifycap"
	hostplugincap "lina-core/internal/service/plugin/internal/hostservices/internal/plugincap"
	hostsessioncap "lina-core/internal/service/plugin/internal/hostservices/internal/sessioncap"
	hostusercap "lina-core/internal/service/plugin/internal/hostservices/internal/usercap"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/capability"
	capabilityai "lina-core/pkg/plugin/capability/aicap"
	capabilityaitext "lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/capability/bizctxcap"
	capabilitydictcap "lina-core/pkg/plugin/capability/dictcap"
	capabilityfilecap "lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/i18ncap"
	capabilityinfracap "lina-core/pkg/plugin/capability/infracap"
	capabilityjobcap "lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	capabilitymanifest "lina-core/pkg/plugin/capability/manifestcap"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/tenantcap"
	capabilitytenantcap "lina-core/pkg/plugin/capability/tenantcap"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
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
	Current(ctx context.Context) bizctxcap.CurrentContext
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
	// Send validates the notify channel and creates unified notify message and delivery records.
	Send(ctx context.Context, in notify.SendInput) (*notify.SendOutput, error)
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
	plugincap.LifecycleRunner
}

// directory implements the source-plugin host service directory.
type directory struct {
	apiDoc        apidoccap.Service // apiDoc exposes localized API-documentation route text.
	auth          authcap.Service   // auth exposes authentication and authorization sub capabilities.
	ai            capabilityai.Service
	users         capabilityusercap.Service
	bizCtx        bizctxcap.Service // bizCtx exposes read-only request business context.
	cache         kvcache.Service   // cache owns the shared runtime-selected KV backend.
	dict          capabilitydictcap.Service
	files         capabilityfilecap.Service
	hostConfig    hostconfigcap.Service
	i18n          i18ncap.Service // i18n exposes runtime translation lookups.
	infra         capabilityinfracap.Service
	jobs          capabilityjobcap.Service
	manifest      manifestcap.ServiceFactory
	notifications capabilitynotifycap.Service
	org           capabilityorgcap.Service
	admin         capability.AdminServices
	plugins       hostplugincap.Service
	route         routecap.Service // route exposes dynamic route metadata lookups.
	sessions      capabilitysessioncap.Service
	tenant        capabilitytenantcap.Service
	tenantFilter  tenantcap.PluginTableFilterService // tenantFilter exposes plugin table tenant filtering.
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
	hostConfigSvc hostconfigcap.Service,
	scopeSvc datascope.Service,
	i18nSvc RuntimeI18nService,
	pluginStateSvc plugincap.StateService,
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
	tenantFilterSvc, err := capabilitytenantcap.NewPluginTableFilter(bizCtxAdapter, tenantSvc)
	if err != nil {
		return nil, gerror.Wrap(err, "create plugin tenant filter service failed")
	}
	var (
		i18nAdapter         = newI18nAdapter(i18nSvc)
		userDomain          = hostusercap.New(tenantFilterSvc)
		tokenDomain         = newAuthAdapter(authTokenIssuer)
		authzDomain         = hostauthzcap.New()
		dictDomain          = hostdictcap.New(tenantFilterSvc, i18nAdapter)
		runtimeConfigDomain = hostruntimeconfigcap.New(tenantFilterSvc)
		fileDomain          = hostfilecap.New(tenantFilterSvc)
		sessionDomain       = hostsessioncap.New(authSvc, scopeSvc, sessionStore, tenantSvc)
		notificationDomain  = hostnotifycap.New(notifySvc)
		jobDomain           = hostjobcap.New(tenantFilterSvc)
		infraDomain         = hostinfracap.New()
		pluginConfigFactory = plugincap.NewConfigFactory("", "")
		pluginLifecycle     = plugincap.NewLifecycle(pluginLifecycleRunner)
		pluginState         = plugincap.NewState(pluginStateSvc)
		pluginDomain        = hostplugincap.New(pluginConfigFactory, pluginState, pluginLifecycle)
	)
	return &directory{
		apiDoc:        newAPIDocAdapter(apiDocSvc),
		auth:          authcap.New(tokenDomain, authzDomain),
		ai:            capabilityai.New(aiTextSvc),
		users:         userDomain,
		bizCtx:        bizCtxAdapter,
		cache:         kvCacheSvc,
		dict:          dictDomain,
		files:         fileDomain,
		hostConfig:    hostConfigSvc,
		i18n:          i18nAdapter,
		infra:         infraDomain,
		jobs:          jobDomain,
		manifest:      capabilitymanifest.NewFactory("", sourcePluginEmbeddedFiles),
		notifications: notificationDomain,
		org:           orgSvc,
		admin: &adminDirectory{
			users:    userDomain,
			auth:     authcap.NewAdmin(authzDomain),
			dict:     dictDomain,
			config:   runtimeConfigDomain,
			files:    fileDomain,
			sessions: sessionDomain,
			notify:   notificationDomain,
			plugins:  pluginDomain,
			jobs:     jobDomain,
			infra:    infraDomain,
		},
		plugins:      pluginDomain,
		route:        newRouteAdapter(),
		sessions:     sessionDomain,
		tenant:       tenantSvc,
		tenantFilter: tenantFilterSvc,
	}, nil
}

// sourcePluginEmbeddedFiles resolves source-plugin embedded assets without
// making manifestcap depend on pluginhost.
func sourcePluginEmbeddedFiles(pluginID string) fs.FS {
	sourcePlugin, ok := pluginhost.GetSourcePlugin(strings.TrimSpace(pluginID))
	if !ok || sourcePlugin == nil {
		return nil
	}
	return sourcePlugin.GetEmbeddedFiles()
}
