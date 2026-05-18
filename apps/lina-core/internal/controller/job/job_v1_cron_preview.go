// This file implements cron preview for scheduled jobs.

package job

import (
	"context"
	"time"

	"lina-core/api/job/v1"
)

// CronPreview handles requests that preview upcoming cron trigger times.
func (c *ControllerV1) CronPreview(ctx context.Context, req *v1.CronPreviewReq) (res *v1.CronPreviewRes, err error) {
	times, err := c.jobMgmtSvc.PreviewCron(ctx, req.Expr, req.Timezone)
	if err != nil {
		return nil, err
	}
	items := make([]string, 0, len(times))
	for _, item := range times {
		items = append(items, item.Format(time.RFC3339))
	}
	return &v1.CronPreviewRes{Times: items}, nil
}
