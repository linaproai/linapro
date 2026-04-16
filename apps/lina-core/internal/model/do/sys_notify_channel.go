// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysNotifyChannel is the golang structure of table sys_notify_channel for DAO operations like Where/Data.
type SysNotifyChannel struct {
	g.Meta      `orm:"table:sys_notify_channel, do:true"`
	Id          any         // 主键ID
	ChannelKey  any         // 通道标识
	Name        any         // 通道名称
	ChannelType any         // 通道类型：inbox=站内信 email=邮件 webhook=Webhook
	Status      any         // 状态：1=启用 0=停用
	ConfigJson  any         // 通道配置JSON
	Remark      any         // 备注
	CreatedAt   *gtime.Time // 创建时间
	UpdatedAt   *gtime.Time // 更新时间
	DeletedAt   *gtime.Time // 删除时间
}
