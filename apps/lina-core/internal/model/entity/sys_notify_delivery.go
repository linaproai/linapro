// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysNotifyDelivery is the golang structure for table sys_notify_delivery.
type SysNotifyDelivery struct {
	Id             int64       `json:"id"             orm:"id"              description:"主键ID"`
	MessageId      int64       `json:"messageId"      orm:"message_id"      description:"通知消息ID"`
	ChannelKey     string      `json:"channelKey"     orm:"channel_key"     description:"投递通道标识"`
	ChannelType    string      `json:"channelType"    orm:"channel_type"    description:"投递通道类型"`
	RecipientType  string      `json:"recipientType"  orm:"recipient_type"  description:"接收者类型：user=用户 email=邮箱 webhook=Webhook"`
	RecipientKey   string      `json:"recipientKey"   orm:"recipient_key"   description:"接收者标识，如用户ID邮箱地址或Webhook标识"`
	UserId         int64       `json:"userId"         orm:"user_id"         description:"站内信用户ID，非站内信时为0"`
	DeliveryStatus int         `json:"deliveryStatus" orm:"delivery_status" description:"投递状态：0=待发送 1=成功 2=失败"`
	IsRead         int         `json:"isRead"         orm:"is_read"         description:"是否已读：0=未读 1=已读"`
	ReadAt         *gtime.Time `json:"readAt"         orm:"read_at"         description:"已读时间"`
	ErrorMessage   string      `json:"errorMessage"   orm:"error_message"   description:"失败原因"`
	SentAt         *gtime.Time `json:"sentAt"         orm:"sent_at"         description:"发送完成时间"`
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"      description:"创建时间"`
	UpdatedAt      *gtime.Time `json:"updatedAt"      orm:"updated_at"      description:"更新时间"`
	DeletedAt      *gtime.Time `json:"deletedAt"      orm:"deleted_at"      description:"删除时间"`
}
