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
	Id          any         // Primary key ID
	ChannelKey  any         // Channel key
	Name        any         // Channel name
	ChannelType any         // Channel type: inbox=in-app message, email=email, webhook=webhook
	Status      any         // Status: 1=enabled, 0=disabled
	ConfigJson  any         // Channel configuration JSON
	Remark      any         // Remark
	CreatedAt   *gtime.Time // Creation time
	UpdatedAt   *gtime.Time // Update time
	DeletedAt   *gtime.Time // Deletion time
}
