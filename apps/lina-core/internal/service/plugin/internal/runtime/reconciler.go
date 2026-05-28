// This file implements the leader-aware dynamic-plugin reconciler. Management
// APIs persist the desired host state, while the primary node archives the
// staged artifact, performs migrations and menu switches, advances generation,
// and updates per-node convergence rows.

package runtime

import (
	"context"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/pkg/logger"
)

// runtimeReconcilerRevisionPollInterval is the clustered background cadence for
// checking the shared reconciler revision. Full scans run only when the revision
// changes or when the low-frequency safety sweep interval elapses.
const runtimeReconcilerRevisionPollInterval = 2 * time.Second

const (
	// runtimeReconcilerStaleReconcilingAfter is the conservative window after
	// which a leftover transient host state is treated as abandoned.
	runtimeReconcilerStaleReconcilingAfter = 5 * time.Minute
	// runtimeReconcilerDistributedLockLease bounds one per-plugin primary-side
	// reconcile lease in clustered deployments.
	runtimeReconcilerDistributedLockLease  = 30 * time.Minute
	runtimeReconcilerDistributedLockReason = "plugin-runtime-reconcile"
)

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

	s.runReconcilerTickSafely(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runReconcilerTickSafely(ctx)
		}
	}
}

// runReconcilerTickSafely isolates panics from one background tick so the loop
// can continue and stale transient rows can be recovered by later ticks.
func (s *serviceImpl) runReconcilerTickSafely(ctx context.Context) {
	defer func() {
		if panicValue := recover(); panicValue != nil {
			logger.Errorf(ctx, "dynamic plugin reconciler tick panic recovered panic=%v stack=%s", panicValue, string(debug.Stack()))
		}
	}()
	s.runReconcilerTick(ctx)
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

// RefreshInstalledRuntimePluginReleases repairs already installed dynamic
// releases whose active archive no longer matches the staged same-version
// artifact. It deliberately skips installs, uninstalls, and enablement changes.
func (s *serviceImpl) RefreshInstalledRuntimePluginReleases(ctx context.Context) error {
	reconcileMu.Lock()
	defer reconcileMu.Unlock()

	registries, err := s.listRuntimeRegistries(ctx)
	if err != nil {
		return err
	}
	if !s.isPrimaryNode() {
		return nil
	}

	var firstErr error
	for _, registry := range registries {
		if err = s.reconcilePrimaryPluginWithLock(ctx, registry, s.refreshInstalledRuntimePluginRelease); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// refreshInstalledRuntimePluginRelease executes the same-version refresh path
// only when an installed dynamic release is already bound to the discovered
// manifest version and the archived artifact/snapshot is stale.
func (s *serviceImpl) refreshInstalledRuntimePluginRelease(ctx context.Context, registry *entity.SysPlugin) error {
	if registry == nil ||
		catalog.NormalizeType(registry.Type) != catalog.TypeDynamic ||
		registry.Installed != catalog.InstalledYes {
		return nil
	}
	desiredManifest, err := s.catalogSvc.GetDesiredManifest(registry.PluginId)
	if err != nil {
		return err
	}
	if desiredManifest == nil || catalog.NormalizeType(desiredManifest.Type) != catalog.TypeDynamic {
		return nil
	}
	if strings.TrimSpace(desiredManifest.Version) != strings.TrimSpace(registry.Version) {
		return nil
	}
	if !s.shouldRefreshInstalledRelease(ctx, registry, desiredManifest) {
		return nil
	}
	desiredState := strings.TrimSpace(registry.DesiredState)
	if desiredState == "" {
		desiredState = catalog.BuildStableHostState(registry)
	}
	return s.applyRefresh(ctx, registry, desiredManifest, desiredState)
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
		return s.notifyReconcilerChanged(ctx, runtimeChangeReasonDesiredStateChanged)
	}
	if err := s.publishReconcilerChanged(ctx, runtimeChangeReasonDesiredStateChanged, false); err != nil {
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
		if err = s.reconcilePrimaryPluginWithLock(ctx, registry, s.reconcilePluginIfNeeded); err != nil {
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

	var err error
	registry, err = s.recoverStaleReconciling(ctx, registry)
	if err != nil {
		return err
	}
	if registry == nil {
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

// reconcilePrimaryPluginWithLock serializes primary-only lifecycle side
// effects. Single-node mode relies on the process mutex; clustered mode also
// acquires a per-plugin distributed lock before running the supplied callback.
func (s *serviceImpl) reconcilePrimaryPluginWithLock(
	ctx context.Context,
	registry *entity.SysPlugin,
	fn func(context.Context, *entity.SysPlugin) error,
) error {
	if registry == nil || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return nil
	}
	if fn == nil {
		return nil
	}
	locked, unlock, err := s.lockRuntimeReconcilePlugin(ctx, registry.PluginId)
	if err != nil {
		return err
	}
	if !locked {
		logger.Debugf(ctx, "skip dynamic plugin reconcile because per-plugin lock is held plugin=%s", registry.PluginId)
		return nil
	}
	defer unlock()
	return fn(ctx, registry)
}

// reconcilePrimaryPluginWithRequiredLock serializes explicit caller-initiated
// lifecycle side effects. Unlike the background reconciler, callers must see a
// lock-conflict error instead of observing a false success.
func (s *serviceImpl) reconcilePrimaryPluginWithRequiredLock(
	ctx context.Context,
	registry *entity.SysPlugin,
	fn func(context.Context, *entity.SysPlugin) error,
) error {
	if registry == nil || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return nil
	}
	if fn == nil {
		return nil
	}
	locked, unlock, err := s.lockRuntimeReconcilePlugin(ctx, registry.PluginId)
	if err != nil {
		return err
	}
	if !locked {
		return gerror.Newf("dynamic plugin reconciler lock is held: %s", registry.PluginId)
	}
	defer unlock()
	return fn(ctx, registry)
}

// recoverStaleReconciling restores abandoned transient host states before a
// primary-side reconcile pass. Fresh reconciling rows are skipped so another
// node or request can finish the in-flight lifecycle side effects.
func (s *serviceImpl) recoverStaleReconciling(ctx context.Context, registry *entity.SysPlugin) (*entity.SysPlugin, error) {
	if registry == nil || strings.TrimSpace(registry.CurrentState) != catalog.HostStateReconciling.String() {
		return registry, nil
	}
	if registry.UpdatedAt != nil && runtimeReconcilerUpdatedAtAge(registry.UpdatedAt, time.Now()) < runtimeReconcilerStaleReconcilingAfter {
		logger.Debugf(ctx, "skip fresh dynamic plugin reconciling state plugin=%s updatedAt=%s", registry.PluginId, registry.UpdatedAt.Format(time.RFC3339))
		return nil, nil
	}
	restored, err := s.restoreStableState(ctx, registry)
	if err != nil {
		return nil, err
	}
	if restored != nil {
		logger.Warningf(ctx, "restored stale dynamic plugin reconciling state plugin=%s", restored.PluginId)
		if err = s.notifyReconcilerChanged(ctx, runtimeChangeReasonStaleReconcilingRestored); err != nil {
			return nil, err
		}
	}
	return restored, nil
}

// runtimeReconcilerUpdatedAtAge computes an age for database TIMESTAMP values.
// PostgreSQL TIMESTAMP columns can be decoded with a UTC location while carrying
// the database wall-clock time; when that makes the instant appear in the future,
// reinterpret the wall-clock fields in the local runtime location.
func runtimeReconcilerUpdatedAtAge(updatedAt *time.Time, now time.Time) time.Duration {
	if updatedAt == nil {
		return runtimeReconcilerStaleReconcilingAfter
	}
	age := now.Sub(*updatedAt)
	if age >= 0 {
		return age
	}
	localWallClock := time.Date(
		updatedAt.Year(),
		updatedAt.Month(),
		updatedAt.Day(),
		updatedAt.Hour(),
		updatedAt.Minute(),
		updatedAt.Second(),
		updatedAt.Nanosecond(),
		now.Location(),
	)
	return now.Sub(localWallClock)
}

// lockRuntimeReconcilePlugin acquires the per-plugin lifecycle side-effect lock.
// Single-node mode is already protected by the process-wide reconcile mutex, so
// only clustered deployments need the distributed locker backend.
func (s *serviceImpl) lockRuntimeReconcilePlugin(ctx context.Context, pluginID string) (bool, func(), error) {
	if !s.isClusterModeEnabled() {
		return true, func() {}, nil
	}
	lockSvc := s.reconcilerLockSvc
	if lockSvc == nil {
		return false, nil, gerror.New("dynamic plugin reconciler lock service is not configured")
	}

	lockName := runtimeReconcilerLockName(pluginID)
	holder := s.runtimeReconcilerLockHolder()
	instance, ok, err := lockSvc.Lock(ctx, lockName, holder, runtimeReconcilerDistributedLockReason, runtimeReconcilerDistributedLockLease)
	if err != nil {
		return false, nil, err
	}
	if !ok || instance == nil {
		return false, func() {}, nil
	}
	return true, func() {
		if unlockErr := instance.Unlock(ctx); unlockErr != nil {
			logger.Warningf(ctx, "release dynamic plugin reconciler lock failed plugin=%s lock=%s err=%v", pluginID, lockName, unlockErr)
		}
	}, nil
}

// runtimeReconcilerLockName builds the stable per-plugin reconciler lock name.
func runtimeReconcilerLockName(pluginID string) string {
	return "plugin-runtime-reconcile:" + strings.TrimSpace(pluginID)
}

// runtimeReconcilerLockHolder returns a unique lock holder for this attempt.
func (s *serviceImpl) runtimeReconcilerLockHolder() string {
	nodeID := strings.TrimSpace(s.currentNodeID())
	if nodeID == "" {
		nodeID = "local-node"
	}
	return nodeID + ":" + guid.S()
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
