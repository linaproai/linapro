// This file validates scheduled-job cron expressions and timezone inputs.

package jobmgmt

import (
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/robfig/cron/v3"
)

var (
	fiveFieldCronParser = cron.NewParser(
		cron.Minute |
			cron.Hour |
			cron.Dom |
			cron.Month |
			cron.Dow,
	)
	sixFieldCronParser = cron.NewParser(
		cron.Second |
			cron.Minute |
			cron.Hour |
			cron.Dom |
			cron.Month |
			cron.Dow,
	)
)

// normalizeCronExpression validates one cron expression and returns the trimmed
// expression plus its parsed schedule for reuse by preview and save flows.
func normalizeCronExpression(expr string) (string, cron.Schedule, error) {
	trimmedExpr := strings.TrimSpace(expr)
	if trimmedExpr == "" {
		return "", nil, gerror.New("定时表达式不能为空")
	}
	if len(trimmedExpr) > 128 {
		return "", nil, gerror.New("定时表达式长度不能超过128个字符")
	}

	fields := strings.Fields(trimmedExpr)
	switch len(fields) {
	case 5:
		schedule, err := fiveFieldCronParser.Parse(strings.Join(fields, " "))
		if err != nil {
			return "", nil, gerror.Newf("定时表达式格式不正确：%s", sanitizeCronParseError(err))
		}
		return strings.Join(fields, " "), schedule, nil
	case 6:
		if fields[0] == "#" {
			return "", nil, gerror.New("6段定时表达式的秒位必须填写具体值，5段表达式无需手工填写#")
		}
		schedule, err := sixFieldCronParser.Parse(strings.Join(fields, " "))
		if err != nil {
			return "", nil, gerror.Newf("定时表达式格式不正确：%s", sanitizeCronParseError(err))
		}
		return strings.Join(fields, " "), schedule, nil
	default:
		return "", nil, gerror.New("定时表达式仅支持5段或6段")
	}
}

// normalizeJobTimezone validates one job timezone input and applies the
// scheduled-job default when callers omit the field.
func normalizeJobTimezone(timezone string) (string, *time.Location, error) {
	trimmedTimezone := strings.TrimSpace(timezone)
	if trimmedTimezone == "" {
		trimmedTimezone = "Asia/Shanghai"
	}

	location, err := time.LoadLocation(trimmedTimezone)
	if err != nil {
		return "", nil, gerror.Newf("任务时区不合法：%s", trimmedTimezone)
	}
	return trimmedTimezone, location, nil
}

// sanitizeCronParseError keeps parser errors readable in API responses.
func sanitizeCronParseError(err error) string {
	if err == nil {
		return ""
	}
	return strings.Join(strings.Fields(err.Error()), " ")
}
