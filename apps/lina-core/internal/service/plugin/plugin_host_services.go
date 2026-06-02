// This file exposes the startup facade for source-plugin host service
// adapters while keeping concrete adapter implementations inside plugin internals.

package plugin

import (
	"context"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/notify"
	"lina-core/internal/service/plugin/internal/hostservices"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/capability"
	capabilityaitext "lina-core/pkg/plugin/capability/ai/aitext"
	"lina-core/pkg/plugin/capability/contract"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	capabilitytenantcap "lina-core/pkg/plugin/capability/tenantcap"
)

// HostAPIDocResolver defines the apidoc route-text slice required by
// source-plugin host service adapters.
type HostAPIDocResolver interface {
	// ResolveRouteText resolves one route's localized module tag and operation summary.
	ResolveRouteText(ctx context.Context, input apidoc.RouteTextInput) apidoc.RouteTextOutput
	// ResolveRouteTexts resolves multiple route texts with a single apidoc catalog load.
	ResolveRouteTexts(ctx context.Context, inputs []apidoc.RouteTextInput) []apidoc.RouteTextOutput
	// FindRouteTitleOperationKeys finds operation key bases whose localized module tag matches keyword.
	FindRouteTitleOperationKeys(ctx context.Context, keyword string) []string
}

// HostAuthSessionRevoker defines the session revocation slice required by
// source-plugin session adapters.
type HostAuthSessionRevoker interface {
	// RevokeSession writes a shared revoke marker and removes one online session by token ID.
	RevokeSession(ctx context.Context, tokenID string) error
}

// HostTenantTokenIssuer defines the tenant-token handoff slice required by
// source-plugin auth adapters.
type HostTenantTokenIssuer interface {
	// IssueTenantToken consumes a pre-login token and issues a tenant-bound token pair.
	IssueTenantToken(ctx context.Context, in auth.TenantTokenIssueInput) (*auth.TenantTokenOutput, error)
	// ReissueTenantTokenFromBearer parses the current bearer token and issues a new tenant-bound token pair.
	ReissueTenantTokenFromBearer(ctx context.Context, tokenString string, tenantID int) (*auth.TenantTokenOutput, error)
	// IssueImpersonationToken signs and registers one host-owned impersonation token.
	IssueImpersonationToken(ctx context.Context, in auth.ImpersonationTokenIssueInput) (*auth.ImpersonationTokenOutput, error)
	// RevokeImpersonationToken validates and revokes one host impersonation token.
	RevokeImpersonationToken(ctx context.Context, tokenString string, tenantID int) error
}

// HostBizContextProvider defines the read-only request context projection
// required by source-plugin adapters.
type HostBizContextProvider interface {
	// Current returns the plugin-visible read-only projection of the current business context.
	Current(ctx context.Context) contract.CurrentContext
}

// HostRuntimeI18nService defines the runtime translation slice required by
// source-plugin host service adapters.
type HostRuntimeI18nService interface {
	// GetLocale returns the effective request locale.
	GetLocale(ctx context.Context) string
	// Translate resolves one runtime message key in the current request locale.
	Translate(ctx context.Context, key string, fallback string) string
	// ExportMessages exports flat runtime messages for one locale.
	ExportMessages(ctx context.Context, locale string) i18nsvc.MessageExportOutput
}

// HostNotifyPublisher defines the notification slice required by source-plugin adapters.
type HostNotifyPublisher interface {
	// SendNoticePublication sends one published notice through the built-in inbox channel.
	SendNoticePublication(ctx context.Context, in notify.NoticePublishInput) (*notify.SendOutput, error)
	// DeleteBySource removes notify records for the given business source identifiers.
	DeleteBySource(ctx context.Context, sourceType notify.SourceType, sourceIDs []string) error
}

// NewHostServices creates source-plugin host service adapters from startup-owned
// shared services and delegates concrete adapter construction to the plugin
// hostservices subcomponent. Callers must pass the same shared service instances
// used by HTTP startup, WASM host services, middleware, and plugin lifecycle
// orchestration; this facade does not create replacement runtime services.
func NewHostServices(
	apiDocSvc HostAPIDocResolver,
	authSvc HostAuthSessionRevoker,
	authTokenIssuer HostTenantTokenIssuer,
	bizCtxSvc HostBizContextProvider,
	hostConfigSvc contract.HostConfigService,
	scopeSvc datascope.Service,
	i18nSvc HostRuntimeI18nService,
	pluginStateSvc contract.PluginStateService,
	pluginLifecycleRunner contract.PluginLifecycleRunner,
	sessionStore session.Store,
	aiTextSvc capabilityaitext.Service,
	orgSvc capabilityorgcap.Service,
	tenantSvc capabilitytenantcap.RuntimeService,
	notifySvc HostNotifyPublisher,
	kvCacheSvc kvcache.Service,
) (capability.Services, error) {
	return hostservices.New(
		apiDocSvc,
		authSvc,
		authTokenIssuer,
		bizCtxSvc,
		hostConfigSvc,
		scopeSvc,
		i18nSvc,
		pluginStateSvc,
		pluginLifecycleRunner,
		sessionStore,
		aiTextSvc,
		orgSvc,
		tenantSvc,
		notifySvc,
		kvCacheSvc,
	)
}
