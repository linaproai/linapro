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
	"lina-core/internal/service/cachecoord/revisionctrl"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/migration"
	"lina-core/internal/service/plugin/internal/openapi"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/internal/service/plugin/internal/wasm"
	rolesvc "lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/capability"
	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	"lina-core/pkg/plugin/pluginhost"
)

// TopologyProvider abstracts cluster topology information needed by the reconciler.
type TopologyProvider interface {
	// IsClusterModeEnabled reports whether multi-node cluster mode is active.
	IsClusterModeEnabled() bool
	// IsPrimaryNode reports whether this host instance is the designated primary node.
	IsPrimaryNode() bool
	// CurrentNodeID returns the stable host-unique identifier for the current node.
	CurrentNodeID() string
}

// MenuManager abstracts menu-sync operations so the runtime package does not
// directly import the integration package (which depends on runtime).
type MenuManager interface {
	// SyncPluginMenusAndPermissions synchronizes all plugin menus and dynamic route
	// permission entries for the given manifest.
	SyncPluginMenusAndPermissions(ctx context.Context, manifest *catalog.Manifest) error
	// SyncPluginMenus synchronizes only the declared manifest menus, skipping
	// route-permission entries. Used during rollback to restore a previous menu state.
	SyncPluginMenus(ctx context.Context, manifest *catalog.Manifest) error
	// DeletePluginMenusByManifest removes all plugin-owned menu rows for the given manifest.
	DeletePluginMenusByManifest(ctx context.Context, manifest *catalog.Manifest) error
}

// ResourceReferenceManager abstracts resource-reference synchronization without
// importing the integration package into runtime.
type ResourceReferenceManager interface {
	// SyncPluginResourceReferences keeps sys_plugin_resource_ref aligned with the
	// current governance resource index derived from the given manifest.
	SyncPluginResourceReferences(ctx context.Context, manifest *catalog.Manifest) error
}

// HookDispatcher abstracts hook event dispatch so the runtime package does not
// depend on the integration package directly.
type HookDispatcher interface {
	// DispatchPluginHookEvent fires a lifecycle hook event to all registered listeners.
	DispatchPluginHookEvent(
		ctx context.Context,
		event pluginhost.ExtensionPoint,
		values map[string]interface{},
	) error
}

// JwtConfigProvider provides JWT configuration for dynamic route token validation.
type JwtConfigProvider interface {
	// GetJwtSecret returns the JWT signing secret used to validate bearer tokens.
	GetJwtSecret(ctx context.Context) string
	// GetSessionTimeout returns the runtime-effective online-session timeout.
	GetSessionTimeout(ctx context.Context) (time.Duration, error)
}

// UploadSizeProvider provides the runtime-effective upload size ceiling in MB.
type UploadSizeProvider interface {
	// GetUploadMaxSize returns the runtime-effective upload size ceiling in MB.
	GetUploadMaxSize(ctx context.Context) (int64, error)
}

// UserContextSetter injects authenticated user information into the request context.
type UserContextSetter interface {
	// SetUser populates the context with the resolved token, user identity, and client type fields.
	SetUser(ctx context.Context, tokenID string, userID int, username string, status int, clientType string)
	// SetTenant populates the resolved request tenant.
	SetTenant(ctx context.Context, tenantID int)
	// SetUserAccess populates cached access-snapshot fields for downstream plugin integrations.
	SetUserAccess(ctx context.Context, dataScope int, dataScopeUnsupported bool, unsupportedDataScope int)
}

// RoleAccessProjector exposes the role-owned token access projection used by
// dynamic route authentication before requests enter guest code.
type RoleAccessProjector interface {
	// BuildDynamicRouteAccessProjection returns a token-bound permission and
	// data-scope projection using role module cache and freshness semantics.
	BuildDynamicRouteAccessProjection(
		ctx context.Context,
		tokenID string,
		userID int,
		tenantID int,
	) (*rolesvc.DynamicRouteAccessProjection, error)
}

// userImpersonationSetter is optionally implemented by user context adapters
// that can preserve tenant impersonation metadata.
type userImpersonationSetter interface {
	// SetImpersonation populates platform impersonation metadata.
	SetImpersonation(ctx context.Context, actingUserID int, tenantID int, actingAsTenant bool, isImpersonation bool)
}

// PermissionMenuFilter filters button-type permission menus based on plugin enablement.
type PermissionMenuFilter interface {
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

// StorageCleanupServicesProvider returns the startup-owned capability directory
// used by dynamic uninstall storage cleanup.
type StorageCleanupServicesProvider interface {
	// StorageCleanupServices returns the current shared capability directory.
	StorageCleanupServices() capability.Services
}

// ArtifactService defines runtime WASM artifact parsing and validation operations.
type ArtifactService interface {
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
}

// RuntimeStateQueryService defines public runtime-state query operations.
type RuntimeStateQueryService interface {
	// ListRuntimeStates returns public plugin runtime states for shell slot rendering.
	ListRuntimeStates(ctx context.Context) (*RuntimeStateListOutput, error)
}

// DynamicRouteService defines dynamic plugin route execution and host dispatch operations.
type DynamicRouteService interface {
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
}

// LifecycleReconcileService defines runtime lifecycle convergence operations
// needed by install, enable, disable, and upgrade flows.
type LifecycleReconcileService interface {
	// ReconcileDynamicPluginRequest implements lifecycle.RuntimeOrchestrator.
	// It submits a desired-state transition to the reconciler loop.
	ReconcileDynamicPluginRequest(ctx context.Context, pluginID string, desiredState string, options DynamicReconcileOptions) error
	// EnsureRuntimeArtifactAvailable implements lifecycle.RuntimeOrchestrator.
	// It verifies the WASM artifact is present for the given lifecycle action label.
	EnsureRuntimeArtifactAvailable(manifest *catalog.Manifest, actionLabel string) error
	// ShouldRefreshInstalledDynamicRelease implements lifecycle.RuntimeOrchestrator.
	// It type-asserts registry to *store.PluginRecord then delegates to the private helper.
	ShouldRefreshInstalledDynamicRelease(
		ctx context.Context,
		registry interface{},
		manifest *catalog.Manifest,
	) bool
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

// RuntimeRegistryService defines registry-backed dynamic plugin projection and
// artifact-state query operations.
type RuntimeRegistryService interface {
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
}

// RuntimeProjectionService defines node and release state projection operations.
type RuntimeProjectionService interface {
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
}

// RuntimeReconcilerService defines background and on-demand runtime reconciliation operations.
type RuntimeReconcilerService interface {
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
}

// RuntimeUpgradeService defines explicit dynamic-plugin runtime upgrade operations.
type RuntimeUpgradeService interface {
	// UpgradeDynamicPluginRequest runs the version-switching upgrade path for one
	// installed dynamic plugin. Discovery and background reconciliation must not
	// call this method implicitly.
	UpgradeDynamicPluginRequest(ctx context.Context, pluginID string) error
}

// DynamicLifecyclePreconditionService defines dynamic lifecycle callback execution.
type DynamicLifecyclePreconditionService interface {
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
}

// DynamicJobService defines dynamic-plugin built-in Jobs discovery and
// execution operations.
type DynamicJobService interface {
	// DiscoverJobContracts runs the dynamic plugin Jobs declaration entry point.
	DiscoverJobContracts(ctx context.Context, manifest *catalog.Manifest) ([]*bridgecontract.JobContract, error)
	// ExecuteDeclaredJob runs one declared dynamic-plugin job through the active runtime.
	ExecuteDeclaredJob(ctx context.Context, manifest *catalog.Manifest, contract *bridgecontract.JobContract) error
}

// RuntimeLifecycleService defines dynamic plugin uninstall and emergency cleanup operations.
type RuntimeLifecycleService interface {
	// Uninstall executes uninstall lifecycle for an installed dynamic plugin.
	Uninstall(ctx context.Context, pluginID string) error
	// UninstallWithOptions executes uninstall lifecycle for an installed dynamic
	// plugin using one explicit cleanup policy snapshot.
	UninstallWithOptions(ctx context.Context, pluginID string, purgeStorageData bool) error
	// ForceUninstallMissingArtifact clears host governance for an installed
	// dynamic plugin whose staging and active release artifacts are unavailable.
	ForceUninstallMissingArtifact(ctx context.Context, registry *store.PluginRecord) error
}

// ActiveManifestService defines active dynamic release manifest loading operations.
type ActiveManifestService interface {
	// LoadActiveDynamicPluginManifest implements catalog.DynamicManifestLoader.
	// It returns the currently active dynamic-plugin manifest reloaded from the stable
	// release archive so live traffic sees the stable version during staged upgrades.
	LoadActiveDynamicPluginManifest(ctx context.Context, registry *store.PluginRecord) (*catalog.Manifest, error)
}

// DynamicPackageService defines runtime WASM package upload and storage operations.
type DynamicPackageService interface {
	// UploadDynamicPackage validates one runtime wasm package and writes it into the
	// configured plugin.dynamic.storagePath directory.
	UploadDynamicPackage(ctx context.Context, in *DynamicUploadInput) (out *DynamicUploadOutput, err error)
	// StoreUploadedPackage is the exported form of storeUploadedPackage for cross-package access.
	StoreUploadedPackage(ctx context.Context, filename string, content []byte, overwriteSupport bool) (*DynamicUploadOutput, error)
}

// Service defines the runtime service contract by composing runtime sub-capabilities.
type Service interface {
	ArtifactService
	RuntimeStateQueryService
	DynamicRouteService
	LifecycleReconcileService
	RuntimeRegistryService
	RuntimeProjectionService
	RuntimeReconcilerService
	RuntimeUpgradeService
	DynamicLifecyclePreconditionService
	DynamicJobService
	RuntimeLifecycleService
	ActiveManifestService
	DynamicPackageService
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
	// openapiSvc projects dynamic routes into the host OpenAPI document.
	openapiSvc openapi.Service
	// topology provides cluster topology information.
	topology TopologyProvider
	// menuMgr handles plugin menu and permission synchronization.
	menuMgr MenuManager
	// resourceRefMgr handles release-scoped governance resource references.
	resourceRefMgr ResourceReferenceManager
	// hookDispatcher fires lifecycle hook events.
	hookDispatcher HookDispatcher
	// jwtConfig provides the JWT signing secret for route token validation.
	jwtConfig JwtConfigProvider
	// uploadSize provides the runtime-effective upload size ceiling for package uploads.
	uploadSize UploadSizeProvider
	// userCtx injects the authenticated user identity into the request context.
	userCtx UserContextSetter
	// sessionStore validates online-session hot state for dynamic route requests.
	sessionStore session.Store
	// roleAccess projects role-owned token permissions for dynamic route requests.
	roleAccess RoleAccessProjector
	// menuFilter filters button-type permission menus by plugin enablement.
	menuFilter PermissionMenuFilter
	// cacheChangeNotifier publishes runtime cache changes after successful convergence.
	cacheChangeNotifier CacheChangeNotifier
	// reconcilerLockSvc serializes primary lifecycle side effects per dynamic plugin.
	reconcilerLockSvc locker.Service
	// dependencyValidator checks candidate release dependency constraints before
	// dynamic lifecycle side effects.
	dependencyValidator DependencyValidator
	// storageCleanupServices resolves plugin-scoped storage cleanup capability
	// views for dynamic uninstall resource purging.
	storageCleanupServices StorageCleanupServicesProvider
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
	i18nSvc runtimeI18nService
}

// runtimeI18nService defines the narrow i18n capability plugin runtime needs.
type runtimeI18nService interface {
	// Translate returns one runtime translation key with caller-provided fallback text.
	Translate(ctx context.Context, key string, fallback string) string
	// TranslateDynamicPluginSourceText resolves inactive dynamic-plugin source text from its release artifact.
	TranslateDynamicPluginSourceText(ctx context.Context, pluginID string, key string, sourceText string) string
	// InvalidateRuntimeBundleCache clears cached runtime bundles for one explicit scope.
	InvalidateRuntimeBundleCache(scope i18nsvc.InvalidateScope)
}

// New creates a new runtime Service with the given sub-service dependencies.
func New(
	catalogSvc catalog.Service,
	storeSvc store.Service,
	migrationSvc migration.Service,
	frontendSvc frontend.Service,
	openapiSvc openapi.Service,
	i18nSvc runtimeI18nService,
	reconcilerLockSvc locker.Service,
	topology TopologyProvider,
	menuMgr MenuManager,
	resourceRefMgr ResourceReferenceManager,
	hookDispatcher HookDispatcher,
	jwtConfig JwtConfigProvider,
	uploadSize UploadSizeProvider,
	userCtx UserContextSetter,
	sessionStore session.Store,
	roleAccess RoleAccessProjector,
	menuFilter PermissionMenuFilter,
	cacheChangeNotifier CacheChangeNotifier,
	dependencyValidator DependencyValidator,
	storageCleanupServices StorageCleanupServicesProvider,
	wasmRuntime wasm.Runtime,
) Service {
	service := &serviceImpl{
		catalogSvc:                 catalogSvc,
		storeSvc:                   storeSvc,
		migrationSvc:               migrationSvc,
		frontendSvc:                frontendSvc,
		openapiSvc:                 openapiSvc,
		topology:                   topology,
		menuMgr:                    menuMgr,
		resourceRefMgr:             resourceRefMgr,
		hookDispatcher:             hookDispatcher,
		jwtConfig:                  jwtConfig,
		uploadSize:                 uploadSize,
		userCtx:                    userCtx,
		sessionStore:               sessionStore,
		roleAccess:                 roleAccess,
		menuFilter:                 menuFilter,
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
