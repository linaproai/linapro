// This file validates scheduled-job cron expressions and timezone inputs.

package jobmgmt

import (
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"lina-core/pkg/bizerr"
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
		return "", nil, bizerr.NewCode(CodeJobCronExpressionRequired)
	}
	if len(trimmedExpr) > 128 {
		return "", nil, bizerr.NewCode(CodeJobCronExpressionTooLong)
	}

	fields := strings.Fields(trimmedExpr)
	switch len(fields) {
	case 5:
		schedule, err := fiveFieldCronParser.Parse(strings.Join(fields, " "))
		if err != nil {
			return "", nil, bizerr.NewCode(
				CodeJobCronExpressionInvalid,
				bizerr.P("reason", sanitizeCronParseError(err)),
			)
		}
		return strings.Join(fields, " "), schedule, nil
	case 6:
		if fields[0] == "#" {
			return "", nil, bizerr.NewCode(CodeJobCronSecondsRequired)
		}
		schedule, err := sixFieldCronParser.Parse(strings.Join(fields, " "))
		if err != nil {
			return "", nil, bizerr.NewCode(
				CodeJobCronExpressionInvalid,
				bizerr.P("reason", sanitizeCronParseError(err)),
			)
		}
		return strings.Join(fields, " "), schedule, nil
	default:
		return "", nil, bizerr.NewCode(CodeJobCronFieldCountInvalid)
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
		return "", nil, bizerr.NewCode(CodeJobTimezoneInvalid, bizerr.P("timezone", trimmedTimezone))
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
