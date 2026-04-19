// This file declares the scheduled job reset endpoint placeholder.

package job

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/api/job/v1"
)

// Reset handles requests that reset scheduled job execution counters.
func (c *ControllerV1) Reset(ctx context.Context, req *v1.ResetReq) (res *v1.ResetRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
