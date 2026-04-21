// Package monitoronline exposes host online-user handlers to source plugins
// without granting direct access to internal controller packages.
package monitoronline

import (
	"context"

	v1 "lina-core/api/monitor/v1"
	internalmonitor "lina-core/internal/controller/monitor"
)

// OnlineForceLogout returns the host online-user force-logout handler.
func OnlineForceLogout() func(ctx context.Context, req *v1.OnlineForceLogoutReq) (*v1.OnlineForceLogoutRes, error) {
	controller := internalmonitor.NewV1()
	return controller.OnlineForceLogout
}

// OnlineList returns the host online-user list handler.
func OnlineList() func(ctx context.Context, req *v1.OnlineListReq) (*v1.OnlineListRes, error) {
	controller := internalmonitor.NewV1()
	return controller.OnlineList
}
