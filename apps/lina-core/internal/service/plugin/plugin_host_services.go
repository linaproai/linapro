// This file exposes startup facades for source-plugin host service adapters
// and dynamic-plugin Wasm host service dispatchers while keeping concrete
// adapter implementations inside plugin internals. It also contains host-side
// helper projections used by host-service authorization review.

package plugin

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/hostlock"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/notify"
	"lina-core/internal/service/plugin/internal/capabilityhost"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/internal/service/session"
	"lina-core/pkg/dialect"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability"
	capabilityaitext "lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
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

// NewLocalStorageProvider creates the host built-in plugin storage provider.
func NewLocalStorageProvider(rootDir string) storagecap.Provider {
	return capabilityhost.NewLocalStorageProvider(rootDir)
}

// NewStorageProviderRuntime creates provider selection runtime state from plugin
// enablement dependencies.
func NewStorageProviderRuntime(pluginStateSvc plugincap.StateService) storagecap.ProviderRuntime {
	return &storageProviderRuntime{pluginStateSvc: pluginStateSvc}
}

// storageProviderRuntime adapts plugin state to storagecap runtime.
type storageProviderRuntime struct {
	pluginStateSvc plugincap.StateService
}

// ProviderPluginAvailable reports whether a provider plugin may serve calls.
func (r *storageProviderRuntime) ProviderPluginAvailable(ctx context.Context, pluginID string) bool {
	if r == nil || r.pluginStateSvc == nil {
		return false
	}
	return r.pluginStateSvc.IsProviderEnabled(ctx, pluginID)
}

// NewHostServices creates source-plugin host service adapters from startup-owned
// shared services and delegates concrete adapter construction to the plugin
// capabilityhost subcomponent. Callers must pass the same shared service instances
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
	tenantSvc tenantspi.RuntimeService,
	notifySvc HostNotifyPublisher,
	kvCacheSvc kvcache.Service,
	lockSvc hostlock.Service,
	storageRuntime storagecap.ProviderRuntime,
	localStorageProvider storagecap.Provider,
) (capability.Services, error) {
	return capabilityhost.New(
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
		lockSvc,
		storageRuntime,
		localStorageProvider,
	)
}

// newWasmHostServiceRuntime creates the dynamic-plugin host-service dispatcher
// runtime from startup-owned shared services.
func newWasmHostServiceRuntime(
	hostServices capability.Services,
	configFactory plugincap.ConfigServiceFactory,
	hostConfigSvc hostconfigcap.Service,
	manifestFactory manifestcap.ServiceFactory,
) (wasm.Runtime, error) {
	runtime, err := wasm.NewRuntime(
		hostServices,
		configFactory,
		hostConfigSvc,
		manifestFactory,
	)
	if err != nil {
		return nil, gerror.Wrap(err, "create wasm host service runtime failed")
	}
	return runtime, nil
}

// dataTableMetadataSchema is the host schema used by PostgreSQL metadata
// lookups.
const dataTableMetadataSchema = "public"

// ResolveDataTableComments resolves host-side table comments for the given
// data-table names. It degrades to an empty map when metadata lookup is
// unavailable so plugin list APIs are not blocked by optional schema comments.
func (s *serviceImpl) ResolveDataTableComments(ctx context.Context, tables []string) map[string]string {
	normalizedTables := normalizeDataTableNames(tables)
	if len(normalizedTables) == 0 {
		return map[string]string{}
	}

	db := g.DB()
	dbDialect, err := dialect.FromDatabase(db)
	if err != nil {
		logger.Warningf(ctx, "resolve plugin data table comments skipped: %v", err)
		return map[string]string{}
	}

	metas, err := dbDialect.QueryTableMetadata(ctx, db, dataTableMetadataSchema, normalizedTables)
	if err != nil {
		logger.Warningf(ctx, "resolve plugin data table comments failed schema=%s tables=%v err=%v", dataTableMetadataSchema, normalizedTables, err)
		return map[string]string{}
	}
	return dataTableCommentsFromMetadata(metas)
}

// dataTableCommentsFromMetadata maps dialect table metadata to the comment map
// consumed by plugin governance projections.
func dataTableCommentsFromMetadata(metas []dialect.TableMeta) map[string]string {
	comments := make(map[string]string, len(metas))
	for _, meta := range metas {
		tableName := strings.TrimSpace(meta.TableName)
		tableComment := strings.TrimSpace(meta.TableComment)
		if tableName == "" || tableComment == "" {
			continue
		}
		comments[tableName] = tableComment
	}
	return comments
}

// normalizeDataTableNames trims blanks and de-duplicates table names before
// they are used in metadata lookup queries.
func normalizeDataTableNames(tables []string) []string {
	if len(tables) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(tables))
	normalized := make([]string, 0, len(tables))
	for _, table := range tables {
		name := strings.TrimSpace(table)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		normalized = append(normalized, name)
	}
	return normalized
}
