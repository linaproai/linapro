// This file orchestrates explicit plugin runtime upgrade execution after the
// management API has collected operator confirmation.

package plugin

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
)

const (
	// runtimeUpgradeDistributedLockLease bounds orphaned upgrade locks after a
	// node crash while remaining longer than normal plugin SQL/governance phases.
	runtimeUpgradeDistributedLockLease = 30 * time.Minute
	// runtimeUpgradeDistributedLockReason records the owner purpose in coordination backends.
	runtimeUpgradeDistributedLockReason = "plugin-runtime-upgrade"
)

// ExecuteRuntimeUpgrade runs one explicit runtime upgrade after confirmation.
func (s *serviceImpl) ExecuteRuntimeUpgrade(
	ctx context.Context,
	pluginID string,
	options RuntimeUpgradeOptions,
) (*RuntimeUpgradeResult, error) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return nil, bizerr.NewCode(CodePluginNotFound, bizerr.P("pluginId", normalizedPluginID))
	}
	if !options.Confirmed {
		return nil, bizerr.NewCode(
			CodePluginRuntimeUpgradeConfirmationRequired,
			bizerr.P("pluginId", normalizedPluginID),
		)
	}
	unlock, err := s.lockRuntimeUpgrade(ctx, normalizedPluginID)
	if err != nil {
		return nil, err
	}
	defer unlock()

	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}

	targetManifest, registry, projection, err := s.loadRuntimeUpgradeExecutionState(ctx, normalizedPluginID)
	if err != nil {
		return nil, err
	}
	if !runtimeUpgradeCanExecute(projection.State) {
		return nil, bizerr.NewCode(
			CodePluginRuntimeUpgradeUnavailable,
			bizerr.P("pluginId", normalizedPluginID),
			bizerr.P("runtimeState", projection.State.String()),
		)
	}
	if catalog.NormalizeType(targetManifest.Type) == catalog.TypeDynamic {
		if err = s.ensureDynamicPluginUpgradeLifecyclePreconditionAllowed(
			ctx,
			registry,
			targetManifest,
			options.Authorization,
		); err != nil {
			return nil, err
		}
	}
	if err = s.markRuntimeUpgradeRunning(ctx, registry); err != nil {
		return nil, err
	}

	result := &RuntimeUpgradeResult{
		PluginID:          normalizedPluginID,
		FromVersion:       projection.EffectiveVersion,
		ToVersion:         projection.DiscoveredVersion,
		EffectiveVersion:  projection.EffectiveVersion,
		DiscoveredVersion: projection.DiscoveredVersion,
	}

	if err = s.executeRuntimeUpgradeByType(ctx, registry, targetManifest, options); err != nil {
		if restoreErr := s.restoreRuntimeUpgradeStableState(ctx, normalizedPluginID); restoreErr != nil {
			logger.Warningf(
				ctx,
				"restore runtime upgrade stable state after failure failed plugin=%s err=%v",
				normalizedPluginID, restoreErr,
			)
		}
		s.invalidateRuntimeUpgradeCaches(ctx, normalizedPluginID, targetManifest.Type, "runtime_upgrade_failed")
		if syncErr := s.syncEnabledSnapshotFromRegistry(ctx, normalizedPluginID); syncErr != nil {
			logger.Warningf(
				ctx,
				"sync plugin snapshot after runtime upgrade failure failed plugin=%s err=%v",
				normalizedPluginID, syncErr,
			)
		}
		if markErr := s.markRuntimeCacheChanged(ctx, "runtime_upgrade_failed"); markErr != nil {
			logger.Warningf(
				ctx,
				"mark runtime cache changed after upgrade failure failed plugin=%s err=%v",
				normalizedPluginID, markErr,
			)
		}
		return nil, bizerr.WrapCode(
			err,
			CodePluginRuntimeUpgradeExecutionFailed,
			bizerr.P("pluginId", normalizedPluginID),
			bizerr.P("fromVersion", result.FromVersion),
			bizerr.P("toVersion", result.ToVersion),
		)
	}
	if catalog.NormalizeType(targetManifest.Type) == catalog.TypeDynamic {
		s.executeDynamicPluginUpgradeLifecycleNotification(
			ctx,
			registry,
			targetManifest,
			options.Authorization,
		)
	}

	if err = s.syncEnabledSnapshotFromRegistry(ctx, normalizedPluginID); err != nil {
		return nil, err
	}
	s.invalidateRuntimeUpgradeCaches(ctx, normalizedPluginID, targetManifest.Type, "runtime_upgrade_succeeded")
	refreshedManifest, refreshedRegistry, refreshedProjection, err := s.loadRuntimeUpgradeExecutionState(
		ctx,
		normalizedPluginID,
	)
	if err != nil {
		return nil, err
	}
	result.Executed = true
	result.RuntimeState = RuntimeUpgradeState(refreshedProjection.State)
	result.EffectiveVersion = refreshedProjection.EffectiveVersion
	result.DiscoveredVersion = refreshedProjection.DiscoveredVersion
	if refreshedManifest == nil && refreshedRegistry != nil {
		result.EffectiveVersion = refreshedRegistry.Version
		result.DiscoveredVersion = refreshedRegistry.Version
	}
	return result, nil
}

// runtimeUpgradeCanExecute reports whether the explicit management upgrade
// endpoint may run side effects for a runtime-upgrade state.
func runtimeUpgradeCanExecute(state catalog.RuntimeUpgradeState) bool {
	return state == catalog.RuntimeUpgradeStatePendingUpgrade ||
		state == catalog.RuntimeUpgradeStateUpgradeFailed
}

// loadRuntimeUpgradeExecutionState re-reads the target manifest, registry row,
// and version-drift projection from authoritative storage.
func (s *serviceImpl) loadRuntimeUpgradeExecutionState(
	ctx context.Context,
	pluginID string,
) (*catalog.Manifest, *entity.SysPlugin, catalog.RuntimeUpgradeProjection, error) {
	targetManifest, err := s.loadDesiredManifestForPreview(pluginID)
	if err != nil {
		return nil, nil, catalog.RuntimeUpgradeProjection{}, err
	}
	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return nil, nil, catalog.RuntimeUpgradeProjection{}, err
	}
	if registry == nil {
		return nil, nil, catalog.RuntimeUpgradeProjection{}, bizerr.NewCode(
			CodePluginNotFound,
			bizerr.P("pluginId", pluginID),
		)
	}
	projection, err := s.catalogSvc.BuildRuntimeUpgradeState(ctx, registry, targetManifest)
	if err != nil {
		return nil, nil, catalog.RuntimeUpgradeProjection{}, err
	}
	return targetManifest, registry, projection, nil
}

// markRuntimeUpgradeRunning records an observable in-progress state before the
// upgrade starts. Projection later reports pending/failed/normal from the
// authoritative release and version state after execution completes.
func (s *serviceImpl) markRuntimeUpgradeRunning(ctx context.Context, registry *entity.SysPlugin) error {
	if registry == nil {
		return bizerr.NewCode(CodePluginNotFound, bizerr.P("pluginId", ""))
	}
	return s.catalogSvc.SetRegistryRuntimeState(ctx, registry.PluginId, do.SysPlugin{
		CurrentState: catalog.RuntimeUpgradeStateUpgradeRunning.String(),
	})
}

// restoreRuntimeUpgradeStableState clears transient runtime-upgrade markers
// after a failed explicit upgrade so projection can expose upgrade_failed or
// pending_upgrade from release/version state instead of staying in running.
func (s *serviceImpl) restoreRuntimeUpgradeStableState(ctx context.Context, pluginID string) error {
	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil {
		return nil
	}
	stableState := catalog.BuildStableHostState(registry)
	return s.catalogSvc.SetRegistryRuntimeState(ctx, registry.PluginId, do.SysPlugin{
		DesiredState: stableState,
		CurrentState: stableState,
	})
}

// executeRuntimeUpgradeByType dispatches the confirmed upgrade to the source or
// dynamic implementation while keeping common state validation in the root facade.
func (s *serviceImpl) executeRuntimeUpgradeByType(
	ctx context.Context,
	registry *entity.SysPlugin,
	targetManifest *catalog.Manifest,
	options RuntimeUpgradeOptions,
) error {
	if targetManifest == nil {
		return bizerr.NewCode(CodePluginNotFound, bizerr.P("pluginId", ""))
	}

	switch catalog.NormalizeType(targetManifest.Type) {
	case catalog.TypeSource:
		_, err := s.UpgradeSourcePlugin(ctx, targetManifest.ID)
		return err
	case catalog.TypeDynamic:
		if err := s.persistDynamicPluginAuthorization(ctx, targetManifest, options.Authorization); err != nil {
			return err
		}
		if registry != nil && registry.Installed == catalog.InstalledYes {
			return s.runtimeSvc.UpgradeDynamicPluginRequest(ctx, targetManifest.ID)
		}
		return bizerr.NewCode(CodePluginNotInstalled)
	default:
		return bizerr.NewCode(
			CodePluginRuntimeUpgradeTypeUnsupported,
			bizerr.P("pluginId", targetManifest.ID),
			bizerr.P("pluginType", targetManifest.Type),
		)
	}
}

// lockRuntimeUpgrade serializes explicit runtime upgrades for one plugin within
// the current process and, when cluster mode is enabled, across all nodes via
// the configured coordination lock store.
func (s *serviceImpl) lockRuntimeUpgrade(ctx context.Context, pluginID string) (func(), error) {
	localUnlock := s.lockRuntimeUpgradeLocal(pluginID)
	if s == nil || s.topology == nil || !s.topology.IsEnabled() {
		return localUnlock, nil
	}
	if s.runtimeUpgradeLockStore == nil {
		localUnlock()
		return nil, bizerr.NewCode(
			CodePluginRuntimeUpgradeLockUnavailable,
			bizerr.P("pluginId", pluginID),
		)
	}

	lockName := runtimeUpgradeDistributedLockName(pluginID)
	owner := runtimeUpgradeDistributedLockOwner(s.topology)
	handle, ok, err := s.runtimeUpgradeLockStore.Acquire(
		ctx,
		lockName,
		owner,
		runtimeUpgradeDistributedLockReason,
		runtimeUpgradeDistributedLockLease,
	)
	if err != nil {
		localUnlock()
		return nil, bizerr.WrapCode(
			err,
			CodePluginRuntimeUpgradeLockUnavailable,
			bizerr.P("pluginId", pluginID),
		)
	}
	if !ok || handle == nil {
		localUnlock()
		return nil, bizerr.NewCode(
			CodePluginRuntimeUpgradeAlreadyRunning,
			bizerr.P("pluginId", pluginID),
		)
	}

	return func() {
		if releaseErr := s.runtimeUpgradeLockStore.Release(ctx, handle); releaseErr != nil {
			logger.Warningf(
				ctx,
				"release runtime upgrade distributed lock failed plugin=%s lock=%s err=%v",
				pluginID, lockName, releaseErr,
			)
		}
		localUnlock()
	}, nil
}

// lockRuntimeUpgradeLocal serializes explicit runtime upgrades for one plugin
// within the current process.
func (s *serviceImpl) lockRuntimeUpgradeLocal(pluginID string) func() {
	if s == nil {
		return func() {}
	}
	s.runtimeUpgradeLocksMu.Lock()
	lock := s.runtimeUpgradeLocks[pluginID]
	if lock == nil {
		lock = &sync.Mutex{}
		s.runtimeUpgradeLocks[pluginID] = lock
	}
	s.runtimeUpgradeLocksMu.Unlock()
	lock.Lock()
	return lock.Unlock
}

// runtimeUpgradeDistributedLockName builds the cluster-wide lock name for one plugin upgrade.
func runtimeUpgradeDistributedLockName(pluginID string) string {
	return "plugin-runtime-upgrade:" + strings.TrimSpace(pluginID)
}

// runtimeUpgradeDistributedLockOwner builds a unique owner for one acquisition
// so concurrent requests from the same node cannot re-enter the same lock.
func runtimeUpgradeDistributedLockOwner(topology Topology) string {
	nodeID := "local-node"
	if topology != nil && strings.TrimSpace(topology.NodeID()) != "" {
		nodeID = strings.TrimSpace(topology.NodeID())
	}
	return nodeID + ":" + guid.S()
}
