// This file defines topology and dependency adapters used while wiring the facade.

package plugin

import (
	"context"

	"lina-core/internal/service/bizctx"
	configsvc "lina-core/internal/service/config"
)

// runtimeTopologyAdapter adapts plugin.Topology to runtime.TopologyProvider.
type runtimeTopologyAdapter struct{ t Topology }

func (a *runtimeTopologyAdapter) IsClusterModeEnabled() bool { return a.t.IsEnabled() }
func (a *runtimeTopologyAdapter) IsPrimaryNode() bool        { return a.t.IsPrimary() }
func (a *runtimeTopologyAdapter) CurrentNodeID() string      { return a.t.NodeID() }

// lifecycleTopologyAdapter adapts plugin.Topology to lifecycle.TopologyProvider.
type lifecycleTopologyAdapter struct{ t Topology }

func (a *lifecycleTopologyAdapter) IsPrimaryNode() bool { return a.t.IsPrimary() }

// integrationTopologyAdapter adapts plugin.Topology to integration.TopologyProvider.
type integrationTopologyAdapter struct{ t Topology }

func (a *integrationTopologyAdapter) IsPrimaryNode() bool { return a.t.IsPrimary() }

// jwtConfigAdapter adapts configsvc.Service to runtime.JwtConfigProvider.
type jwtConfigAdapter struct{ svc configsvc.Service }

func (a *jwtConfigAdapter) GetJwtSecret(ctx context.Context) string {
	return a.svc.GetJwt(ctx).Secret
}

// userCtxAdapter adapts bizctx.Service to runtime.UserContextSetter.
type userCtxAdapter struct{ svc bizctx.Service }

func (a *userCtxAdapter) SetUser(ctx context.Context, tokenID string, userID int, username string, status int) {
	a.svc.SetUser(ctx, tokenID, userID, username, status)
}

// bizCtxAdapter adapts bizctx.Service to integration.BizCtxProvider.
type bizCtxAdapter struct{ svc bizctx.Service }

func (a *bizCtxAdapter) GetUserId(ctx context.Context) int {
	bizUser := a.svc.Get(ctx)
	if bizUser == nil {
		return 0
	}
	return bizUser.UserId
}
