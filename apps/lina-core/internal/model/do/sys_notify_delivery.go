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
	Id             any         // Primary key ID
	TenantId       any         // Owning tenant ID, 0 means PLATFORM
	MessageId      any         // Notification message ID
	ChannelKey     any         // Delivery channel key
	ChannelType    any         // Delivery channel type
	RecipientType  any         // Recipient type: user=user, email=email, webhook=webhook
	RecipientKey   any         // Recipient key such as user ID, email address, or webhook identifier
	UserId         any         // In-app message user ID, 0 for non-in-app delivery
	DeliveryStatus any         // Delivery status: 0=pending, 1=succeeded, 2=failed
	IsRead         any         // Read flag: 0=unread, 1=read
	ReadAt         *gtime.Time // Read time
	ErrorMessage   any         // Failure reason
	SentAt         *gtime.Time // Send completion time
	CreatedAt      *gtime.Time // Creation time
	UpdatedAt      *gtime.Time // Update time
	DeletedAt      *gtime.Time // Deletion time
}
