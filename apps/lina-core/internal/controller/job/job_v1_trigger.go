// This file declares the scheduled job manual trigger endpoint placeholder.

package job

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/api/job/v1"
)

// Trigger handles requests that trigger one scheduled job immediately.
func (c *ControllerV1) Trigger(ctx context.Context, req *v1.TriggerReq) (res *v1.TriggerRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
