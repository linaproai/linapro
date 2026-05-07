// This file defines topology and dependency adapters used while wiring the facade.

package plugin

import (
	"context"

	"lina-core/internal/service/bizctx"
	configsvc "lina-core/internal/service/config"
)

// runtimeTopologyAdapter adapts plugin.Topology to runtime.TopologyProvider.
type runtimeTopologyAdapter struct{ t Topology }

// IsClusterModeEnabled reports whether clustered runtime reconciliation is active.
func (a *runtimeTopologyAdapter) IsClusterModeEnabled() bool { return a.t.IsEnabled() }

// IsPrimaryNode reports whether the current host owns primary-only runtime work.
func (a *runtimeTopologyAdapter) IsPrimaryNode() bool { return a.t.IsPrimary() }

// CurrentNodeID returns the current host node identifier.
func (a *runtimeTopologyAdapter) CurrentNodeID() string { return a.t.NodeID() }

// lifecycleTopologyAdapter adapts plugin.Topology to lifecycle.TopologyProvider.
type lifecycleTopologyAdapter struct{ t Topology }

// IsPrimaryNode reports whether lifecycle mutations may run on this node.
func (a *lifecycleTopologyAdapter) IsPrimaryNode() bool { return a.t.IsPrimary() }

// integrationTopologyAdapter adapts plugin.Topology to integration.TopologyProvider.
type integrationTopologyAdapter struct{ t Topology }

// IsPrimaryNode reports whether integration tasks should run on this node.
func (a *integrationTopologyAdapter) IsPrimaryNode() bool { return a.t.IsPrimary() }

// jwtConfigAdapter adapts configsvc.Service to runtime.JwtConfigProvider.
type jwtConfigAdapter struct{ svc configsvc.Service }

// GetJwtSecret returns the runtime JWT signing secret.
func (a *jwtConfigAdapter) GetJwtSecret(ctx context.Context) string {
	return a.svc.GetJwtSecret(ctx)
}

// uploadSizeAdapter adapts configsvc.Service to runtime.UploadSizeProvider.
type uploadSizeAdapter struct{ svc configsvc.Service }

// GetUploadMaxSize returns the runtime-effective upload size limit in MB.
func (a *uploadSizeAdapter) GetUploadMaxSize(ctx context.Context) (int64, error) {
	return a.svc.GetUploadMaxSize(ctx)
}

// userCtxAdapter adapts bizctx.Service to runtime.UserContextSetter.
type userCtxAdapter struct{ svc bizctx.Service }

// SetUser injects authenticated user identity into the request context.
func (a *userCtxAdapter) SetUser(ctx context.Context, tokenID string, userID int, username string, status int) {
	a.svc.SetUser(ctx, tokenID, userID, username, status)
}

// SetUserAccess injects cached access-snapshot fields into the request context.
func (a *userCtxAdapter) SetUserAccess(ctx context.Context, dataScope int, dataScopeUnsupported bool, unsupportedDataScope int) {
	a.svc.SetUserAccess(ctx, dataScope, dataScopeUnsupported, unsupportedDataScope)
}

// bizCtxAdapter adapts bizctx.Service to integration.BizCtxProvider.
type bizCtxAdapter struct{ svc bizctx.Service }

// GetUserId returns the current request user ID for integration-layer helpers.
func (a *bizCtxAdapter) GetUserId(ctx context.Context) int {
	bizUser := a.svc.Get(ctx)
	if bizUser == nil {
		return 0
	}
	return bizUser.UserId
}

// GetDataScope returns the current request user's effective role data-scope.
func (a *bizCtxAdapter) GetDataScope(ctx context.Context) int {
	bizUser := a.svc.Get(ctx)
	if bizUser == nil {
		return 0
	}
	return bizUser.DataScope
}

// GetDataScopeUnsupported returns the unsupported data-scope state from the current request.
func (a *bizCtxAdapter) GetDataScopeUnsupported(ctx context.Context) (bool, int) {
	bizUser := a.svc.Get(ctx)
	if bizUser == nil {
		return false, 0
	}
	return bizUser.DataScopeUnsupported, bizUser.UnsupportedDataScope
}
