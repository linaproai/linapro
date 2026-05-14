// This file declares water snap task submission DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SubmitSnapReq defines the request for submitting one asynchronous watermark snapshot task.
type SubmitSnapReq struct {
	g.Meta      `path:"/water/snaps/{deviceType}/{deviceId}" method:"post" tags:"水印服务" summary:"提交水印截图任务" dc:"提交一张截图图片，按媒体策略异步添加水印。" permission:"water:management:submit"`
	DeviceType  string `json:"deviceType" v:"required#设备类型不能为空" dc:"设备类型" eg:"gb"`
	DeviceId    string `json:"deviceId" v:"required#设备ID不能为空" dc:"设备ID" eg:"34020000001320000001"`
	ErrorCode   string `json:"errorCode" dc:"上游错误码" eg:""`
	DeviceCode  string `json:"deviceCode" dc:"设备国标ID" eg:"34020000001320000001"`
	ChannelCode string `json:"channelCode" dc:"通道编码" eg:"34020000001320000001"`
	DeviceIdx   string `json:"deviceIdx" dc:"设备索引" eg:"1"`
	Image       string `json:"image" v:"required#图片不能为空" dc:"base64图片或data URL" eg:"data:image/png;base64,..."`
	ImageName   string `json:"imageName" dc:"图片名称" eg:"snap.png"`
	ImagePath   string `json:"imagePath" dc:"原始图片路径" eg:"/tmp/snap.png"`
	AccessNode  string `json:"accessNode" dc:"接入节点" eg:"node-a"`
	AcceptNode  string `json:"acceptNode" dc:"接收节点" eg:"node-b"`
	UploadUrl   string `json:"uploadUrl" dc:"上传地址" eg:""`
	CallbackUrl string `json:"callbackUrl" dc:"结果回调地址" eg:"https://example.com/callback"`
	Url         string `json:"url" dc:"兼容hotgo旧回调字段" eg:"https://example.com/callback"`
	User        string `json:"user" dc:"业务用户ID" eg:"user-a"`
	Tenant      string `json:"tenant" v:"required#媒体租户ID不能为空" dc:"媒体业务租户ID" eg:"tenant-a"`
}

// SubmitSnapRes defines the response for submitting one asynchronous task.
type SubmitSnapRes struct {
	Success bool   `json:"success" dc:"是否提交成功" eg:"true"`
	TaskId  string `json:"taskId" dc:"水印任务ID" eg:"wm_123"`
	Status  string `json:"status" dc:"任务状态" eg:"queued"`
}
