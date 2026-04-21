// Package monitorserver exposes the host server-monitor handler to source
// plugins without granting direct access to internal controller packages.
package monitorserver

import (
	"context"

	v1 "lina-core/api/monitor/v1"
	internalmonitor "lina-core/internal/controller/monitor"
)

// ServerMonitor returns the host server-monitor query handler.
func ServerMonitor() func(ctx context.Context, req *v1.ServerMonitorReq) (*v1.ServerMonitorRes, error) {
	controller := internalmonitor.NewV1()
	return controller.ServerMonitor
}
