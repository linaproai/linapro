// This file declares water task query DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetTaskReq defines the request for querying one watermark task.
type GetTaskReq struct {
	g.Meta `path:"/water/tasks/{taskId}" method:"get" tags:"水印服务" summary:"查询水印任务" dc:"根据任务ID查询最近水印任务状态。" permission:"water:management:query"`
	TaskId string `json:"taskId" v:"required#任务ID不能为空" dc:"任务ID" eg:"wm_123"`
}

// GetTaskRes defines one watermark task status.
type GetTaskRes struct {
	TaskId       string `json:"taskId" dc:"任务ID" eg:"wm_123"`
	Status       string `json:"status" dc:"任务状态" eg:"success"`
	Success      bool   `json:"success" dc:"是否处理成功" eg:"true"`
	Message      string `json:"message" dc:"任务消息" eg:"处理完成"`
	Error        string `json:"error" dc:"错误信息" eg:""`
	Tenant       string `json:"tenant" dc:"媒体租户ID" eg:"tenant-a"`
	DeviceId     string `json:"deviceId" dc:"设备ID" eg:"34020000001320000001"`
	StrategyId   int64  `json:"strategyId" dc:"策略ID" eg:"1"`
	StrategyName string `json:"strategyName" dc:"策略名称" eg:"默认策略"`
	Source       string `json:"source" dc:"策略来源" eg:"global"`
	SourceLabel  string `json:"sourceLabel" dc:"策略来源说明" eg:"全局策略"`
	Image        string `json:"image" dc:"输出图片data URL" eg:"data:image/png;base64,..."`
	CreatedAt    string `json:"createdAt" dc:"创建时间" eg:"2026-05-14 10:00:00"`
	UpdatedAt    string `json:"updatedAt" dc:"更新时间" eg:"2026-05-14 10:00:01"`
	DurationMs   int64  `json:"durationMs" dc:"处理耗时毫秒" eg:"12"`
}
