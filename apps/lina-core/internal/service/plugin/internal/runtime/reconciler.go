// This file implements the leader-aware dynamic-plugin reconciler. Management
// APIs persist the desired host state, while the primary node archives the
// staged artifact, performs migrations and menu switches, advances generation,
// and updates per-node convergence rows.

package runtime

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/pkg/logger"
)

// runtimeReconcilerRevisionPollInterval is the clustered background cadence for
// checking the shared reconciler revision. Full scans run only when the revision
// changes or when the low-frequency safety sweep interval elapses.
const runtimeReconcilerRevisionPollInterval = 2 * time.Second

// Background reconciler singletons ensure only one reconcile loop and one
// convergence pass run at a time inside the current process.
var (
	reconcilerOnce sync.Once
	reconcileMu    sync.Mutex
)

// StartRuntimeReconciler starts the background loop that keeps dynamic-plugin
// desired state, active release, and current-node projection converged.
func (s *serviceImpl) StartRuntimeReconciler(ctx context.Context) {
	if !s.isClusterModeEnabled() {
		return
	}
	reconcilerOnce.Do(func() {
		go s.runReconciler(context.WithoutCancel(ctx))
	})
}

// runReconciler executes the periodic background convergence loop used by
// clustered deployments.
func (s *serviceImpl) runReconciler(ctx context.Context) {
	ticker := time.NewTicker(runtimeReconcilerRevisionPollInterval)
	defer ticker.Stop()

	s.runReconcilerTick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runReconcilerTick(ctx)
		}
	}
}

// runReconcilerTick executes one revision-gated background reconcile check.
func (s *serviceImpl) runReconcilerTick(ctx context.Context) {
	decision, err := s.nextBackgroundReconcileDecision(ctx)
	if err != nil {
		logger.Warningf(ctx, "dynamic plugin reconciler revision check failed: %v", err)
		return
	}
	if !decision.shouldRun {
		return
	}
	if err = s.ReconcileRuntimePlugins(ctx); err != nil {
		logger.Warningf(ctx, "dynamic plugin reconciler tick failed reason=%s revision=%d err=%v", decision.reason, decision.revision, err)
		return
	}
	s.markBackgroundReconcileObserved(decision.revision, time.Now())
	logger.Debugf(ctx, "dynamic plugin reconciler tick completed reason=%s revision=%d", decision.reason, decision.revision)
}

// ReconcileRuntimePlugins runs one convergence pass. It is safe to call from
// both the background loop and synchronous management flows.
func (s *serviceImpl) ReconcileRuntimePlugins(ctx context.Context) error {
	reconcileMu.Lock()
	defer reconcileMu.Unlock()

	registries, err := s.listRuntimeRegistries(ctx)
	if err != nil {
		return err
	}

	isPrimary := s.isPrimaryNode()

	var firstErr error
	for _, registry := range registries {
		if err = s.reconcileRuntimeRegistry(ctx, registry, isPrimary); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// reconcileDynamicPluginRequest records the requested target state and lets the
// primary node converge the addressed plugin immediately.
func (s *serviceImpl) reconcileDynamicPluginRequest(
	ctx context.Context,
	pluginID string,
	desiredState catalog.HostState,
) error {
	if err := s.updateDesiredState(ctx, pluginID, desiredState); err != nil {
		return err
	}
	if !s.isPrimaryNode() {
		return s.notifyReconcilerChanged(ctx, "desired_state_changed")
	}
	if err := s.publishReconcilerChanged(ctx, "desired_state_changed", false); err != nil {
		return err
	}
	return s.reconcileRuntimePlugin(ctx, pluginID)
}

// reconcileRuntimePlugin converges one target plugin synchronously for
// management requests. Unlike the background full scan, it must not fail a
// user-triggered install/refresh because some unrelated staged dynamic plugin is
// temporarily broken in the shared registry during other tests or uploads.
func (s *serviceImpl) reconcileRuntimePlugin(ctx context.Context, pluginID string) error {
	reconcileMu.Lock()
	defer reconcileMu.Unlock()

	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil {
		return gerror.New("plugin does not exist")
	}
	return s.reconcileRuntimeRegistry(ctx, registry, true)
}

// reconcileRuntimeRegistry converges one runtime registry row, optionally
// performing primary-only lifecycle work before updating current-node state.
func (s *serviceImpl) reconcileRuntimeRegistry(
	ctx context.Context,
	registry *entity.SysPlugin,
	isPrimary bool,
) error {
	if registry == nil {
		return nil
	}

	pluginID := registry.PluginId

	// Refresh the registry against current artifact presence before any lifecycle
	// action so missing or newly restored packages are reflected consistently.
	refreshedRegistry, err := s.reconcileRegistryArtifactState(ctx, registry)
	if err != nil {
		logger.Warningf(ctx, "reconcile runtime registry artifact state failed plugin=%s err=%v", pluginID, err)
		return err
	}
	if refreshedRegistry == nil {
		return nil
	}
	registry = refreshedRegistry

	if isPrimary {
		// Only the primary node mutates shared lifecycle state such as release
		// activation, migrations, and desired/current host states.
		if err = s.reconcilePluginIfNeeded(ctx, registry); err != nil {
			logger.Warningf(ctx, "reconcile dynamic plugin failed plugin=%s err=%v", pluginID, err)
			return err
		}
		// Reload after lifecycle work so node projection sees the latest release
		// binding, generation, and stable host state.
		registry, err = s.catalogSvc.GetRegistry(ctx, registry.PluginId)
		if err != nil {
			logger.Warningf(ctx, "reload dynamic plugin registry failed plugin=%s err=%v", pluginID, err)
			return err
		}
	}
	if registry == nil {
		return nil
	}
	if err = s.reconcileCurrentNodeProjection(ctx, registry); err != nil {
		logger.Warningf(ctx, "reconcile current node projection failed plugin=%s err=%v", pluginID, err)
		return err
	}
	return nil
}

// reconcilePluginIfNeeded selects the smallest convergence action for the current
// registry row: install, upgrade, same-version refresh, state toggle, or uninstall.
func (s *serviceImpl) reconcilePluginIfNeeded(ctx context.Context, registry *entity.SysPlugin) error {
	if registry == nil || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return nil
	}

	desiredState := strings.TrimSpace(registry.DesiredState)
	if desiredState == "" {
		desiredState = catalog.BuildStableHostState(registry)
	}
	stableState := catalog.BuildStableHostState(registry)
	if desiredState == catalog.HostStateUninstalled.String() {
		if registry.Installed != catalog.InstalledYes {
			return nil
		}
		return s.applyUninstall(ctx, registry)
	}

	desiredManifest, err := s.catalogSvc.GetDesiredManifest(registry.PluginId)
	if err != nil {
		return err
	}
	if desiredManifest == nil || catalog.NormalizeType(desiredManifest.Type) != catalog.TypeDynamic {
		return gerror.New("dynamic plugin desired manifest does not exist")
	}

	if registry.Installed != catalog.InstalledYes {
		return s.applyInstall(ctx, registry, desiredManifest, desiredState)
	}
	if strings.TrimSpace(desiredManifest.Version) != strings.TrimSpace(registry.Version) {
		// Version drift is intentionally left as a pending runtime upgrade. Only
		// the explicit management API is allowed to run upgrade side effects.
		return nil
	}
	if s.shouldRefreshInstalledRelease(ctx, registry, desiredManifest) {
		// Same semantic version can still require refresh when the staged artifact,
		// archive bytes, or synthesized checksum changed after a rebuild.
		return s.applyRefresh(ctx, registry, desiredManifest, desiredState)
	}
	if desiredState != stableState {
		return s.applyStateToggle(ctx, registry, desiredManifest, desiredState)
	}
	return nil
}

// reconcileCurrentNodeProjection verifies the current node can serve the active
// dynamic plugin state and then persists the node-local convergence snapshot.
func (s *serviceImpl) reconcileCurrentNodeProjection(ctx context.Context, registry *entity.SysPlugin) error {
	if registry == nil || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return nil
	}

	// Enabled dynamic plugins must prove their active manifest and optional
	// frontend bundle still load on this node before we mark the node converged.
	if registry.Installed == catalog.InstalledYes && registry.Status == catalog.StatusEnabled && registry.ReleaseId > 0 {
		manifest, err := s.loadActiveManifest(ctx, registry)
		if err != nil {
			return s.syncNodeProjection(ctx, nodeProjectionInput{
				PluginID:     registry.PluginId,
				ReleaseID:    registry.ReleaseId,
				DesiredState: registry.DesiredState,
				CurrentState: catalog.NodeStateFailed.String(),
				Generation:   registry.Generation,
				Message:      err.Error(),
			})
		}
		if frontend.HasFrontendAssets(manifest) {
			if err = s.ensureFrontendBundle(ctx, manifest); err != nil {
				return s.syncNodeProjection(ctx, nodeProjectionInput{
					PluginID:     registry.PluginId,
					ReleaseID:    registry.ReleaseId,
					DesiredState: registry.DesiredState,
					CurrentState: catalog.NodeStateFailed.String(),
					Generation:   registry.Generation,
					Message:      err.Error(),
				})
			}
		}
	}

	return s.syncNodeProjection(ctx, nodeProjectionInput{
		PluginID:     registry.PluginId,
		ReleaseID:    registry.ReleaseId,
		DesiredState: registry.DesiredState,
		CurrentState: registry.CurrentState,
		Generation:   registry.Generation,
		Message:      "Current node converged to host plugin generation.",
	})
}
