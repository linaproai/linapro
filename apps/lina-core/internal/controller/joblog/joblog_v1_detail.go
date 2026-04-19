// This file declares the scheduled job log detail endpoint placeholder.

package joblog

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/api/joblog/v1"
)

// Detail handles scheduled job log detail lookup requests.
func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
