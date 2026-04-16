// This file persists current-node lifecycle projections so the host can track
// the observed plugin state of each release on the local node.

package runtime

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
)

// nodeProjectionInput collects the state fields used by the node projection upsert.
type nodeProjectionInput struct {
	PluginID     string
	ReleaseID    int
	DesiredState string
	CurrentState string
	Generation   int64
	Message      string
}

// SyncPluginNodeState implements catalog.NodeStateSyncer.
// It updates the current node projection of one plugin lifecycle state.
func (s *serviceImpl) SyncPluginNodeState(
	ctx context.Context,
	pluginID string,
	version string,
	installed int,
	enabled int,
	message string,
) error {
	if !s.isClusterModeEnabled() {
		return nil
	}

	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil {
		_, releaseErr := s.catalogSvc.GetRelease(ctx, pluginID, version)
		if releaseErr != nil {
			return releaseErr
		}
		return s.syncNodeProjection(ctx, nodeProjectionInput{
			PluginID:     pluginID,
			ReleaseID:    0,
			DesiredState: catalog.DeriveNodeState(installed, enabled),
			CurrentState: catalog.DeriveNodeState(installed, enabled),
			Generation:   int64(1),
			Message:      message,
		})
	}

	desiredState := strings.TrimSpace(registry.DesiredState)
	if desiredState == "" {
		desiredState = catalog.DeriveNodeState(installed, enabled)
	}
	currentState := strings.TrimSpace(registry.CurrentState)
	if currentState == "" {
		currentState = catalog.DeriveNodeState(installed, enabled)
	}
	generation := registry.Generation
	if generation <= 0 {
		generation = 1
	}
	return s.syncNodeProjection(ctx, nodeProjectionInput{
		PluginID:     registry.PluginId,
		ReleaseID:    registry.ReleaseId,
		DesiredState: desiredState,
		CurrentState: currentState,
		Generation:   generation,
		Message:      message,
	})
}

// GetPluginNodeState implements catalog.NodeStateSyncer.
// It returns the latest node projection row for one plugin/node pair.
func (s *serviceImpl) GetPluginNodeState(ctx context.Context, pluginID string, nodeID string) (*entity.SysPluginNodeState, error) {
	if !s.isClusterModeEnabled() {
		return nil, nil
	}

	var state *entity.SysPluginNodeState
	err := dao.SysPluginNodeState.Ctx(ctx).
		Where(do.SysPluginNodeState{
			PluginId: pluginID,
			NodeKey:  nodeID,
		}).
		Scan(&state)
	return state, err
}

// CurrentNodeID implements catalog.NodeStateSyncer.
func (s *serviceImpl) CurrentNodeID() string {
	return s.currentNodeID()
}

// SyncPluginReleaseRuntimeState implements catalog.ReleaseStateSyncer.
// It updates the active release row to reflect current registry state.
func (s *serviceImpl) SyncPluginReleaseRuntimeState(ctx context.Context, registry *entity.SysPlugin) error {
	if registry == nil || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return nil
	}

	release, err := s.catalogSvc.GetRelease(ctx, registry.PluginId, registry.Version)
	if err != nil {
		return err
	}
	if release == nil {
		return nil
	}

	status := catalog.BuildReleaseStatus(registry.Installed, registry.Status)
	return s.catalogSvc.UpdateReleaseState(ctx, release.Id, status, "")
}

// syncNodeProjection upserts the node state row for the given plugin/node pair.
func (s *serviceImpl) syncNodeProjection(ctx context.Context, in nodeProjectionInput) error {
	if !s.isClusterModeEnabled() {
		return nil
	}

	pluginID := strings.TrimSpace(in.PluginID)
	nodeKey := s.currentNodeID()
	desiredState := strings.TrimSpace(in.DesiredState)
	if desiredState == "" {
		desiredState = catalog.NodeStateUninstalled.String()
	}
	currentState := strings.TrimSpace(in.CurrentState)
	if currentState == "" {
		currentState = desiredState
	}
	generation := in.Generation
	if generation <= 0 {
		generation = 1
	}

	data := do.SysPluginNodeState{
		PluginId:        pluginID,
		ReleaseId:       in.ReleaseID,
		NodeKey:         nodeKey,
		DesiredState:    desiredState,
		CurrentState:    currentState,
		Generation:      generation,
		LastHeartbeatAt: gtime.Now(),
		ErrorMessage:    strings.TrimSpace(in.Message),
	}

	existing, err := s.GetPluginNodeState(ctx, pluginID, nodeKey)
	if err != nil {
		return err
	}
	if existing == nil {
		_, err = dao.SysPluginNodeState.Ctx(ctx).Data(data).Insert()
		return err
	}

	_, err = dao.SysPluginNodeState.Ctx(ctx).
		Where(do.SysPluginNodeState{Id: existing.Id}).
		Data(data).
		Update()
	return err
}
