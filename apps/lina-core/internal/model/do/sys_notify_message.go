// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysNotifyMessage is the golang structure of table sys_notify_message for DAO operations like Where/Data.
type SysNotifyMessage struct {
	g.Meta       `orm:"table:sys_notify_message, do:true"`
	Id           any         // Primary key ID
	TenantId     any         // Owning tenant ID, 0 means PLATFORM
	PluginId     any         // Source plugin ID, empty for host built-in flows
	SourceType   any         // Source type: notice=notice, plugin=plugin, system=system
	SourceId     any         // Source business ID
	CategoryCode any         // Message category: notice=notification, announcement=announcement, other=other
	Title        any         // Message title
	Content      any         // Message body
	PayloadJson  any         // Extended payload JSON
	SenderUserId any         // Sender user ID
	CreatedAt    *gtime.Time // Creation time
}
