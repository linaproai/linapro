// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysNotifyChannel is the golang structure for table sys_notify_channel.
type SysNotifyChannel struct {
	Id          int64       `json:"id"          orm:"id"           description:"主键ID"`
	ChannelKey  string      `json:"channelKey"  orm:"channel_key"  description:"通道标识"`
	Name        string      `json:"name"        orm:"name"         description:"通道名称"`
	ChannelType string      `json:"channelType" orm:"channel_type" description:"通道类型：inbox=站内信 email=邮件 webhook=Webhook"`
	Status      int         `json:"status"      orm:"status"       description:"状态：1=启用 0=停用"`
	ConfigJson  string      `json:"configJson"  orm:"config_json"  description:"通道配置JSON"`
	Remark      string      `json:"remark"      orm:"remark"       description:"备注"`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:"创建时间"`
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"   description:"更新时间"`
	DeletedAt   *gtime.Time `json:"deletedAt"   orm:"deleted_at"   description:"删除时间"`
}
