// This file exposes startup facades for source-plugin host service adapters
// and dynamic-plugin Wasm host service dispatchers while keeping concrete
// adapter implementations inside plugin internals.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/hostlock"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/notify"
	"lina-core/internal/service/plugin/internal/hostservices"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/capability"
	capabilityaitext "lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/plugincap"
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
	Current(ctx context.Context) bizctxcap.CurrentContext
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
	// Send validates the notify channel and creates unified notify message and delivery records.
	Send(ctx context.Context, in notify.SendInput) (*notify.SendOutput, error)
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
	hostConfigSvc hostconfigcap.Service,
	scopeSvc datascope.Service,
	i18nSvc HostRuntimeI18nService,
	pluginStateSvc plugincap.StateService,
	pluginLifecycleRunner plugincap.LifecycleRunner,
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

// ConfigureWasmHostServices wires dynamic-plugin host-service dispatchers to
// the same runtime-owned services used by the host HTTP process.
func ConfigureWasmHostServices(
	kvCacheSvc kvcache.Service,
	lockSvc hostlock.Service,
	notifySvc notify.Service,
	configSvc configsvc.PluginConfigReader,
	hostServices capability.Services,
	configFactory plugincap.ConfigServiceFactory,
	hostConfigSvc hostconfigcap.Service,
	manifestFactory manifestcap.ServiceFactory,
) error {
	if err := wasm.ConfigureCacheHostService(kvCacheSvc); err != nil {
		return gerror.Wrap(err, "configure wasm cache host service failed")
	}
	if err := wasm.ConfigureLockHostService(lockSvc); err != nil {
		return gerror.Wrap(err, "configure wasm lock host service failed")
	}
	if err := wasm.ConfigureNotifyHostService(notifySvc); err != nil {
		return gerror.Wrap(err, "configure wasm notify host service failed")
	}
	if err := wasm.ConfigureStorageHostService(configSvc); err != nil {
		return gerror.Wrap(err, "configure wasm storage host service failed")
	}
	if err := wasm.ConfigureAITextHostService(hostServices); err != nil {
		return gerror.Wrap(err, "configure wasm ai text host service failed")
	}
	if err := wasm.ConfigureOrgHostService(hostServices); err != nil {
		return gerror.Wrap(err, "configure wasm org host service failed")
	}
	if err := wasm.ConfigureTenantHostService(hostServices); err != nil {
		return gerror.Wrap(err, "configure wasm tenant host service failed")
	}
	if err := wasm.ConfigureConfigHostService(configFactory); err != nil {
		return gerror.Wrap(err, "configure wasm config host service failed")
	}
	if err := wasm.ConfigureHostConfigService(hostConfigSvc); err != nil {
		return gerror.Wrap(err, "configure wasm host config service failed")
	}
	if err := wasm.ConfigureManifestHostService(manifestFactory); err != nil {
		return gerror.Wrap(err, "configure wasm manifest host service failed")
	}
	return nil
}
