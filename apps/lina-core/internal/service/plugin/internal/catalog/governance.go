// This file builds governance summary projections from release, migration,
// resource-reference, and node-state records for the plugin management UI.

package catalog

import (
	"context"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// GovernanceSnapshot aggregates the review-oriented governance data shown in the plugin management UI.
type GovernanceSnapshot struct {
	// ReleaseVersion is the version string of the currently active release.
	ReleaseVersion string
	// LifecycleState is the derived lifecycle state key.
	LifecycleState string
	// NodeState is the current node state projection.
	NodeState string
	// ResourceCount is the number of resource reference rows for the active release.
	ResourceCount int
	// MigrationState is the review-friendly migration state key.
	MigrationState string
}

// BuildGovernanceSnapshot loads the current governance projection for one plugin version.
func (s *serviceImpl) BuildGovernanceSnapshot(
	ctx context.Context,
	pluginID string,
	version string,
	pluginType string,
	installed int,
	enabled int,
) (*GovernanceSnapshot, error) {
	snapshot := &GovernanceSnapshot{
		ReleaseVersion: version,
		LifecycleState: DeriveLifecycleState(pluginType, installed, enabled),
		NodeState:      DeriveNodeState(installed, enabled),
		MigrationState: MigrationStateNone.String(),
	}

	release, err := s.GetRelease(ctx, pluginID, version)
	if err != nil {
		return nil, err
	}
	if release == nil {
		release, err = s.GetActiveRelease(ctx, pluginID)
		if err != nil {
			return nil, err
		}
	}
	if release != nil && strings.TrimSpace(release.ReleaseVersion) != "" {
		snapshot.ReleaseVersion = release.ReleaseVersion
	}

	if release != nil {
		resourceCount, countErr := dao.SysPluginResourceRef.Ctx(ctx).
			Where(do.SysPluginResourceRef{
				PluginId:  pluginID,
				ReleaseId: release.Id,
			}).
			Count()
		if countErr != nil {
			return nil, countErr
		}
		snapshot.ResourceCount = resourceCount
	}

	if s.nodeStateSyncer != nil {
		nodeState, stateErr := s.nodeStateSyncer.GetPluginNodeState(ctx, pluginID, s.nodeStateSyncer.CurrentNodeID())
		if stateErr != nil {
			return nil, stateErr
		}
		if nodeState != nil && strings.TrimSpace(nodeState.CurrentState) != "" {
			snapshot.NodeState = nodeState.CurrentState
		}
	}

	if release != nil {
		latestMigration, migrationErr := s.getLatestMigration(ctx, pluginID, release.Id)
		if migrationErr != nil {
			return nil, migrationErr
		}
		if latestMigration != nil {
			snapshot.MigrationState = DeriveMigrationState(latestMigration)
		}
	}

	return snapshot, nil
}

// getLatestMigration returns the newest migration record for one plugin release.
func (s *serviceImpl) getLatestMigration(ctx context.Context, pluginID string, releaseID int) (*entity.SysPluginMigration, error) {
	if releaseID <= 0 {
		return nil, nil
	}
	var migration *entity.SysPluginMigration
	err := dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId:  pluginID,
			ReleaseId: releaseID,
		}).
		OrderDesc(dao.SysPluginMigration.Columns().Id).
		Scan(&migration)
	return migration, err
}

// DeriveMigrationState converts the latest migration row into the review-friendly state key.
func DeriveMigrationState(migration *entity.SysPluginMigration) string {
	if migration == nil {
		return MigrationStateNone.String()
	}
	if migration.Status == MigrationExecutionStatusSucceeded.String() {
		return MigrationStateSucceeded.String()
	}
	return MigrationStateFailed.String()
}
