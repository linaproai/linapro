// This file declares the scheduled job log cancel endpoint placeholder.

package joblog

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/api/joblog/v1"
)

// Cancel handles scheduled job log cancellation requests.
func (c *ControllerV1) Cancel(ctx context.Context, req *v1.CancelReq) (res *v1.CancelRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
