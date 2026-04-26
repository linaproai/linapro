package v1

import "github.com/gogf/gf/v2/frame/g"

// CronPreviewReq defines the request for previewing one cron expression.
type CronPreviewReq struct {
	g.Meta   `path:"/job/cron-preview" method:"get" tags:"Job Scheduling / Scheduled Jobs" summary:"Preview timed expression" dc:"Preview the last 5 trigger times based on the input 5-segment or 6-segment Cron expression and time zone, which is used for task form verification and display; the 5-segment expression is parsed at minute granularity, and # seconds will be added when running." permission:"system:job:list"`
	Expr     string `json:"expr" v:"required|length:1,128" dc:"The Cron expression to be previewed supports 5-segment and 6-segment writing." eg:"17 3 * * *"`
	Timezone string `json:"timezone" d:"Asia/Shanghai" dc:"Time zone identifier used when cron parses" eg:"Asia/Shanghai"`
}

// CronPreviewRes defines the response for previewing one cron expression.
type CronPreviewRes struct {
	Times []string `json:"times" dc:"List of the last 5 trigger times, using RFC3339 format" eg:"[\"2026-04-20T03:17:00+08:00\"]"`
}
