// This file declares the cron preview endpoint placeholder for scheduled jobs.

package job

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/api/job/v1"
)

// CronPreview handles requests that preview upcoming cron trigger times.
func (c *ControllerV1) CronPreview(ctx context.Context, req *v1.CronPreviewReq) (res *v1.CronPreviewRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
