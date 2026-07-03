// Package runtime provides the dynamic plugin execution environment: WASM artifact
// parsing, upload handling, background reconciliation, per-node state projection,
// and route dispatch for enabled dynamic plugins.
package runtime

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord/revisionctrl"
	"lina-core/internal/service/cluster"
	configsvc "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/migration"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/capability"
	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	"lina-core/pkg/plugin/pluginhost"
)

// IntegrationService abstracts the integration service methods runtime needs
// while avoiding a package import cycle with plugin/internal/integration.
type IntegrationService interface {
	// SyncPluginMenusAndPermissions synchronizes all plugin menus and dynamic route
	// permission entries for the given manifest.
	SyncPluginMenusAndPermissions(ctx context.Context, manifest *catalog.Manifest) error
	// SyncPluginMenus synchronizes only the declared manifest menus, skipping
	// route-permission entries. Used during rollback to restore a previous menu state.
	SyncPluginMenus(ctx context.Context, manifest *catalog.Manifest) error
	// DeletePluginMenusByManifest removes all plugin-owned menu rows for the given manifest.
	DeletePluginMenusByManifest(ctx context.Context, manifest *catalog.Manifest) error
	// SyncPluginResourceReferences keeps sys_plugin_resource_ref aligned with the
	// current governance resource index derived from the given manifest.
	SyncPluginResourceReferences(ctx context.Context, manifest *catalog.Manifest) error
	// DispatchPluginHookEvent fires a lifecycle hook event to all registered listeners.
	DispatchPluginHookEvent(
		ctx context.Context,
		event pluginhost.ExtensionPoint,
		values map[string]interface{},
	) error
	// FilterPermissionMenus returns only the menus that pass plugin-level enablement checks.
	FilterPermissionMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu
	// CanExposeBusinessEntries reports whether a plugin can expose business entries in the current tenant context.
	CanExposeBusinessEntries(ctx context.Context, pluginID string) bool
}

// CacheChangeNotifier publishes successful dynamic runtime cache mutations to
// the root plugin facade.
type CacheChangeNotifier interface {
	// MarkRuntimeCacheChanged records one cache-affecting runtime change.
	MarkRuntimeCacheChanged(ctx context.Context, reason string) error
	// PublishPluginChange records one plugin-scoped runtime change.
	PublishPluginChange(ctx context.Context, pluginID string, pluginType string, reason string) error
}

// DependencyValidator validates candidate dynamic plugin releases before the
// reconciler switches effective release state or runs lifecycle side effects.
type DependencyValidator interface {
	// ValidateDynamicPluginCandidate verifies candidate dependencies and
	// reverse-dependency version safety for one dynamic lifecycle action.
	ValidateDynamicPluginCandidate(ctx context.Context, manifest *catalog.Manifest) error
}

// DynamicReconcileOptions carries explicit one-shot reconcile decisions from
// lifecycle into runtime without using context values for business control flow.
type DynamicReconcileOptions struct {
	// DesiredManifest optionally carries a caller-validated desired manifest for
	// the addressed plugin. It is used by targeted management requests that have
	// already applied one-shot request decisions such as install mode.
	DesiredManifest *catalog.Manifest
	// InstallMockData requests loading optional mock-data SQL during install.
	InstallMockData bool
}

// Service defines the dynamic-plugin runtime contract used by other plugin
// packages. Cross-package dependency seams stay as dedicated constructor
// interfaces above; runtime's own methods are declared directly here instead
// of being grouped through self-composition interfaces.
type Service interface {
	// ParseRuntimeWasmArtifact reads one WASM artifact file and extracts all embedded custom sections.
	// It implements catalog.ArtifactParser and returns validation/parse errors
	// without mutating registry, cache, i18n, data-scope, or bridge state.
	ParseRuntimeWasmArtifact(filePath string) (*catalog.ArtifactSpec, error)
	// ParseRuntimeWasmArtifactContent parses one WASM artifact from an in-memory byte slice.
	// It implements catalog.ArtifactParser and returns validation/parse errors
	// without mutating registry, cache, i18n, data-scope, or bridge state.
	ParseRuntimeWasmArtifactContent(filePath string, content []byte) (*catalog.ArtifactSpec, error)
	// ValidateRuntimeArtifact loads and validates the WASM artifact for a dynamic plugin source directory.
	// It implements catalog.ArtifactParser and must reject artifacts whose
	// manifest, bridge metadata, routes, i18n source text, or frontend assets do
	// not match the source manifest.
	ValidateRuntimeArtifact(manifest *catalog.Manifest, rootDir string) error

	// ListRuntimeStates returns public plugin runtime states for shell slot rendering.
	ListRuntimeStates(ctx context.Context) (*RuntimeStateListOutput, error)

	// ExecuteDynamicRoute is the exported form of executeDynamicRoute for cross-package access.
	ExecuteDynamicRoute(
		ctx context.Context,
		manifest *catalog.Manifest,
		request *bridgecontract.BridgeRequestEnvelopeV1,
	) (*bridgecontract.BridgeResponseEnvelopeV1, error)
	// RegisterDynamicRouteDispatcher binds the fixed-prefix dispatcher into one host
	// router group so dynamic routes reuse the standard RouterGroup registration flow.
	RegisterDynamicRouteDispatcher(group *ghttp.RouterGroup)
	// PrepareDynamicRouteMiddleware resolves the active dynamic route contract and
	// caches host-owned runtime state on the request before later middlewares run.
	PrepareDynamicRouteMiddleware(r *ghttp.Request)
	// AuthenticateDynamicRouteMiddleware applies host-owned login and permission
	// governance for the matched dynamic route before bridge execution starts.
	AuthenticateDynamicRouteMiddleware(r *ghttp.Request)
	// DispatchDynamicRoute dispatches one fixed-prefix request into the active release
	// of one dynamic plugin. Matching always happens against the archived active manifest
	// so staged uploads cannot affect live traffic before reconcile.
	DispatchDynamicRoute(
		ctx context.Context,
		in *DynamicRouteDispatchInput,
	) (*bridgecontract.BridgeResponseEnvelopeV1, error)

	// ReconcileDynamicPluginRequest submits a desired-state transition to the reconciler loop.
	ReconcileDynamicPluginRequest(ctx context.Context, pluginID string, desiredState string, options DynamicReconcileOptions) error
	// EnsureRuntimeArtifactAvailable verifies the WASM artifact is present for the given lifecycle action label.
	EnsureRuntimeArtifactAvailable(manifest *catalog.Manifest, actionLabel string) error
	// ShouldRefreshInstalledDynamicRelease reports whether the installed dynamic release is stale.
	ShouldRefreshInstalledDynamicRelease(
		ctx context.Context,
		registry interface{},
		manifest *catalog.Manifest,
	) bool

	// BuildPluginItem returns a PluginItem projection for one manifest + registry pair.
	// Used by the plugin facade SyncAndList coordination method.
	BuildPluginItem(ctx context.Context, manifest *catalog.Manifest, registry *store.PluginRecord) *PluginItem
	// BuildPluginSummaryItem returns the lightweight management-list projection
	// for one manifest + registry pair.
	BuildPluginSummaryItem(ctx context.Context, manifest *catalog.Manifest, registry *store.PluginRecord) *PluginItem
	// BuildPluginItemReadOnly returns one detail projection without mutating
	// governance state when a dynamic artifact is missing from storage.
	BuildPluginItemReadOnly(ctx context.Context, manifest *catalog.Manifest, registry *store.PluginRecord) *PluginItem
	// BuildRuntimeItems returns PluginItems for dynamic plugins present in the registry
	// but absent from the given manifest map. Used by the plugin facade SyncAndList.
	BuildRuntimeItems(ctx context.Context, covered map[string]struct{}) ([]*PluginItem, error)
	// BuildRuntimeSummaryItemsReadOnly returns lightweight dynamic PluginItems
	// without mutating missing artifacts back into governance tables.
	BuildRuntimeSummaryItemsReadOnly(ctx context.Context, covered map[string]struct{}) ([]*PluginItem, error)
	// BuildRuntimeItemsReadOnly returns dynamic PluginItems without reconciling
	// missing artifacts back into governance tables.
	BuildRuntimeItemsReadOnly(ctx context.Context, covered map[string]struct{}) ([]*PluginItem, error)
	// CheckIsInstalled reports whether a plugin is installed after reconciling artifact state.
	// Used by the plugin facade UpdateStatus guard.
	CheckIsInstalled(ctx context.Context, pluginID string) (bool, error)
	// HasArtifactStorageFile is the exported form of hasArtifactStorageFile for cross-package access.
	HasArtifactStorageFile(ctx context.Context, pluginID string) (bool, string, error)

	// SyncPluginNodeState implements catalog.NodeStateSyncer.
	// It updates the current node projection of one plugin lifecycle state.
	SyncPluginNodeState(
		ctx context.Context,
		pluginID string,
		version string,
		installed int,
		enabled int,
		message string,
	) error
	// GetPluginNodeState implements catalog.NodeStateSyncer.
	// It returns the latest node projection row for one plugin/node pair.
	GetPluginNodeState(ctx context.Context, pluginID string, nodeID string) (*entity.SysPluginNodeState, error)
	// CurrentNodeID implements catalog.NodeStateSyncer.
	CurrentNodeID() string
	// SyncPluginReleaseRuntimeState implements catalog.ReleaseStateSyncer.
	// It updates the active release row to reflect current registry state.
	SyncPluginReleaseRuntimeState(ctx context.Context, registry *store.PluginRecord) error

	// StartRuntimeReconciler starts the background loop that keeps dynamic-plugin
	// desired state, active release, and current-node projection converged.
	StartRuntimeReconciler(ctx context.Context)
	// ReconcileRuntimePlugins runs one convergence pass. It is safe to call from
	// both the background loop and synchronous management flows.
	ReconcileRuntimePlugins(ctx context.Context) error
	// RefreshInstalledRuntimePluginReleases repairs same-version installed
	// dynamic releases whose archived artifact or snapshot is stale while
	// avoiding install or state-toggle side effects for unrelated registry rows.
	RefreshInstalledRuntimePluginReleases(ctx context.Context) error

	// UpgradeDynamicPluginRequest runs the version-switching upgrade path for one
	// installed dynamic plugin. Discovery and background reconciliation must not
	// call this method implicitly.
	UpgradeDynamicPluginRequest(ctx context.Context, pluginID string) error

	// RunDynamicLifecycleCallback executes one dynamic lifecycle handler when declared.
	RunDynamicLifecycleCallback(
		ctx context.Context,
		manifest *catalog.Manifest,
		input DynamicLifecycleInput,
	) (*DynamicLifecycleDecision, error)
	// RunDynamicLifecyclePrecondition executes one dynamic Before* handler when declared.
	RunDynamicLifecyclePrecondition(
		ctx context.Context,
		manifest *catalog.Manifest,
		input DynamicLifecycleInput,
	) (*DynamicLifecycleDecision, error)

	// DiscoverJobContracts runs the dynamic plugin Jobs declaration entry point.
	DiscoverJobContracts(ctx context.Context, manifest *catalog.Manifest) ([]*bridgecontract.JobContract, error)
	// ExecuteDeclaredJob runs one declared dynamic-plugin job through the active runtime.
	ExecuteDeclaredJob(ctx context.Context, manifest *catalog.Manifest, contract *bridgecontract.JobContract) error

	// Uninstall executes uninstall lifecycle for an installed dynamic plugin.
	Uninstall(ctx context.Context, pluginID string) error
	// UninstallWithOptions executes uninstall lifecycle for an installed dynamic
	// plugin using one explicit cleanup policy snapshot.
	UninstallWithOptions(ctx context.Context, pluginID string, purgeStorageData bool) error
	// ForceUninstallMissingArtifact clears host governance for an installed
	// dynamic plugin whose staging and active release artifacts are unavailable.
	ForceUninstallMissingArtifact(ctx context.Context, registry *store.PluginRecord) error

	// LoadActiveDynamicPluginManifest implements catalog.DynamicManifestLoader.
	// It returns the currently active dynamic-plugin manifest reloaded from the stable
	// release archive so live traffic sees the stable version during staged upgrades.
	LoadActiveDynamicPluginManifest(ctx context.Context, registry *store.PluginRecord) (*catalog.Manifest, error)

	// UploadDynamicPackage validates one runtime wasm package and writes it into the
	// configured plugin.dynamic.storagePath directory.
	UploadDynamicPackage(ctx context.Context, in *DynamicUploadInput) (out *DynamicUploadOutput, err error)
	// StoreUploadedPackage is the exported form of storeUploadedPackage for cross-package access.
	StoreUploadedPackage(ctx context.Context, filename string, content []byte, overwriteSupport bool) (*DynamicUploadOutput, error)
}

// Ensure serviceImpl satisfies the composed runtime contract used by other plugin packages.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// catalogSvc provides manifest discovery and manifest asset access.
	catalogSvc catalog.Service
	// storeSvc provides plugin governance registry and release projections.
	storeSvc store.Service
	// migrationSvc executes install/uninstall SQL migration support.
	migrationSvc migration.Service
	// frontendSvc manages in-memory frontend bundles.
	frontendSvc frontend.Service
	// topology provides cluster topology information.
	topology cluster.Service
	// integrationSvc handles runtime integration side effects.
	integrationSvc IntegrationService
	// configSvc provides runtime configuration for dynamic routes and package uploads.
	configSvc configsvc.Service
	// userCtx injects the authenticated user identity into the request context.
	userCtx bizctx.Service
	// sessionStore validates online-session hot state for dynamic route requests.
	sessionStore session.Store
	// roleAccess projects role-owned token permissions for dynamic route requests.
	roleAccess role.Service
	// cacheChangeNotifier publishes runtime cache changes after successful convergence.
	cacheChangeNotifier CacheChangeNotifier
	// reconcilerLockSvc serializes primary lifecycle side effects per dynamic plugin.
	reconcilerLockSvc locker.Service
	// dependencyValidator checks candidate release dependency constraints before
	// dynamic lifecycle side effects.
	dependencyValidator DependencyValidator
	// storageCleanupServices resolves plugin-scoped storage cleanup capability views for dynamic uninstall resource purging.
	storageCleanupServices capability.Services
	// wasmRuntime owns dynamic-plugin WASM execution and host-call dependencies.
	wasmRuntime wasm.Runtime
	// reconcilerRevisionObserved records the reconciler revision consumed by this runtime service.
	reconcilerRevisionObserved *revisionctrl.ObservedRevision
	// reconcilerRevisionCtrl coordinates cluster-wide dynamic-plugin reconciler wake-up.
	reconcilerRevisionCtrl *revisionctrl.Controller
	// reconcilerOnce starts the background reconciler loop once per runtime service instance.
	reconcilerOnce sync.Once
	// reconcileMu serializes convergence passes for this runtime service instance.
	reconcileMu sync.Mutex
	// reconcilerSafetyMu protects the last full-sweep timestamp.
	reconcilerSafetyMu sync.Mutex
	// lastReconcilerSweepAt records the last successful background full-scan pass.
	lastReconcilerSweepAt time.Time
	// i18nSvc localizes plugin metadata and invalidates dynamic-plugin bundles.
	i18nSvc i18nsvc.Service
}

// New creates a new runtime Service with the given sub-service dependencies.
func New(
	catalogSvc catalog.Service,
	storeSvc store.Service,
	migrationSvc migration.Service,
	frontendSvc frontend.Service,
	i18nSvc i18nsvc.Service,
	reconcilerLockSvc locker.Service,
	topology cluster.Service,
	integrationSvc IntegrationService,
	configSvc configsvc.Service,
	userCtx bizctx.Service,
	sessionStore session.Store,
	roleAccess role.Service,
	cacheChangeNotifier CacheChangeNotifier,
	dependencyValidator DependencyValidator,
	storageCleanupServices capability.Services,
	wasmRuntime wasm.Runtime,
) Service {
	service := &serviceImpl{
		catalogSvc:                 catalogSvc,
		storeSvc:                   storeSvc,
		migrationSvc:               migrationSvc,
		frontendSvc:                frontendSvc,
		topology:                   topology,
		integrationSvc:             integrationSvc,
		configSvc:                  configSvc,
		userCtx:                    userCtx,
		sessionStore:               sessionStore,
		roleAccess:                 roleAccess,
		cacheChangeNotifier:        cacheChangeNotifier,
		reconcilerLockSvc:          reconcilerLockSvc,
		dependencyValidator:        dependencyValidator,
		storageCleanupServices:     storageCleanupServices,
		wasmRuntime:                wasmRuntime,
		reconcilerRevisionObserved: revisionctrl.NewObservedRevision(),
		i18nSvc:                    i18nSvc,
	}
	service.configureReconcilerRevisionController()
	return service
}
