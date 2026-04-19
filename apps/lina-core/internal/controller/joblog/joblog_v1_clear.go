// This file declares the scheduled job log clear endpoint placeholder.

package joblog

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/api/joblog/v1"
)

// Clear handles scheduled job log cleanup requests.
func (c *ControllerV1) Clear(ctx context.Context, req *v1.ClearReq) (res *v1.ClearRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
