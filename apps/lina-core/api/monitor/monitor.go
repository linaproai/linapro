// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package monitor

import (
	"context"

	"lina-core/api/monitor/v1"
)

type IMonitorV1 interface {
	OnlineForceLogout(ctx context.Context, req *v1.OnlineForceLogoutReq) (res *v1.OnlineForceLogoutRes, err error)
	OnlineList(ctx context.Context, req *v1.OnlineListReq) (res *v1.OnlineListRes, err error)
	ServerMonitor(ctx context.Context, req *v1.ServerMonitorReq) (res *v1.ServerMonitorRes, err error)
}
