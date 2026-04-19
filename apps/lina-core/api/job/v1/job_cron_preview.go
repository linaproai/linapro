package v1

import "github.com/gogf/gf/v2/frame/g"

// CronPreviewReq defines the request for previewing one cron expression.
type CronPreviewReq struct {
	g.Meta   `path:"/job/cron-preview" method:"get" tags:"定时任务管理" summary:"预览 Cron 表达式" dc:"根据输入的 Cron 表达式和时区预览最近5次触发时间，用于任务表单校验与展示" permission:"system:job:list"`
	Expr     string `json:"expr" v:"required|length:1,128" dc:"待预览的 Cron 表达式" eg:"17 3 * * *"`
	Timezone string `json:"timezone" d:"Asia/Shanghai" dc:"Cron 解析时使用的时区标识" eg:"Asia/Shanghai"`
}

// CronPreviewRes defines the response for previewing one cron expression.
type CronPreviewRes struct {
	Times []string `json:"times" dc:"最近5次触发时间列表，使用 RFC3339 格式" eg:"[\"2026-04-20T03:17:00+08:00\"]"`
}
