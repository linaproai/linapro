// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysNotifyDelivery is the golang structure of table sys_notify_delivery for DAO operations like Where/Data.
type SysNotifyDelivery struct {
	g.Meta         `orm:"table:sys_notify_delivery, do:true"`
	Id             any         // 主键ID
	MessageId      any         // 通知消息ID
	ChannelKey     any         // 投递通道标识
	ChannelType    any         // 投递通道类型
	RecipientType  any         // 接收者类型：user=用户 email=邮箱 webhook=Webhook
	RecipientKey   any         // 接收者标识，如用户ID邮箱地址或Webhook标识
	UserId         any         // 站内信用户ID，非站内信时为0
	DeliveryStatus any         // 投递状态：0=待发送 1=成功 2=失败
	IsRead         any         // 是否已读：0=未读 1=已读
	ReadAt         *gtime.Time // 已读时间
	ErrorMessage   any         // 失败原因
	SentAt         *gtime.Time // 发送完成时间
	CreatedAt      *gtime.Time // 创建时间
	UpdatedAt      *gtime.Time // 更新时间
	DeletedAt      *gtime.Time // 删除时间
}
