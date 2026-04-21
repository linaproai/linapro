// Package monitor declares the HTTP controller contract exposed by the
// monitor-online source plugin.
package monitor

import (
	"context"

	"lina-plugin-monitor-online/backend/api/monitor/v1"
)

// IMonitorV1 defines the monitor-online HTTP handlers.
type IMonitorV1 interface {
	// OnlineForceLogout invalidates one active online-user session.
	OnlineForceLogout(ctx context.Context, req *v1.OnlineForceLogoutReq) (res *v1.OnlineForceLogoutRes, err error)
	// OnlineList returns one paginated online-user list.
	OnlineList(ctx context.Context, req *v1.OnlineListReq) (res *v1.OnlineListRes, err error)
}
