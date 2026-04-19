// This file implements cron expression preview for scheduled jobs.

package jobmgmt

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/robfig/cron/v3"
)

// PreviewCron returns the next five fire times for one cron expression.
func (s *serviceImpl) PreviewCron(_ context.Context, expr string, timezone string) ([]time.Time, error) {
	trimmedExpr := strings.TrimSpace(expr)
	if trimmedExpr == "" {
		return nil, gerror.New("Cron 表达式不能为空")
	}

	trimmedTimezone := strings.TrimSpace(timezone)
	if trimmedTimezone == "" {
		trimmedTimezone = "Asia/Shanghai"
	}
	location, err := time.LoadLocation(trimmedTimezone)
	if err != nil {
		return nil, gerror.Wrap(err, "任务时区不合法")
	}

	parser := cron.NewParser(
		cron.SecondOptional |
			cron.Minute |
			cron.Hour |
			cron.Dom |
			cron.Month |
			cron.Dow |
			cron.Descriptor,
	)
	schedule, err := parser.Parse(trimmedExpr)
	if err != nil {
		return nil, gerror.Wrap(err, "Cron 表达式不合法")
	}

	current := time.Now().In(location)
	result := make([]time.Time, 0, 5)
	for len(result) < 5 {
		current = schedule.Next(current)
		result = append(result, current)
	}
	return result, nil
}
