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
	"lina-core/internal/service/cachecoord"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	filesvc "lina-core/internal/service/file"
	"lina-core/internal/service/hostlock"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/jobmeta"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/notify"
	"lina-core/internal/service/plugin/internal/capabilityhost"
	hostconfigadapter "lina-core/internal/service/plugin/internal/hostconfig"
	"lina-core/internal/service/plugin/internal/manifestresource"
	"lina-core/internal/service/plugin/internal/pluginconfig"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/internal/service/role"
	storagesvc "lina-core/internal/service/storage"
	usersvc "lina-core/internal/service/user"
	"lina-core/pkg/dialect"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capregistry"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/pluginhost"
)

// NewHostConfigService creates the plugin-visible HostConfig capability from
// the startup-owned host configuration reader.
func NewHostConfigService(reader configsvc.Service) hostconfigcap.Service {
	return hostconfigadapter.NewStaticCapabilityAdapter(reader)
}

// PluginConfigFactory creates plugin-scoped configuration service views for
// source-plugin capability directories and dynamic-plugin WASM host services.
type PluginConfigFactory = pluginconfig.Factory

// pluginStateLookup narrows the root plugin service to the enablement lookups
// needed by host-service adapters and provider runtime guards.
type pluginStateLookup interface {
	// IsEnabled reports whether one plugin is enabled in the current scope.
	IsEnabled(ctx context.Context, pluginID string) bool
	// IsProviderEnabled reports whether one plugin may serve provider calls.
	IsProviderEnabled(ctx context.Context, pluginID string) bool
	// IsEnabledAuthoritative reports persisted plugin enablement bypassing local snapshots.
	IsEnabledAuthoritative(ctx context.Context, pluginID string) bool
}

// NewPluginConfigFactory creates a plugin configuration factory with optional
// root overrides.
func NewPluginConfigFactory(productionRoot string, developmentRoot string) PluginConfigFactory {
	return pluginconfig.NewFactory(productionRoot, developmentRoot)
}

// NewPluginConfigFactoryWithHostStaticConfig creates a plugin configuration
// factory that checks host static plugin.<plugin-id> sections before file and
// artifact sources.
func NewPluginConfigFactoryWithHostStaticConfig(
	productionRoot string,
	developmentRoot string,
	hostStatic configsvc.Service,
) PluginConfigFactory {
	return pluginconfig.NewFactoryWithHostStaticConfig(productionRoot, developmentRoot, hostStatic)
}

// NewLocalStorageProvider creates the host built-in plugin storage provider.
func NewLocalStorageProvider(storageSvc storagesvc.Service) storagecap.Provider {
	return capabilityhost.NewLocalStorageProvider(storageSvc)
}

// NewStorageProviderRuntime creates provider selection runtime state from plugin
// enablement dependencies.
func NewStorageProviderRuntime(pluginStateSvc pluginStateLookup) storagecap.ProviderRuntime {
	return &storageProviderRuntime{pluginStateSvc: pluginStateSvc}
}

// storageProviderRuntime adapts plugin state to storagecap runtime.
type storageProviderRuntime struct {
	pluginStateSvc pluginStateLookup
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
	pluginConfigFactory PluginConfigFactory,
	storageRuntime storagecap.ProviderRuntime,
	localStorageProvider storagecap.Provider,
) (capability.Services, error) {
	return capabilityhost.New(
		apiDocSvc,
		authSvc,
		bizCtxSvc,
		roleAccessSvc,
		hostConfigSvc,
		scopeSvc,
		cacheCoordSvc,
		i18nSvc,
		pluginStateSvc,
		pluginLifecycleSvc,
		userSvc,
		fileSvc,
		jobSvc,
		orgSvc,
		tenantSvc,
		notifySvc,
		kvCacheSvc,
		lockSvc,
		pluginConfigFactory,
		storageRuntime,
		localStorageProvider,
	)
}

// newWasmHostServiceRuntime creates the dynamic-plugin host-service dispatcher
// runtime from startup-owned shared services.
func newWasmHostServiceRuntime(
	hostServices capability.Services,
	ownerCapabilities *capregistry.Registry,
	configFactory PluginConfigFactory,
	hostConfigSvc hostconfigcap.Service,
	manifestFactory manifestresource.Factory,
) (wasm.Runtime, error) {
	runtime, err := wasm.NewRuntime(
		hostServices,
		ownerCapabilities,
		configFactory,
		hostConfigSvc,
		manifestFactory,
	)
	if err != nil {
		return nil, gerror.Wrap(err, "create wasm host service runtime failed")
	}
	return runtime, nil
}

// buildSourceCapabilityRegistry builds the startup owner capability registry
// from source-plugin declarations registered before plugin service startup.
func buildSourceCapabilityRegistry() (*capregistry.Registry, error) {
	registry := capregistry.NewRegistry()
	for _, definition := range pluginhost.ListSourcePlugins() {
		if definition == nil {
			continue
		}
		for _, descriptor := range definition.GetCapabilityDescriptors() {
			if err := validateSourceCapabilityDescriptorOwner(definition.ID(), descriptor); err != nil {
				return nil, err
			}
			if err := registry.Register(descriptor); err != nil {
				return nil, err
			}
		}
	}
	return registry, nil
}

func validateSourceCapabilityDescriptorOwner(pluginID string, descriptor capregistry.Descriptor) error {
	declaringPluginID := strings.TrimSpace(pluginID)
	ownerPluginID := strings.TrimSpace(descriptor.OwnerPluginID)
	if declaringPluginID == "" || ownerPluginID == "" || ownerPluginID != declaringPluginID {
		return gerror.Newf(
			"capability descriptor owner must match declaring source plugin: plugin=%s owner=%s service=%s version=%s",
			declaringPluginID,
			ownerPluginID,
			strings.TrimSpace(descriptor.Service),
			strings.TrimSpace(descriptor.Version),
		)
	}
	return nil
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
