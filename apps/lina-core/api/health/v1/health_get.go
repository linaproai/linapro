// This file defines DTOs for the anonymous host health probe API.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// GetReq requests the current host health status.
type GetReq struct {
	g.Meta `path:"/health" method:"get" tags:"Host Runtime Operations" summary:"Get host health status" dc:"Run a lightweight database probe and return whether the host process can serve traffic. This endpoint is anonymous so container probes and load balancers can call it before authentication."`
}

// GetRes describes the host health probe result.
type GetRes struct {
	Status string `json:"status" dc:"Health status: ok=host can serve traffic, unavailable=health probe failed" eg:"ok"`
	Mode   string `json:"mode,omitempty" dc:"Deployment mode: single=standalone host, master=cluster primary node, slave=cluster secondary node" eg:"single"`
	Reason string `json:"reason,omitempty" dc:"Sanitized failure reason, omitted when the status is ok" eg:"database probe failed"`
}
