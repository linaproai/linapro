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
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/openapi"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/internal/service/pluginruntimecache"
	"lina-core/internal/service/session"
	bridgecontract "lina-core/pkg/pluginbridge/contract"
	"lina-core/pkg/pluginhost"
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
	// SetUser populates the context with the resolved token and user identity fields.
	SetUser(ctx context.Context, tokenID string, userID int, username string, status int)
	// SetTenant populates the resolved request tenant.
	SetTenant(ctx context.Context, tenantID int)
	// SetUserAccess populates cached access-snapshot fields for downstream plugin integrations.
	SetUserAccess(ctx context.Context, dataScope int, dataScopeUnsupported bool, unsupportedDataScope int)
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
	// IsEnabled reports whether a plugin is enabled in the current tenant context.
	IsEnabled(ctx context.Context, pluginID string) bool
}

// CacheChangeNotifier publishes successful dynamic runtime cache mutations to
// the root plugin facade.
type CacheChangeNotifier interface {
	// MarkRuntimeCacheChanged records one cache-affecting runtime change.
	MarkRuntimeCacheChanged(ctx context.Context, reason string) error
}

// DependencyValidator validates candidate dynamic plugin releases before the
// reconciler switches effective release state or runs lifecycle side effects.
type DependencyValidator interface {
	// ValidateDynamicPluginCandidate verifies candidate dependencies and
	// reverse-dependency version safety for one dynamic lifecycle action.
	ValidateDynamicPluginCandidate(ctx context.Context, manifest *catalog.Manifest) error
}

// ArtifactService defines runtime WASM artifact parsing and validation operations.
type ArtifactService interface {
	// ParseRuntimeWasmArtifact reads one WASM artifact file and extracts all embedded custom sections.
	// It implements the catalog.ArtifactParser interface.
	ParseRuntimeWasmArtifact(filePath string) (*catalog.ArtifactSpec, error)
	// ParseRuntimeWasmArtifactContent parses one WASM artifact from an in-memory byte slice.
	// It implements the catalog.ArtifactParser interface.
	ParseRuntimeWasmArtifactContent(filePath string, content []byte) (*catalog.ArtifactSpec, error)
	// ValidateRuntimeArtifact loads and validates the WASM artifact for a dynamic plugin source directory.
	// It implements the catalog.ArtifactParser interface.
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

// DynamicCronService defines dynamic plugin cron discovery and execution operations.
type DynamicCronService interface {
	// DiscoverCronContracts runs the reserved guest-side cron registration entry
	// point and collects all declared dynamic-plugin cron contracts.
	DiscoverCronContracts(
		ctx context.Context,
		manifest *catalog.Manifest,
	) ([]*bridgecontract.CronContract, error)
	// ExecuteDeclaredCronJob runs one declared dynamic-plugin cron job through
	// the active runtime bridge.
	ExecuteDeclaredCronJob(
		ctx context.Context,
		manifest *catalog.Manifest,
		contract *bridgecontract.CronContract,
	) error
}

// LifecycleReconcileService defines runtime lifecycle convergence operations
// needed by install, enable, disable, and upgrade flows.
type LifecycleReconcileService interface {
	// ReconcileDynamicPluginRequest implements lifecycle.ReconcileProvider.
	// It submits a desired-state transition to the reconciler loop.
	ReconcileDynamicPluginRequest(ctx context.Context, pluginID string, desiredState string) error
	// EnsureRuntimeArtifactAvailable implements lifecycle.ReconcileProvider.
	// It verifies the WASM artifact is present for the given lifecycle action label.
	EnsureRuntimeArtifactAvailable(manifest *catalog.Manifest, actionLabel string) error
	// ShouldRefreshInstalledDynamicRelease implements lifecycle.ReconcileProvider.
	// It type-asserts registry to *entity.SysPlugin then delegates to the private helper.
	ShouldRefreshInstalledDynamicRelease(
		ctx context.Context,
		registry interface{},
		manifest *catalog.Manifest,
	) bool
}

// RuntimeRegistryService defines registry-backed dynamic plugin projection and
// artifact-state query operations.
type RuntimeRegistryService interface {
	// BuildPluginItem returns a PluginItem projection for one manifest + registry pair.
	// Used by the plugin facade SyncAndList coordination method.
	BuildPluginItem(ctx context.Context, manifest *catalog.Manifest, registry *entity.SysPlugin) *PluginItem
	// BuildRuntimeItems returns PluginItems for dynamic plugins present in the registry
	// but absent from the given manifest map. Used by the plugin facade SyncAndList.
	BuildRuntimeItems(ctx context.Context, covered map[string]struct{}) ([]*PluginItem, error)
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
	SyncPluginReleaseRuntimeState(ctx context.Context, registry *entity.SysPlugin) error
}

// RuntimeReconcilerService defines background and on-demand runtime reconciliation operations.
type RuntimeReconcilerService interface {
	// StartRuntimeReconciler starts the background loop that keeps dynamic-plugin
	// desired state, active release, and current-node projection converged.
	StartRuntimeReconciler(ctx context.Context)
	// ReconcileRuntimePlugins runs one convergence pass. It is safe to call from
	// both the background loop and synchronous management flows.
	ReconcileRuntimePlugins(ctx context.Context) error
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

// RuntimeLifecycleService defines dynamic plugin uninstall and emergency cleanup operations.
type RuntimeLifecycleService interface {
	// Uninstall executes uninstall lifecycle for an installed dynamic plugin.
	Uninstall(ctx context.Context, pluginID string) error
	// UninstallWithOptions executes uninstall lifecycle for an installed dynamic
	// plugin using one explicit cleanup policy snapshot.
	UninstallWithOptions(ctx context.Context, pluginID string, purgeStorageData bool) error
	// ForceUninstallMissingArtifact clears host governance for an installed
	// dynamic plugin whose staging and active release artifacts are unavailable.
	ForceUninstallMissingArtifact(ctx context.Context, registry *entity.SysPlugin) error
}

// ActiveManifestService defines active dynamic release manifest loading operations.
type ActiveManifestService interface {
	// LoadActiveDynamicPluginManifest implements catalog.DynamicManifestLoader.
	// It returns the currently active dynamic-plugin manifest reloaded from the stable
	// release archive so live traffic sees the stable version during staged upgrades.
	LoadActiveDynamicPluginManifest(ctx context.Context, registry *entity.SysPlugin) (*catalog.Manifest, error)
}

// DependencyWiringService defines provider wiring operations for runtime integrations.
type DependencyWiringService interface {
	// SetTopology wires the cluster topology provider.
	SetTopology(t TopologyProvider)
	// SetMenuManager wires the menu synchronization provider.
	SetMenuManager(m MenuManager)
	// SetHookDispatcher wires the lifecycle hook dispatcher.
	SetHookDispatcher(d HookDispatcher)
	// SetJwtConfigProvider wires the JWT configuration provider for route token validation.
	SetJwtConfigProvider(p JwtConfigProvider)
	// SetUploadSizeProvider wires the upload-size provider for dynamic package uploads.
	SetUploadSizeProvider(p UploadSizeProvider)
	// SetUserContextSetter wires the user-context injection provider.
	SetUserContextSetter(p UserContextSetter)
	// SetSessionStore wires the online-session store used for dynamic route requests.
	SetSessionStore(store session.Store)
	// SetPermissionMenuFilter wires the plugin-level permission menu filter.
	SetPermissionMenuFilter(f PermissionMenuFilter)
	// SetRuntimeCacheChangeNotifier wires cluster cache revision publication.
	SetRuntimeCacheChangeNotifier(n CacheChangeNotifier)
	// SetDependencyValidator wires release dependency validation.
	SetDependencyValidator(v DependencyValidator)
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
	DynamicCronService
	LifecycleReconcileService
	RuntimeRegistryService
	RuntimeProjectionService
	RuntimeReconcilerService
	RuntimeUpgradeService
	DynamicLifecyclePreconditionService
	RuntimeLifecycleService
	ActiveManifestService
	DependencyWiringService
	DynamicPackageService
}

// Ensure serviceImpl satisfies the composed runtime contract used by other plugin packages.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// catalogSvc provides manifest, registry, and release access.
	catalogSvc catalog.Service
	// lifecycleSvc provides install/uninstall SQL migration support.
	lifecycleSvc lifecycle.Service
	// frontendSvc manages in-memory frontend bundles.
	frontendSvc frontend.Service
	// openapiSvc projects dynamic routes into the host OpenAPI document.
	openapiSvc openapi.Service
	// topology provides cluster topology information.
	topology TopologyProvider
	// menuMgr handles plugin menu and permission synchronization.
	menuMgr MenuManager
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
	// menuFilter filters button-type permission menus by plugin enablement.
	menuFilter PermissionMenuFilter
	// cacheChangeNotifier publishes runtime cache changes after successful convergence.
	cacheChangeNotifier CacheChangeNotifier
	// dependencyValidator checks candidate release dependency constraints before
	// dynamic lifecycle side effects.
	dependencyValidator DependencyValidator
	// reconcilerRevisionObserved records the reconciler revision consumed by this runtime service.
	reconcilerRevisionObserved *pluginruntimecache.ObservedRevision
	// reconcilerRevisionCtrl coordinates cluster-wide dynamic-plugin reconciler wake-up.
	reconcilerRevisionCtrl *pluginruntimecache.Controller
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
	lifecycleSvc lifecycle.Service,
	frontendSvc frontend.Service,
	openapiSvc openapi.Service,
	i18nSvc runtimeI18nService,
) Service {
	return &serviceImpl{
		catalogSvc:                 catalogSvc,
		lifecycleSvc:               lifecycleSvc,
		frontendSvc:                frontendSvc,
		openapiSvc:                 openapiSvc,
		sessionStore:               session.NewDBStore(),
		reconcilerRevisionObserved: pluginruntimecache.NewObservedRevision(),
		i18nSvc:                    i18nSvc,
	}
}

// SetTopology wires the cluster topology provider.
func (s *serviceImpl) SetTopology(t TopologyProvider) {
	s.topology = t
	s.configureReconcilerRevisionController()
}

// SetMenuManager wires the menu synchronization provider.
func (s *serviceImpl) SetMenuManager(m MenuManager) {
	s.menuMgr = m
}

// SetHookDispatcher wires the lifecycle hook dispatcher.
func (s *serviceImpl) SetHookDispatcher(d HookDispatcher) {
	s.hookDispatcher = d
}

// SetJwtConfigProvider wires the JWT configuration provider for route token validation.
func (s *serviceImpl) SetJwtConfigProvider(p JwtConfigProvider) {
	s.jwtConfig = p
}

// SetUploadSizeProvider wires the upload-size provider for dynamic package uploads.
func (s *serviceImpl) SetUploadSizeProvider(p UploadSizeProvider) {
	s.uploadSize = p
}

// SetUserContextSetter wires the user-context injection provider.
func (s *serviceImpl) SetUserContextSetter(p UserContextSetter) {
	s.userCtx = p
}

// SetSessionStore wires the online-session store used for dynamic route requests.
func (s *serviceImpl) SetSessionStore(store session.Store) {
	if store != nil {
		s.sessionStore = store
	}
}

// SetPermissionMenuFilter wires the plugin-level permission menu filter.
func (s *serviceImpl) SetPermissionMenuFilter(f PermissionMenuFilter) {
	s.menuFilter = f
}

// SetRuntimeCacheChangeNotifier wires cluster cache revision publication.
func (s *serviceImpl) SetRuntimeCacheChangeNotifier(n CacheChangeNotifier) {
	s.cacheChangeNotifier = n
}

// SetDependencyValidator wires release dependency validation.
func (s *serviceImpl) SetDependencyValidator(v DependencyValidator) {
	s.dependencyValidator = v
}

// isClusterModeEnabled is a nil-safe wrapper around the topology provider.
func (s *serviceImpl) isClusterModeEnabled() bool {
	if s.topology == nil {
		return false
	}
	return s.topology.IsClusterModeEnabled()
}

// isPrimaryNode is a nil-safe wrapper around the topology provider.
func (s *serviceImpl) isPrimaryNode() bool {
	if s.topology == nil {
		return false
	}
	return s.topology.IsPrimaryNode()
}

// currentNodeID is a nil-safe wrapper around the topology provider.
func (s *serviceImpl) currentNodeID() string {
	if s.topology == nil {
		return ""
	}
	return s.topology.CurrentNodeID()
}

// dispatchHookEvent is a nil-safe wrapper for hook event dispatch.
func (s *serviceImpl) dispatchHookEvent(
	ctx context.Context,
	event pluginhost.ExtensionPoint,
	values map[string]interface{},
) error {
	if s.hookDispatcher == nil {
		return nil
	}
	return s.hookDispatcher.DispatchPluginHookEvent(ctx, event, values)
}

// syncPluginMenusAndPermissions is a nil-safe wrapper for menu synchronization.
func (s *serviceImpl) syncPluginMenusAndPermissions(ctx context.Context, manifest *catalog.Manifest) error {
	if s.menuMgr == nil {
		return nil
	}
	return s.menuMgr.SyncPluginMenusAndPermissions(ctx, manifest)
}

// syncPluginMenus is a nil-safe wrapper for partial menu synchronization (rollback path).
func (s *serviceImpl) syncPluginMenus(ctx context.Context, manifest *catalog.Manifest) error {
	if s.menuMgr == nil {
		return nil
	}
	return s.menuMgr.SyncPluginMenus(ctx, manifest)
}

// deletePluginMenusByManifest is a nil-safe wrapper for menu deletion.
func (s *serviceImpl) deletePluginMenusByManifest(ctx context.Context, manifest *catalog.Manifest) error {
	if s.menuMgr == nil {
		return nil
	}
	return s.menuMgr.DeletePluginMenusByManifest(ctx, manifest)
}

// ensureFrontendBundle delegates to frontendSvc to guarantee an in-memory bundle exists.
func (s *serviceImpl) ensureFrontendBundle(ctx context.Context, manifest *catalog.Manifest) error {
	if s.frontendSvc == nil {
		return nil
	}
	return s.frontendSvc.EnsureBundle(ctx, manifest)
}

// validateFrontendMenuBindings delegates frontend menu binding validation.
func (s *serviceImpl) validateFrontendMenuBindings(ctx context.Context, manifest *catalog.Manifest) error {
	if s.frontendSvc == nil {
		return nil
	}
	return s.frontendSvc.ValidateRuntimeFrontendMenuBindings(ctx, manifest)
}

// invalidateRuntimeCaches removes cached runtime frontend assets and runtime i18n
// bundles after one plugin lifecycle change. Only the dynamic-plugin sector for
// the affected plugin is invalidated; host and source-plugin sectors stay hot
// for unrelated locales and plugins.
func (s *serviceImpl) invalidateRuntimeCaches(ctx context.Context, pluginID string, reason string) {
	wasm.InvalidateAllCache(ctx)
	if s.frontendSvc != nil {
		s.frontendSvc.InvalidateBundle(ctx, pluginID, reason)
	}
	if s.i18nSvc != nil {
		s.i18nSvc.InvalidateRuntimeBundleCache(i18nsvc.InvalidateScope{
			Sectors:         []i18nsvc.Sector{i18nsvc.SectorDynamicPlugin},
			DynamicPluginID: pluginID,
		})
	}
}

// notifyRuntimeCacheChanged publishes a successful dynamic runtime mutation to
// other cluster nodes through the root plugin facade.
func (s *serviceImpl) notifyRuntimeCacheChanged(ctx context.Context, reason string) error {
	if s.cacheChangeNotifier == nil {
		return nil
	}
	return s.cacheChangeNotifier.MarkRuntimeCacheChanged(ctx, reason)
}
