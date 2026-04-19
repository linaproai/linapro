// This file declares the scheduled job delete endpoint placeholder.

package job

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/api/job/v1"
)

// Delete handles scheduled job deletion requests.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
