// This file implements the anonymous host health probe controller.

package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/api/health/v1"
	"lina-core/internal/dao"
	"lina-core/pkg/logger"
)

// Health response statuses.
const (
	healthStatusOK          = "ok"
	healthStatusUnavailable = "unavailable"
)

// Health response deployment modes.
const (
	healthModeSingle = "single"
	healthModeMaster = "master"
	healthModeSlave  = "slave"
)

// defaultHealthProbeTimeout is used only when the config service is absent.
const defaultHealthProbeTimeout = 5 * time.Second

// Get returns the current host health status for anonymous probes.
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	timeout := defaultHealthProbeTimeout
	if c != nil && c.configSvc != nil {
		if cfg := c.configSvc.GetHealth(ctx); cfg != nil && cfg.Timeout > 0 {
			timeout = cfg.Timeout
		}
	}

	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if _, err = dao.SysUser.Ctx(probeCtx).Limit(1).Count(); err != nil {
		logger.Warningf(ctx, "health database probe failed: %v", err)
		if request := g.RequestFromCtx(ctx); request != nil {
			request.Response.WriteStatus(http.StatusServiceUnavailable)
		}
		return &v1.GetRes{
			Status: healthStatusUnavailable,
			Reason: "database probe failed",
		}, nil
	}

	return &v1.GetRes{
		Status: healthStatusOK,
		Mode:   c.resolveMode(),
	}, nil
}

// resolveMode maps cluster state into the public health response mode.
func (c *ControllerV1) resolveMode() string {
	if c == nil || c.clusterSvc == nil || !c.clusterSvc.IsEnabled() {
		return healthModeSingle
	}
	if c.clusterSvc.IsPrimary() {
		return healthModeMaster
	}
	return healthModeSlave
}
