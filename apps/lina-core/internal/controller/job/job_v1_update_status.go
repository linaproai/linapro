// This file declares the scheduled job status update endpoint placeholder.

package job

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/api/job/v1"
)

// UpdateStatus handles scheduled job status change requests.
func (c *ControllerV1) UpdateStatus(ctx context.Context, req *v1.UpdateStatusReq) (res *v1.UpdateStatusRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
