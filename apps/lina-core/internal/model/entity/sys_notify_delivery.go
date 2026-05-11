// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysNotifyDelivery is the golang structure for table sys_notify_delivery.
type SysNotifyDelivery struct {
	Id             int64       `json:"id"             orm:"id"              description:"Primary key ID"`
	TenantId       int         `json:"tenantId"       orm:"tenant_id"       description:"Owning tenant ID, 0 means PLATFORM"`
	MessageId      int64       `json:"messageId"      orm:"message_id"      description:"Notification message ID"`
	ChannelKey     string      `json:"channelKey"     orm:"channel_key"     description:"Delivery channel key"`
	ChannelType    string      `json:"channelType"    orm:"channel_type"    description:"Delivery channel type"`
	RecipientType  string      `json:"recipientType"  orm:"recipient_type"  description:"Recipient type: user=user, email=email, webhook=webhook"`
	RecipientKey   string      `json:"recipientKey"   orm:"recipient_key"   description:"Recipient key such as user ID, email address, or webhook identifier"`
	UserId         int64       `json:"userId"         orm:"user_id"         description:"In-app message user ID, 0 for non-in-app delivery"`
	DeliveryStatus int         `json:"deliveryStatus" orm:"delivery_status" description:"Delivery status: 0=pending, 1=succeeded, 2=failed"`
	IsRead         int         `json:"isRead"         orm:"is_read"         description:"Read flag: 0=unread, 1=read"`
	ReadAt         *gtime.Time `json:"readAt"         orm:"read_at"         description:"Read time"`
	ErrorMessage   string      `json:"errorMessage"   orm:"error_message"   description:"Failure reason"`
	SentAt         *gtime.Time `json:"sentAt"         orm:"sent_at"         description:"Send completion time"`
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"      description:"Creation time"`
	UpdatedAt      *gtime.Time `json:"updatedAt"      orm:"updated_at"      description:"Update time"`
	DeletedAt      *gtime.Time `json:"deletedAt"      orm:"deleted_at"      description:"Deletion time"`
}
