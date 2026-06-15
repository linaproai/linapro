// This file coordinates plugin runtime cache freshness, enabled snapshots, and
// runtime revision publishing across cluster nodes.

package plugin

import (
	"context"
	"path/filepath"
	"strings"

	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cachecoord/revisionctrl"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/pkg/logger"
)

// pluginI18nService defines the i18n methods needed by plugin lifecycle,
// runtime cache refresh, and source-plugin reason rendering paths.
type pluginI18nService interface {
	// GetLocale returns the effective request locale stored in business context.
	GetLocale(ctx context.Context) string
	// BundleVersion returns the per-locale runtime translation bundle version.
	BundleVersion(locale string) uint64
	// InvalidateRuntimeBundleCache clears cached runtime bundles for one explicit scope.
	InvalidateRuntimeBundleCache(scope i18nsvc.InvalidateScope)
	// Translate renders one runtime i18n key in the current request locale.
	Translate(ctx context.Context, key string, fallback string) string
}

// pluginChangePublishInput identifies one successful plugin governance change
// and the derived runtime caches that must observe it.
type pluginChangePublishInput struct {
	pluginID   string
	pluginType string
	reason     string
}

// newRuntimeCacheRevisionController creates the cluster-aware revision
// controller used by the root plugin service.
func newRuntimeCacheRevisionController(
	topology Topology,
	cacheCoordSvc cachecoord.Service,
	integrationSvc pluginRuntimeIntegrationRefresher,
	frontendSvc pluginRuntimeFrontendInvalidator,
	i18nSvc pluginI18nService,
	managementListInvalidator pluginManagementListInvalidator,
	openapiInvalidator pluginOpenAPIProjectionInvalidator,
	wasmRuntime wasm.Runtime,
	catalogSvc catalog.Service,
	storeSvc store.Service,
	runtimeArtifactReconciler pluginRuntimeArtifactReconciler,
) *revisionctrl.Controller {
	clusterEnabled := false
	if topology != nil {
		clusterEnabled = topology.IsEnabled()
	}
	return revisionctrl.NewControllerWithCoordinator(
		clusterEnabled,
		cacheCoordSvc,
		revisionctrl.NewObservedRevision(),
		func(ctx context.Context, revision int64) error {
			if integrationSvc != nil {
				if err := integrationSvc.RefreshEnabledSnapshot(ctx); err != nil {
					return err
				}
			}
			if frontendSvc != nil {
				frontendSvc.InvalidateAllBundles(ctx, "cluster_runtime_revision_changed")
			}
			if managementListInvalidator != nil {
				managementListInvalidator.InvalidateManagementListCache(ctx, "cluster_runtime_revision_changed")
			}
			if openapiInvalidator != nil {
				openapiInvalidator.InvalidateProjectionCache(ctx, "cluster_runtime_revision_changed")
			}
			if wasmRuntime != nil {
				if runtimeArtifactReconciler != nil {
					runtimeArtifactReconciler.ReconcileActiveDynamicArtifactCaches(ctx, "cluster_runtime_revision_changed")
				} else {
					invalidateActiveDynamicWasmCaches(ctx, wasmRuntime, catalogSvc, storeSvc, "cluster_runtime_revision_changed")
				}
			}
			if i18nSvc != nil {
				i18nSvc.InvalidateRuntimeBundleCache(i18nsvc.InvalidateScope{
					Sectors: []i18nsvc.Sector{
						i18nsvc.SectorSourcePlugin,
						i18nsvc.SectorDynamicPlugin,
					},
				})
			}
			return nil
		},
	)
}

// pluginRuntimeIntegrationRefresher narrows the integration cache refresh dependency.
type pluginRuntimeIntegrationRefresher interface {
	// RefreshEnabledSnapshot rebuilds the in-memory plugin enablement snapshot.
	RefreshEnabledSnapshot(ctx context.Context) error
}

// pluginRuntimeFrontendInvalidator narrows the frontend bundle invalidation dependency.
type pluginRuntimeFrontendInvalidator interface {
	// InvalidateAllBundles removes every cached runtime frontend bundle.
	InvalidateAllBundles(ctx context.Context, reason string)
}

// pluginManagementListInvalidator narrows the root read-model invalidation callback.
type pluginManagementListInvalidator interface {
	// InvalidateManagementListCache clears the plugin management read model.
	InvalidateManagementListCache(ctx context.Context, reason string)
}

// pluginOpenAPIProjectionInvalidator narrows dynamic route OpenAPI cache invalidation.
type pluginOpenAPIProjectionInvalidator interface {
	// InvalidateProjectionCache clears cached dynamic route OpenAPI projections.
	InvalidateProjectionCache(ctx context.Context, reason string)
}

// pluginRuntimeArtifactReconciler reconciles dynamic artifact-derived caches
// after a peer runtime revision change.
type pluginRuntimeArtifactReconciler interface {
	// ReconcileActiveDynamicArtifactCaches invalidates only active dynamic artifacts
	// whose path changed or disappeared since the previous reconciliation.
	ReconcileActiveDynamicArtifactCaches(ctx context.Context, reason string)
}

// ensureRuntimeCacheFresh synchronizes plugin runtime caches with the shared
// cluster revision before read paths consume process-local snapshots.
func (s *serviceImpl) ensureRuntimeCacheFresh(ctx context.Context) error {
	if s == nil || s.runtimeCacheRevisionCtrl == nil {
		return nil
	}
	return s.runtimeCacheRevisionCtrl.EnsureFresh(ctx)
}

// ensureRuntimeCacheFreshBestEffort logs revision refresh failures for methods
// that cannot return an error to their caller.
func (s *serviceImpl) ensureRuntimeCacheFreshBestEffort(ctx context.Context, operation string) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		logger.Warningf(ctx, "refresh plugin runtime cache failed operation=%s err=%v", operation, err)
	}
}

// MarkRuntimeCacheChanged publishes one successful runtime cache mutation to
// other cluster nodes. It implements the dynamic runtime cache-change notifier.
func (s *serviceImpl) MarkRuntimeCacheChanged(ctx context.Context, reason string) error {
	_, err := s.publishPluginChange(ctx, pluginChangePublishInput{reason: reason})
	return err
}

// PublishPluginChange publishes one successful plugin-scoped mutation to other
// cluster nodes. It implements the lifecycle and runtime cache-change notifier.
func (s *serviceImpl) PublishPluginChange(
	ctx context.Context,
	pluginID string,
	pluginType string,
	reason string,
) error {
	_, err := s.publishPluginChange(ctx, pluginChangePublishInput{
		pluginID:   pluginID,
		pluginType: pluginType,
		reason:     reason,
	})
	return err
}

// markRuntimeCacheChanged bumps the shared plugin runtime cache revision in
// cluster mode and is a no-op in single-node deployments.
func (s *serviceImpl) markRuntimeCacheChanged(ctx context.Context, reason string) (int64, error) {
	return s.publishPluginChange(ctx, pluginChangePublishInput{reason: reason})
}

// publishPluginChange is the single root-facade publication path for plugin
// governance mutations. It invalidates local derived caches, clears the
// management read model, and publishes the shared plugin-runtime revision so
// cluster peers observe the same change.
func (s *serviceImpl) publishPluginChange(ctx context.Context, input pluginChangePublishInput) (int64, error) {
	if s == nil {
		return 0, nil
	}
	reason := input.reason
	s.invalidateRuntimeUpgradeCaches(ctx, input.pluginID, input.pluginType, reason)
	s.InvalidateManagementListCache(ctx, reason)
	if s.openapiSvc != nil {
		s.openapiSvc.InvalidateProjectionCache(ctx, reason)
	}
	if s.runtimeCacheRevisionCtrl == nil {
		return 0, nil
	}
	revision, err := s.runtimeCacheRevisionCtrl.MarkChanged(ctx)
	if err != nil {
		return 0, err
	}
	if revision > 0 {
		logger.Debugf(ctx, "plugin runtime cache revision bumped reason=%s revision=%d", reason, revision)
	}
	return revision, nil
}

// syncEnabledSnapshotFromRegistry refreshes the in-memory enablement snapshot
// for one plugin using the latest registry row after a lifecycle transition.
func (s *serviceImpl) syncEnabledSnapshotFromRegistry(ctx context.Context, pluginID string) error {
	return s.syncEnabledSnapshotStateFromRegistry(ctx, pluginID)
}

// syncEnabledSnapshotStateFromRegistry updates only the in-memory enabled
// snapshot for the same registry state.
func (s *serviceImpl) syncEnabledSnapshotStateFromRegistry(
	ctx context.Context,
	pluginID string,
) error {
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil || registry.Installed != plugintypes.InstalledYes {
		s.integrationSvc.DeletePluginEnabledState(pluginID)
		return nil
	}
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	runtimeState, err := s.storeSvc.BuildRuntimeUpgradeState(ctx, registry, manifest)
	if err != nil {
		return err
	}
	enabled := registry.Status == plugintypes.StatusEnabled &&
		store.RuntimeStateAllowsBusinessEntry(runtimeState.State)
	s.integrationSvc.SetPluginEnabledState(pluginID, enabled)
	return nil
}

// syncEnabledSnapshotAndPublishRuntimeChange updates local enablement, publishes
// the runtime revision, and lets capability providers observe the refreshed
// platform enabled snapshot at use time.
func (s *serviceImpl) syncEnabledSnapshotAndPublishRuntimeChange(
	ctx context.Context,
	pluginID string,
	reason string,
) error {
	if err := s.syncEnabledSnapshotStateFromRegistry(ctx, pluginID); err != nil {
		return err
	}
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	pluginType := ""
	if registry != nil {
		pluginType = registry.Type
	}
	_, err = s.publishPluginChange(ctx, pluginChangePublishInput{
		pluginID:   pluginID,
		pluginType: pluginType,
		reason:     reason,
	})
	return err
}

// invalidateRuntimeUpgradeCaches clears this node's plugin-scoped derived
// runtime caches after an explicit upgrade succeeds or fails. Cluster peers
// receive the same mutation through the shared plugin-runtime revision.
func (s *serviceImpl) invalidateRuntimeUpgradeCaches(ctx context.Context, pluginID string, pluginType string, reason string) {
	if s == nil {
		return
	}
	normalizedPluginID := strings.TrimSpace(pluginID)
	normalizedType := plugintypes.NormalizeType(pluginType)
	if normalizedPluginID == "" {
		if s.frontendSvc != nil {
			s.frontendSvc.InvalidateAllBundles(ctx, reason)
		}
		if s.wasmRuntime != nil {
			s.wasmRuntime.InvalidateAllCache(ctx)
		}
		if s.i18nSvc != nil {
			s.i18nSvc.InvalidateRuntimeBundleCache(i18nsvc.InvalidateScope{
				Sectors: []i18nsvc.Sector{
					i18nsvc.SectorSourcePlugin,
					i18nsvc.SectorDynamicPlugin,
				},
			})
		}
		return
	}
	if s.frontendSvc != nil {
		s.frontendSvc.InvalidateBundle(ctx, normalizedPluginID, reason)
	}
	if normalizedType == plugintypes.TypeDynamic && s.wasmRuntime != nil {
		s.invalidateDynamicWasmCacheForPlugin(ctx, normalizedPluginID, reason)
	}
	if s.i18nSvc == nil {
		return
	}
	switch normalizedType {
	case plugintypes.TypeSource:
		s.i18nSvc.InvalidateRuntimeBundleCache(i18nsvc.InvalidateScope{
			Sectors:        []i18nsvc.Sector{i18nsvc.SectorSourcePlugin},
			SourcePluginID: normalizedPluginID,
		})
	case plugintypes.TypeDynamic:
		s.i18nSvc.InvalidateRuntimeBundleCache(i18nsvc.InvalidateScope{
			Sectors:         []i18nsvc.Sector{i18nsvc.SectorDynamicPlugin},
			DynamicPluginID: normalizedPluginID,
		})
	default:
		s.i18nSvc.InvalidateRuntimeBundleCache(i18nsvc.InvalidateScope{
			Sectors: []i18nsvc.Sector{
				i18nsvc.SectorSourcePlugin,
				i18nsvc.SectorDynamicPlugin,
			},
		})
	}
}

// invalidateDynamicWasmCacheForPlugin removes compiled WASM modules for one
// dynamic plugin's active artifact. Unknown paths fall back to a full invalidation
// because executing stale code is less safe than recompiling unrelated plugins.
func (s *serviceImpl) invalidateDynamicWasmCacheForPlugin(ctx context.Context, pluginID string, reason string) {
	if s == nil || s.wasmRuntime == nil {
		return
	}
	artifactPath, err := activeDynamicArtifactPath(ctx, s.catalogSvc, s.storeSvc, pluginID)
	if err != nil {
		logger.Warningf(ctx, "resolve dynamic plugin artifact for cache invalidation failed plugin=%s reason=%s err=%v", pluginID, reason, err)
		s.wasmRuntime.InvalidateAllCache(ctx)
		return
	}
	if strings.TrimSpace(artifactPath) == "" {
		return
	}
	s.wasmRuntime.InvalidateCache(ctx, artifactPath)
}

// invalidateActiveDynamicWasmCaches reconciles peer-observed runtime revision
// changes by invalidating currently active dynamic artifact paths. This keeps
// the operation bounded by installed dynamic plugin count instead of dropping
// all compiled modules unconditionally on every peer revision.
func invalidateActiveDynamicWasmCaches(
	ctx context.Context,
	wasmRuntime wasm.Runtime,
	catalogSvc catalog.Service,
	storeSvc store.Service,
	reason string,
) {
	if wasmRuntime == nil || storeSvc == nil {
		return
	}
	registries, err := storeSvc.ListAllRegistries(ctx)
	if err != nil {
		logger.Warningf(ctx, "list registries for dynamic wasm cache reconciliation failed reason=%s err=%v", reason, err)
		wasmRuntime.InvalidateAllCache(ctx)
		return
	}
	for _, registry := range registries {
		if registry == nil ||
			plugintypes.NormalizeType(registry.Type) != plugintypes.TypeDynamic ||
			registry.Installed != plugintypes.InstalledYes ||
			registry.ReleaseId <= 0 {
			continue
		}
		artifactPath, pathErr := activeDynamicArtifactPath(ctx, catalogSvc, storeSvc, registry.PluginId)
		if pathErr != nil {
			logger.Warningf(ctx, "resolve active dynamic artifact during cache reconciliation failed plugin=%s reason=%s err=%v", registry.PluginId, reason, pathErr)
			continue
		}
		if strings.TrimSpace(artifactPath) != "" {
			wasmRuntime.InvalidateCache(ctx, artifactPath)
		}
	}
}

// ReconcileActiveDynamicArtifactCaches invalidates dynamic artifact-derived
// caches whose active artifact path changed or disappeared since the previous
// peer revision observation.
func (s *serviceImpl) ReconcileActiveDynamicArtifactCaches(ctx context.Context, reason string) {
	if s == nil || s.wasmRuntime == nil {
		return
	}
	current, err := buildActiveDynamicArtifactSnapshot(ctx, s.catalogSvc, s.storeSvc)
	if err != nil {
		logger.Warningf(ctx, "build active dynamic artifact snapshot failed reason=%s err=%v", reason, err)
		s.wasmRuntime.InvalidateAllCache(ctx)
		return
	}
	s.activeDynamicArtifactSnapshotMu.Lock()
	previous := s.activeDynamicArtifactSnapshot
	if previous == nil {
		previous = map[string]string{}
	}
	for pluginID, oldPath := range previous {
		newPath := current[pluginID]
		if oldPath != "" && oldPath != newPath {
			s.wasmRuntime.InvalidateCache(ctx, oldPath)
		}
	}
	for pluginID, newPath := range current {
		if newPath == "" || previous[pluginID] == newPath {
			continue
		}
		s.wasmRuntime.InvalidateCache(ctx, newPath)
	}
	s.activeDynamicArtifactSnapshot = current
	s.activeDynamicArtifactSnapshotMu.Unlock()
}

// buildActiveDynamicArtifactSnapshot returns active dynamic artifact paths by
// plugin ID without parsing artifact content.
func buildActiveDynamicArtifactSnapshot(
	ctx context.Context,
	catalogSvc catalog.Service,
	storeSvc store.Service,
) (map[string]string, error) {
	out := make(map[string]string)
	if storeSvc == nil {
		return out, nil
	}
	registries, err := storeSvc.ListAllRegistries(ctx)
	if err != nil {
		return nil, err
	}
	for _, registry := range registries {
		if registry == nil ||
			plugintypes.NormalizeType(registry.Type) != plugintypes.TypeDynamic ||
			registry.Installed != plugintypes.InstalledYes ||
			registry.ReleaseId <= 0 {
			continue
		}
		artifactPath, pathErr := activeDynamicArtifactPath(ctx, catalogSvc, storeSvc, registry.PluginId)
		if pathErr != nil {
			return nil, pathErr
		}
		if strings.TrimSpace(artifactPath) != "" {
			out[strings.TrimSpace(registry.PluginId)] = artifactPath
		}
	}
	return out, nil
}

// activeDynamicArtifactPath resolves the active dynamic release package path to
// an absolute filesystem path without parsing the WASM artifact.
func activeDynamicArtifactPath(
	ctx context.Context,
	catalogSvc catalog.Service,
	storeSvc store.Service,
	pluginID string,
) (string, error) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" || storeSvc == nil {
		return "", nil
	}
	registry, err := storeSvc.GetRegistry(ctx, normalizedPluginID)
	if err != nil {
		return "", err
	}
	if registry == nil ||
		plugintypes.NormalizeType(registry.Type) != plugintypes.TypeDynamic ||
		registry.ReleaseId <= 0 {
		return "", nil
	}
	release, err := storeSvc.GetRegistryRelease(ctx, registry)
	if err != nil || release == nil {
		return "", err
	}
	packagePath := strings.TrimSpace(release.PackagePath)
	if packagePath == "" {
		return "", nil
	}
	if filepath.IsAbs(packagePath) {
		return filepath.Clean(packagePath), nil
	}
	if catalogSvc == nil {
		return "", nil
	}
	storageDir, err := catalogSvc.RuntimeStorageDir(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(storageDir, filepath.FromSlash(packagePath))), nil
}
