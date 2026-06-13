// This file defines topology contracts and dependency adapters used while wiring the facade.

package plugin

import (
	"context"
	"time"

	"lina-core/internal/service/bizctx"
	configsvc "lina-core/internal/service/config"
)

// Topology defines the cluster semantics required by plugin runtime behavior.
type Topology interface {
	// IsEnabled reports whether the host is running in clustered mode.
	IsEnabled() bool
	// IsPrimary reports whether the current node is the primary node.
	IsPrimary() bool
	// NodeID returns the stable identifier of the current node.
	NodeID() string
}

// singleNodeTopology provides the default topology used when clustering is disabled.
type singleNodeTopology struct{}

// IsEnabled reports false because the default topology is always single-node.
func (singleNodeTopology) IsEnabled() bool {
	return false
}

// IsPrimary reports true because the only node is also the primary node.
func (singleNodeTopology) IsPrimary() bool {
	return true
}

// NodeID returns the stable placeholder node identifier for single-node mode.
func (singleNodeTopology) NodeID() string {
	return "local-node"
}

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

// IsClusterModeEnabled reports whether lifecycle startup waits should use
// primary-node coordination.
func (a *lifecycleTopologyAdapter) IsClusterModeEnabled() bool { return a.t.IsEnabled() }

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

// GetSessionTimeout returns the runtime-effective session timeout.
func (a *jwtConfigAdapter) GetSessionTimeout(ctx context.Context) (time.Duration, error) {
	return a.svc.GetSessionTimeout(ctx)
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
func (a *userCtxAdapter) SetUser(ctx context.Context, tokenID string, userID int, username string, status int, clientType string) {
	a.svc.SetUser(ctx, tokenID, userID, username, status, clientType)
}

// SetTenant injects the resolved tenant into the request context.
func (a *userCtxAdapter) SetTenant(ctx context.Context, tenantID int) {
	a.svc.SetTenant(ctx, tenantID)
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
