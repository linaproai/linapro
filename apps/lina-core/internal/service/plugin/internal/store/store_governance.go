// This file builds governance summary projections from release, migration,
// resource-reference, and node-state records for the plugin management UI.

package store

import (
	"context"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/plugintypes"
)

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
		LifecycleState: plugintypes.DeriveLifecycleState(pluginType, installed, enabled),
		NodeState:      plugintypes.DeriveNodeState(installed, enabled),
		MigrationState: plugintypes.MigrationStateNone.String(),
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

	if s.nodeIDProvider != nil {
		nodeID := strings.TrimSpace(s.nodeIDProvider.CurrentNodeID())
		if nodeID != "" {
			var nodeState *NodeStateRecord
			stateErr := dao.SysPluginNodeState.Ctx(ctx).
				Where(do.SysPluginNodeState{
					PluginId: pluginID,
					NodeKey:  nodeID,
				}).
				Scan(&nodeState)
			if stateErr != nil {
				return nil, stateErr
			}
			if nodeState != nil && strings.TrimSpace(nodeState.CurrentState) != "" {
				snapshot.NodeState = nodeState.CurrentState
			}
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
func (s *serviceImpl) getLatestMigration(ctx context.Context, pluginID string, releaseID int) (*MigrationRecord, error) {
	if releaseID <= 0 {
		return nil, nil
	}
	var migration *MigrationRecord
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
func DeriveMigrationState(migration *MigrationRecord) string {
	if migration == nil {
		return plugintypes.MigrationStateNone.String()
	}
	if migration.Status == plugintypes.MigrationExecutionStatusSucceeded.String() {
		return plugintypes.MigrationStateSucceeded.String()
	}
	return plugintypes.MigrationStateFailed.String()
}
