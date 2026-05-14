// This file declares synchronous water preview DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// PreviewReq defines the request for synchronously previewing watermark output.
type PreviewReq struct {
	g.Meta      `path:"/water/preview" method:"post" tags:"水印服务" summary:"预览水印效果" dc:"同步处理一张图片并返回水印效果，用于管理页测试。" permission:"water:management:preview"`
	Tenant      string `json:"tenant" v:"required#媒体租户ID不能为空" dc:"媒体业务租户ID" eg:"tenant-a"`
	DeviceId    string `json:"deviceId" dc:"设备ID" eg:"34020000001320000001"`
	DeviceCode  string `json:"deviceCode" dc:"设备国标ID" eg:"34020000001320000001"`
	ChannelCode string `json:"channelCode" dc:"通道编码" eg:"34020000001320000001"`
	Image       string `json:"image" v:"required#图片不能为空" dc:"base64图片或data URL" eg:"data:image/png;base64,..."`
}

// PreviewRes defines the synchronous watermark preview result.
type PreviewRes struct {
	Success      bool   `json:"success" dc:"是否成功" eg:"true"`
	Status       string `json:"status" dc:"处理状态" eg:"success"`
	Message      string `json:"message" dc:"结果说明" eg:"处理完成"`
	Image        string `json:"image" dc:"输出图片data URL" eg:"data:image/png;base64,..."`
	StrategyId   int64  `json:"strategyId" dc:"策略ID" eg:"1"`
	StrategyName string `json:"strategyName" dc:"策略名称" eg:"默认策略"`
	Source       string `json:"source" dc:"策略来源" eg:"global"`
	SourceLabel  string `json:"sourceLabel" dc:"策略来源说明" eg:"全局策略"`
	DurationMs   int64  `json:"durationMs" dc:"处理耗时毫秒" eg:"12"`
}
